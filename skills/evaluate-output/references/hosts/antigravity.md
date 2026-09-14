# Antigravity (`agy`) tool mapping

Source: the superpowers plugin's `skills/using-superpowers/references/antigravity-tools.md`, read 2026-09-14 at `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/`.

That reference maps two actions only, subagent dispatch and task tracking, and names the file-writing and background-process tools in passing. It names no reading, search, listing, command, or URL-fetching tool, so those rows are marked unknown rather than guessed.

| Action the skill requests | Antigravity tool |
| --- | --- |
| Read a file | Not named in the source; take it from your own tool list and record which tool you used. |
| Search file contents | Not named in the source; take it from your own tool list. |
| List files | Not named in the source; take it from your own tool list. |
| Run a command | Not named in the source. `manage_task` is not it: that tool manages background processes (`list`, `kill`, `status`, `send_input`). Use the command tool your session actually offers, and if there is none, every execution criterion is UNKNOWN and the report says so. |
| Fetch a URL | Not named in the source; without a fetch tool, criteria that need a retrieved source stay UNKNOWN. |
| Dispatch a subagent | `invoke_subagent` with a built-in `TypeName`: `self` for full-capability work, `research` for a read-only pass |
| Invoke the v-eval core | The session's command tool, running the binary with its arguments. |

Notes. Antigravity has no todo tool, so a multi-step evaluation tracks its steps in a task artifact: `write_to_file` with `IsArtifact: true` and `ArtifactMetadata.ArtifactType: "task"`, edited with `replace_file_content` or `multi_replace_file_content` as steps complete. The evidence you gather is written with the same file tools; keep it out of the artifact under evaluation.

Invoke the v-eval core: run `veval …` through the host's command tool; on Windows the binary is `veval.exe`.
