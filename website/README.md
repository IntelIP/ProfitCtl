# ProfitCtl Website

Static product website for ProfitCTL, with a real CLI terminal demo, a separate
cost calculator, and local documentation at `/docs/`.

## Env

- `PUBLIC_REPO_URL` for GitHub button links
- `PUBLIC_SITE_ORIGIN` for canonical metadata, the sitemap, and crawler guidance.
  The production origin defaults to `https://profitctl.com`.
  Set it to the actual public origin before deploying elsewhere.
- Cloudflare Pages: set `GO_VERSION=1.26.8` to match `go.mod`.

## Run

Install the Go version specified by `go.mod` and Bun 1.3.3. Both development and
production builds compile the current CLI for the browser.
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
headings, structured data, and local links. The terminal runs the actual CLI as
WebAssembly in a worker, against the fictional files in `examples/website_demo/`.
It validates a baseline, simulates rising usage, and compares a cheaper route.
The story loops automatically. Selecting a step holds its result. Reduced-motion
settings show a single comparison without looping. Playback pauses when the demo is off screen or
the page is hidden. No provider calls or local-file access are enabled.

The build generates compressed WebAssembly, its matching Go runtime and
license, and scenario downloads in `public/demo/`. These assets are ignored by
Git and rebuilt for each deployment. The loader accepts either an already
decoded HTTP response or a compressed file. The real WebAssembly commands and
read-only workspace are checked by `scripts/check-terminal-demo.mjs`.

The separate calculator remains a simplified illustration. Neither it nor the
terminal scenario uses live provider prices.

The page uses the ProfitCTL wordmark. Browser icons share its forest and lime
colors and are served at stable public paths. Development credit appears on
both pages and in their structured metadata: IntelIP and Hudson Aikins.

Local source and a successful build do not establish publication
or search indexing. Deployment and domain changes require maintainer approval.

The homepage and documentation follow the browser light or dark preference through
native CSS, including cards, controls, text, and browser theme colors.
