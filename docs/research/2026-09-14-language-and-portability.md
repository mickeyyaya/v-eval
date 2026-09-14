# Implementation language for the deterministic core and CLI

Research date and access date: 2026-09-14.

Method: Web-research agent. 29 sources; about eight read in full. Scope was extended mid-run to cover macOS, Linux, and Windows portability and the six target agent CLIs.

Facts, inferences, single-source claims, and unverified items are labeled inside the memo; a label applies to the sentence or row it sits in.

---

Research memo, 2026-09-14 (all URLs accessed that day). Facts cited; inferences labelled.

## 1. Sources (title — author, date; establishes; limits)

| # | Source | Establishes | Limits |
|---|---|---|---|
| S1 | anthropics/skills (Anthropic) github.com/anthropics/skills — docx, skill-creator `scripts/` | All bundled scripts are plain `.py` (`#!/usr/bin/env python3`); docx SKILL.md: `docx` npm is "preinstalled". No uv/PEP 723. | No stated language policy. |
| S2 | Agent Skills spec agentskills.io/specification | Scripts "Be self-contained or clearly document dependencies"; "Supported languages depend on the agent implementation. Common options include Python, Bash, and JavaScript"; `compatibility: Requires Python 3.14+ and uv` example. | Language-agnostic. |
| S3 | Skill best practices (Anthropic) platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices | "Prefer scripts for deterministic operations"; "Claude API: Has no network access and no runtime package installation"; "Avoid assuming tools are installed". Python examples only. | No uv/binary guidance. |
| S4 | Claude Code setup code.claude.com/docs/en/setup | Native binary; npm package "installs the same native binary"; "does not itself invoke Node"; Git for Windows optional, PowerShell fallback. | Ships no Python/Node. |
| S5 | Codex: github.com/openai/codex (`codex-rs`); Discussion #1174 (2025-05-30); InfoQ 2025-06-04 | Rust rewrite: "Zero-dependency Install — currently Node v22+ is required"; curl/PowerShell installers, GitHub release binaries. | — |
| S6 | Gemini CLI install geminicli.com/docs/get-started/installation | "Node.js 20.0.0+"; npm/brew/npx; no standalone binary. | — |
| S7 | Agent SDK overview code.claude.com/docs/en/agent-sdk/overview | "available as a library for Python and TypeScript only. To drive the same agent loop from another language, run the CLI as a subprocess with the `-p` flag and `--output-format json`." | — |
| S8 | MCP SDKs modelcontextprotocol.io/docs/sdk | Tier 1: TypeScript, Python, C#, Go, Rust; Tier 2: Java, Ruby; Tier 3: Swift, PHP, Kotlin. | — |
| S9 | Plugins reference code.claude.com/docs/en/plugins-reference | `${CLAUDE_PLUGIN_ROOT}` for "Scripts, binaries, and config files"; plugins copied to cache; Node deps auto-installed (`npm ci --ignore-scripts`, 60 s); `${CLAUDE_PLUGIN_DATA}` for "Python dependencies" via SessionStart hook. | No Python/Go auto-install. |
| S10 | Plugin marketplaces code.claude.com/docs/en/plugin-marketplaces | "claude.ai rejects a plugin that has [a top-level `bin/`]"; use `scripts/`. | — |
| S11 | Plugin evals code.claude.com/docs/en/plugin-evals | Sandbox strips hooks/env (allowlist + `EVAL_*`); "Native Windows has no backend, so run shell-granting suites under WSL2"; plugin hooks "run outside the agent's sandbox". | — |
| S12 | Hooks reference code.claude.com/docs/en/hooks | PreToolUse/PostToolUse "on every tool call"; "All matching hooks run in parallel"; exec form (`args`): "There is no shell"; Windows exec form "requires `command` to resolve to a real executable such as a `.exe`" — `.cmd/.bat` shims "can't be spawned without a shell"; default shell bash, PowerShell if no Git Bash. | — |
| S13 | oxagen #2843 (macanderson, 2026-09-10) github.com/macanderson/oxagen/issues/2843 | Measured `node -e 0` = 43 ms; Node hook p50 108 / p95 139 ms; "no Node executable can meet the [30 ms] budget" → compiled binary. | One machine. |
| S14 | Lily, dev.to 2026-08-20 "Your Claude Code Hooks Are Costing You Minutes" | Hooks fire dozens of times/session; "Python3 startup can cost tens to hundreds of milliseconds". | No numbers. |
| S15 | HN thread news.ycombinator.com/item?id=46230192; bdrung/startup-time | `python3 -c` 22 ms (M-series); `uv run python -c` 58 ms vs 13 ms direct; RPi3: Go 4.1 ms, Python 198 ms. | UNVERIFIED user data; no Windows. |
| S16 | Howl github.com/ai-screams/howl | Go statusline for Claude Code: ~10 ms cold start, 5.6 MB, CGO off, mac/linux releases. | Self-reported. |
| S17 | promptcellar PR #14 (2026-05-06) github.com/dominiek/promptcellar-for-claude-code/pull/14 | Gitignored Go binaries broke marketplace install; fix: committed shims download release tarball on first SessionStart, sha256-verify, `exec`. | Unix only. |
| S18 | JamesPrial/go-plugin-release | Cross-compiles 5 targets incl. windows-amd64, checksums, OS/arch wrapper, orphan `releases` branch, "No CGO". | 0 stars. |
| S19 | spencerbeggs/claude-binary-plugin; Bun `--compile` docs bun.sh/docs/bundler/executables | TS hooks → Bun single-file exe, 8 targets; Bun admits "binary is still way too big". | — |
| S20 | promptfoo assertions promptfoo.dev/docs/configuration/expected-outputs | Custom assertions: `javascript`, `python`, `ruby`, `file://*.js\|*.py`. No shell/exec type. | — |
| S21 | Library repos (GitHub) | Go: owenrumney/go-sarif v3 (schema `Validate()`), santhosh-tekuri/jsonschema v6 (1.26k★, passes official suite), joshdk/go-junit. Py: python-jsonschema (5k★), sarif-python-om (last push 2024-04). TS: ajv (14.8k★), node-sarif-builder (3.2M wk dl), @microsoft/sarif (2026-06). CTRF ref impl TS, schema normative. Inspect/DeepEval Python; openevals Py+TS. | — |
| S22 | Anhaia dev.to 2026-04-18; thedailyagent 2026-05-06; Archit Singh 2026-02-28 "Designing CLI Tools for AI Agents"; Bright Coding 2026-06-04 | Practitioner consensus: static binary, no runtime env, `--json`, fast start. Field-guide cold-start table (Py 200–500 ms, Go <10, Rust <5) is unsourced; Gopher "12 ms" self-reported at 3% parity. | Opinion pieces. |
| S23 | Codex hooks developers.openai.com/codex/hooks; openai/codex #24453 (2026-05-25); mem0 #6181 | Command hooks; `commandWindows` override; plugin hooks need trust; native Windows runs commands in cmd/PowerShell (POSIX syntax fails); #24453: Windows `command_execution` not emitting PreToolUse. | Bug status unknown. |
| S24 | Gemini CLI hooks geminicli.com/docs/hooks/reference; skills geminicli.com/docs/cli/skills | Only `type: "command"`; timeout 60 000 ms; PowerShell example; Agent Skills in `.agents/skills`. | Script execution unspecified. |
| S25 | Antigravity CLI antigravity.google/docs/cli/plugins, /install | `agy` "runs natively on macOS, Linux, and Windows" (`%LOCALAPPDATA%\agy\bin`); plugins with `hooks.json`, markdown skills, terminal tool, headless JSON. | Hook shell semantics UNVERIFIED. |
| S26 | Hermes Agent hermes-agent.nousresearch.com/docs (windows-native; creating-skills) | Python agent; Windows terminal tool "runs commands through Git Bash" (bundled PortableGit); agentskills.io-compatible with `platforms:` and `${HERMES_SKILL_DIR}`; "Prefer stdlib Python, curl". | — |
| S27 | Ollama CLI docs.ollama.com/cli | `ollama launch` delegates to Claude Code/Codex/OpenCode/Droid; Ollama exposes only model/tool-call APIs. | Cannot run scripts. |
| S28 | Python docs docs.python.org/3/using/windows.html; MS Learn nodejs-on-windows; MS devblogs "Tar and Curl Come to Windows" (2018) | "Windows does not include a system supported installation of Python"; Node must be installed; `curl.exe`/`tar` in-box since build 17063. | — |
| S29 | GoReleaser goreleaser.com/customization/checksum, /builds/go; cargo-dist book | GOOS/GOARCH matrix; sha256 checksums file; cargo-dist installers + manifests from a tag. | — |

## 2. Findings (facts, then inferences)

**Q1.** Anthropic ships Python scripts and relies on a preinstalled container (S1, S3); the spec is language-agnostic (S2). *Inference:* this reflects Anthropic's sandbox, not a portable-host guarantee.

**Q2.** Claude Code and Codex are native binaries with no Node/Python (S4, S5); Gemini needs Node 20+ (S6); Antigravity is a native binary (S25); Hermes installs its own Python/Node/Git Bash (S26); Ollama is a server (S27); stock Windows has neither Python nor Node (S28). **None of node/python/go is guaranteed across the six** — Node only under Gemini, Python only under Hermes.

**Q3.** Agent SDK is Python/TS; other languages drive `claude -p --output-format json` (S7). Go and Rust are MCP Tier 1 (S8). *Inference:* a Go core becomes an MCP server natively; the standalone-agent phase needs a thin Python/TS shell or subprocess control.

**Q4.** Docs allow bundled "binaries" (S9) but claude.ai rejects top-level `bin/` (S10); Node deps auto-install, Python deps need a SessionStart hook into `${CLAUDE_PLUGIN_DATA}` (S9). Field pattern: gitignored binaries break installs; fix is download-on-first-SessionStart with sha256 (S17, S18). Eval sandbox runs hooks outside itself and refuses Bash on native Windows (S11).

**Q5.** SARIF/JSON Schema/JUnit are mature in Go (S21); promptfoo assertions are JS/Python only (S20); Inspect/DeepEval are Python. *Inference:* a non-Python/JS core needs ~10-line shell-out adapters — cheap.

**Q6/Q7.** Hooks fire per tool call, in parallel (S12). Node floor 43 ms, real hook p95 139 ms (S13); Python 13–22 ms warm on fast hardware, ~200 ms slow (S15, UNVERIFIED); `uv run` adds ~45 ms cached (S15); Go ~4–12 ms (S15, S16, S22, self-reported). *Inference:* only compiled binaries reliably stay under ~30 ms.

## 3. Windows and cross-CLI requirements

**(a) Windows.** Claude Code: bash not guaranteed (Git for Windows optional); hooks fall back to PowerShell; exec-form hooks need a real `.exe` and cannot spawn npm `.cmd` shims (S4, S12); `claude plugin eval` refuses Bash on native Windows (S11). Codex: hooks run in cmd/PowerShell, use `commandWindows`, POSIX syntax fails, and a Windows hook-dispatch bug is open (S23). Gemini: command hooks with PowerShell examples; Node present by definition (S24). Antigravity: native Windows binary; hook shell semantics undocumented (S25). Hermes: bundled Git Bash (S26). Python/Node absent on stock Windows (S28). **Only a Go/Rust `.exe` runs from every CLI's hook without a shell or interpreter** (S12 exec form; S23 `commandWindows`).

**(b) Cross-compile/distribution.** GoReleaser builds the GOOS/GOARCH matrix and a sha256 checksums file; cargo-dist is the Rust equivalent (S29). Keep the plugin source-only; a SessionStart hook (bash on Unix; `shell: "powershell"` / `commandWindows` on Windows) downloads `v-eval_<os>_<arch>` into `${CLAUDE_PLUGIN_DATA}` with in-box `curl`/`tar` (S28), verifies the checksum, and every other hook calls that path in exec form (S9, S12, S17, S18). Never commit binaries (S10, S17).

**(d) CLIs that cannot run scripts.** Ollama cannot; it delegates to a harness (S27). The other five expose a shell tool; Antigravity headless soft-denies shell unless granted (S25). *Inference:* expose the core as an MCP server too, so tool-call-only harnesses can reach it.

## 4. Decision matrix

| | Go | Python | TypeScript | Rust |
|---|---|---|---|---|
| Host-runtime guarantee | None needed | Not guaranteed (S28) | Only under Gemini (S6) | None needed |
| Plugin distribution | Release binaries + SessionStart download (S17, S18) | Needs `uv`; `${CLAUDE_PLUGIN_DATA}` install (S9) | Auto `npm ci` in cache (S9) — best-supported | Same as Go (cargo-dist) |
| Hook startup latency | ~4–12 ms (S15, S16) | 13–200 ms; +45 ms via `uv run` (S15) | 43 ms floor, ~110 ms real (S13) | <10 ms (S22, unsourced) |
| Eval adapters | Shell-out for promptfoo/Inspect | Native for Inspect/DeepEval/promptfoo | Native for promptfoo/openevals/CTRF | Shell-out |
| MCP / Agent SDK | MCP Tier 1; Agent SDK via `claude -p` (S7, S8) | Both native | Both native | MCP Tier 1; no Agent SDK |
| Maintainer fit (140-pkg Go) | Highest | Second toolchain | Second toolchain | New language |
| Single-binary portability | Yes, CGO off | No | Bun compile, large (S19) | Yes |
| Windows without extra installs | Yes (`.exe`) | No | No (except Gemini) | Yes |
| No-shell invocation from hooks | Yes: exec form / `commandWindows` (S12, S23) | No (needs `python`) | No (`node`; `.cmd` shims fail, S12) | Yes |

Shell is excluded: absent on native Windows under Claude Code/Codex (S4, S23) and refused in Windows evals (S11).

## 5. Recommendation

**Go** — single static binary (`CGO_ENABLED=0`), GoReleaser + checksums, downloaded per OS/arch into `${CLAUDE_PLUGIN_DATA}` on first SessionStart.

1. Only option meeting both hard requirements (stock Windows, no-shell hook invocation) across all six CLIs.
2. Hooks fire on every tool call in parallel; Go's ~10 ms start is the only class that meets sub-30 ms budgets (S12, S13).
3. SARIF, JSON Schema, JUnit and MCP Tier 1 are first-class in Go (S8, S21).
4. Maintainer already runs 140 Go packages — zero toolchain cost.
5. Codex chose a native binary for "zero-dependency install" (S5); the ecosystem is converging there.

**Strongest counter-argument:** the Agent SDK is Python/TS only (S7), Anthropic's own skills are Python (S1), and promptfoo/Inspect/DeepEval assertions are Python/JS (S20). A Python core would be native to every eval framework and to the standalone-agent phase, with `uv run --script` for self-containment — at ~45 ms per hook (S15) and with `uv`/Python required on Windows, which no CLI installs (S28).

## 6. Hybrid: Go core + thin Python/TS adapters

Core binary: schema validation, verdict aggregation, provenance, JUnit/CTRF import, SARIF export, `--json` output (S22), later `v-eval mcp` via go-sdk. Adapters: a PEP 723 Python scorer for promptfoo/Inspect that runs `v-eval judge --json`; a 20-line JS `file://` promptfoo assertion; the standalone agent as a Python Agent SDK script invoking the binary, or Go driving `claude -p` (S7). **Costs:** two release channels to sync; adapters need their interpreter (acceptable — they run only where promptfoo/Inspect already run); a versioned JSON contract between core and adapters; doubled CI matrix.

## 7. Gaps

- No Windows latency measurements for Go vs `python.exe` vs `uv run` (S15 is Unix, user-reported).
- Antigravity hook execution semantics undocumented; Codex Windows PreToolUse bug (#24453) status unknown.
- Exec-form `args` in plugin `hooks/hooks.json` assumed from the shared schema (S12) — verify.
- claude.ai `bin/` rejection (S10) means the binary must live under `scripts/` or `${CLAUDE_PLUGIN_DATA}`; untested end-to-end.
- Go cold-start figures (S15, S16, S22) are self-reported; no independent benchmark.
