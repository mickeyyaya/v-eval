# Delivery: public GitHub repository, MIT license, documents-first initial commit

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

## Context and Problem Statement

REQ-26 requires a GitHub project. At the start of the session the working tree had no commits, no remote, and no repository; MIT sat in the scaffold as a working choice, not a decision. Should the repository be created now, public or private, under which license, and what goes into the first commit?

## Decision Drivers

* REQ-01 and REQ-26: open source, on GitHub.
* The maintainer's instruction that every piece of research and design be written as detailed documents before declaring work done.
* License compatibility with the tools v-eval will adapt to or export for (promptfoo, Inspect, smevals, structured-evaluation are MIT; DeepEval, Ragas, evolve-loop are Apache-2.0).
* Pilot and anchor cases must never be public.

## Considered Options

Repository:

* Create public `mickeyyaya/v-eval` now; first commit after the documents are written
* Create private first, flip to public at first release
* Commit the current docs as-is now, reconcile later
* Do not create it yet

License:

* MIT
* Apache-2.0
* Dual: MIT for skill and docs, Apache-2.0 for the Go core

## Decision Outcome

Chosen options: "Create public `mickeyyaya/v-eval` now; first commit after the documents are written" and "MIT". The repository was created on 2026-09-14 at <https://github.com/mickeyyaya/v-eval> with `origin` set and nothing pushed. The first commit contains the reconciled docs, research reports, decision records, and architecture documents; existing docs are reconciled in place rather than left contradictory. Pilot and anchor cases live outside the public tree.

### Consequences

* Good, because the documents land in a real history that can be reviewed and contributed to.
* Good, because MIT matches the scaffold and the majority of adjacent eval tooling, so adapters and exports mix without notices.
* Bad, because MIT carries no explicit patent grant; if the Go core attracts corporate contributors this may need revisiting.
* Bad, because a public repository from the first commit exposes a design-stage project; the README must keep stating that nothing is implemented.

## Confirmation

`git remote -v` shows `origin` at the URL above; `LICENSE` is the MIT text; the first commit contains `docs/research/`, `docs/decisions/`, and `docs/architecture/`; `.gitignore` excludes `anchors/` and any local case store. Partially done: repository and remote exist; the first commit does not yet.

## Pros and Cons of the Options

### Public now, documents-first commit

* Good, because satisfies REQ-26 immediately and matches the documents-first instruction.
* Bad, because the design stage is visible.

### Private first

* Good, because iteration without an audience.
* Bad, because delays the open-source signal.

### Commit current docs as-is

* Good, because preserves the starting point.
* Bad, because the first commit would contradict several decisions taken the same day.

### Do not create yet

* Good, because nothing shared until the spec is reviewed.
* Bad, because the requirement stays open.

### MIT

* Good, because shortest text, most familiar to skill authors, matches adjacent MIT tools.
* Bad, because no explicit patent grant.

### Apache-2.0

* Good, because patent grant and contribution terms; matches evolve-loop.
* Bad, because heavier for a skill-first project.

### Dual license

* Good, because each part gets the fitting terms.
* Bad, because two license files and per-directory notices.

## More Information

* Related requirements: REQ-01, REQ-26, REQ-27. Informed by the [research synthesis](../research/2026-09-14-synthesis.md) and [frameworks research](../research/frameworks.md).
* Repository: <https://github.com/mickeyyaya/v-eval> (created 2026-09-14).
* License files observed 2026-09-14: promptfoo MIT <https://github.com/promptfoo/promptfoo/blob/main/LICENSE> ; Inspect MIT <https://github.com/UKGovernmentBEIS/inspect_ai/blob/main/LICENSE> ; DeepEval Apache-2.0 <https://github.com/confident-ai/deepeval/blob/main/LICENSE.md> ; structured-evaluation MIT <https://github.com/plexusone/structured-evaluation/blob/main/LICENSE> ; evolve-loop Apache-2.0 per its plugin manifest, read locally.
* Revisit when corporate contributions to the Go core make a patent grant material, or when the first release changes the visibility needs.
