# Start at rungs 2 to 4 together: skill, Go core CLI, agent definition, plugin

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Implemented in part on 2026-09-15: rungs 1 and 2 stand as `schema/`, `core/`, `cmd/veval/`, `internal/version/`, and `skills/evaluate-output/`. Rungs 3 and 4 -- `agents/`, `profiles/`, and the plugin manifests -- do not exist. This record fixes the starting form and the repository layout.

## Context and Problem Statement

The maintainer asked which structure lets the skill evolve into an agent or an evaluation service that adapts to wider cases without rewrites. Prior art splits into skills that carry judgment and procedure (evals-skills, obra/superpowers) and CLIs that carry reproducible execution and grading (smevals, `claude plugin eval`). No existing skill or tool emits per-criterion PASS / FAIL / UNKNOWN / ERROR / NOT_APPLICABLE with evidence. At which rung of the growth ladder should v-eval start, and in what layout?

| Rung | Form | Adds |
| --- | --- | --- |
| 1 | Skill: SKILL.md, references, scripts | Judgment procedure, portable across CLIs |
| 2 | Skill + core CLI | Reproducible rollup, JSON report, SARIF export, test-result import |
| 3 | Agent definition + isolation profile | Independent subagent, restricted tools, different model family than the builder |
| 4 | Plugin | Slash command, hooks, optional MCP server, installable on Claude and Codex, testable by `claude plugin eval` |
| 5 | Service: MCP or HTTP | Sandboxed execution, persisted baselines, locked anchor set, callable from CI |

## Decision Drivers

* Growth without rewrites requires three separated layers: versioned contract and report schema, deterministic core with a CLI, and a thin host-specific judgment layer; every later form wraps the same core.
* The evolve-loop prose-heading drift incident shows a model-derived rollup is fragile; software must own it.
* `claude plugin eval` only targets plugins with a manifest.
* The maintainer wants the agent and plugin forms available from the first release.

## Considered Options

* Rung 2: skill + core CLI in the layered layout
* Rung 1: skill only, add the core later
* Rungs 2 to 4 together: skill, agent, and plugin from day one
* Rung 5: service-first

## Decision Outcome

Chosen option: "Rungs 2 to 4 together", because the maintainer wants the isolation profile, command, and hooks from the first release, and the layered layout makes the service rung an additive folder later. Repository layout:

```text
v-eval/
  schema/                  contract and report JSON Schema, versioned
  core/  cmd/veval/        deterministic library and CLI (Go)
  skills/evaluate-output/  SKILL.md, references/ (per-harness tool maps), scripts/
  agents/  profiles/       agent definition and isolation profile
  .claude-plugin/ .codex-plugin/ commands/ hooks/
  server/                  later: MCP or HTTP service
  evals/  anchors/         skill regression cases; locked labeled set, never public
  docs/
```

### Consequences

* Good, because the plugin manifests make the skill testable under `claude plugin eval` ([0018](0018-skill-regression-testing.md)).
* Good, because the agent form gets an isolation profile modeled on evolve-loop's from the start.
* Good, because the service rung is a new folder, not a rewrite.
* Bad, because packaging work is front-loaded before any pilot case is evaluated.
* Bad, because hooks must already obey the no-shell, cross-OS rule ([0005](0005-go-core-binary.md), [0020](0020-portability-constraints.md)).

## Confirmation

The repository contains the listed folders; the skill's text references `veval` commands for validation, aggregation, and rendering; `claude plugin eval` can resolve the plugin. Implemented in part on 2026-09-15: the schema, core, command line, and skill folders exist, and `skills/evaluate-output/SKILL.md` names every `veval` subcommand, which `cmd/veval/skill_contract_test.go` fails on when it stops. No plugin manifest exists, so `claude plugin eval` has nothing to resolve.

## Pros and Cons of the Options

### Rung 2: skill + core CLI

* Good, because it is the smallest form where software owns pass/fail.
* Bad, because no isolation profile or manifests until later.

### Rung 1: skill only

* Good, because it is fastest to publish and fully portable.
* Bad, because the model re-derives the rollup each run.
* Bad, because it is not testable by `claude plugin eval`.

### Rungs 2 to 4 together

* Good, because it is the most complete first release.
* Good, because agent and plugin wrappers reuse the same core.
* Bad, because it front-loads packaging.

### Rung 5: service-first

* Good, because it best serves CI and multi-host use.
* Bad, because it is the heaviest start and needs hosting and sandbox decisions now.

## More Information

* Related requirements: REQ-01, REQ-04, REQ-12, REQ-13; portability in [0020](0020-portability-constraints.md). Informed by [skill packaging research](../research/2026-09-14-skill-packaging.md) and [local prior art](../research/2026-09-14-local-prior-art.md).
* Agent Skills specification: <https://agentskills.io/specification> (accessed 2026-09-14).
* Claude Code plugins reference: <https://code.claude.com/docs/en/plugins-reference> (accessed 2026-09-14). Plugin evals: <https://code.claude.com/docs/en/plugin-evals> (accessed 2026-09-14).
* ai-evals-course/evals-skills, dual Claude and Codex manifests: <https://github.com/ai-evals-course/evals-skills> (accessed 2026-09-14).
* prime-radiant-inc/smevals: <https://github.com/prime-radiant-inc/smevals> (accessed 2026-09-14).
* evolve-loop evaluator profile, `runtime/.evolve/profiles/evaluator.json`, read locally 2026-09-14.
* Revisit when the pilot shows the agent or plugin rungs are unused, or a service form is needed for CI first.
