# First release (v0.1.0) and HTML report design pass

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Publish v-eval 0.1.0 on GitHub Releases from a tag, with binaries for six targets, checksums, release notes taken from the changelog, and the two example reports rendered by the released binary; and make the HTML report read like a polished document first.

**Architecture:** No new packages. The HTML renderer's template and stylesheet change; the Go code around it does not. Release tooling is one GitHub Actions workflow, one GoReleaser addition (`release.extra_files`), and one standard-library Python tool with unit tests under `tools/release/`.

**Tech Stack:** Go 1.23 standard library; GoReleaser v2; GitHub Actions; Python 3 standard library for the changelog extractor (decision 0022).

**Spec:** [decision 0025](../../decisions/0025-tag-driven-release.md), [decision 0021](../../decisions/0021-html-report-every-evaluation.md), [packaging design](../../architecture/packaging-and-portability.md), `.superpowers/sdd/2026-09-15-first-release/html-design-brief.md` (the design brief, reproduced in summary in the report-schema design's HTML section after this plan).

## Global Constraints

- Independent project (decision 0001): no dependency on any other maintainer project.
- Renders add no information (decision 0006); the report is evidence-first (decision 0013); the HTML is one self-contained file with no script and no external resource, light and dark (decision 0021).
- Every script is portable across macOS, Linux, and Windows (decision 0020); the release workflow itself runs on ubuntu-latest and cross-compiles.
- Strict TDD; simplifier, code reviewer, and architect on every change; merge to `main` fast-forward after the reviews and green CI; the tag is pushed only from `main`.
- Version: `v0.1.0` (the schema is 0.1.0; this is the first tagged release).

## Task A: HTML report design pass

Files: `core/render/templates/report.html.tmpl`, `core/render/html.go`, `core/render/funcs.go`, `core/render/html_test.go`, `core/render/renderers_test.go`, HTML goldens. The Markdown renderer and goldens are untouched.

- [ ] Tests first: nav anchors equal the rendered section ids in order; the overall badge sits in a hero card before the first section; blocked-by ids render as chips; a print stylesheet hides the nav; the footer names the report id; the criteria heading carries the count.
- [ ] Template and stylesheet: masthead, sticky section nav, two-column facts grid, evidence cards with kind-tinted tags, forensics severity badges, zebra tables, footer, print rules; contrast at least 4.5:1 for every badge and tag in both themes.
- [ ] Goldens regenerated; browser check at 1100px and 400px in both themes by the controller.
- [ ] Commit: `feat(render): report design pass — masthead, section nav, evidence cards, print styles`

## Task B: Release tooling

Files: `.github/workflows/release.yml`, `.goreleaser.yaml`, `.gitignore`, `tools/release/release_notes.py`, `tools/release/test_release_notes.py`, `.github/workflows/docs.yml` (runs the new tests on three operating systems).

- [ ] Tests first for the extractor: finds `## [0.1.0] - 2026-09-15`, prints the body up to the next level-two heading without the trailing link references, exits 2 with a message when the version is absent, ignores fenced code, runs as `python tools/release/release_notes.py CHANGELOG.md 0.1.0`.
- [ ] Workflow: trigger `push: tags: ['v*']`; `permissions: contents: write`; checkout with `fetch-depth: 0`; `setup-go` with `go-version-file: go.mod`; `go test ./...`; extract notes to `build/release-notes.md`; render `build/examples/example-report-<fixture>.html` from both fixtures with `go run ./cmd/veval render --format html`; `goreleaser/goreleaser-action@v6` with `args: release --clean --release-notes build/release-notes.md`.
- [ ] `.goreleaser.yaml`: `release: extra_files: - glob: build/examples/*.html`; `build/` git-ignored; `goreleaser check` passes; a local `goreleaser release --snapshot --clean --skip=publish` proves the extra files and notes are picked up.
- [ ] Commit: `ci: tag-driven release with notes from the changelog and rendered example reports`

## Task C: Documents and the release

Files: `CHANGELOG.md` (`[0.1.0] - 2026-09-15`), `CITATION.cff`, `README.md` (Install), `docs/architecture/packaging-and-portability.md`, `docs/architecture/report-schema.md` (HTML design), `ROADMAP.md`, `docs/decisions/README.md`, `docs/decisions/0025-tag-driven-release.md`.

- [ ] Documents updated; register, links, markdownlint clean.
- [ ] Reviews: simplifier and code reviewer per task, architect once over both; fixes applied.
- [ ] Merge `--ff-only` to `main`, push, `go` and `docs` workflows green on three operating systems.
- [ ] `git tag -a v0.1.0 -m "v-eval 0.1.0"`, push the tag, watch `release.yml`.
- [ ] Verify the published release: assets listed, checksum of a downloaded archive matches, `veval version` on the extracted binary names `0.1.0`, the attached example report opens and matches the golden.

## Verification

1. `gh release view v0.1.0` shows six archives, `checksums.txt`, and two example reports.
2. The darwin_arm64 archive's sha256 matches `checksums.txt`; `veval version` prints `veval 0.1.0 (<sha>) schema 0.1.0`.
3. The attached `example-report-extended.html` renders offline with the new design in light and dark, and its bytes equal the output of the same binary run locally on the fixture.
