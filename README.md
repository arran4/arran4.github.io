# arran4.github.io

This repository contains the source of my personal website hosted on GitHub Pages. It is built with [Hugo](https://gohugo.io/). The homepage has repository-owned templates, CSS and JavaScript; legacy subpages still rely on the Toha theme until their separate migration.

## Building locally

You need Hugo **v0.170.1**, Go and (for the existing Toha-based subpages) Node/npm **plus Dart Sass**. The Extended edition was removed in v0.170.0; Toha's old Sass assets now require a separately installed `sass` binary that supports the embedded protocol. CI pins the official Dart Sass **1.108.2** Linux archive by SHA-256. This dependency is temporary and is not needed by the homepage itself. Run the following commands from the project root:

```bash
hugo mod npm pack && npm install && hugo --minify
```

`hugo mod npm pack` vendors the theme's JavaScript dependencies, `npm install` installs them, and `hugo --minify` generates the static site into the `public/` directory. You can preview the site by opening `public/index.html` or by running `hugo server`.

## GitHub Pages workflow

Deployment is automated via the workflow in `.github/workflows/gh-pages.yml`. It triggers on pushes to the `main` branch, pull requests, and on a nightly schedule. The workflow performs these steps:

1. Check out the repository with submodules and full history.
2. Set up Hugo using the specified version.
3. Run the same build commands as above to produce the `public/` directory.
4. Upload and deploy the generated `public/` artifact using the GitHub Pages Actions integration when the workflow runs on `main`.

The published site is served from `https://arran4.github.io/`.

## License

This content is licensed under the [Creative Commons Attribution 4.0 International License](LICENSE).
## Homepage v2: releases-first, articles and public resources

The homepage is a **releases-first, repository-owned Hugo view**: `layouts/home.html`, `layouts/baseof.home.html`, `static/css/workbench-v2.css` and `static/js/home.js`. No external client-side UI dependencies or live browser API calls are used. System/dark/light switching and keyboard-accessible project search/filtering run entirely in vanilla JavaScript.

- `tools/refresh-releases/repos.txt` lists ten approved **public** GitHub release sources. `data/releases.json` is a checked-in snapshot containing actual version tags and release timestamps. On scheduled/explicit workflow runs the standard-library-only `go run ./tools/refresh-releases` updater checks the public unauthenticated GitHub Releases API with a timeout and keeps last-known-good entries on HTTP 403, network failure or missing releases. Offline HTTP fixture tests are in `tools/refresh-releases`. Snapshot timestamps are visible; this is not a claim of real-time availability.
- `data/articles.yaml` lists already-published technical blog articles for the homepage. The first **Introducing arran4.github.io** story is an unpublished `draft: true` under `content/articles/introducing-arran4-github-io/`. Its approval and publication are coordinated via [public-relations #78](https://github.com/arran4/public-relations/issues/78) and the [website editorial-pipeline issue](https://github.com/arran4/arran4.github.io/issues/113).
- Curated Awesome/reference collections have a separate section. The compact combined public Pages and app directory shows twelve items initially, then supports category filtering and search. A site URL that cannot be verified falls back to its public GitHub repository URL.
- `data/home.yaml` disables support/publication links by default. `?preview-support=1` displays a review-only panel. No unverified Substack URL is published.
- The [owner-assigned photography issue](https://github.com/arran4/arran4.github.io/issues/111) is deferred; significant [upstream contributions](https://github.com/arran4/arran4.github.io/issues/112) are not inferred solely from forks.
- The homepage change **does not remove Toha, its Hugo module, or NPM** (or the Dart Sass build binary needed by the legacy theme on Hugo 0.170+): `/repos/`, `/stars/`, other applications, privacy/404 and older generated pages continue using those assets until [the full legacy parity migration](https://github.com/arran4/arran4.github.io/issues/110) is complete.

This website is the intended future first-party location for owner-reviewed Introducing, Getting started and What's new articles, as recorded in [public-relations](https://github.com/arran4/public-relations/issues/77). Release tags and publication approval remain distinct.

## Published sites and custom homepage

The home page is a custom, accessible visual project index backed by a curated `data/pages.yaml` (Pages-enabled public repositories) and `data/apps.yaml` (the site's eight existing public app landing pages, regardless of whether their individual source repositories are public).
This manually reviewed static file lists **all 33 public repositories found with Pages enabled on 2026-11-17**, including this website. Entries include a repository, category, description, featured flag and URL. Five Pages-enabled repositories have no confirmed published URL; they remain visible in the inventory with a source link but are not assigned invented web addresses.

Update `data/pages.yaml` through a normal code review whenever a new Pages site is launched, renamed, archived, or removed. This snapshot **does not refresh itself**. The historical `featured` and `featured_order` metadata remain in the Pages catalogue for compatibility, but **homepage feature status now comes from editorial article selection**, not menu location or repository prominence. Update `data/apps.yaml` when an on-site app landing page is added or changed. **No Pages API calls occur at build time.** Curated direct links also appear in the existing theme navigation so a user can recover after a broken relative link. The comprehensive `/pages/` directory is generated for reference but intentionally not added to the primary menu.

The `/repos/` catalogue still queries the **unauthenticated public** `https://api.github.com/users/arran4/repos` endpoint when available. If any page fails to load (including HTTP 403/rate limits), the build logs a warning, discards the potentially incomplete API result and uses the 33 reviewed public repositories in `data/pages.yaml` instead. This intentionally narrows the fallback to known public Pages repositories rather than publishing an empty or partial catalogue; cached Pages descriptions and site links remain functional. The fallback metadata date refers to the 2026-11-17 Pages inventory review, not to the repository's last commit. Set `HUGO_REPOS_OFFLINE=1` when running Hugo to exercise this fallback without requesting the repository endpoint. The homepage verification checks the already rendered output; the prior extra forced-fallback Hugo rebuild has been removed because it rerendered thousands of pages and duplicate Go snapshot tests provide better coverage. It matches entries from `data/pages.yaml` to attach a direct site link where the published URL is known. No GitHub token is passed to those requests, and the workflow's single public star count request is also unauthenticated. The nightly GitHub Actions job rebuilds even if star counts are unchanged. The separate star-count snapshot refresh is unauthenticated and best-effort; if public requests are rate-limited, the last checked-in counts are retained instead of failing the build. The existing public starred-repository API is subject to unauthenticated rate limits: when it fails, its list stops fetching and Hugo logs a warning rather than aborting the entire site. It may be incomplete during those runs; this does not affect the curated Pages and apps indexes.

### Design and verification

The bespoke homepage and its styles are kept in `layouts/home.html`, `layouts/baseof.home.html` and `static/css/workbench.css`. Legacy app/privacy pages and the old Toha-backed repository browser are retained to preserve their paths.

To check the existing Pages directory:
1. Review every public Pages-enabled repository against `data/pages.yaml` and resolve the five missing site URLs where possible.
2. Run the Hugo build (`hugo mod npm pack && npm install && hugo --minify`), then inspect `public/index.html`, `public/pages/index.html`, `public/404.html` and the generated `/repos/` entries.
3. Verify all links, keyboard navigation, colour contrast, mobile layout and the light/dark presentations before merging.
