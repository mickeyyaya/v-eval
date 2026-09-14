# Adaptive learning: a local precedent bank with a guardrail skeleton; rule ladder later

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

The maintainer asked whether the skill can be adaptive, learning from the user's questions and corrections to improve accuracy locally. REQ-22 to REQ-24 already require measured self-improvement. The research found what measurably works, what does not, and how learning from a single user goes wrong. Which adaptive design should v1 ship?

## Decision Drivers

* Learning must be local (privacy, no service, works with ollama) and must run on every CLI and OS.
* Retrieval of past corrected cases lifted three of four ollama-class judges by 5 to 12 points and hurt one.
* Run-time reflection has negative information gain on subjective evaluation; sampling beats it.
* Memory-driven agents learned "what pleased the user" and lost 45% of their refusals; follow-up rebuttals flip verdicts 56 to 86% of the time; a visible prior score blocks 48% of corrections; in-sample gains of +0.18 collapsed to +0.04 held-out.
* The learner must never edit the trusted contract or read the anchor set.

## Considered Options

* Precedent bank plus guardrail skeleton now; rule ladder later
* Rule ledger with promotion gates from day one
* Offline prompt optimization driven
* No learning in v1: logging and anchor set only

## Decision Outcome

Chosen option: "Precedent bank plus guardrail skeleton now; rule ladder later", because it has the smallest misevolution surface, every learned item is one deletable row, and it is the external-evidence mechanism the reflection literature supports. Design:

* Log every evaluation and correction to a local SQLite store with provenance: artifact hash, criterion id, v-eval verdict, user reaction, rationale, model and prompt hash, skill version, timestamp, project id.
* Embed `criterion + excerpt + correction` with a local embedder (ollama, `nomic-embed-text` class) into sqlite-vec; at judge time retrieve at most five train-split precedents for the same criterion, with prior scores and verdict text stripped.
* Guardrails: a frozen human-labeled anchor set the learner never reads; corrections are re-judged cold (fresh context, no prior verdict, no user text) before entering the example pool; rationale before verdict; the contract is immutable to the learner; per-row provenance and a 90-day staleness flag; versioned bundle digest.
* Later: candidate-to-active rules with auto-disable once anchors exceed about fifty cases. Prompt optimization is a maintainer tool run offline, never a runtime feature. Run-time reflection is not used as an accuracy lever.

### Consequences

* Good, because learning is auditable and reversible per row.
* Good, because it runs identically on every harness with a script and SQLite.
* Bad, because generalization is limited to similar cases; the local embedder's recall is the ceiling.
* Bad, because the anchor set must be authored before any accuracy claim, and sub-ten-point changes at fifty labels are within noise.

## Confirmation

The store schema exists with the provenance fields; a test asserts anchors are never returned by retrieval; a test asserts a correction without a cold re-judge cannot enter the pool; the promotion path is gated on non-regressing anchor TPR and TNR. None exists yet.

## Pros and Cons of the Options

### Precedent bank plus guardrail skeleton

* Good, because reversible, auditable, portable, and supported by the evidence on external signal.
* Bad, because limited generalization.

### Rule ledger with promotion gates from day one

* Good, because it generalizes faster and is cheap at evaluation time.
* Bad, because rules encode the user's blind spots and need the anchor set first.

### Offline prompt optimization

* Good, because largest measured gains (GEPA +6 to +13 points).
* Bad, because 79 to 737 rollouts or 200 plus labels, opaque rewrites, overfits to one user, cannot run inside a portable skill.

### No learning in v1

* Good, because safest.
* Bad, because the skill does not adapt at all.

## More Information

* Related requirements: REQ-22, REQ-23, REQ-24, and REQ-37 and REQ-38 in [requirements](../requirements.md). Informed by [adaptive skills research](../research/2026-09-14-adaptive-skills.md) and [evaluator learning research](../research/2026-09-14-evaluator-learning.md).
* Dussert, govllm, five few-shot examples per criterion on ollama-class judges: <https://arxiv.org/abs/2605.24737> (2026-05-23; accessed 2026-09-14; single source, n about 44).
* Tao et al., reflection has negative information gain on subjective evaluation: <https://arxiv.org/abs/2607.28908> (2026-07-31; accessed 2026-09-14).
* Shao et al., "Your Agent May Misevolve", ICLR 2026: <https://arxiv.org/abs/2509.26354> (accessed 2026-09-14). Kim and Khashabi, sycophancy under rebuttal: <https://arxiv.org/abs/2509.16533> (accessed 2026-09-14). Kapetanovic et al., anchoring bias in judges: <https://arxiv.org/html/2608.25869> (accessed 2026-09-14).
* Agrawal et al., GEPA, ICLR 2026: <https://arxiv.org/abs/2507.19457> (accessed 2026-09-14). Zhang et al., "Who Grades the Grader?": <https://arxiv.org/html/2607.12790v1> (2026-07-14; accessed 2026-09-14). PROCTOR: <https://arxiv.org/html/2609.02246> (2026-09-02; accessed 2026-09-14).
* sqlite-vec: <https://github.com/asg017/sqlite-vec> (accessed 2026-09-14). ECC continuous-learning-v2 instinct schema: <https://github.com/affaan-m/everything-claude-code/blob/main/skills/continuous-learning-v2/SKILL.md> (accessed 2026-09-14).
* Revisit when the anchor set exceeds about fifty labels (enable the rule ladder) or about one hundred (consider auto-promotion), or when retrieval shows no measured gain for a model.
