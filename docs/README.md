# The v-eval handbook

Start with the [consolidated requirements](requirements.md) for the product's intended scope, including mandatory skill policies and measured recursive self-improvement. The research below supports those requirements; it does not replace them.

Learn how to decide whether AI-generated content, context, and code actually satisfy their intended purpose. This handbook connects research and practitioner experience to a proposed evaluation workflow. The repository is a research and skill prototype; an automated evaluation program is not implemented yet.

## Choose a reading path

| Your question | Start here | Then read |
| --- | --- | --- |
| How do I judge an AI answer or generated context? | [Learn evaluation](learn-evaluation.md) | [Evaluation perspectives](evaluation-views.md) |
| What do experienced engineers disagree about? | [Evaluation diversity and scoring](deep-research.md) | [Practitioner reading list](practitioner-reading-list.md) |
| Can software check an output without an LLM judge? | [Deterministic tools in 2026](deterministic-tools-2026.md) | [Evaluation contract](../templates/evaluation-contract.md) |
| How can AI game the evaluator? | [Gaming and defenses](gaming-and-defenses.md) | [Comparison experiments](comparison.md) |
| What are the established metrics and tools? | [Research overview](research.md) | Detailed notes for [content/context](research/content-context.md), [code](research/code.md), and [frameworks](research/frameworks.md) |
| Why build v-eval when tools already exist? | [Comparison and evidence requirements](comparison.md) | [Proposed design](design.md) and [roadmap](../ROADMAP.md) |
| What did the 2026 research conclude, and what was decided? | [Research synthesis](research/2026-09-14-synthesis.md) | [Research index](research/README.md) and [decision records](decisions/README.md) |
| How will v-eval be built? | [Architecture](architecture/README.md) | [Decision records](decisions/README.md) and [proposed design](design.md) |
| Can the skill learn from my corrections without drifting? | [Adaptive skills](research/2026-09-14-adaptive-skills.md) | [Evaluator learning](research/2026-09-14-evaluator-learning.md) |
| Which language and packaging work on every agent CLI and operating system? | [Language and portability](research/2026-09-14-language-and-portability.md) | [Skill packaging](research/2026-09-14-skill-packaging.md) |

For a first sitting, read the learning guide and try its fictional source-checking example. For deeper study, read the scoring report, choose a few original practitioner posts, and complete the exercises. For implementation planning, use the tools guide and comparison protocol to decide what can be reused.

## What was decided on 2026-09-14

After the research above was complete, the maintainer answered the open design questions one at a time. The answers are recorded as twenty-two decision records, each with the options considered, the decision, its consequences, and links to the evidence. They fix the relationship to evolve-loop, the scoring policy, the first slice, the packaging rungs, the implementation language, the report contract, the verdict vocabulary, the evidence rule, the intake classifier, the first adapters, isolation, gaming-trace detectors, the evaluator-not-auditor stance, local learning, pilot cases and labeling, the absence of a separate judge in v1, self-improvement governance, skill regression tests, delivery and license, portability, the HTML report, and the exemption of repository maintenance tooling from the product's no-interpreter rule. Start at [the decisions index](decisions/README.md); the [research synthesis](research/2026-09-14-synthesis.md) maps each finding to the record it informed. The [architecture](architecture/README.md) documents describe the resulting pipeline. None of this is implemented yet.

## Practice on a real artifact

Write its intended job and required outcomes using the contract template. Select only relevant perspectives. Collect the strongest available evidence and record each criterion as PASS, FAIL, UNKNOWN, ERROR, or justified NOT_APPLICABLE. A missing observation must remain visible.

Try the [draft skill](../skills/evaluate-output/SKILL.md) on [the code example input](../examples/code-review/input.md), then compare your reasoning with [the worked report](../examples/code-review/report.md). That example is an inspection exercise, not a benchmark of evaluator quality.

## Read claims at the right strength

A paper reports an experiment under specified conditions. An engineer's essay reports experience, interpretation, or opinion. A tool's documentation establishes an available interface. A synthetic example illustrates a mechanism. None of those alone proves that v-eval improves decisions.

The handbook distinguishes these evidence types, preserves publication/access dates where available, and links to originals. It includes historical foundations and sources available through 2026-09-14; it is a curated collection, not every relevant publication. Inaccessible social posts are recorded as discovery gaps rather than verified quotations.
