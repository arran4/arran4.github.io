package main

import (
  "context"
  "fmt"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"
  "time"
)
func TestParseAllowlistRejectsOutsideOwnerAndDuplicates(t *testing.T) {
  good,err:=parseAllowlist("# comment\narran4/header\n\narran4/md2png\n")
  if err!=nil || len(good)!=2 {t.Fatalf("valid allowlist: %v %v",good,err)}
  for _,input:=range []string{"someoneelse/repo","arran4/private/extra","arran4/header\narran4/header"} {
    if _,err:=parseAllowlist(input);err==nil {t.Errorf("accepted invalid input %q",input)}
  }
}
func TestRefreshRetainsLastKnownGoodOnRateLimit(t *testing.T) {
  server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
    if strings.Contains(r.URL.Path,"/header/") {
      w.Header().Set("Content-Type","application/json")
      fmt.Fprint(w, `{"tag_name":"v0.0.7","published_at":"2026-11-19T01:00:00Z","html_url":"https://github.com/arran4/header/releases/tag/v0.0.7"}`)
      return
    }
    w.WriteHeader(http.StatusForbidden)
  }))
  defer server.Close()
  before:=snapshot{CheckedAt:"2026-11-18",Items:[]release{
    {Repo:"arran4/header",Name:"Header",Tag:"v0.0.6",Published:"2026-11-10T00:00:00Z",Summary:"Metadata CLI"},
    {Repo:"arran4/md2png",Name:"md2png",Tag:"v0.5.22",Published:"2026-11-11T00:00:00Z",Summary:"Markdown rendering"},
  }}
  now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
  got,count:=refresh(context.Background(),server.Client(),server.URL,[]string{"arran4/md2png","arran4/header"},before,now)
  if count!=1 || got.CheckedAt!="2026-11-19" || len(got.Items)!=2 {t.Fatalf("unexpected refresh: %d %#v",count,got)}
  if got.Items[0].Repo!="arran4/header" || got.Items[0].Tag!="v0.0.7" {t.Fatalf("new release did not sort first: %#v",got.Items)}
  if got.Items[1].Tag!="v0.5.22" || got.Items[1].Summary!="Markdown rendering" {t.Fatal("cached entry was lost")}
}
func TestRefreshAllRateLimitedKeepsTimestamp(t *testing.T) {
  server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(http.StatusForbidden)}))
  defer server.Close()
  before:=snapshot{CheckedAt:"2026-11-18",Items:[]release{{Repo:"arran4/header",Name:"Header",Tag:"v0.0.6",Published:"2026-11-10T00:00:00Z"}}}
  got,count:=refresh(context.Background(),server.Client(),server.URL,[]string{"arran4/header"},before,time.Now())
  if count!=0 || got.CheckedAt!=before.CheckedAt || len(got.Items)!=1 || got.Items[0].Tag!="v0.0.6" {t.Fatalf("lost cache: %#v %d",got,count)}
}
func TestFetchRejectsIncorrectReleaseLink(t *testing.T) {
  server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
    fmt.Fprint(w,`{"tag_name":"v2","published_at":"2026-11-19T01:00:00Z","html_url":"https://example.com/something"}`)
  }))
  defer server.Close()
  if _,err:=fetch(context.Background(),server.Client(),server.URL,"arran4/header");err==nil {t.Fatal("accepted wrong release URL")}
}

func TestUnchangedReleaseKeepsSnapshotDate(t *testing.T) {
  server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
    fmt.Fprint(w, `{"tag_name":"v1","published_at":"2026-11-19T01:00:00Z","html_url":"https://github.com/arran4/header/releases/tag/v1"}`)
  }))
  defer server.Close()
  before:=snapshot{CheckedAt:"2026-11-19",Items:[]release{{Repo:"arran4/header",Name:"Header",Tag:"v1",Published:"2026-11-19T01:00:00Z",URL:"https://github.com/arran4/header/releases/tag/v1",Summary:"Existing summary"}}}
  got,count:=refresh(context.Background(),server.Client(),server.URL,[]string{"arran4/header"},before,time.Date(2026,11,27,0,0,0,0,time.UTC))
  if count!=1 || got.CheckedAt!=before.CheckedAt || got.Items[0]!=before.Items[0] {t.Fatalf("unchanged release caused snapshot churn: %#v",got)}
}
