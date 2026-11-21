# Automatic design review in Codex and Claude Code

Ask “Should we move Condere report assembly behind a queue?” In an approved Condere checkout, the hook can look up the prepared `condere-queue` comparison. The host receives its source-qualified summary along with questions about workload, failure recovery, security and cost per successful outcome. It can request a full inline comparison through `profitctl_review_design`.

A billing-email queue and a Redis cache for report assembly are different decisions. For either one, the host gets guidance to evaluate the actual workload using sourced assumptions, or identify the missing measurement. Everyday requests such as listing services and updating tests proceed silently.

## Configure the two clients

Use the Python interpreter of an installed ProfitCtl agent build. Choose owner-controlled locations for the Unix socket, routes and backups. In the commands below, set `agent_python` to that interpreter and `socket_path` to the running design review service's socket. Run from the ProfitCtl checkout when copying the example.

```sh
agent_python="/absolute/path/to/the/installed/venv/bin/python"
socket_path="/absolute/path/to/the/running/design-review.sock"
hook_dir="$HOME/.local/share/profitctl/design-hooks"
mkdir -p -m 700 "$hook_dir"
install -m 600 examples/design/hook-routes.json "$hook_dir/routes.json"
```

Edit `roots` in the copied file to point at the actual project checkout. The prepared profile name must match one on the service. The route file and its directory must belong to the current user and permit writes only from that user. An empty `routes` list enables generic reviews everywhere.

| Route field | How it affects matching |
| --- | --- |
| `profile` | A service profile slug such as `condere-queue`. |
| `project` | A project name matched as a complete term, ignoring case. |
| `roots` | Absolute, approved directories. The hook process and the event's reported working directory must both fall under the same root. |
| `workloads` | Names for this particular workload, such as `report assembly` and `assemble reports`. A matching workload and a proposed change must appear together. |
| `topics` | Names for that change, such as `queue` or `message broker`. An existing or rejected queue does not qualify as the proposal. |
| `infer_project` | Optional and false by default. Set true for a dedicated project root to allow “Should we put report assembly behind a queue?” A shared directory requires an explicit project name. |

The outer document has integer `version` 1 and a `routes` array. Pricing requires one matching profile. Ambiguous scopes, routes without workload terms, mismatched projects, and a service outage all leave the host with generic guidance. Keep workload terms as specific as the economic model. Use separate records for dedicated and shared roots.

```sh
"$agent_python" -I -m profitctl_agent.design_hook_config install \
  --codex-config "$HOME/.codex/hooks.json" \
  --claude-config "$HOME/.claude/settings.json" \
  --python "$agent_python" --socket "$socket_path" \
  --routes "$hook_dir/routes.json" --backup-dir "$hook_dir/backups"
```

The installer appends two command handlers per client. It preserves existing hooks, checks for concurrent edits, and prints the private backup folder when it changes settings. Use the same arguments for a repeat installation or uninstallation. In Codex, use `/hooks` to review and trust the two new definitions. Claude Code also lists them under `/hooks`. Start fresh client sessions after changing the configuration. [Codex hooks](https://developers.openai.com/codex/hooks/) · [Claude Code hooks](https://code.claude.com/docs/en/hooks)

```sh
"$agent_python" -I -m profitctl_agent.design_hook_config uninstall \
  --codex-config "$HOME/.codex/hooks.json" \
  --claude-config "$HOME/.claude/settings.json" \
  --python "$agent_python" --socket "$socket_path" \
  --routes "$hook_dir/routes.json" --backup-dir "$hook_dir/backups"
```

Uninstallation restores the original file when it still matches the installed version. If settings have evolved, it removes matching ProfitCtl handlers and preserves later changes. An operator-edited ProfitCtl handler requires inspection.

## Hook behavior

`UserPromptSubmit` classifies the current prompt locally and can return `additionalContext`. `Stop` looks at the native `last_assistant_message` to catch a design proposed by the assistant. An incomplete design review can receive one continuation. The native `stop_hook_active` flag bounds this correction. A response that already addresses economics, failure behavior and a trust control finishes normally.

The hook reads up to 64 KiB from stdin. It ignores unsupported events and oversized inputs. It passes only the fixed review action and an owner-configured profile slug to the local socket, with a one-second request timeout. It evaluates text in memory, ignores transcript paths, and stores no prompt or transcript. The local service summary is bounded to 1,400 characters. Isolated Python mode keeps project files and `PYTHONPATH` out of module resolution. The hook has a two-second internal deadline and a four-second native command limit.
