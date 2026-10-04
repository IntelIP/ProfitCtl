import { defineConfig } from "astro/config";

export default defineConfig({
  site: process.env.PUBLIC_SITE_ORIGIN || "https://profitctl.com",
  trailingSlash: "always",
  compressHTML: true,
  devToolbar: {
    enabled: false,
  },
});
