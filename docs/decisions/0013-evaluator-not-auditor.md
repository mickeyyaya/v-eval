# v-eval is an evaluator, not a verdict auditor: evidence first, verdicts derived

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented. This record fixes report ordering, completeness of collection, and tone.

## Context and Problem Statement

While selecting forensic detectors, the maintainer stated the governing principle: "you are evaluator, not the verdict auditor, our goal is to collect all the information and reflect the true hidden underneath the data." Earlier axes point the same way: never trust a summary, act as a detective. An auditor stops at pass or fail; an evaluator keeps collecting until the picture is complete. What does this change in the product?

## Decision Drivers

* The product's value is revealing what is actually there, not issuing a gate decision.
* Observations not tied to any criterion still carry information and must not be dropped.
* The five-state vocabulary and overall status remain useful summaries, but summaries must not bury evidence.
* Learning must optimize revealed-information completeness and accuracy, not agreement with the user.

## Considered Options

* Evidence-first report with derived verdicts, completeness as a first-class quality
* Verdict-first report with evidence as supporting detail (the conventional audit layout)
* Verdicts only, evidence on request

## Decision Outcome

Chosen option: "Evidence-first report with derived verdicts, completeness as a first-class quality". Report ordering: observations and evidence (everything found, including items tied to no criterion), then the claim-to-verification table, then per-criterion results, then the derived overall status, then the improvement brief, then routing rationale and provenance. The report states what was inspected, what was not, and what remains unknown with the same prominence as failures. Tone is investigative, not adversarial: findings are observed facts with locations and quotes. Gaming traces are reported as observations with evidence, then mapped to integrity criteria.

### Consequences

* Good, because a reader sees the evidence before the label and can disagree with the label.
* Good, because "not inspected" and "unknown" become visible outputs, which supports the abstention measurements in [0015](0015-pilot-cases-and-labeling.md).
* Bad, because reports are longer and the overall status is not the first line; the HTML renderer ([0021](0021-html-report-every-evaluation.md)) needs a compact status strip near the top without demoting evidence.
* Bad, because an observations section can attract noise; the pilot must measure whether readers find it useful.

## Confirmation

The report schema has an `observations` array independent of criteria and an `inspection_scope` object with inspected and not-inspected paths; the renderers emit sections in the fixed order above; a review checklist for reports asks whether any observation was dropped because it did not change a verdict. None exists yet.

## Pros and Cons of the Options

### Evidence-first with derived verdicts

* Good, because it matches the maintainer's principle and the verify-over-summary axis.
* Good, because completeness of collection becomes measurable.
* Bad, because longer reports.

### Verdict-first with supporting evidence

* Good, because conventional and quick to scan.
* Bad, because evidence that did not change the verdict tends to be omitted.

### Verdicts only

* Good, because shortest.
* Bad, because contradicts REQ-09 and the principle entirely.

## More Information

* Related requirements: REQ-09, REQ-11, REQ-13, and REQ-30 in [requirements](../requirements.md). Informed by the [research synthesis](../research/2026-09-14-synthesis.md) and [deep research](../deep-research.md).
* Hamel Husain, "'It's Hard to Eval' Is a Product Smell", expose assumptions, provenance, intermediate work, and unresolved items: <https://hamel.dev/blog/posts/eval-smell/> (2026-06-29; accessed 2026-09-14).
* Anthropic, "Demystifying evals for AI agents", read the transcripts, grade final state: <https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents> (2026-01-09; accessed 2026-09-14).
* Husain and Shankar, "AI Evals: Everything You Need to Know", error discovery before metric building: <https://hamel.dev/blog/posts/evals-faq/> (2025-05-28, modified 2026-09-01; accessed 2026-09-14).
* Andrew Ng, error analysis through pipeline stages: <https://www.deeplearning.ai/the-batch/improve-agentic-performance-with-evals-and-error-analysis-part-2/> (2025-10-22; accessed 2026-09-14).
* Revisit when pilot readers report the observations section is noise, or when a consumer needs verdict-first output (which a renderer option, not a schema change, should provide).
