# Claude Code tool mapping

Source: Claude Code's own tool set, read from the tools available to this skill's host session, 2026-09-14.

| Action the skill requests | Claude Code tool |
| --- | --- |
| Read a file | `Read` |
| Search file contents | `Grep` |
| List files | `Glob` |
| Run a command | `Bash` |
| Fetch a URL | `WebFetch` |
| Dispatch a subagent | `Agent` |
| Invoke the v-eval core | `Bash`, running the binary with its arguments |

Notes. `Read` returns the file with line numbers, which is what a criterion's file-and-line citation quotes. `Grep` and `Glob` search without reading whole files; a match is a pointer, not evidence, until you open it. `Agent` gives a cold context for a re-judge, which is the only way to re-decide a criterion without carrying the first decision into the second. A permission prompt that is denied is an observation: record the action as not executed, and leave the criteria that depended on it UNKNOWN.

Invoke the v-eval core: run `veval …` through the host's command tool; on Windows the binary is `veval.exe`.
