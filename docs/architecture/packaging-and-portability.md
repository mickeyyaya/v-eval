# Packaging and portability

Status: design of 2026-09-14, partly implemented. The schema, the core, the command line, the skill with its per-host reference files, the three-OS CI matrix, and the GoReleaser configuration exist; the agent and isolation profile, the plugin manifests and hooks, and the MCP server do not. Decisions: [decision 0004 start at rungs 2 to 4](../decisions/0004-start-at-rungs-2-to-4.md), [decision 0005 Go core binary](../decisions/0005-go-core-binary.md), [decision 0020 portability constraints](../decisions/0020-portability-constraints.md), [decision 0018 skill regression testing](../decisions/0018-skill-regression-testing.md). Research: [skill packaging memo](../research/2026-09-14-skill-packaging.md), [language and portability memo](../research/2026-09-14-language-and-portability.md). All sources accessed 2026-09-14. Items marked **proposal** were not decided by the maintainer. Repository maintenance tooling: [decision 0022](../decisions/0022-repository-maintenance-tooling.md).

## Constraints

Two hard requirements from the maintainer shape everything here. The skill must be general enough to run on all major agent CLIs, explicitly including Antigravity and ollama-backed agents. Every script and program must run on macOS, Linux, and Windows. The research established that no interpreter is guaranteed across the six target hosts: Claude Code and Codex are native binaries that ship neither Node nor Python, Gemini CLI requires Node, Hermes bundles its own Python and Git Bash, Antigravity is a native binary, and a stock Windows machine has neither Python nor Node. On Windows, Claude Code's exec-form hooks require a real `.exe` and cannot spawn npm `.cmd` shims (<https://code.claude.com/docs/en/hooks>), Codex runs hooks in cmd or PowerShell with a `commandWindows` override (<https://developers.openai.com/codex/hooks>), and `claude plugin eval` refuses Bash on native Windows (<https://code.claude.com/docs/en/plugin-evals>). Hooks fire on every tool call in parallel, so startup time matters; measured Node hooks sit near a 43 ms floor and 139 ms p95, compiled binaries near 4 to 12 ms.

Consequences: no bash on the required path; the deterministic core is a single static Go binary per OS and architecture; the skill text is plain and portable; host-specific mappings live in reference files.

## Layered layout

The principle is separation of three layers so that later forms are wrappers, not rewrites: the contract and report schema (data, versioned, host-independent); the deterministic core (library plus CLI); and the judgment layer (a thin, host-specific prompt or persona).

```text
v-eval/
  schema/                    exists: report JSON Schema v0.1.0, versioned, embedded in the core
  core/  cmd/veval/          exists: Go library and CLI with validate, aggregate, render, export sarif;
                             classify, adapters, and detectors are still to come
  internal/version/          exists: version and build digest stamped at link time
  skills/evaluate-output/    exists: SKILL.md and references/hosts/ (six per-harness mappings);
                             scripts/ (thin launchers) not needed so far
  .github/workflows/         exists: go.yml (three-OS matrix) and docs.yml
  .goreleaser.yaml           exists: six operating-system and architecture targets
  agents/  profiles/         later: evaluator persona and isolation profile (rung 3)
  .claude-plugin/ .codex-plugin/ commands/ hooks/    later: plugin manifests, slash command, hooks (rung 4)
  server/                    later: MCP server exposure (rung 5)
  evals/  anchors/           later: regression cases; locked labeled set kept out of the public tree
  tools/                     exists: maintainer-only tooling, never shipped (decision 0022)
  docs/                      exists
```

## Growth rungs

| Rung | Form | Adds | Status |
| --- | --- | --- | --- |
| 1 | Skill: SKILL.md, references, scripts | Judgment procedure, portable to any Agent Skills host | In the first release |
| 2 | Skill plus core CLI | Reproducible rollup, JSON report, HTML, SARIF, adapters, detectors | In the first release |
| 3 | Agent definition plus isolation profile | Runs as an independent subagent with restricted tools and a different model family than the builder, modeled on evolve-loop's evaluator profile | In the first release |
| 4 | Plugin | Slash command, hooks that refuse "done" without a valid report, installable on Claude Code and Codex, testable by `claude plugin eval` | In the first release |
| 5 | Service: MCP or HTTP | Sandboxed execution, persisted baselines, callable from CI and from tool-call-only harnesses | Later |

The maintainer chose to ship rungs 2 to 4 together. Rung 5 is planned because ollama cannot run scripts at all; it exposes only model and tool-call APIs and delegates to a harness (<https://docs.ollama.com/cli>), so an MCP surface is the way an ollama-backed agent reaches the core.

## Binary distribution

The plugin stays source-only; binaries are never committed, and claude.ai rejects plugins with a top-level `bin/` directory (<https://code.claude.com/docs/en/plugin-marketplaces>). The pattern:

1. GoReleaser builds the GOOS and GOARCH matrix with `CGO_ENABLED=0` and publishes a sha256 checksums file with every tagged release (<https://goreleaser.com/customization/checksum>). evolve-loop already ships this way.
2. A SessionStart hook downloads `veval_<os>_<arch>` into the plugin data directory (`${CLAUDE_PLUGIN_DATA}` on Claude Code; the equivalent per host) using in-box tools: `curl` and `tar` on macOS, Linux, and Windows builds since 17063. The hook is written twice, a shell form for Unix and a PowerShell or `commandWindows` form for Windows, and does nothing else.
3. The hook verifies the checksum before marking the binary usable and records the version.
4. Every other hook, script, and command invokes the binary directly in exec form with arguments, never through a shell, so it works on Windows without Git Bash.
5. A field-observed failure mode is avoided: gitignored binaries broke marketplace installs for at least one plugin, fixed by exactly this download-on-first-run pattern.

**Proposal**: the binary location and version are also discoverable through `veval --print-self`, so a harness that already has the binary on PATH can skip the download.

## Windows rules

- No bash on the required path. Shell scripts may exist as conveniences on Unix only.
- Hook definitions use exec form with a real executable; `.cmd` and `.bat` shims are not used.
- Paths are handled by the core, never by shell string manipulation; line endings are normalized at intake.
- `claude plugin eval` suites that need shell access run under WSL2 in CI; the deterministic report-grading cases need none.
- Windows is in the CI matrix from the first commit.

## Per-harness reference files in the skill

The skill speaks in actions; a reference file per harness maps actions to that harness's tools, following the pattern superpowers uses for Codex, Gemini, Antigravity, Hermes, and Pi. SKILL.md itself contains no Claude-only tool names on the required path and stays within the Agent Skills spec (<https://agentskills.io/specification>: `name`, `description`, optional `license`, `compatibility`, `metadata`, body under roughly 500 lines, `scripts/`, `references/`).

The mapping tables now live with the skill, one file per harness, under [`skills/evaluate-output/references/hosts/`](../../skills/evaluate-output/references/hosts/): `claude-code.md`, `codex.md`, `gemini-cli.md`, `antigravity.md`, `hermes.md`, and `ollama.md`. Each file names its source and its read date, and says which cells the source did not name rather than guessing. The table that stood here, with cells marked "fill in", is superseded by those files.

Where a harness lacks a capability, the skill applies the rule's intent with what exists, as the evo:fable overlay already does: a fresh subagent becomes a structured self-pass, background monitoring becomes explicit re-checks. Where a reference file and the harness's actual tool list disagree, the tool list wins and the tool actually used is recorded in the report's provenance.

## Multi-CLI reference: evolve-loop's bridge

evolve-loop's Go bridge already drives `claude`, `claude-p`, `claude-tmux`, `codex`, `codex-tmux`, `agy`, `agy-tmux`, `gemini`, `ollama`, `ollama-tmux`, and `antigravity` identifiers, with per-CLI preflight rules such as permission modes being Claude-only ([local prior art](../research/2026-09-14-local-prior-art.md)). v-eval's rung 3 agent and the regression runner in [evaluating-v-eval.md](evaluating-v-eval.md) should reuse or mirror that driver set rather than invent a new one; this is a **proposal** pending a look at whether the bridge can be imported as a package.

## CI matrix

From the first commit: macOS, Linux, and Windows runners; Go build and tests with the race detector on all three; GoReleaser dry run producing every target; markdownlint and a link checker for the docs; the deterministic report-grading regression cases on all three OSs; the plugin-eval suite on Linux and, for shell-granting cases, WSL2. A release is blocked if any OS fails.

## Proposals awaiting maintainer confirmation

- `veval --print-self` and the exact plugin data directory per host.
- Reusing evolve-loop's bridge as an imported package versus mirroring its driver list.
- Whether the Codex plugin manifest ships in the first release or after the Claude Code plugin is validated.
