# Roadmap

Status: proposed sequence, reconciled with the decisions of 2026-09-14. No dates are implied. Each stage has an exit condition that must be demonstrated, not asserted. Decision records are under [docs/decisions](docs/decisions/README.md); designs under [docs/architecture](docs/architecture/README.md).

## Stage 0. Foundation (complete, 2026-09-14)

- Cited landscape research, practitioner review, deterministic-tools survey, standards register, and the 2026-09-14 research reports.
- Consolidated requirements with sources U1 to U19.
- Twenty-two decision records in MADR format.
- Architecture design: pipeline, report schema, learning loop, forensics, packaging and portability, evaluating v-eval.
- Draft skill (draft-2), contract template, worked example, community files, documentation workflow.

Exit condition met: every decision has a record, every design links to its decision and research, and the repository states what is not implemented.

## Stage 1. Walking skeleton

Done:

- `schema/`: report JSON Schema v0.1.0 from [the report schema design](docs/architecture/report-schema.md), embedded in the core. The contract is the report schema's `contract` section rather than a file of its own ([decision 0024](docs/decisions/0024-contract-inside-report-schema.md)).
- `core/` and `cmd/veval/`: Go binary with `version`, `validate`, `aggregate`, `render` (Markdown, HTML), and `export sarif`; the evidence-shape rule for PASS; the five-state rollup with its rule recorded; counts and coverage.
- Renderers: Markdown and a self-contained HTML file that reads offline in light and dark themes.
- SARIF 2.1.0 export. The two golden logs were validated by the maintainer on 2026-09-15 with an external JSON Schema validator against the OASIS errata01 schema; no SARIF validator runs in CI yet.
- `skills/evaluate-output/`: skill draft-3 writes the JSON report and calls the core when present; per-harness reference files for Claude Code, Codex, Gemini CLI, Antigravity, Hermes, and ollama-backed agents, under `skills/evaluate-output/references/hosts/`.
- Continuous integration on macOS, Linux, and Windows; GoReleaser configuration for the six operating-system and architecture targets.
- First official release, v0.1.0 (2026-09-15), cut by tag through `release.yml` with binaries for six targets, checksums, notes from the changelog, and the two example reports rendered by the released binary ([decision 0025](docs/decisions/0025-tag-driven-release.md)); the HTML report received a design pass first.

Remaining:

- `agents/` and `profiles/`: evaluator definition with a read-only, no-network isolation profile.
- `.claude-plugin/` and `.codex-plugin/`: manifests, a slash command, and a SessionStart hook that downloads the checksummed binary for the host OS and architecture into the plugin data directory.

Exit condition: the worked example is evaluated end to end on all three operating systems from at least two CLIs; the JSON validates; the HTML renders offline; a PASS without a qualifying citation is rejected by the core.

Exit condition status, stated honestly: the JSON validates, the HTML renders offline, and a PASS without a qualifying citation is rejected by the core, all demonstrated. The worked example has been evaluated end to end on macOS in the maintainer's own session, and the test suite covers the same path on all three operating systems in CI once `main` is green; that is one CLI, not two, so the exit condition is not met.

## Stage 2. Pilot set, anchors, and skill tests

- Mine evolve-loop cycle history for real cases, including the tautological-grader and classifier-drift incidents and ordinary passing cycles as benign controls; add synthetic variants (subtle violation, missing evidence, valid alternative implementation, ambiguous requirement).
- Label with the protocol in [decision 0015](docs/decisions/0015-pilot-cases-and-labeling.md); keep disagreement records; lock the anchor set, encrypted, with a canary, never in the public tree.
- Portable skill regression cases runnable under `claude plugin eval` and through the core's runner on other CLIs.

Exit condition: a labeled pilot set with published counts by origin and split; the skill-test runner drives at least two CLIs; a baseline false-acceptance, false-rejection, and abstention table exists with denominators.

## Stage 3. Adapters, classifier, and forensics

- Adapters: commit-bound test evidence; diff scope and integrity; claim-to-source support.
- Classifier: deterministic detection, versioned perspective profiles, provisional proposals, routing rationale in every report.
- Forensic pass: test and grader tampering; evidence provenance mismatch; hardcoding and special-casing.
- Isolation levels recorded on every executed check.
- Port the source-register generator from Python to an internal Go tool run with `go run ./tools/...` and delete the Python ([decision 0022](docs/decisions/0022-repository-maintenance-tooling.md)).

Exit condition: rerunning the pilot reproduces the reports byte-for-byte for deterministic parts; each detector catches its seeded cases without flagging the benign controls; results are compared with Stage 2's baseline.

## Stage 4. Learning loop and self-improvement governance

- Reward records for every human reaction; local precedent bank (SQLite, sqlite-vec, ollama embeddings); cold re-judge; retrieval of examples for similar cases with prior scores stripped.
- Self-improvement loop: auto-propose, auto-reject on the held-out gate, human-promote; budgets; append-only audit log.
- Measurement: paired before-and-after on the anchor set with intervals; false acceptance, false rejection, and abstention per version.

Exit condition: at least one promoted change shows a non-regressing anchor result and a measured improvement on held-out cases, with the reaction records that drove it.

## Stage 5. Extension

- Optional pinned per-criterion judge, validated on the anchor set before use.
- MCP service exposure of the core for tool-call-only harnesses.
- evolve-loop optional phase using the published verdict mapping.
- Rule ladder (candidate to active to disabled) on top of the precedent bank.
- Static-analysis SARIF import; document and generated-context profiles beyond the first adapters; judge-directed text and eval-awareness detectors.
- The comparison experiment in [docs/comparison.md](docs/comparison.md) against ordinary review and configured frameworks.

## Not planned

A hosted dashboard, a model leaderboard, an automatic repair loop, and a default composite score are not milestones. Expansion depends on measured demand and on the exit conditions above.
