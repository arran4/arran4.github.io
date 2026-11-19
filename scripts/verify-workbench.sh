#!/usr/bin/env bash
set -euo pipefail

# Validate output rather than merely checking that Hugo returned zero.
for page in public/index.html public/pages/index.html public/404.html public/repos/index.html; do
  test -s "$page" || { echo "Missing generated output: $page" >&2; exit 1; }
done

# Toha v4.13.0 publishes a read-only static/sitemap.xml from the module cache.
# The site-local override must be writable and replaced with Hugo's real sitemap.
test -s public/sitemap.xml || { echo "Missing generated sitemap" >&2; exit 1; }
test -w public/sitemap.xml || { echo "Generated sitemap is read-only" >&2; exit 1; }
if grep -Fq 'This writable project-local file overrides' public/sitemap.xml; then
  echo "Sitemap placeholder was not replaced by Hugo" >&2
  exit 1
fi
grep -Fq '<urlset' public/sitemap.xml || { echo "Expected XML sitemap missing" >&2; exit 1; }

grep -Fq 'wb-feature-grid' public/index.html || { echo "Custom homepage was not rendered" >&2; exit 1; }
grep -Fq 'Published sites' public/index.html || { echo "Published-site shortcuts missing" >&2; exit 1; }
grep -Fq 'Apps and project pages' public/index.html || { echo "App directory missing" >&2; exit 1; }
grep -Fq 'https://arran4.github.io/gentoo-packages/' public/index.html || { echo "Featured Gentoo Packages link missing" >&2; exit 1; }
for app in arrans-movie-night-scheduler aurelies-dinners cbz-viewer cook-times gizmo-card-creator lojban-messenger-dictionary protein-calculator rpg-gym-stats; do
  test -s "public/$app/index.html" || { echo "Missing public app landing page: $app" >&2; exit 1; }
  grep -Fq "/$app/" public/index.html || { echo "App shortcut missing: $app" >&2; exit 1; }
done
if grep -Eq 'github.com/arran4/(cbz_viewer|cards)(["/]|$)' public/index.html; then
  echo "Homepage must not link to private app source repositories" >&2
  exit 1
fi
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
  # Temporary compatibility check until the first daily snapshot PR has merged.
  offline_dir=$(mktemp -d)
  trap 'rm -rf "$offline_dir"' EXIT
  HUGO_REPOS_OFFLINE=1 hugo --minify --destination "$offline_dir" --quiet
  test -s "$offline_dir/repos/index.html" || { echo "Legacy repository fallback missing" >&2; exit 1; }
  grep -Fq 'Visit site' "$offline_dir/repos/index.html" || {
    echo "Legacy repository fallback has no published-site shortcuts" >&2
    exit 1
  }
fi

echo "Custom homepage, Pages directory, recovery links, and public-only catalogue verified."
