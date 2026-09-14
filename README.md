# v-eval

Repository: <https://github.com/mickeyyaya/v-eval>

**[Project requirements](docs/requirements.md)** consolidate the maintainer's requested scope and take priority over preliminary implementation proposals.

Evaluate AI-generated work against the intent, design, sources, and acceptance criteria that define success.

**Status: walking skeleton.** The report schema v0.1.0 (`schema/`), the Go core (`core/`: validation, aggregation, canonical JSON and digests, Markdown and self-contained HTML renderers, SARIF 2.1.0 export), and the `veval` command line (`cmd/veval/`) exist, with continuous integration on macOS, Linux, and Windows. The adapters, the classifier, the forensic detectors, the learning loop, the agent profile, and the plugin manifests are designed but not built, and there is no validated quality model yet. The repository also carries cited research, a consolidated requirements record, twenty-two decision records, an architecture design, a proposed evaluation contract, a draft agent skill, and a worked example.

**[Open the learning handbook](docs/README.md)** for a guided path through generated context, scoring, deterministic tools, evaluation gaming, and practitioner research.

## The question

An AI can produce plausible code or polished writing without solving the intended problem. v-eval asks:

> What was required, what evidence supports each requirement, and what remains unverified?

The proposed first use case is reviewing a code change against a developer's intent, design, and tests. Documents and retrieved context use the same criterion-and-evidence concept, with different checks. This scope is a proposal for discussion.

## Principles

The maintainer set these axes on 2026-09-14; the [decision records](docs/decisions/README.md) trace how each shaped the design.

- **Verify; never trust a summary.** A candidate's description, commit message, or "tests passed" is a claim to check against the underlying context, not evidence. A result of PASS must cite something the evaluator actually opened or ran.
- **Evaluator, not auditor.** The goal is to collect all the information and reflect what is hidden beneath the data. Observations and evidence lead; verdicts are a derived summary.
- **Detective stance toward gaming.** Verifying the supplied result is the first step; the second is looking for traces of tampering, mismatched provenance, and hardcoding. Isolation is recorded, not mandatory.
- **Portable.** The skill must work across agent CLIs, including Antigravity and ollama-backed agents, and every script or program must run on macOS, Linux, and Windows.
- **Learn locally from human feedback, with measured improvement.** Corrections and reactions are the signal; a locked anchor set measures whether the evaluator got more accurate rather than more agreeable.

## What an evaluation should show

| Criterion | Method | Result | Evidence |
| --- | --- | --- | --- |
| Preserve input order | Code inspection | FAIL | The implementation sorts its output, contradicting the brief. |
| Add no dependencies | Code inspection | PASS | The complete supplied function uses only built-ins. |
| Regression suite passes | Test execution | UNKNOWN | No runnable suite or verified run was supplied. |

This is an illustrative report, not an executed test result. See the [worked example](examples/code-review/report.md).

A result applies to its stated criteria and inspected artifact. It is not a universal quality score. Required failures cannot be averaged away; missing evidence remains visible.

## Try the draft skill

Give a tool-capable assistant access to this checkout and a request such as:

```text
Read skills/evaluate-output/SKILL.md and use it to evaluate my change.
Intent: <the requested outcome>
Artifacts: <files or diff to review>
Design: <relevant design documents, if available>
Criteria: <acceptance requirements>
Tests: <existing suite and execution environment, if available>
Write a report with evidence for each criterion and distinguish checks
you executed from code inspection and supplied claims.
```

The skill is a prompt workflow executed by the host assistant, not a sandbox or a standalone program. Its behavior depends on the assistant and available tools. To try the supplied example, use [input.md](examples/code-review/input.md) without first reading the worked report.

## Read and contribute

- [Start here: learn to evaluate generated context and other AI output](docs/learn-evaluation.md)
- [Research index: the 2026-09-14 decision research and the foundational notes](docs/research/README.md)
- [Decision records](docs/decisions/README.md)
- [Architecture](docs/architecture/README.md)
- [Deep research: evaluation diversity and scoring](docs/deep-research.md)
- [Original practitioner posts and social sources](docs/practitioner-reading-list.md)
- [Deterministic evaluation tools and 2026 research](docs/deterministic-tools-2026.md)
- [Research: benchmarks, metrics, and open-source projects](docs/research.md)
- [How v-eval compares, and how we would prove an improvement](docs/comparison.md)
- [Adapt evaluation to a particular perspective](docs/evaluation-views.md)
- [AI gaming, evaluator attacks, and proposed defenses](docs/gaming-and-defenses.md)
- [Proposed design and decisions to discuss](docs/design.md)
- [Evaluation contract template](templates/evaluation-contract.md)
- [Draft evaluation skill](skills/evaluate-output/SKILL.md)
- [Roadmap](ROADMAP.md)
- [Contribution guide](CONTRIBUTING.md)

Existing frameworks already support custom evaluation and rubrics. The proposed contribution is a convenient workflow connecting project requirements to inspectable evidence. We will test whether it adds value over an adapter or recipe for an existing framework before building a new execution engine.

## Repository layout

```text
README.md            this file
LICENSE              MIT
CONTRIBUTING.md      how to contribute cases, research, and proposals
CODE_OF_CONDUCT.md   community expectations
SECURITY.md          how to report a vulnerability
CHANGELOG.md         notable changes by version
ROADMAP.md           proposed sequence, no dates implied
CITATION.cff         how to cite the project
.editorconfig        shared editor settings
.gitignore           ignored paths
.markdownlint-cli2.yaml  Markdown lint configuration used by CI
.goreleaser.yaml     release build matrix for the six binary targets
.github/             issue and pull request templates, workflows
schema/              report JSON Schema v0.1.0, embedded in the core
core/                Go library: report types, validation, aggregation, renderers, SARIF export
cmd/veval/           the veval command line: version, validate, aggregate, render, export
internal/            build metadata shared by the core and the command line
docs/
  README.md          handbook index and reading paths
  requirements.md    consolidated requirements with IDs
  design.md          proposed design
  research/          dated, cited research memos and the foundational notes
  decisions/         one MADR-style record per decision
  architecture/      pipeline, report contract, learning loop, packaging
skills/              the evaluation skill in Agent Skills format, with per-host reference files
tools/docs/          maintainer-only documentation tooling with tests (Python; decision 0022; not shipped)
templates/           evaluation contract template
examples/            worked example input and report
```

## License

[MIT](LICENSE). Referenced papers, datasets, and third-party projects retain their own licenses; they are linked, not bundled.
