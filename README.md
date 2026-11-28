# arran4.github.io

This repository contains the source of my personal website hosted on GitHub Pages. It is built with the [Hugo](https://gohugo.io/) static site generator and uses the Toha theme.

## Building locally

You need the extended version of Hugo, Go, and Node/npm installed. Run the following commands from the project root:

```bash
hugo mod npm pack && npm install && hugo --minify
```

`hugo mod npm pack` vendors the theme's JavaScript dependencies, `npm install` installs them, and `hugo --minify` generates the static site into the `public/` directory. You can preview the site by opening `public/index.html` or by running `hugo server`.

## GitHub Pages workflow and metadata refresh

The existing `.github/workflows/gh-pages.yml` handles both publishing and the daily refresh; there is no separate low-value verification workflow.

- **Push to `main`:** test the Go refresh utility, build and verify Hugo, then deploy Pages.
- **Pull requests:** test and build but do not deploy.
- **Daily at 00:05 UTC / 11:05 AEDT:** run the Go refresher. Authenticate public API queries with the workflow's short-lived `GITHUB_TOKEN`, rate-limit requests with 2-second pagination pauses, respect bounded `Retry-After` waits, and fail without publishing partial data.
- **Changed metadata:** build and verify Hugo with the validated snapshots, commit only the three generated data files directly to `main` using a normal non-force push, then deploy the verified artifact in the same workflow. No changes means no commit or rebuild. Failed retrieval, validation, build or push prevents deployment.
- **Manual refresh:** run the existing Pages workflow on `main` with `workflow_dispatch` and `refresh=true`. Use `refresh=false` to rebuild and deploy the committed data (including recovery after a deployment failure).

The refresh uses read-only repository access. Only the separate snapshot commit job requests `contents: write`; only deployment requests `pages: write` and `id-token: write`. No automated PR creation/approval permission, workflow dispatch permission, or PAT is needed. Hugo builds do not receive the API token. The repository must allow the workflow token to push the generated files to `main`; a rejected push fails safely without force-pushing or deploying.

The Go command is `go run ./cmd/refresh-github-snapshots`. It writes `data/github/repositories.json` (all public user-owned repos), `data/github/starred.json` (public starred repos), and `data/stars.json` (star counts for public owned repos). Snapshots are sorted, retain stable GitHub `node_id` URLs and topic metadata, reject private or duplicated records, and refuse unexpected >10% reductions without review. It collects and validates *both* paginated lists before writing anything. Tests use local HTTP fixtures and run without GitHub access.

The initial complete snapshots must be generated, built, verified and committed directly by the first successful refresh run. Until then the Hugo adapters retain their temporary, legacy public API/curated fallback. Once both snapshots land in `main`, the website uses only the local snapshot files, never querying the GitHub API during Hugo generation. A follow-up cleanup should remove this transitional fallback.

The published site is https://arran4.github.io/.

## License

This content is licensed under the [Creative Commons Attribution 4.0 International License](LICENSE).
## Published sites and custom homepage

The home page is a custom, accessible visual project index backed by a curated `data/pages.yaml` (Pages-enabled public repositories) and `data/apps.yaml` (the site's eight existing public app landing pages, regardless of whether their individual source repositories are public).
This manually reviewed static file lists **all 33 public repositories found with Pages enabled on 2026-11-17**, including this website. Entries include a repository, category, description, featured flag and URL. Five Pages-enabled repositories have no confirmed published URL; they remain visible in the inventory with a source link but are not assigned invented web addresses.

Update `data/pages.yaml` through a normal code review whenever a new Pages site is launched, renamed, archived, or removed. This snapshot **does not refresh itself**. Featured projects are deliberately selected using `featured` and `featured_order`: Blog, Gentoo Packages, Header and go-subcommand. KLookup, the overlay and other work remain in the full directory without special prominence. Update `data/apps.yaml` when an on-site app landing page is added or changed. **No Pages API calls occur at build time.** Curated direct links also appear in the existing theme navigation so a user can recover after a broken relative link. The comprehensive `/pages/` directory is generated for reference but intentionally not added to the primary menu.

The `/repos/` catalogue and `/stars/` catalogue use their complete checked-in snapshots when available. The full repository browser is distinct from the deliberately curated 33-entry public Pages directory and must not silently collapse into that narrower catalogue. Repositories retain the original `node_id` paths, dates, tags, languages, archived flags and public source links. The Pages directory continues to be maintained manually via `data/pages.yaml`.

### Design and verification

The bespoke homepage and its styles are kept in `layouts/home.html`, `layouts/baseof.home.html` and `static/css/workbench.css`. Legacy app/privacy pages and the old Toha-backed repository browser are retained to preserve their paths.

To check this change:
1. Review every public Pages-enabled repository against `data/pages.yaml` and resolve the five missing site URLs where possible.
2. Run the Hugo build (`hugo mod npm pack && npm install && hugo --minify`), then inspect `public/index.html`, `public/pages/index.html`, `public/404.html` and the generated `/repos/` entries.
3. Verify all links, keyboard navigation, colour contrast, mobile layout and the light/dark presentations before merging.
