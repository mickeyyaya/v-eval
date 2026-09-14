# Codex tool mapping

Source: the superpowers plugin's `skills/using-superpowers/references/codex-tools.md`, read 2026-09-14 at `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/`.

That reference documents Codex's multi-agent and shell behavior; it names no separate file-reading, search, listing, or URL-fetching tools, and its own environment-detection recipe runs read-only `git` commands as shell commands. The rows below say which cells come from the source and which do not.

| Action the skill requests | Codex tool |
| --- | --- |
| Read a file | Not named in the source. Codex reads through the shell (`cat`, `sed -n`) unless your tool list offers a file tool; confirm before citing. |
| Search file contents | Not named in the source. Use `grep` or `rg` through the shell. |
| List files | Not named in the source. Use `ls` or `find` through the shell. |
| Run a command | The shell tool, as the source's environment-detection commands are run. Hooks on Windows take the `commandWindows` override. |
| Fetch a URL | Not named in the source. Without a fetch tool, criteria that need a retrieved source stay UNKNOWN with the reason stated. |
| Dispatch a subagent | `spawn_agent` with `fork_turns: "none"` for a clean context (requires `multi_agent = true` in `~/.codex/config.toml`); resume a child with `followup_task`; wait with `wait_agent` in bounded stretches; `list_agents` reconciles. Set `model` and `reasoning_effort` together on every spawn. |
| Invoke the v-eval core | The shell tool, running the binary with its arguments. |

The source's own instruction applies to this table too: trust your actual tool list over any table when they disagree, and record the tool you used in the report's provenance.

Invoke the v-eval core: run `veval …` through the host's command tool; on Windows the binary is `veval.exe`.
