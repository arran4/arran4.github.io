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

The build produces static HTML and assets in `public/`. Go and Python are used
only by optional public GitHub metadata refresh tooling, not to render the
frontend. The GitHub Pages workflow builds PRs without deploying; a successful
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
  snapshot replacement is [PR #107](https://github.com/arran4/arran4.github.io/pull/107).
- The release-first homepage and the native Introducing / Getting started /
  What's new article workflow are being developed independently in
  [PR #115](https://github.com/arran4/arran4.github.io/pull/115).
  Only approved articles should appear publicly.

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

## Checks and migration

Run `bash scripts/verify-workbench.sh` **after** building. The checks validate
generated pages, local CSS, the 404, app/privacy links, public-only catalogue
rules and absence of Toha/NPM source dependencies. The workflow runs one full
Hugo build; it does not rerender thousands of pages merely for a duplicate
fallback check.

[PR #117](https://github.com/arran4/arran4.github.io/pull/117)
removes the external theme independently of PRs #107 and #115. All target
`main`; their workflow/configuration/template changes must be reconciled in
merge order, without reintroducing removed frontend dependencies. The public
metadata snapshot workflow is a separate concern from the UI theme.

## License

This content is licensed under the [Creative Commons Attribution 4.0
International License](LICENSE).
