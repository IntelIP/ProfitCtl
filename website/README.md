# ProfitCtl Website

Static product website for ProfitCTL, with a real CLI terminal demo, a separate
cost calculator, and local documentation at `/docs/`.

## Env

- `PUBLIC_REPO_URL` for GitHub button links
- `PUBLIC_SITE_ORIGIN` for canonical metadata, the sitemap, and crawler guidance.
  The production origin defaults to `https://profitctl.com`.
  Set it to the actual public origin before deploying elsewhere.
- Cloudflare Pages: set `GO_VERSION=1.26.8` to match `go.mod`.
- `PUBLIC_OPENPANEL_CLIENT_ID`: the public ID of the ProfitCtl **write** client.
  Set it in the Cloudflare Pages production build environment to enable analytics.
  Leave it unset for development and previews. Never use the MCP read client or
  put a client secret in a `PUBLIC_*` variable.

## Analytics

The shared layout loads the [official OpenPanel script](https://openpanel.dev/docs/sdks/script)
asynchronously on `profitctl.com` and `www.profitctl.com`. It records page views
and outgoing link clicks on the homepage and documentation. It does not identify
visitors by name or email, or enable custom attribute tracking.

The script stays unloaded on local and preview domains, when the public client
ID is absent, and when Do Not Track or Global Privacy Control is enabled.
Browser blockers can also prevent collection. The MCP credentials in Doppler
are for reading reports and are never included in the website.

Run `bun run test:analytics` for the domain, privacy, and duplicate-load checks.
These checks also run during the build. Local checks do not prove live event
receipt; verify a page view in ProfitCtl's OpenPanel project after an approved
production deployment.

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

The header and footer pair the ProfitCTL wordmark with the approved bold-plus
calculator mark. Browser icons and the sharing thumbnail use muted sage and
forest colors. Icon links and the sharing image are versioned to refresh caches. Development credit appears on
both pages and in their structured metadata: IntelIP and Hudson Aikins.

Local source and a successful build do not establish publication
or search indexing. Deployment and domain changes require maintainer approval.

The homepage and documentation follow the browser light or dark preference through
native CSS, including cards, controls, text, and browser theme colors.
