import { createDemoFilesystem } from "./cli-filesystem";
import { terminalStory } from "./terminal-story";

type GoRuntime = { argv: string[]; env: Record<string, string>; importObject: WebAssembly.Imports; exit: (code: number) => void; run: (instance: WebAssembly.Instance) => Promise<void> };
const host = globalThis as unknown as { fs: unknown; process: unknown; Go: new () => GoRuntime };
let initialization: Promise<{ module: WebAssembly.Module; files: Record<string, string> }> | undefined;

async function initialize() {
  if (!globalThis.WebAssembly) throw new Error("This browser cannot run the local demo. Download a scenario and run it with the CLI.");
  const [binary, scenarios] = await Promise.all([fetch("/demo/profitctl.wasm.gz"), fetch("/demo/scenarios.json")]);
  if (!binary.ok || !scenarios.ok || !binary.body) throw new Error("The demo could not download. Please retry.");
  let bytes = await binary.arrayBuffer();
  const signature = new Uint8Array(bytes, 0, 2);
  // Some static hosts decode gzip during fetch; others serve the compressed file.
  if (signature[0] === 0x1f && signature[1] === 0x8b) {
    if (!globalThis.DecompressionStream) throw new Error("This browser cannot load the demo. Download a scenario and run it with the CLI.");
    bytes = await new Response(new Blob([bytes]).stream().pipeThrough(new DecompressionStream("gzip"))).arrayBuffer();
  }
  const files = await scenarios.json() as Record<string, string>;
  host.process = { getuid: () => -1, getgid: () => -1, geteuid: () => -1, getegid: () => -1, getgroups: () => [], pid: -1, ppid: -1, cwd: () => "/", chdir: () => { throw new Error("Read-only demo workspace"); }, umask: () => 0 };
  host.fs = createDemoFilesystem(files, () => {});
  const runtimeUrl = "/demo/wasm_exec.js";
  await import(/* @vite-ignore */ runtimeUrl);
  return { module: await WebAssembly.compile(bytes), files };
}

let busy = false;
self.onmessage = async (event: MessageEvent<number>) => {
  const step = terminalStory[event.data];
  if (!Number.isInteger(event.data) || !step || busy) return;
  busy = true;
  try {
    const { module, files } = await (initialization ??= initialize());
    let output = "";
    let exitCode = 0;
    host.fs = createDemoFilesystem(files, (text) => { output += text; });
    const runtime = new host.Go();
    runtime.argv = ["profitctl", ...step.args];
    runtime.env = { NO_COLOR: "1" };
    runtime.exit = (code) => { exitCode = code; };
    await runtime.run(await WebAssembly.instantiate(module, runtime.importObject));
    self.postMessage({ output, exitCode });
  } catch (failure) {
    self.postMessage({ error: failure instanceof Error ? failure.message : "The demo could not run. Please retry." });
  } finally { busy = false; }
};
