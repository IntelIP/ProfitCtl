import assert from "node:assert/strict";
import { webcrypto } from "node:crypto";
import { readFile } from "node:fs/promises";
import { createContext, runInContext } from "node:vm";
import { gunzipSync } from "node:zlib";
import { createDemoFilesystem } from "../src/scripts/cli-filesystem.ts";
import { terminalStory } from "../src/scripts/terminal-story.ts";

const output = new URL("../public/demo/", import.meta.url);
const [compressed, runtimeSource, files] = await Promise.all([
  readFile(new URL("profitctl.wasm.gz", output)),
  readFile(new URL("wasm_exec.js", output), "utf8"),
  readFile(new URL("scenarios.json", output), "utf8").then(JSON.parse),
]);
assert.ok(compressed.length < 25 * 1024 * 1024);
const module = await WebAssembly.compile(gunzipSync(compressed));
async function run(args) {
  let transcript = "";
  let exitCode = 0;
  const context = createContext({
    fs: createDemoFilesystem(files, (text) => { transcript += text; }),
    process: { cwd: () => "/", getuid: () => -1, getgid: () => -1, geteuid: () => -1, getegid: () => -1, getgroups: () => [], pid: -1, ppid: -1, umask: () => 0 },
    crypto: webcrypto, performance, TextEncoder, TextDecoder, Uint8Array, ArrayBuffer, DataView, WebAssembly,
    setTimeout, clearTimeout, console,
  });
  runInContext(runtimeSource, context);
  const runtime = new context.Go();
  runtime.argv = ["profitctl", ...args];
  runtime.env = { NO_COLOR: "1" };
  runtime.exit = (code) => { exitCode = code; };
  await runtime.run(await WebAssembly.instantiate(module, runtime.importObject));
  return { output: transcript, exitCode };
}
for (const step of terminalStory) {
  const result = await run(step.args);
  assert.equal(result.exitCode, step.exitCode, step.title + "\n" + result.output);
  assert.ok(result.output.length > 0, step.title);
}
const comparisons = await run(["compare", "demo/baseline.yml", "demo/growth.yml", "demo/optimized.yml", "--json"]);
assert.equal(comparisons.exitCode, 1, comparisons.output);
const comparison = JSON.parse(comparisons.output);
assert.deepEqual(comparison.scenarios.map((scenario) => scenario.name), ["baseline", "growth", "optimized"]);
for (const [index, margin] of [66.5, -29.5, 60.5].entries()) {
  assert.ok(Math.abs(comparison.scenarios[index].operating_margin - margin) < 1e-8);
}
assert.deepEqual(comparison.scenarios.map((scenario) => scenario.covenants_passed), [true, false, true]);
assert.equal(comparison.scenarios[1].revenue, comparison.scenarios[2].revenue);
assert.ok(comparison.scenarios[2].cost_per_user < comparison.scenarios[1].cost_per_user);
const replay = await run(terminalStory[1].args);
assert.equal(replay.exitCode, 0, replay.output);
const unavailable = await run(["validate", "-f", "/etc/passwd"]);
assert.equal(unavailable.exitCode, 2);
assert.match(unavailable.output, /failed to read config/);
console.log("Real WebAssembly CLI: all story commands, a repeated run, and workspace isolation verified.");
