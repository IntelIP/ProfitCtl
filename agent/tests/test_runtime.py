import json
import os
import sqlite3
import subprocess
import sys
from unittest.mock import patch

from case_fixture import EconomicCaseFixture
from profitctl_agent.economics import EconomicsError, RevisionConflict, load_baseline, save_revision
from profitctl_agent.investigator import READ_TOOLS
from profitctl_agent.runtime import InvestigationError, InvestigationService


class RecoveryTest(EconomicCaseFixture):
    def setUp(self):
        super().setUp()
        env = patch.dict(os.environ, {"PROFITCTL_AGENT_MODEL": "scripted"})
        env.start()
        self.addCleanup(env.stop)
        self.service = InvestigationService(self.root, profitctl_binary=self.binary)
        self.start_args = {"run_id": "condere", "case_id": "research", "question": "Should every case get deep research?"}
        self.answer = {"run_id": "condere", "expected_revision": 1, "action": "supply",
                       "evidence": {"kind": "assumed", "source": "Illustrative acceptance sensitivity",
                                    "updates": [{"option": "Research every request", "accepted": 9}]}}

    def test_process_loss_after_go_publication_before_checkpoint_acknowledgement(self):
        paused = self.service.start(**self.start_args)
        self.assertEqual(paused["status"], "awaiting_evidence")
        self.assertEqual(set(paused["analysis_tools"]), READ_TOOLS)
        self.assertEqual(paused["base_revision"], 1)
        self.assertEqual(paused["baseline_report"]["options"][0]["usd_per_accepted"], 7)
        self.assertAlmostEqual(paused["baseline_report"]["options"][1]["break_even_acceptance_rate"], 0.8)
        self.assertEqual(self.service.start(**self.start_args)["checkpoint_id"], paused["checkpoint_id"])

        # The process exits at the real side-effect boundary. Its SQLite lease
        # disappears along with it, while both the checkpoint and Go record survive.
        child = """
import json, os, sys
from profitctl_agent import economics
from profitctl_agent.runtime import InvestigationService
original = economics.save_revision
def lose_ack(*args):
    original(*args)
    os._exit(72)
economics.save_revision = lose_ack
InvestigationService(sys.argv[1], profitctl_binary=sys.argv[2]).resume(**json.loads(sys.argv[3]))
"""
        process = subprocess.run([sys.executable, "-c", child, str(self.root), str(self.binary),
                                  json.dumps(self.answer)], capture_output=True, text=True, timeout=45)
        self.assertEqual(process.returncode, 72, process.stderr)
        restarted = InvestigationService(self.root, profitctl_binary=self.binary)
        checkpoint = restarted.status(run_id="condere")
        self.assertEqual(checkpoint["status"], "interrupted")
        self.assertEqual(checkpoint["decision_publication"], "unconfirmed")
        self.assertIsNone(checkpoint["decision_revision"])
        before_retry = load_baseline(self.root, self.binary, "research")["record"]
        self.assertEqual(before_retry["revision"], 2)
        done = restarted.resume(**self.answer)
        self.assertEqual((done["status"], done["decision_revision"]), ("completed", 2))
        self.assertAlmostEqual(done["report"]["options"][1]["usd_per_accepted"], 56 / 9)
        self.assertEqual(restarted.resume(**self.answer)["checkpoint_id"], done["checkpoint_id"])
        self.assertEqual(load_baseline(self.root, self.binary, "research")["record"], before_retry)
        self.assertEqual(load_baseline(self.root, self.binary, "research", 1)["record"], self.first)

    def test_reassessment_failure_exposes_saved_input_and_an_exact_recovery_call(self):
        self.service.start(**self.start_args)
        with patch.object(self.service, "_analyze", side_effect=RuntimeError("provider rate limit")):
            with self.assertRaisesRegex(RuntimeError, "rate limit"):
                self.service.resume(**self.answer)
        restarted = InvestigationService(self.root, profitctl_binary=self.binary)
        saved = restarted.status(run_id=self.answer["run_id"])
        self.assertEqual(saved["status"], "interrupted")
        self.assertEqual(saved["pending_steps"], ["reassess"])
        owner = saved["recorded_owner_response"]
        self.assertEqual((owner["action"], owner["expected_revision"]), ("supply", 1))
        self.assertEqual(owner["evidence"]["source"], self.answer["evidence"]["source"])
        self.assertEqual(owner["evidence"]["updates"][0]["accepted"], 9)
        self.assertEqual(saved["decision_publication"], "not_started")
        self.assertIsNone(saved["decision_revision"])
        self.assertEqual(load_baseline(self.root, self.binary, "research")["record"], self.first)
        retry = saved["retry_tool_call"]
        self.assertEqual(retry, {"name": "profitctl_resume_investigation",
                                 "arguments": {"run_id": self.answer["run_id"], **owner}})
        done = restarted.resume(**retry["arguments"])
        self.assertEqual(done["recorded_owner_response"], owner)
        self.assertEqual((done["decision_publication"], done["decision_revision"]), ("published", 2))

    def test_decline_stale_answer_and_busy_writer(self):
        self.service.start(**self.start_args)
        with self.assertRaisesRegex(InvestigationError, "expected case revision"):
            self.service.resume(**dict(self.answer, expected_revision=2))
        with self.assertRaisesRegex(InvestigationError, "another investigation"):
            self.service.start(**dict(self.start_args, question="Different question"))
        run_dir = self.root / ".profitctl" / "investigations" / "condere"
        conn = sqlite3.connect(run_dir / "lease.sqlite3")
        try:
            conn.execute("BEGIN IMMEDIATE")
            with self.assertRaisesRegex(InvestigationError, "Another process"):
                self.service.resume(**self.answer)
            self.assertEqual(self.service.status(run_id="condere")["status"], "awaiting_evidence")
        finally:
            conn.close()
        result = self.service.resume(run_id="condere", action="decline", expected_revision=1)
        self.assertEqual(result["status"], "declined")
        with self.assertRaisesRegex(InvestigationError, "already been recorded"):
            self.service.resume(**self.answer)
        self.assertEqual(load_baseline(self.root, self.binary, "research")["record"], self.first)

    def test_unknown_and_symlinked_run_are_rejected(self):
        with self.assertRaisesRegex(InvestigationError, "No saved investigation"):
            self.service.status(run_id="unknown")
        with self.assertRaisesRegex(InvestigationError, "run or case ID"):
            self.service.status(run_id="../elsewhere")
        store = self.root / ".profitctl" / "investigations"
        store.mkdir(mode=0o700)
        (store / "condere").symlink_to(self.root, target_is_directory=True)
        with self.assertRaisesRegex(InvestigationError, "without symlinks"):
            self.service.start(**self.start_args)

    def test_changed_input_or_forecast_blocks_resume_before_another_model_turn(self):
        self.service.start(**self.start_args)
        original = self.source.read_bytes()
        self.source.write_bytes(original + b"\n")
        with patch.object(self.service, "_analyze", side_effect=AssertionError("model was called")):
            with self.assertRaises(EconomicsError):
                self.service.resume(**self.answer)
            self.source.write_bytes(original)
            updated_case = dict(self.first["case"], recommendation="Await more data from the owner.")
            save_revision(self.root, self.binary, "research", 1, updated_case)
            with self.assertRaises(RevisionConflict):
                self.service.resume(**self.answer)
        self.assertEqual(self.service.status(run_id="condere")["status"], "awaiting_evidence")
