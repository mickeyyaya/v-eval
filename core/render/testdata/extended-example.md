# v-eval report: service_change git:9f2c1a4e5b6d7c8f9a0b1c2d3e4f5a6b7c8d9e0f

- Report id: sha256:8133cd5af3691431144e38ab5e39c701833b2a53e6a08040139195d374b55465
- Created: 2026-09-14T00:03:00Z by fixture on any/any
- Schema 0.1.0, v-eval 0.0.0-dev, skill revision draft-2
- Contract: extended-example version 1 (provisional)
- Artifact paths: examples/extended/service.go, examples/extended/service_test.go
- Requested outcome: Enforce the vendor's published per-key rate limit in the request path, keep the regression suite passing, and ship a reproducible container image.

## Status

Overall: INCOMPLETE

- Rule applied: unresolved contract conflict
- Blocked by: none
- Contract conflict (unresolved) between E1, E4: The suite is run against the worktree toolchain while the image is built from a pinned one; the brief does not say which toolchain the release is accepted on.

## Observations

- **O1** (assistant) The request path enforces the limit per API key, matching the published documentation. Criteria: E2.
  - [kind inspection, isolation none] examples/extended/service.go:48-52 — const rateLimitPerMinute = 100 // per API key

- **O2** (adapter) The evaluation host has no container toolchain, so the reproducibility and throughput checks could not run. Criteria: E3, E4.
  - [kind execution, isolation none] docker --version (exit 127) — zsh: command not found: docker

## Claims

| ID | Claim | Status | Location | Verification | Notes |
| --- | --- | --- | --- | --- | --- |
| CL1 | The regression suite passes on this revision. | verified | pull request description, paragraph 1 | [kind execution, isolation worktree] go test ./... (exit 0) — ok  github.com/example/service  1.204s; 42 tests, 0 failures | Re-run by the evaluator in a worktree checked out at this revision. |
| CL2 | Throughput improved to 610 requests per second. | unverified | pull request description, paragraph 2 | none | No workload, tool, or log accompanies the number. |

## Criteria

| ID | Result | Method | Evidence and reasoning | Next action |
| --- | --- | --- | --- | --- |
| E1 | PASS | execution | [kind execution, isolation worktree] go test ./... (exit 0) — ok  github.com/example/service  1.204s; 42 tests, 0 failures<br>The suite ran in a worktree checked out at this revision and exited zero with no failures reported. The conclusion covers the suite as supplied, not untested paths. | None. |
| E2 | PASS | source_verification | [kind inspection, isolation none] Clients may issue at most 100 requests per minute per API key. from https://api.example.test/docs/rate-limits (2026-08-30) — The vendor states the limit as 100 requests per minute per API key.<br>[kind inspection, isolation none] examples/extended/service.go:48-52 — const rateLimitPerMinute = 100 // per API key<br>The enforced constant is 100 per minute per key, the same limit and the same unit the cited documentation states at the version read. | None. |
| E3 | UNKNOWN | deterministic_check | [kind supplied, isolation none] Benchmark number quoted in the pull request description, paragraph 2 — Quotes 610 requests per second; no workload, tool, or log accompanies the number. (supplied)<br>No benchmark ran on this host, so the throughput dimension carries no measured value and the target is neither met nor missed.<br>Shared cause with E4.<br>Dimensions: throughput. | Run the stated workload and record throughput with the tool and version that produced it. |
| E4 | ERROR | execution | [kind execution, isolation none] docker build --no-cache . (exit 127) — zsh: command not found: docker<br>The build never started: the container toolchain is absent from the evaluation host. Reproducibility is untested here, not refuted.<br>Shared cause with E3. | Re-run the evaluation on a host that has the container toolchain installed. |
| E5 | NOT_APPLICABLE | static_inspection | [kind inspection, isolation none] examples/extended/vendor.txt:1-1 — # no vendored third-party sources at this revision<br>This revision vendors no third-party sources, so the header requirement has nothing to apply to. | None while the revision vendors nothing. |
| E6 | UNKNOWN | rubric_judgment | [kind inspection, isolation none] examples/extended/service.go:1-1 — Modification time 2026-09-14T00:04:12Z, after the evaluation started at 2026-09-14T00:03:00Z.<br>[kind judgment, isolation none] Integrity rubric applied to the modification-time drift above — The rubric cannot separate a benign checkout touch from a content change without the pre-evaluation tree digest.<br>A tracked file's modification time moved during the evaluation, and no pre-evaluation tree digest was recorded, so whether the artifact changed cannot be settled either way. | Record a tree digest before evaluation begins, then re-run and compare. |

## Forensics

- **F1** mtime-drift 0.3 on E6: severity suspicious, disposition open.
  - Benign alternative: A checkout or a formatter touched the file without changing its contents.
  - [kind inspection, isolation none] examples/extended/service.go:1-1 — Modification time 2026-09-14T00:04:12Z, after the evaluation started at 2026-09-14T00:03:00Z.

## Dimensions

- **throughput** requests_per_second version 1 (established_measure), defined at examples/extended/brief.md#throughput
  - Value: not measured; threshold at or above 500 from examples/extended/brief.md#throughput
  - Measured by bench unknown on workload steady 60s at 8 concurrent workers, direction higher_is_better, range 0 and above
  - Not measured on this host: the quoted number has no workload or tool behind it, so no comparison to the target is possible.
  - [kind supplied, isolation none] Benchmark number quoted in the pull request description, paragraph 2 — Quotes 610 requests per second; no workload, tool, or log accompanies the number. (supplied)

## Counts

- Required: 4 applicable — 2 PASS, 0 FAIL, 1 UNKNOWN, 1 ERROR, 1 not applicable
- Optional: 1 applicable — 0 PASS, 0 FAIL, 1 UNKNOWN, 0 ERROR, 0 not applicable
- Coverage: 2/5 applicable criteria assessed. Coverage is not a correctness score.

## Improvement

- **Reproducibility of the container image is untested: the toolchain is absent from the evaluation host.**
  - Locations: examples/extended/Dockerfile
  - Change: Run the evaluation on a host with the container toolchain, or pin a builder image the evaluator can fetch itself.
  - Preserve: Do not change the published rate limit.
  - Verify by: Two builds of this revision print the same image digest.
  - Criteria: E4

- **Throughput is quoted in prose but never measured.**
  - Locations: pull request description, paragraph 2
  - Change: Run the stated workload and record the measurement with the tool and version that produced it.
  - Preserve: none
  - Verify by: The throughput dimension carries a value, a unit, and a named tool.
  - Criteria: E3

- **The evaluation cannot tell a benign touch from a modification of the artifact.**
  - Locations: examples/extended/service.go
  - Change: Record a tree digest and file timestamps before evaluation begins.
  - Preserve: Do not write into the artifact tree during evaluation.
  - Verify by: The integrity detector compares pre- and post-evaluation tree digests.
  - Criteria: E6

## Limitations

- Not inspected: deployment manifests, which this contract does not cover
- Not executed: container image build, benchmark workload
- Unknown metadata: pre-evaluation file timestamps, the toolchain the release is accepted on
- Assumptions: The vendor documentation read at version 2026-08-30 is the version the change targets.

## Routing

Service change with a supplied test suite and a cited external limit, so execution and source verification both ran; the container and benchmark adapters had no toolchain to run on.

- Supplied: brief x1 from user, artifact x2 from candidate, candidate_claim x1 from candidate
- Profiles: requirements-and-tests 0.1, release-readiness 0.2
- Adapters run: go-test 0.1 over examples/extended/service_test.go; mtime-drift 0.3 over examples/extended/service.go
- Adapters skipped: container-build, missing container toolchain on the evaluation host
- Ambiguity:
  - Does the rate limit apply per API key or per client address? Resolved: Read as per API key: the brief and the vendor documentation both state the limit per key.

## Provenance

- Tools: veval 0.0.0-dev, go 1.23.1, mtime-drift 0.3
- Environment: any/any
  - runtime go 1.23.1
- Isolation levels used: none, worktree
- Evidence digest: sha256:14f86f9069d39c8b3be2115a6fb9cbd61b063ee7fa4e9aaa218c111786620d9d
- Commands:
  - go test ./... in /tmp/veval-worktree, exit 0, isolation worktree, 2026-09-14T00:03:10Z to 2026-09-14T00:03:12Z, log logs/e1-go-test.log
  - docker --version in /tmp/veval-worktree, exit 127, isolation none, 2026-09-14T00:06:00Z to 2026-09-14T00:06:00Z, log logs/o2-docker-version.log
  - docker build --no-cache . in /tmp/veval-worktree, exit 127, isolation none, 2026-09-14T00:06:30Z to 2026-09-14T00:06:30Z, log logs/e4-docker-build.log

## Learning

- Precedents retrieved:
  - sha256:4b2f8d1c6e0a9735bd4c1e8f206a3b5c7d9e0f1a2b3c4d5e6f708192a3b4c5d6 for E6, similarity 0.82
- Reward records created: rr-2026-09-14-e6-integrity-unknown
