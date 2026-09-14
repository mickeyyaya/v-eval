# Portability constraints: every major agent CLI, and macOS, Linux, and Windows

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Implemented in part on 2026-09-15: `.github/workflows/go.yml` builds and tests on macOS, Linux, and Windows; `skills/evaluate-output/references/hosts/` carries one mapping file per target harness; the core is a single dependency-free Go binary with no shell on any path. The MCP server and the hooks do not exist. This record promotes two maintainer statements to requirements.

## Context and Problem Statement

During the language discussion the maintainer stated two constraints: "the skill needs to be general enough to be adapted through all major CLIs, including antigravity and ollama and etc..." and "for the script and program, it must be portable across different OSs, such as MacOS, Linux, Windows." Neither was in the requirements document. What do they require of the skill, the core, hooks, and CI?

## Decision Drivers

* Target CLIs: Claude Code, Codex, Gemini CLI, Antigravity (`agy`), Hermes Agent, and ollama-backed agents. Their tool names, hook mechanisms, and script support differ; ollama cannot run scripts at all.
* Target OSs: macOS, Linux, Windows. Stock Windows has neither Python nor Node; Claude Code exec-form hooks need a real executable; Codex uses `commandWindows`; plugin eval refuses Bash on native Windows.
* Two local patterns already solve parts of this: obra/superpowers ships one SKILL.md plus per-harness tool-mapping reference files for Codex, Gemini, Antigravity, Hermes, and Pi; evolve-loop's Go bridge drives claude, codex, agy, gemini, antigravity, and ollama.

## Considered Options

* Promote both statements to requirements with concrete rules for skill, core, hooks, and CI
* Treat them as preferences and handle case by case
* Support Claude Code first, port later

## Decision Outcome

Chosen option: "Promote both statements to requirements with concrete rules". Rules:

* Skill: a plain Agent Skills SKILL.md with no host-specific tool names on the required path; host mappings live in `skills/evaluate-output/references/<harness>-tools.md` following the superpowers pattern; on hosts lacking a capability, the skill degrades to inspect-and-report-UNKNOWN rather than failing.
* Core: a single static binary per OS and architecture ([0005](0005-go-core-binary.md)); an MCP server for tool-call-only harnesses.
* Scripts and hooks: no bash on the required path; no POSIX-only assumptions (paths, line endings, `awk` or `grep` pipelines); exec-form invocation of the binary; `commandWindows` where a host requires it.
* CI: build and test on macOS, Linux, and Windows from the first commit; a hook smoke test per OS.
* Tests: the multi-CLI runner in [0018](0018-skill-regression-testing.md) exercises at least two CLIs in CI.

### Consequences

* Good, because every later decision has a portability test: does it still work on Codex, Gemini, Antigravity, Hermes, an ollama agent, and Windows?
* Good, because the reference-file pattern keeps SKILL.md within the Agent Skills size guidance.
* Bad, because Antigravity hook semantics and Codex's Windows hook dispatch are undocumented or have open bugs; some paths will be verified only by trying.
* Bad, because a three-OS CI matrix and per-harness reference files add maintenance.

## Confirmation

CI matrix includes windows-latest, macos-latest, ubuntu-latest; a lint rejects `#!/bin/bash` or `.sh` on the required path; `references/` contains one mapping file per target harness; the skill passes the Agent Skills `skills-ref validate`. Implemented in part on 2026-09-15: `.github/workflows/go.yml` runs on windows-latest, macos-latest, and ubuntu-latest, and `cmd/veval/skill_contract_test.go` fails when a host mapping under `skills/evaluate-output/references/hosts/` goes missing. No shell-script lint runs, because no script exists on the required path to lint; the skill has not been put through `skills-ref validate`.

## Pros and Cons of the Options

### Promote to requirements with concrete rules

* Good, because portability is enforced, not hoped for.
* Bad, because more CI and files.

### Preferences, case by case

* Good, because less upfront work.
* Bad, because portability regressions arrive silently.

### Claude Code first, port later

* Good, because fastest first release.
* Bad, because porting a shell-dependent design later is a rewrite.

## More Information

* Related requirements: REQ-01, REQ-04, and REQ-35 and REQ-36 in [requirements](../requirements.md). Informed by [language and portability research](../research/2026-09-14-language-and-portability.md) and [local prior art](../research/2026-09-14-local-prior-art.md).
* Agent Skills specification: <https://agentskills.io/specification> (accessed 2026-09-14). Gemini CLI skills: <https://geminicli.com/docs/cli/skills> (accessed 2026-09-14). Antigravity CLI plugins: <https://antigravity.google/docs/cli/plugins> (accessed 2026-09-14). Hermes skills: <https://hermes-agent.nousresearch.com/docs/user-guide/features/skills> (accessed 2026-09-14).
* Claude Code hooks (Windows exec form): <https://code.claude.com/docs/en/hooks> (accessed 2026-09-14). Plugin evals (native Windows): <https://code.claude.com/docs/en/plugin-evals> (accessed 2026-09-14). Codex hooks: <https://developers.openai.com/codex/hooks> (accessed 2026-09-14).
* obra/superpowers per-harness reference files (`references/antigravity-tools.md`, `hermes-tools.md`, and others), read locally 2026-09-14; evolve-loop bridge drivers, `go/internal/bridge/`, read locally 2026-09-14.
* Revisit when a target CLI changes its skill or hook contract, or when Windows CI cannot exercise a required path.
