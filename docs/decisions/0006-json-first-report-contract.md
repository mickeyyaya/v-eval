# JSON-first report contract; Markdown and HTML are renders; SARIF export

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Implemented 2026-09-15: `schema/report.schema.json` v0.1.0, `core/report` for decoding, validation, and aggregation, `core/render` for the two renders, `core/export/sarif.go` for the export, and `cmd/veval` for the command line. The [contract template](../../templates/evaluation-contract.md) is still Markdown and has not been regenerated from the schema. Three points of the outcome below were settled differently in the building; see the amendments.

## Context and Problem Statement

The repository's design deferred a machine-readable schema until real cases existed. Two findings changed that. evolve-loop's June 2026 architecture review found that all six of its phase classifiers derived verdicts by grepping prose headings from LLM-authored reports; the headings drifted twice, no golden test existed, and a valid report was classified FAIL. The schema research found no existing format carrying per-criterion five-state verdicts plus evidence plus execution provenance. What is the canonical report contract?

## Decision Drivers

* Consumers must never parse prose to learn a verdict.
* Artifact identity (commit, content hash) and execution provenance (command, exit code, environment) must be first-class.
* Interoperability with CI and code-scanning tools.
* Avoid adopting an unmaintained or ill-fitting schema.

## Considered Options

* JSON-first versioned schema; Markdown and HTML rendered; SARIF export
* Markdown-first with a lint
* Adopt structured-evaluation's schema

## Decision Outcome

Chosen option: "JSON-first versioned schema; Markdown and HTML rendered; SARIF export", because it makes every verdict machine-readable, keeps provenance first-class, and reuses a standards-body format for interchange. The canonical artifact is a JSON document validated against a versioned schema in `schema/`. The Go core validates it, computes counts and overall status from it, and renders Markdown and HTML ([0021](0021-html-report-every-evaluation.md)). Markdown is never parsed back. The SARIF 2.1.0 export maps PASS to `pass`, FAIL to `fail`, NOT_APPLICABLE to `notApplicable`, UNKNOWN to `review`, ERROR to `executionSuccessful: false` with a notification; provenance maps to `invocation` and `versionControlDetails`, artifact identity to `artifact.hashes`. structured-evaluation is not adopted; a lossy export may follow.

### Consequences

* Good, because every gate or classifier consuming v-eval reads JSON fields.
* Good, because SARIF gives GitHub code-scanning and similar ingestion for free.
* Bad, because two exporters and a schema version policy must be maintained.
* Bad, because the existing Markdown template and worked example must be regenerated.

## Confirmation

A JSON Schema file in `schema/` with a version; a core test that renders Markdown and HTML from a fixture and never parses them; a SARIF export validated by a SARIF schema validator. Implemented in part on 2026-09-15: `schema/report.schema.json` with `schema.Version`, `core/render/renderers_test.go` and the goldens under `core/render/testdata/` rendering both formats from the two fixtures without parsing either back, and `core/export/sarif_test.go` over the golden logs. The export has not been run through an external SARIF schema validator; the tests assert the shape the specification calls for, not the specification itself.

## Amendments (2026-09-15)

Three points of the Decision Outcome were settled differently when the export was built.

* `versionControlProvenance[]` is not emitted. SARIF requires `repositoryUri` on every entry and a report carries no repository URI, so such an entry could only ever be invalid. A git revision travels in `run.properties.veval.vcs_revision_id` instead; the `artifacts[].hashes` half for a `content:sha256:` revision is unchanged (`core/export/sarif.go`).
* Each invocation's `executionSuccessful` is that command's own exit status and nothing else: a command that exited zero did exit zero however the rest of the evaluation went. An `ERROR` criterion is a fact about the run rather than about any one command, so it is attached to none: every errored criterion is named in a `toolExecutionNotifications` entry on a single synthesized invocation, appended last, which names no command line (`core/export/sarif.go`).
* The canonical JSON form this record rests on -- the encoding, the two digests taken over it, and the requirement that a second implementation reproduce the bytes -- is defined in [report-schema.md](../architecture/report-schema.md#canonical-form-and-digests), not here.

## Pros and Cons of the Options

### JSON-first; renders; SARIF export

* Good, because verdict rollup and provenance are enforced in code.
* Good, because SARIF carries four of five states plus provenance natively.
* Bad, because two mappers to maintain.

### Markdown-first with a lint

* Good, because it is human-friendly and matches the current template.
* Bad, because every consumer parses prose; rollup cannot be enforced.

### Adopt structured-evaluation's schema

* Good, because Go types, JSON Schema, Zod bindings, and a lint/check/render CLI exist under MIT.
* Bad, because per-criterion verdicts are only `pass | partial | fail`, and lint rejects other values.
* Bad, because it has no commit hash, content hash, command, exit code, or environment fields.
* Bad, because zero stars, a single maintainer, fourteen releases in seven months, and a stale example.

## More Information

* Related requirements: REQ-09, REQ-13, REQ-21. Informed by [report schema research](../research/2026-09-14-report-schemas.md) and [local prior art](../research/2026-09-14-local-prior-art.md).
* OASIS, SARIF v2.1.0 plus Errata 01: <https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/sarif-v2.1.0-errata01-os-complete.html> (2023-08-28; accessed 2026-09-14).
* plexusone/structured-evaluation, v0.14.0: <https://github.com/plexusone/structured-evaluation> (2026-08-15; accessed 2026-09-14).
* EvalEval, Every Eval Ever schema: <https://github.com/evaleval/every_eval_ever> (accessed 2026-09-14).
* OpenTelemetry GenAI evaluation event, Development status: <https://github.com/open-telemetry/semantic-conventions-genai/blob/main/docs/gen-ai/gen-ai-events.md> (accessed 2026-09-14).
* evolve-loop, `runtime/docs/research/verdict-and-gate-proxy-failure-class-2026-06-03.md`, read locally 2026-09-14.
* Revisit when a standards-body evaluation report format reaches a stable release covering the five states and provenance, or when SARIF's `review` kind proves insufficient.
