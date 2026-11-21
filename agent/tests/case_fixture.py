import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
from tempfile import TemporaryDirectory
import unittest

from profitctl_agent.economics import load_baseline


REPO = Path(__file__).resolve().parents[2]


class EconomicCaseFixture(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        tmp = TemporaryDirectory()
        cls.addClassCleanup(tmp.cleanup)
        if configured := os.getenv("PROFITCTL_TEST_BINARY"):
            cls.binary = Path(configured).resolve()
        else:
            cls.binary = Path(tmp.name) / "profitctl"
            subprocess.run(["go", "build", "-o", cls.binary, "."], cwd=REPO, check=True)

    def setUp(self):
        workspace = TemporaryDirectory()
        self.addCleanup(workspace.cleanup)
        self.root = Path(workspace.name)
        self.source = self.root / "cohort.json"
        shutil.copyfile(REPO / "examples/system/accepted-outcomes.json", self.source)
        digest = hashlib.sha256(self.source.read_bytes()).hexdigest()
        case = {
            "project": "Illustrative research", "question": "Which depth?",
            "workload": "Research assessments", "window": "Matched cohort of ten requests",
            "source_revision": "Public teaching fixture v1",
            "alternatives": [{"name": s, "cost": "Fixture", "quality": "Same acceptance rule"}
                             for s in ["Selective research", "Research every request"]],
            "evidence": [{"kind": "assumed", "source": "Public fixture", "finding": "Synthetic inputs."}],
            "recommendation": "Keep selective research pending evidence.",
            "reversal_condition": "Every request clears the threshold with equivalent quality.",
            "missing": ["Confirm measured customer acceptance."],
            "calculations": [{"args": ["system", "outcomes", "cohort.json"],
                              "inputs": [{"path": "cohort.json", "sha256": digest}]}],
        }
        result = subprocess.run([self.binary, "decision", "create", "research", "--input", "-"],
                                cwd=self.root, input=json.dumps(case), text=True, capture_output=True, check=True)
        self.first = json.loads(result.stdout)
        self.baseline = load_baseline(self.root, self.binary, "research")
