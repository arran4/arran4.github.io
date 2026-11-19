package main

import (
 "context"
 "encoding/json"
 "fmt"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func sample(name string)repository{
 return repository{NodeID:"R_test_"+name,Name:name,FullName:"arran4/"+name,UpdatedAt:"2026-11-17T00:00:00Z",Owner:owner{Login:"arran4",HTMLURL:"https://github.com/arran4",AvatarURL:"https://github.com/arran4.png"},HTMLURL:"https://github.com/arran4/"+name,DefaultBranch:"main",Topics:[]string{"go"},Stars:7}
}
func serveJSON(t *testing.T,fn func(http.ResponseWriter,*http.Request))*httptest.Server{
 t.Helper();return httptest.NewServer(http.HandlerFunc(fn))
}
func newClient(s *httptest.Server)client{
 return client{http:s.Client(),baseURL:s.URL,pause:0,sleep:func(time.Duration){},maxPages:3}
}
func TestCompletePaginationAndStableMetadata(t *testing.T){
 rows:=make([]repository,100)
 for i:=range rows{rows[i]=sample(fmt.Sprintf("project-%03d",i))}
 calls:=0
 s:=serveJSON(t,func(w http.ResponseWriter,r *http.Request){
  calls++
  if !strings.Contains(r.URL.RawQuery,"sort=full_name"){t.Errorf("repository sort not deterministic: %s",r.URL.RawQuery)}
  if r.URL.Query().Get("page")=="1"{json.NewEncoder(w).Encode(rows)}else{json.NewEncoder(w).Encode([]repository{sample("project-100")})}
 });defer s.Close()
 got,err:=newClient(s).collect(context.Background(),"repos")
 if err!=nil{t.Fatal(err)}
 if len(got)!=101||calls!=2{t.Fatalf("got %d repositories across %d requests",len(got),calls)}
 if got[0].NodeID!="R_test_project-000"||got[100].FullName!="arran4/project-100"{t.Fatal("lost stable identifiers")}
}
func TestRateLimitDoesNotReplacePriorSnapshots(t *testing.T){
 dir:=t.TempDir();old:=[]byte("[{\"node_id\":\"preserve\"}]\n")
 if err:=os.MkdirAll(filepath.Join(dir,"github"),0755);err!=nil{t.Fatal(err)}
 path:=filepath.Join(dir,"github","repositories.json")
 if err:=os.WriteFile(path,old,0644);err!=nil{t.Fatal(err)}
 s:=serveJSON(t,func(w http.ResponseWriter,r *http.Request){
  if strings.Contains(r.URL.Path,"/starred"){w.Header().Set("X-RateLimit-Remaining","0");w.WriteHeader(http.StatusForbidden)}else{json.NewEncoder(w).Encode([]repository{sample("ok")})}
 });defer s.Close()
 err:=run(context.Background(),newClient(s),dir)
 if err==nil{t.Fatal("expected failed refresh")}
 current,err:=os.ReadFile(path)
 if err!=nil||string(current)!=string(old){t.Fatalf("snapshot changed despite failed refresh: %v",err)}
}
func TestRejectPrivateAndDuplicateRepositories(t *testing.T){
 for _,tc:=range []struct{name string;items []repository}{
  {"private",func()[]repository{r:=sample("secret");r.Private=true;return []repository{r}}()},
  {"duplicate",[]repository{sample("duplicate"),sample("duplicate")}},
  {"different owner",func()[]repository{r:=sample("fork");r.Owner.Login="someoneelse";return []repository{r}}()},
 }{
  t.Run(tc.name,func(t *testing.T){
   s:=serveJSON(t,func(w http.ResponseWriter,r *http.Request){json.NewEncoder(w).Encode(tc.items)});defer s.Close()
   if _,err:=newClient(s).collect(context.Background(),"repos");err==nil{t.Fatal("invalid snapshot accepted")}
  })
 }
}
func TestRejectLargeUnexpectedShrink(t *testing.T){
 path:=filepath.Join(t.TempDir(),"existing.json")
 old:=make([]repository,100)
 for i:=range old{old[i]=sample(fmt.Sprintf("repo-%d",i))}
 b,_:=marshal(old)
 if err:=os.WriteFile(path,b,0644);err!=nil{t.Fatal(err)}
 if err:=guardSnapshot(path,old[:80]);err==nil{t.Fatal("large unexpected deletion accepted")}
 if err:=guardSnapshot(path,old[:95]);err!=nil{t.Fatal(err)}
}
func TestSnapshotOrderAndZeroCounts(t *testing.T){
 dir:=t.TempDir();b,a:=sample("z"),sample("a");a.Stars=0
 if err:=writeSnapshots(dir,[]repository{b,a},[]repository{sample("starred")});err!=nil{t.Fatal(err)}
 data,err:=os.ReadFile(filepath.Join(dir,"stars.json"));if err!=nil{t.Fatal(err)}
 var counts map[string]int
 if err:=json.Unmarshal(data,&counts);err!=nil{t.Fatal(err)}
 if len(counts)!=2||counts["arran4/z"]!=7||counts["arran4/a"]!=0{t.Fatal("star counts mismatch")}
}

func TestSecondaryRateLimitRetryAfter(t *testing.T) {
 attempts:=0
 var delays []time.Duration
 s:=serveJSON(t,func(w http.ResponseWriter,r *http.Request){
  attempts++
  if attempts==1{
   w.Header().Set("Retry-After","3")
   w.WriteHeader(http.StatusTooManyRequests)
   return
  }
  json.NewEncoder(w).Encode([]repository{sample("success")})
 });defer s.Close()
 c:=newClient(s)
 c.sleep=func(d time.Duration){delays=append(delays,d)}
 repos,err:=c.collect(context.Background(),"repos")
 if err!=nil{t.Fatal(err)}
 if len(repos)!=1||attempts!=2||len(delays)!=1||delays[0]!=3*time.Second{t.Fatalf("unexpected retry: repos=%d attempts=%d delays=%v",len(repos),attempts,delays)}
}
func TestExcessiveRetryAfterFailsWithoutWaiting(t *testing.T) {
 var waited bool
 s:=serveJSON(t,func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Retry-After","3600")
  w.WriteHeader(http.StatusTooManyRequests)
 });defer s.Close()
 c:=newClient(s);c.sleep=func(time.Duration){waited=true}
 if _,err:=c.collect(context.Background(),"repos");err==nil||waited{t.Fatal("unbounded rate-limit wait accepted")}
}
