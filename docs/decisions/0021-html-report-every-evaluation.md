# Every evaluation renders a clean, self-contained HTML report

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented.

## Context and Problem Statement

The maintainer requested: "make sure every evaluation will generate human readable html report with clean and neat structure format." With JSON canonical ([0006](0006-json-first-report-contract.md)), what does the HTML renderer have to be, and what does it contain?

## Decision Drivers

* Every evaluation, not only some, must produce the report; it is a fixed pipeline output.
* It must open offline on any OS with no external assets, because reports travel with the evaluated artifact and CI archives them.
* The evaluator-not-auditor ordering ([0013](0013-evaluator-not-auditor.md)) must be visible in the layout.
* Both `claude plugin eval` and smevals already produce self-contained HTML, so this is the expected form.

## Considered Options

* Self-contained single-file HTML rendered by the Go core from the JSON, fixed section order, light and dark
* HTML rendered from the Markdown render
* Server-rendered or hosted report viewer

## Decision Outcome

Chosen option: "Self-contained single-file HTML rendered by the Go core from the JSON, fixed section order, light and dark". One file, inline CSS, no CDN, no scripts required to read it, theme-aware. Fixed section order: a compact status strip; observations and evidence; the claim-to-verification table; per-criterion table with evidence excerpts and citations; dimension measurements with units; unknowns, not-inspected scope, and limitations; improvement brief; routing rationale; provenance including isolation levels, tool versions, and hashes. Markdown is a sibling render, not the HTML's source.

### Consequences

* Good, because a reader without the CLI can inspect evidence next to every verdict.
* Good, because the same renderer serves the skill, the agent, the plugin, and CI archives.
* Bad, because a template and its accessibility and readability must be maintained in the core.
* Bad, because long claim tables and evidence excerpts make some reports large; the renderer needs collapsible sections that degrade without scripts.

## Confirmation

The core's render command emits `report.html` for every fixture; a test asserts no external URL is referenced for assets; a test asserts the section order; a manual readability check on the pilot reports in light and dark themes. None exists yet.

## Pros and Cons of the Options

### Self-contained HTML from JSON in the core

* Good, because offline, portable, and identical across hosts.
* Bad, because template maintenance in Go.

### HTML from the Markdown render

* Good, because one template less.
* Bad, because Markdown loses structure the schema has, and the chain re-introduces prose as an intermediate.

### Hosted viewer

* Good, because richer interaction.
* Bad, because a service dependency contradicts local-first operation.

## More Information

* Related requirements: REQ-09, REQ-19, and REQ-34 in [requirements](../requirements.md). Informed by [skill packaging research](../research/2026-09-14-skill-packaging.md) and [report schema research](../research/2026-09-14-report-schemas.md).
* Claude Code plugin evals, self-contained HTML report output: <https://code.claude.com/docs/en/plugin-evals> (accessed 2026-09-14).
* prime-radiant-inc/smevals, `build` static HTML reports: <https://github.com/prime-radiant-inc/smevals> (accessed 2026-09-14).
* Hamel Husain, "'It's Hard to Eval' Is a Product Smell", interfaces that expose provenance and intermediate work: <https://hamel.dev/blog/posts/eval-smell/> (2026-06-29; accessed 2026-09-14).
* v-eval [industry scorecard](../industry-scorecard.md), worked table layout for dimension measurements.
* Revisit when pilot readers report the layout hides evidence, or when a consumer needs an interactive viewer.
