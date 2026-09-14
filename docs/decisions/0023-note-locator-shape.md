# A fourth locator shape, `note`, for evidence that was neither opened nor run

* Status: accepted
* Deciders: maintainer (mickeyyaya), at approval of the core walking-skeleton plan
* Date: 2026-09-15

Implemented 2026-09-15: `core/report/locator.go`, `schema/report.schema.json`, `core/render`, and `core/export/sarif.go`.

## Context and Problem Statement

[Decision 0008](0008-evidence-policy-verify-over-summary.md) fixes three locator groups -- a file and its line range, a command with its working directory, exit status, and log, a passage with its source and dates -- and each of the three names a place the evaluator went. [Decision 0013](0013-evaluator-not-auditor.md) requires the report to carry everything found, including what changes no verdict. A report therefore has to carry material the evaluator did not observe: the candidate's own summary, a benchmark number quoted from a pull request description, a precedent retrieved from the bank. With three shapes only, such a record has no locator it can honestly fill, and validation rejects it. The evaluator is then pushed to drop the observation or to dress it as a citation it cannot support. What shape does non-qualifying evidence take?

## Decision Drivers

* Nothing found is dropped because it changes no verdict ([0013](0013-evaluator-not-auditor.md)).
* A summary can never be the basis for `PASS` ([0008](0008-evidence-policy-verify-over-summary.md)); the rule must not be weakened to make room for the summary.
* A locator that names no place at all is worse than a labelled one: a reader cannot then tell a citation from a paraphrase.
* The distinction must be machine-checkable in the core, not a convention in the prompt.

## Considered Options

* A fourth shape, `note`, complete on its own and never qualifying for `PASS`
* An empty locator permitted on non-qualifying evidence
* Three shapes only; such material recorded outside the evidence arrays

## Decision Outcome

Chosen option: "A fourth shape, `note`", because it keeps non-qualifying material inside the evidence arrays without touching the citation rule. A `note` locator carries one required field, a plain-language pointer to where the material came from, for example "Benchmark number quoted in the pull request description, paragraph 2". `Shape()` reports `note` only when the note field is the one group set, so a note mixed with half a file citation is still no shape at all. `Qualifies()` is false for `note`, so `evidence.pass_requires_observed_locator` is untouched: a criterion whose only evidence is a note cannot be `PASS`. Renderers and the SARIF export mark the shape beside the observation.

### Consequences

* Good, because candidate-supplied summaries and retrieved precedents get a locator they can fill honestly, and so stay inside the evidence arrays that `provenance.evidence_digest` covers.
* Good, because the `PASS` rule is stated once and not relaxed: the shape that cannot support `PASS` is the shape the core names.
* Good, because an empty or half-filled locator remains a violation, so "nobody recorded where this came from" and "this came from the pull request description" stay different facts.
* Bad, because a note is free text and nothing can check it; a reader must trust that it says where the material came from.
* Bad, because a fourth shape is a fourth branch in the schema, both renderers, and the export.

## Confirmation

`core/report/locator.go` defines `ShapeNote` and leaves it out of `Qualifies()`; `core/report/locator_test.go` covers the shape and a `PASS` resting only on a note; `schema/report.schema.json` carries the fourth branch of the locator alternatives and `core/report/schema_drift_test.go` holds those branches equal to the Go fields; the extended fixture `core/report/testdata/extended-example.json` uses the shape. Implemented 2026-09-15.

## Pros and Cons of the Options

### A fourth shape, `note`

* Good, because non-qualifying evidence has a writable, checkable place.
* Good, because the citation rule stays exactly as decision 0008 states it.
* Bad, because the note's content is unverifiable prose.

### An empty locator on non-qualifying evidence

* Good, because no new shape to carry through the schema, the renderers, and the export.
* Bad, because "the evaluator did not record where this came from" and "this came from the pull request description" become the same report.

### Record such material outside the evidence arrays

* Good, because the evidence arrays stay uniformly strong.
* Bad, because the evidence digest would not cover it, and decision 0013's completeness rule would be served through a second, weaker channel.

## More Information

* Related requirements: REQ-05, REQ-13, and REQ-28 in [requirements](../requirements.md). Follows [decision 0008](0008-evidence-policy-verify-over-summary.md) and [decision 0013](0013-evaluator-not-auditor.md).
* Raised as Assumption 2 of the [core walking-skeleton plan](../superpowers/plans/2026-09-14-core-walking-skeleton.md) and confirmed at approval.
* Shipped field names and the exactly-one-complete-group rule: [report schema, evidence shape](../architecture/report-schema.md#evidence-shape).
* Revisit when a note is used where a qualifying citation was available, or when a structured non-qualifying locator -- a retrieved precedent id, say -- earns a shape of its own.
