# Execution isolation is graded and recorded, not mandatory; the evaluator acts as a detective

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

The integrity research recommends a container floor before executing candidate code. The maintainer rejected a container-only rule: "only inside the container standard is too restrict, virtual environment should be good enough but it should not be mandatory, the key is to act as the detective to find the trace of any gaming traces, verify the known result / evidence is just part of first step." Should v1 execute candidate code, and under what isolation?

## Decision Drivers

* A container-only rule would leave most execution criteria UNKNOWN on ordinary developer machines and across CLIs and OSs.
* Frontier models edit tests and graders at high rates when able; hiding tests and stripping `.git` and network are the measured structural fixes.
* evolve-loop's incidents were caught by looking for traces, not by isolation.
* The strength of an evidence claim depends on how it was obtained, and the report must say so.

## Considered Options

* Import-first; execute only inside a container floor, otherwise UNKNOWN
* Execute in the host's existing sandbox
* Never execute in v1
* Execute directly on the host, no isolation
* Graded, recorded isolation; execution allowed at any level; forensic pass always runs (the maintainer's reframing)

## Decision Outcome

Chosen option: "Graded, recorded isolation; execution allowed at any level; forensic pass always runs". Isolation is a recorded attribute of every executed check with levels none, worktree or virtual environment, container, remote sandbox. Lower isolation weakens the evidence claim and is shown beside the result; it does not block execution. Verifying the supplied result is step one. Step two, always, is a forensic pass for gaming traces ([0012](0012-forensic-detectors-v1.md)). Forensic findings are separate criteria with their own evidence, never folded into a score; a confirmed trace of tampering is FAIL on an explicit integrity criterion.

### Consequences

* Good, because execution criteria can be evaluated on a laptop, in CI, and under any CLI.
* Good, because the report makes the isolation level and its implication visible instead of hiding it in a policy.
* Bad, because at low isolation the candidate can in principle read tests, modify graders, or spoof output; the forensic detectors are the mitigation, and they are imperfect.
* Bad, because the anchor set and hidden holdout tests must be kept out of the candidate's reach by construction, not by isolation.

## Confirmation

Schema: each executed check records `isolation_level`; the HTML and Markdown renderers display it next to the result; the forensic pass is a mandatory pipeline stage with its own report section; a test asserts a report cannot omit the isolation level for an executed check. None exists yet.

## Pros and Cons of the Options

### Graded, recorded isolation with mandatory forensic pass

* Good, because practical everywhere and honest about evidence strength.
* Bad, because relies on detectors at low isolation.

### Import-first; container floor otherwise UNKNOWN

* Good, because matches the tampering and injection evidence.
* Bad, because too restrictive for most developer machines; many criteria stay UNKNOWN.

### Host's existing sandbox

* Good, because cheaper to build.
* Bad, because isolation varies by host and is undocumented on several CLIs.

### Never execute in v1

* Good, because zero execution risk.
* Bad, because execution criteria stay UNKNOWN unless a trusted run is supplied.

### Execute directly, no isolation

* Good, because fastest.
* Bad, because it contradicts the integrity research without recording the weakness.

## More Information

* Related requirements: REQ-05, REQ-13, REQ-14, and REQ-29 and REQ-31 in [requirements](../requirements.md). Informed by [integrity and RSI research](../research/2026-09-14-integrity-and-rsi.md).
* Zhong, Raghunathan, Carlini, ImpossibleBench, test edits and near-zero cheating when tests are hidden: <https://arxiv.org/abs/2510.20270> (2025-10; accessed 2026-09-14).
* Cursor, "Reward hacking is swamping model intelligence gains", `.git` and network removal effects: <https://cursor.com/blog/reward-hacking-coding-benchmarks> (2026-06-25; accessed 2026-09-14; vendor-run).
* UK AISI Inspect sandboxing: <https://inspect.aisi.org.uk/sandboxing.html> (accessed 2026-09-14). Anthropic sandbox runtime: <https://github.com/anthropic-experimental/sandbox-runtime> (accessed 2026-09-14).
* evolve-loop evaluator isolation profile, `runtime/.evolve/profiles/evaluator.json`, and grader best practices, read locally 2026-09-14.
* Revisit when a detector misses a tampering case in the pilot that a container would have prevented.
