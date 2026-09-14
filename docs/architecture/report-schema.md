# Report schema: the JSON-first contract

Status: design of 2026-09-14, implemented as schema v0.1.0. This is the human-readable specification; the machine-checkable JSON Schema lives at [`schema/report.schema.json`](../../schema/report.schema.json), is embedded in the core, and is versioned with it. Decisions: [decision 0006 JSON-first](../decisions/0006-json-first-report-contract.md), [decision 0007 verdicts](../decisions/0007-verdict-vocabulary.md), [decision 0008 evidence policy](../decisions/0008-evidence-policy-verify-over-summary.md), [decision 0013 evaluator-not-auditor](../decisions/0013-evaluator-not-auditor.md), [decision 0021 HTML report](../decisions/0021-html-report-every-evaluation.md). Field names are no longer proposals: the schema file exists and a drift test in `core/report` fails when the Go types and the schema disagree. Statements still marked **proposal** below are the ones the schema file did not settle.

## Why JSON first

evolve-loop derived phase verdicts by grepping headings out of prose reports; the headings drifted twice and no golden test caught it ([local prior art](../research/2026-09-14-local-prior-art.md)). The schema research found no existing format that carries a five-state per-criterion verdict, evidence, and execution provenance together, and found SARIF 2.1.0 the only mature standard covering four of the five states plus provenance ([report schemas memo](../research/2026-09-14-report-schemas.md)). So v-eval owns a schema, validates every report against it, and exports to SARIF; Markdown and HTML are renders.

## Top-level structure

The order below is also the render order. Observations and evidence come before derived results because the product is an evaluator, not an auditor.

| # | Section | Required | Purpose |
| --- | --- | --- | --- |
| 1 | `identity` | yes | What was evaluated, by which v-eval, from which bundle |
| 2 | `contract` | yes | The criteria and their status (user-specified, provisional, approved) |
| 3 | `routing` | yes | Classifier rationale: what was supplied, profiles used, adapters run and skipped |
| 4 | `observations` | yes | Everything found, including items tied to no criterion |
| 5 | `claims` | yes | Claim-to-verification table for every candidate claim |
| 6 | `criteria` | yes | Per-criterion results with evidence |
| 7 | `forensics` | yes | Detector findings mapped to integrity criteria |
| 8 | `dimensions` | no | Native-unit metrics with definitions and threshold sources |
| 9 | `counts` | yes | Raw counts and assessment coverage |
| 10 | `status` | yes | Derived overall status and the rule that produced it |
| 11 | `improvement` | yes | Agent improvement brief |
| 12 | `limitations` | yes | What was not inspected, not executed, or remains unknown |
| 13 | `provenance` | yes | Tools, versions, commands, environment, isolation levels, timestamps |
| 14 | `learning` | no | References to precedents retrieved and reward records created |

## Field tables

### identity

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `report_id` | string | yes | Content hash of the canonical JSON minus this field |
| `schema_version` | string | yes | Semver of the report schema |
| `veval_version` | string | yes | Core version and build digest |
| `skill_revision` | string | yes | SKILL.md digest |
| `artifact` | object | yes | `kind`, `revision` (commit plus dirty flag, or content hash), `paths[]`, `bundle_digest` |
| `task` | object | yes | `requested_outcome`, `intended_user`, `brief_ref` |
| `created_at` | RFC 3339 | yes | |
| `host` | object | yes | `cli`, `model` if exposed, `os`, `arch` |

### contract

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `contract_id` | string | yes | |
| `contract_version` | string | yes | Digest of the contract as evaluated |
| `status` | enum | yes | `user_specified`, `provisional`, `approved` |
| `conflicts[]` | object | yes | Each conflict between brief, design, and tests, with resolution or `unresolved` |
| `criteria[]` | object | yes | `id`, `requirement`, `source_ref`, `required` (bool), `applicability`, `methods_allowed[]`, `acceptance_rule`, `expected_evidence`, `provisional` (bool) |

### routing

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `supplied[]` | object | yes | Source type, count, origin label per context-model source |
| `profiles[]` | object | yes | Profile name and version applied |
| `adapters_run[]` | object | yes | Name, version, inputs used |
| `adapters_skipped[]` | object | yes | Name and the missing input |
| `ambiguity[]` | object | yes | Questions the rules could not answer and how they were resolved |
| `rationale` | string | yes | Plain-language summary a reader can audit |

### observations

Each entry: `id`, `text`, `evidence[]` (see evidence shape), `criterion_ids[]` (may be empty), `origin` (`assistant`, `adapter`, `detector`). No observation is dropped because it changes no verdict.

### claims

Each entry: `claim_id`, `text`, `location` (where the candidate said it), `status` enum `verified`, `contradicted`, `unverified`, `not_checkable`, `verification[]` (evidence records describing what was opened or run), `notes`. Every candidate claim detected at intake must appear here; a summary that is not enumerated is a schema violation.

### criteria

Each entry: `id`, `result` enum, `method_used` enum, `evidence[]`, `reasoning` (auditable justification, not private chain of thought), `next_action`, `shared_cause_with[]` (other criterion IDs exposing the same defect), `dimension_refs[]`.

### forensics

Each entry: `finding_id`, `detector`, `detector_version`, `criterion_id` (an integrity criterion), `severity` enum `observed`, `suspicious`, `confirmed`, `evidence[]`, `benign_alternative`, `disposition` (`open`, `explained`, `confirmed`). See [forensics.md](forensics.md).

### dimensions

Each entry: `dimension`, `metric`, `metric_version`, `authority_type` enum `formal_standard`, `established_measure`, `vendor_rating`, `project_rubric`, `definition_ref`, `unit`, `range`, `direction`, `value` (nullable; missing is null, never zero), `threshold`, `threshold_source`, `tool`, `tool_version`, `workload`, `evidence[]`, `interpretation`. No composite is computed ([decision 0002](../decisions/0002-no-composite-score.md)).

### counts and status

`counts`: `required` and `optional` objects each with `applicable`, `pass`, `fail`, `unknown`, `error`, `not_applicable`; `coverage` object with `numerator`, `denominator`, and `undefined` (true when denominator is zero).

`status`: `overall` enum `PASS`, `FAIL`, `INCOMPLETE`, `ADVISORY`; `rule_applied` string; `blocked_by[]` criterion IDs; `advisory` bool when the task requested review only.

### improvement, limitations, provenance, learning

`improvement[]`: `issue`, `locations[]`, `suggested_change`, `constraints_to_preserve[]`, `verify_by`, `linked_criteria[]`.

`limitations`: `not_inspected[]`, `not_executed[]`, `unknown_metadata[]`, `assumptions[]`.

`provenance`: `tools[]` (name, version), `commands[]` (command, cwd, exit_status, started_at, ended_at, log_ref, isolation), `environment` (os, arch, runtime versions), `isolation_levels_used[]`, `evidence_digest`.

`learning`: `precedents_retrieved[]` (precedent IDs, criterion, similarity), `reward_records_created[]`.

## Evidence shape

Every evidence record has `kind`, `locator`, `observation`, `provenance`, `isolation`, `origin`. The core enforces the following rule when computing results ([decision 0008](../decisions/0008-evidence-policy-verify-over-summary.md)):

- A criterion result of `PASS` requires at least one evidence record whose `origin` is `observed` and whose `locator` is one of: `{file, line_start, line_end}`, `{command, cwd, exit_status, log_ref}`, or `{passage, source_ref, source_date_or_version, access_date}`.
- Evidence with `origin` `candidate_supplied` may appear on any criterion but is never sufficient for `PASS` and is labeled as supplied in every render.
- Evidence with `kind` `judgment` must carry `rubric_version` and, when exposed, `model`; unknown metadata is recorded as unknown, not omitted.

The evidence hierarchy, strongest first, is: observed execution, direct inspection, supplied logs, candidate summary. Renders show the kind beside every evidence line.

## Enums

| Enum | Values |
| --- | --- |
| Criterion result | `PASS`, `FAIL`, `UNKNOWN`, `ERROR`, `NOT_APPLICABLE` |
| Overall status | `PASS`, `FAIL`, `INCOMPLETE`, `ADVISORY` |
| Method used | `execution`, `deterministic_check`, `static_inspection`, `source_verification`, `rubric_judgment`, `human_judgment` |
| Evidence kind | `execution`, `inspection`, `supplied`, `judgment` |
| Evidence origin | `observed`, `candidate_supplied`, `retrieved` |
| Isolation level | `none`, `worktree`, `container`, `remote_sandbox` |
| Claim status | `verified`, `contradicted`, `unverified`, `not_checkable` |
| Forensic severity | `observed`, `suspicious`, `confirmed` |
| Authority type | `formal_standard`, `established_measure`, `vendor_rating`, `project_rubric` |

`ADVISORY` as an overall status is a **proposal**: the existing skill says advisory-only reviews must be identified as advisory rather than claiming acceptance, and a distinct status makes that machine-checkable.

## Overall status derivation

For required, applicable, non-provisional criteria: any `FAIL` gives `FAIL`; otherwise any `UNKNOWN` or `ERROR` gives `INCOMPLETE`; otherwise `PASS` when at least one such criterion exists. Zero required applicable criteria, any provisional required criterion, or any unresolved contract conflict gives `INCOMPLETE`. A confirmed forensic finding on a required integrity criterion is a `FAIL` on that criterion and follows the same rule. The rule applied is written into `status.rule_applied` so the derivation is auditable.

## Mapping to SARIF 2.1.0

The export targets the OASIS SARIF 2.1.0 errata01 specification (<https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/sarif-v2.1.0-errata01-os-complete.html>, accessed 2026-09-14).

| v-eval | SARIF |
| --- | --- |
| Criterion | `tool.driver.rules[]` with `id` = criterion ID, `shortDescription` = requirement |
| `PASS` | `result.kind: pass` |
| `FAIL` | `result.kind: fail`, `level` from `required` |
| `NOT_APPLICABLE` | `result.kind: notApplicable` |
| `UNKNOWN` | `result.kind: review`, `properties.veval_result: UNKNOWN` |
| `ERROR` | `invocation.executionSuccessful: false` plus a `toolExecutionNotifications` entry, `properties.veval_result: ERROR` |
| Evidence locator | `result.locations[].physicalLocation` (file, region) or `result.properties.evidence` |
| Commands | `invocations[]` with `commandLine`, `workingDirectory`, `exitCode`, `startTimeUtc`, `endTimeUtc`, `environmentVariables` |
| Artifact revision | `run.properties.veval.vcs_revision_id`. `versionControlProvenance[]` is not emitted: SARIF requires `repositoryUri` on every entry and the report carries no repository URI, so the git revision goes into the run properties until it does. |
| Artifact hashes | `artifacts[].hashes` |
| Forensic finding | A result on the integrity rule with `properties.veval_severity` |
| Overall status, counts | `run.properties.veval` object |

Two v-eval states have no exact SARIF kind (`UNKNOWN`, `ERROR`) and are carried in `properties`, so the export is near-lossless but not symmetric. Import from SARIF is a separate, deferred adapter.

## Mapping to evolve-loop verdict tokens

| v-eval overall | evolve-loop audit token | Note |
| --- | --- | --- |
| `PASS` | `PASS` | |
| `FAIL` | `FAIL` | |
| `INCOMPLETE` | `WARN` | evolve-loop blocks ship on WARN and reviews case by case |
| `ADVISORY` | `SKIPPED` | **proposal**; advisory reviews do not gate |
| Criterion `NOT_APPLICABLE` | `SKIPPED` | at criterion level only |
| Criterion `ERROR` | no equivalent | carried in the report; surfaced as WARN with reason |

This mapping supports [decision 0001](../decisions/0001-independent-of-evolve-loop.md): v-eval stays independent and evolve-loop can consume the report through this table.

## HTML renderer

Every evaluation produces one HTML file ([decision 0021](../decisions/0021-html-report-every-evaluation.md)). Constraints: a single self-contained file, readable offline, no external scripts, styles, or fonts, light and dark rendering, no dependency on a browser at render time (the core writes the file). Fixed section order:

1. Header: artifact identity, revision, bundle digest, v-eval version, date, host.
2. Overall status with the rule applied and the blocking criteria, kept short.
3. Observations: everything found, grouped by source, with evidence excerpts.
4. Claim-to-verification table.
5. Criteria results table: ID, requirement, result, method, evidence with kind and isolation badges, next action.
6. Forensic findings with benign alternatives and dispositions.
7. Dimension metrics in native units with definition and threshold source.
8. Counts and coverage with denominators.
9. Improvement brief.
10. Limitations and not inspected.
11. Routing rationale.
12. Provenance: tools, commands, environment, isolation levels.
13. Learning references, when present.

The status block sits near the top for orientation but is deliberately brief; the evidence sections carry the weight, so the summary cannot bury what was found.

## Versioning

The schema is semver versioned. A report records the schema version it was written against. Renderers and the SARIF exporter declare the schema versions they accept. Breaking changes require a migration note under `schema/` and a new decision record.
