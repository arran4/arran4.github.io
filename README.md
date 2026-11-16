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
## GitHub Pages site catalogue

The [Pages directory](https://arran4.github.io/pages/) lists public sites published by repositories owned by `arran4`. It is generated at Hugo build time from the same paginated GitHub repository feed as the [repository catalogue](https://arran4.github.io/repos/).

For repositories reporting `has_pages: true`, Hugo fetches the Pages API's published `html_url`, which handles custom domains and does not confuse arbitrary repository homepage links with GitHub Pages. The published site is linked from the Pages directory, its repository card, and its detail page. Private Pages sites are not exposed.

The nightly GitHub Actions run always builds and deploys the site, even when star counts have not changed, so additions and URL changes are picked up without manual edits. Page discovery errors stop the build rather than silently publishing an incomplete directory. Pull requests build the site for validation without deploying.
