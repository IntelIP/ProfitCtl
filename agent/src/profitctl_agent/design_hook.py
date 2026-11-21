"""Native Codex and Claude Code design-review hooks. Hook payloads stay in memory."""
import argparse
import json
import os
from pathlib import Path
import re
import signal
import stat
import sys

MAX_INPUT = 64 * 1024
MAX_ROUTES = 32 * 1024
CONTINUATION = "ProfitCtl design review. Revisit the proposed change once. "
GUIDANCE = (
    "Compare the current design with the proposed change for the same workload. "
    "State the load or latency target, the failure and recovery behavior, and who operates it. "
    "Apply Security+ principles to the changed trust boundary. Identify the data exposed, "
    "one credible threat and a least-privilege control. Compare the cost per successful outcome "
    "and identify the demand or reliability threshold that would change the decision."
)
INLINE = (
    "Use profitctl_review_design inline with sourced workload and unit-price assumptions when available. "
    "Otherwise name the missing measurements or rates and how to obtain them. "
    "Only give dollar figures supported by this design's evidence."
)

INTENT = re.compile(
    r"\b(?:add(?:ing)?|adopt(?:ing)?|architect(?:ing)?|build(?:ing)?|compar(?:e|ing)|"
    r"consider(?:ing)?|creat(?:e|ing)|decoupl(?:e|ing)|deploy(?:ing)?|design(?:ing)?|"
    r"distribut(?:e|ing)|evaluat(?:e|ing)|extract(?:ing)?|introduc(?:e|ing)|isolat(?:e|ing)|"
    r"migrat(?:e|ing)|mov(?:e|ing)|partition(?:ing)?|propos(?:e|al|ing)|recommend(?:ation|ing)?|"
    r"replac(?:e|ing)|replicat(?:e|ing)|scal(?:e|ing)|separat(?:e|ing)|shard(?:ing)?|"
    r"split(?:ting)?|switch(?:ing)?|use|using|put|insert|enqueue|offload(?:ing)?|"
    r"should (?:run|live|have|be)|spin up|stand up|roll out)\b", re.I
)
COMPONENT = re.compile(
    r"\b(?:microservices?|services?|service boundar(?:y|ies)|queues?|message brokers?|event buses?|"
    r"workers?|pipelines?|databases?|data ?stores?|object stor(?:age|es?)|stores?|caches?|clusters?|"
    r"regions?|replicas?|shards?|load balancers?|partitions?|serverless|multi[- ]region|"
    r"replication|storage tiers?|availability zones?|event streams?|monoliths?|architecture)\b", re.I
)
HOUSEKEEPING = re.compile(
    r"\b(?:add|write|update|fix|review|build|create)\s+(?:an? |the |some |unit |integration )*"
    r"(?:tests?|docs?|documentation|comments?|typos?|readmes?|diagrams?|changelogs?)\b", re.I
)
EVERYDAY = re.compile(
    r"\b(?:(?:songs?|tracks?|episodes?) (?:to|in) (?:my|the|our) queue|(?:grocery|clothing|corner) store|"
    r"customer service|room service|wait(?:ing)? in (?:the|a) queue)\b", re.I
)
NEGATED_TOPIC = re.compile(
    r"\b(?:without|rather than|instead of|avoid(?:ing)?|remov(?:e|ing)|drop|no|not|existing|current|"
    r"away from)(?:\s+[\w'’-]+){0,4}\s*$", re.I
)
OTHER_PROJECT = re.compile(
    r"(?:\b(?i:for|in|within)\s+([A-Z][\w.-]*)(?=[,:]|['’]s\b)|"
    r"\b([A-Z][\w.-]*)['’]s\b|\b(?i:project|repository|workspace)\s+[\"']?([A-Z][\w.-]*))"
)
REVIEW_SIGNALS = [
    re.compile(r"\b(?:cost|spend|unit price|budget|monthly|break[- ]even)\b", re.I),
    re.compile(r"\b(?:retries|retry|recovery|failure|availability|durability|rollback|idempotenc[ey]|backpressure)\b", re.I),
    re.compile(r"\b(?:trust boundar(?:y|ies)|least privilege|least-privilege|authorization|access control|IAM|encryption)\b", re.I),
]
SLUG = re.compile(r"[a-z0-9][a-z0-9-]{0,63}")
TERM = re.compile(r"[\w][\w .+/-]{0,63}")


def _proposal_clauses(text):
    for clause in re.split(r"[.!?;\n]", text):
        if EVERYDAY.search(clause):
            continue
        for intent in INTENT.finditer(clause):
            fragment = clause[max(0, intent.start() - 80):intent.end() + 150]
            if COMPONENT.search(fragment) and not HOUSEKEEPING.match(clause, intent.start()):
                yield clause
                break


def _path(text):
    if not isinstance(text, str) or len(text) > 4096 or "\x00" in text:
        return None
    path = Path(text)
    return path if path.is_absolute() and ".." not in path.parts else None


def _trusted_routes(path):
    if not path.is_absolute() or path.parent.is_symlink():
        raise ValueError("Use an absolute path in an owner-controlled directory.")
    parent = path.parent.stat()
    if parent.st_uid != os.getuid() or parent.st_mode & 0o022:
        raise ValueError("The routes directory must be owner-controlled.")
    # A misplaced FIFO or symlink must never block or redirect a native hook.
    fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
    with os.fdopen(fd, "rb") as stream:
        info = os.fstat(stream.fileno())
        if (not stat.S_ISREG(info.st_mode) or info.st_uid != os.getuid()
                or info.st_mode & 0o022 or info.st_size > MAX_ROUTES):
            raise ValueError("Use an owner-controlled regular routes file.")
        raw = stream.read(MAX_ROUTES + 1)
    if len(raw) > MAX_ROUTES:
        raise ValueError("Routes exceed the supported size.")
    data = json.loads(raw)
    if (not isinstance(data, dict) or set(data) != {"version", "routes"} or type(data["version"]) is not int
            or data["version"] != 1 or not isinstance(data["routes"], list) or len(data["routes"]) > 32):
        raise ValueError("Invalid routes document.")
    for route in data["routes"]:
        if (not isinstance(route, dict) or not {"profile", "project", "roots", "topics"} <= set(route)
                or set(route) - {"profile", "project", "roots", "topics", "workloads", "infer_project"}
                or not isinstance(route["profile"], str) or not SLUG.fullmatch(route["profile"])
                or not isinstance(route["project"], str) or not TERM.fullmatch(route["project"])
                or not isinstance(route["roots"], list) or not 1 <= len(route["roots"]) <= 16
                or any(_path(root) in (None, Path("/")) for root in route["roots"])
                or not isinstance(route["topics"], list) or not 1 <= len(route["topics"]) <= 16
                or any(not isinstance(t, str) or not TERM.fullmatch(t) for t in route["topics"])
                or not isinstance(route.get("workloads", []), list) or len(route.get("workloads", [])) > 16
                or any(not isinstance(t, str) or not TERM.fullmatch(t) for t in route.get("workloads", []))
                or type(route.get("infer_project", False)) is not bool):
            raise ValueError("Invalid design route.")
    return data["routes"]


def _mentions(text, term):
    return re.search(r"(?<!\w)" + re.escape(term) + r"(?!\w)", text, re.I) is not None


def _in_root(actual, claimed, root):
    # Resolve only an owner-configured root. The event can name neither a file nor a directory to open.
    declared = Path(root)
    canonical = declared.resolve()
    return actual.is_relative_to(canonical) and (
        claimed.is_relative_to(declared) or claimed.is_relative_to(canonical)
    )


def _proposes_topic(clause, term):
    for match in re.finditer(r"(?<!\w)" + re.escape(term) + r"(?!\w)", clause, re.I):
        before, after = clause[:match.start()], clause[match.end():]
        if NEGATED_TOPIC.search(before):
            continue
        if re.search(r"\breplac\w*\b[^.!?;]{0,120}$", before, re.I) and re.match(r"\s+(?:with|by)\b", after, re.I):
            continue
        direct = re.search(r"\b(?:add(?:ing)?|deploy(?:ing)?|install(?:ing)?|adopt(?:ing)?|use)\b.{0,130}$", before, re.I)
        if direct and COMPONENT.search(direct[0]) and not re.search(r"\b(?:via|through|behind)\s+(?:an?\s+)?$", before, re.I):
            continue
        return True
    return False


def _profile(text, clauses, cwd, actual, routes_path):
    claimed = _path(cwd)
    if claimed is None:
        return None
    routes = _trusted_routes(routes_path)
    scoped = [r for r in routes if any(_in_root(actual, claimed, root) for root in r["roots"])]
    named = {r["project"].casefold() for r in routes if _mentions(text, r["project"])}
    explicit = {word.casefold() for hit in OTHER_PROJECT.finditer(text) for word in hit.groups() if word}
    if explicit - named:
        return None
    if not named:
        if len({r["project"].casefold() for r in scoped}) != 1:
            return None
        scoped = [r for r in scoped if r.get("infer_project")]
    elif len(named) == 1:
        scoped = [r for r in scoped if r["project"].casefold() in named]
    else:
        return None
    matching = {r["profile"] for r in scoped if any(
        any(_mentions(clause, work) for work in r.get("workloads", []))
        and any(_proposes_topic(clause, topic) for topic in r["topics"]) for clause in clauses
    )}
    return matching.pop() if len(matching) == 1 else None


def handle(event, socket_path, routes_path):
    if not isinstance(event, dict):
        return None
    name = event.get("hook_event_name")
    if name == "Stop":
        if event.get("stop_hook_active") is not False:
            return None
        text = event.get("last_assistant_message")
    elif name == "UserPromptSubmit":
        text = event.get("prompt")
    else:
        return None
    if not isinstance(text, str) or len(text) > MAX_INPUT or text.startswith(CONTINUATION):
        return None
    clauses = list(_proposal_clauses(text))
    if not clauses:
        return None
    if name == "Stop" and all(signal.search(text) for signal in REVIEW_SIGNALS):
        return None
    guidance = GUIDANCE
    if re.search(r"\b(?:queue|queues|broker|brokers)\b", text, re.I):
        guidance += " For a queue, cover duplicate delivery, bounded retries and how to detect a growing backlog."
    try:
        profile = _profile(text, clauses, event.get("cwd"), Path.cwd(), routes_path)
        if profile and socket_path.is_absolute():
            from .design_service import request
            report = request(socket_path, {"action": "review", "profile": profile}, timeout=1.0)
            summary = report.get("summary") if isinstance(report, dict) else None
            if not isinstance(summary, str) or not summary.strip():
                raise ValueError("No profile summary.")
            safe = " ".join("".join(c if c.isprintable() else " " for c in summary[:1400]).split())
            guidance = f"Prepared ProfitCtl profile {profile}. Use the following summary as evidence. {safe}\n" + guidance
        else:
            guidance += " " + INLINE
    except (OSError, ValueError, ImportError, RecursionError, RuntimeError):
        guidance += " The prepared local review is unavailable. " + INLINE
    if name == "Stop":
        return {"decision": "block", "reason": CONTINUATION + guidance}
    return {"hookSpecificOutput": {"hookEventName": name, "additionalContext": guidance}}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--socket", type=Path, required=True)
    parser.add_argument("--routes", type=Path, required=True)
    args = parser.parse_args()

    def expired(*_):
        raise TimeoutError("Hook deadline expired.")

    signal.signal(signal.SIGALRM, expired)
    signal.setitimer(signal.ITIMER_REAL, 2.0)
    try:
        data = sys.stdin.buffer.read(MAX_INPUT + 1)
        if len(data) <= MAX_INPUT:
            result = handle(json.loads(data), args.socket, args.routes)
            if result:
                print(json.dumps(result, ensure_ascii=False))
    except (OSError, ValueError, RecursionError):
        pass
    finally:
        signal.setitimer(signal.ITIMER_REAL, 0)


if __name__ == "__main__":
    main()
