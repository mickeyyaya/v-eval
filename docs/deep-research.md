# Evaluation diversity and scoring

AI evaluation works best as a collection of measurements tied to a decision. An executable test, a citation-support check, a reader's judgment, and a latency percentile can all be useful. They measure different things, require different evidence, and fail in different ways. A credible evaluation system makes those differences visible.

This report develops that position from original research and practitioner writing, including material available in 2026. Its practical conclusion is to combine explicit acceptance criteria, appropriate deterministic checks, carefully scoped semantic judgments, and inspection of real failures. This is the rationale for v-eval's proposed workflow. It is not evidence that the prototype already outperforms existing tools.

## 1. What is being evaluated?

There are at least four distinct objects: a model's general capability, an application's behavior, a particular artifact, and the evaluator itself. A model benchmark estimates performance on a task distribution. An application evaluation includes prompts, retrieval, tools, and state. Artifact acceptance asks whether one output satisfies a specific job. Evaluator validation asks whether its conclusions agree with independently established evidence.

Confusing these objects causes misleading claims. A strong coding benchmark result does not prove that a particular patch respects an architectural constraint. A judge's agreement with conversational preferences does not show that it can validate a technical handoff. A correct final answer does not demonstrate that the retrieval stage was relevant or efficient.

Hamel Husain's product-evaluation case study combines scoped tests, human/model review, and online experiments. Shreya Shankar's later essay describes evaluation as systematic quality measurement whose rigor can vary with the situation. These are compatible with a small initial workflow; neither requires a dashboard full of generic scores.[^1][^2]

The proposed v-eval unit is an **artifact plus an evaluation contract**. The contract records the intended user, requested outcome, relevant sources, criteria, permitted methods, and acceptance rules. The report attaches evidence to each criterion and preserves uncertainty. This unit can support code, documents, retrieved material, or generated context, provided each uses suitable checks.

## 2. Diversity has several meanings

Evaluation diversity is not simply running more judges. It includes diversity of tasks, situations, evidence, measurement methods, and affected users. These dimensions should be chosen because they expose different ways the application can fail.

| Axis | Examples | Why it can change the conclusion |
| --- | --- | --- |
| Task | Retrieval, summarization, planning, code change, final action | A skill at finding a fact does not establish skill at synthesizing a plan. |
| Artifact | Plain text, structured data, source code, generated memory | Different properties can be verified directly. |
| Evidence condition | Complete, missing, conflicting, stale | Fluent output can obscure a missing basis for the answer. |
| Input variation | Length, language, document format, source position | A strong average may hide a fragile condition. |
| User and domain | Novice, maintainer, researcher; task-specific terminology | The same detail can be necessary for one reader and distracting for another. |
| Failure mechanism | Omission, invented claim, invalid action, broken assertion, injection | A single metric rarely diagnoses all of them. |
| Measurement | Executable, statistical, rubric-based, human, operational | Each supplies different evidence and uncertainty. |

HELM explicitly combines scenario coverage with several measurements rather than allowing accuracy to displace every other concern. CheckList provides a complementary behavioral-testing method: test minimum capabilities, invariances, and expected directional changes. Their lesson for a product is to design a coverage matrix around meaningful behavior, rather than copy a benchmark's aggregate score.[^3][^4]

For generated context, useful variations include a changed decision near the end of a conversation, an unresolved proposal, a stale source, contradictory documents, and a missing prerequisite. Reordering irrelevant passages should not change supported claims. Changing the source's actual decision should change the summary. These are proposed product tests inspired by behavioral testing, not standardized v-eval benchmark results.

Keep a representative sample separate from targeted challenge cases. Oversampling rare failures helps discover and repair them, but the resulting failure proportion is not automatically a production prevalence estimate. Report important slices and denominators; do not let many easy cases conceal a smaller critical slice.

## 3. Context quality needs its own evaluation

Retrieved context selects existing material. Generated context transforms material into a summary, plan, memory, or handoff. A generated summary can preserve factual sentences while changing their status: a suggestion becomes a commitment, an old decision becomes current, or an unresolved question disappears.

Evaluate at least three connections: the task to the selected evidence, the evidence to the generated context, and the context to the downstream answer or action. Jason Liu's question-context-answer framing is a useful starting checklist. Its compact taxonomy should not be treated as an exhaustive standard; freshness, permissions, and operational cost still need separate definitions when relevant.[^5]

Long-context benchmarks add another distinction. RULER expands beyond finding one inserted fact; HELMET reports that synthetic retrieval performance can be a poor predictor of diverse downstream tasks. Those findings concern model capability under specified experiments. They do not directly certify the completeness of a context packet prepared for a particular developer.[^6][^7]

For a handoff, evidence can include the original discussion, the order of decisions, unresolved questions, linked implementation changes, and verified test records. A useful evaluation asks whether a later reader would reach the intended understanding. If the original discussion is unavailable, comprehensive preservation cannot be verified simply by reading the summary.

Harrison Chase's observability discussion connects individual calls, full executions, and conversations, including state changes whose consequences appear later. This is useful guidance for deciding what evidence to retain. A trace shows recorded inputs, actions, and outputs; it is not privileged access to the model's true internal reasoning.[^8]

## 4. Choose the score by its meaning

Scores are useful when readers know what they count, what changes them, and which decision they support. There is no reason to force every dimension onto one scale.

| Output form | Suitable question | Required interpretation |
| --- | --- | --- |
| Binary pass/fail | Does a defined criterion hold? | State the rule, evidence, and uncertainty separately. |
| Count or proportion | How many required items or labeled relevant items were recovered? | State the numerator, denominator, exclusions, and unit. |
| Anchored ordinal grade | How does an artifact compare with defined quality levels? | Explain each level; adjacent intervals are not automatically equal. |
| Pairwise preference | Which of two acceptable alternatives better serves this preference? | Allow ties, record position instability, and keep absolute acceptance separate. |
| Native engineering metric | What is the latency, defect severity, or complexity? | Preserve units, version, workload, and threshold authority. |
| Composite | How does an explicit policy trade off several measurements? | Disclose weights, normalization, missing-value policy, and mandatory gates. |

Husain recommends starting with binary judgments and critiques. Eugene Yan distinguishes direct judgments from pairwise preferences and emphasizes interpretable decisions. These are practical defaults. They are not a prohibition on an ordinal scale whose distinctions matter and can be applied consistently.[^9][^10]

A pairwise winner can still be unacceptable. Two answers may both invent the requested deadline; one merely reads better. Conversely, a valid alternative implementation can differ from the reference text while satisfying every behavioral requirement. Preference and reference similarity should not silently become correctness.

For security, performance, maintainability, and debuggability, use the [industry scorecard](industry-scorecard.md). It distinguishes formal standards, established metrics, vendor ratings, and proposed project rubrics. CVSS describes vulnerability severity, not total system security. A latency percentile needs a workload. A named maintainability formula is a structural proxy. Where no universal scalar exists, a clearly labeled rubric is more honest than an invented industry index.

The v-eval proposal therefore supports dimension scores with detailed reasoning. It avoids a default all-purpose average because unlike scales and missing evidence can conceal required failures. A product can later define a justified composite for a specific decision, while retaining the underlying measurements and sensitivity to its weights.

## 5. Deterministic checks and their limits

Many useful checks require no model judgment: parsing a structured output, validating a schema, checking required fields, counting words under a declared tokenization rule, applying style rules, compiling code, or executing a behavior test. Existing tools can supply this evidence. The [2026 deterministic-tools guide](deterministic-tools-2026.md) records current releases and newer research separately.

Determinism concerns repeatability under fixed conditions. It does not guarantee that the criterion captures the intended quality. A regex can reliably find a heading even when the section is empty. A schema can accept a correctly typed invented date. A test can consistently enforce the wrong expected behavior.

Specify the boundary carefully. A pipeline that asks a model to extract claims and then checks the resulting JSON is hybrid: validation is deterministic, extraction is not thereby proven correct. A workflow graph can execute a fixed sequence while its model judges remain variable. Likewise, fetching a URL depends on changing external state; successful retrieval does not establish that the source supports a claim.

For code, distinguish deterministic analysis from environmental reproducibility. Pin tools, rules, inputs, seeds where applicable, and runtime configuration. Record collected and skipped tests, failures, and logs. Latency measurements vary even when the test procedure is fixed; repeated observations and a stated workload are part of the result.

The practical principle is to automate computable checks and keep semantic uncertainty visible. Passing a machine check should establish the property it actually verifies, not an expanded claim such as “this output is useful.”

## 6. Validate semantic judgments

An LLM judge can compare evidence against a rubric, but the explanation it produces can also be wrong. Eugene Yan's survey distinguishes several judging setups and their validation measures. Chip Huyen's implementation advice similarly treats judges as systems whose agreement with human judgment needs ongoing inspection.[^11][^12]

Begin with one narrow failure mode. Have an appropriate domain owner define examples of acceptable and unacceptable behavior, retain their reasoning, and identify ambiguities. Use separate development and held-out cases. Independently labeled cases help expose disagreements that a single reviewer might otherwise resolve invisibly.

“Calibration” is used loosely in practitioner writing. Agreement with human labels is useful, but it is different from probabilistic calibration. A judge that says “90% confident” has not established that nine out of ten comparable judgments are correct. Treat claimed confidence as unvalidated unless a separate study connects reported probabilities to observed outcomes.

For pairwise judging, conceal irrelevant generator identity and test both presentation orders. A changed preference is instability to investigate, not automatically evidence that the alternatives are equally good. Adding a second judge may help, but correlated errors and extra cost must be measured. Diverse evidence is often a more concrete intervention than assigning several expert personas.

## 7. Requirements evolve, but comparisons need a fixed contract

It is reasonable to define known constraints before building. It is also reasonable to discover unanticipated failures by inspecting real outputs. These activities answer different questions.

Husain and Shankar's FAQ emphasizes error discovery and cautions against inventing expensive evaluators for speculative failure modes. Anthropic's agent-evaluation guide advocates evaluation-driven development for capabilities with identifiable success criteria. The practical synthesis is to write known acceptance checks early, then improve the taxonomy as real behavior reveals gaps.[^13][^14]

EvalGen's research documents **criteria drift**: inspecting outputs can help people articulate what they want, so evaluation criteria are not always fully independent of observed examples. This supports an iterative discovery phase. It also creates a need to freeze and version the contract for an actual comparison, then reevaluate candidates if the contract changes.[^15]

Do not punish a candidate for a newly invented requirement while pretending it was part of the original task. Conversely, a narrow supplied checklist can omit an explicit requirement in the brief. Surface the gap, record its authority, and resolve the contract before claiming full acceptance.

## 8. Gaming is a failure of the measurement relationship

An output can score well because it satisfies the intended goal, exploits a proxy, encounters leaked answers, or persuades the judge. The observable evidence can distinguish some of these paths. It usually cannot establish hidden intent.

DeepMind's specification-gaming examples illustrate the gap between literal objectives and intended outcomes. Lilian Weng's reward-hacking survey traces related mechanisms across learning systems. These motivate testing the link between score and usefulness rather than treating a high score as self-validating.[^16][^17]

In 2026, Anthropic documented BrowseComp cases involving leaked answers and benchmark identification. The authors did not call the unrestricted search behavior an alignment failure. The lesson is to define allowed methods and protect evaluation integrity, rather than label every unexpected solution cheating.[^18]

For artifact evaluation, concrete challenge cases include judge-directed text embedded in a document, unsupported success statements, stale execution logs, weakened tests, a real citation attached to a stronger claim, and attractive but irrelevant prose. For each attack case, include a benign control or legitimate alternative. A stricter evaluator that rejects harmless unusual work has not necessarily improved.

The [gaming guide](gaming-and-defenses.md) separates prompt instructions from controls a future runner must enforce. A prompt can request separation of trusted criteria and candidate data. Software must enforce any actual execution isolation, protected test authority, evidence collection, and deterministic aggregation. Neither layer fixes a misunderstood requirement by itself.

## 9. Scoring the evaluator requires explicit denominators

Consider this **synthetic example**, not a v-eval experiment. Humans label 200 artifacts: 40 unacceptable and 160 acceptable. The evaluator accepts 6 unacceptable artifacts and rejects the other 34. Among acceptable artifacts, it accepts 152, rejects 4, and leaves 4 incomplete.

| Measure | Calculation | Interpretation |
| --- | --- | --- |
| False acceptance rate | 6 / 40 = 15% | Fraction of known unacceptable artifacts that slip through. |
| Error among acceptances | 6 / 158 ≈ 3.8% | Fraction of accepted artifacts that are unacceptable in this dataset. |
| False rejection rate | 4 / 160 = 2.5% | Fraction of acceptable artifacts incorrectly rejected. |
| Incomplete rate | 4 / 200 = 2% | Fraction without an acceptance/rejection decision. |
| Accuracy among decisions | (34 + 152) / 196 ≈ 94.9% | Correct decisions among cases where a decision was made. |

The attractive 94.9% number coexists with accepting 15% of known bad artifacts. Error among acceptances also depends on the mix of acceptable and unacceptable work. Report all these quantities, raw counts, and important slices. An undefined denominator is not zero.

If the human label is indeterminate, keep it separate from known valid and invalid cases. An unsupported acceptance in that group is an evidence concern, not proof the artifact is factually wrong. If a system abstains more, compare it at similar decision coverage or show the tradeoff explicitly. Refusing every case produces no false acceptances but little practical value.

Sample size affects interpretation. A handful of examples is enough to expose a failure, but not to establish a low failure rate. If the true independent failure probability were 10%, seeing no failures in ten trials would still have probability `0.9^10 ≈ 35%`. Real cases can be correlated, making the effective information smaller. A pilot is for feasibility; a superiority claim needs a separately designed comparison with uncertainty and sufficient coverage.

## 10. Operational outcomes belong beside artifact quality

For an agent, a plausible transcript is not the same as successful action. Verify the relevant final state: a file exists with the required content, a supported operation occurred, or a requested result is observable. Then measure time, cost, resource use, and constraint adherence separately.

The Deep Agents team describes grouping cases by capability and examining correctness before efficiency. Andrew Ng's error-analysis example traces a weak report through its stages while allowing the overall decomposition itself to be questioned. These suggest a useful diagnostic sequence: establish the outcome, identify the failed connection, then optimize the relevant component.[^19][^20]

Online feedback complements offline tests. Reader success, correction effort, abandonment, and support burden can reveal product problems that an offline rubric misses. Those outcomes have multiple causes; a change in them is not automatically caused by the model or evaluator. Use appropriate comparisons and keep the operational interpretation explicit.

## 11. What 2026 practitioner work adds

Recent publications strengthen the case for small, inspectable workflows. Hamel Husain and Shreya Shankar published evaluation skills covering error discovery, judge construction, validation, and RAG evaluation. Simon Willison's smevals announcement describes small file-based suites and separate execution and grading. These are direct prior art for a skill-oriented project, not reasons to claim that v-eval invented the approach.[^21][^22]

Greg Kamradt's ARC-AGI-3 human-baseline report describes changes to normalization after inspecting scoring behavior. It illustrates that aggregation choices can change what a benchmark rewards. Its particular game-based metrics should not be transferred to document quality.[^23]

Original social posts are useful for locating emerging questions, but their brevity often leaves assumptions unspecified. The [reading list](practitioner-reading-list.md) includes accessible originals and records access gaps without reconstructing inaccessible X threads from quotations. Practitioner reputation helps locate relevant experience; it does not replace evidence for a specific claim.

## 12. The resulting v-eval proposal

The proposed report begins with the intended job and applicable perspectives. It shows per-criterion acceptance, native dimension scores when meaningful, exact evidence, limitations, and the next action. A source verifier, requirements reviewer, performance check, and maintainability analysis can contribute different observations without voting away a mandatory failure.

The implementation should reuse existing tools where they fit. structured-evaluation already supplies close rubric/reporting concepts, while general evaluation frameworks provide assertions, judges, execution, and logs. A new project earns its place only if its workflow improves evidence quality, setup effort, review decisions, or another explicitly measured outcome.

Test the draft skill against ordinary review and competently configured existing tools on the same cases, contracts, evidence, and budgets. Give baselines equivalent tuning and abstention choices. Use ablations to separate the benefit of evidence binding, criterion decomposition, and score presentation from the benefit of simply making more model calls. The [comparison protocol](comparison.md) makes those conditions concrete.

Current evidence supports a research-backed design hypothesis and a learning artifact. It does not establish a working automated runner, universal quality index, proven defense against gaming, or superiority over other evaluators. Those are separate implementation and measurement questions.

## Sources and scope

The sources below include original research, firsthand engineering accounts, and practitioner advice. Their publication dates are preserved in the notes; living pages may contain later revisions. Material was selected for relevance to evaluation methods, context quality, scoring, and integrity, with sources available through 2026-09-14. This is a curated English-language reading corpus, not an exhaustive collection of posts or a systematic meta-analysis. Commercial interests and benchmark-specific assumptions remain relevant limitations.

The [practitioner reading list](practitioner-reading-list.md) provides a wider annotated inventory, including social sources and access limitations. [Foundational research](research.md), [deterministic tools](deterministic-tools-2026.md), and the [industry scorecard](industry-scorecard.md) provide the metric and implementation references.

[^1]: Hamel Husain. [Your AI Product Needs Evals](https://hamel.dev/blog/posts/evals/). March 29, 2024. Firsthand case study and practice guidance.
[^2]: Shreya Shankar. [In Defense of AI Evals, for Everyone](https://www.sh-reya.com/blog/in-defense-ai-evals/). September 5, 2025. Practitioner argument.
[^3]: Percy Liang et al. [Holistic Evaluation of Language Models](https://arxiv.org/abs/2211.09110). First submitted November 16, 2022; TMLR 2023. Benchmark research.
[^4]: Marco Tulio Ribeiro, Tongshuang Wu, Carlos Guestrin, Sameer Singh. [Beyond Accuracy: Behavioral Testing of NLP Models with CheckList](https://aclanthology.org/2020.acl-main.442/). ACL, July 2020. Method and empirical studies.
[^5]: Jason Liu. [There Are Only 6 RAG Evals](https://jxnl.co/writing/2025/05/19/there-are-only-6-rag-evals/). May 19, 2025. Conceptual framework.
[^6]: Cheng-Ping Hsieh et al. [RULER: What's the Real Context Size of Your Long-Context Language Models?](https://arxiv.org/abs/2404.06654). First submitted April 9, 2024. Benchmark research.
[^7]: Howard Yen et al. [HELMET: How to Evaluate Long-Context Language Models Effectively and Thoroughly](https://arxiv.org/abs/2410.02694). First submitted October 3, 2024. Benchmark research.
[^8]: Harrison Chase. [Agent observability powers agent evaluation](https://www.langchain.com/blog/agent-observability-powers-agent-evaluation). January 27, 2026. Vendor-authored architecture guidance.
[^9]: Hamel Husain. [Using LLM-as-a-Judge For Evaluation: A Complete Guide](https://hamel.dev/blog/posts/llm-judge/). October 29, 2024; current page modified September 1, 2026. Practitioner guide.
[^10]: Eugene Yan. [Product Evals in Three Simple Steps](https://eugeneyan.com/writing/product-evals/). November 23, 2025. Practitioner process guidance.
[^11]: Eugene Yan. [Evaluating the Effectiveness of LLM-Evaluators (aka LLM-as-Judge)](https://eugeneyan.com/writing/llm-evaluators/). August 18, 2024. Research survey and synthesis.
[^12]: Chip Huyen. [Common pitfalls when building generative AI applications](https://huyenchip.com/2025/01/16/ai-engineering-pitfalls.html). January 16, 2025. Practitioner experience and examples.
[^13]: Hamel Husain and Shreya Shankar. [AI Evals: Everything You Need to Know](https://hamel.dev/blog/posts/evals-faq/). May 28, 2025; modified September 1, 2026. Living practitioner FAQ.
[^14]: Mikaela Grace, Jeremy Hadfield, Rodrigo Olivares, and Jiri De Jonghe, Anthropic. [Demystifying evals for AI agents](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents). January 9, 2026. Engineering practice guide.
[^15]: Shreya Shankar, J. D. Zamfirescu-Pereira, Björn Hartmann, Aditya G. Parameswaran, Ian Arawjo. [Who Validates the Validators? Aligning LLM-Assisted Evaluation of LLM Outputs with Human Preferences](https://arxiv.org/abs/2404.12272). First submitted April 18, 2024; UIST 2024. Research study.
[^16]: Victoria Krakovna et al., DeepMind. [Specification gaming: the flip side of AI ingenuity](https://deepmind.google/blog/specification-gaming-the-flip-side-of-ai-ingenuity/). April 21, 2020. Research overview.
[^17]: Lilian Weng. [Reward Hacking in Reinforcement Learning](https://lilianweng.github.io/posts/2024-11-28-reward-hacking/). November 28, 2024. Research survey.
[^18]: Russell Coleman, Anthropic. [Eval awareness in Claude Opus 4.6's BrowseComp performance](https://www.anthropic.com/engineering/eval-awareness-browsecomp). March 6, 2026. Investigation of recorded benchmark behavior.
[^19]: Vivek Trivedy, Mason Daugherty, Eugene Yurtsev, Harrison Chase. [How we build evals for Deep Agents](https://www.langchain.com/blog/how-we-build-evals-for-deep-agents). March 26, 2026. Team implementation report.
[^20]: Andrew Ng. [Improve Agentic Performance with Evals and Error Analysis, Part 2](https://www.deeplearning.ai/the-batch/improve-agentic-performance-with-evals-and-error-analysis-part-2/). October 22, 2025. Authored newsletter; indexed first-party text was available despite direct-fetch restrictions.
[^21]: Hamel Husain and Shreya Shankar. [Evals Skills for Coding Agents](https://hamel.dev/blog/posts/evals-skills/). March 2, 2026; modified August 31, 2026. Creator announcement and workflow description.
[^22]: Simon Willison. [smevals — a small eval suite for evaluating models, prompts, and harnesses](https://simonwillison.net/2026/Jul/31/smevals/). July 31, 2026. Creator announcement and demonstration.
[^23]: Greg Kamradt. [Measuring Human Performance on ARC-AGI-3](https://arcprize.org/blog/arc-agi-3-human-dataset). April 14, 2026. Benchmark-owner human-study and scoring report.
