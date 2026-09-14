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

### Not yet implemented

- The Go core, the plugin manifests, the agent profile, the adapters, the HTML renderer, the precedent bank, and the pilot set are designed but not built.
