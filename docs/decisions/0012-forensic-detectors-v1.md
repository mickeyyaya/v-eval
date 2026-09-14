# Forensic detectors in v1: tampering, provenance mismatch, hardcoding

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

With the detective stance fixed in [0011](0011-graded-isolation-detective-stance.md), which gaming-trace detectors should the forensic pass carry first? The maintainer added a framing that governs how findings are reported: "you are evaluator, not the verdict auditor, our goal is to collect all the information and reflect the true hidden underneath the data" ([0013](0013-evaluator-not-auditor.md)).

## Decision Drivers

* Test and grader tampering is the documented failure class in evolve-loop (cycles 102 to 111) and in METR's observations.
* Stale, fabricated, or unbound logs are the concrete threat cases in the [gaming guide](../gaming-and-defenses.md).
* Hardcoded outputs and special-casing are the cheapest way to satisfy visible examples.
* Judge-directed text and eval-awareness are real but rarer in a local developer workflow, and their detection is less mature.

## Considered Options

* Test and grader tampering
* Evidence provenance mismatch
* Hardcoding and special-casing
* Judge-directed text and eval awareness

The maintainer selected the first three.

## Decision Outcome

Chosen options: "Test and grader tampering", "Evidence provenance mismatch", and "Hardcoding and special-casing". Judge-directed text and eval-awareness detection are deferred to v2. Detector findings are observations with locations and quotes, mapped to explicit integrity criteria; they are never folded into a score.

Detector outlines:

* Tampering: hash tests, graders, CI configuration, and eval definitions before and after; flag edits, weakened assertions, added skips, zero collected tests, and mutation kill-rate drops, reusing the evolve-loop mutate-eval approach.
* Provenance mismatch: logs or results whose revision, command, environment, or timestamps do not match the artifact; planned-but-not-run checks; spoofed or unbound tool output.
* Hardcoding: outputs matching visible examples verbatim, branches keyed on test inputs, operator overloading or monkey-patching of comparisons, with hidden-holdout re-runs to confirm.

### Consequences

* Good, because the three detectors cover the failure classes with the strongest local and published evidence.
* Good, because each finding lands in the observations section first, then maps to an integrity criterion, matching the evaluator-not-auditor principle.
* Bad, because prompted monitors reach only 42 to 65% on complex hacks; detectors will miss cases and must say what they did not check.
* Bad, because judge-directed text remains undetected in v1; the classifier's immunity to it ([0009](0009-deterministic-intake-classifier.md)) is the only mitigation.

## Confirmation

Each detector has fixture pairs (clean versus tampered, current versus stale log, honest versus hardcoded) with benign controls; the pilot reports detection and false-positive counts per detector; findings appear in the observations section of the report. None exists yet.

## Pros and Cons of the Options

### Test and grader tampering

* Good, because it is the maintainer's own documented failure class with a working mutation gate.
* Bad, because hashing only detects changes inside the reviewed tree.

### Evidence provenance mismatch

* Good, because it directly implements "logs are not the source of truth".
* Bad, because it depends on the artifact and logs carrying revision metadata.

### Hardcoding and special-casing

* Good, because it catches the cheapest gaming path.
* Bad, because hidden-holdout re-runs need execution and can be legitimately specialized code.

### Judge-directed text and eval awareness

* Good, because it is a real attack class.
* Bad, because detection is immature and adaptive attacks defeat text filters; deferred.

## More Information

* Related requirements: REQ-14, and REQ-29 and REQ-30 in [requirements](../requirements.md). Informed by [integrity and RSI research](../research/2026-09-14-integrity-and-rsi.md) and [gaming and defenses](../gaming-and-defenses.md).
* METR, MALT dataset, monitors at AUROC 0.96 on natural hacks: <https://metr.org/blog/2025-10-14-malt-dataset-of-natural-and-prompted-behaviors/> (2025-10-14; accessed 2026-09-14).
* Gabor, Lynch, Rosenfeld, EvilGenie, test-edit flags and holdout tests: <https://arxiv.org/abs/2511.21654> (v2 2026-05-17; accessed 2026-09-14).
* Stryker mutation score definition: <https://stryker-mutator.io/docs/mutation-testing-elements/mutant-states-and-metrics/> (accessed 2026-09-14).
* Thaman, Reward Hacking Benchmark: <https://arxiv.org/abs/2605.02964> (2026-05-03; accessed 2026-09-14; single source).
* Anthropic, eval awareness in BrowseComp: <https://www.anthropic.com/engineering/eval-awareness-browsecomp> (2026-03-06; accessed 2026-09-14).
* evolve-loop, `runtime/docs/eval-grader-best-practices.md`, mutate-eval kill-rate gate, read locally 2026-09-14.
* Revisit when pilot data shows judge-directed text in real cases, or when a detector's false-positive rate blocks legitimate work.
