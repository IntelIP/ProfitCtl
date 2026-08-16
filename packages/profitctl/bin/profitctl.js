#!/usr/bin/env bun

import { existsSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { join } from "node:path";
import { nativeTarget } from "../lib/platform.js";

const packageRoot = fileURLToPath(new URL("../", import.meta.url));

let target;
try {
  target = nativeTarget();
} catch (error) {
  console.error(error.message);
  process.exit(1);
}

const executable = join(packageRoot, "native", target.directory, target.executable);
if (!existsSync(executable)) {
  console.error(`ProfitCtl native binary is missing: ${executable}`);
  console.error("From a source checkout, run: bun run build:native");
  process.exit(1);
}

const result = spawnSync(executable, process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 1);
