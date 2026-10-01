import assert from "node:assert/strict";
import { readFile, stat } from "node:fs/promises";

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

const routes = ["/", "/docs/"];
const pages = new Map();
for (const route of routes) {
  const page = await readFile(new URL(route === "/" ? "index.html" : "docs/index.html", output), "utf8");
  const url = new URL(route, origin).href;
  pages.set(route, page);
  assert.ok(page.includes(`rel="canonical" href="${url}"`), `Missing canonical: ${route}`);
  assert.ok(page.includes(`property="og:url" content="${url}"`), `Missing social URL: ${route}`);
  assert.ok(page.includes(`property="og:image" content="${new URL("/images/profitctl-social.jpg", origin).href}"`));
  assert.ok(page.includes('name="twitter:card" content="summary_large_image"'));
  assert.equal([...page.matchAll(/<h1\b/g)].length, 1, `Expected one primary heading: ${route}`);
  assert.ok(page.match(/<title>([^<]+)<\/title>/)?.[1].length > 20);
  assert.ok(page.match(/name="description" content="([^"]+)"/)?.[1].length > 80);
  assert.ok(sitemap.includes(`<loc>${url}</loc>`), `Missing sitemap entry: ${route}`);
  const data = JSON.parse(page.match(/<script[^>]*type="application\/ld\+json"[^>]*>(.*?)<\/script>/s)[1]);
  assert.equal(data["@type"], route === "/" ? "SoftwareApplication" : "TechArticle");
  assert.equal(data.url, url);
}

for (const [route, page] of pages) {
  for (const match of page.matchAll(/<a\b[^>]*href="([^"]+)"/g)) {
    const link = new URL(match[1], new URL(route, origin));
    if (link.origin !== new URL(origin).origin) continue;
    const target = pages.get(link.pathname);
    assert.ok(target, `Unknown internal destination: ${match[1]}`);
    if (link.hash) assert.ok(target.includes(`id="${decodeURIComponent(link.hash.slice(1))}"`), `Missing linked section: ${match[1]}`);
  }
}
assert.ok(html.includes('id="demo"') && html.includes("Interactive example"));
assert.ok((await stat(new URL("images/profitctl-social.jpg", output))).size > 10000);
console.log("Both pages: metadata, social image, structured data, sitemap, and internal destinations verified.");
