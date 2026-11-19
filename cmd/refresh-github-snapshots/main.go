// Command refresh-github-snapshots updates public GitHub metadata independently of Hugo builds.
package main

import (
 "context"
 "encoding/json"
 "errors"
 "flag"
 "fmt"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "sort"
 "strconv"
 "strings"
 "time"
)
const pageSize=100
type owner struct {
 Login string `json:"login"`
 HTMLURL string `json:"html_url"`
 AvatarURL string `json:"avatar_url"`
}
type repository struct {
 NodeID string `json:"node_id"`
 Name string `json:"name"`
 FullName string `json:"full_name"`
 Description string `json:"description"`
 UpdatedAt string `json:"updated_at"`
 Owner owner `json:"owner"`
 Language string `json:"language"`
 Homepage string `json:"homepage"`
 HasPages bool `json:"has_pages"`
 HTMLURL string `json:"html_url"`
 Topics []string `json:"topics"`
 DefaultBranch string `json:"default_branch"`
 Private bool `json:"private"`
 Archived bool `json:"archived"`
 Fork bool `json:"fork"`
 Stars int `json:"stargazers_count"`
}
type client struct {
 http *http.Client
 baseURL string
 token string
 pause time.Duration
 sleep func(time.Duration)
 maxPages int
}
func (c client) page(ctx context.Context, collection string, number int) ([]repository,error) {
 endpoint:=fmt.Sprintf("%s/users/arran4/%s?per_page=%d&page=%d",strings.TrimRight(c.baseURL,"/"),collection,pageSize,number)
 if collection=="repos" { endpoint+="&sort=full_name&direction=asc" }
 for attempt:=0;attempt<3;attempt++ {
  req,err:=http.NewRequestWithContext(ctx,http.MethodGet,endpoint,nil)
  if err!=nil{return nil,err}
  req.Header.Set("Accept","application/vnd.github+json")
  req.Header.Set("User-Agent","arran4-github-pages-snapshot")
  req.Header.Set("X-GitHub-Api-Version","2022-11-28")
  if c.token!="" {req.Header.Set("Authorization","Bearer "+c.token)}
  resp,err:=c.http.Do(req)
  if err==nil {
   payload,readErr:=io.ReadAll(io.LimitReader(resp.Body,12<<20));resp.Body.Close()
   if readErr!=nil{return nil,fmt.Errorf("read %s page %d: %w",collection,number,readErr)}
   if resp.StatusCode==http.StatusOK {
    var batch []repository
    if err:=json.Unmarshal(payload,&batch);err!=nil{return nil,fmt.Errorf("invalid %s page %d: %w",collection,number,err)}
    if len(batch)>pageSize{return nil,fmt.Errorf("%s page %d exceeds page size",collection,number)}
    return batch,nil
   }
   if resp.StatusCode!=http.StatusForbidden && resp.StatusCode!=http.StatusTooManyRequests && resp.StatusCode<500{return nil,fmt.Errorf("%s page %d: HTTP %d",collection,number,resp.StatusCode)}
   if resp.Header.Get("X-RateLimit-Remaining")=="0" {return nil,fmt.Errorf("%s page %d: HTTP %d, primary rate limit exhausted; retaining previous snapshots",collection,number,resp.StatusCode)}
   if attempt==2{return nil,fmt.Errorf("%s page %d: HTTP %d after bounded retries",collection,number,resp.StatusCode)}
   wait:=time.Duration(2<<attempt)*time.Second
   if after:=resp.Header.Get("Retry-After");after!="" {
    seconds,parseErr:=strconv.Atoi(after)
    if parseErr!=nil||seconds<0||seconds>30{return nil,fmt.Errorf("%s page %d: retry-after %q exceeds bounded retry policy",collection,number,after)}
    wait=time.Duration(seconds)*time.Second
   }
   fmt.Fprintf(os.Stderr,"::warning::%s page %d returned HTTP %d; retrying after %s\n",collection,number,resp.StatusCode,wait)
   c.sleep(wait);continue
  }
  if attempt==2{return nil,fmt.Errorf("%s page %d: %w",collection,number,err)}
  c.sleep(time.Duration(2<<attempt)*time.Second)
 }
 return nil,errors.New("unreachable retry limit")
}
func (c client) collect(ctx context.Context,collection string) ([]repository,error) {
 result:=make([]repository,0)
 for n:=1;n<=c.maxPages;n++ {
  batch,err:=c.page(ctx,collection,n)
  if err!=nil{return nil,err}
  result=append(result,batch...)
  if len(batch)<pageSize{break}
  if n==c.maxPages{return nil,fmt.Errorf("%s exceeded %d pages; snapshot incomplete",collection,c.maxPages)}
  c.sleep(c.pause)
 }
 if len(result)==0{return nil,fmt.Errorf("%s returned an empty list",collection)}
 seenNames:=make(map[string]bool,len(result));seenIDs:=make(map[string]bool,len(result))
 for i:=range result {
  r:=&result[i]
  if r.NodeID==""||r.Name==""||r.FullName==""||r.Owner.Login==""||r.DefaultBranch==""||r.HTMLURL!="https://github.com/"+r.FullName||r.Private||r.Stars<0{return nil,fmt.Errorf("%s returned invalid/non-public repository %q",collection,r.FullName)}
  if collection=="repos"&&(r.Owner.Login!="arran4"||!strings.HasPrefix(r.FullName,"arran4/")){return nil,fmt.Errorf("public repositories unexpectedly contain %q",r.FullName)}
  if _,err:=time.Parse(time.RFC3339,r.UpdatedAt);err!=nil{return nil,fmt.Errorf("%s has invalid updated_at: %w",r.FullName,err)}
  if seenNames[r.FullName]||seenIDs[r.NodeID]{return nil,fmt.Errorf("%s has duplicate repository %q",collection,r.FullName)}
  seenNames[r.FullName]=true;seenIDs[r.NodeID]=true
  if r.Topics==nil{r.Topics=[]string{}}
 }
 sort.Slice(result,func(i,j int)bool{return result[i].FullName<result[j].FullName})
 return result,nil
}
func marshal(value any)([]byte,error){
 b,err:=json.MarshalIndent(value,"","  ")
 if err!=nil{return nil,err}
 return append(b,'\n'),nil
}
func guardSnapshot(path string,next []repository)error{
 previous,err:=os.ReadFile(path)
 if errors.Is(err,os.ErrNotExist){return nil}
 if err!=nil{return err}
 var old []repository
 if err:=json.Unmarshal(previous,&old);err!=nil{return fmt.Errorf("read existing %s: %w",path,err)}
 if len(old)>0&&len(next)*10<len(old)*9{return fmt.Errorf("%s shrank from %d to %d entries (>10%%); not replacing until reviewed",path,len(old),len(next))}
 return nil
}
// Stage all outputs before replacing any destination. An API error never changes snapshots.
func writeSnapshots(directory string,repos,starred []repository)error{
 repoPath:=filepath.Join(directory,"github","repositories.json")
 starPath:=filepath.Join(directory,"github","starred.json")
 if err:=guardSnapshot(repoPath,repos);err!=nil{return err}
 if err:=guardSnapshot(starPath,starred);err!=nil{return err}
 counts:=map[string]int{}
 for _,r:=range repos{counts[r.FullName]=r.Stars}
 payloads:=map[string]any{repoPath:repos,starPath:starred,filepath.Join(directory,"stars.json"):counts}
 tmpFiles:=map[string]string{}
 defer func(){for _,p:=range tmpFiles{os.Remove(p)}}()
 for dst,data:=range payloads{
  if err:=os.MkdirAll(filepath.Dir(dst),0755);err!=nil{return err}
  b,err:=marshal(data);if err!=nil{return err}
  f,err:=os.CreateTemp(filepath.Dir(dst),".snapshot-*");if err!=nil{return err}
  tmpFiles[dst]=f.Name()
  if _,err=f.Write(b);err!=nil{f.Close();return err}
  if err:=f.Close();err!=nil{return err}
 }
 for dst,tmp:=range tmpFiles{if err:=os.Rename(tmp,dst);err!=nil{return err}}
 return nil
}
func run(ctx context.Context,c client,directory string)error{
 repos,err:=c.collect(ctx,"repos")
 if err!=nil{return fmt.Errorf("repository snapshot: %w",err)}
 starred,err:=c.collect(ctx,"starred")
 if err!=nil{return fmt.Errorf("starred snapshot: %w",err)}
 return writeSnapshots(directory,repos,starred)
}
func main(){
 baseURL:=flag.String("api-url","https://api.github.com","GitHub API base URL (tests)")
 dir:=flag.String("data-dir","data","Hugo data directory")
 delay:=flag.Duration("page-delay",2*time.Second,"minimum pause between paginated API requests")
 maxPages:=flag.Int("max-pages",50,"maximum pages per catalogue")
 flag.Parse()
 if *maxPages<1||*delay<0{fmt.Fprintln(os.Stderr,"invalid pagination settings");os.Exit(2)}
 token:=os.Getenv("GH_TOKEN");if token==""{token=os.Getenv("GITHUB_TOKEN")}
 c:=client{http:&http.Client{Timeout:20*time.Second},baseURL:*baseURL,token:token,pause:*delay,sleep:time.Sleep,maxPages:*maxPages}
 ctx,cancel:=context.WithTimeout(context.Background(),8*time.Minute);defer cancel()
 if err:=run(ctx,c,*dir);err!=nil{fmt.Fprintf(os.Stderr,"::error::%v\n",err);os.Exit(1)}
 fmt.Println("Validated and refreshed complete public repository, starred and star-count snapshots")
}
