import { execFileSync } from "node:child_process";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { gzipSync } from "node:zlib";

const repository = fileURLToPath(new URL("../../", import.meta.url));
const output = new URL("../public/demo/", import.meta.url);
const temporary = await mkdtemp(join(tmpdir(), "profitctl-browser-"));
try {
  execFileSync("go", ["build", "-ldflags=-s -w -X main.version=browser-demo", "-o", join(temporary, "profitctl.wasm"), "."], {
    cwd: repository,
    env: { ...process.env, GOOS: "js", GOARCH: "wasm", CGO_ENABLED: "0" },
    stdio: "inherit",
  });
  const compressed = gzipSync(await readFile(join(temporary, "profitctl.wasm")), { level: 9 });
  if (compressed.length > 25 * 1024 * 1024) throw new Error("Browser demo exceeds the Pages asset limit.");
  await mkdir(output, { recursive: true });
  await writeFile(new URL("profitctl.wasm.gz", output), compressed);
  const goRoot = execFileSync("go", ["env", "GOROOT"], { cwd: repository, encoding: "utf8" }).trim();
  for (const [source, name] of [["lib/wasm/wasm_exec.js", "wasm_exec.js"], ["LICENSE", "GO-LICENSE.txt"]]) {
    const destination = new URL(name, output);
    await rm(destination, { force: true });
    await writeFile(destination, await readFile(join(goRoot, source)));
  }
  const scenarios = {};
  for (const name of ["baseline", "growth", "optimized"]) {
    const text = await readFile(join(repository, "examples/website_demo", `${name}.yml`), "utf8");
    scenarios[`demo/${name}.yml`] = text;
    await writeFile(new URL(`${name}.yml`, output), text);
  }
  await writeFile(new URL("scenarios.json", output), JSON.stringify(scenarios));
  console.log(`Real ProfitCTL browser demo built (${(compressed.length / 1024 / 1024).toFixed(2)} MiB compressed).`);
} finally {
  await rm(temporary, { recursive: true, force: true });
}
