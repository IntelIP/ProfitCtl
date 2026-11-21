"""Exercise the actual CLI and companion from a locally built release archive."""

import hashlib
import json
from pathlib import Path
import platform
import subprocess
import sys
import tarfile
import tempfile
import zipfile


def smoke(release_dir, tag):
    os_name = platform.system().lower()
    arch = {"aarch64": "arm64", "arm64": "arm64", "amd64": "amd64", "x86_64": "amd64"}[
        platform.machine().lower()
    ]
    stem = f"profitctl_{tag}_{os_name}_{arch}"
    windows = os_name == "windows"
    archive = release_dir / (stem + (".zip" if windows else ".tar.gz"))
    checksums = {
        name: digest
        for digest, name in (line.split() for line in (release_dir / "SHA256SUMS").read_text().splitlines())
    }
    assert hashlib.sha256(archive.read_bytes()).hexdigest() == checksums[archive.name]
    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        binaries = []
        for base in ("profitctl", "profitctl-standards"):
            name = base + (".exe" if windows else "")
            member = f"{stem}/{name}"
            if windows:
                with zipfile.ZipFile(archive) as package:
                    data = package.read(member)
            else:
                with tarfile.open(archive) as package:
                    data = package.extractfile(member).read()
            binary = root / name
            binary.write_bytes(data)
            binary.chmod(0o700)
            binaries.append(binary)

        cli, standards = binaries

        def run(*args, code=0, raw=False):
            proc = subprocess.run([str(cli), *args], capture_output=True, text=True, timeout=30)
            assert proc.returncode == code, (args, proc.returncode, proc.stderr)
            return proc.stdout.strip() if raw else json.loads(proc.stdout)

        assert run("--version", raw=True) == f"profitctl {tag}"
        margin = run("simulate", "-f", "test/fixtures/monthly_margin_regression.yml", "--json")
        assert margin["margin"]["gross"] == 80 and margin["revenue"]["total"] == 500
        folder = Path("examples/system")
        for option, total in (("cheaper", 141.5), ("adverse", 276.5)):
            report = run("system", "compare", folder / "condere-synthetic-baseline.json",
                         folder / f"condere-synthetic-{option}.json")
            assert (report["baseline"]["total_cost_usd"], report["proposed"]["total_cost_usd"]) == (191.5, total)
        partial = run("system", "evaluate", folder / "condere-production.json", code=1)
        assert partial["total_cost_usd"] is None and partial["missing_inputs"]
        receipts = run("system", "receipts", folder / "condere-sanitized-receipts.json", code=1)
        assert receipts["reported_cost_usd"] == 0.02 and receipts["total_cost_usd"] is None
        cohorts = run("system", "outcomes", folder / "accepted-outcomes.json", code=1)
        assert cohorts["options"][1]["usd_per_accepted"] is None
        assert abs(cohorts["options"][1]["break_even_acceptance_rate"] - 0.8) < 1e-12
        judged = subprocess.check_output([standards, "test/fixtures/agent_recommendation_valid.md"], timeout=30)
        assert json.loads(judged)["passed"]

        workspace = root / "cases"
        workspace.mkdir(mode=0o700)
        fixture = Path("examples/decisions/forecast.json")
        created = run("decision", "create", "airport-cache-demo", "--workspace", workspace, "--input", fixture)
        assert created == run("decision", "read", "airport-cache-demo", "--workspace", workspace)
        assert created["revision"] == 1
    print(f"{stem}: verified archive digest, native CLI, companion and saved decision")


if __name__ == "__main__":
    smoke(Path(sys.argv[1]), sys.argv[2])
