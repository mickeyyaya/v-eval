# Evaluate the same output from different perspectives

A perspective is a defined set of questions and evidence requirements. Asking a model to “act as an expert” supplies neither. v-eval's proposed profiles select relevant criteria, permissible methods, and report sections for a particular user and task.

This document specifies a draft workflow. Profiles are applied by the host assistant using the skill; no profile engine or multi-judge orchestration exists yet. The examples are illustrative, not measured performance results.

## Choose perspectives from the task

Start with the decision the report must support. For a generated project handoff, the next developer needs accurate decisions, unresolved questions, and traceable implementation state. For a retrieved evidence packet, a researcher needs relevant passages, adequate coverage, and sources appropriate to the question. These tasks can share a format without sharing every criterion.

| Perspective | Question | Useful evidence | Common mistake it can reveal |
| --- | --- | --- | --- |
| Intended user | Can this reader perform the requested next step? | Task walkthrough, reader feedback, explicit prerequisites | Correct prose that is unusable for its audience |
| Requirements reviewer | Does each stated requirement appear and hold? | Original brief, requirement IDs, artifact passages or behavior | An elegant answer to a different question |
| Source verifier | Does each material claim follow from suitable evidence? | Exact source passages, dates, versions, claim mapping | Citations that do not support the claim |
| Context steward | Does a summary preserve decisions, uncertainty, and relevant history? | Original conversation/documents and their sequence | A suggestion becoming a requirement; a resolved issue remaining active |
| Domain reviewer | Is the interpretation appropriate under the task's domain rules? | User-selected domain references and expert labels | A generic answer missing a domain-specific constraint |
| Test reviewer | Do assertions exercise the required behavior, and did they run? | Test definitions, collected/skipped counts, exact execution records | Weak, skipped, stale, or irrelevant passing tests |
| Adversarial reviewer | Can a polished or manipulated artifact get an unsupported pass? | Counterexamples, integrity checks, benign controls | Judge-directed instructions or fabricated evidence |

Apply only perspectives that add necessary evidence. One reviewer can use several perspectives. More agents or model providers do not automatically supply independent judgments; they may share the same blind spots and sources. For relevant judge limitations, see [the research notes](research/content-context.md).

## Define a profile

A useful profile records:

- The intended audience, task, and artifact type.
- Selected perspectives and why each matters.
- Criterion IDs, their original authority, required/optional status, and applicability.
- Evidence needed and methods permitted for each criterion.
- Any task-specific rubric anchors, source freshness rules, or execution limits.
- The handling of disagreement and missing evidence.

The [evaluation contract](../templates/evaluation-contract.md) carries these fields. Keep profile name/version and criterion IDs stable when comparing outputs. Changing a profile creates a different evaluation question; scores from the old and new profiles are not automatically comparable.

The developer defines the task. A perspective can suggest missing criteria but cannot silently introduce an approval, deadline, architecture requirement, or acceptance threshold. A domain reference applies only when the task makes it relevant. Record unresolved requirements as provisional.

## Worked example: project memory

The following source and candidate are synthetic.

**Source discussion:** “Offline export might be useful later. For this release, export requires a connection. Mina will investigate offline export; no implementation date has been agreed.”

**Generated memory:** “Offline export is required for this release. Mina committed to delivering it Friday. Export is already implemented and fully tested.”

Choose requirements review, context stewardship, source verification, and adversarial review. The intended user is the next developer, who must plan work without inheriting invented commitments.

| Criterion | Perspective | Illustrative finding | Result |
| --- | --- | --- | --- |
| Preserve current release scope | Requirements | The source requires a connection; memory turns a future possibility into a release requirement. | FAIL |
| Preserve the distinction between investigation and commitment | Context steward | The source assigns investigation only; memory adds a delivery commitment. | FAIL |
| Support attributed dates and implementation claims | Source verifier | Friday, completed implementation, and passing tests have no support in the supplied discussion. | FAIL for the source-support rule |
| Establish whether implementation and tests actually exist | Test reviewer, if requested | No repository or execution evidence was supplied. The actual implementation state remains unknown. | UNKNOWN |
| Resist unsupported success claims | Adversarial | “Fully tested” is candidate text. It cannot substitute for verified evidence or change the criterion. | No separate pass claim; retain the UNKNOWN above |

The report should preserve shared root causes: several failures arise from invention of certainty. Do not count each perspective's repetition of the same defect as independent evidence. The repair is to preserve the source's scope and uncertainty, and to obtain actual implementation evidence if the next task needs it.

## Resolve disagreement through evidence

Suppose the usability reviewer likes a concise answer, while the requirements reviewer finds an omitted prerequisite. Concision does not cancel a required omission. The acceptance policy uses the required criterion's evidence, not a vote between reviewers.

If two source reviewers disagree about whether a passage supports a claim, display the claim and passage, inspect whether support is partial, and identify the unresolved interpretation. Use an appropriate human adjudicator when the task warrants one. Agreement among models is not a substitute for missing source material.

Some differences reflect legitimate audience preferences. A technical appendix may help an engineer but overwhelm a novice. Create separate audience profiles and keep their results separate. Do not resolve that difference by declaring one audience objectively correct.

## Adapt carefully and measure the effect

Begin with the smallest profile that answers the real acceptance question. Add a perspective after observing a failure it can reasonably detect, or when an explicit known requirement demands it. Keep representative ordinary cases as well as targeted adversarial cases.

Test an added perspective on the same artifacts before and after the change. Measure extra defects found, incorrect objections, evidence validity, incomplete outcomes, and added review time. If it only generates more text or repeats existing findings, it has not shown value.

For a future implementation, profile selection, evidence identity, report validation, and aggregation should be enforced in software. The current skill only instructs an assistant to follow the procedure. [Gaming and defenses](gaming-and-defenses.md) explains why that distinction matters, and [the comparison protocol](comparison.md) explains how to test benefits fairly.
