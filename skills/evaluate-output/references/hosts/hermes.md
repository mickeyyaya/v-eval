# Hermes tool mapping

Source: the superpowers plugin's `skills/using-superpowers/references/hermes-tools.md`, read 2026-09-14 at `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/`.

| Action the skill requests | Hermes tool |
| --- | --- |
| Read a file | `read_file` |
| Search file contents | `search_files` |
| List files | `terminal` with `find` |
| Run a command | `terminal` |
| Fetch a URL | `web_extract(urls=[...])`, with `web_search(query=...)` to locate one |
| Dispatch a subagent | `delegate_task(goal=..., context=..., toolsets=[...], role="leaf")`; if it is unavailable, do the pass inline rather than inventing a tool call |
| Invoke the v-eval core | `terminal`, running the binary with its arguments |

Notes. `write_file` creates a file and `patch` edits one; the skill writes only its own report. The `todo` tool tracks the steps of a long evaluation.

Invoke the v-eval core: run `veval …` through the host's command tool; on Windows the binary is `veval.exe`.
