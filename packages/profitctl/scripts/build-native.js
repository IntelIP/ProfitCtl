import { chmodSync, mkdirSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { join } from "node:path";
import { nativeTarget } from "../lib/platform.js";

const packageRoot = fileURLToPath(new URL("../", import.meta.url));
const repositoryRoot = fileURLToPath(new URL("../../../", import.meta.url));
const target = nativeTarget();
const outputDirectory = join(packageRoot, "native", target.directory);
const output = join(outputDirectory, target.executable);

mkdirSync(outputDirectory, { recursive: true });

const result = spawnSync(
  "go",
  ["build", "-trimpath", "-o", output, "."],
  { cwd: repositoryRoot, stdio: "inherit" },
);

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

if (process.platform !== "win32") {
  chmodSync(output, 0o755);
}

console.log(`Built ${output}`);
