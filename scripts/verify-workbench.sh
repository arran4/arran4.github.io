#!/usr/bin/env bash
set -euo pipefail

# Validate output rather than merely checking that Hugo returned zero.
for page in public/index.html public/pages/index.html public/404.html public/repos/index.html; do
  test -s "$page" || { echo "Missing generated output: $page" >&2; exit 1; }
done

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
grep -Rq 'Published site' public/repos/ || { echo "Repository cards have no published-site shortcuts" >&2; exit 1; }

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
# These assertions verify the local layouts and absence of the old dependencies
# without running another expensive full Hugo render.
test -s public/css/site-pages.css || { echo "Owned page styles are missing" >&2; exit 1; }
test -s public/stars/index.html || { echo "Starred-repository landing page missing" >&2; exit 1; }
test -s public/cbz-viewer/privacy/index.html || { echo "App privacy URL missing" >&2; exit 1; }
test -s public/index.xml || { echo "RSS output missing" >&2; exit 1; }
grep -Fq 'PUBLIC GITHUB' public/repos/index.html || { echo "Local repository template not rendered" >&2; exit 1; }
grep -Fq 'PUBLIC GITHUB' public/stars/index.html || { echo "Local stars template not rendered" >&2; exit 1; }
if grep -Eiq 'hugo-toha|bootstrap|fontawesome|node_modules|popper' public/repos/index.html; then
  echo "Repository browser still has theme dependencies" >&2
  exit 1
fi
test ! -e package.json && test ! -e package-lock.json || { echo "NPM manifests should be removed" >&2; exit 1; }
! grep -Fq 'github.com/hugo-toha/toha' go.mod || { echo "Toha module is still required" >&2; exit 1; }
! grep -Eq '(hugo mod npm|npm install|sass)' .github/workflows/gh-pages.yml || { echo "Theme tools still used by CI" >&2; exit 1; }

echo "Custom homepage, Pages directory, recovery links, and public-only catalogue verified."
