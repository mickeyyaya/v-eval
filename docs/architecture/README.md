# Architecture

Status: design documents written 2026-09-14 from the decisions recorded in [../decisions/](../decisions/) and the research in [../research/](../research/). Nothing here is implemented. Where a detail was not decided by the maintainer it is marked **proposal**.

The documents describe one product: an evaluator that takes any supplied history, collects all the information it can, reflects what is hidden underneath the data, and only then derives verdicts ([decision 0013](../decisions/0013-evaluator-not-auditor.md)). The host assistant performs semantic inspection and writes evidence; a small deterministic core validates that evidence, runs adapters and detectors, computes results, and renders reports ([decision 0004](../decisions/0004-start-at-rungs-2-to-4.md), [decision 0005](../decisions/0005-go-core-binary.md)).

| Document | What it specifies |
| --- | --- |
| [pipeline.md](pipeline.md) | The intake-to-report pipeline: intake of any history data, a typed context model, a deterministic classifier that routes to criteria and adapters, the adapter and detector interfaces, verdict computation, and renderers. Includes the trust boundaries and illustrative Go interfaces. |
| [report-schema.md](report-schema.md) | The JSON-first report contract that every renderer consumes: sections, fields, enums, the evidence-shape rule enforced for any PASS, mappings to SARIF 2.1.0 and to evolve-loop's verdict tokens, and the fixed section order of the HTML report. |
| [learning-loop.md](learning-loop.md) | The adaptive layer: human reactions as reward records, a local precedent bank, cold re-judging before admission, a locked anchor set, deterministic splits, the later rule ladder, measurement, and governance of the self-improvement loop. |
| [forensics.md](forensics.md) | The detective pass that runs after results are verified: the three first-release detectors, the evidence they produce, the integrity criteria they feed, false-positive handling, and graded isolation levels. |
| [packaging-and-portability.md](packaging-and-portability.md) | The layered repository layout and growth rungs, Go binary distribution across macOS, Linux, and Windows, per-harness reference files for Claude Code, Codex, Gemini CLI, Antigravity, Hermes, and ollama-backed agents, and the CI matrix. |
| [evaluating-v-eval.md](evaluating-v-eval.md) | How v-eval's own skill and core are regression-tested and measured: portable case format with several runners, the pilot set, the labeling protocol, the anchor set, and metrics with denominators. |

Reading order for a first pass: pipeline, report schema, forensics, learning loop, packaging, evaluating. The [design overview](../design.md) summarizes these; the [requirements](../requirements.md) remain the authority on scope.
