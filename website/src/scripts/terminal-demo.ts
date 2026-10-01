import { terminalStory } from "./terminal-story";

export function initializeTerminalDemo() {
  const root = document.querySelector<HTMLElement>("[data-terminal-demo]");
  if (!root) return;
  const command = root.querySelector<HTMLElement>("[data-terminal-command]")!;
  const output = root.querySelector<HTMLElement>("[data-terminal-output]")!;
  const status = root.querySelector<HTMLElement>("[data-terminal-status]")!;
  const detail = root.querySelector<HTMLElement>("[data-terminal-detail]")!;
  const toggle = root.querySelector<HTMLButtonElement>("[data-terminal-toggle]")!;
  const download = root.querySelector<HTMLAnchorElement>("[data-terminal-download]")!;
  const steps = [...root.querySelectorAll<HTMLButtonElement>("[data-terminal-step]")];
  const reducedMotion = matchMedia("(prefers-reduced-motion: reduce)");
  const playback = new EventTarget();
  let visible = false;
  let paused = false;
  let controller: AbortController | undefined;
  let worker: Worker | undefined;
  let failed = false;
  const playing = () => !paused && visible && !document.hidden;

  function changed(signal: AbortSignal, timeout?: number) {
    return new Promise<void>((resolve, reject) => {
      const cleanup = () => { clearTimeout(timer); playback.removeEventListener("change", done); signal.removeEventListener("abort", abort); };
      const done = () => { cleanup(); resolve(); };
      const abort = () => { cleanup(); reject(signal.reason); };
      const timer = timeout === undefined ? undefined : setTimeout(done, timeout);
      playback.addEventListener("change", done, { once: true });
      signal.addEventListener("abort", abort, { once: true });
      if (signal.aborted) abort();
    });
  }
  async function wait(milliseconds: number, signal: AbortSignal) {
    while (milliseconds > 0) {
      if (signal.aborted) throw signal.reason;
      if (!playing()) { await changed(signal); continue; }
      const start = performance.now();
      await changed(signal, milliseconds);
      milliseconds -= performance.now() - start;
    }
  }
  function update() {
    toggle.textContent = failed ? "Retry demo" : !controller ? "Run demo" : paused ? "Resume" : "Pause";
    toggle.setAttribute("aria-label", toggle.textContent + " terminal demo");
    playback.dispatchEvent(new Event("change"));
  }
  function execute(index: number, signal: AbortSignal) {
    worker ??= new Worker(new URL("./terminal.worker.ts", import.meta.url), { type: "module" });
    return new Promise<{ output: string; exitCode: number }>((resolve, reject) => {
      const cleanup = () => { clearTimeout(timer); signal.removeEventListener("abort", abort); };
      const abort = () => { cleanup(); worker?.terminate(); worker = undefined; reject(signal.reason); };
      const timer = setTimeout(() => { cleanup(); worker?.terminate(); worker = undefined; reject(new Error("The demo took too long. Please retry or run the downloaded scenario locally.")); }, 45000);
      signal.addEventListener("abort", abort, { once: true });
      worker!.onerror = () => { cleanup(); reject(new Error("The demo could not start. Please retry or run the scenario locally.")); };
      worker!.onmessage = (event) => {
        cleanup();
        if (event.data.error) reject(new Error(event.data.error));
        else resolve(event.data);
      };
      worker!.postMessage(index);
    });
  }
  async function run(first: number, signal: AbortSignal) {
    let index = first;
    let cycles = 0;
    try {
      while (!signal.aborted) {
        await wait(1, signal);
        const step = terminalStory[index]!;
        steps.forEach((button, position) => button.setAttribute("aria-current", position === index ? "step" : "false"));
        detail.textContent = step.detail;
        download.href = `/demo/${step.scenario}.yml`;
        command.textContent = "$ ";
        output.textContent = "";
        root!.dataset.state = "running";
        const text = `profitctl ${step.args.join(" ")}`;
        if (reducedMotion.matches) command.textContent += text;
        else for (const character of text) { await wait(18, signal); command.textContent += character; }
        status.textContent = "Running ProfitCTL…";
        const result = await execute(index, signal);
        if (result.exitCode !== step.exitCode) throw new Error("The scenario returned an unexpected result. Download it to inspect the inputs.");
        const transcript = result.output.replace(/\x1b\[[0-9;]*m/g, "");
        for (const line of transcript.trimEnd().split("\n")) {
          await wait(reducedMotion.matches ? 1 : 45, signal);
          output.textContent += line + "\n";
        }
        const marginFailed = step.args[0] === "simulate" && result.exitCode === 1;
        root!.dataset.state = marginFailed ? "risk" : "pass";
        status.textContent = marginFailed ? "Margin target failed" : step.args[0] === "compare" ? "Comparison complete" : "Command complete";
        await wait(index === 3 ? 8000 : 5500, signal);
        index++;
        if (index === terminalStory.length) {
          root!.dataset.cycles = String(++cycles);
          if (reducedMotion.matches) { status.textContent = "Demo complete"; controller = undefined; update(); return; }
          index = 0;
        }
      }
    } catch (failure) {
      if (signal.aborted) return;
      failed = true;
      root!.dataset.state = "error";
      status.textContent = "Demo unavailable";
      output.textContent = failure instanceof Error ? failure.message : "Please retry or run the scenario locally.";
      worker?.terminate(); worker = undefined; controller = undefined;
      update();
    }
  }
  function launch(index = 0) {
    controller?.abort();
    controller = new AbortController();
    paused = false;
    failed = false;
    visible = true;
    root!.dataset.cycles = "0";
    update();
    void run(index, controller.signal);
  }
  toggle.addEventListener("click", () => { if (!controller) launch(); else { paused = !paused; update(); } });
  root.querySelector<HTMLButtonElement>("[data-terminal-replay]")!.addEventListener("click", () => launch());
  steps.forEach((button, index) => button.addEventListener("click", () => launch(index)));
  const observer = new IntersectionObserver(([entry]) => {
    visible = Boolean(entry?.isIntersecting);
    update();
    if (visible && !controller && !failed && !reducedMotion.matches) launch();
  }, { threshold: 0.15 });
  observer.observe(root);
  document.addEventListener("visibilitychange", update);
  window.addEventListener("pagehide", () => { controller?.abort(); worker?.terminate(); observer.disconnect(); }, { once: true });
  update();
}
