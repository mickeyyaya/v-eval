# Practitioner reading list: evaluation diversity and scoring

**30 unique reviewed sources; accessed 2026-09-14.** This selective corpus combines original blogs, surveys, engineering accounts, interviews, and two readable social posts. It establishes neither consensus nor v-eval superiority.

Use [the main report](deep-research.md) for synthesis, [research index](research.md) for papers/frameworks, and [standards reference](standards-reference.md), [industry scorecard](industry-scorecard.md), and [deterministic-tools guide](deterministic-tools-2026.md) for measurements. The [CSV inventory](practitioner-sources.csv) preserves full metadata and access notes.

Surveys synthesize experiments; implementations demonstrate functionality; anecdotes suggest patterns; announcements state aims. None proves effectiveness. Dates identify publication, with updates marked. Eugene Yan's dates use page metadata. Hamel's judge guide is deduplicated; `/evals/index.html` is normalized to `/evals/`.

## Scoring and judge validity

| Source, author, publication | Evidence kind | Takeaway and limit |
| --- | --- | --- |
| R01: [Evaluating the Effectiveness of LLM-Evaluators (aka LLM-as-Judge)](https://eugeneyan.com/writing/llm-evaluators/) — Eugene Yan; 2024-08-18 | Literature survey | Choose direct, pairwise, or reference judgments for the decision. Compare classification errors as well as agreement. Surveyed tasks and models differ; no method is universally best. |
| R02: [AlignEval: Building an App to Make Evals Easy, Fun, and Automated](https://eugeneyan.com/writing/aligneval/) — Eugene Yan; 2024-10-27 | Implementation report; small demonstration | Inspect and label examples before optimizing a judge, then check separate test data. The working tool demonstrates feasibility; its small experiment does not establish general reliability. |
| R03: [Product Evals in Three Simple Steps](https://eugeneyan.com/writing/product-evals/) — Eugene Yan; 2025-11-23 | Practitioner guidance | Use interpretable criteria, dimension-specific judges, and order checks for comparisons. Failure-enriched examples help debugging but do not estimate production prevalence; binary scoring remains a practical default. |
| R04: [Common pitfalls when building generative AI applications](https://huyenchip.com/2025/01/16/ai-engineering-pitfalls.html) — Chip Huyen; 2025-01-16 | Experience and case synthesis | Keep independent human inspection alongside automated judges and revisit alignment after changes. The advice comes from experience; review cadence and sample sizes must fit the application. |
| R05: [Extrinsic Hallucinations in LLMs](https://lilianweng.github.io/posts/2024-07-07-hallucination/) — Lilian Weng; 2024-07-07 | Literature survey | Separate source consistency, external factuality, and uncertainty. Atomic claims make verification inspectable, but extraction and sources can fail; this survey does not measure v-eval accuracy. |
| R06: [Reward Hacking in Reinforcement Learning](https://lilianweng.github.io/posts/2024-11-28-reward-hacking/) — Lilian Weng; 2024-11-28 | Literature survey | Challenge whether score improvements preserve the intended outcome. Protect independent checks against feedback-loop exploitation. Training experiments motivate tests but cannot supply failure rates for static reviews. |
| R07: [Using LLM-as-a-Judge For Evaluation: A Complete Guide](https://hamel.dev/blog/posts/llm-judge/) — Hamel Husain; 2024-10-29; updated 2026-09-01 | Practitioner guide | Start with expert decisions and critiques, then validate a judge and inspect failure segments. A coherent single-expert rubric may still miss other stakeholders or mandatory constraints. |
| R08: [Measuring Human Performance on ARC-AGI-3](https://arcprize.org/blog/arc-agi-3-human-dataset) — Greg Kamradt; 2026-04-14 | Controlled study and scoring analysis | Inspect how normalization and score caps reward actual behavior; human replay data can expose scoring defects. Findings concern this benchmark, not universal document or code quality. |
| R09: [AI Evals: Everything You Need to Know](https://hamel.dev/blog/posts/evals-faq/) — Hamel Husain; Shreya Shankar; 2025-05-28; updated 2026-09-01 | Practitioner synthesis | Turn important observed failures into targeted product evaluations and maintain realistic examples. This living FAQ offers opinionated guidance, not a standard or a universal prescription for every team. |

## Context and agent behavior

| Source, author, publication | Evidence kind | Takeaway and limit |
| --- | --- | --- |
| R10: [Evaluating Long-Context Question & Answer Systems](https://eugeneyan.com/writing/qa-evals/) — Eugene Yan; 2025-06-22 | Literature survey | Evaluate faithfulness, usefulness, citations, and answerability separately across evidence positions and documents. Benchmark findings guide test design; local transfer and claim-extraction quality still require validation. |
| R11: [Agents](https://huyenchip.com/2025/01/07/agents.html) — Chip Huyen; 2025-01-07 | Literature and conceptual synthesis | Distinguish goal achievement, planning, tool failures, constraints, and efficiency. A valid call need not accomplish the task; the evolving taxonomy is guidance rather than an exhaustive specification. |
| R12: [There Are Only 6 RAG Evals](https://jxnl.co/writing/2025/05/19/there-are-only-6-rag-evals/) — Jason Liu; 2025-05-19 | Conceptual framework | Diagnose question, context, and answer relationships instead of one quality score. The six-part taxonomy is a useful checklist, not a completeness theorem covering freshness, permissions, and cost. |
| R13: [Agent observability powers agent evaluation](https://www.langchain.com/blog/agent-observability-powers-agent-evaluation) — Harrison Chase; 2026-01-27 | Architecture advice and anecdotes | Retain calls, executions, conversations, and state changes to trace delayed memory failures. Recorded behavior supports diagnosis; it does not reveal a model's true internal reasoning or prove causality. |
| R14: [How we build evals for Deep Agents](https://www.langchain.com/blog/how-we-build-evals-for-deep-agents) — Vivek Trivedy; Mason Daugherty; Eugene Yurtsev; Harrison Chase; 2026-03-26 | Firsthand implementation account | Group cases by relevant capabilities, distinguish plumbing from model behavior, and inspect correctness before efficiency. The described process is useful prior art, not a controlled comparison of frameworks. |
| R15: [Pressure Testing GPT-4 & Claude 2.1 Long Context](https://mail.gregkamradt.com/posts/pressure-testing-gpt-4-claude-2-1-long-context) — Greg Kamradt; Date unverified | Small controlled retrieval probe | Vary context length and evidence position to expose retrieval failures. Historical needle retrieval does not establish current model thresholds, multi-document synthesis, or useful summaries; absolute publication date remains unverified. |
| R16: [Improve Agentic Performance with Evals and Error Analysis, Part 2](https://www.deeplearning.ai/the-batch/improve-agentic-performance-with-evals-and-error-analysis-part-2/) — Andrew Ng; 2025-10-22 | Practitioner guidance and worked example | Trace weak reports through search and synthesis, while questioning the workflow decomposition itself. Improving isolated stages may not improve the whole; the example supplies guidance, not comparative evidence. |
| R17: [Demystifying evals for AI agents](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents) — Mikaela Grace; Jeremy Hadfield; Rodrigo Olivares; Jiri De Jonghe; 2026-01-09 | Firsthand practice guide | Write checks for known success criteria and expand coverage as agent failures emerge. Match grading methods to the task; the engineering guidance does not certify a particular external evaluator. |
| R18: [Eval awareness in Claude Opus 4.6's BrowseComp performance](https://www.anthropic.com/engineering/eval-awareness-browsecomp) — Russell Coleman; 2026-03-06 | Observed benchmark trajectories | Record allowed search methods and inspect answer leakage or benchmark identification. The investigated unrestricted behavior was not labeled an alignment failure; findings apply to the reported configuration. |

## Product evaluation and implementation

| Source, author, publication | Evidence kind | Takeaway and limit |
| --- | --- | --- |
| R19: [Hard Truths From the AI Trenches](https://jxnl.co/writing/2025/03/05/ai-from-the-trenches/) — Jason Liu; 2025-03-05 | Consulting anecdotes and opinion | Measure the information users actually need, then connect technical errors to consequences. Consulting anecdotes motivate prioritization; their outcomes do not establish causal effects or general performance guarantees. |
| R20: [Your AI Product Needs Evals](https://hamel.dev/blog/posts/evals/) — Hamel Husain; 2024-03-29 | Firsthand case study | Combine scoped tests, trace review, human/model evaluation, and online experiments at appropriate cadences. Offline success supports iteration but does not establish better user behavior by itself. |
| R21: [We Iterate on Models. We Can Iterate on Evals, Too](https://www.deeplearning.ai/the-batch/we-iterate-on-models-we-can-iterate-on-evals-too) — Andrew Ng; 2025-04-16 | Practitioner recommendation | Begin with a small useful evaluation and improve it through human comparisons and iteration. Starter examples support learning; they are not sufficient evidence of broad reliability. |
| R22: [A Field Guide to Rapidly Improving AI Products](https://hamel.dev/blog/posts/field-guide/) — Hamel Husain; 2025-03-24 | Practitioner process guide | Organize improvement around real failures, trusted evaluation, and repeatable experiments. The proposed process transfers ideas across teams, but does not prove universal efficacy or prescribe a fixed roadmap. |
| R23: [“It's Hard to Eval” Is a Product Smell](https://hamel.dev/blog/posts/eval-smell/) — Hamel Husain; 2026-06-29 | Consulting examples and design sketches | Expose assumptions, provenance, intermediate work, and unresolved items so users can verify results. The sketches propose useful interfaces; their effect on review time has not been demonstrated here. |
| R24: [smevals — a small eval suite for evaluating models, prompts, and harnesses](https://simonwillison.net/2026/Jul/31/smevals/) — Simon Willison; 2026-07-31 | Implementation description and demo | Small file-based suites, separate execution and grading, and browsable reports reduce workflow complexity. An implemented harness establishes functionality; its graders still need task-specific validation. |
| R25: [Evals Skills for Coding Agents](https://hamel.dev/blog/posts/evals-skills/) — Hamel Husain; Shreya Shankar; 2026-03-02; updated 2026-08-31 | Implemented skill workflow description | Existing skills cover error discovery, judge construction, validation, and RAG evaluation. Treat these as direct packaging prior art; a skill collection alone does not prove improved decisions. |
| R26: [In Defense of AI Evals, for Everyone](https://www.sh-reya.com/blog/in-defense-ai-evals/) — Shreya Shankar; 2025-09-05 | Practitioner argument | Adjust evaluation rigor to expertise, familiarity, and the decision at hand. This is the author's argument; it cannot substitute for inaccessible social posts she discusses or establish consensus. |

## Interviews and original social posts

| Source, author, publication | Evidence kind | Takeaway and limit |
| --- | --- | --- |
| R27: [Multi-Turn RL for Multi-Hour Agents — with Will Brown, Prime Intellect](https://www.latent.space/p/willccbb) — Will Brown; swyx; 2025-05-23 | Speaker-attributed opinion | Consider evaluator incentives, independent research, and the value of new tasks. The discussion raises governance questions; it is not evidence that any named organization manipulated benchmark rankings. |
| R28: [Reality: The Final Eval — Lukas Petersson and Axel Backlund of Andon Labs](https://www.latent.space/p/andon) — Lukas Petersson; Axel Backlund; swyx; Vibhu; 2026-06-04 | Benchmark creators' accounts | Long-running economic environments expose different behaviors from short question-answer tasks. Creators' interpretations are distinct from benchmark data; profit alone cannot certify truthful output or compliance with every constraint. |
| R29: [Request for a worked example of iterating prompts with simple evals [descriptive title]](https://bsky.app/profile/simonwillison.net/post/3lbzccosg5c23) — Simon Willison; 2024-11-28 | Practitioner request/opinion | Show successive outputs, handwritten checks, and prompt revisions in tutorials. This original request identifies a learning need; it provides no measured evidence that a particular evaluation approach works. |
| R30: [Launching Every Eval Ever: Toward a Common Language for AI Eval Reporting](https://bsky.app/profile/eval-eval.bsky.social/post/3mf2techg422c) — EvalEval Coalition; 2026-02-17 | Initiative announcement | Investigate shared reporting vocabulary and reusable evaluation records before inventing another format. The announcement establishes the initiative's stated aim, not completeness, broad adoption, or a validated industry standard. |

## Access limits and unverified leads

Greg Kamradt's long-context newsletter was read, but its absolute publication date remains **unverified**. References to 2023 models do not establish publication year. R15 has a blank CSV date; historical results cannot establish current model limits.

The Andrew Ng letters were read through indexed first-party text; direct fetching returned HTTP 403. Willison's Bluesky text was read through [the public CDN](https://web-cdn.bsky.app/profile/simonwillison.net/post/3lbzccosg5c23). The CSV retains these distinctions and both social posts' exact UTC timestamps.

Four X originals returned HTTP 403 and were **NOT READ**. Their wording and dates remain unverified. They are leads only, excluded from the supporting-source CSV; secondhand descriptions cannot establish their authors' positions.

| Unread lead | Status |
| --- | --- |
| [Andrej Karpathy](https://x.com/karpathy/status/1795873666481402010) | NOT READ |
| [Ben Hylak](https://x.com/benhylak/status/1962949215195431372) | NOT READ |
| [swyx](https://x.com/swyx/status/1963725773355057249) | NOT READ |
| [Jason Liu](https://x.com/jxnlco/status/1963295889121837466) | NOT READ |

## Turn recommendations into experiments

Compare v-eval with competent skills and harnesses on identical tasks. Preserve scoring semantics; measure wrong decisions, incomplete judgments, and review effort. Practitioner advice motivates experiments, not conclusions.
