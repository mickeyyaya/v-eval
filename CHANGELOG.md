# Changelog

All notable changes to this project are documented in this file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project will follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once a first release is tagged.

## [Unreleased]

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
- Documentation tooling: `tools/docs/gen_sources.py` and `tools/docs/check_links.py` also skip `.superpowers`, so git-ignored scratch is never counted in the source register or link-checked.

### Changed

- `.markdownlint-cli2.yaml` ignores `**/testdata/**`: renderer goldens are generated output, not prose.

### Not yet implemented

- The plugin manifests, the agent profile, the isolation profile, the adapters, the classifier, the forensic detectors, the precedent bank, and the pilot set are designed but not built.
