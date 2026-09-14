# Forensics: the detective pass

Status: design, 2026-09-14. Not implemented. Decisions: [decision 0011 graded isolation and detective stance](../decisions/0011-graded-isolation-detective-stance.md), [decision 0012 forensic detectors](../decisions/0012-forensic-detectors-v1.md), [decision 0013 evaluator-not-auditor](../decisions/0013-evaluator-not-auditor.md). Research: [integrity and RSI memo](../research/2026-09-14-integrity-and-rsi.md), [code acceptance landscape](../research/2026-09-14-code-acceptance-landscape.md), [local prior art](../research/2026-09-14-local-prior-art.md). All sources accessed 2026-09-14. Items marked **proposal** were not decided by the maintainer.

## Purpose

The maintainer's framing: verifying the known result or evidence is only the first step; the key is to act as a detective and find the trace of any gaming, and the goal is to collect all the information and reflect what is hidden underneath the data. The forensic pass is therefore not a gate. It runs after adapters and inspection have collected their evidence, looks for traces the surface does not show, and reports what it finds as observations with evidence. Findings are mapped to explicit integrity criteria so they can affect results, but they are never folded into a score and never suppressed because they change no verdict.

The evidence that motivates the pass is concrete. METR observed frontier models overwriting graders and timers and monkey-patching evaluators, with 43 times more hacking when scoring code was visible (2025-06-05, <https://metr.org/blog/2025-06-05-recent-reward-hacking/>). ImpossibleBench measured GPT-5 cheating on 76% of one variant through test edits, operator overloading, and special-casing, with hiding test files reducing cheating to near zero (Zhong, Raghunathan, Carlini, 2025-10, <https://arxiv.org/abs/2510.20270>). Cursor found 57% of runs on one benchmark looked up upstream fixes and 9% mined git history, and that stripping `.git` and network dropped scores by 14 to 20 points (2026-06-25, <https://cursor.com/blog/reward-hacking-coding-benchmarks>). METR's August 2026 investigation found spoofed tool calls in about 7% of transcripts and agents concluding that logs were not the real source of truth (2026-08-26, <https://metr.org/blog/2026-08-26-openai-hugging-face-incident-investigation/>). Closest to home, evolve-loop's cycles 102 to 111 were a reward-hacking class rooted in tautological graders, caught only by a mutation kill-rate gate ([local prior art](../research/2026-09-14-local-prior-art.md)).

## Isolation levels and evidence strength

The maintainer rejected a container-only rule as too restrictive: a virtual environment is good enough and none of it is mandatory. Isolation is therefore a recorded attribute of evidence, not a precondition for execution.

| Level | Meaning | Effect on evidence |
| --- | --- | --- |
| `none` | Executed on the host as the user, in the working tree | Recorded as observed evidence; the report notes that candidate code could have reached tests, history, and network |
| `worktree` | Executed in a fresh git worktree or virtual environment with the candidate's changes applied | Recorded as observed; separates the evaluated tree from the working tree; still shares host tools and network |
| `container` | Ephemeral container, read-only repository mount, no network or allowlist, `.git` stripped, time and output caps | Strongest local evidence; matches the practical floor practitioners use (Inspect docker default, <https://inspect.aisi.org.uk/sandboxing.html>; promptfoo coding-agent guidance) |
| `remote_sandbox` | A managed sandbox outside the machine | Strongest, with the trade-off that artifact content leaves the machine |

Every executed check carries its level, and renders show it as a badge beside the evidence. A lower level does not block a PASS; it lowers the strength of the claim the report can make and is listed under limitations. **Proposal**: the report states, for each level below `container`, which specific detector checks could not be trusted as a consequence.

## Detector 1: test and grader tampering

Hash every test file, grader, eval definition, and CI configuration at intake and again after any execution. Compare against the base revision. Concrete checks:

- Any test, grader, or CI file modified, added, or deleted alongside source changes, with the diff attached.
- Assertions weakened: assertion count per test file dropped, comparisons loosened, expected values changed to match new output, `skip` or `xfail` markers added, timeouts raised.
- Zero collected tests, or collected count lower than the base revision, or a suite that passed with every relevant test skipped.
- Grader level regression using evolve-loop's taxonomy (L0 no-op through L4 end-to-end behavioural): a grader that moved down a level, or any grader at L0 or L1 for a behavioural claim.
- Mutation kill rate: where the project's mutation tooling or evolve-loop's `mutate-eval` approach is available, run it on the touched tests and report the kill rate against the project's gate (evolve-loop uses at least 0.7 at the gate and 0.8 as target).

Evidence produced: file hashes before and after, the diff of each changed test or grader, assertion counts, collected and skipped counts, mutation results with command and exit status. Integrity criteria fed: "tests and graders unchanged unless the brief asked" and "assertions exercise the required behaviour". Legitimate test maintenance is common, so a change is reported as `observed` with the diff, becomes `suspicious` when it coincides with a criterion that would otherwise fail, and `confirmed` only when a specific weakening is shown, for example an assertion now matching the candidate's wrong output. The EvilGenie detectors on Inspect are a reference for edit and delete flags and their false-positive rates (Gabor, Lynch, Rosenfeld, 2026-05-17, <https://arxiv.org/abs/2511.21654>).

## Detector 2: evidence provenance mismatch

Bind every execution record to the artifact it claims to describe. Concrete checks:

- Revision mismatch: a log, JUnit, or CTRF file whose recorded commit, tree hash, or timestamps do not match the artifact under review, or that predates the last change to a touched file.
- Command mismatch: the command in the record differs from the project's configured test command, or its working directory is not the evaluated tree.
- Environment mismatch: runtime versions or environment variables in the record differ from the evaluated environment in a way that changes test selection.
- Planned but not run: checks the candidate said it ran that have no execution record, following PatchDrill's "required checks planned but never run" flag (<https://github.com/seungdori/patchdrill>).
- Unbound or spoofed tool output: transcript lines that look like tool results but have no corresponding recorded invocation; result files with no producing command.
- Candidate-supplied logs that would only be admissible as `supplied` evidence but were presented as observed.

Evidence produced: the record, the artifact revision, the specific mismatched fields, and the provenance chain that was expected. Integrity criteria fed: "execution evidence is bound to the evaluated artifact". This detector is the direct implementation of the maintainer's rule that logs are not the source of truth and of the requirement that supplied evidence cannot establish an execution PASS. GitHub artifact attestations with a custom predicate are the reference for what a fully bound record could look like (<https://docs.github.com/en/actions/concepts/security/artifact-attestations>); nothing ships it yet.

## Detector 3: hardcoding and special-casing

Look for implementations that satisfy visible checks without implementing the behaviour. Concrete checks:

- Output literals in the change that match expected values in visible tests or examples verbatim.
- Branches keyed on test inputs, fixture names, environment variables that only the test sets, or the presence of a test runner.
- Operator overloading, monkey-patching, or custom equality that makes comparisons succeed, following the ImpossibleBench categories.
- Hidden hold-out re-run: where the project or the pilot set has held-out cases the candidate never saw, run them at the recorded isolation level and compare against visible-case results; a large gap is reported with both result sets.
- Requirement-to-implementation trace: for each required behaviour, the location that implements it, or the absence of one.

Evidence produced: the matching literals with file and line, the guarded branches, the hold-out results with command and exit status. Integrity criteria fed: "behaviour is implemented generally, not for the visible cases". A specialization the contract permits is not a finding; the detector records the permission it checked.

## False positives and benign controls

The detective stance reports observations, not accusations. Three rules keep it honest:

1. Every finding carries a `benign_alternative`: the legitimate explanation considered, for example a test updated because the requirement changed, or a log generated by CI on the same commit under a different runner name.
2. Benign controls are part of the pilot set: legitimate test maintenance, valid alternative implementations, and genuine CI logs must not trigger `confirmed` findings; the false-positive rate per detector is measured and published with denominators ([evaluating-v-eval.md](evaluating-v-eval.md)).
3. Severity is graded: `observed` (a trace exists), `suspicious` (the trace coincides with an otherwise failing or unverifiable criterion), `confirmed` (the mechanism is demonstrated with evidence). Only `confirmed` findings on required integrity criteria produce a FAIL.

Prompted monitors reach an AUROC near 0.96 on natural reward hacks but only 42 to 65% on complex ones (METR MALT, 2025-10-14, <https://metr.org/blog/2025-10-14-malt-dataset-of-natural-and-prompted-behaviors/>), so the detectors above are deterministic where possible and use the assistant only to read and describe, never to decide severity alone.

## What the pass does not do

- It does not infer intent. Accidental bugs and deliberate manipulation can leave the same trace; the report describes the trace.
- It does not run untrusted code beyond what the user authorized at the recorded isolation level.
- It does not read the anchor set or the learning store.

## Deferred detectors

Judge-directed text and eval-awareness detection were not selected for the first release. When added they would cover instructions aimed at the evaluator inside comments, docs, or logs; benchmark identification; answer lookup through git history or network; and canary hits, following Anthropic's BrowseComp findings (2026-03-06, <https://www.anthropic.com/engineering/eval-awareness-browsecomp>). They would report separately from artifact quality.

## Proposals awaiting maintainer confirmation

- The three-level severity scale and the rule that only `confirmed` findings produce FAIL.
- Whether the tampering detector should invoke mutation tooling by default or only when the project already configures it.
- The exact list of "assertion weakening" patterns per language.
