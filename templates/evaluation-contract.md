# Evaluation contract

Copy and fill this template for a real task. This is the human-readable form; the machine-readable schema under `schema/` will follow the pilot cases and is specified in [the report schema](../docs/architecture/report-schema.md). The contract is fixed for a comparison run, and no learned material may edit it.

## Task and authority

- Contract ID and version:
- Requested outcome and intended user:
- Original instruction or brief:
- Artifact scope and revision (commit plus dirty state, content hash, or document version):
- History data supplied (briefs, transcripts, diffs, logs, CI output, sources, prior reports), each with origin and date:
- Relevant design and trusted sources, with versions and dates:
- Evaluation profile and version; selected perspectives and the reason for each:
- Conflicting instructions and their resolution, if any:
- Contract status: user-specified, provisional, or approved; record the actual source of approval if applicable.

## Criteria

Repeat a row for each distinct requirement. Required means it can block acceptance; optional means advisory. Integrity criteria for the forensic pass are listed here too, so a confirmed tampering trace has a criterion to fail.

| ID | Requirement and source | Required? | Applicability condition | Perspective and evaluation method | Acceptance rule | Expected evidence |
| --- | --- | --- | --- | --- | --- | --- |
| C1 | State the observable requirement and link its origin | Yes / No | State when it applies | Name the perspective; test / deterministic / inspection / source / rubric / human | Define pass and fail conditions, including rubric anchors if used | Identify the evidence needed |
| I1 | Tests, graders, and evaluation configuration are unchanged, or changes are reviewed and justified | Yes | Always | Adversarial reviewer; deterministic hash comparison | FAIL on unreviewed modification, weakened assertion, added skip, or zero collected tests | Before and after hashes, diff of test files |
| I2 | Supplied execution evidence matches this exact artifact | Yes | When logs or results are supplied | Evidence integrity; provenance comparison | FAIL on revision, command, or environment mismatch; UNKNOWN when provenance is absent | Revision, command, environment, timestamps |

## Execution and data boundaries

- Existing test or check commands and their trusted origin:
- Authorized execution environment and its isolation level (none, worktree or virtual environment, container, remote sandbox):
- Source or artifact data permitted for any remote judge:
- Checks unavailable in this environment:

## Dimension metrics, when requested

| Dimension / criterion | Metric and version | Authority type and definition source | Unit / range / direction | Threshold and its source | Tool, configuration, and scope |
| --- | --- | --- | --- | --- | --- |
| Select a relevant perspective | Use a named measure | Formal standard / established metric / vendor rating / project rubric; link its definition | Preserve native units | State an industry threshold if defined, otherwise an explicit project budget | Record artifact, workload, source corpus, and tool version |

Attach measured values, supporting evidence, interpretation, and limitations to the evaluation record. If a measurement is unavailable, leave its value absent and state why. Missing measurements are not zero. No composite is produced by default. See [the industry scorecard](../docs/industry-scorecard.md).

## Evaluation record

- Evaluation date; evaluator, tool, and model identity when available:
- Skill, rubric, core, and profile versions:
- Exact artifact and contract identity:
- Routing rationale (perspectives, criteria, and adapters selected or skipped, and why):
- Executed commands, working directories, isolation levels, collected and skipped tests, exit statuses, and logs:
- Supplied evidence whose provenance is not verified:
- Learned material referenced (precedent IDs), if any:

### Observations

List everything found during collection, including items tied to no criterion, each with a location or quoted passage.

### Claim-to-verification table

| Claim (from summary, description, or log) | Origin | Verification action | Result | Evidence reference |
| --- | --- | --- | --- | --- |
| Quote the claim | Candidate / tool / user | Command run, file opened, passage compared, or none | Verified / Contradicted / UNVERIFIED | Location, command and exit status, or passage |

### Per-criterion results

| Criterion ID | Result | Method actually used | Isolation level | Evidence reference and observation | Limitation or next action |
| --- | --- | --- | --- | --- | --- |
| C1 | PASS / FAIL / UNKNOWN / ERROR / NOT_APPLICABLE | Distinguish execution, inspection, and judgment | none / worktree / container / remote | Point to a file and line, source passage, or run record; a PASS needs an opened-or-ran citation | State unresolved evidence or the specific repair |

### Forensic findings

| Detector | Observation | Evidence | Integrity criterion | Result |
| --- | --- | --- | --- | --- |
| Tampering / provenance / hardcoding | Describe the observable fact, not a motive | Hashes, diffs, mismatched fields, matched literals | I1 / I2 / ... | PASS / FAIL / UNKNOWN |

Use the acceptance policy in [the design](../docs/design.md#results-and-acceptance). Show required and optional counts separately, exclusions with reasons, coverage with numerator and denominator, overall status, and the scope of the conclusion. Do not fill missing run results with assumptions.
