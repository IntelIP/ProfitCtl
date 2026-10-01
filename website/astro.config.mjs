import { defineConfig } from "astro/config";

export default defineConfig({
  site: process.env.PUBLIC_SITE_ORIGIN || "https://profitctl.intelip.co",
  trailingSlash: "always",
  compressHTML: true,
  devToolbar: {
    enabled: false,
  },
});
