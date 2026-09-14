# Worked evaluation report

**Overall: FAIL.** The supplied function violates case-insensitive deduplication and order preservation. Regression execution is unverified.

This is an illustrative inspection report authored with the example, not a measured validation of the skill. No code or test suite was executed. Artifact identity: the complete function under [Candidate artifact in input.md](input.md#candidate-artifact), as shipped in the same repository revision as this report. Contract: the five supplied acceptance criteria in that input; all are required and applicable.

| ID | Result | Method | Evidence and reasoning | Next action |
| --- | --- | --- | --- | --- |
| C1 | FAIL | Static inspection | The candidate strips whitespace but puts original-case strings in a set. For `[' A@x.test ', 'a@x.test']`, the two stripped strings remain distinct by Python string comparison. This is a reasoned counterexample, not an executed result. | Deduplicate using a normalized comparison key. |
| C2 | FAIL | Static inspection | `sorted(...)` replaces input order with sorting order. `['z@x.test', 'a@x.test']` therefore cannot preserve first-seen order. | Append first occurrences to an output list while tracking seen keys. |
| C3 | FAIL | Static inspection | `['A@x.test', 'a@x.test']` retains both spellings, including the second occurrence. This violates the rule to keep only the first occurrence's spelling for an equivalent address. | Store the first trimmed value for each normalized key. |
| C4 | PASS | Static inspection | The complete supplied function uses built-ins and a string method; it imports or calls no third-party dependency. The conclusion covers only this snippet. | None. |
| C5 | UNKNOWN | Execution evidence unavailable | The candidate author's statement supplies no tests, command, environment, logs, or verified artifact binding. | Provide and run the regression suite against the exact artifact in an authorized environment. |

Required counts: **5 applicable; 1 PASS; 3 FAIL; 1 UNKNOWN; 0 ERROR; 0 excluded.** Optional criteria: none. Assessment coverage: **4/5**. This fraction is not a correctness score.

Passing a future regression suite would not by itself resolve C1–C3 unless its assertions exercise the stated requirements. Findings may overlap in root cause; they should not be counted as three independent implementation defects.
