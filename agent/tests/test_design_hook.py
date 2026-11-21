import contextlib
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import types
import unittest
from unittest.mock import Mock, patch

from profitctl_agent import design_hook as hook


QUESTION = "Should we separate Condere's report assembly stage from the main service with a queue?"
PROPOSAL = "I recommend splitting Condere report assembly into a worker behind a queue."


class NativeHook(unittest.TestCase):
    def setUp(self):
        directory = self.enterContext(tempfile.TemporaryDirectory())
        self.directory = Path(directory).resolve()
        self.checkout = self.directory / "project"
        self.checkout.mkdir()
        self.socket = self.directory / "service.sock"
        self.routes = self.directory / "routes.json"
        self.route = {"profile": "condere-queue", "project": "Condere", "roots": [str(self.checkout)],
                      "topics": ["queue", "message broker"], "workloads": ["report assembly", "assemble reports"]}
        self.write_routes([self.route])
        service = types.ModuleType("profitctl_agent.design_service")
        service.request = self.request = Mock(return_value={"summary": "Modeled estimate USD 18 monthly."})
        self.enterContext(patch.dict(sys.modules, {service.__name__: service}))
        self.enterContext(contextlib.chdir(self.checkout))

    def write_routes(self, routes):
        self.routes.write_text(json.dumps({"version": 1, "routes": routes}))
        self.routes.chmod(0o600)

    def event(self, event="UserPromptSubmit", text=QUESTION, **fields):
        payload = {"hook_event_name": event, "cwd": str(self.checkout), "transcript_path": "/never/read",
                   "session_id": "unused", **fields}
        if event == "Stop":
            payload.setdefault("stop_hook_active", False)
            payload["last_assistant_message"] = text
        else:
            payload["prompt"] = text
        return hook.handle(payload, self.socket, self.routes)

    def test_scoped_native_prompt_and_response_content(self):
        extra = self.event(text=QUESTION + " Private unrelated material 79_xyz.")["hookSpecificOutput"]
        self.assertEqual(extra["hookEventName"], "UserPromptSubmit")
        self.assertIn("USD 18", extra["additionalContext"])
        self.assertIn("least-privilege", extra["additionalContext"])
        self.assertNotIn("79_xyz", str(extra))
        self.request.assert_called_once_with(self.socket, {"action": "review", "profile": "condere-queue"}, timeout=1.0)

    def test_assistant_originated_proposal_and_one_stop_correction(self):
        result = self.event("Stop", PROPOSAL)
        self.assertEqual(result["decision"], "block")
        self.assertIn("USD 18", result["reason"])
        self.assertIsNone(self.event(text=result["reason"]))
        self.assertIsNone(self.event("Stop", PROPOSAL, stop_hook_active=True))
        self.assertIsNone(self.event("Stop", PROPOSAL, stop_hook_active="false"))
        self.assertIsNone(self.event("Stop", PROPOSAL + " Cost is unmeasured. Duplicate retries need idempotency. "
                                     "Enforce least privilege for queue readers."))
        self.request.assert_called_once()

    def test_generic_proposals_do_not_use_prepared_prices(self):
        for text in ["Add two database replicas for the east region.",
                     "Move audit events into a dedicated data store.",
                     "Separate the export service and add a queue.",
                     "Condere should migrate the store to a second region."]:
            with self.subTest(text=text):
                context = self.event(text=text)["hookSpecificOutput"]["additionalContext"]
                for term in ("profitctl_review_design", "Security+", "missing measurements", "failure"):
                    self.assertIn(term, context)
                self.assertNotIn("USD 18", context)
        self.request.assert_not_called()

    def test_named_project_requires_both_the_workload_and_the_proposed_change(self):
        for proposal in ["Condere should add a queue for billing emails.",
                         "Deploy a Redis cache for Condere report assembly.",
                         "Add a Redis cache to the Condere report assembly queue.",
                         "Condere should replace the report assembly queue with a Redis cache.",
                         "Use a cache for Condere report assembly instead of a queue.",
                         "Add a queue for Condere billing emails. Deploy a cache for report assembly."]:
            with self.subTest(proposal=proposal):
                result = self.event(text=proposal)["hookSpecificOutput"]["additionalContext"]
                self.assertIn("inline", result)
                self.assertNotIn("USD", result)
        legacy = {k: v for k, v in self.route.items() if k != "workloads"}
        self.write_routes([legacy])
        self.assertIn("inline", self.event()["hookSpecificOutput"]["additionalContext"])
        self.request.assert_not_called()

    def test_dedicated_directory_can_identify_a_project_unless_other_project_is_named(self):
        natural = "Should we put report assembly behind a queue?"
        self.assertIn("inline", self.event(text=natural)["hookSpecificOutput"]["additionalContext"])
        self.write_routes([{**self.route, "infer_project": True}])
        self.assertIn("USD 18", self.event(text=natural)["hookSpecificOutput"]["additionalContext"])
        for prompt in ["For OtherApp, put report assembly behind a queue.",
                       "Move OtherApp's report assembly behind a queue.",
                       "We fixed Condere yesterday. For OtherApp, put report assembly behind a queue."]:
            self.assertIn("inline", self.event(text=prompt)["hookSpecificOutput"]["additionalContext"])
        self.write_routes([{**self.route, "infer_project": True},
                           {**self.route, "project": "OtherApp", "profile": "other-queue"}])
        self.assertIn("inline", self.event(text=natural)["hookSpecificOutput"]["additionalContext"])
        self.request.assert_called_once()

    def test_wrong_project_directory_scope_multiple_projects_and_prompt_paths(self):
        other = self.directory / "other"
        other.mkdir()
        for cwd in [str(other), str(self.checkout) + "-copy", str(self.checkout / ".." / "project"), ".", None]:
            with self.subTest(cwd=cwd):
                self.assertIn("inline", self.event(cwd=cwd)["hookSpecificOutput"]["additionalContext"])
        with contextlib.chdir(other):
            self.assertIn("inline", self.event()["hookSpecificOutput"]["additionalContext"])
        self.write_routes([self.route, {**self.route, "project": "Second", "profile": "second-queue"}])
        self.assertIn("inline", self.event(text=QUESTION + " Compare Second too.")["hookSpecificOutput"]["additionalContext"])
        self.write_routes([self.route, {**self.route, "profile": "ambiguous-other-report"}])
        self.assertIn("inline", self.event()["hookSpecificOutput"]["additionalContext"])
        self.request.assert_not_called()

    def test_unrelated_transcript_and_unknown_event_silent(self):
        for text in ("Show today's calendar.", "Please add a test for the queue.", "List the running services.",
                     "Explain the difference between a queue and a database.", "Add a song to my queue.",
                     "Use the grocery store near home.", "", None):
            self.assertIsNone(self.event(text=text))
        target = self.checkout / "transcript.json"
        target.write_text(PROPOSAL)
        self.assertIsNone(self.event("Stop", None, transcript_path=str(target)))
        self.assertIsNone(self.event("SubagentStop", PROPOSAL))
        self.request.assert_not_called()

    def test_outage_and_untrusted_routes_use_generic_help(self):
        self.request.side_effect = OSError("private error 77_x")
        context = self.event()["hookSpecificOutput"]["additionalContext"]
        self.assertIn("unavailable", context)
        self.assertNotIn("77_x", context)
        self.request.reset_mock()
        self.routes.chmod(0o666)
        self.assertIn("inline", self.event()["hookSpecificOutput"]["additionalContext"])
        self.routes.unlink()
        self.routes.symlink_to(self.checkout/"routes.json")
        (self.checkout/"routes.json").write_text(json.dumps({"version": 1, "routes": [self.route]}))
        self.assertIn("inline", self.event()["hookSpecificOutput"]["additionalContext"])
        self.request.assert_not_called()

    def test_cli_silence_for_malformed_or_oversize_input_and_bounded_summary(self):
        module = str(Path(hook.__file__).resolve())
        for raw in (b"{bad", b"[" * 3500, b"x" * (hook.MAX_INPUT+1),
                    json.dumps({"hook_event_name": "UserPromptSubmit", "prompt": "How are you?"}).encode()):
            got = subprocess.run([sys.executable, module, "--socket", str(self.socket), "--routes", str(self.routes)],
                                 input=raw, capture_output=True, check=True, timeout=4)
            self.assertEqual((got.stdout, got.stderr), (b"", b""))
        self.request.return_value = {"summary": "𝄞" * 10000 + "secret-tail"}
        result = self.event()
        self.assertLess(len(json.dumps(result, ensure_ascii=False).encode()), 8192)
        self.assertNotIn("secret-tail", str(result))
