# v-eval is an independent product; evolve-loop adopts it later

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented. This record fixes the relationship so later design work does not silently assume either coupling or isolation.

## Context and Problem Statement

The maintainer's sibling project, evolve-loop, already ships an `evaluator` skill (evo plugin v22.21.0, Apache-2.0, Go 1.23). It scores six weighted dimensions on a 0.0 to 1.0 scale, computes a weighted composite mapped to STRONG / ADEQUATE / NEEDS WORK / CRITICAL, runs a perturbation-based gaming test, and executes under a real isolation profile (read-only repository, denied write paths, no network, budget and turn caps, a different CLI family from the builder). Its audit phase uses the verdict tokens PASS / FAIL / WARN / SKIPPED with an ALL-PASS rule.

v-eval's [design](../design.md) rejects a default composite score and proposes five per-criterion states. The two projects overlap heavily in purpose and disagree in scoring policy, and neither referenced the other before this decision. Should v-eval replace the evolve-loop evaluator, be consumed by it, or stay independent?

## Decision Drivers

* REQ-01 and REQ-04 require an open-source program with a skill as its core interface, which a library-only form would not satisfy.
* The composite-score disagreement must be resolved somewhere without forcing a rewrite of evolve-loop.
* evolve-loop's user-definable phases (merged 2026-05-27) let an operator add an optional phase as pure data under `.evolve/phases/<name>/`, so later adoption needs no core change in either project.
* evolve-loop's cycle history is the best local source of real pilot cases and should remain usable.

## Considered Options

* Independent product; evolve-loop adopts later via a published verdict mapping
* Replace evolve-loop's evaluator
* Library consumed by evolve-loop only
* Fully separate, no interoperability planned

## Decision Outcome

Chosen option: "Independent product; evolve-loop adopts later", because it keeps v-eval's scoring policy intact, satisfies REQ-01 and REQ-04, and still allows evolve-loop to adopt v-eval as an optional data-defined phase without either project changing its core.

v-eval publishes this verdict mapping for evolve-loop consumers:

| v-eval result | evolve-loop audit token |
| --- | --- |
| overall PASS | PASS |
| overall FAIL | FAIL |
| overall INCOMPLETE | WARN |
| criterion NOT_APPLICABLE | SKIPPED |
| criterion ERROR | no equivalent; surfaced as WARN with an explicit reason |

### Consequences

* Good, because the [report schema](../architecture/report-schema.md) needs only a reason field on ERROR to make the mapping lossless.
* Good, because v-eval's isolation profile ([0004](0004-start-at-rungs-2-to-4.md)) can use evolve-loop's profile as a reference without a dependency.
* Good, because evolve-loop cycles remain the primary pilot-case source ([0015](0015-pilot-cases-and-labeling.md)).
* Bad, because no code is shared, so any duplication between the two Go codebases is accepted for now.
* Bad, because evolve-loop's evaluator keeps its composite score; two tools from one maintainer disagree until evolve-loop adopts v-eval.

## Confirmation

A documented mapping table lives in the schema documentation, and a core unit test asserts that every v-eval overall and per-criterion state maps to exactly one evolve-loop token or an explicit "no equivalent" reason. No such test exists yet.

## Pros and Cons of the Options

### Independent product; evolve-loop adopts later

* Good, because v-eval's five-state, no-composite design stays intact.
* Good, because adoption is a data-defined optional phase in evolve-loop, not a core change.
* Neutral, because interop is a mapping, not shared code.
* Bad, because evolve-loop's isolation profile is not reused directly.

### Replace evolve-loop's evaluator

* Good, because one maintainer would run one evaluator.
* Bad, because it forces reconciling the composite score and six-dimension rubric now.
* Bad, because it couples v-eval's release cadence to evolve-loop's.

### Library consumed by evolve-loop only

* Good, because it is the narrowest scope.
* Bad, because it contradicts REQ-01 and REQ-04.

### Fully separate, no interoperability planned

* Good, because it is the simplest to reason about.
* Bad, because it wastes real pilot cases and a proven isolation profile.

## More Information

* Related requirements: REQ-01, REQ-04, REQ-24, REQ-25 in [requirements](../requirements.md). Informed by [local prior art](../research/2026-09-14-local-prior-art.md) and the [research synthesis](../research/2026-09-14-synthesis.md).
* evolve-loop evaluator and audit skills, `runtime/skills/evaluator/SKILL.md` and `runtime/skills/audit/SKILL.md`, read locally 2026-09-14; project at <https://github.com/mickeyyaya/evolve-loop> (accessed 2026-09-14).
* evolve-loop user-definable phases, merged to main at 395a770 on 2026-05-27 (session memory, read locally 2026-09-14).
* FIRST, CVSS v4.0 FAQ, caution against merging unrelated dimensions into one score: <https://www.first.org/cvss/v4.0/faq> (accessed 2026-09-14).
* Revisit when evolve-loop retires its evaluator in favour of v-eval, when a shared Go module would remove real duplication, or when the mapping proves lossy on real cycles.
