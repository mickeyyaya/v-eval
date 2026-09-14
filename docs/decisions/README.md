# Decision records

This directory holds v-eval's architecture decision records in the [MADR](https://adr.github.io/madr/) format. Each record states the context, the options considered with their tradeoffs, the maintainer's decision, its consequences, how compliance will be confirmed, and the evidence behind it. New records start from [adr-template.md](adr-template.md) and take the next number.

Records 0001 to 0022 were decided on 2026-09-14 in a maintainer session; 0023 and 0024 were decided on 2026-09-15 while the core walking skeleton was built, and 0025 at the first release the same day. Each record states what of it is built in a dated line under its title and again in its Confirmation section, and amends its own text in a dated Amendments section where the building settled something differently. Where a record says neither, nothing it describes is implemented.

| ID | Title | Status | Date | Related requirements |
| --- | --- | --- | --- | --- |
| [0001](0001-independent-of-evolve-loop.md) | v-eval is an independent product; evolve-loop adopts it later | accepted | 2026-09-14 | REQ-01, REQ-04, REQ-24, REQ-25 |
| [0002](0002-no-composite-score.md) | No composite score; verdicts plus native dimension scores | accepted | 2026-09-14 | REQ-06, REQ-07, REQ-08, REQ-09 |
| [0003](0003-first-slice-code-change.md) | First implemented slice is a code change against intent, design, and tests | accepted | 2026-09-14 | REQ-02, REQ-03, REQ-05 |
| [0004](0004-start-at-rungs-2-to-4.md) | Start at rungs 2 to 4 together: skill, Go core CLI, agent definition, plugin | accepted | 2026-09-14 | REQ-01, REQ-04, REQ-12, REQ-13 |
| [0005](0005-go-core-binary.md) | Go core binary, cross-compiled, with thin adapters | accepted | 2026-09-14 | REQ-01, REQ-04; portability |
| [0006](0006-json-first-report-contract.md) | JSON-first report contract; Markdown and HTML are renders; SARIF export | accepted | 2026-09-14 | REQ-09, REQ-13, REQ-21 |
| [0007](0007-verdict-vocabulary.md) | Five per-criterion states and a three-state overall result, with published mappings | accepted | 2026-09-14 | REQ-09, REQ-12, REQ-13 |
| [0008](0008-evidence-policy-verify-over-summary.md) | Evidence policy: never trust a summary; PASS requires an opened-or-ran citation | accepted | 2026-09-14 | REQ-05, REQ-13; REQ-28 |
| [0009](0009-deterministic-intake-classifier.md) | Deterministic intake classifier with profile rules; model only for ambiguity | accepted | 2026-09-14 | REQ-02, REQ-05, REQ-10; REQ-32, REQ-33 |
| [0010](0010-first-adapters.md) | First adapters: commit-bound test evidence, diff-scope and integrity, claim-to-source support | accepted | 2026-09-14 | REQ-02, REQ-03, REQ-05, REQ-13 |
| [0011](0011-graded-isolation-detective-stance.md) | Execution isolation is graded and recorded, not mandatory; the evaluator acts as a detective | accepted | 2026-09-14 | REQ-05, REQ-13, REQ-14; REQ-29, REQ-31 |
| [0012](0012-forensic-detectors-v1.md) | Forensic detectors in v1: tampering, provenance mismatch, hardcoding | accepted | 2026-09-14 | REQ-14; REQ-29 |
| [0013](0013-evaluator-not-auditor.md) | v-eval is an evaluator, not a verdict auditor: evidence first, verdicts derived | accepted | 2026-09-14 | REQ-09, REQ-11, REQ-13; REQ-30 |
| [0014](0014-adaptive-learning-precedent-bank.md) | Adaptive learning: a local precedent bank with a guardrail skeleton; rule ladder later | accepted | 2026-09-14 | REQ-22, REQ-23, REQ-24; REQ-37, REQ-38 |
| [0015](0015-pilot-cases-and-labeling.md) | Pilot cases from evolve-loop history plus synthetic variants; sole labeler with a blind second opinion | accepted | 2026-09-14 | REQ-20, REQ-22, REQ-24 |
| [0016](0016-no-separate-judge-v1.md) | No separate LLM judge in v1: the host assistant inspects, the core decides | accepted | 2026-09-14 | REQ-04, REQ-13 |
| [0017](0017-rsi-governance-human-promote.md) | Self-improvement loop: auto-propose, auto-reject, human-promote; human reactions are the reward | accepted | 2026-09-14 | REQ-22, REQ-23, REQ-24 |
| [0018](0018-skill-regression-testing.md) | Skill regression testing: plugin eval with ablation, plus a portable case format for every CLI | accepted | 2026-09-14 | REQ-04, REQ-22, REQ-24 |
| [0019](0019-delivery-public-repo-mit.md) | Delivery: public GitHub repository, MIT license, documents-first initial commit | accepted | 2026-09-14 | REQ-01, REQ-26, REQ-27 |
| [0020](0020-portability-constraints.md) | Portability constraints: every major agent CLI, and macOS, Linux, and Windows | accepted | 2026-09-14 | REQ-01, REQ-04; REQ-35, REQ-36 |
| [0021](0021-html-report-every-evaluation.md) | Every evaluation renders a clean, self-contained HTML report | accepted | 2026-09-14 | REQ-09, REQ-19; REQ-34 |
| [0022](0022-repository-maintenance-tooling.md) | Repository maintenance tooling lives under tools/ and is exempt from the product's no-interpreter rule; lychee is the CI link authority | accepted | 2026-09-14 | REQ-21, REQ-36, REQ-39 |
| [0023](0023-note-locator-shape.md) | A fourth locator shape, `note`, for evidence that was neither opened nor run | accepted | 2026-09-15 | REQ-05, REQ-13; REQ-28 |
| [0024](0024-contract-inside-report-schema.md) | The evaluation contract is a section of the report schema, not a schema of its own | accepted | 2026-09-15 | REQ-09, REQ-13, REQ-21 |
| [0025](0025-tag-driven-release.md) | Releases are cut by tag, built by GoReleaser in CI, and ship the rendered example reports | accepted | 2026-09-15 | REQ-34, REQ-36 |

Requirements stated by the maintainer on 2026-09-14 carry identifiers REQ-28 to REQ-39 in [requirements.md](../requirements.md). Research memos referenced by the records live in [../research/](../research/) with the `2026-09-14-` prefix.
