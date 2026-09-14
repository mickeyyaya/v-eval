# First adapters: commit-bound test evidence, diff-scope and integrity, claim-to-source support

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

The initial question was which single deterministic adapter to ship first. The maintainer redirected it: "think outside of coding with all the history data we received as input for evaluation. Test cases and test result should be just one of deterministic adapter, build an abstract layer to generalize the context and classifier to decide which criterions and adapter should be direct to." With the classifier fixed in [0009](0009-deterministic-intake-classifier.md), which adapters ship behind the common interface in the first release?

## Decision Drivers

* Test evidence must be bound to the exact commit; supplied logs are not evidence.
* Diff structure is the cheapest deterministic signal and the source of NOT_APPLICABLE.
* The first release must not be code-only in shape; documents and generated context need a claim-support path.
* Each adapter declares what it needs and what it produces, so the classifier can route to it.

## Considered Options

* Commit-bound test evidence
* Diff-scope and integrity
* Claim-to-source support
* Static-analysis SARIF import

The maintainer selected the first three.

## Decision Outcome

Chosen options: "Commit-bound test evidence", "Diff-scope and integrity", and "Claim-to-source support", because together they cover behavioural evidence, structural and integrity evidence, and non-code evidence, while sharing one adapter interface. Static-analysis SARIF import is deferred because SARIF ingestion is straightforward once the export path in [0006](0006-json-first-report-contract.md) exists.

Adapter contracts, in outline:

* Test evidence: run the mapped command in a clean checkout of the exact commit, or import JUnit or CTRF bound to that commit; record tree hash, commit, command, working directory, environment, collected and skipped counts, exit status, result digest. PASS only at that SHA; ERROR if not run; UNKNOWN if no test maps to the criterion.
* Diff-scope and integrity: paths touched and untouched, tests or CI configuration changed alongside source, lockfile drift, requirement-ID citations in the diff and test names. Yields NOT_APPLICABLE when a criterion's paths are untouched and feeds the tampering detectors in [0012](0012-forensic-detectors-v1.md).
* Claim-to-source: enumerate material claims in a document or generated context, locate supporting passages in the supplied sources, record source date and version, mark unsupported claims; source support is distinguished from independent truth.

### Consequences

* Good, because the three adapters exercise the interface with execution, structure, and source-support evidence.
* Good, because the claim-to-source adapter is the direct implementation of the verify-over-summary axis for non-code history ([0008](0008-evidence-policy-verify-over-summary.md)).
* Bad, because three adapters before any pilot case is graded is more surface than the roadmap's stage 1 assumed.
* Bad, because static-analysis findings are unavailable until the SARIF import lands.

## Confirmation

Each adapter has a manifest declaring required inputs and produced evidence types; the classifier test suite routes fixture bundles to the expected adapters; each adapter has tests for its PASS, FAIL, UNKNOWN, ERROR, and where applicable NOT_APPLICABLE paths. None exists yet.

## Pros and Cons of the Options

### Commit-bound test evidence

* Good, because nobody ships it and it implements "logs are not the source of truth".
* Bad, because it needs a clean checkout or a trusted import.

### Diff-scope and integrity

* Good, because cheap, fully deterministic, and the source of NOT_APPLICABLE.
* Bad, because it does not verify behaviour.

### Claim-to-source support

* Good, because it covers documents and generated context from the first release.
* Bad, because claim extraction is model-assisted and must be measured separately from support checking.

### Static-analysis SARIF import

* Good, because SARIF is already the interchange format.
* Bad, because deferred; least distinctive.

## More Information

* Related requirements: REQ-02, REQ-03, REQ-05, REQ-13. Informed by [code acceptance research](../research/2026-09-14-code-acceptance-landscape.md) and [deterministic tools](../deterministic-tools-2026.md).
* GitHub artifact attestations and `actions/attest` custom predicates: <https://docs.github.com/en/actions/concepts/security/artifact-attestations> and <https://github.com/actions/attest> (accessed 2026-09-14).
* PatchDrill, hash-stamped evidence manifest and "planned but never run" checks: <https://github.com/seungdori/patchdrill> (accessed 2026-09-14).
* traceSDD, orphan-requirement set difference: <https://arxiv.org/html/2606.30689v1> (2026-06-28; accessed 2026-09-14; single source).
* Gao et al., ALCE, citation recall and precision: <https://aclanthology.org/2023.emnlp-main.398/> (accessed 2026-09-14). Min et al., FActScore: <https://aclanthology.org/2023.emnlp-main.741/> (accessed 2026-09-14).
* Terminal-Bench 2.0, final-state grading with CTRF output: <https://arxiv.org/abs/2601.11868> (accessed 2026-09-14).
* Revisit when the pilot shows one adapter dominates and the others are unused, or when SARIF import becomes a blocker for security criteria.
