# The evaluation contract is a section of the report schema, not a schema of its own

* Status: accepted
* Deciders: maintainer (mickeyyaya), during the core walking skeleton
* Date: 2026-09-15

Implemented 2026-09-15: `schema/report.schema.json` (the `contract` section) and `core/report/types.go`.

## Context and Problem Statement

[Decision 0004](0004-start-at-rungs-2-to-4.md) lays out `schema/` as "contract and report JSON Schema, versioned", and [decision 0006](0006-json-first-report-contract.md) makes the report JSON canonical. When the schema was written, the contract turned out to be one section of a report: `contract`, carrying `contract_id`, `contract_version`, `status`, `conflicts[]`, and `criteria[]`. Should it also exist as a file of its own, referenced by the report schema?

## Decision Drivers

* In the walking skeleton a contract is never evaluated on its own; it is always read, validated, and digested as part of a report.
* Two schema files describing one object drift apart, and the drift test can hold only one of them equal to the Go types.
* The report id covers the contract as it was evaluated; a contract validated elsewhere, against another file, would not be the contract the report names.
* The classifier ([decision 0009](0009-deterministic-intake-classifier.md)) will one day propose a contract before any report exists.

## Considered Options

* The contract is the `contract` section of the report schema
* A separate `schema/contract.schema.json`, pulled into the report schema by reference
* Both: a standalone file generated from the report schema

## Decision Outcome

Chosen option: "The contract is the `contract` section of the report schema", because it leaves one canonical record and one drift test. `core/report/schema_drift_test.go` holds `schema/report.schema.json` equal to the Go types field by field and enum by enum; a second file would need a second test of its own or would drift unchecked, and the two would disagree exactly where it matters, in what a criterion is allowed to say.

### Consequences

* Good, because there is one file, one version, and one drift test.
* Good, because a contract can never state a criterion shape a report cannot carry.
* Bad, because a tool that wants to validate a contract alone has to carry the report schema and point at the `contract` section within it.
* Bad, because decision 0004's layout line, "contract and report JSON Schema", now names one file rather than two; the [roadmap](../../ROADMAP.md) records the folding.
* A standalone contract schema may still be extracted when the classifier plan needs to emit and validate a contract before any report exists. The extraction is mechanical, because the section already stands on its own inside the file.

## Confirmation

`schema/` contains exactly one schema file, `report.schema.json`; `core/report/schema_drift_test.go` covers `contract` and its criteria alongside every other section; `veval validate` reports contract violations by their path within the report. Implemented 2026-09-15.

## Pros and Cons of the Options

### The `contract` section of the report schema

* Good, because one canonical record and one drift test.
* Good, because the contract the report id covers is the contract that was validated.
* Bad, because standalone contract validation needs the whole file.

### A separate contract schema file

* Good, because a classifier could emit and validate a contract with no report around it.
* Bad, because two files describing one object need two drift tests, and nothing yet emits a bare contract.

### Both, with the standalone file generated

* Good, because it serves both readers without hand-maintained duplication.
* Bad, because a generator and its staleness check are more machinery than the one consumer that does not yet exist.

## More Information

* Related requirements: REQ-09, REQ-13, REQ-21. Follows [decision 0006](0006-json-first-report-contract.md); the layout it narrows is in [decision 0004](0004-start-at-rungs-2-to-4.md).
* Section contents and field names: [report schema, contract](../architecture/report-schema.md#contract).
* Revisit when the classifier in [decision 0009](0009-deterministic-intake-classifier.md) needs to produce a contract before a report exists, or when a second tool has to validate a contract without one.
