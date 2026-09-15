# The project site is one static page, deployed by workflow, with the example reports rendered at deploy time

* Status: accepted
* Deciders: maintainer (mickeyyaya), at release 0.1.1
* Date: 2026-09-15

Implemented 2026-09-15: `site/index.html`, `.github/workflows/pages.yml`, published at <https://mickeyyaya.github.io/v-eval/>.

## Context and Problem Statement

The repository's documents are written for contributors and follow the decision records. Someone meeting v-eval for the first time needs a shorter introduction: what the skill is, why it insists on evidence, what a report looks like, and how to install it. Where does that page live, what may it claim, and how does it show a report without copying test data by hand?

## Decision Drivers

* The page must not claim more than the code does; the same rule as every other document ([decision 0006](0006-json-first-report-contract.md) for renders, the documentation rule for prose).
* The HTML report is the product's face; the page should show real reports, rendered by the current code, not screenshots or hand-edited copies.
* Portability and independence: no third-party scripts, fonts, or analytics; nothing that depends on another project ([decision 0001](0001-independent-of-evolve-loop.md), [decision 0020](0020-portability-constraints.md)).
* Maintenance cost: one file a maintainer can read whole.

## Considered Options

1. One self-contained HTML page under `site/`, deployed to GitHub Pages by a workflow that renders the two example reports beside it with `go run ./cmd/veval`.
2. A static-site generator with a theme and a content directory.
3. No site; the README is the introduction.

## Decision Outcome

Option 1. `site/index.html` is one file: system font stacks, custom properties on `:root` with a dark override, no script, no external resource. Every figure on it quotes the repository's fixtures verbatim, every command is one the binary accepts, and the stats are counts of things in the tree. `pages.yml` runs on pushes to `main` that touch the site, the renderers, the command line, or the schema: it renders `example-report-code-review.html` and `example-report-service-change.html` from the fixtures into `site/examples/` (git-ignored) and deploys the directory with the Pages actions under `pages: write` and `id-token: write` permissions. The page is fact-checked against the repository before it is published, like a document.

### Consequences

* Good: the introduction and the example reports can never drift from the code, because the reports are rendered at deploy time and the page is reviewed as a document.
* Good: the page reads offline and prints; a reader on any operating system sees the same page.
* Bad: no analytics, no search, no comments; a change to the page is a commit and a review like any other.
* Bad: the page's copy repeats claims that live in the documents; when a document changes, the page must be re-read (the fact-check list in the ledger is the checklist).

## Confirmation

`curl -sI https://mickeyyaya.github.io/v-eval/` returns 200; the page contains no `<script`; the two example links resolve; the Fig. 1 texts equal `core/report/testdata/worked-example.json`; the `pages` workflow is green on `main`.

Confirmed on 2026-09-15: run 34915006447 deployed; the index and both examples answer 200; the page carries no script; the C3 text on the page is the fixture's, including "after trimming".

## Pros and Cons of the Options

### One self-contained page deployed by workflow

* Good, because it is reviewable, portable, and honest about what it shows.
* Bad, because long-form documentation stays in the repository rather than on the site.

### A static-site generator

* Good, because many pages and navigation come cheaply.
* Bad, because it adds a toolchain, a theme, and a build the project does not otherwise need, and the example reports would be copied rather than rendered.

### README only

* Good, because there is nothing to maintain.
* Bad, because the README is written for contributors, and a rendered report cannot be shown in it.

## More Information

* GitHub Pages with GitHub Actions: <https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages> (accessed 2026-09-15); the `configure-pages`, `upload-pages-artifact`, and `deploy-pages` actions.
* Related: [decision 0021](0021-html-report-every-evaluation.md) (the report the page shows), [decision 0025](0025-tag-driven-release.md) (the same two reports are attached to every release).
