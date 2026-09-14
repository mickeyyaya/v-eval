# Research: how to evaluate AI-generated work

Reviewed 2026-09-14 using original papers, official project documentation, and repository license files. This is a selected landscape review, not an exhaustive survey or a performance benchmark. Recommendations are our design inferences; publication results retain their original task and experimental scope.

## Start with the question being measured

“Is this good?” combines several questions that need different evidence:

| Question | Appropriate evidence | Common mistake |
| --- | --- | --- |
| Did it follow the instruction? | Explicit requirements, deterministic checks, anchored semantic rubric | Rewarding polished writing that misses the task |
| Are its claims supported? | Claim-to-source passages and support judgments | Assuming that a citation's existence proves support |
| Is the information correct and current? | Suitable authoritative sources, versions, dates, domain review | Equating faithfulness to supplied context with truth |
| Is it complete enough to use? | Required information and intended user task | Measuring only factual precision; omissions can still make it useless |
| Is the retrieved context useful? | Task relevance, labeled retrieval coverage and ranking | Treating a good final answer as proof that retrieval was good |
| Does the code behave correctly? | Assertions, execution, boundary cases, regression checks | Equating a test pass with satisfying an unspecified requirement |
| Can the evaluator be trusted? | Independent labels, adversarial cases, error and abstention rates | Treating the judge's own confidence as measured reliability |

The distinction between instruction compliance and presentation is supported by [LLMBar](https://arxiv.org/abs/2310.07641). Separating source faithfulness from relevance follows [RAGAs](https://aclanthology.org/2024.eacl-demo.16/). These sources motivate separate checks; they do not establish v-eval's effectiveness.

## Learning path

1. Read [Learn evaluation](learn-evaluation.md) for definitions, a complete context example, and small metric calculations.
2. Read the metric and paper tables below, then follow the detailed notes for the artifact you care about.
3. Try the [contract template](../templates/evaluation-contract.md) on a real task and identify what evidence is missing.
4. Read [the comparison and experiment plan](comparison.md) to understand why a feature claim differs from measured superiority.
5. Inspect [the proposed design](design.md) and [draft skill](../skills/evaluate-output/SKILL.md). Their procedures are hypotheses to validate, not proof that evaluation has become objective.

## Influential metrics and benchmarks

| Family | Examples and original sources | Appropriate role in v-eval |
| --- | --- | --- |
| Reference similarity | [BLEU, 2002](https://aclanthology.org/P02-1040/); [ROUGE, 2004](https://aclanthology.org/W04-1013/); [BERTScore, 2019/2020](https://arxiv.org/abs/1904.09675) | Optional task-specific similarity baselines. They do not directly verify truth or user value. |
| Factual support and citation | [FActScore, 2023](https://aclanthology.org/2023.emnlp-main.741/); [ALCE, 2023](https://aclanthology.org/2023.emnlp-main.398/) | Decompose claims and check support; measure omissions separately from supported-claim precision. |
| Retrieval quality | [Recall@k, MRR, nDCG definitions](https://ir-measur.es/en/latest/measures.html) | Measure retrieval against relevance judgments; record cutoff, unit, labels, and implementation. |
| RAG quality | [RAGAs, 2024](https://aclanthology.org/2024.eacl-demo.16/) | Separate answer faithfulness, relevance, and context quality. Reference-free metrics are estimators. |
| Explicit instruction following | [IFEval, 2023](https://arxiv.org/abs/2311.07911); [CodeIF, 2025](https://arxiv.org/abs/2502.19166) | Check verifiable constraints directly; use separate evidence for semantic requirements. |
| Model-based judgment | [G-Eval, 2023](https://aclanthology.org/2023.emnlp-main.153/); [Prometheus 2, 2024](https://arxiv.org/abs/2405.01535) | Optional rubric judges calibrated on the actual task population. |
| Evaluating the judge | [MT-Bench/Chatbot Arena study, 2023](https://arxiv.org/abs/2306.05685); [LLMBar, 2023/2024](https://arxiv.org/abs/2310.07641) | Test position, presentation and instruction-compliance failures; preference agreement is not proof of truth. |
| Function-level coding | [HumanEval/pass@k, 2021](https://arxiv.org/abs/2107.03374); [EvalPlus, 2023](https://arxiv.org/abs/2305.01210) | Learn from executable evaluation and stronger test suites. A model's benchmark score does not accept a particular patch. |
| Repository and fresh coding tasks | [SWE-bench, 2023/2024](https://arxiv.org/abs/2310.06770); [LiveCodeBench, 2024](https://arxiv.org/abs/2403.07974) | Learn artifact/environment binding and contamination-aware dataset design. |
| Test adequacy and maintainability | [Stryker mutation metrics](https://stryker-mutator.io/docs/mutation-testing-elements/mutant-states-and-metrics/); [Maintainability Index](https://learn.microsoft.com/en-us/visualstudio/code-quality/code-metrics-maintainability-index-range-and-meaning?view=visualstudio) | Separate test sensitivity and structural proxies from functional acceptance. |
| Quality and security taxonomies | [ISO/IEC 25010:2023 public overview](https://www.iso.org/standard/78176.html); [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/) | Choose relevant criteria; these are not a universal automatic quality score or a certification provided by v-eval. |

For long context windows, [RULER (2024)](https://arxiv.org/abs/2404.06654) tests retrieval, tracing, and aggregation at different lengths. It evaluates a model's ability to use context; it does not tell you whether a particular supplied context packet contains the right information. That is a separate task-specific check.

**Current benchmark caveat:** OpenAI's [2026-02-23 reassessment of SWE-bench Verified](https://openai.com/index/why-we-no-longer-evaluate-swe-bench-verified/) documents test flaws and contamination and explains its move away from the benchmark for frontier reporting. This is a reason to audit evaluation tasks, not to conclude that all executable benchmarks are useless. The audit's selected-subset findings should not be generalized to every task.

Read detailed source notes for [content and context](research/content-context.md) and [code quality](research/code.md). These include dates, exact source links, limitations, and suggested uses.

## Existing projects to study before building

| Project | What to learn or reuse | License observed on 2026-09-14 |
| --- | --- | --- |
| [promptfoo](https://github.com/promptfoo/promptfoo) | Declarative assertions, providers, local evaluation, CI | [MIT](https://github.com/promptfoo/promptfoo/blob/main/LICENSE) |
| [DeepEval](https://github.com/confident-ai/deepeval) | Python test cases, G-Eval rubrics, RAG and agent metrics | [Apache-2.0](https://github.com/confident-ai/deepeval/blob/main/LICENSE.md) |
| [Ragas](https://github.com/vibrantlabsai/ragas) | Context, answer and factuality metrics | [Apache-2.0](https://github.com/vibrantlabsai/ragas/blob/main/LICENSE) |
| [Inspect AI](https://github.com/UKGovernmentBEIS/inspect_ai) | Tasks, scorers, execution logs and configurable sandboxes | [MIT](https://github.com/UKGovernmentBEIS/inspect_ai/blob/main/LICENSE) |
| [OpenAI Evals repository](https://github.com/openai/evals) | Extensible evaluation cases and recorded results | [MIT framework; separate dataset notices](https://github.com/openai/evals/blob/main/LICENSE.md) |
| [TruLens](https://github.com/truera/trulens) | App traces and composable feedback functions | [MIT](https://github.com/truera/trulens/blob/main/LICENSE) |
| [structured-evaluation](https://github.com/plexusone/structured-evaluation) | Requirement traceability, evidence rules, report schema and blocking criteria | [MIT](https://github.com/plexusone/structured-evaluation/blob/main/LICENSE) |
| [Phoenix / phoenix-evals](https://github.com/Arize-ai/phoenix) | Adjacent tracing, datasets and experiments ecosystem | [Elastic-2.0](https://github.com/Arize-ai/phoenix/blob/main/LICENSE), source-available with restrictions |

The Phoenix distinction matters when choosing dependencies; a public source repository is not necessarily licensed like MIT or Apache software. This table summarizes inspected files, not legal advice or a license compatibility audit. Hosted offerings and bundled datasets can have different terms.

[The detailed framework comparison](research/frameworks.md) covers integration fit and links directly to implementation/docs. **structured-evaluation is the closest discovered prior art:** its [rubrics](https://github.com/plexusone/structured-evaluation/blob/main/docs/features/rubrics.md) already address requirement references and evidence discipline. Generic rubrics, evidence explanations, and blocking criteria are not new contributions by v-eval.

## What the research suggests building

Our proposed contribution is a learnable workflow that connects a user's requirement to an inspectable observation and an acceptance decision. It may ultimately be a skill and adapter over existing tools. Whether it deserves a standalone program is an open product question.

Three design choices follow from the reviewed failure modes: keep independent dimensions visible, distinguish insufficient evidence from a demonstrated failure, and evaluate the evaluator on held-out human-reviewed cases. These choices make the report easier to audit in principle; their accuracy, usability, and cost benefits must be measured.

Anthropic's [2026 practical guide to agent evaluation](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents) also recommends combining executable, model-based, and human assessment and checking actual outcomes. We use that as supporting practice guidance, not as validation of this prototype.

The current repository contains research, teaching material, and prompt instructions. It contains no benchmark results demonstrating that v-eval outperforms an existing evaluator. [The comparison plan](comparison.md) explains the evidence needed before making that claim.
