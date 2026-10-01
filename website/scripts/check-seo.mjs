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
  assert.equal(data.author.name, "Hudson Aikman");
  assert.equal(data.publisher.name, "Intel IP");
  assert.ok(page.includes('name="author" content="Intel IP and Hudson Aikman"'));
  assert.ok(page.includes('Developed by <a href="https://github.com/IntelIP">Intel IP</a> and Hudson Aikman.'));
  assert.ok(page.includes('>GitHub <span aria-hidden="true">↗</span></a>'));
  assert.ok(page.includes('rel="icon" type="image/png" sizes="96x96" href="/favicon-96.png"'), `Missing crawlable favicon: ${route}`);
  assert.ok(page.includes('rel="icon" type="image/svg+xml" sizes="any" href="/favicon.svg"'));
  assert.ok(page.includes('rel="shortcut icon" href="/favicon.ico"'));
  assert.ok(page.includes('rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png"'));
  assert.ok(page.includes('name="theme-color" content="#20251d"'));
  assert.ok(!page.includes('class="brand-mark"'), `Old letter badge: ${route}`);
  assert.ok(page.includes('aria-label="ProfitCTL home">ProfitCTL</a>'));
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
for (const [file, size] of [["favicon-96.png", 96], ["icons/profitctl-192.png", 192], ["apple-touch-icon.png", 180]]) {
  const icon = await readFile(new URL(file, output));
  assert.equal(icon.subarray(1, 4).toString(), "PNG", `Invalid icon: ${file}`);
  assert.equal(icon.readUInt32BE(16), size, `Wrong icon width: ${file}`);
  assert.equal(icon.readUInt32BE(20), size, `Wrong icon height: ${file}`);
}
const legacyIcon = await readFile(new URL("favicon.ico", output));
assert.equal(legacyIcon.readUInt16LE(2), 1);
assert.equal(legacyIcon.readUInt16LE(4), 4);
assert.ok((await readFile(new URL("favicon.svg", output), "utf8")).includes('viewBox="0 0 96 96"'));
assert.ok(schema.image.includes(new URL("/icons/profitctl-192.png", origin).href));
console.log("Both pages: metadata, brand icons, social image, structured data, sitemap, and internal destinations verified.");
