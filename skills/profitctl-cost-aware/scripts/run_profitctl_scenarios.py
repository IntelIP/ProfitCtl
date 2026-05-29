#!/usr/bin/env python3
"""Run bundled ProfitCtl scenarios and return compact JSON for agents."""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


DEFAULT_PROFITCTL_REPO = Path("/Users/hudson/Documents/GitHub/IntelIP/ProfitCtl")
SKILL_DIR = Path(__file__).resolve().parents[1]
TEMPLATE_DIR = SKILL_DIR / "references" / "templates"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--template",
        action="append",
        dest="templates",
        help="Template name without .yml, or path to a scenario YAML file. Repeatable.",
    )
    parser.add_argument(
        "--compare",
        action="store_true",
        help="Run profitctl compare when at least two scenarios are selected.",
    )
    parser.add_argument(
        "--profitctl-repo",
        default=os.environ.get("PROFITCTL_REPO", str(DEFAULT_PROFITCTL_REPO)),
        help="ProfitCtl repo path for go-run fallback.",
    )
    return parser.parse_args()


def resolve_profitctl(repo: Path) -> tuple[list[str], Path | None]:
    binary = shutil.which("profitctl")
    if binary:
        return [binary], None
    if (repo / "go.mod").exists() and (repo / "main.go").exists():
        return ["go", "run", "."], repo
    raise SystemExit(
        json.dumps(
            {
                "status": "error",
                "reason": "profitctl not found on PATH and ProfitCtl repo fallback is unavailable",
                "profitctl_repo": str(repo),
            },
            indent=2,
        )
    )


def resolve_template(raw: str) -> Path:
    candidate = Path(raw).expanduser()
    if candidate.exists():
        return candidate.resolve()

    for suffix in ("", ".yml", ".yaml"):
        bundled = TEMPLATE_DIR / f"{raw}{suffix}"
        if bundled.exists():
            return bundled.resolve()

    available = sorted(path.stem for path in TEMPLATE_DIR.glob("*.yml"))
    raise SystemExit(
        json.dumps(
            {
                "status": "error",
                "reason": f"template not found: {raw}",
                "available_templates": available,
            },
            indent=2,
        )
    )


def run_command(command: list[str], cwd: Path | None) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        command,
        cwd=str(cwd) if cwd else None,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )


def compact_simulation(data: dict[str, object]) -> dict[str, object]:
    scenario = data.get("scenario", {}) or {}
    costs = data.get("costs", {}) or {}
    fixed = costs.get("fixed", {}) if isinstance(costs, dict) else {}
    margin = data.get("margin", {}) or {}
    stress = data.get("stress_test", {}) or {}
    p95 = stress.get("p95", {}) if isinstance(stress, dict) else {}
    revenue = data.get("revenue", {}) or {}
    covenants = data.get("covenants", {}) or {}

    return {
        "users": scenario.get("users"),
        "growth_factor": scenario.get("growth_factor"),
        "revenue": revenue.get("total"),
        "fixed_monthly_cost": fixed.get("monthly"),
        "total_cost": costs.get("total") if isinstance(costs, dict) else None,
        "gross_margin": margin.get("gross") if isinstance(margin, dict) else None,
        "p95_margin": p95.get("margin") if isinstance(p95, dict) else None,
        "cost_per_user": margin.get("cost_per_user") if isinstance(margin, dict) else None,
        "p95_cost_per_user": p95.get("cost_per_user") if isinstance(p95, dict) else None,
        "covenants_passed": covenants.get("passed") if isinstance(covenants, dict) else None,
        "violations": covenants.get("violations", []) if isinstance(covenants, dict) else [],
    }


def scenario_name(path: Path) -> str:
    return path.stem


def main() -> int:
    args = parse_args()
    selected = args.templates or [
        "cloudflare-workers-ai-saas",
        "cloud-run-ai-saas",
        "vercel-ai-saas",
    ]
    repo = Path(args.profitctl_repo).expanduser().resolve()
    profitctl_cmd, cwd = resolve_profitctl(repo)

    output: dict[str, object] = {
        "status": "ok",
        "profitctl_command": " ".join(profitctl_cmd),
        "scenarios": [],
    }

    with tempfile.TemporaryDirectory(prefix="profitctl-skill-") as tmp_raw:
        tmp = Path(tmp_raw)
        scenario_paths: list[Path] = []

        for raw in selected:
            source = resolve_template(raw)
            destination = tmp / source.name
            shutil.copy2(source, destination)
            scenario_paths.append(destination)

            validate = run_command(profitctl_cmd + ["validate", "-f", str(destination)], cwd)
            simulate = run_command(profitctl_cmd + ["simulate", "-f", str(destination), "--json"], cwd)

            scenario_result: dict[str, object] = {
                "name": scenario_name(destination),
                "source_template": str(source),
                "validate_exit_code": validate.returncode,
                "simulate_exit_code": simulate.returncode,
                "validate_stdout": validate.stdout.strip(),
            }

            if simulate.stdout.strip():
                try:
                    parsed = json.loads(simulate.stdout)
                    scenario_result["summary"] = compact_simulation(parsed)
                except json.JSONDecodeError:
                    scenario_result["simulate_stdout"] = simulate.stdout.strip()

            if validate.stderr.strip():
                scenario_result["validate_stderr"] = validate.stderr.strip()
            if simulate.stderr.strip():
                scenario_result["simulate_stderr"] = simulate.stderr.strip()

            cast_scenarios = output["scenarios"]
            assert isinstance(cast_scenarios, list)
            cast_scenarios.append(scenario_result)

        if args.compare and len(scenario_paths) >= 2:
            compare = run_command(
                profitctl_cmd + ["compare", *[str(path) for path in scenario_paths]],
                cwd,
            )
            output["compare"] = {
                "exit_code": compare.returncode,
                "stdout": compare.stdout.strip(),
                "stderr": compare.stderr.strip(),
            }

    print(json.dumps(output, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
