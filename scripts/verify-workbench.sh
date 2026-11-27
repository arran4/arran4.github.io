#!/usr/bin/env bash
set -euo pipefail

# Validate output rather than merely checking that Hugo returned zero.
for page in public/index.html public/pages/index.html public/404.html public/repos/index.html; do
  test -s "$page" || { echo "Missing generated output: $page" >&2; exit 1; }
done

# Preserve the generated sitemap after removing the legacy theme.
test -s public/sitemap.xml || { echo "Missing generated sitemap" >&2; exit 1; }
grep -Fq '<urlset' public/sitemap.xml || { echo "Expected XML sitemap missing" >&2; exit 1; }

grep -Eq 'id="?releases"?([[:space:]>])' public/index.html || { echo "Releases-first section missing" >&2; exit 1; }
grep -Eq 'id="?articles"?([[:space:]>])' public/index.html || { echo "Articles section missing" >&2; exit 1; }
grep -Eq 'id="?resources"?([[:space:]>])' public/index.html || { echo "Resource collections missing" >&2; exit 1; }
grep -Eq 'id="?directory"?([[:space:]>])' public/index.html || { echo "Searchable directory missing" >&2; exit 1; }
grep -Fq 'data-catalogue-item' public/index.html || { echo "Public directory cards missing" >&2; exit 1; }
test -s public/css/workbench-v2.css && test -s public/js/home.js || { echo "Local homepage assets missing" >&2; exit 1; }
grep -Fq 'Latest releases' public/index.html || { echo "Release entries missing" >&2; exit 1; }
grep -Fq 'CBZ Viewer' public/index.html || { echo "App directory entry missing" >&2; exit 1; }
grep -Fq 'https://arran4.github.io/gentoo-packages/' public/index.html || { echo "Gentoo Packages directory shortcut missing" >&2; exit 1; }
for app in arrans-movie-night-scheduler aurelies-dinners cbz-viewer cook-times gizmo-card-creator lojban-messenger-dictionary protein-calculator rpg-gym-stats; do
  test -s "public/$app/index.html" || { echo "Missing public app landing page: $app" >&2; exit 1; }
  grep -Fq "/$app/" public/index.html || { echo "App shortcut missing: $app" >&2; exit 1; }
done
if grep -Eq 'github.com/arran4/(cbz_viewer|cards)(["/]|$)' public/index.html; then
  echo "Homepage must not link to private app source repositories" >&2
  exit 1
fi
test ! -e public/articles/introducing-arran4-github-io/index.html || { echo "Unapproved draft introduction was published" >&2; exit 1; }
grep -Fq 'arran4/md2png' public/index.html || { echo "Cached md2png release link missing" >&2; exit 1; }
grep -Fq 'Awesome Terminal Multiplexers' public/index.html || { echo "Curated resources missing" >&2; exit 1; }
grep -Fq 'https://arran4.github.io/klookup/' public/index.html || { echo "Expected direct link missing" >&2; exit 1; }
grep -Fq 'Public GitHub Pages sites' public/pages/index.html || { echo "Reference directory was not rendered" >&2; exit 1; }
grep -Fq 'That page isn' public/404.html || { echo "Recovery 404 page was not rendered" >&2; exit 1; }
grep -Rq 'Published site' public/repos/ || { echo "Repository cards have no published-site shortcuts" >&2; exit 1; }

if grep -Fq 'href="/pages/"' public/index.html; then
  echo "The reference-only /pages/ directory should not be in the homepage navigation" >&2
  exit 1
fi
if grep -R 'GITHUB_TOKEN' content/ >/dev/null; then
  echo "A content generator still references GITHUB_TOKEN" >&2
  exit 1
fi
# The refresh job intentionally authenticates its metadata requests.
# Reject credentials only in the Hugo *build* job, never in the refresh job.
build_job=$(awk '
  /^  build:$/ { inside = 1; next }
  inside && /^  [[:alnum:]_-]+:$/ { exit }
  inside { print }
' .github/workflows/gh-pages.yml)
test -n "$build_job" || { echo "Pages workflow has no build job" >&2; exit 1; }
if printf '%s\n' "$build_job" | grep -Eq 'secrets\.|GH_TOKEN[[:space:]]*:|GITHUB_TOKEN[[:space:]]*:'; then
  echo "Hugo build job must not receive an API credential" >&2
  exit 1
fi

repos=$(grep '^  - repo:' data/pages.yaml | sed -E 's/^[^"]*"([^"]+)".*/\1/')
test -n "$repos" || { echo "Static Pages inventory is empty" >&2; exit 1; }
if test "$(printf '%s\n' "$repos" | sort | uniq -d | wc -l)" -ne 0; then
  echo "Duplicate repository entry in static Pages catalogue" >&2
  exit 1
fi
if printf '%s\n' "$repos" | grep -Ev '^arran4/[A-Za-z0-9_.-]+$'; then
  echo "Pages catalogue contains non-arran4 repository entries" >&2
  exit 1
fi
# Once the complete snapshots are bootstrapped, ensure Hugo generated the
# complete repository/starred catalogues, not only the curated Pages subset.
if test -s data/github/repositories.json && test -s data/github/starred.json; then
  expected_repos=$(jq -e 'if type == "array" and length > 0 then length else error("invalid repositories snapshot") end' data/github/repositories.json)
  expected_stars=$(jq -e 'if type == "array" and length > 0 then length else error("invalid starred snapshot") end' data/github/starred.json)
  actual_repos=$(find public/repos -mindepth 2 -type f -name index.html | wc -l)
  actual_stars=$(find public/stars -mindepth 2 -type f -name index.html | wc -l)
  if test "$actual_repos" -lt "$expected_repos"; then
    echo "Only $actual_repos/$expected_repos repository pages rendered" >&2
    exit 1
  fi
  if test "$actual_stars" -lt "$expected_stars"; then
    echo "Only $actual_stars/$expected_stars starred pages rendered" >&2
    exit 1
  fi
  echo "Complete offline GitHub catalogues rendered from validated snapshots."
else
  # Before bootstrap, the already completed Hugo build is sufficient.
  # Do not rerun all ~14,000 pages solely to exercise a temporary fallback.
  grep -Fq 'go-subcommand' public/repos/index.html || {
    echo "Legacy repository catalogue missing reviewed projects" >&2
    exit 1
  }
fi

# These assertions verify the local layouts and absence of the old dependencies
# without running another expensive full Hugo render.
test -s public/css/site-pages.css || { echo "Owned page styles are missing" >&2; exit 1; }
test -s public/stars/index.html || { echo "Starred-repository landing page missing" >&2; exit 1; }
for app in arrans-movie-night-scheduler aurelies-dinners cbz-viewer cook-times gizmo-card-creator lojban-messenger-dictionary protein-calculator rpg-gym-stats; do
  test -s "public/$app/privacy/index.html" || { echo "App privacy URL missing: $app" >&2; exit 1; }
done
test -s public/index.xml || { echo "RSS output missing" >&2; exit 1; }
grep -Fq 'PUBLIC GITHUB' public/repos/index.html || { echo "Local repository template not rendered" >&2; exit 1; }
grep -Fq 'PUBLIC GITHUB' public/stars/index.html || { echo "Local stars template not rendered" >&2; exit 1; }
if grep -Eiq '<(link|script)[^>]*(href|src)=[^ >]*(bootstrap|fontawesome|node_modules|popper|hugo-toha)' public/repos/index.html; then
  echo "Repository browser still has theme dependencies" >&2
  exit 1
fi
test ! -e package.json && test ! -e package-lock.json || { echo "NPM manifests should be removed" >&2; exit 1; }
! grep -Fq 'github.com/hugo-toha/toha' go.mod || { echo "Toha module is still required" >&2; exit 1; }
! grep -Eq '(hugo mod npm|npm install|sass)' .github/workflows/gh-pages.yml || { echo "Theme tools still used by CI" >&2; exit 1; }


echo "Releases-first homepage, local templates, public snapshots, privacy pages and draft exclusion verified."
