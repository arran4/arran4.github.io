// Command refresh-releases refreshes a curated public GitHub release snapshot.
// No token, third-party Go dependencies, or hidden/private repository lookups.
package main

import (
  "bufio"
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "net/http"
  "os"
  "path/filepath"
  "regexp"
  "sort"
  "strings"
  "time"
)

const (
  cataloguePath = "data/releases.json"
  allowlistPath = "tools/refresh-releases/repos.txt"
  githubAPI = "https://api.github.com"
)

type release struct {
  Repo string `json:"repo"`
  Name string `json:"name"`
  Tag string `json:"tag"`
  Published string `json:"published"`
  Summary string `json:"summary"`
  URL string `json:"url"`
}
type snapshot struct {
  CheckedAt string `json:"checked_at"`
  Items []release `json:"items"`
}
type apiRelease struct {
  Tag string `json:"tag_name"`
  Published string `json:"published_at"`
  URL string `json:"html_url"`
  Draft bool `json:"draft"`
}
var validRepo = regexp.MustCompile(`^arran4/[A-Za-z0-9_.-]+$`)

func parseAllowlist(content string) ([]string,error) {
  scanner:=bufio.NewScanner(strings.NewReader(content))
  seen:=make(map[string]bool)
  var repos []string
  for scanner.Scan() {
    repo:=strings.TrimSpace(scanner.Text())
    if repo=="" || strings.HasPrefix(repo,"#") {continue}
    if !validRepo.MatchString(repo) {return nil,fmt.Errorf("invalid public repository: %q",repo)}
    if seen[repo] {return nil,fmt.Errorf("duplicate repository: %q",repo)}
    seen[repo]=true
    repos=append(repos,repo)
  }
  return repos,scanner.Err()
}
func load(path string) (snapshot,error) {
  var data snapshot
  raw,err:=os.ReadFile(path)
  if err!=nil {return data,err}
  if err:=json.Unmarshal(raw,&data);err!=nil{return data,err}
  if len(data.Items)==0{return data,errors.New("empty existing release snapshot")}
  return data,nil
}
func fetch(ctx context.Context, client *http.Client, base, repo string) (apiRelease,error) {
  var out apiRelease
  if !validRepo.MatchString(repo) {return out,fmt.Errorf("disallowed repository %q",repo)}
  url:=strings.TrimRight(base,"/")+"/repos/"+repo+"/releases/latest"
  req,err:=http.NewRequestWithContext(ctx,http.MethodGet,url,nil)
  if err!=nil{return out,err}
  req.Header.Set("User-Agent","arran4-github-pages-public-release-index")
  req.Header.Set("Accept","application/vnd.github+json")
  resp,err:=client.Do(req)
  if err!=nil{return out,err}
  defer resp.Body.Close()
  if resp.StatusCode!=http.StatusOK {
    return out,fmt.Errorf("public GitHub API HTTP %d",resp.StatusCode)
  }
  if err:=json.NewDecoder(io.LimitReader(resp.Body,1<<20)).Decode(&out);err!=nil{return out,err}
  if out.Draft || out.Tag=="" || out.Published=="" {return out,errors.New("missing published release data")}
  if _,err:=time.Parse(time.RFC3339,out.Published);err!=nil{return out,err}
  expected:="https://github.com/"+repo+"/releases/tag/"
  if !strings.HasPrefix(out.URL,expected) {return out,errors.New("release URL does not point to allowed public repository")}
  return out,nil
}
func refresh(ctx context.Context,client *http.Client,base string,repos []string,before snapshot,now time.Time) (snapshot,int) {
  existing:=make(map[string]release,len(before.Items))
  for _,r:=range before.Items {existing[r.Repo]=r}
  result:=snapshot{CheckedAt:before.CheckedAt}
  successful:=0
  for _,repo:=range repos {
    previous,ok:=existing[repo]
    fetched,err:=fetch(ctx,client,base,repo)
    if err!=nil {
      fmt.Fprintf(os.Stderr,"::warning::Release refresh for %s unavailable (%s); keeping last-known-good data\n",repo,err)
      if ok {result.Items=append(result.Items,previous)}
      continue
    }
    if !ok {
      previous=release{Repo:repo,Name:strings.TrimPrefix(repo,"arran4/"),Summary:"See the official GitHub release notes."}
    }
    previous.Tag=fetched.Tag
    previous.Published=fetched.Published
    previous.URL=fetched.URL
    result.Items=append(result.Items,previous)
    successful++
  }
  sort.SliceStable(result.Items,func(i,j int)bool{return result.Items[i].Published>result.Items[j].Published})
  if successful>0 {result.CheckedAt=now.UTC().Format("2006-01-02")}
  return result,successful
}
func persist(path string,value snapshot) error {
  raw,err:=json.MarshalIndent(value,"","  ")
  if err!=nil{return err}
  raw=append(raw,'\n')
  file,err:=os.CreateTemp(filepath.Dir(path),".releases-*.json")
  if err!=nil{return err}
  defer os.Remove(file.Name())
  if err:=file.Chmod(0644);err!=nil{file.Close();return err}
  if _,err:=file.Write(raw);err!=nil{file.Close();return err}
  if err:=file.Close();err!=nil{return err}
  return os.Rename(file.Name(),path)
}
func main(){
  raw,err:=os.ReadFile(allowlistPath)
  if err!=nil {fmt.Fprintln(os.Stderr,err);os.Exit(1)}
  repos,err:=parseAllowlist(string(raw))
  if err!=nil {fmt.Fprintln(os.Stderr,err);os.Exit(1)}
  current,err:=load(cataloguePath)
  if err!=nil {fmt.Fprintln(os.Stderr,err);os.Exit(1)}
  client:=&http.Client{Timeout:8*time.Second}
  ctx,cancel:=context.WithTimeout(context.Background(),90*time.Second)
  defer cancel()
  next,count:=refresh(ctx,client,githubAPI,repos,current,time.Now())
  if count==0 {
    fmt.Fprintln(os.Stderr,"::warning::No new public release data; preserving the existing snapshot")
    return
  }
  if err:=persist(cataloguePath,next);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
  fmt.Printf("Refreshed %d of %d public repositories (%d cached or unavailable)\n",count,len(repos),len(repos)-count)
}
