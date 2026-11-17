#!/usr/bin/env bash
set -euo pipefail

# Validate output rather than merely checking that Hugo returned zero.
for page in public/index.html public/pages/index.html public/404.html public/repos/index.html; do
  test -s "$page" || { echo "Missing generated output: $page" >&2; exit 1; }
done

grep -Fq 'wb-feature-grid' public/index.html || { echo "Custom homepage was not rendered" >&2; exit 1; }
grep -Fq 'Published sites' public/index.html || { echo "Published-site shortcuts missing" >&2; exit 1; }
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
echo "Custom homepage, Pages directory, recovery links, and public-only catalogue verified."
