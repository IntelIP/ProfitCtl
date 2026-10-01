# ProfitCtl Website

Static product website for ProfitCtl, with an interactive example and local
documentation at `/docs/`.

## Env

- `PUBLIC_REPO_URL` for GitHub button links
- `PUBLIC_SITE_ORIGIN` for canonical metadata, the sitemap, and crawler guidance.
  The proposed production origin defaults to `https://profitctl.intelip.co`.
  Set it to the actual public origin before deploying elsewhere.

## Run

```bash
bun install --frozen-lockfile
bun run dev
```

## Build

```bash
bun run build
bun run typecheck
```

The build produces static files in `dist/`, including `robots.txt` and
`sitemap.xml`. It checks the demo calculations, page metadata, social image,
headings, structured data, and local links. The browser demo uses illustrative
costs; it does not run the CLI simulation or fetch live provider prices.

The page uses the ProfitCTL wordmark. Browser icons share its forest and lime
colors and are served at stable public paths. Development credit appears on
both pages and in their structured metadata: IntelIP and Hudson Aikins.

Local source and a successful build do not establish publication
or search indexing. Deployment and domain changes require maintainer approval.
