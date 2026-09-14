# Evaluation-report schemas: adopt or define

Research date and access date: 2026-09-14.

Method: Web-research agent. Eight source groups: the structured-evaluation repository files (fetched and read), Every Eval Ever, SARIF 2.1.0, OpenTelemetry GenAI conventions, Inspect eval logs, promptfoo outputs, DeepEval test cases, and 2026 standardization papers.

Facts, inferences, single-source claims, and unverified items are labeled inside the memo; a label applies to the sentence or row it sits in.

---

Date: 2026-09-14. All URLs accessed 2026-09-14. Facts are from fetched repo files/specs unless marked **INFERENCE** or **UNVERIFIED**.

## 1. PlexusOne `structured-evaluation`

**Identity.** github.com/plexusone/structured-evaluation. Go 1.24, MIT. Created 2026-01-26; last push 2026-08-15; tags v0.1.0..**v0.14.0** (released 2026-08-15). **0 stars, 0 forks, 0 open issues.** Contributors: `grokify` (123 commits) + dependabot (1); recent commits are `Co-Authored-By: Claude Opus 4.8`. Self-description: "A reusable evaluation framework for LLM-as-Judge and multi-agent workflows." Docs: plexusone.dev/structured-evaluation.

**What it contains.** Packages `rubric/` (rubric definitions + reports), `claims/` (fact-verification), `summary/` (GO/NO-GO), `combine/` (DAG aggregation of agent results), `render/` (terminal/markdown/box/html), `schema/` (generated JSON Schema), `cmd/sevaluation` (CLI), `ts/` (Zod). `go.mod` depends only on `invopop/jsonschema`, `cobra`, `yaml.v3` — **no LLM client, HTTP, or collector**. It is schema + decision logic + lint + rendering; the judge/evaluator is external.

**Core types (rubric/rubric.go, category.go, finding.go, criteria.go, report.go).**

- `RubricSet{ID, Name, Version, EvaluationType(analytic|holistic), PassCriteria RubricPassCriteria, Categories []Category, JudgePromptTemplate, JudgeInstructions []string, Metadata}`
- `Category{ID, Name, Description, Weight, Required, Class CriterionClass, Blocking bool, Evaluation EvaluationMethod, Scale{Type categorical|checklist|binary|likert,...}, EvaluationPrompt, Examples, Criteria []Criterion}`
- `Criterion{ID, Name, Weight, Class, Blocking, Evaluation, Pass/Partial/Fail CriterionLevel{Description, Indicators []string}}`
- `CriterionClass` = `leadership_principle | specification_quality | implementation_readiness | deterministic_integrity`; `EvaluationMethod` = `deterministic | semantic | human` (v0.14.0). Invariant INV-3: `leadership_principle` must not be `Blocking`.
- Report `Rubric{SchemaVersion "v2", Metadata ReportMetadata, ReviewType, Judge *JudgeMetadata, RubricID, RubricVersion, IntScore(1-5), Confidence(0-1), Pass bool, Blocking []ReasonCode, Categories []CategoryResult, Findings []Finding, PassCriteria, Decision, OverallDecision, NextSteps, Summary, Extensions map[string]any}`
- `CategoryResult{Category, Score ScoreValue, IntScore, NumericScore, Confidence, Severity, ReasonCodes, Reasoning, Evidence []string, Findings, ChecklistResults}`
- `Finding{ID, Category, Code ReasonCode, Severity, Title, Description, Recommendation, Location string, Evidence string, Owner, Effort}`; `Severity` = `critical|high|medium|low|info` (critical/high are blocking).

**Verdict vocabulary.** Per-category `ScoreValue` is **only `pass | partial | fail`** (`schema/enums.json`). There is **no UNKNOWN, ERROR, or NOT_APPLICABLE** at criterion level; `lint` rejects any other value (`INVALID_ENUM`). `DecisionStatus` (report level) = `pass | conditional | fail | human_review`. Uncertainty is modelled as `Confidence` float (+ `HasLowConfidence`, default 0.7) and `human_review`. Other packages: `claims.Verdict` = `verified | unverified | needs-review | rejected`; `summary.Status` = `GO | WARN | NO-GO | SKIP` (SKIP is the nearest analogue to NOT_APPLICABLE); `combine.AgentResult.Error string` is the only error slot.

**Missing evidence.** `Evidence` fields are optional free strings; the schema does not distinguish "no evidence" from "negative evidence". The only treatment is the v0.14.0 `JudgeInstructions` example string "Distinguish missing evidence from negative evidence" — a prompt hint, not a field. Rubric-report lint checks enums, `metadata.document`/`reviewType`, finding titles, count accuracy and decision consistency; it does **not** require evidence. (The claims package *does* lint evidence: `verified-requires-url/quote`, corroboration, staleness.)

**Aggregation (`EvaluateResults`, report.go).** critical/high findings > `maxFindingsSeverity` → `fail`; `minCategoriesPassing` (`all` | `all_required` | N) → `fail`; medium over limit → `conditional` (Passed=false); any `partial` → `conditional` (Passed=true); `MinIntScore` gate; `IntScore` = weighted mean of category IntScores; `Confidence` = min across categories. Summary: `NO-GO > WARN > GO > SKIP`. **INFERENCE:** I found no code path in `criteria.go`/`report.go` that consults `Category.Blocking`; v0.14.0 validates it (INV-3) but enforcement appears left to consumers.

**CLI (`docs/cli/commands.md`).** `sevaluation render <file> --format={terminal|markdown|detailed|box|json|html}`; `lint <file> [--strict] [--format=json]` (exit 1 on errors); `check <file>` (exit 0 pass / 1 fail-or-conditional); `validate <file>`; `schema generate|show`; `version`. Report type auto-detected from top-level `categories|teams|claims`.

**TypeScript/Zod.** npm `@plexusone/structured-evaluation` (version-locked to Go tag), generated from the Go-emitted JSON Schema, `.strict()`. Known limitation (ts/README.md): the JSON Schema declares **no `required` fields** (confirmed: `rubric.schema.json` has none), so every field is `.optional()`.

**Other observations.** `examples/evaluation-report.json` uses stale pre-v0.4 field names (`score: 9.0`, `max_score`, `status: "needs_improvement"`, snake_case) that would fail current lint. 14 minor releases in 7 months.

## 2. Artifact identity and provenance in structured-evaluation

grep over `rubric/ summary/ combine/ claims/ schema/` for `commit|sha256|hash|provenance|exit code|environment|revision|digest|fingerprint` found **no artifact-identity or execution-provenance fields**. What exists:

- `ReportMetadata{Document, DocumentID, DocumentTitle, DocumentVersion, GeneratedAt, GeneratedBy, ReviewerID}` — free-form strings; no commit or content hash.
- `JudgeMetadata{JudgeID, Model, ModelProvider, ModelVersion, PromptTemplate, PromptVersion, SystemPrompt ("or hash/reference"), Temperature, MaxTokens, RubricID, RubricVersion, EvaluatedAt, Latency, TokensUsed, TraceID, SpanID}` — judge provenance only.
- `NextSteps.RerunCommand` — the command to run *next*, not what was run.
- `summary.TaskResult{ID, Status, Detail, DurationMs, Metadata map}`; `combine.AgentResult{AgentID, StepID, Inputs, Outputs, Tasks, Status, ExecutedAt, AgentModel, Duration, Error}` — no exit status or env.
- Escape hatch: `Rubric.Extensions map[string]any`.

## 3. EvalEval "Every Eval Ever" (EEE)

Announced 2026-02-17 (evalevalai.com blog; Batzner, Choshen, Ghosh et al.); paper arXiv:2606.14516 (June 2026). Repo `evaleval/every_eval_ever`: MIT, 114 stars, 49 forks, pushed 2026-09-10. Spec = `eval.schema.json` (schema_version 0.2.x; datastore README says 0.2.0, project page shows 0.2.2 — **discrepancy**) with required `schema_version, evaluation_id, retrieved_timestamp, source_metadata, model_info, eval_library, evaluation_results[] {metric_config{lower_is_better, score_type, min/max}, score_details}`; optional `detailed_evaluation_results{file_path, checksum}` → companion `instance_level_eval.schema.json` requiring `sample_id, interaction_type, input, answer_attribution, evaluation{score, is_correct}` plus `sample_hash`, `error`. Converters for Inspect, HELM, lm-eval; validation via Pydantic on HF PRs; datastore 22,235 models / 2,273 benchmarks; ACL 2026 shared task. **Establishes:** a real, adopted schema for *benchmark results per model*. **Does not establish:** fit for per-criterion verdicts on a code artifact — `model_info` is required, the verdict is `score`+`is_correct`, and there is no criterion/evidence/command provenance.

## 4. Other candidate formats

| Format | Spec URL | Per-criterion verdict | Evidence | Provenance / identity | 5-state fit |
|---|---|---|---|---|---|
| **SARIF 2.1.0** (OASIS Standard, errata01 2023-08-28) | docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/sarif-v2.1.0-errata01-os-complete.html | `result.ruleId` + `result.kind ∈ {pass, fail, notApplicable, open, review, informational}`, `level` | `message`, `locations`, `fingerprints`, `properties` bag | `invocation{commandLine, arguments, environmentVariables, exitCode, executionSuccessful, start/endTimeUtc, workingDirectory}`, `versionControlDetails{repositoryUri, revisionId, branch}`, `artifact.hashes` | PASS/FAIL/NOT_APPLICABLE native; UNKNOWN→`review`/`open`; ERROR→`executionSuccessful=false` + notifications (**UNVERIFIED** in this fetch: `toolExecutionNotifications` §3.20) |
| **OTel GenAI `gen_ai.evaluation.result` event** (Status: Development) | github.com/open-telemetry/semantic-conventions-genai/blob/main/docs/gen-ai/gen-ai-events.md | `gen_ai.evaluation.name` (Required), `score.value`, `score.label` (e.g. `pass`/`fail`) | `gen_ai.evaluation.explanation` (free text) | trace/span parent, `gen_ai.response.id`, `error.type`; evaluator identity **not yet standardized** — PR #359 (`evaluator.type/id/version`, `reference_set.id`) opened 2026-07-03, **still open** 2026-09-08; issue #185 (2026-05-21) proposes an `evaluation` span | Label is free-form (any 5 states); no artifact hash; telemetry, not a file format. Event added in semconv v1.38.0 (Oct 2025) per J. Hodge blog 2026-07-17 (**secondary source**) |
| **Inspect AI EvalLog** (v2; `.eval` zip or `.json`; `inspect log schema`) | inspect.aisi.org.uk/eval-logs.html; src/inspect_ai/log/_log.py | `samples[].scores{name: Score{value, answer, explanation, metadata, history}}`; value constants `C/I/P/N` (correct/incorrect/partial/noanswer) | `explanation`, `metadata`, messages/events | `eval: EvalSpec{run_id, task, task_version, task_file, model, model_args, config, revision: EvalRevision{type:"git", origin, commit, dirty}, packages}`; `status ∈ started\|success\|cancelled\|error`;`sample.error`,`limit` | N≈UNKNOWN, sample `error`≈ERROR; no NOT_APPLICABLE; `model` required (model-centric) |
| **promptfoo** results JSON (`version: 3`) | promptfoo.dev/docs/configuration/reference/ and /outputs/ | `gradingResult.componentResults[]{pass, score, reason, assertion{type,value}}` | `reason`, `metadata.renderedGradingPrompt` | `config` embedded (redacted), `provider`, `timestamp`, `traceId`; `failureReason 0 none/1 assertion/2 error`, `metadata.graderError` | ERROR distinct; no UNKNOWN/NOT_APPLICABLE; no artifact hash |
| **DeepEval** (`LLMTestCase` → `TestResult`/`MetricData`) | deepeval.com/docs/evaluation-test-cases; deepeval/tracing/api.py | one `MetricData{name, threshold, success, score, reason, strict_mode, flaky, error, evaluation_model, evaluation_cost, tokens, verbose_logs}` per metric | `reason`, `verbose_logs` | `evaluation_model`, cost; test-case fields `input, actual_output, expected_output, context, retrieval_context, tools_called, expected_tools, metadata, tags` | ERROR via `error`; no UNKNOWN/NOT_APPLICABLE; no commit/hash |

## 5. 2026 standardization discussion

- **EEE** (Feb/Jun 2026) and **Evaluation Cards** (arXiv:2606.09809, 2026-06-08, Ghosh, Reuel et al.) — same coalition; benchmark-reporting, includes a "provenance and risk" signal; not code-review oriented.
- **ReproEvalCard** (ACL 2026 short, Pattnayak & Bhatia) — checklist of artifacts (prompts, judge configs, traces) for reproducible LLM-pipeline evals; audit of 55 papers; a reporting standard, not a JSON schema.
- **Evidence Package Specification v0.1** (YenkLabs, 2026-06-29, draft) — envelope with `model.config_fingerprint sha256`, `source_blob_sha256`, `replay_hash`, `evidence_bundle.merkle_root`, `verification.policy_version`, `evidence_class ∈ failure|verified|disputed|benchmark_artifact|replay_trace`. Closest in spirit to "evidence + hash binding"; single-vendor, legal-citation first; adoption **UNVERIFIED**.
- **OTel semconv-genai** issues #79/#185 and PR #359 (May–Sep 2026) — active work on evaluator provenance and `test.suite.run.id`/`test.case.id`; unmerged.
- **SARIF for AI review**: Microsoft "Codename MDASH" delivers AI findings as SARIF + HTML (learn.microsoft.com FAQ; page date **UNVERIFIED**); small OSS tools adding SARIF output (mmlqm/ai-code-security-review, rdlugs/roborak#57). No 2026 proposal to replace SARIF for AI code-review findings was found.

## 6. Schema decision options

**A. Adopt structured-evaluation as-is.**
Pros: Go + generated JSON Schema + Zod; ready lint/check/render CLI; `Class`/`Evaluation(deterministic|semantic|human)`/`Blocking` taxonomy maps well to v-eval's deterministic-vs-judged criteria; severity/finding/decision logic already written; MIT.
Cons: verdict enum is `pass|partial|fail` — UNKNOWN/ERROR/NOT_APPLICABLE cannot be expressed per criterion without abusing `Confidence`/`human_review` or `Extensions`, and `lint` rejects other values; no commit/content hash or command/env/exit provenance; document-centric (`metadata.document` filename), not diff/commit-centric; 0 stars, single maintainer, AI-co-authored, 14 releases in 7 months, stale example, no `required` in schema, `Blocking` apparently unenforced (inference). Adoption risk is high.

**B. Define v-eval's own schema; ship exporters to SARIF (primary) and structured-evaluation (secondary).**
Pros: keeps the 5-state verdict, per-criterion evidence, artifact identity (commit + content hash) and execution provenance first-class. SARIF mapping is nearly lossless: PASS→`kind:pass`, FAIL→`fail`, NOT_APPLICABLE→`notApplicable`, UNKNOWN→`review`, ERROR→`executionSuccessful:false` + notification; provenance→`invocation.commandLine/exitCode/environmentVariables`, `versionControlDetails.revisionId`, `artifact.hashes`; gets GitHub code-scanning and MDASH-style ingestion. structured-evaluation export is lossy (UNKNOWN/ERROR → `human_review` + `Extensions`) but gives their renderers/`check`. Optional OTel `gen_ai.evaluation.result` emission for tracing.
Cons: two mappers to maintain; lossy paths must be documented; v-eval owns versioning.

**C. Own schema only.**
Pros: smallest surface now. Cons: no interop; likely re-derives SARIF's `invocation`/`versionControlDetails` anyway; no CI dashboard ingestion.

**Assessment (inference):** B. Nothing found carries per-criterion 5-state verdict + evidence + artifact/execution provenance natively; SARIF is the only mature, standards-body format that covers 4 of 5 states plus provenance, and structured-evaluation's shape is a poor fit for the missing states and is not yet adopted.

## 7. Gaps not resolved

- Did not run `sevaluation`; `Category.Blocking` enforcement is inferred from reading `criteria.go`/`report.go` only, not tests or renderers.
- OTel event's v1.38.0 introduction date rests on a secondary blog; SARIF `toolExecutionNotifications` not re-fetched.
- EEE current `schema_version` (0.2.0 vs 0.2.2) unresolved; no external adopters of structured-evaluation found (absence of evidence only).
- promptfoo top-level JSON layout (`results.results` vs `outputs`) taken from docs, not source types.
- MDASH FAQ publication date unverified.

## Sources (all accessed 2026-09-14)

1. plexusone/structured-evaluation — README, `rubric/*.go`, `summary/*.go`, `combine/aggregate.go`, `claims/verdict.go`, `schema/*.json`, `docs/cli/commands.md`, `docs/concepts/pass-criteria.md`, `docs/releases/v0.14.0.md`, `ts/README.md`, GitHub API (repo/tags/releases/contributors/commits). Author grokify; last commit 2026-08-15. Establishes types/CLI/status; does not establish external adoption.
2. EvalEval, "Every Eval Ever: Toward a Common Language for AI Eval Reporting", Batzner et al., 2026-02-17, evalevalai.com/infrastructure/2026/02/17/everyevalever-launch/; arXiv:2606.14516 (2026-06); github.com/evaleval/every_eval_ever (schemas). Establishes schema + adoption; not code-review fit.
3. OASIS, SARIF v2.1.0 errata01 complete, 2023-08-28, URL in §4.
4. OpenTelemetry semantic-conventions-genai, `docs/gen-ai/gen-ai-events.md`; PR #359 (2026-07-03, open); issue #185 (2026-05-21); issue #79. J. Hodge, "State of the OTel GenAI semconv (July 2026)", 2026-07-17 (secondary).
5. UK AISI Inspect, "Log Files" (inspect.aisi.org.uk/eval-logs.html); `src/inspect_ai/log/_log.py`, `scorer/_metric.py` (main).
6. promptfoo, "Configuration Reference" and "Output Formats" (promptfoo.dev/docs/configuration/reference, /outputs); site docs dated 2024-01-15 (outputs) — version may lag.
7. Confident AI DeepEval, "Single-Turn Test Case" (2026-09-01), `deepeval/tracing/api.py`, `deepeval/evaluate/types.py` (main).
8. Ghosh, Reuel et al., "Evaluation Cards", arXiv:2606.09809, 2026-06-08. Pattnayak & Bhatia, "ReproEvalCard", ACL 2026 short (2026.acl-short.22). YenkLabs, "Evidence Package Specification v0.1", 2026-06-29. Microsoft Learn, "Codename MDASH FAQ" (date UNVERIFIED).
