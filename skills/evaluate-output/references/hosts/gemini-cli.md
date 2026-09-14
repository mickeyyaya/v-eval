# Gemini CLI tool mapping

Source: the superpowers plugin's `skills/using-superpowers/references/gemini-tools.md`, read 2026-09-14 at `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/`.

| Action the skill requests | Gemini CLI tool |
| --- | --- |
| Read a file | `read_file`, or `read_many_files` for several at once |
| Search file contents | `grep_search` |
| List files | `list_directory`, or `glob` to find files by name |
| Run a command | `run_shell_command` |
| Fetch a URL | `web_fetch` |
| Dispatch a subagent | `invoke_agent` with `agent_name: "generalist"`, or the `@generalist` chat shortcut; parallel dispatches are several `invoke_agent` calls in one response |
| Invoke the v-eval core | `run_shell_command`, running the binary with its arguments |

Notes. `write_file` creates a file and `replace` edits one; the skill writes only its own report and never edits the artifact under evaluation. `ask_user` is the way to put a contract conflict back to the user rather than resolving it yourself.

Invoke the v-eval core: run `veval …` through the host's command tool; on Windows the binary is `veval.exe`.
