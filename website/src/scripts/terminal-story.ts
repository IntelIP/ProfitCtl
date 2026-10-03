export const terminalStory = [
  { title: "Check the scenario", detail: "A fictional AI SaaS starts with 1,000 users, a $20 plan, and a 60% margin target.", args: ["validate", "-f", "demo/baseline.yml"], exitCode: 0, scenario: "baseline" },
  { title: "Run the baseline", detail: "At 60 research requests per user, the product clears its margin target.", args: ["simulate", "-f", "demo/baseline.yml"], exitCode: 0, scenario: "baseline" },
  { title: "Let usage grow", detail: "Usage rises to 300 requests. Revenue stays the same, while the margin target fails.", args: ["simulate", "-f", "demo/growth.yml"], exitCode: 1, scenario: "growth" },
  { title: "Compare a cheaper route", detail: "At the same usage, a cheaper request route brings the margin back above 60%.", args: ["compare", "demo/baseline.yml", "demo/growth.yml", "demo/optimized.yml"], exitCode: 1, scenario: "optimized" },
];
