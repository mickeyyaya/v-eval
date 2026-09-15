# Changelog

All notable changes to this project are documented in this file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html) from the first tagged release, 0.1.0.

## [Unreleased]

### Added

- `veval version` falls back to the Go build information when nothing was stamped: a `go install` build names its module version, and a checkout build names its commit (with `-dirty` when the tree was modified).
- Validation reports the schema-version mismatch alongside the unknown-field error when a report from a newer schema is read, so the reader learns which build to use instead of only which field was unknown.
- One-line skill installation through the cross-agent `skills` installer (`npx skills add mickeyyaya/v-eval`), verified byte-identical to the repository, documented in `README.md`.

### Fixed

- HTML report: each fact keeps its label and value together across column breaks (`-webkit-column-break-inside` beside `break-inside`).

### Changed

- `veval <command> -h` prints that command's usage on standard output and exits 0, matching `veval -h`; unknown flags and missing values still exit 2 on standard error.

## [0.1.0] - 2026-09-15

First tagged release: the core walking skeleton, the `veval` command line, the skill wired to it, and the release process itself.

### Added

- Consolidated requirements record with sources U1 to U19 (`docs/requirements.md`).
- Research reports with citations and access dates (`docs/research/`).
- Twenty-two decision records in MADR format (`docs/decisions/`).
- Architecture design: pipeline, report schema, learning loop, forensics, packaging and portability, evaluating v-eval (`docs/architecture/`).
- Learning handbook, evaluation contract template, draft skill, worked example.
- Community files: code of conduct (Contributor Covenant 2.1), security policy, contributing guide, citation file, editor config, issue and pull request templates.
- Documentation workflow (markdownlint and link check on every push and pull request) and maintainer tooling under `tools/docs/` (`vdocs` front and standalone `gen_sources` and `check_links` command lines) for the generated source register (`docs/research/sources.md`, verified in CI with `--check`) and a local link check, with unit tests that run on macOS, Linux, and Windows in CI (decision 0022).
- Report JSON Schema v0.1.0 (`schema/report.schema.json`), embedded in the core binary, and the matching Go types (`core/report`), with a drift test that fails when the types and the schema disagree.
- Report validation rules (`core/report`): the evidence rule, which refuses a criterion `PASS` unless at least one evidence record is `observed` and carries a file-and-line, command-and-exit-status, or quoted-passage locator (decision 0008); and the invented-rollup check, which refuses a `status.overall` that the recorded criteria do not derive.
- Aggregation (`core/report`): required and optional counts, assessment coverage with its denominator, and the overall status with the rule that produced it written into `status.rule_applied`, so the derivation is auditable (decision 0007).
- Canonical JSON encoding without HTML escaping, the section-tagged evidence digest, and the content-addressed report ID (`core/report`).
- Renderers (`core/render`): Markdown and a self-contained HTML report that reads offline in light and dark themes, with no external scripts, styles, or fonts (decision 0021). Formats are `md` and `html`.
- SARIF 2.1.0 export (`core/export`) against the OASIS errata01 schema, with `UNKNOWN` and `ERROR` carried in `properties` and the git revision carried as `run.properties.veval.vcs_revision_id`.
- The `veval` command line (`cmd/veval`): `version`, `validate`, `aggregate`, `render`, and `export sarif`; exit 0 for success, 1 for a report that fails its rules, 2 for a usage or input error; `-` reads standard input, `-o -` writes standard output, and `--` terminates the flags.
- Continuous integration for Go on macOS, Linux, and Windows (`.github/workflows/go.yml`) and a GoReleaser configuration covering six operating-system and architecture targets (`.goreleaser.yaml`).
- Skill wiring: `skills/evaluate-output/` calls the `veval` core when it is present, with per-host reference files for Claude Code, Codex, Gemini CLI, Antigravity, Hermes, and ollama-backed agents (`skills/evaluate-output/references/hosts/`), and a contract test that fails when a command name in SKILL.md drifts from the CLI.
- Documentation tooling: `tools/docs/gen_sources.py` and `tools/docs/check_links.py` also skip `.superpowers` and `build`, so git-ignored scratch and release build output are never counted in the source register or link-checked.
- Decision record 0023: a fourth locator shape, `note`, for evidence that was neither opened nor run; it never qualifies for `PASS` (`docs/decisions/0023-note-locator-shape.md`).
- Decision record 0024: the evaluation contract is the report schema's `contract` section rather than a schema file of its own, so there is one canonical record and one drift test (`docs/decisions/0024-contract-inside-report-schema.md`).
- HTML report design pass (`core/render`): a masthead with the overall status as a hero card and blocked-by chips, a sticky section navigation, a two-column facts grid, evidence cards with kind-tinted tags, forensic severity badges, zebra tables, a footer naming the report, and print rules; still one file, no script, light and dark, and no sideways scroll at phone width.
- Release process (decision 0025, `docs/decisions/0025-tag-driven-release.md`): a tag `v*` runs `.github/workflows/release.yml`, which tests, extracts this file's section for the tag with `tools/release/release_notes.py`, renders the two example reports with the freshly built binary, and publishes six archives, `checksums.txt`, and the reports through GoReleaser.
- Install instructions in `README.md`: download an archive from the release, verify its checksum, run `veval version`.
- The canonical form is documented (`docs/architecture/report-schema.md`): the encoding, the normalization of nil arrays and objects to empty ones, how `identity.report_id` and `provenance.evidence_digest` are built, and the requirement that a second implementation reproduce the bytes rather than an equivalent document.

### Changed

- `.markdownlint-cli2.yaml` ignores `**/testdata/**`: renderer goldens are generated output, not prose.
- Skill revision `draft-3` (`skills/evaluate-output/SKILL.md`): the `ERROR` rule is stated per criterion; `status.advisory` is named as the one status field the author writes, because acceptance not being requested is an input to the derivation rather than a product of it; `veval version` runs first and its output is recorded in `provenance.tools`; and reactions are appended to a reactions file beside the report, one JSON object per line, until the core gains the command in Stage 4.
- Decision records 0002, 0004 to 0008, 0013, 0020, and 0021 replace "Nothing described here is implemented" with a dated line naming the implementing paths, and 0006, 0008, 0013, and 0022 carry dated amendments where the building settled something differently: SARIF provenance and invocation success, the shipped locator field names, inspection scope as `limitations.not_inspected[]`, and the Python tooling staying until its Go port is scheduled.
- Documentation reconciled with the shipped code: the four locator shapes and the exactly-one-complete-group rule, the SARIF mapping rows (no `environmentVariables`, per-invocation `executionSuccessful`, `ERROR` notifications on a synthesized invocation), schema versioning as exact equality with no compatibility range yet, a confirmed forensic finding as something the core requires to be `FAIL` already, `skill_revision` as the SKILL.md front-matter label, the byte convention for a `content:sha256:` revision (`docs/architecture/report-schema.md`); the CI description and the unsigned, un-notarized release binaries (`docs/architecture/packaging-and-portability.md`); and `internal/` as build metadata used by the command line (`README.md`).

### Fixed

- Decoding is strict in both directions (`core/report/json.go`): a report followed by a second JSON value, or by any text that is not JSON, is rejected rather than read up to the first value.
- `criteria.error_is_operational` is checked per criterion and against that criterion's own evidence -- an `execution` record, or a command locator reporting a non-zero exit status -- with no report-global fallback, so one failed command elsewhere can no longer justify every `ERROR`.
- Two rules close gaps against `schema/report.schema.json`: `criteria.nonempty`, which refuses a contract that states no criterion and a report that reaches no result, and `range.valid`, which holds `routing.supplied[].count` at zero or more and a retrieved precedent's similarity between 0 and 1; `required.nonempty` now also covers `dimensions[].definition_ref`.
- SARIF: each command invocation's `executionSuccessful` is that command's own exit status, and every `ERROR`-criterion notification moves to one synthesized invocation appended last that names no command line; a `content:sha256:` revision whose digest is empty or not 64 lowercase hexadecimal characters yields no `artifacts[]` entry.
- `veval -h` and `veval --help` print the usage on standard output and exit 0; typing no subcommand still prints it on standard error and exits 2.
- A leading UTF-8 byte order mark is stripped from a report read from a file or from standard input, so a report an editor saved with one is no longer reported as malformed JSON.
- The advisory authoring path is covered end to end through the command line: a report whose `status.advisory` is true aggregates to the `ADVISORY` overall status and renders as advisory.

### Not yet implemented

- The plugin manifests, the agent profile, the isolation profile, the adapters, the classifier, the forensic detectors, the precedent bank, and the pilot set are designed but not built.

[Unreleased]: https://github.com/mickeyyaya/v-eval/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mickeyyaya/v-eval/releases/tag/v0.1.0
