# arran4.github.io

This repository contains the source of my personal website hosted on GitHub Pages. It is built with the [Hugo](https://gohugo.io/) static site generator and uses the Toha theme.

## Building locally

You need the extended version of Hugo and Node/npm installed. Run the following commands from the project root:

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
