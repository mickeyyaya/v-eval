# No separate LLM judge in v1: the host assistant inspects, the core decides

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

Should the first release include a separate LLM judge beyond the host assistant's own inspection? The integrity research found that judge injection defenses fall to adaptive attack, that small local judges are near random on hard code pairs, and that any judge must be validated on a labeled anchor set before its prompt is trusted.

## Decision Drivers

* Software must own pass/fail; a judge may only label evidence.
* No anchor set exists yet to validate any judge.
* The skill must work identically on every CLI, including ollama-backed agents.
* A provider dependency contradicts local-first operation.

## Considered Options

* No separate judge in v1; host assistant inspects, core decides
* One pinned per-criterion judge, opt-in
* Multi-judge panel

## Decision Outcome

Chosen option: "No separate judge in v1; host assistant inspects, core decides", because a judge cannot be validated before the anchor set exists and because the host assistant already runs the skill on every CLI. The assistant performs semantic inspection and writes evidence in the schema's shape; the Go core validates evidence shape ([0008](0008-evidence-policy-verify-over-summary.md)), runs adapters, and computes results. A pinned per-criterion judge, called with one criterion and minimal hunks, rationale first, forced JSON, model and prompt hash recorded, is added only after the anchor set exists.

### Consequences

* Good, because no provider dependency and no unvalidated judge in the first release.
* Good, because the evidence rule already forces the assistant to cite what it opened or ran.
* Bad, because the semantic quality of v1 depends entirely on the host model; on weak local models many criteria will land on UNKNOWN.
* Bad, because there is no second opinion at evaluation time; the pilot's blind second label ([0015](0015-pilot-cases-and-labeling.md)) is the only comparison.

## Confirmation

The skill text contains no instruction to call an external model; the core has no provider client in v1; the roadmap lists the judge adapter under the stage that follows anchor-set creation. None exists yet.

## Pros and Cons of the Options

### No separate judge in v1

* Good, because works on every CLI and needs no validation set first.
* Bad, because semantic quality depends on the host model.

### One pinned per-criterion judge, opt-in

* Good, because gives an isolatable second opinion with provenance now.
* Bad, because ships unvalidated and adds a provider dependency.

### Multi-judge panel

* Good, because several opinions.
* Bad, because correlated errors, position bias, and cost with no reliable independence.

## More Information

* Related requirements: REQ-04, REQ-13. Informed by [integrity and RSI research](../research/2026-09-14-integrity-and-rsi.md) and [evaluator learning research](../research/2026-09-14-evaluator-learning.md).
* Nasr, Carlini et al., "The Attacker Moves Second": <https://arxiv.org/abs/2510.09023> (accessed 2026-09-14). Narisetty et al., out-of-band enforcement: <https://arxiv.org/abs/2606.26479> (accessed 2026-09-14).
* Jiang et al., CodeJudgeBench, non-thinking judges below 60% on code pairs: <https://arxiv.org/abs/2507.10535> (2025-07; accessed 2026-09-14).
* Laddha et al., SLMJury: <https://arxiv.org/html/2606.07810> (2026-06; accessed 2026-09-14).
* Zheng et al., "Judging LLM-as-a-Judge", position, verbosity, and self-enhancement bias: <https://arxiv.org/abs/2306.05685> (accessed 2026-09-14).
* Anthropic, "Demystifying evals for AI agents", isolated per-dimension judges: <https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents> (accessed 2026-09-14).
* Revisit when the anchor set exists and a candidate judge can be validated on it with TPR and TNR floors.
