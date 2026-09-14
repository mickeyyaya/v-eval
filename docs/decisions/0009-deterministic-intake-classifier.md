# Deterministic intake classifier with profile rules; model only for ambiguity

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

The maintainer redirected the pipeline away from a code-only shape: evaluation input is all history data received (briefs, chat transcripts, diffs, logs, CI output, retrieved passages, prior reports), and the system needs "an abstract layer to generalize the context and classifier to decide which criterions and adapter should be direct to." How should that classifier decide which criteria and adapters apply to a bundle of history?

## Decision Drivers

* Routing must be auditable: the report must say why each criterion and adapter was selected or skipped.
* Text inside the history may contain judge-directed instructions; routing must not be steerable by it.
* Weak local models must produce the same routing as strong ones.
* The [evaluation perspectives](../evaluation-views.md) already define profiles as sets of questions and evidence requirements.

## Considered Options

* Deterministic detection plus profile rules, model only for ambiguity, rationale recorded
* Model-driven classifier
* User-selected profiles only

## Decision Outcome

Chosen option: "Deterministic detection plus profile rules, model only for ambiguity, rationale recorded", because it is auditable, not gameable by history text, and identical across models. The core detects what was supplied (diff present, test files present, CI logs, sources with dates, transcript, prior report) and maps it through versioned perspective profiles to applicable criteria and runnable adapters. A model proposes only when the mapping is ambiguous, and its proposals are labeled provisional until confirmed; provisional criteria prevent an overall PASS. The routing rationale is written into every report.

### Consequences

* Good, because routing decisions are cheap to correct and easy to measure, which makes them the safest place for local learning ([0014](0014-adaptive-learning-precedent-bank.md)).
* Good, because the same profile files drive the skill text and the core.
* Bad, because profiles must be authored and versioned before the classifier is useful; the first release needs at least code, document, and context profiles.
* Bad, because messy mixed history may fall to the ambiguity path often at first.

## Confirmation

Profiles live as versioned data under `schema/` or `profiles/`; a core test feeds fixture bundles and asserts the selected criteria, adapters, and rationale; a test feeds a bundle containing judge-directed text and asserts routing is unchanged. None exists yet.

## Pros and Cons of the Options

### Deterministic detection plus profile rules

* Good, because auditable and identical across models.
* Good, because not steerable by injected text.
* Bad, because profiles must be authored first.

### Model-driven classifier

* Good, because it handles messy history flexibly.
* Bad, because unverifiable and sensitive to injected text.
* Bad, because it degrades badly on ollama-class models.

### User-selected profiles only

* Good, because simplest and fully predictable.
* Bad, because it ignores mixed history and pushes routing work onto the user.

## More Information

* Related requirements: REQ-02, REQ-05, REQ-10, and REQ-32 and REQ-33 in [requirements](../requirements.md). Informed by [integrity and RSI research](../research/2026-09-14-integrity-and-rsi.md) and [evaluator learning research](../research/2026-09-14-evaluator-learning.md).
* Shi et al., JudgeDeceiver, injected candidate text steering a judge: <https://arxiv.org/abs/2403.17710> (2024-03-26; accessed 2026-09-14).
* Ding et al., "Rubrics as an Attack Surface", criterion-preserving rubric edits shift accuracy up to 27.9%: <https://arxiv.org/abs/2602.13576> (2026-02-14; accessed 2026-09-14).
* Kumar, SWE-PRBench, more context hurts diff judgment: <https://arxiv.org/abs/2603.26130> (2026-03-27; accessed 2026-09-14).
* Laddha et al., SLMJury, small judges swing 10 to 25 points under ordering and persona: <https://arxiv.org/html/2606.07810> (2026-06; accessed 2026-09-14).
* v-eval [evaluation perspectives](../evaluation-views.md); evolve-loop signal-plane routing (session memory `project_user_defined_phases`, read locally 2026-09-14).
* Revisit when pilot data shows the ambiguity path dominates, or when a validated judge makes model-assisted routing measurably better.
