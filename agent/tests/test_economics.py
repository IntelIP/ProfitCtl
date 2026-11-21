import hashlib
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest

from case_fixture import EconomicCaseFixture
from profitctl_agent.economics import (
    EconomicsError, RevisionConflict, load_baseline, prepare_revision, save_revision, validate_evidence,
)


class EconomicsTest(EconomicCaseFixture):
    def test_partial_calculation_revises_with_its_actual_input_and_idempotent_write(self):
        current = self.baseline["report"]["options"]
        self.assertEqual(current[0]["usd_per_accepted"], 7)
        self.assertIsNone(current[1]["usd_per_accepted"])
        self.assertAlmostEqual(current[1]["break_even_acceptance_rate"], 0.8)
        evidence = {"kind": "assumed", "source": "Public sensitivity fixture",
                    "updates": [{"option": current[1]["name"], "accepted": 9, "review_minutes": 15}]}
        change = prepare_revision(self.root, self.binary, self.baseline, "run-1", evidence)
        self.assertAlmostEqual(change["report"]["options"][1]["usd_per_accepted"], 56/9)
        self.assertIn("Research every request", change["case"]["recommendation"])
        self.assertNotIn("choice", change["case"])
        self.assertEqual(change["case"]["evidence"][-2]["kind"], "assumed")
        self.assertIn("review_minutes", change["case"]["evidence"][-2]["finding"])
        self.assertTrue(all(c["basis"] == "modeled" for o in change["report"]["options"] for c in o["costs"]))
        pin = change["calculation"]["inputs"][0]
        artifact = self.root / pin["path"]
        self.assertEqual(hashlib.sha256(artifact.read_bytes()).hexdigest(), pin["sha256"])
        self.assertEqual(artifact.stat().st_mode & 0o777, 0o600)
        self.assertEqual((artifact.parent / ".gitignore").read_text(), "*\n")
        written = save_revision(self.root, self.binary, "research", 1, change["case"])
        # Go already wrote revision 2; the calling process lost the acknowledgement.
        retry = prepare_revision(self.root, self.binary, self.baseline, "run-1", evidence)
        self.assertEqual(save_revision(self.root, self.binary, "research", 1, retry["case"]), written)
        self.assertEqual(load_baseline(self.root, self.binary, "research", 1)["record"], self.first)
        self.assertEqual(load_baseline(self.root, self.binary, "research")["report"], change["report"])
        conflicting = dict(change["case"], recommendation="A different revision.")
        with self.assertRaises(RevisionConflict):
            save_revision(self.root, self.binary, "research", 1, conflicting)

    def test_evidence_type_counts_scope_and_single_charge_replacement(self):
        name = self.baseline["report"]["options"][1]["name"]
        owner = {"kind": "observed", "source": "Named receipt for the scoped cohort",
                 "updates": [{"option": name, "costs": [{"id": "delivery", "usd": 50, "basis": "invoice"}]}]}
        changed = prepare_revision(self.root, self.binary, self.baseline, "one", owner)
        candidate = changed["report"]["options"][1]
        self.assertEqual(candidate["known_subtotal_usd"], 62)
        self.assertAlmostEqual(candidate["break_even_acceptance_rate"], 6.2/7)
        self.assertIsNone(candidate["usd_per_accepted"])
        self.assertEqual([c["basis"] for c in candidate["costs"]], ["invoice", "modeled"])
        mixed = dict(owner, updates=[{"option": name, "accepted": 8,
                                      "costs": [{"id": "review", "usd": 18, "basis": "modeled"}]}])
        changes = prepare_revision(self.root, self.binary, self.baseline, "mixed", mixed)
        entries = [e for e in changes["case"]["evidence"] if e["source"] == owner["source"]]
        self.assertEqual([e["kind"] for e in entries], ["observed", "modeled"])
        self.assertNotIn('"usd":18', entries[0]["finding"])
        with self.assertRaisesRegex(EconomicsError, "assumed charge"):
            validate_evidence(self.baseline, dict(owner, kind="assumed"))
        for invalid in [
            {"option": name, "accepted": 11}, {"option": "Another period", "accepted": 1},
            {"option": name, "costs": [{"id": "new", "basis": "invoice", "usd": 1}]},
            {"option": name, "costs": [{"id": "review", "basis": "modeled", "usd": True}]},
            {"option": name, "review_minutes": -1}, {"option": name},
        ]:
            with self.subTest(invalid=invalid), self.assertRaises(ValueError):
                validate_evidence(self.baseline, dict(owner, updates=[invalid]))

    def test_update_replaces_the_selected_copy_of_an_outcome_calculation(self):
        case = self.first["case"]
        case["calculations"] *= 2
        save_revision(self.root, self.binary, "research", 1, case)
        current = load_baseline(self.root, self.binary, "research")
        self.assertEqual(current["calculation_index"], 1)
        evidence = {"kind": "assumed", "source": "Scenario", "updates": [
            {"option": "Research every request", "accepted": 9}]}
        change = prepare_revision(self.root, self.binary, current, "duplicate", evidence)
        save_revision(self.root, self.binary, "research", 2, change["case"])
        self.assertEqual(load_baseline(self.root, self.binary, "research")["report"], change["report"])

    def test_input_must_still_match_even_if_a_symlink_has_identical_bytes(self):
        original = self.source.read_bytes()
        self.source.write_bytes(original + b"\n")
        with self.assertRaisesRegex(EconomicsError, "digest"):
            load_baseline(self.root, self.binary, "research")
        with TemporaryDirectory() as outside:
            target = Path(outside) / "cohort.json"
            target.write_bytes(original)
            self.source.unlink()
            self.source.symlink_to(target)
            with self.assertRaisesRegex(EconomicsError, "redirects"):
                load_baseline(self.root, self.binary, "research")


if __name__ == "__main__":
    unittest.main()
