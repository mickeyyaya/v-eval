# No composite score; verdicts plus native dimension scores

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

REQ-08 asks for dimension scores where a meaningful metric is defined, together with its definition and source. The [design](../design.md) and the [industry scorecard](../industry-scorecard.md) argue that severity, latency, maintainability, and content-quality measures use different units and cannot be averaged without hiding a required failure. evolve-loop's evaluator, by contrast, computes a weighted composite from six 0.0 to 1.0 dimensions. Should v-eval ever emit a single number?

## Decision Drivers

* A required failure must never be averaged away by unrelated strengths.
* REQ-08 and REQ-09 require each score to carry its definition, authority, and reasoning.
* Missing measurements must remain visibly absent, not become zero.
* The learning loop must optimize accuracy against labels, not a scalar that can be gamed.

## Considered Options

* No composite; per-criterion verdicts plus native dimension scores
* Opt-in composite with disclosed weights and a required-failure veto
* Always emit a composite like evolve-loop

## Decision Outcome

Chosen option: "No composite; per-criterion verdicts plus native dimension scores", because it is the only option consistent with every document in the repository and with the cited guidance against merging unrelated dimensions. A report carries per-criterion results in the five-state vocabulary ([0007](0007-verdict-vocabulary.md)), an overall PASS / FAIL / INCOMPLETE derived by the documented policy, and native dimension measurements, each with metric name, version, definition, unit, range, direction, observed value, tool and configuration, threshold and its authority, and applicability. An opt-in composite is deferred, not rejected forever; if added, it must disclose weights and normalization and preserve the required-failure veto.

### Consequences

* Good, because the [report schema](../architecture/report-schema.md) has no `composite` field and no averaging logic to validate.
* Good, because the verdict mapping to evolve-loop ([0001](0001-independent-of-evolve-loop.md)) maps status only.
* Good, because the learning loop ([0014](0014-adaptive-learning-precedent-bank.md)) has no scalar to overfit.
* Bad, because there is no one-line comparability across projects or over time.
* Bad, because renderers must present several units side by side without implying a total.

## Confirmation

The schema forbids a top-level numeric summary; a schema test rejects any report that adds one. The HTML and Markdown renderers keep units visible in the dimension table. Neither exists yet.

## Pros and Cons of the Options

### No composite; per-criterion verdicts plus native dimension scores

* Good, because required failures stay visible.
* Good, because each measurement keeps its definition and threshold origin.
* Bad, because no single comparable number exists.

### Opt-in composite with disclosed weights and a required-failure veto

* Good, because projects that need a number can declare one transparently.
* Neutral, because it is one schema section and one calculation.
* Bad, because it invites use before any composite has been validated against developer decisions.

### Always emit a composite like evolve-loop

* Good, because it is familiar and comparable.
* Bad, because it contradicts the repository's design and the cited guidance.
* Bad, because attractive formatting can compensate numerically for a critical failure.

## More Information

* Related requirements: REQ-06, REQ-07, REQ-08, REQ-09. Informed by the [research synthesis](../research/2026-09-14-synthesis.md) and the [industry scorecard](../industry-scorecard.md).
* FIRST, CVSS v4.0 FAQ: <https://www.first.org/cvss/v4.0/faq> (accessed 2026-09-14).
* Hamel Husain, "Using LLM-as-a-Judge For Evaluation", binary judgments and critiques over scales: <https://hamel.dev/blog/posts/llm-judge/> (2024-10-29, modified 2026-09-01; accessed 2026-09-14).
* Eugene Yan, "Product Evals in Three Simple Steps": <https://eugeneyan.com/writing/product-evals/> (2025-11-23; accessed 2026-09-14).
* Nakajima, "Regimes", in-sample +0.18 collapsing to held-out +0.04: <https://arxiv.org/pdf/2606.10241> (2026-06-08; accessed 2026-09-14).
* Microsoft Maintainability Index definition, showing tool-specific formulas: <https://learn.microsoft.com/en-us/visualstudio/code-quality/code-metrics-maintainability-index-range-and-meaning?view=visualstudio> (accessed 2026-09-14).
* Revisit when the pilot shows a user-tested need for a comparable number with a proposed transparent policy, or when a CI consumer cannot ingest per-dimension measurements.
