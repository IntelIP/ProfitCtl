#!/usr/bin/env python3
"""Run source-backed ProfitCtl assessments and bundled scenarios as compact JSON."""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


SKILL_DIR = Path(__file__).resolve().parents[1]
TEMPLATE_DIR = SKILL_DIR / "references" / "templates"
DEFAULT_ASSESS_MODEL = "openai/gpt-5.6-terra"


class HelperError(Exception):
    """A compact, user-actionable helper failure."""

    def __init__(self, reason: str, **details: object) -> None:
        super().__init__(reason)
        self.reason = reason
        self.details = details

    def payload(self) -> dict[str, object]:
        return {"status": "error", "reason": self.reason, **self.details}


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
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
    evidence = parser.add_mutually_exclusive_group()
    evidence.add_argument(
        "--assess-path",
        help="Run a required source-backed assessment for this codebase path.",
    )
    evidence.add_argument(
        "--assessment",
        help="Read an existing assessment JSON without provider calls.",
    )
    parser.add_argument(
        "--model",
        default=DEFAULT_ASSESS_MODEL,
        help="OpenRouter model id for --assess-path.",
    )
    parser.add_argument(
        "--profitctl-repo",
        default=os.environ.get("PROFITCTL_REPO"),
        help="Optional source checkout for a go-run developer fallback.",
    )
    args = parser.parse_args(argv)
    if (args.assess_path or args.assessment) and (args.templates or args.compare):
        parser.error("assessment and scenario modes must be run separately")
    return args


def executable_name() -> str:
    return "profitctl.exe" if os.name == "nt" else "profitctl"


def is_executable(path: Path) -> bool:
    return path.is_file() and (os.name == "nt" or os.access(path, os.X_OK))


def resolve_profitctl(
    repo_raw: str | None,
    skill_dir: Path = SKILL_DIR,
) -> tuple[list[str], Path | None]:
    bundled = skill_dir / "bin" / executable_name()
    if is_executable(bundled):
        return [str(bundled)], None

    binary = shutil.which("profitctl")
    if binary:
        return [binary], None

    repo_candidates: list[Path] = []
    if repo_raw:
        repo_candidates.append(Path(repo_raw).expanduser().resolve())
    if len(skill_dir.parents) >= 2:
        repo_candidates.append(skill_dir.parents[1])

    for repo in repo_candidates:
        if (
            shutil.which("go")
            and (repo / "go.mod").is_file()
            and (repo / "main.go").is_file()
        ):
            return ["go", "run", "."], repo

    raise HelperError(
        "profitctl not found on PATH or in the installed skill",
        next_action="Install the ProfitCtl release package or add profitctl to PATH.",
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
    raise HelperError(
        f"template not found: {raw}",
        available_templates=available,
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


def load_json(path: Path) -> dict[str, object]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise HelperError(
            f"assessment JSON cannot be read: {error}",
            assessment=str(path),
        ) from error
    if not isinstance(value, dict):
        raise HelperError("assessment JSON must contain an object")
    return value


def compact_assessment(
    data: dict[str, object],
    evidence_mode: str,
) -> dict[str, object]:
    providers: list[dict[str, object]] = []
    for raw_provider in data.get("providers", []) or []:
        if not isinstance(raw_provider, dict):
            continue
        providers.append(
            {
                "name": raw_provider.get("name"),
                "provider": raw_provider.get("provider"),
                "official_domain": raw_provider.get("official_domain"),
                "evidence": raw_provider.get("evidence", []),
            }
        )

    sourced_rates: list[dict[str, object]] = []
    for raw_receipt in data.get("pricing_receipts", []) or []:
        if not isinstance(raw_receipt, dict):
            continue
        sourced_rates.append(
            {
                "provider": raw_receipt.get("provider"),
                "role": raw_receipt.get("role"),
                "source_url": raw_receipt.get("url"),
                "captured_at": raw_receipt.get("captured_at"),
                "highlights": raw_receipt.get("highlights", []),
            }
        )

    draft = data.get("draft", {}) or {}
    if not isinstance(draft, dict):
        draft = {}
    assumptions = draft.get("assumptions", []) or []
    inferred_scale = [
        assumption
        for assumption in assumptions
        if isinstance(assumption, dict) and assumption.get("source") == "inferred"
    ]
    repo_assumptions = [
        assumption
        for assumption in assumptions
        if isinstance(assumption, dict) and assumption.get("source") == "repo_detected"
    ]
    recommendation = draft.get("recommendation", {}) or {}

    return {
        "evidence_mode": evidence_mode,
        "facts": {
            "path": data.get("path"),
            "model": data.get("model"),
            "analyzed_files": data.get("analyzed_files"),
            "providers": providers,
            "repo_detected_scale": repo_assumptions,
        },
        "sourced_rates": sourced_rates,
        "inferred_scale": inferred_scale,
        "economics": {
            "cost_lines": draft.get("cost_lines", []),
            "estimated_monthly_cost_usd": data.get("estimated_monthly_cost_usd"),
        },
        "next_action": recommendation,
    }


def run_assessment(
    assess_path: str,
    model: str,
    repo_raw: str | None,
) -> tuple[dict[str, object], str]:
    missing_env = [
        name
        for name in ("OPENROUTER_API_KEY", "EXA_API_KEY")
        if not os.environ.get(name, "").strip()
    ]
    if missing_env:
        raise HelperError(
            "source-backed assessment credentials are missing",
            missing_env=missing_env,
            next_action="Set the missing keys and rerun; do not substitute unsourced prices.",
        )

    profitctl_cmd, cwd = resolve_profitctl(repo_raw)
    with tempfile.TemporaryDirectory(prefix="profitctl-assess-") as tmp_raw:
        report_path = Path(tmp_raw) / "assessment.json"
        command = profitctl_cmd + [
            "assess",
            "--path",
            str(Path(assess_path).expanduser().resolve()),
            "--model",
            model,
            "--out",
            str(report_path),
        ]
        result = run_command(command, cwd)
        if result.returncode != 0:
            raise HelperError(
                "profitctl assess failed",
                exit_code=result.returncode,
                stderr=result.stderr.strip(),
            )
        return compact_assessment(load_json(report_path), "live_assessment"), " ".join(
            profitctl_cmd
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
        "cost_per_user": margin.get("cost_per_user")
        if isinstance(margin, dict)
        else None,
        "p95_cost_per_user": p95.get("cost_per_user")
        if isinstance(p95, dict)
        else None,
        "covenants_passed": covenants.get("passed")
        if isinstance(covenants, dict)
        else None,
        "violations": covenants.get("violations", [])
        if isinstance(covenants, dict)
        else [],
    }


def scenario_name(path: Path) -> str:
    return path.stem


def run_scenarios(
    selected: list[str],
    compare_requested: bool,
    repo_raw: str | None,
) -> tuple[dict[str, object], int]:
    profitctl_cmd, cwd = resolve_profitctl(repo_raw)
    output: dict[str, object] = {
        "profitctl_command": " ".join(profitctl_cmd),
        "scenarios": [],
    }
    failed = False

    with tempfile.TemporaryDirectory(prefix="profitctl-skill-") as tmp_raw:
        tmp = Path(tmp_raw)
        scenario_paths: list[Path] = []

        for raw in selected:
            source = resolve_template(raw)
            destination = tmp / source.name
            shutil.copy2(source, destination)
            scenario_paths.append(destination)

            validate = run_command(
                profitctl_cmd + ["validate", "-f", str(destination)],
                cwd,
            )
            simulate = run_command(
                profitctl_cmd + ["simulate", "-f", str(destination), "--json"],
                cwd,
            )
            failed = failed or validate.returncode != 0 or simulate.returncode not in (0, 1)

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

        if compare_requested and len(scenario_paths) >= 2:
            compare = run_command(
                profitctl_cmd + ["compare", *[str(path) for path in scenario_paths]],
                cwd,
            )
            failed = failed or compare.returncode not in (0, 1)
            output["compare"] = {
                "exit_code": compare.returncode,
                "stdout": compare.stdout.strip(),
                "stderr": compare.stderr.strip(),
            }

    return output, 3 if failed else 0


def main(argv: list[str] | None = None) -> int:
    try:
        args = parse_args(argv)
        output: dict[str, object] = {"status": "ok"}

        if args.assessment:
            output["assessment"] = compact_assessment(
                load_json(Path(args.assessment).expanduser().resolve()),
                "saved_assessment",
            )
            print(json.dumps(output, indent=2, sort_keys=True))
            return 0

        if args.assess_path:
            assessment, command = run_assessment(
                args.assess_path,
                args.model,
                args.profitctl_repo,
            )
            output["profitctl_command"] = command
            output["assessment"] = assessment
            print(json.dumps(output, indent=2, sort_keys=True))
            return 0

        selected = args.templates or [
            "cloudflare-workers-ai-saas",
            "cloud-run-ai-saas",
            "vercel-ai-saas",
        ]
        scenario_output, exit_code = run_scenarios(
            selected,
            args.compare,
            args.profitctl_repo,
        )
        output.update(scenario_output)
        if exit_code != 0:
            output["status"] = "error"
        print(json.dumps(output, indent=2, sort_keys=True))
        return exit_code
    except HelperError as error:
        print(json.dumps(error.payload(), indent=2, sort_keys=True))
        return 2


if __name__ == "__main__":
    sys.exit(main())
