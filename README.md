# arran4.github.io

Source for [arran4.github.io](https://arran4.github.io/), a Hugo-based project
hub with public software links, technical writing, reference collections and
application landing pages.

The site's frontend is owned by this repository. **There is no Toha theme, NPM
package tree, Bootstrap, Font Awesome, Sass build step or external JavaScript
framework.** Existing page routes are retained.

## Build and preview

Use Hugo **v0.170.1**:

```sh
hugo --minify
hugo server
```

The build produces static HTML and assets in `public/`. Go is used by the public GitHub metadata refresh tooling, not to render the frontend. The GitHub Pages workflow builds PRs without deploying; a successful
build on `main` publishes the static site.

## Content sources

- `data/pages.yaml`: a **manually reviewed snapshot** of public
  GitHub Pages sites, including a separate `/pages/` reference directory.
  Unverified URLs are never invented. Changes require a normal PR.
- `data/apps.yaml`: the eight public app landing pages hosted by this
  website. A public landing page does not imply that the application's source
  repository is public.
- `/repos/` and `/stars/`: public GitHub repository/starred-repository
  indexes. Their generated detail URLs use GitHub's stable IDs when available.
  The existing unauthenticated API builder may be rate-limited; `/repos/`
  falls back to the reviewed Pages inventory, while `/stars/` remains
  available with a clear empty state rather than a 404. The reviewed, durable
  snapshot replacement was [PR #107](https://github.com/arran4/arran4.github.io/pull/107).
- `data/releases.json`: selected public GitHub release metadata from the
  allowlist in `tools/refresh-releases/repos.txt`. The same daily Go snapshot
  job refreshes this cache, retains prior entries on API failures, and proposes
  changes in the existing data-only PR. No release metadata changes means
  no timestamp-only churn.
- `data/articles.yaml`: reviewed, already published technical writing.
  Native Introducing / Getting started / What's new articles require editorial
  approval; drafts do not appear in the deployed site.
- `data/home.yaml`: optional support destinations, disabled by default.
  Use `?preview-support=1` to review the panel.

## First-party layouts

- `layouts/home.html`, `layouts/baseof.home.html`,
  `static/css/workbench.css`: custom homepage.
- `layouts/_default/`, `layouts/repos/`, `layouts/stars/`,
  `layouts/*-tags/`, `layouts/pages/`, and
  `layouts/partials/catalogue/`: all other shared views and public
  repository details, with local styles in `static/css/site-pages.css`.
- `layouts/<app>/single.html` and `static/<app>/style.css`:
  existing standalone application pages and their privacy URLs.
- `layouts/index.rss.xml`: owned-site RSS, without obsolete sample
  GoHugo.org news items.

The site does not fetch fonts, stylesheets or scripts from a third-party UI
framework at runtime. Normal external links to GitHub, project documentation
and other public destinations remain.

## GitHub metadata and verification

The merged metadata pipeline runs once daily or with manual `refresh=true`.
It tests and runs `go run ./cmd/refresh-github-snapshots`, validates public
repository and starred records, refreshes allowlisted public releases, paces requests, and proposes a single
data-only PR when metadata changes. It uses the workflow token only in the
refresh job. Hugo builds do not receive it. Once both complete snapshots
are committed, builds use local metadata; until then the transitional API
fallback remains. See #108 for bootstrap and fallback removal.

After building, run `bash scripts/verify-workbench.sh`. It verifies generated
pages, public catalogue completeness when snapshots exist, local CSS,
app/privacy routes, RSS and the absence of external theme build dependencies.
Only one full Hugo render is required.

## License

This content is licensed under the [Creative Commons Attribution 4.0
International License](LICENSE).
