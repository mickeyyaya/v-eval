# Repository maintenance tooling is exempt from the product's no-interpreter rule

* Status: accepted
* Deciders: maintainer (mickeyyaya), on an architecture-review recommendation
* Date: 2026-09-14

## Context and Problem Statement

[Decision 0020](0020-portability-constraints.md) requires every product script and program to run on macOS, Linux, and Windows with no interpreter or shell on the required path, because the skill's `scripts/` launchers run inside agent hooks on user machines. The repository also needs maintainer-only tooling: a generator for the source register (`docs/research/sources.md`, which lists every external link cited in the documentation) and a fast local link check. Those tools run only on maintainer machines and in continuous integration, never on a user's machine or inside a hook. Where should they live, what may they be written in, and how do they relate to the product's rule and to the CI link checker?

## Decision Drivers

* The maintainer's instruction to keep every source link and to verify it, which the generator makes checkable.
* The no-interpreter rule must stay unambiguous for contributors: one word, `scripts/`, must not mean both "strictest zone" and "exempt zone".
* The Go core does not exist yet; writing Go before `go.mod` exists is the wrong order, and a released binary must not carry maintainer-only subcommands as a public contract.
* CI already runs lychee for link and anchor checking; two checkers with different slug rules would disagree by construction.

## Considered Options

* A. Keep Python tooling under `scripts/docs/`.
* B. Move the tooling to `tools/docs/`, declare the exemption, keep lychee as the CI authority for links, and plan a port to an internal Go tool once `core/` exists.
* C. Add `veval docs sources` and `veval docs links` subcommands to the future Go core.
* D. No tooling; rely on CI and manual bibliography maintenance.

## Decision Outcome

Chosen option: **B**.

* Maintainer tooling lives under `tools/`, never under `scripts/`, and is never shipped in the plugin or on the skill's required path. It may use any toolchain available to maintainers and CI; today that is Python 3 with the standard library only.
* The exemption is scoped: decision 0020 governs the product's required path; this record governs repository maintenance only.
* Each tool is invocable on its own: `gen_sources.py` and `check_links.py` carry their own command lines and never import each other; `vdocs.py` is a routing front only. Tests are written unittest-style with the standard library so the tooling has no dependency; they also run under pytest, which remains the project's framework for product code.
* The source-register generator is the one tool with no off-the-shelf substitute and is kept, with a `--check` mode so CI fails when the committed register is stale.
* Link checking has one authority in CI: lychee with fragment checking. The local link checker is a convenience for a fast pre-commit pass, documented as advisory, with its simplified anchor algorithm stated; it does not run in CI.
* When `core/` exists, the generator is ported to an internal Go tool run with `go run ./tools/...`, not a `veval` subcommand, and the Python is deleted.

### Consequences

* Good: the no-interpreter rule stays crisp; contributors see the exemption in one place; the register cannot silently drift.
* Good: no maintainer-only commands leak into the released binary.
* Bad: a second toolchain exists until the Go port; the local link checker and lychee can disagree on unusual anchors, which is why lychee decides.

## Confirmation

* `tools/docs/` exists; no Python appears under `scripts/` or `skills/*/scripts/`.
* CI runs the tooling's unit tests on macOS, Linux, and Windows and runs the generator in `--check` mode.
* The local link checker's docstring states that lychee in CI is the authority.

## Amendments (2026-09-15)

`core/` now exists, so the last bullet of the Decision Outcome -- port the source-register generator to an internal Go tool run with `go run ./tools/...` and delete the Python -- has become due. It has not been done. The Python tooling under `tools/docs/` stays as it is until the port is scheduled, and the port is now a Roadmap Stage 3 item. Doing it inside the walking skeleton would have put a second unrelated thing at risk in one branch, and the register is verified in CI by `vdocs.py register . --check` either way.

### One third-party package, in CI only

`tools/schema/validate.py` validates JSON instances against a JSON Schema and depends on the third-party `jsonschema` package: the standard library has no JSON Schema validator, and writing one is not repository maintenance. This is the one exception to "standard library only" above, and it is scoped the same way the exemption itself is: the package is installed only in continuous integration (`.github/workflows/go.yml`, job `schemas`, pinned to `jsonschema==4.26.0`) and by a maintainer on demand with `pip install jsonschema`. The product and every other tool stay standard-library. The tool's own unit tests need no package: the missing-package path is tested with the import mocked absent, and the tests that validate real instances are skipped, with a reason, wherever the package is not importable, so they run on all three operating systems in `docs.yml` without installing the package. The job validates the SARIF fixtures under `core/export/testdata/` against the OASIS SARIF 2.1.0 schema, fetched into `build/` at run time, and the report fixtures under `core/report/testdata/` against `schema/report.schema.json`.

## Pros and Cons of the Options

### A. `scripts/docs/`

* Good: no move needed.
* Bad: the same directory name denotes the strictest and the exempt zones; the architecture document lists only `skills/evaluate-output/scripts/`.

### B. `tools/docs/` with a declared exemption

* Good: unambiguous placement; smallest change; keeps the register verifiable now.
* Bad: Python remains until the Go port.

### C. Subcommands of the Go core

* Good: single toolchain eventually.
* Bad: public compatibility surface for maintainer chores; premature before the core exists.

### D. No tooling

* Good: nothing to maintain.
* Bad: the source-preservation instruction becomes an unverifiable claim.

## More Information

* Architecture review of the maintenance scripts, 2026-09-14, in-session (recommendations 1 to 5 adopted; deletion of the local link checker replaced by demotion to advisory because lychee parity could not be verified locally on that day).
* [Decision 0020](0020-portability-constraints.md); [packaging and portability](../architecture/packaging-and-portability.md); the documentation workflow at `.github/workflows/docs.yml`.
* lychee fragment checking: <https://github.com/lycheeverse/lychee> (accessed 2026-09-14).
