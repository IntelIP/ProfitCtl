import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { runInNewContext } from "node:vm";

const layout = readFileSync(new URL("../src/layouts/ProductLayout.astro", import.meta.url), "utf8");
const script = layout.match(/<script is:inline data-openpanel-init[^>]*>([\s\S]*?)<\/script>/)?.[1];
assert.ok(script, "Shared layout must contain the OpenPanel initializer");

function visit(hostname, navigator = {}, existingClient) {
  const loaded = [];
  const window = existingClient ? { op: existingClient } : {};
  const context = {
    window,
    location: { hostname },
    navigator,
    openPanelClientId: "public-test-client",
    document: {
      createElement: () => ({}),
      head: { appendChild: (element) => loaded.push(element) },
    },
  };
  runInNewContext(script, context, { timeout: 1000 });
  return { loaded, window, context };
}

for (const host of ["localhost", "127.0.0.1", "profitctl.pages.dev", "preview.profitctl.pages.dev", "profitctl.com.example.org"]) {
  const result = visit(host);
  assert.equal(result.loaded.length, 0, `Do not load analytics on ${host}`);
  assert.equal(result.window.op, undefined, `Do not queue events on ${host}`);
}

for (const signal of [{ doNotTrack: "1" }, { globalPrivacyControl: true }]) {
  const result = visit("profitctl.com", signal);
  assert.equal(result.loaded.length, 0, "Privacy signals must prevent the analytics request");
  assert.equal(result.window.op, undefined, "Privacy signals must prevent queued events");
}

for (const host of ["profitctl.com", "www.profitctl.com"]) {
  const result = visit(host);
  assert.equal(result.loaded.length, 1, `Load the provider once on ${host}`);
  assert.equal(result.loaded[0].src, "https://openpanel.dev/op1.js");
  assert.equal(result.loaded[0].async, true);
  const queue = JSON.parse(JSON.stringify(result.window.op.q));
  assert.deepEqual(queue, [["init", {
    clientId: "public-test-client",
    trackScreenViews: true,
    trackOutgoingLinks: true,
    trackAttributes: false,
  }]]);
  runInNewContext(script, result.context, { timeout: 1000 });
  assert.equal(result.loaded.length, 1, "Repeated initialization must not duplicate the provider");
}

const existing = () => {};
const result = visit("profitctl.com", {}, existing);
assert.equal(result.loaded.length, 0, "Preserve an existing OpenPanel client");
assert.equal(result.window.op, existing);
console.log("Analytics domain, privacy, public configuration, and duplicate-load checks passed.");
