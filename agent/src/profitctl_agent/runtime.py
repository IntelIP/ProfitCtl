"""Durable investigation state, human input, and replay-safe decision publication."""

import os
import re
import sqlite3
import stat
from contextlib import closing, contextmanager
from pathlib import Path
from typing import TypedDict

from langgraph.checkpoint.serde.jsonplus import JsonPlusSerializer
from langgraph.checkpoint.sqlite import SqliteSaver
from langgraph.graph import END, START, StateGraph
from langgraph.types import Command, interrupt
from langsmith import tracing_context

from . import economics
from .investigator import analyze, configured_model


class InvestigationError(ValueError):
    """An actionable investigation request or run-state error."""


class Run(TypedDict, total=False):
    run_id: str
    case_id: str
    question: str
    model_mode: str
    baseline: dict
    analysis: str
    analysis_tools: list
    conversation: list
    owner: dict
    prepared: dict
    decision_revision: int


def _id(value):
    if not isinstance(value, str) or not re.fullmatch(r"[a-z0-9][a-z0-9-]{0,63}", value):
        raise InvestigationError("A run or case ID must be 1 to 64 lowercase letters, numbers or hyphens.")
    return value


def _directory(path, *, create, private=True):
    if create:
        path.mkdir(mode=0o700, exist_ok=True)
    info = path.lstat()
    if not stat.S_ISDIR(info.st_mode):
        raise InvestigationError("The checkpoint path must be a directory without symlinks.")
    if private and os.name == "posix" and (info.st_uid != os.getuid() or info.st_mode & 0o077):
        raise InvestigationError("Investigation directories must be private to the current OS user.")


def _private_file(path, *, create, initial=b""):
    if create:
        try:
            fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        except FileExistsError:
            pass
        else:
            with os.fdopen(fd, "wb") as target:
                target.write(initial)
    info = path.lstat()
    if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1:
        raise InvestigationError("Investigation files must be regular files without links.")
    if os.name == "posix" and (info.st_uid != os.getuid() or info.st_mode & 0o077):
        raise InvestigationError("Investigation files must be private to the current OS user.")


def _pending(state):
    baseline = state["baseline"]
    return {"expected_revision": baseline["record"]["revision"],
            "missing": baseline["record"]["case"].get("missing", []),
            "options": [{"name": entry["name"], "count_source": entry["count_source"],
                         "cost_ids": [charge["id"] for charge in entry["costs"]]} for entry in baseline["input"]["options"]],
            "actions": ["supply", "decline"]}


def _owner_input(state):
    return {"owner": interrupt(_pending(state))}


def _view(snapshot):
    state = snapshot.values
    if not state:
        raise InvestigationError("No saved investigation exists for this run ID.")
    status = "interrupted"
    if "decision_revision" in state:
        status = "completed"
    elif state.get("owner", {}).get("action") == "decline":
        status = "declined"
    elif snapshot.interrupts:
        status = "awaiting_evidence"
    result = {k: state[k] for k in ("run_id", "case_id", "question", "model_mode", "analysis",
                                    "analysis_tools", "decision_revision") if k in state}
    result.update(status=status, checkpoint_id=snapshot.config["configurable"]["checkpoint_id"])
    if "baseline" in state:
        baseline = state["baseline"]
        result.update(base_revision=baseline["record"]["revision"], baseline_report=baseline["report"],
                      calculation=baseline["calculation"])
    if status == "awaiting_evidence":
        result["pending_evidence"] = snapshot.interrupts[0].value
    if status == "completed":
        result.update(report=state["prepared"]["report"], calculation=state["prepared"]["calculation"],
                      result=state["prepared"]["case"]["recommendation"])
    return result


class InvestigationService:
    def __init__(self, workspace, *, profitctl_binary="profitctl", model=None):
        self.workspace = Path(workspace).resolve(strict=True)
        if not self.workspace.is_dir():
            raise InvestigationError("The configured workspace must be a directory.")
        self.binary = profitctl_binary
        self.model = model
        self.mode = (getattr(model, "model_name", None) or type(model).__name__) if model is not None else os.environ.get("PROFITCTL_AGENT_MODEL", "")

    def _analyze(self, state, *, followup):
        if not self.mode or self.mode != state["model_mode"]:
            raise InvestigationError("Use the investigation's originally configured model to run this step.")
        model = self.model if self.model is not None else configured_model(self.mode)
        if followup:
            data = state["prepared"]
            case = data["case"]
            phase = "reassess with the owner's supplied evidence"
        else:
            data = state["baseline"]
            case = data["record"]["case"]
            phase = "inspect the original forecast and identify the next evidence needed"
        return analyze(model, state["question"], case, data["calculation"], data["report"],
                       state.get("conversation", []), phase)

    def _graph(self, saver):
        flow = StateGraph(Run)
        flow.add_node("calculate", lambda s: {"baseline": economics.load_baseline(self.workspace, self.binary, s["case_id"])})
        flow.add_node("investigate", lambda s: self._analyze(s, followup=False))
        flow.add_node("request_evidence", _owner_input)
        flow.add_node("recalculate", lambda s: {"prepared": economics.prepare_revision(
            self.workspace, self.binary, s["baseline"], s["run_id"], s["owner"]["evidence"])})
        flow.add_node("reassess", lambda s: self._analyze(s, followup=True))
        flow.add_node("record_decision", lambda s: {"decision_revision": economics.save_revision(
            self.workspace, self.binary, s["case_id"], s["baseline"]["record"]["revision"], s["prepared"]["case"])["revision"]})
        flow.add_edge(START, "calculate")
        flow.add_edge("calculate", "investigate")
        flow.add_edge("investigate", "request_evidence")
        flow.add_conditional_edges("request_evidence", lambda s: s["owner"]["action"], {"decline": END, "supply": "recalculate"})
        flow.add_edge("recalculate", "reassess")
        flow.add_edge("reassess", "record_decision")
        flow.add_edge("record_decision", END)
        return flow.compile(checkpointer=saver)

    @contextmanager
    def _session(self, run_id, *, create=False, write=False):
        root = self.workspace / ".profitctl"
        store = root / "investigations"
        run = store / _id(run_id)
        db, lock = run / "checkpoints.sqlite3", run / "lease.sqlite3"
        try:
            _directory(root, create=create, private=False)
            _directory(store, create=create)
            if create:
                _private_file(store / ".gitignore", create=True, initial=b"*\n")
            _directory(run, create=create)
            _private_file(db, create=create)
            _private_file(lock, create=create)
        except FileNotFoundError as exc:
            raise InvestigationError("No saved investigation exists for this run ID.") from exc
        # A separate SQLite reservation serializes the entire read/advance operation
        # for this run. The OS releases it when a process is killed.
        with closing(sqlite3.connect(lock, timeout=0)) as lease, closing(sqlite3.connect(db, check_same_thread=False)) as conn:
            if write:
                try:
                    lease.execute("BEGIN IMMEDIATE")
                except sqlite3.OperationalError as exc:
                    if exc.sqlite_errorcode in (sqlite3.SQLITE_BUSY, sqlite3.SQLITE_LOCKED):
                        raise InvestigationError("Another process is advancing this investigation. Retry its status.") from exc
                    raise
            saver = SqliteSaver(conn, serde=JsonPlusSerializer(pickle_fallback=False, allowed_msgpack_modules=()))
            config = {"configurable": {"thread_id": run_id}, "recursion_limit": 20}
            with tracing_context(enabled=False):
                yield self._graph(saver), config

    def start(self, *, run_id, case_id, question):
        _id(case_id)
        if not isinstance(question, str) or not question.strip() or len(question) > 4000:
            raise InvestigationError("Supply an investigation question of up to 4,000 characters.")
        if not self.mode:
            raise InvestigationError("Select PROFITCTL_AGENT_MODEL before starting an investigation.")
        with self._session(run_id, create=True, write=True) as (graph, cfg):
            snapshot = graph.get_state(cfg)
            if snapshot.values:
                if (case_id, question) != (snapshot.values["case_id"], snapshot.values["question"]):
                    raise InvestigationError("This run ID already belongs to another investigation.")
                if snapshot.interrupts or not snapshot.next:
                    return _view(snapshot)
                value = None  # Replay only the outstanding node of a crashed run.
            else:
                value = {"run_id": run_id, "case_id": case_id, "question": question, "model_mode": self.mode}
            graph.invoke(value, cfg, durability="sync")
            return _view(graph.get_state(cfg))

    def status(self, *, run_id):
        with self._session(run_id) as (graph, cfg):
            return _view(graph.get_state(cfg))

    def resume(self, *, run_id, expected_revision, action, evidence=None):
        if type(expected_revision) is not int or expected_revision < 1:
            raise InvestigationError("Supply the positive base revision shown in the evidence request.")
        if action not in ("supply", "decline") or ((action == "decline") != (evidence is None)):
            raise InvestigationError("Supply evidence with action supply, or decline without evidence.")
        if action == "supply":
            evidence = economics.OwnerEvidence.model_validate(evidence).model_dump(mode="json", exclude_none=True)
        owner = {"action": action, "expected_revision": expected_revision, "evidence": evidence}
        with self._session(run_id, write=True) as (graph, cfg):
            snapshot = graph.get_state(cfg)
            state = snapshot.values
            if "baseline" not in state:
                raise InvestigationError("Start or retry this investigation before supplying evidence.")
            if expected_revision != state["baseline"]["record"]["revision"]:
                raise InvestigationError("The expected case revision does not match the investigation's forecast.")
            if state.get("owner") is not None and state["owner"] != owner:
                raise InvestigationError("An owner response has already been recorded for this investigation.")
            if not snapshot.next:
                return _view(snapshot)
            if not snapshot.interrupts and state.get("owner") is None:
                raise InvestigationError("Retry start with the same run ID to reach the pending evidence request.")
            if action == "supply":
                if snapshot.interrupts:
                    economics.verify_current_baseline(self.workspace, self.binary, state["case_id"], state["baseline"])
                economics.validate_evidence(state["baseline"], evidence)
            value = Command(resume=owner) if snapshot.interrupts else None
            graph.invoke(value, cfg, durability="sync")
            return _view(graph.get_state(cfg))
