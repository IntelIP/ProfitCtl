"""Bounded access to Go's economics calculator and immutable decision revisions."""

from contextlib import ExitStack
from copy import deepcopy
import hashlib
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import stat
import subprocess
from tempfile import TemporaryDirectory
from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, StringConstraints


class EconomicsError(ValueError):
    """The requested case, evidence, or calculation could not be used safely."""


class CaseNotFound(EconomicsError):
    """The requested Go case or revision is unavailable."""


class RevisionConflict(EconomicsError):
    """Another writer has already published different content at this revision."""


class _StrictInput(BaseModel):
    model_config = ConfigDict(extra="forbid", strict=True, allow_inf_nan=False)


_Label = Annotated[str, StringConstraints(strip_whitespace=True, min_length=1,
                                         max_length=512, pattern=r"^[^\x00-\x1f\x7f]+$")]
_Count = Annotated[int, Field(ge=0, le=10**12)]
_Nonnegative = Annotated[float, Field(ge=0)]


class ChargeUpdate(_StrictInput):
    id: _Label = Field(description="Existing charge ID from the selected option.")
    usd: _Nonnegative = Field(description="Replacement cohort total for that charge in USD.")
    basis: Literal["invoice", "reported", "modeled"]


class OptionUpdate(_StrictInput):
    option: str = Field(min_length=1, max_length=2048, description="Exact option name returned by the cohort calculator.")
    attempts: _Count | None = None
    accepted: _Count | None = None
    review_minutes: _Nonnegative | None = Field(default=None, description="Total review minutes for this cohort.")
    costs: list[ChargeUpdate] = Field(default_factory=list, max_length=128)


class OwnerEvidence(_StrictInput):
    """A named owner-supplied measurement or assumption for the original cohort."""

    kind: Literal["observed", "assumed"]
    source: _Label = Field(description="Named run, review record, invoice, or stated scenario assumption.")
    updates: list[OptionUpdate] = Field(min_length=1, max_length=4)


_MAX_INPUT = 2 << 20
_ARTIFACTS = "profitctl-investigation-inputs"


def _go(workspace, binary, args, *, stdin=None, outcomes=False):
    executable = shutil.which(str(binary))
    if not executable:
        raise EconomicsError("The configured ProfitCtl executable is unavailable.")
    try:
        run = subprocess.run([str(Path(executable).resolve()), *args], cwd=workspace,
                             input=stdin, capture_output=True, timeout=30, check=False,
                             env={"PATH": os.defpath, "LANG": "C"})
    except (OSError, subprocess.TimeoutExpired) as exc:
        raise EconomicsError("The local ProfitCtl command could not complete.") from exc
    if run.returncode == 4:
        raise RevisionConflict("The case changed after this investigation began. Reconcile with the latest revision.")
    if run.returncode == 1 and args[:2] == ["decision", "read"]:
        raise CaseNotFound("No saved decision matches this case ID and revision.")
    if run.returncode != 0 and not (outcomes and run.returncode == 1):
        raise EconomicsError(run.stderr.decode(errors="replace").strip()[:600] or "ProfitCtl rejected the operation.")
    try:
        value = json.loads(run.stdout)
    except (ValueError, UnicodeDecodeError) as exc:
        raise EconomicsError("ProfitCtl returned an unreadable result.") from exc
    if outcomes and (value.get("schema_version") != "profitctl.outcome-comparison/v1"
                     or value.get("status") != ("incomplete" if run.returncode else "complete")):
        raise EconomicsError("ProfitCtl returned an unexpected outcome comparison.")
    return value


def _case_id(value):
    if not isinstance(value, str) or not re.fullmatch(r"[a-z0-9][a-z0-9-]{0,63}", value):
        raise EconomicsError("Use the lowercase case ID of a saved ProfitCtl decision.")
    return value


def _regular_bytes(parent_fd, name):
    flags = os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK
    with os.fdopen(os.open(name, flags, dir_fd=parent_fd), "rb") as source:
        info = os.fstat(source.fileno())
        if not stat.S_ISREG(info.st_mode) or info.st_size > _MAX_INPUT:
            raise EconomicsError("A calculation input must be a regular file of at most 2 MiB.")
        data = source.read(_MAX_INPUT + 1)
        if len(data) > _MAX_INPUT:
            raise EconomicsError("The calculation input exceeds 2 MiB.")
        return data


def _pinned_bytes(workspace, pin):
    if not hasattr(os, "O_NOFOLLOW") or os.open not in os.supports_dir_fd:
        raise EconomicsError("Run the standalone investigation on macOS or Linux. Its local input store requires "
                             "no-follow directory descriptors. Existing ProfitCtl Go scenario commands remain available.")
    parts = pin["path"].split("/")
    if any(not part or part.startswith((".", "-")) or "\\" in part or ":" in part for part in parts):
        raise EconomicsError("The case must pin a nonhidden workspace-relative input.")
    try:
        with ExitStack() as files:
            flags = os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW
            folder = os.open(workspace, flags)
            files.callback(os.close, folder)
            for part in parts[:-1]:
                folder = os.open(part, flags, dir_fd=folder)
                files.callback(os.close, folder)
            data = _regular_bytes(folder, parts[-1])
    except OSError as exc:
        raise EconomicsError("A pinned input is missing, unreadable, or redirects outside its saved location.") from exc
    if hashlib.sha256(data).hexdigest() != pin["sha256"]:
        raise EconomicsError("A calculation input changed after its digest was recorded.")
    return data


def _calculate(workspace, binary, data):
    # Go reads the exact verified bytes even when a workspace input changes concurrently.
    with TemporaryDirectory(prefix="profitctl-cohort-") as tmp:
        path = Path(tmp) / "cohort.json"
        path.write_bytes(data)
        return _go(workspace, binary, ["system", "outcomes", str(path)], outcomes=True)


def _read_record(workspace, binary, case_id, revision):
    return _go(workspace, binary, ["decision", "read", _case_id(case_id),
                                   "--workspace", ".", "--revision", str(revision)])


def verify_current_baseline(workspace, profitctl_binary, case_id, baseline):
    """Invalidate stale owner input before releasing the interrupt or calling a model."""
    if _read_record(workspace, profitctl_binary, case_id, 0) != baseline["record"]:
        raise RevisionConflict("The case changed while awaiting evidence. Review the latest revision.")
    for pin in baseline["calculation"]["inputs"]:
        _pinned_bytes(workspace, pin)


def load_baseline(workspace, profitctl_binary, case_id, revision=0):
    """Read a saved revision and replay its most recent pinned outcome comparison."""
    record = _read_record(workspace, profitctl_binary, case_id, revision)
    calculations = record["case"].get("calculations", [])
    for index in range(len(calculations) - 1, -1, -1):
        calc = calculations[index]
        args = calc["args"]
        if args[:1] == ["profitctl"]:
            args = args[1:]
        if len(args) == 3 and args[:2] == ["system", "outcomes"]:
            inputs = {p["path"]: _pinned_bytes(workspace, p) for p in calc["inputs"]}
            data = inputs[args[2]]
            report = _calculate(workspace, profitctl_binary, data)
            return {"record": record, "calculation": calc, "calculation_index": index,
                    "input": json.loads(data), "report": report}
    raise EconomicsError("This case needs a pinned 'system outcomes FILE' calculation before investigation.")


def _updated_cohort(baseline, evidence):
    cohort = deepcopy(baseline["input"])
    options = {opt["name"]: opt for opt in cohort["options"]}
    seen = set()
    for update in evidence.updates:
        if update.option in seen or update.option not in options:
            raise EconomicsError("Each update must name a distinct option in the original cohort.")
        seen.add(update.option)
        option = options[update.option]
        if update.attempts is None and update.accepted is None and update.review_minutes is None and not update.costs:
            raise EconomicsError("Supply a count, a review duration, or an existing charge to update.")
        counts = {name: getattr(update, name) for name in ("accepted", "attempts") if getattr(update, name) is not None}
        if counts:
            option.update(counts)
            if option.get("attempts") is not None and option.get("accepted") is not None and option["accepted"] > option["attempts"]:
                raise EconomicsError("The accepted count cannot exceed the attempt count for this cohort.")
            fields = ", ".join(counts)
            option["count_source"] = f"{fields} ({evidence.kind}): {evidence.source}."
            if len(counts) == 1:
                original = baseline["record"]
                option["count_source"] += f" Other count retained from {original['id']} revision {original['revision']}."
        charges = {c["id"]: c for c in option["costs"]}
        replaced = set()
        for charge in update.costs:
            if charge.id not in charges or charge.id in replaced:
                raise EconomicsError("Name each replaced charge once using an ID in the original option.")
            if evidence.kind == "assumed" and charge.basis != "modeled":
                raise EconomicsError("An assumed charge must retain the modeled basis.")
            replaced.add(charge.id)
            charges[charge.id].update(usd=charge.usd, basis=charge.basis, source=evidence.source)
    return cohort


def validate_evidence(baseline, evidence):
    """Check the owner's input against the original options without writing or running a model."""
    _updated_cohort(baseline, OwnerEvidence.model_validate(evidence))


def _publish_bytes(folder, name, data):
    temp = ".pending-" + secrets.token_hex(12)
    fd = os.open(temp, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600, dir_fd=folder)
    try:
        with os.fdopen(fd, "wb") as f:
            f.write(data)
            f.flush()
            os.fsync(f.fileno())
        try:
            os.link(temp, name, src_dir_fd=folder, dst_dir_fd=folder)
        except FileExistsError:
            if _regular_bytes(folder, name) != data:
                raise EconomicsError("An existing investigation artifact has unexpected content.")
        os.fsync(folder)
    finally:
        os.unlink(temp, dir_fd=folder)


def _save_input(workspace, data):
    digest = hashlib.sha256(data).hexdigest()
    flags = os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW
    with ExitStack() as files:
        root = os.open(workspace, flags)
        files.callback(os.close, root)
        try:
            os.mkdir(_ARTIFACTS, 0o700, dir_fd=root)
            os.fsync(root)
        except FileExistsError:
            pass
        folder = os.open(_ARTIFACTS, flags, dir_fd=root)
        files.callback(os.close, folder)
        if os.fstat(folder).st_mode & 0o077:
            raise EconomicsError("The investigation input directory must be owner-only.")
        _publish_bytes(folder, ".gitignore", b"*\n")
        _publish_bytes(folder, digest + ".json", data)
    path = f"{_ARTIFACTS}/{digest}.json"
    return {"args": ["system", "outcomes", path], "inputs": [{"path": path, "sha256": digest}]}


def _conclusion(report, previous, evidence_kind):
    options = report["options"]
    lines = []
    scenario = evidence_kind == "assumed" or any(c["basis"] == "modeled" for o in options for c in o["costs"])
    basis = "Under the stated assumptions and scoped costs" if scenario else "Within the stated cohort and cost scope"
    priced = [o for o in options if o["usd_per_accepted"] is not None]
    for opt in options:
        name = opt["name"]
        cost, rate = opt["usd_per_accepted"], opt["break_even_acceptance_rate"]
        if cost is not None:
            lines.append(f"{name}: USD {cost:.4g} per accepted outcome.")
        else:
            lines.append(f"{name}: cost per accepted outcome remains unknown.")
        if rate is not None and rate > 1:
            lines.append(f"{name} cannot match the baseline's unit cost even at full acceptance "
                         f"(break-even would require {rate:.2%}).")
        elif rate is not None:
            lines.append(f"{name} needs {rate:.2%} acceptance to match the baseline at its stated cost per attempt.")
    if len(priced) == len(options) and len(options) > 1:
        best = min(o["usd_per_accepted"] for o in priced)
        names = ", ".join(o["name"] for o in priced if o["usd_per_accepted"] == best)
        recommendation = f"{basis}, {names} has the lowest calculated cost per accepted outcome. Confirm comparable quality before acting."
    else:
        recommendation = previous or "Collect the remaining evidence before choosing an option."
    return recommendation, " ".join(lines)


def _evidence_entries(evidence, run_id, revision):
    # Counts and costs can come from the same file with different provenance.
    groups = {}
    for update in evidence.updates:
        fields = update.model_dump(exclude_none=True, exclude={"option", "costs"})
        if fields:
            groups.setdefault(evidence.kind, []).append({"option": update.option, **fields})
        for charge in update.costs:
            kind = "modeled" if charge.basis == "modeled" else "observed"
            groups.setdefault(kind, []).append({"option": update.option, "charge": charge.model_dump()})
    return [{"kind": kind, "source": evidence.source,
             "finding": f"Investigation {run_id} updates case revision {revision} with owner-supplied inputs: "
                        + json.dumps(items, ensure_ascii=False, separators=(",", ":"))}
            for kind, items in groups.items()]


def prepare_revision(workspace, profitctl_binary, baseline, run_id, evidence):
    """Calculate an approved input update and prepare a replayable case. The decision write is separate."""
    evidence = OwnerEvidence.model_validate(evidence)
    cohort = _updated_cohort(baseline, evidence)
    for pin in baseline["calculation"]["inputs"]:
        _pinned_bytes(workspace, pin)
    data = (json.dumps(cohort, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode()
    report = _calculate(workspace, profitctl_binary, data)
    try:
        calc = _save_input(workspace, data)
    except OSError as exc:
        raise EconomicsError("Could not safely store the revised input in the selected workspace.") from exc
    record = baseline["record"]
    case = deepcopy(record["case"])
    case["calculations"][baseline["calculation_index"]] = calc
    case["evidence"].extend(_evidence_entries(evidence, run_id, record["revision"]))
    recommendation, figures = _conclusion(report, case.get("recommendation"), evidence.kind)
    case["evidence"].append({"kind": "modeled", "source": calc["args"][-1],
                             "finding": f"Calculated with ProfitCtl from the pinned cohort. {figures}"})
    case["recommendation"] = recommendation
    case.setdefault("reversal_condition", "Reassess if comparable quality, included cost, or acceptance changes.")
    missing = []
    for opt in report["options"]:
        missing.extend(f"{opt['name']}: {gap}" for gap in opt["limitations"])
    # The forecast's qualitative questions can survive a numerical update.
    # Keep the original wording as an explicit prior request, separately from present unknowns.
    previous = [item.removeprefix("Prior case request: ") for item in case.get("missing", [])
                if not item.startswith("Current cohort: ")]
    case["missing"] = ["Prior case request: " + s for s in previous]
    if missing:
        current = "Current cohort: " + " ".join(missing)
        if len(case["missing"]) < 32:
            case["missing"].insert(0, current)
        else:
            case["evidence"].append({"kind": "unknown", "source": calc["args"][-1], "finding": current})
    return {"case": case, "calculation": calc, "report": report}


def save_revision(workspace, profitctl_binary, case_id, expected_revision, case):
    """Publish with Go's immutable compare-and-set operation. A lost acknowledgement can be retried."""
    return _go(workspace, profitctl_binary, ["decision", "revise", _case_id(case_id),
               "--workspace", ".", "--expected", str(expected_revision), "--input", "-"],
               stdin=json.dumps(case, allow_nan=False).encode())
