# Pilot cases from evolve-loop history plus synthetic variants; sole labeler with a blind second opinion

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

Every accuracy claim, the anchor set, and the learning loop rest on labeled cases. Where do the first pilot cases come from, and who labels them independently?

## Decision Drivers

* Real cases with known outcomes exist in evolve-loop's cycle history, including the tautological-grader incidents and the prose-heading classifier drift.
* Synthetic variants are needed for controlled failure modes: subtle violation, missing evidence, valid alternative implementation, ambiguous requirement.
* The maintainer is a solo labeler; the research allows one accountable labeler but requires a detector that fails differently to break self-evaluation coupling.
* Disagreement records must be kept; held-out cases must never enter prompt tuning; the set must not be public.

## Considered Options

Case source:

* evolve-loop history plus synthetic variants
* Public benchmark slices (SWE-Gate, SWE-PRBench)
* Synthetic only
* Mixed: all three

Labeling:

* Maintainer as sole labeler with a disclosed blind second opinion from a different model family
* Maintainer plus one independent human reviewer
* Maintainer alone, no second opinion

## Decision Outcome

Chosen options: "evolve-loop history plus synthetic variants" and "Maintainer as sole labeler with a disclosed blind second opinion from a different model family". Cases are mined from evolve-loop cycles, with ordinary passing cycles as benign controls, plus seeded variants. The maintainer labels each criterion and overall acceptability before seeing v-eval's output. A second label comes from a model of a different family than any judge used, blind to the maintainer's label and to v-eval's verdict; disagreements are adjudicated by the maintainer and both original labels are retained. Cases are split deterministically by hash at creation into train, dev, and test; the test split plus extra items form the locked anchor set, stored outside the public tree, encrypted at rest, with a canary.

### Consequences

* Good, because cases are real, local, and already documented with outcomes.
* Good, because a disagreement record exists from the first case.
* Bad, because the set is code-heavy and English-only; document and context cases are few until those slices are validated.
* Bad, because a model second opinion shares blind spots with any model judge; it is a disclosed limitation, not independence.

## Confirmation

A pilot manifest lists each case with origin, split, labels, second-opinion label, and adjudication note; a test asserts no anchor case appears in any prompt, example pool, or public directory; reported metrics show false acceptance, false rejection, abstention, and decided-case accuracy with denominators. None exists yet.

## Pros and Cons of the Options

### evolve-loop history plus synthetic variants

* Good, because real and local with known outcomes.
* Bad, because code-heavy.

### Public benchmark slices

* Good, because externally labeled and comparable.
* Bad, because code-only, licences and contamination need checking, no context cases.

### Synthetic only

* Good, because fully controlled and publishable.
* Bad, because measures only imagined failure modes.

### Mixed

* Good, because broadest coverage.
* Bad, because most labeling work before anything is measured.

### Sole labeler with blind model second opinion

* Good, because practical for a solo maintainer and breaks self-evaluation coupling partially.
* Bad, because not independent in the human sense.

### Maintainer plus one human

* Good, because strongest evidence for published claims.
* Bad, because depends on recruiting a second person.

### Maintainer alone

* Good, because fastest.
* Bad, because no disagreement record.

## More Information

* Related requirements: REQ-20, REQ-22, REQ-24. Informed by [evaluator learning research](../research/2026-09-14-evaluator-learning.md), [integrity and RSI research](../research/2026-09-14-integrity-and-rsi.md), and the [comparison protocol](../comparison.md).
* Husain and Shankar, evals FAQ, 100 to 200 labels per failure mode, one accountable labeler, held-out TPR and TNR: <https://hamel.dev/blog/posts/evals-faq/> (modified 2026-09-01; accessed 2026-09-14). validate-evaluator skill: <https://www.skills.sh/hamelsmu/evals-skills/validate-evaluator> (accessed 2026-09-14).
* Zhang et al., "Who Grades the Grader?", detectors that fail differently from the evaluated model: <https://arxiv.org/html/2607.12790v1> (2026-07-14; accessed 2026-09-14).
* METR, MALT, label provenance for natural versus prompted behaviour: <https://metr.org/blog/2025-10-14-malt-dataset-of-natural-and-prompted-behaviors/> (accessed 2026-09-14).
* Lee et al., correct reporting of judge evaluations with calibration sets: <https://arxiv.org/abs/2511.21140> (2025-11; accessed 2026-09-14).
* OpenAI, SWE-bench Verified audit, 59.4% flawed items in a hard subset, so audit your own set: <https://openai.com/index/why-we-no-longer-evaluate-swe-bench-verified/> (accessed 2026-09-14).
* Revisit when a second human labeler becomes available, or when the pilot needs document and context cases the evolve-loop history cannot supply.
