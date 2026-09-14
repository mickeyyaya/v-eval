# Evaluating v-eval itself

Status: design, 2026-09-14. Not implemented. Decisions: [decision 0015 pilot cases and labeling](../decisions/0015-pilot-cases-and-labeling.md), [decision 0018 skill regression testing](../decisions/0018-skill-regression-testing.md), [decision 0017 RSI governance](../decisions/0017-rsi-governance-human-promote.md), [decision 0020 portability constraints](../decisions/0020-portability-constraints.md). Research: [skill packaging memo](../research/2026-09-14-skill-packaging.md), [integrity and RSI memo](../research/2026-09-14-integrity-and-rsi.md), [evaluator learning memo](../research/2026-09-14-evaluator-learning.md). The comparison protocol, denominators, and fairness rules already written in [../comparison.md](../comparison.md) apply unchanged and are not repeated here. All sources accessed 2026-09-14. Items marked **proposal** were not decided by the maintainer.

## What is being measured

Four objects, kept separate as the handbook insists: the skill's behaviour on a host (did it fire, did it follow the procedure, did it write a valid report), the core's correctness (validation, aggregation, adapters, detectors, renderers), the evaluator's accuracy against independent labels (false acceptance, false rejection, abstention, evidence validity, forensic false positives), and the learning loop's effect over versions. Each has its own tests and its own denominators.

## Skill regression tests

The maintainer accepted `claude plugin eval` with ablation as a runner, with the condition that the tests must be adaptable across all LLM CLIs. The design therefore has one portable case format and several runners.

**Case format** (**proposal**): a directory per case holding `prompt.md` (the evaluation request), `inputs/` (the history bundle: brief, diff, tests, logs, sources), `expected.json` (expected per-criterion results, expected claim statuses, expected forensic findings, and required evidence locators), and `graders/` expressed as deterministic checks over the produced report JSON. The same case is expressible as a `claude plugin eval` case (`prompt.md` plus `graders/*.md` using `file_exists`, `regex`, and `tool_used: Skill`), and as an Agent Skills `evals.json` entry (prompt, files, assertions) per the skill-creation guidance (<https://agentskills.io/skill-creation/evaluating-skills>). A small generator in the core emits both from the portable case.

**Runners:**

1. `claude plugin eval` on Claude Code: three runs per case with and without the plugin, the with/without delta reported, `--threshold`, `--max-cost-usd`, JSON output with `schemaVersion: 1` (<https://code.claude.com/docs/en/plugin-evals>). The ablation delta shows whether the skill changes behaviour at all; the deterministic graders on the report file show whether it changed it correctly. Practitioners report the command may be early-access gated for some accounts, so it is one runner, not the only one.
2. Core-shipped runner: `veval selftest` drives any supported CLI headless (Claude Code `-p --output-format json`, Codex, Gemini, Antigravity, Hermes, or an ollama-backed agent through the MCP surface), applies the same graders to the produced report, and records host, model, and timings. This is the runner that satisfies the cross-CLI condition; it mirrors evolve-loop's bridge driver set ([packaging-and-portability.md](packaging-and-portability.md)).
3. Agent Skills `evals.json` for hosts that implement the skill-creator flow.

Grading is deterministic: the report validates against the schema; expected results, claim statuses, and forensic findings match; every PASS carries an admissible evidence locator; the HTML renders. An LLM grader is used only for a short summary field, if at all, because judging the judge with a judge adds the noise the plugin-eval documentation itself warns about.

## Core tests

Ordinary Go tests with the race detector on macOS, Linux, and Windows: schema validation against valid and deliberately invalid reports; aggregation against a table of criterion-result combinations including zero applicable criteria, provisional criteria, and unresolved conflicts; each adapter against fixtures for pass, fail, missing input (must yield UNKNOWN), and operational failure (must yield ERROR); each detector against a tampered and a benign fixture; renderers against golden files; the SARIF export against the schema. The golden-file contract tests are the drift alarm evolve-loop lacked.

## The pilot set

Source, per the maintainer: evolve-loop history plus synthetic variants. Cases are mined from real evolve-loop cycles, including the tautological-grader incidents, the prose-heading classifier drift, stale-log and skipped-test cases, and ordinary passing cycles as benign controls; each is reduced to a minimal reproduction with the intent, artifact, criteria, evidence, and outcome. Synthetic variants are added per base case: a subtle violation, missing or conflicting evidence, and a valid alternative implementation, following the shape already described in [../comparison.md](../comparison.md). Every case is tagged with its origin, artifact kind, and the failure modes it exercises.

The set is never public. It is stored encrypted with a canary item whose perfect score would prove leakage, following the mitigations Anthropic used after finding a model decrypting benchmark answers (2026-03-06, <https://www.anthropic.com/engineering/eval-awareness-browsecomp>). Cases are assigned to train, dev, and test splits deterministically by hash at creation ([learning-loop.md](learning-loop.md)). OpenAI's audits found 59.4% of hard SWE-bench Verified items and about 30% of SWE-bench Pro items had flawed tests or prompts (2026-02-23, <https://openai.com/index/why-we-no-longer-evaluate-swe-bench-verified/>; 2026-07-08, <https://openai.com/index/separating-signal-from-noise-coding-evaluations/>), so every case that v-eval fails is audited for a case defect before it counts against v-eval.

## Labeling protocol

The maintainer is the sole accountable labeler, with a disclosed second opinion from a model of a different family. The second label is produced blind: the model sees the case and the contract, not the maintainer's label and not v-eval's result. Disagreements are adjudicated by the maintainer and the original labels are kept. Labels are binary per criterion with a rationale, following the Husain and Shankar guidance of 100 to 200 labels per failure mode and a split of roughly 10 to 20 percent train, 40 to 45 dev, 40 to 45 test (updated 2026-09-01, <https://hamel.dev/blog/posts/evals-faq/>). An indeterminate label is allowed where the contract or evidence is insufficient and is kept out of the accept and reject denominators. This arrangement is a feasibility protocol; published accuracy claims would need a second human labeler, and the reports say so.

## The anchor set

The test split plus additional labeled cases form the locked anchor set. It is re-scored on every v-eval version and before any promotion; the learner, retrieval, and tuning never read it. Its purpose is drift attribution: if anchor scores move while the artifact population did not, the evaluator changed (Li, 2026-06-13, <https://arxiv.org/abs/2606.15474>; Zhang et al., 2026-07-14, <https://arxiv.org/html/2607.12790v1>). Anchor results are published per version with sample sizes.

## Metrics, with denominators

Let U be labeled unacceptable cases, A acceptable, X indeterminate, N their sum, as in [../comparison.md](../comparison.md).

| Metric | Definition | Reported with |
| --- | --- | --- |
| False acceptance | accepted cases in U / U | raw counts, Wilson interval |
| False rejection | rejected cases in A / A | raw counts; INCOMPLETE in A shown separately |
| Abstention | INCOMPLETE cases / N, also within A, U, X | raw counts |
| Decided-case accuracy | correct accept or reject among A and U / decisions among A and U | always beside abstention |
| Evidence validity | PASS results whose cited evidence, when independently opened, supports the result / PASS results | audited sample size |
| Claim-table completeness | candidate claims enumerated / candidate claims present in the bundle | per case |
| Forensic false-positive rate | benign-control cases with a `confirmed` finding / benign-control cases | per detector |
| Forensic detection rate | seeded-tamper cases with at least a `suspicious` finding / seeded cases | per detector |
| Cost and time | tokens, tool calls, wall-clock per case | median and tail |

A zero denominator is undefined, not zero. Paired before-and-after comparisons across versions use McNemar's test; at fifty cases a rate carries roughly plus or minus eleven points, so changes under ten points are reported as unproven. Accuracy among decided cases is never shown without the abstention rate, so a version cannot look better by refusing hard cases.

## Cadence

Core tests and deterministic skill cases run on every commit on all three OSs. The plugin-eval suite and the cross-CLI runner run on every release candidate and nightly at a capped cost. The anchor set is re-scored on every version and before any learned material is promoted. Results are appended to a versioned results log in the repository, with the case set digest, so a reader can see which cases a number was computed on.

## Proposals awaiting maintainer confirmation

- The portable case directory layout and the name `veval selftest`.
- The nightly cadence and its cost cap.
- Whether cross-CLI runs on hosts without headless modes are attempted or recorded as not run.
