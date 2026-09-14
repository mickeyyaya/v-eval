# Skill regression testing: plugin eval with ablation, plus a portable case format for every CLI

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

How is v-eval's own skill regression-tested across releases? Claude Code 2.1.270 on the maintainer's machine ships `claude plugin eval` with a with/without ablation arm, cost ceilings, repeated runs, and JSON output; practitioners report it may be early-access gated for some accounts. The maintainer accepted the recommended option with a condition: "it should be able to be adapted across all LLM CLIs."

## Decision Drivers

* A first-party harness with ablation already exists and should be used where available.
* The skill must be testable on Codex, Gemini CLI, Antigravity, Hermes, and ollama-backed agents, not only Claude Code.
* Grading must be deterministic on the written report file; an LLM grading the judge is noise.
* Cases must be one canonical set, expressible in each harness's format.

## Considered Options

* `claude plugin eval` with ablation plus a portable evals.json, grading the report deterministically
* Own Go test harness only
* Manual runs on the worked example

## Decision Outcome

Chosen option: "`claude plugin eval` with ablation plus a portable case format", extended per the maintainer's condition. Canonical cases live in `evals/` in one format the Go core owns; generators emit `claude plugin eval` cases (prompt.md, case.yaml, graders) and Agent Skills `evals.json`. The Go core also ships a runner that can drive any CLI headless (`claude -p`, `codex exec`, Gemini CLI, `agy`, Hermes, and ollama-backed agents through a harness), using evolve-loop's bridge drivers as the reference implementation. Each run produces a JSON report; graders check that the skill fired, that the report validates, and that expected verdicts and evidence citations are present. Under plugin eval, `tool_used: Skill` is the fired indicator and regex or file_exists graders check the report.

### Consequences

* Good, because one canonical case set serves every harness.
* Good, because the with/without ablation delta is available on Claude Code at no extra cost.
* Bad, because the multi-CLI runner is real software to build and maintain, and per-CLI headless flags change.
* Bad, because plugin eval's gating status is unverified for this account; the portable path is the guaranteed one.

## Confirmation

`evals/` contains canonical cases; generators produce both target formats and a test round-trips them; a CI job runs the portable runner against at least Claude Code headless and one other CLI; results are JSON reports graded deterministically. None exists yet.

## Pros and Cons of the Options

### Plugin eval plus portable case format and multi-CLI runner

* Good, because reuses first-party ablation and cost controls.
* Good, because portable across CLIs and OSs.
* Bad, because two generators and a runner to maintain.

### Own Go test harness only

* Good, because fully portable and CLI-agnostic.
* Bad, because rebuilds ablation, cost caps, and reporting.

### Manual runs on the worked example

* Good, because it is what the repository does today.
* Bad, because no regression signal between releases.

## More Information

* Related requirements: REQ-04, REQ-22, REQ-24; portability in [0020](0020-portability-constraints.md). Informed by [skill packaging research](../research/2026-09-14-skill-packaging.md) and [language and portability research](../research/2026-09-14-language-and-portability.md).
* Claude Code plugin evals: <https://code.claude.com/docs/en/plugin-evals> (accessed 2026-09-14); local `claude plugin eval --help` on version 2.1.270, read 2026-09-14.
* Agent Skills evaluating-skills workflow and `evals.json`: <https://agentskills.io/skill-creation/evaluating-skills> (accessed 2026-09-14). skill-creator SKILL.md: <https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md> (accessed 2026-09-14).
* Agent SDK overview, driving the agent loop from another language via `claude -p --output-format json`: <https://code.claude.com/docs/en/agent-sdk/overview> (accessed 2026-09-14). Ollama CLI: <https://docs.ollama.com/cli> (accessed 2026-09-14).
* evolve-loop bridge drivers for claude, codex, agy, gemini, antigravity, and ollama, `go/internal/bridge/`, read locally 2026-09-14.
* prime-radiant-inc/smevals, runner and grader separation: <https://github.com/prime-radiant-inc/smevals> (accessed 2026-09-14).
* Revisit when plugin eval is confirmed available or gated for this account, or when a CLI changes its headless interface.
