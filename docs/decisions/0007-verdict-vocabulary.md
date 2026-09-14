# Five per-criterion states and a three-state overall result, with published mappings

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

The [design](../design.md#results-and-acceptance) proposes PASS, FAIL, UNKNOWN, ERROR, and NOT_APPLICABLE per criterion and PASS, FAIL, INCOMPLETE overall. evolve-loop uses PASS / FAIL / WARN / SKIPPED. SARIF uses `pass`, `fail`, `notApplicable`, `review`, `open`, `informational`. The research found no prior art with UNKNOWN or NOT_APPLICABLE; smevals' rule that a harness failure is never graded is the only analogue to ERROR. Which vocabulary should the schema fix?

## Decision Drivers

* Missing evidence, a tool crash, and a contract-defined exclusion are different facts and must stay distinguishable.
* Anthropic's 2026 work supports a sanctioned, machine-readable abstain label.
* Consumers in evolve-loop and CI tools need a documented mapping.
* The vocabulary is a differentiator of v-eval and should not be diluted for interop.

## Considered Options

* Keep five per-criterion states plus PASS / FAIL / INCOMPLETE overall; publish mappings
* Align to evolve-loop's PASS / FAIL / WARN / SKIPPED
* Use SARIF kinds directly

## Decision Outcome

Chosen option: "Keep five per-criterion states plus PASS / FAIL / INCOMPLETE overall; publish mappings", because it preserves the distinctions the design and the integrity research say matter, and because UNKNOWN and NOT_APPLICABLE are the differentiator. Per criterion: PASS, FAIL, UNKNOWN, ERROR, NOT_APPLICABLE. Overall, for required applicable criteria: any FAIL gives FAIL; otherwise any UNKNOWN or ERROR gives INCOMPLETE; otherwise all PASS with at least one such criterion gives PASS; provisional criteria or unresolved contract conflicts prevent PASS. Mappings to evolve-loop ([0001](0001-independent-of-evolve-loop.md)) and to SARIF kinds ([0006](0006-json-first-report-contract.md)) are published beside the schema.

### Consequences

* Good, because an evidence gap and an operational failure never collapse into one label.
* Good, because the schema enumerates exactly these values and rejects others.
* Bad, because two mapping tables must be maintained and their lossy cells documented (ERROR has no evolve-loop equivalent; UNKNOWN and ERROR share SARIF's `review` and `executionSuccessful`).
* Bad, because small local models must be forced into structured output to use the vocabulary reliably.

## Confirmation

The JSON Schema enumerates the five per-criterion values and three overall values; a core test covers the overall-status policy on fixtures including provisional criteria and contract conflicts; mapping tables have round-trip tests. None exists yet.

## Pros and Cons of the Options

### Keep five states; publish mappings

* Good, because no prior art carries UNKNOWN or NOT_APPLICABLE; this is the differentiator.
* Good, because a sanctioned abstain label is supported by 2026 evidence.
* Bad, because two mappings to maintain.

### Align to evolve-loop's tokens

* Good, because direct interoperability with the audit phase.
* Bad, because WARN merges missing evidence with a tool crash.

### Use SARIF kinds directly

* Good, because standards-native.
* Bad, because `review` and `open` do not separate UNKNOWN from ERROR, and INCOMPLETE has no home.

## More Information

* Related requirements: REQ-09, REQ-12, REQ-13. Informed by [report schema research](../research/2026-09-14-report-schemas.md) and [code acceptance research](../research/2026-09-14-code-acceptance-landscape.md).
* Anthropic Alignment Science, "Agentic Misalignment in Summer 2026", structured DECLINE_TO_LABEL as a sanctioned abstain: <https://alignment.anthropic.com/2026/agentic-misalignment-summer-2026/> (2026-07-13; accessed 2026-09-14).
* Anthropic, "Demystifying evals for AI agents", grader taxonomy: <https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents> (2026-01-09; accessed 2026-09-14).
* prime-radiant-inc/smevals, harness failure is never graded: <https://github.com/prime-radiant-inc/smevals> (accessed 2026-09-14).
* UK AISI Inspect log format, C/I/P/N values and `error`: <https://inspect.aisi.org.uk/eval-logs.html> (accessed 2026-09-14).
* OASIS SARIF v2.1.0 result kinds: <https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/sarif-v2.1.0-errata01-os-complete.html> (accessed 2026-09-14).
* Revisit when a consumer needs a state the vocabulary cannot express, or when pilot data shows UNKNOWN is overused as a refuge.
