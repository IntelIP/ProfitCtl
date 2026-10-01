import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";

const output = new URL("../dist/", import.meta.url);
const [html, robots, sitemap] = await Promise.all(
  ["index.html", "robots.txt", "sitemap.xml"].map((file) =>
    readFile(new URL(file, output), "utf8"),
  ),
);
const origin = process.env.PUBLIC_SITE_ORIGIN || "https://profitctl.intelip.co";
const canonical = new URL("./", origin).href;
assert.ok(html.includes(`rel="canonical" href="${canonical}"`));
assert.ok(html.includes(`property="og:url" content="${canonical}"`));
assert.ok(html.includes('name="robots" content="index, follow"'));
assert.ok(robots.includes("User-agent: *\nAllow: /\n"));
assert.ok(robots.includes(`Sitemap: ${new URL("sitemap.xml", origin)}`));
assert.ok(sitemap.includes(`<loc>${canonical}</loc>`));
const schema = JSON.parse(
  html.match(/<script[^>]*type="application\/ld\+json"[^>]*>(.*?)<\/script>/s)[1],
);
assert.equal(schema["@type"], "SoftwareApplication");
assert.equal(schema.url, canonical);
console.log(`Search metadata and crawler files verified for ${canonical}`);
