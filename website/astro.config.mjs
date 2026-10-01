import { defineConfig } from "astro/config";

export default defineConfig({
  site: process.env.PUBLIC_SITE_ORIGIN || "https://profitctl.intelip.co",
  compressHTML: true,
  devToolbar: {
    enabled: false,
  },
});
