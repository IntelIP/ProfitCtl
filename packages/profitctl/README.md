# ProfitCtl for Bun

This package exposes the native ProfitCtl CLI through Bun without downloading binaries from GitHub at install time.

From the canonical repository:

```bash
cd packages/profitctl
bun run build:native
bun run profitctl -- --help
```

Release packaging bundles the supported macOS, Linux, and Windows binaries into one tarball. Public registry publication remains disabled while the canonical ProfitCtl repository is private.
