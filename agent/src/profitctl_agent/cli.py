"""Command-line and stdio entry points for a local ProfitCtl investigation."""

import argparse
import asyncio
import json
import os
import sys
from collections.abc import Callable
from functools import partial
from pathlib import Path
from typing import TYPE_CHECKING, Annotated, Any, Literal

from mcp.server.fastmcp import FastMCP
from mcp.types import ToolAnnotations
from pydantic import Field, ValidationError

from .economics import OwnerEvidence

if TYPE_CHECKING:
    from .design_service import SocketClient
    from .runtime import InvestigationService


Identifier = Annotated[str, Field(pattern=r"^[a-z0-9][a-z0-9-]{0,63}$")]
Question = Annotated[str, Field(min_length=1, max_length=4000)]
Revision = Annotated[int, Field(ge=1, strict=True)]
Action = Literal["supply", "decline"]


def _invoke(operation: Callable[..., dict], **kwargs: Any) -> dict:
    try:
        return operation(**kwargs)
    except ValueError:
        raise
    except Exception as exc:
        # A provider exception can include request bodies, secrets or local paths.
        raise RuntimeError(
            f"Investigation operation failed ({type(exc).__name__}). "
            "Check the configured Go executable and model connection. "
            "Read the run status before retrying with the same run ID."
        ) from None


def make_server(service: "InvestigationService | SocketClient", *, review_design: Callable[..., dict] | None = None) -> FastMCP:
    server = FastMCP(
        "profitctl-agent",
        instructions=(
            "Start or resume a ProfitCtl investigation in the configured private workspace. "
            "Return the run ID to the operator. Read status before resuming in another client. "
            "Supply only evidence the operator provided for that run."
        ),
        log_level="WARNING",
    )

    @server.tool(annotations=ToolAnnotations(readOnlyHint=False, destructiveHint=False,
                                             idempotentHint=True, openWorldHint=True))
    async def profitctl_start_investigation(
        run_id: Identifier, case_id: Identifier, question: Question
    ) -> dict[str, Any]:
        """Start an investigation for a saved case. Choose a fresh stable run_id and reuse it for identical retries."""
        return await asyncio.to_thread(
            _invoke, service.start, run_id=run_id, case_id=case_id, question=question
        )

    @server.tool(annotations=ToolAnnotations(readOnlyHint=True, openWorldHint=False))
    async def profitctl_investigation_status(run_id: Identifier) -> dict[str, Any]:
        """Read the checkpoint, recorded owner response and decision revision. For an interrupted run, retry the returned retry_tool_call."""
        return await asyncio.to_thread(_invoke, service.status, run_id=run_id)

    @server.tool(annotations=ToolAnnotations(readOnlyHint=False, idempotentHint=True, openWorldHint=True))
    async def profitctl_resume_investigation(
        run_id: Identifier, expected_revision: Revision, action: Action, evidence: OwnerEvidence | None = None,
    ) -> dict[str, Any]:
        """Supply sourced evidence or decline at the base_revision. After an error, read status to see whether the response was saved and how to retry."""
        update = evidence.model_dump(exclude_none=True, exclude_unset=True) if evidence is not None else None
        return await asyncio.to_thread(
            _invoke, service.resume, run_id=run_id, expected_revision=expected_revision,
            action=action, evidence=update,
        )

    if review_design is not None:
        @server.tool(annotations=ToolAnnotations(readOnlyHint=True, idempotentHint=True, openWorldHint=False))
        async def profitctl_review_design(
            profile: Identifier | None = None, proposal: dict[str, Any] | None = None,
        ) -> dict[str, Any]:
            """Review design economics using Go. Supply exactly one registered profile slug or self-contained profitctl.design/v1 proposal object. Return sourced gaps in incomplete reports."""
            return await asyncio.to_thread(review_design, profile=profile, proposal=proposal)

    return server


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="profitctl-agent",
        description="Start and resume economic investigations in a private workspace.",
        epilog="The local runtime requires a POSIX host such as macOS or Linux. "
               "Claude subscription verification uses the default macOS Keychain profile.",
    )
    workspace = os.environ.get("PROFITCTL_WORKSPACE_ROOT")
    parser.add_argument("--workspace", type=Path, default=workspace, required=not workspace,
                        help="Private case workspace (or set PROFITCTL_WORKSPACE_ROOT)")
    parser.add_argument("--profitctl-binary", default="profitctl",
                        help="Decision-enabled Go executable, resolved when the service starts")
    commands = parser.add_subparsers(dest="command", required=True)
    start = commands.add_parser("start", help="Open an investigation for a saved case")
    start.add_argument("run_id", help="New stable ID for this investigation and its retries")
    start.add_argument("--case", dest="case_id", required=True, help="Existing ProfitCtl decision ID")
    start.add_argument("--question", required=True, help="Economic question to investigate")
    status = commands.add_parser("status", help="Read a checkpoint and pending request")
    status.add_argument("run_id")
    resume = commands.add_parser("resume", help="Continue a paused run with an owner response")
    resume.add_argument("run_id")
    resume.add_argument("--expected-revision", required=True, type=int, help="base_revision from status")
    resume.add_argument("--action", required=True, choices=("supply", "decline"))
    inputs = resume.add_mutually_exclusive_group()
    inputs.add_argument("--evidence", help="JSON object containing kind, source and updates")
    inputs.add_argument("--evidence-file", type=Path, help="Read the same JSON object from a local file")
    mcp = commands.add_parser("mcp", help="Serve investigation tools over local stdio")
    mcp.add_argument("--socket", type=Path,
                     help="Route to the persistent local agent and expose its design review tool")
    serve = commands.add_parser("serve", help="Run the persistent agent on a private Unix socket")
    serve.add_argument("--socket", type=Path, required=True, help="Socket in an existing owner-only directory")
    serve.add_argument("--profiles", type=Path, required=True, help="Owner-only directory of registered JSON profiles")
    return parser


def _evidence(args: argparse.Namespace) -> dict | None:
    if args.action == "decline":
        if args.evidence is not None or args.evidence_file is not None:
            raise ValueError("Declining takes no evidence. Choose supply to update this investigation.")
        return None
    raw = args.evidence
    if args.evidence_file is not None:
        with args.evidence_file.open("rb") as stream:
            raw = stream.read(65537)
    if raw is None:
        raise ValueError("Supplying an observation requires --evidence or --evidence-file.")
    if len(raw if isinstance(raw, bytes) else raw.encode()) > 65536:
        raise ValueError("Supply at most 64 KiB of evidence.")
    try:
        return OwnerEvidence.model_validate_json(raw).model_dump(exclude_none=True, exclude_unset=True)
    except ValidationError as exc:
        details = "; ".join(
            f"{'.'.join(map(str, err['loc'])) or 'evidence'}: {err['msg']}"
            for err in exc.errors(include_input=False, include_url=False)
        )
        raise ValueError(f"Invalid evidence. {details}") from None


def main() -> int:
    args = _parser().parse_args()

    try:
        if args.command == "serve":
            from .design_service import serve

            serve(args.workspace, args.profitctl_binary, args.socket, args.profiles)
            return 0
        if args.command == "mcp" and args.socket is not None:
            from .design_service import SocketClient

            client = SocketClient(args.socket)
            make_server(client, review_design=client.review).run(transport="stdio")
            return 0
        from .runtime import InvestigationService

        service = InvestigationService(args.workspace, profitctl_binary=args.profitctl_binary)
        if args.command == "mcp":
            make_server(service).run(transport="stdio")
            return 0
        if args.command == "start":
            operation = partial(service.start, run_id=args.run_id, case_id=args.case_id, question=args.question)
        elif args.command == "resume":
            operation = partial(service.resume, run_id=args.run_id, expected_revision=args.expected_revision,
                                action=args.action, evidence=_evidence(args))
        else:
            operation = partial(service.status, run_id=args.run_id)
        print(json.dumps(_invoke(operation), allow_nan=False))
        return 0
    except (ValueError, OSError, RuntimeError) as exc:
        print(f"profitctl-agent: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
