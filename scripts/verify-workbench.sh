#!/usr/bin/env bash
set -euo pipefail

# Validate output rather than merely checking that Hugo returned zero.
for page in public/index.html public/pages/index.html public/404.html public/repos/index.html; do
  test -s "$page" || { echo "Missing generated output: $page" >&2; exit 1; }
done

grep -Fq 'id="releases"' public/index.html || { echo "Releases-first section missing" >&2; exit 1; }
grep -Fq 'id="articles"' public/index.html || { echo "Articles section missing" >&2; exit 1; }
grep -Fq 'id="resources"' public/index.html || { echo "Resource collections missing" >&2; exit 1; }
grep -Fq 'id="directory"' public/index.html || { echo "Searchable directory missing" >&2; exit 1; }
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
grep -Fq 'v0.5.22' public/index.html || { echo "Cached md2png release link missing" >&2; exit 1; }
grep -Fq 'Awesome Terminal Multiplexers' public/index.html || { echo "Curated resources missing" >&2; exit 1; }
grep -Fq 'https://arran4.github.io/klookup/' public/index.html || { echo "Expected direct link missing" >&2; exit 1; }
grep -Fq 'Public GitHub Pages sites' public/pages/index.html || { echo "Reference directory was not rendered" >&2; exit 1; }
grep -Fq 'That page isn' public/404.html || { echo "Recovery 404 page was not rendered" >&2; exit 1; }
grep -Rq 'Visit site' public/repos/ || { echo "Repository cards have no published-site shortcuts" >&2; exit 1; }

if grep -Fq 'href="/pages/"' public/index.html; then
  echo "The reference-only /pages/ directory should not be in the homepage navigation" >&2
  exit 1
fi
if grep -R 'GITHUB_TOKEN' content/ >/dev/null; then
  echo "A content generator still references GITHUB_TOKEN" >&2
  exit 1
fi
if grep -Fq 'secrets.GITHUB_TOKEN' .github/workflows/gh-pages.yml; then
  echo "Build still passes the repository token to a public data request" >&2
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
# Exercise the repository catalogue without any successful repository API lookup.
# A rate-limited deployment must still publish a useful, public-only /repos/ page.
offline_dir=$(mktemp -d)
trap 'rm -rf "$offline_dir"' EXIT
HUGO_REPOS_OFFLINE=1 hugo --minify --destination "$offline_dir" --quiet
test -s "$offline_dir/repos/index.html" || { echo "Offline repository catalogue missing" >&2; exit 1; }
grep -Fq 'Visit site' "$offline_dir/repos/index.html" || {
  echo "Offline repository catalogue has no published-site shortcuts" >&2
  exit 1
}
grep -Fq 'go-subcommand' "$offline_dir/repos/index.html" || {
  echo "Offline repository catalogue is missing reviewed projects" >&2
  exit 1
}

echo "Releases-first homepage, cached public releases, article draft gate, applications, Pages directory, and public-only catalogue verified."
