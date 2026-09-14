# Self-improvement loop: auto-propose, auto-reject, human-promote; human reactions are the reward

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

REQ-22 to REQ-24 require v-eval to evaluate its own changes with the same checker and to measure improvement rather than assert it. How is the loop triggered, budgeted, and governed, and who promotes a change? The maintainer chose the recommended option and added the governing nuance: "the most important thing is to keep the human reaction as followed the reinforcement learning from human feedback."

## Decision Drivers

* No named public protocol exists for a checker evaluating its own change; the closest published practice is a locked anchor set the loop never reads plus promotion as an auditable event.
* Documented loop failures: timeout self-edit, rubric tag-counter gaming, over-promotion, evaluator drift inverting conclusions, zero accepted edits under a noisy gate.
* Human feedback must be the reward signal; a model's self-assessed score must never drive promotion.
* Bounded execution: hard token and time caps, stopping conditions.

## Considered Options

* Auto-propose, auto-reject, human-promote, with a locked anchor set
* Human-gated batch only
* Two-track auto-promote
* Reuse evolve-loop's cycle as the loop

## Decision Outcome

Chosen option: "Auto-propose, auto-reject, human-promote, with a locked anchor set", with human reactions as the only promotion signal. A scheduled or manual run evaluates v-eval's own change with the same checker, may only discard candidates that fail the held-out gate, and queues survivors with anchor TPR and TNR deltas and at least three seeded runs. The maintainer promotes. Every human reaction (accept, reject, correction, override, question, promote, discard) is captured as a first-class reward record with provenance and is the ground truth for the loop; the anchor set and held-out gate measure whether human-driven changes improved revealed-information accuracy, they do not replace the human. Corrections are still cold re-judged before entering the example pool ([0014](0014-adaptive-learning-precedent-bank.md)). Hard token and time caps apply. Every promote or discard is an append-only audited event. Graduation to auto-promote is considered after about one hundred anchor labels.

### Consequences

* Good, because no unattended ratchet: the loop can only discard on its own.
* Good, because the human reaction log doubles as the learning signal store.
* Bad, because the queue starves under a noisy gate; with tens of items, per-round noise is near five points, so many candidates will be undecidable.
* Bad, because the maintainer is the bottleneck for every promotion.

## Confirmation

A reward-record table with provenance; a promotion log that is append-only; a test that a candidate cannot be promoted without a human record; a test that the loop never reads the anchor set. None exists yet.

## Pros and Cons of the Options

### Auto-propose, auto-reject, human-promote

* Good, because captures published gains without an unattended ratchet.
* Bad, because queue starvation under noise; human bottleneck.

### Human-gated batch only

* Good, because cheapest and safest.
* Bad, because no loop until the human starts one; rubber-stamping is easy.

### Two-track auto-promote

* Good, because scales; closest to published two-anchor designs.
* Bad, because most software, needs a large anchor set, second judge shares blind spots.

### Reuse evolve-loop's cycle

* Good, because proven machinery.
* Bad, because couples v-eval's self-evaluation to evolve-loop's release and profile system, contrary to [0001](0001-independent-of-evolve-loop.md).

## More Information

* Related requirements: REQ-22, REQ-23, REQ-24, and REQ-37 and REQ-38 in [requirements](../requirements.md). Informed by [integrity and RSI research](../research/2026-09-14-integrity-and-rsi.md) and [evaluator learning research](../research/2026-09-14-evaluator-learning.md).
* Nakajima, "Regimes", audited promotion ladder: <https://arxiv.org/pdf/2606.10241> (2026-06-08; accessed 2026-09-14).
* Zhang et al., "Who Grades the Grader?", locked anchor never read by any loop: <https://arxiv.org/html/2607.12790v1> (2026-07-14; accessed 2026-09-14; unreplicated).
* Li, "Who Drifted: the System or the Judge?", frozen anchor re-scoring: <https://arxiv.org/abs/2606.15474> (2026-06-13; accessed 2026-09-14).
* Sakana AI, Darwin Gödel Machine, sandbox plus human oversight: <https://arxiv.org/abs/2505.22954> (2025-05; accessed 2026-09-14).
* Sharma et al., sycophancy in preference data: <https://arxiv.org/abs/2310.13548> (accessed 2026-09-14).
* Revisit when anchors exceed about one hundred labels, or when the queue starves for three consecutive runs.
