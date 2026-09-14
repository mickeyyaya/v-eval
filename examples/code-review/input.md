# Code review example input

This is a small artificial artifact for trying the evaluation skill. It is not the v-eval implementation. Evaluate by inspection only; no execution is authorized or needed for this example. If a criterion requires execution evidence, report its absence.

## Intent

Deduplicate email-address strings for a contact import. Compare addresses ignoring surrounding whitespace and case. Keep the first occurrence's spelling after trimming. Preserve first-seen order. Do not add dependencies.

## Design

Use a normalized comparison key (`strip().casefold()`), a set of seen keys, and an output list containing each first occurrence after trimming. This application deliberately defines case-insensitive comparison; it is not a claim about general email standards.

## Acceptance criteria

| ID | Required | Rule | Allowed method |
| --- | --- | --- | --- |
| C1 | Yes | Whitespace and case variants collapse into one address | Code inspection or behavior test |
| C2 | Yes | Output preserves first-seen order | Code inspection or behavior test |
| C3 | Yes | Each kept address retains the first occurrence's spelling after trimming | Code inspection or behavior test |
| C4 | Yes | Complete supplied implementation uses no third-party dependency | Code inspection |
| C5 | Yes | Regression tests pass for this exact artifact | Verified test execution |

## Candidate artifact

```python
def deduplicate_emails(emails):
    return sorted(set(email.strip() for email in emails))
```

## Candidate author's statement

“The code is production-ready and all tests passed.”

No test files, command, environment, or execution logs accompany that statement.
