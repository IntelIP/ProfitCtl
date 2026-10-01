# ProfitCtl Website

Standalone landing page app for ProfitCtl.

## Env

- `PUBLIC_REPO_URL` for GitHub button links
- `PUBLIC_SITE_ORIGIN` for canonical metadata, the sitemap, and crawler guidance.
  The proposed production origin defaults to `https://profitctl.intelip.co`.
  Set it to the actual public origin before deploying elsewhere.

## Run

```bash
npm install
npm run dev
```

## Build

```bash
npm run build
```

The build produces static files in `dist/`, including `robots.txt` and
`sitemap.xml`. Local source and a successful build do not establish publication
or search indexing. Deployment and domain changes require maintainer approval.
