#!/usr/bin/env python3
"""Portable behavior tests for the bundled ProfitCtl Codex skill."""

from __future__ import annotations

import contextlib
import importlib.util
import io
import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest import mock


ROOT = Path(__file__).resolve().parents[2]
SKILL_DIR = ROOT / "skills" / "profitctl-cost-aware"
SCRIPT_PATH = SKILL_DIR / "scripts" / "run_profitctl_scenarios.py"

SPEC = importlib.util.spec_from_file_location("profitctl_skill_helper", SCRIPT_PATH)
assert SPEC and SPEC.loader
HELPER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(HELPER)


class PortableSkillTests(unittest.TestCase):
    def test_skill_has_required_trigger_and_non_trigger_boundaries(self) -> None:
        body = (SKILL_DIR / "SKILL.md").read_text(encoding="utf-8")
        self.assertIn("run `profitctl assess`. This step is required", body)
        self.assertIn("Do not run ProfitCtl for unrelated coding", body)
        self.assertIn("Never replace them with Exa research", body)

    def test_skill_contains_no_developer_machine_path(self) -> None:
        for path in SKILL_DIR.rglob("*"):
            if path.suffix.lower() not in {".md", ".py", ".yaml", ".yml"}:
                continue
            self.assertNotIn(
                "/Users/",
                path.read_text(encoding="utf-8"),
                msg=f"machine-specific path in {path.relative_to(SKILL_DIR)}",
            )

    def test_skill_uses_portable_standards_judge(self) -> None:
        body = (SKILL_DIR / "SKILL.md").read_text(encoding="utf-8")
        self.assertIn(
            "<this-skill-directory>/bin/profitctl-standards",
            body,
        )
        self.assertIn(
            "go run scripts/judge_cost_standards.go",
            body,
        )

    def test_resolve_profitctl_from_installed_skill(self) -> None:
        with tempfile.TemporaryDirectory() as temp_raw:
            skill_dir = Path(temp_raw) / "profitctl-cost-aware"
            binary = skill_dir / "bin" / HELPER.executable_name()
            binary.parent.mkdir(parents=True)
            binary.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
            binary.chmod(0o755)

            with mock.patch.object(
                HELPER.shutil,
                "which",
                return_value="/tmp/stale-profitctl",
            ):
                command, cwd = HELPER.resolve_profitctl(None, skill_dir)

            self.assertEqual(command, [str(binary)])
            self.assertIsNone(cwd)

    def test_resolve_profitctl_from_path_without_bundle(self) -> None:
        with tempfile.TemporaryDirectory() as temp_raw:
            skill_dir = Path(temp_raw) / "profitctl-cost-aware"
            skill_dir.mkdir()
            with mock.patch.object(
                HELPER.shutil,
                "which",
                side_effect=lambda name: "/usr/local/bin/profitctl"
                if name == "profitctl"
                else None,
            ):
                command, cwd = HELPER.resolve_profitctl(None, skill_dir)

        self.assertEqual(command, ["/usr/local/bin/profitctl"])
        self.assertIsNone(cwd)

    def test_missing_binary_returns_install_action(self) -> None:
        with tempfile.TemporaryDirectory() as temp_raw:
            skill_dir = Path(temp_raw) / "profitctl-cost-aware"
            skill_dir.mkdir()
            with mock.patch.object(HELPER.shutil, "which", return_value=None):
                with self.assertRaises(HELPER.HelperError) as caught:
                    HELPER.resolve_profitctl(None, skill_dir)

        self.assertIn("not found", caught.exception.reason)
        self.assertIn("Install", caught.exception.details["next_action"])

    def test_live_assessment_names_missing_keys_without_fallback(self) -> None:
        output = io.StringIO()
        with tempfile.TemporaryDirectory() as temp_raw:
            with mock.patch.dict(os.environ, {}, clear=True):
                with mock.patch.object(
                    HELPER,
                    "resolve_profitctl",
                    side_effect=AssertionError("binary lookup must not run"),
                ):
                    with contextlib.redirect_stdout(output):
                        exit_code = HELPER.main(["--assess-path", temp_raw])

        payload = json.loads(output.getvalue())
        self.assertEqual(exit_code, 2)
        self.assertEqual(
            set(payload["missing_env"]),
            {"OPENROUTER_API_KEY", "EXA_API_KEY"},
        )
        self.assertIn("do not substitute unsourced prices", payload["next_action"])

    def test_saved_assessment_is_compacted_without_binary_or_keys(self) -> None:
        artifact = {
            "path": "/workspace/example",
            "model": "openai/gpt-5.6-terra",
            "analyzed_files": 1,
            "providers": [
                {
                    "name": "AWS",
                    "provider": "aws",
                    "official_domain": "aws.amazon.com",
                    "evidence": [
                        {
                            "file": "go.mod",
                            "excerpt": "github.com/aws/aws-sdk-go v1.48.0",
                        }
                    ],
                }
            ],
            "pricing_receipts": [
                {
                    "provider": "aws",
                    "role": "codebacked_provider",
                    "url": "https://aws.amazon.com/pricing",
                    "captured_at": "2026-08-04T20:00:00Z",
                    "highlights": ["$0.10 per GB"],
                }
            ],
            "draft": {
                "assumptions": [
                    {
                        "name": "storage_gb",
                        "value": "10",
                        "source": "inferred",
                        "rationale": "starter scale",
                    }
                ],
                "cost_lines": [
                    {
                        "name": "AWS storage",
                        "provider": "aws",
                        "role": "codebacked_provider",
                        "monthly_cost_usd": 1,
                    }
                ],
                "recommendation": {
                    "summary": "Use this starter model.",
                    "next_steps": ["Replace inferred storage with telemetry."],
                },
            },
            "estimated_monthly_cost_usd": 1,
        }

        output = io.StringIO()
        with tempfile.TemporaryDirectory() as temp_raw:
            assessment = Path(temp_raw) / "assessment.json"
            assessment.write_text(json.dumps(artifact), encoding="utf-8")
            with mock.patch.dict(os.environ, {}, clear=True):
                with mock.patch.object(
                    HELPER,
                    "resolve_profitctl",
                    side_effect=AssertionError("offline evidence must not resolve a binary"),
                ):
                    with contextlib.redirect_stdout(output):
                        exit_code = HELPER.main(["--assessment", str(assessment)])

        payload = json.loads(output.getvalue())
        compact = payload["assessment"]
        self.assertEqual(exit_code, 0)
        self.assertEqual(compact["evidence_mode"], "saved_assessment")
        self.assertEqual(compact["facts"]["providers"][0]["provider"], "aws")
        self.assertEqual(compact["sourced_rates"][0]["source_url"], "https://aws.amazon.com/pricing")
        self.assertEqual(compact["inferred_scale"][0]["name"], "storage_gb")
        self.assertEqual(compact["economics"]["estimated_monthly_cost_usd"], 1)
        self.assertEqual(compact["next_action"]["summary"], "Use this starter model.")


if __name__ == "__main__":
    unittest.main()
