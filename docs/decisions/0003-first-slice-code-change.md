# First implemented slice is a code change against intent, design, and tests

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented. This decision selects the first slice; REQ-02 keeps content, context, code, and documentation in scope.

## Context and Problem Statement

The [requirements](../requirements.md#6-decisions) left the first automated use case open between code, documents, and generated or retrieved context. Each needs different evidence. Which artifact type should the first implemented slice target?

## Decision Drivers

* Availability of real, locally labeled cases: the maintainer's evolve-loop history contains documented code cases with known outcomes.
* Availability of external calibration material: SWE-Gate constraint cases and SWE-PRBench review comments are code.
* Existence of deterministic adapters: test execution, diff analysis, and static tools exist for code; almost none exist for generated context.
* The maintainer's later direction that evaluation input is all history data, so the first slice must not harden a code-only pipeline shape.

## Considered Options

* Code change against intent, design, and tests
* Generated context and handoffs
* Documents with claim-to-source support

## Decision Outcome

Chosen option: "Code change against intent, design, and tests", because it has the strongest evidence, tooling, and local cases. To prevent a code-only shape, the intake and adapter abstraction ([0009](0009-deterministic-intake-classifier.md), [0010](0010-first-adapters.md)) is designed for any history data from the start, and a claim-to-source adapter ships in the first release.

### Consequences

* Good, because pilot cases ([0015](0015-pilot-cases-and-labeling.md)) can be assembled from evolve-loop cycles plus synthetic variants immediately.
* Good, because the public worked example in `examples/code-review/` stays the illustration and gains a second valid implementation and one ambiguous requirement.
* Bad, because the skill's document and context sections stay shorter until their slices are validated.
* Bad, because retrieval-quality evaluation with relevance labels is deferred.

## Confirmation

The first pilot set contains code cases with per-criterion labels, and the adapter interface has at least one non-code adapter registered. Checked by reviewing the pilot manifest and the adapter registry once they exist.

## Pros and Cons of the Options

### Code change against intent, design, and tests

* Good, because real cases and deterministic adapters exist now.
* Good, because the best external calibration sets are code.
* Bad, because it risks a code-only mindset if intake is not generalized at the same time.

### Generated context and handoffs

* Good, because it is closest to the maintainer's verify-over-summary axis.
* Bad, because almost no deterministic checks exist.
* Bad, because labeling needs the original conversations.

### Documents with claim-to-source support

* Good, because claim decomposition and citation support are well studied.
* Bad, because evidence collection is mostly model judgment, which the research says to defer.

## More Information

* Related requirements: REQ-02, REQ-03, REQ-05. Informed by [code acceptance research](../research/2026-09-14-code-acceptance-landscape.md) and the [research synthesis](../research/2026-09-14-synthesis.md).
* OpenAI, "Why SWE-bench Verified no longer measures frontier coding capabilities": <https://openai.com/index/why-we-no-longer-evaluate-swe-bench-verified/> (2026-02-23; accessed 2026-09-14).
* He et al., "SWE-Gate": <https://arxiv.org/abs/2609.04167> (2026-09-03; accessed 2026-09-14).
* Kumar, "SWE-PRBench": <https://arxiv.org/abs/2603.26130> (2026-03-27; accessed 2026-09-14; single source).
* Raghavendra et al., "Agentic Rubrics", ACL 2026: <https://aclanthology.org/2026.acl-long.697/> (accessed 2026-09-14).
* evolve-loop, `runtime/docs/eval-grader-best-practices.md`, read locally 2026-09-14.
* Revisit when demand for document or context evaluation outpaces code in the pilot, or when the claim-to-source adapter proves more accurate than the code adapters.
