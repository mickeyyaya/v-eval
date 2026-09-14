# ollama-backed agent tool mapping

Source: written for this skill, 2026-09-14. ollama exposes model and tool-call APIs; it has no tool set of its own to copy.

No tools of its own; runs through a harness (Claude Code, Codex, OpenCode) whose mapping applies, or reaches the core over MCP in Stage 5; without either, every execution criterion is UNKNOWN and the report says so.

| Action the skill requests | ollama-backed agent |
| --- | --- |
| Read a file | The harness's read tool; over MCP, the core's exposed read. Neither present: the evaluation cannot proceed and the report says so. |
| Search file contents | The harness's search tool; otherwise read what you can and record what was not searched. |
| List files | The harness's listing tool; otherwise the scope inspected is whatever was supplied, stated as a limitation. |
| Run a command | The harness's command tool; over MCP, the core's exposed call. Neither present: every execution criterion is UNKNOWN. |
| Fetch a URL | The harness's fetch tool; otherwise source-supported criteria stay UNKNOWN. |
| Dispatch a subagent | The harness's dispatch tool; otherwise a sequential self-pass in one context, labeled as such, since it is weaker evidence than a cold re-judge. |
| Invoke the v-eval core | The harness's command tool, or the core's MCP surface once it exists. |

Notes. A small local model must not compute the rollup or invent evidence: use the core when it is reachable, keep a result UNKNOWN when unsure, and quote rather than paraphrase. Record in the report which harness supplied the tools, because that, not ollama, determines what could be run.

Invoke the v-eval core: run `veval …` through the host's command tool; on Windows the binary is `veval.exe`.
