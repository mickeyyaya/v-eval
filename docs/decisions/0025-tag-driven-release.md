# Releases are cut by tag, built by GoReleaser in CI, and ship the rendered example reports

* Status: accepted
* Deciders: maintainer (mickeyyaya), at the first release
* Date: 2026-09-15

Implemented 2026-09-15: `.github/workflows/release.yml`, `.goreleaser.yaml`, `tools/release/release_notes.py`.

## Context and Problem Statement

[Decision 0005](0005-go-core-binary.md) chose a single static Go binary and [decision 0020](0020-portability-constraints.md) requires it to run on macOS, Linux, and Windows. The walking skeleton shipped a GoReleaser configuration for six targets, but nothing produced a release: no workflow ran on a tag, no release notes existed outside `CHANGELOG.md`, and a user who read about the HTML report had nothing to open. How is an official release produced, and what does it contain beyond binaries?

## Decision Drivers

* Every release must be reproducible from a tag by CI, never from a maintainer's machine, so the published checksums describe a build anyone can rerun.
* The release must be verifiable the way v-eval asks of everything else: the notes name what changed, the checksums bind the archives and the example reports, and the reports are rendered from the tagged commit rather than copied from test data.
* Release tooling follows [decision 0022](0022-repository-maintenance-tooling.md): standard-library Python under `tools/`, unit-tested, runnable on all three operating systems.
* The auto review and release loop (Roadmap Stage 4) will later gate releases on v-eval's own evaluation of the candidate against the previous version; the first release must not build anything that loop would have to undo.

## Considered Options

1. A tag-driven GitHub Actions workflow running GoReleaser, with release notes extracted from `CHANGELOG.md` and the two example reports rendered by the freshly built binary and attached as assets.
2. Manual releases: the maintainer runs GoReleaser locally with a token.
3. Continuous releases on every push to `main` with generated version numbers.

## Decision Outcome

Option 1. A push of a tag matching `v*` runs `release.yml` on ubuntu-latest: it runs the test suite, extracts the tagged version's section from `CHANGELOG.md` with `tools/release/release_notes.py`, renders `build/examples/example-report-code-review.html` and `build/examples/example-report-service-change.html` from the two fixtures with `go run ./cmd/veval` at the tagged commit, and runs `goreleaser release --clean` with the extracted notes. GoReleaser builds the six archives, `checksums.txt` (which also lists the two example reports through `checksum.extra_files`), and attaches the reports through `release.extra_files`. The version and short commit are stamped by ldflags, so `veval version` on a released binary names the tag.

### Consequences

* Good: a release is one command for the maintainer (`git tag -a vX.Y.Z && git push origin vX.Y.Z`) and every artifact is produced by CI from the tagged commit.
* Good: the HTML report is visible from the release page without installing anything.
* Good: `CHANGELOG.md` stays the single source of release notes; the extractor fails the workflow when the tagged version has no section, so an undocumented release cannot ship.
* Bad: binaries are unsigned and un-notarized; macOS Gatekeeper blocks a browser download until the quarantine attribute is removed ([packaging design](../architecture/packaging-and-portability.md#binary-distribution)). Signing is a later decision.
* Bad: the release workflow runs on one operating system. Cross-compilation covers the six targets, but a Windows-only or macOS-only runtime regression is caught by `go.yml` on the merge commit, not by the release job.

## Confirmation

`gh release view vX.Y.Z` lists six archives, `checksums.txt`, and the two example reports; `sha256sum -c --ignore-missing checksums.txt` passes against a downloaded archive; the binary inside prints `veval X.Y.Z (<short commit>) schema <schema version>`; `python tools/release/release_notes.py CHANGELOG.md X.Y.Z` prints the notes the release page shows. The confirmation for each release is recorded here after its workflow has run.

Confirmed for v0.1.0 on 2026-09-15: workflow run 34910783993 published six archives, `checksums.txt` (eight entries), and the two example reports; `shasum -a 256 --check --ignore-missing checksums.txt` passed for `veval_darwin_arm64.tar.gz` and `example-report-service-change.html`; the extracted binary printed `veval 0.1.0 (c591242) schema 0.1.0`; the attached example report was byte-identical to that binary's own render of the fixture; the release body was the `[0.1.0]` section of `CHANGELOG.md`.

Confirmed for v0.1.1 on 2026-09-15: workflow run 34915090805 published the same nine assets; checksums passed; the extracted binary printed `veval 0.1.1 (23e5cab) schema 0.1.0`; the attached service-change report was byte-identical both to that binary's own render and to the copy the project site serves.

## Pros and Cons of the Options

### Tag-driven GoReleaser in CI with notes and example reports

* Good, because the build is reproducible from the tag and the checksums describe CI's build.
* Good, because the notes and the examples are derived from the repository at the tag, not typed into a form.
* Bad, because a tag pushed by mistake publishes a release; the workflow's test step and the changelog extractor are the only guards until the Stage 4 loop gates it.

### Manual local releases

* Good, because nothing new is needed.
* Bad, because the checksums describe one machine's build, the token lives on that machine, and the release cannot be rerun by anyone else.

### Continuous releases on every push

* Good, because users always have the latest build.
* Bad, because the version has no meaning, `CHANGELOG.md` cannot describe a release before it exists, and the Stage 4 loop needs a candidate to evaluate before it becomes a release.

## More Information

* GoReleaser: quick start and GitHub Actions (<https://goreleaser.com/quick-start/>, <https://goreleaser.com/ci/actions/>, accessed 2026-09-15); release customization and `extra_files` (<https://goreleaser.com/customization/release/>, accessed 2026-09-15); checksums (<https://goreleaser.com/customization/checksum>).
* GitHub Actions permissions for `GITHUB_TOKEN` (<https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication>, accessed 2026-09-15).
* Keep a Changelog 1.1.0 (<https://keepachangelog.com/en/1.1.0/>): the section format the extractor reads.
* Related: [decision 0005](0005-go-core-binary.md), [decision 0020](0020-portability-constraints.md), [decision 0022](0022-repository-maintenance-tooling.md), the [packaging design](../architecture/packaging-and-portability.md), and the auto review and release loop in [ROADMAP.md](../../ROADMAP.md) Stage 4.
