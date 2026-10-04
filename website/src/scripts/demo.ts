export interface ExampleInputs {
  users: number;
  requests: number;
  price: number;
}

export function calculateExample({ users, requests, price }: ExampleInputs) {
  if (![users, requests, price].every(Number.isFinite) || users <= 0 || requests < 0 || price <= 0) {
    throw new RangeError("Use positive users and price, and nonnegative requests.");
  }
  const revenue = users * price;
  const option = (requestCost: number) => {
    const cost = 400 + users * (1.5 + requests * requestCost);
    return { cost, unitCost: cost / users, margin: (revenue - cost) / revenue * 100 };
  };
  const baseline = option(0.01);
  const candidate = option(0.08);
  return {
    revenue, baseline, candidate,
    addedCost: candidate.cost - baseline.cost,
    minimumPrice: Math.ceil(candidate.unitCost / 0.4 * 100) / 100,
    passes: candidate.margin >= 60,
  };
}

export const money = (value: number, digits = 0) => value.toLocaleString("en-US", {
  style: "currency", currency: "USD", minimumFractionDigits: digits, maximumFractionDigits: digits,
});
export const percent = (value: number) => `${value.toFixed(1)}%`;

export function initializeDemo() {
  const element = document.querySelector<HTMLElement>("[data-demo]");
  if (!element) return;
  const demo = element;
  const users = demo.querySelector<HTMLInputElement>("#demo-users")!;
  const requests = demo.querySelector<HTMLInputElement>("#demo-requests")!;
  const price = demo.querySelector<HTMLInputElement>("#demo-price")!;
  const presets: Record<string, ExampleInputs> = {
    feature: { users: 1000, requests: 60, price: 20 },
    pricing: { users: 1000, requests: 60, price: 12 },
    growth: { users: 10000, requests: 220, price: 20 },
  };
  const setText = (selector: string, value: string) => {
    demo.querySelector<HTMLElement>(selector)!.textContent = value;
  };
  const buttons = [...demo.querySelectorAll<HTMLButtonElement>("[data-preset]")];

  function update() {
    try {
      const inputs = { users: users.valueAsNumber, requests: requests.valueAsNumber, price: price.valueAsNumber };
      const result = calculateExample(inputs);
      setText("#users-value", inputs.users.toLocaleString("en-US"));
      setText("#requests-value", inputs.requests.toLocaleString("en-US"));
      setText("#price-value", money(inputs.price));
      setText("[data-revenue]", money(result.revenue));
      for (const name of ["baseline", "candidate"] as const) {
        const option = result[name];
        setText(`[data-${name}-cost]`, money(option.cost));
        setText(`[data-${name}-margin]`, percent(option.margin));
        setText(`[data-${name}-unit]`, money(option.unitCost, 2));
        const bar = demo.querySelector<HTMLElement>(`[data-${name}-bar]`)!;
        bar.style.width = `${Math.min(100, option.cost / result.revenue * 100)}%`;
      }
      const verdict = demo.querySelector<HTMLElement>("[data-verdict]")!;
      verdict.dataset.state = result.passes ? "pass" : "risk";
      setText("[data-verdict-title]", result.passes ? "The feature clears your margin target." : "The feature misses your margin target.");
      setText("[data-verdict-detail]", result.passes
        ? `Research adds ${money(result.addedCost)} per month. The modeled margin stays above 60%.`
        : `At this usage, a price of ${money(result.minimumPrice, 2)} per user would meet the 60% target.`);
      setText("[data-demo-error]", "");
    } catch {
      setText("[data-demo-error]", "Choose valid users, usage, and price values to compare the scenarios.");
    }
  }

  buttons.forEach((button) => button.addEventListener("click", () => {
    const preset = presets[button.dataset.preset!];
    users.value = String(preset.users);
    requests.value = String(preset.requests);
    price.value = String(preset.price);
    buttons.forEach((item) => item.setAttribute("aria-pressed", String(item === button)));
    update();
  }));
  [users, requests, price].forEach((input) => input.addEventListener("input", () => {
    buttons.forEach((button) => button.setAttribute("aria-pressed", "false"));
    update();
  }));

  const copyButton = document.querySelector<HTMLButtonElement>("[data-copy-command]");
  copyButton?.addEventListener("click", async () => {
    const code = document.querySelector<HTMLElement>("[data-start-command]")!;
    try {
      await navigator.clipboard.writeText(code.textContent!.trim());
      copyButton.textContent = "Copied";
    } catch {
      copyButton.textContent = "Select the command to copy";
    }
  });

  if (!matchMedia("(prefers-reduced-motion: reduce)").matches && "IntersectionObserver" in window) {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add("is-visible");
          observer.unobserve(entry.target);
        }
      });
    }, { threshold: 0.12 });
    document.querySelectorAll("[data-reveal]").forEach((element) => observer.observe(element));
  }
  update();
}
