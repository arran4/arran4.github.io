# arran4.github.io

This repository contains the source of my personal website hosted on GitHub Pages. It is built with the [Hugo](https://gohugo.io/) static site generator and uses the Toha theme.

## Building locally

You need the extended version of Hugo, Go, and Node/npm installed. Run the following commands from the project root:

```bash
hugo mod npm pack && npm install && hugo --minify
```

`hugo mod npm pack` vendors the theme's JavaScript dependencies, `npm install` installs them, and `hugo --minify` generates the static site into the `public/` directory. You can preview the site by opening `public/index.html` or by running `hugo server`.

## GitHub Pages workflow

Deployment is automated via the workflow in `.github/workflows/gh-pages.yml`. It triggers on pushes to the `main` branch, pull requests, and on a nightly schedule. The workflow performs these steps:

1. Check out the repository with submodules and full history.
2. Set up Hugo using the specified version.
3. Run the same build commands as above to produce the `public/` directory.
4. Deploy the contents of `public/` to the `gh-pages` branch using `peaceiris/actions-gh-pages` when the workflow runs on `main`.

The published site is served from `https://arran4.github.io/`.

## License

This content is licensed under the [Creative Commons Attribution 4.0 International License](LICENSE).
## Published sites and custom homepage

The home page is a custom, accessible visual project index backed by a curated `data/pages.yaml` (Pages-enabled public repositories) and `data/apps.yaml` (the site's eight existing public app landing pages, regardless of whether their individual source repositories are public).
This manually reviewed static file lists **all 33 public repositories found with Pages enabled on 2026-11-17**, including this website. Entries include a repository, category, description, featured flag and URL. Five Pages-enabled repositories have no confirmed published URL; they remain visible in the inventory with a source link but are not assigned invented web addresses.

Update `data/pages.yaml` through a normal code review whenever a new Pages site is launched, renamed, archived, or removed. This snapshot **does not refresh itself**. Featured projects are deliberately selected using `featured` and `featured_order`: Blog, Gentoo Packages, Header and go-subcommand. KLookup, the overlay and other work remain in the full directory without special prominence. Update `data/apps.yaml` when an on-site app landing page is added or changed. **No Pages API calls occur at build time.** Curated direct links also appear in the existing theme navigation so a user can recover after a broken relative link. The comprehensive `/pages/` directory is generated for reference but intentionally not added to the primary menu.

The `/repos/` catalogue still queries the **unauthenticated public** `https://api.github.com/users/arran4/repos` endpoint when available. If any page fails to load (including HTTP 403/rate limits), the build logs a warning, discards the potentially incomplete API result and uses the 33 reviewed public repositories in `data/pages.yaml` instead. This intentionally narrows the fallback to known public Pages repositories rather than publishing an empty or partial catalogue; cached Pages descriptions and site links remain functional. The fallback metadata date refers to the 2026-11-17 Pages inventory review, not to the repository's last commit. Set `HUGO_REPOS_OFFLINE=1` when running Hugo to exercise this fallback without requesting the repository endpoint. The build verification script runs an additional forced-fallback build. It matches entries from `data/pages.yaml` to attach a direct site link where the published URL is known. No GitHub token is passed to those requests, and the workflow's single public star count request is also unauthenticated. The nightly GitHub Actions job rebuilds even if star counts are unchanged. The separate star-count snapshot refresh is unauthenticated and best-effort; if public requests are rate-limited, the last checked-in counts are retained instead of failing the build. The existing public starred-repository API is subject to unauthenticated rate limits: when it fails, its list stops fetching and Hugo logs a warning rather than aborting the entire site. It may be incomplete during those runs; this does not affect the curated Pages and apps indexes.

### Design and verification

The bespoke homepage and its styles are kept in `layouts/home.html`, `layouts/baseof.home.html` and `static/css/workbench.css`. Legacy app/privacy pages and the old Toha-backed repository browser are retained to preserve their paths.

To check this change:
1. Review every public Pages-enabled repository against `data/pages.yaml` and resolve the five missing site URLs where possible.
2. Run the Hugo build (`hugo mod npm pack && npm install && hugo --minify`), then inspect `public/index.html`, `public/pages/index.html`, `public/404.html` and the generated `/repos/` entries.
3. Verify all links, keyboard navigation, colour contrast, mobile layout and the light/dark presentations before merging.
