# Go core binary, cross-compiled, with thin adapters

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

The maintainer asked for a search for the best language for the deterministic core, then added two hard constraints: the skill must work across all major agent CLIs (Claude Code, Codex, Gemini CLI, Antigravity, Hermes, ollama-backed agents), and every script and program must run on macOS, Linux, and Windows. Which language satisfies both?

The research established: Claude Code and Codex are native binaries that ship neither Node nor Python; Gemini CLI requires Node 20+; Antigravity is a native binary; Hermes bundles its own Python and Git Bash; Ollama cannot run scripts at all. A stock Windows machine has neither Python nor Node. Claude Code exec-form hooks on Windows require a real `.exe` and cannot spawn npm `.cmd` shims; Codex runs Windows hooks in cmd or PowerShell; `claude plugin eval` refuses Bash on native Windows. Hooks fire on every tool call, in parallel: one measurement put a Node hook at a 43 ms floor and 139 ms p95, while compiled binaries start in roughly 4 to 12 ms (self-reported). Go is a Tier 1 MCP SDK language with mature SARIF, JSON Schema, and JUnit libraries, and the maintainer already maintains a 140-package Go project.

## Decision Drivers

* No interpreter is guaranteed across the six target CLIs.
* Windows hooks need a real executable and no shell.
* Hook startup latency multiplies across every tool call.
* Later MCP server exposure for tool-call-only harnesses.
* Maintainer fluency and single-binary distribution.
* Counter-driver: the Agent SDK and the eval frameworks are Python or TypeScript.

## Considered Options

* Go core binary + thin adapters
* Rust core binary + thin adapters
* Python with uv-managed scripts
* TypeScript compiled with Bun

## Decision Outcome

Chosen option: "Go core binary + thin adapters", because it is the only option that meets both hard constraints across all six CLIs, meets the hook latency budget, is MCP Tier 1, and costs the maintainer no new toolchain. The core is built with `CGO_ENABLED=0` for each OS and architecture by GoReleaser and published with a sha256 checksums file. Binaries are never committed, because claude.ai rejects plugins with a top-level `bin/`. A SessionStart hook downloads the matching binary into the plugin data directory using in-box tools (`curl` and `tar` exist on Windows since build 17063), verifies the checksum, and every other hook invokes that path in exec form. Adapters to promptfoo, Inspect, and DeepEval are thin Python or JavaScript files only where a framework requires its own language. The core also exposes itself as an MCP server so ollama-backed and other tool-call-only harnesses can reach it.

### Consequences

* Good, because hooks are exec-form binary invocations on every OS with no shell.
* Good, because one release channel serves all CLIs.
* Bad, because framework adapters are a second, versioned contract to keep in sync.
* Bad, because the Agent SDK is not available; the agent rung drives `claude -p` or uses a thin SDK shell.
* Bad, because CI must build and test on three operating systems from the first commit.

## Confirmation

GoReleaser config produces darwin, linux, and windows artifacts with a checksums file; a Windows CI job runs the core's tests and a hook smoke test; the plugin repository contains no committed binary. None exists yet.

## Pros and Cons of the Options

### Go core binary + thin adapters

* Good, because stock Windows and no-shell hook invocation are both satisfied.
* Good, because MCP Tier 1 and mature SARIF, JSON Schema, JUnit libraries.
* Neutral, because eval-framework adapters become small shell-out scripts.
* Bad, because no Agent SDK.

### Rust core binary + thin adapters

* Good, because the same portability profile via cargo-dist.
* Bad, because it is a new language for the maintainer with slower iteration.

### Python with uv-managed scripts

* Good, because it is native to Inspect, DeepEval, promptfoo, and the Agent SDK.
* Bad, because Python is absent on stock Windows and on Claude Code, Codex, and Antigravity hosts.
* Bad, because `uv run` adds about 45 ms per hook and must itself be installed.

### TypeScript compiled with Bun

* Good, because npm dependencies auto-install in Claude Code plugins.
* Bad, because single-file executables are large and Node is guaranteed only under Gemini.
* Bad, because `.cmd` shims fail in Windows hooks.

## More Information

* Related requirements: REQ-01, REQ-04; portability in [0020](0020-portability-constraints.md). Informed by [language and portability research](../research/2026-09-14-language-and-portability.md).
* Claude Code setup: <https://code.claude.com/docs/en/setup> (accessed 2026-09-14). Hooks reference: <https://code.claude.com/docs/en/hooks> (accessed 2026-09-14). Plugin evals: <https://code.claude.com/docs/en/plugin-evals> (accessed 2026-09-14). Plugin marketplaces (rejects top-level `bin/`): <https://code.claude.com/docs/en/plugin-marketplaces> (accessed 2026-09-14).
* Codex hooks with `commandWindows`: <https://developers.openai.com/codex/hooks> (accessed 2026-09-14). Gemini CLI installation: <https://geminicli.com/docs/get-started/installation> (accessed 2026-09-14). Ollama CLI: <https://docs.ollama.com/cli> (accessed 2026-09-14).
* MCP SDK tiers: <https://modelcontextprotocol.io/docs/sdk> (accessed 2026-09-14). Agent SDK overview: <https://code.claude.com/docs/en/agent-sdk/overview> (accessed 2026-09-14).
* Hook latency measurement, one machine: <https://github.com/macanderson/oxagen/issues/2843> (2026-09-10; accessed 2026-09-14). Python on Windows: <https://docs.python.org/3/using/windows.html> (accessed 2026-09-14). GoReleaser checksums: <https://goreleaser.com/customization/checksum> (accessed 2026-09-14).
* Revisit when hosts guarantee a common interpreter, when the Agent SDK becomes essential to the agent rung, or when Windows measurements contradict the latency assumptions.
