# Why this approach might help, and how to compare it fairly

**v-eval has not demonstrated that it outperforms existing evaluation tools.** It contains research, a contract template, a draft skill, and an illustrative review. There is no automated runner or comparative measurement. The proposed advantage is inspectable acceptance decisions: connect requirements to evidence, expose missing information, and explain the next action.

The code example does not settle the product's scope. Understanding evaluation of supplied or retrieved context is also a priority. Context, documents, and code require different evidence. See [the design](design.md) for the proposed policy.

## What exists and what remains proposed

| Capability | Current v-eval status | What would still need implementation or validation |
| --- | --- | --- |
| Explicit requirements and acceptance rules | Markdown contract template | Parsing, version tracking, and validation |
| Criterion-level findings with evidence | Draft skill instructions and worked example | Reliable collection and mechanically checked references |
| PASS / FAIL / UNKNOWN / ERROR distinctions | Written policy followed by the host assistant | Deterministic report validation and aggregation |
| Checking sources, tests, and artifact identity | Instructions to inspect and record evidence | Adapters, reproducible runs, stale-evidence detection |
| Context, document, and code review | Described in the skill | Representative examples and measured reliability for each |
| Better decisions or lower review effort | Product hypothesis | Independent labels and comparative experiments |

The host assistant supplies reasoning and tools. Skill instructions do not guarantee compliance or create a security boundary. Future software could enforce structural invariants; semantic judgments would still require measurement.

## Why the mechanisms are plausible

Consider a context packet intended to help answer a question about a product's current refund policy. It contains clear, relevant prose, but one passage describes last year's policy and another omits an exception. A relevance score alone would not settle whether this packet is sufficient for the requested answer.

An explicit contract could require the current policy version, exception coverage, and identification of conflicts. The report would attach passages to conclusions and mark freshness UNKNOWN when dates are unavailable. These inspectable checks remain fallible if the authoritative source is wrong or unavailable.

| Proposed mechanism | Improvement hypothesis | Limitation to test |
| --- | --- | --- |
| Trace requirements to the user's task | Catch omissions hidden by fluent output | The extracted requirements may misrepresent intent |
| Select evidence appropriate to each criterion | Avoid treating similarity or a test name as proof | Evidence selection can miss the decisive passage/assertion |
| Attach source locations and artifact identity | Reduce unsupported conclusions and stale-log acceptance | A valid reference may still fail to support the conclusion |
| Preserve UNKNOWN and operational ERROR | Reduce acceptance when required evidence is missing | Excessive abstention can make the evaluator unusable |
| Apply mandatory criteria separately from optional quality | Prevent a serious violation from disappearing in an average | Poorly chosen mandatory criteria can reject useful work |

For context evaluation, distinguish retrieval relevance, coverage, source support, and downstream usefulness. Ragas already provides separate context and answer metrics; these dimensions are not new. A useful v-eval experiment would test whether task-specific evidence requirements improve decisions beyond a well-configured metric suite. [Ragas metrics](https://docs.ragas.io/en/stable/concepts/metrics/overview/).

## Existing tools are strong baselines

**promptfoo** already combines mechanical assertions, custom functions, model rubrics, context checks, thresholds, and custom aggregation. **DeepEval** supports explicit criteria/evaluation steps, expected outputs, retrieval context, explanations, and pytest integration. They currently provide executable evaluation infrastructure that v-eval lacks. Neither should be represented by a weak, generic “rate this from 1–10” prompt in a comparison. [promptfoo assertions](https://www.promptfoo.dev/docs/configuration/expected-outputs/), [DeepEval G-Eval](https://deepeval.com/docs/metrics-llm-evals), [DeepEval examples](https://github.com/confident-ai/deepeval).

**structured-evaluation is the closest discovered prior art.** Its implemented Go types and documentation cover requirement traceability, deterministic/semantic/human evaluation methods, blocking criteria, and instructions to cite evidence and distinguish missing from negative evidence. It also supplies report schemas, validation and rendering. Consequently, neither criterion-based evidence nor mixed evaluation methods establishes novelty. Its report layer needs an evaluator/collector around it for this comparison; it would be unfair to score the schema alone as though it were an autonomous judge. [Rubrics](https://github.com/plexusone/structured-evaluation/blob/main/docs/features/rubrics.md), [implemented types](https://github.com/plexusone/structured-evaluation/blob/main/rubric/rubric.go), [CLI and packages](https://github.com/plexusone/structured-evaluation).

Inspect supplies task execution and sandbox integration; TruLens evaluates instrumented applications and agent plans. These capabilities could support this workflow. This research has not established that existing tools cannot implement the approach. [Inspect tasks](https://inspect.aisi.org.uk/tasks.html), [Inspect sandboxing](https://inspect.aisi.org.uk/sandboxing.html), [TruLens](https://github.com/truera/trulens).

The recommended experiment is a thin skill/adapter using existing evaluation infrastructure, with possible structured-evaluation interoperability. Build a separate runner only if prototypes identify unmet requirements or substantially simpler operation. If a recipe for an existing framework delivers equivalent decisions and usability, publishing that recipe is a successful outcome.

## A fair experiment

Start with a public pilot of 30 base tasks: ten context packets, ten documents, and ten code changes. Create three variants per task, including acceptable work, a subtle violation, and missing or conflicting evidence. This proposed 90-case pilot tests feasibility; it is too small to establish general superiority.

1. **Establish labels independently.** Two reviewers label each criterion and overall acceptability before seeing evaluator results. Adjudicate disagreements and preserve the original labels. Allow an indeterminate label where the contract or authoritative evidence is insufficient. Include realistic cases alongside seeded defects, and disclose their proportions.

2. **Separate tuning from measurement.** Split by base-task family, keeping variants and near-duplicates together. Use 20 families for development and ten as held-out cases. Freeze rubrics, prompts, adapters, thresholds, versions, and the analysis plan before revealing held-out outcomes. Broader claims require a larger, separately collected set.

3. **Provide equal inputs and resources.** Give every system the same artifacts, fixed contracts, sources, tools, and execution environment. Match judge model/version, total judge-token or monetary budget, retry limits, and timeouts. Count extraction and verification calls within the budget. Report budget overruns and failures. Separately test criterion extraction from vague intent, using independent reference requirements; do not confound extraction with grading.

4. **Use competent configurations.** Compare ordinary assistant review; promptfoo with criterion-specific assertions and rubrics; DeepEval with explicit evaluation steps; structured-evaluation plus a judge and collector; and the v-eval draft skill. Give baselines equivalent tuning effort, evidence access, mandatory-gate rules, and an abstention option. Later runner results must be a separately versioned condition, not attributed retroactively to the draft.

5. **Test component contributions.** Run paired ablations that remove evidence-reference verification, artifact/version checks, or criterion-level decomposition one at a time while preserving input access and budget. Separately compare aggregation/abstention policies. This distinguishes improvements from extra model calls, conservative thresholds, or report formatting.

6. **Measure decisions and developer effort.** Repeat model-based runs to expose variability. Randomize and blind report order for timed human review; measure correct final decisions and time to identify a repair or missing evidence. Report token/tool cost, setup effort, and median/tail latency. Use paired comparisons and uncertainty intervals grouped by task family; report disagreement and small-sample limitations.

### Denominators matter

Let **U** be independently labeled unacceptable cases, **A** acceptable cases, **X** indeterminate cases, and **N = U + A + X**. Count evaluator outcomes as accepted, rejected, or incomplete; operational errors are incomplete and additionally reported separately.

- **False acceptance rate:** accepted cases in U / all cases in U.
- **Error among acceptances:** accepted cases in U / accepted cases with determinate labels (A or U). This answers a different question from false acceptance.
- **False rejection rate:** rejected cases in A / all cases in A. Report incomplete cases in A separately.
- **Abstention rate:** incomplete cases / N. Also report it within A, U, and X.
- **Unsupported acceptance:** accepted cases in X / all cases in X; do not silently relabel X as incorrect.
- **Decided-case accuracy:** correct accept/reject decisions among A and U / all accept/reject decisions among A and U. Always pair it with abstention.

Show raw counts. A zero denominator is undefined, not zero. Audit cited evidence independently for actual support, not merely valid URLs. Compare false acceptance at similar abstention and cost; a system that refuses everything has not demonstrated useful superiority.

### Results

No experiment has been run.

| Configuration | False acceptance | False rejection / abstention | Evidence validity | Cost / latency / review time |
| --- | --- | --- | --- | --- |
| Ordinary assistant | Not measured | Not measured | Not measured | Not measured |
| Configured promptfoo | Not measured | Not measured | Not measured | Not measured |
| Configured DeepEval | Not measured | Not measured | Not measured | Not measured |
| structured-evaluation + judge/collector | Not measured | Not measured | Not measured | Not measured |
| v-eval draft skill | Not measured | Not measured | Not measured | Not measured |

Evidence for superiority would mean better acceptance decisions or lower developer effort under comparable conditions, with uncertainty reported. Until then, this is a testable workflow proposal and learning project.
