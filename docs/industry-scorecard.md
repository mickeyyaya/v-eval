# Industry metrics and a practical evaluation scorecard

Research checked **2026-09-14**. This is a learning reference and proposed reporting design. v-eval does not currently run these tools or calculate these scores. The numerical example below is synthetic.

Evaluate security, performance, maintainability, and other dimensions numerically using measures that answer the actual questions. Severity 9.3, maintainability 42, and latency 420 milliseconds describe different properties. Averaging them conceals meaning and could let attractive formatting compensate for a critical security failure. [FIRST also cautions against merging unrelated dimensions into cybersecurity scoring](https://www.first.org/cvss/v4.0/faq).

Start with acceptance criteria in [the design](design.md). See [the standards reference](standards-reference.md) for formal frameworks and their status, [the learning guide](learn-evaluation.md) for fundamentals, [the research](research.md) for papers, and [the comparison](comparison.md) for tools.

## 1. Recognize four kinds of measurement

| Kind | Examples | What the label establishes |
| --- | --- | --- |
| Published standard or quality model | FIRST CVSS 4.0; ISO/IEC 25010:2023 | A shared definition or framework within its stated scope. It does not establish that every product passes. |
| Established measurement | Latency percentiles, error rate, cyclomatic complexity, test coverage | A recognizable quantity; counting conventions and acceptable targets still need definitions. |
| Organization or vendor measure | Google Core Web Vitals thresholds; Microsoft Maintainability Index; Sonar ratings; OpenSSF Scorecard | A published implementation or recommendation. Preserve its provider, version, and configuration. |
| Proposed project rubric | A four-check debuggability score; required-context coverage | A local operational definition. Explain its anchors and validate its usefulness. |

[ISO/IEC 25010:2023](https://www.iso.org/standard/78176.html), edition 2, November 2023, provides a nine-characteristic product quality model. Its public abstract describes organizing requirements and evaluations, rather than a scalar product grade. This document uses that public description and makes no ISO compliance claim.

## 2. Security: severity and development practices answer different questions

| Measure | Unit, range, direction | Tool and evidence | Applicability and limitation |
| --- | --- | --- | --- |
| CVSS 4.0 | Severity, 0–10; higher means more severe | FIRST calculator or compatible scorer; identified vulnerability, metric rationale, vector, affected revision | Scores a vulnerability's technical severity, not whole-system security or business risk. |
| OpenSSF Scorecard | Individual checks 0–10; higher means stronger performance against that heuristic | Scorecard scan; repository revision, scan date, permissions, check details, tool version | Evaluates selected repository and supply-chain practices. It does not prove code is secure. |

[CVSS 4.0 specification document 1.2, June 18, 2024](https://www.first.org/cvss/v4.0/specification-document) defines severity bands: **0 none; 0.1–3.9 low; 4.0–6.9 medium; 7.0–8.9 high; 9.0–10 critical**. Report the vector and metric-group nomenclature, such as CVSS-B when only Base metrics are explicitly selected. Threat and Environmental metrics can refine circumstances; Supplemental metrics do not change the number. A missing assessment is unknown, not zero severity. Project policy determines blocking findings. CVSS is owned by FIRST.Org, Inc. and used by permission.

[OpenSSF Scorecard](https://github.com/ossf/scorecard), with [v5.5.0](https://github.com/ossf/scorecard/releases/tag/v5.5.0) observed during this research, combines heuristic checks using risk weighting. Keep individual findings even when reporting its aggregate: identical averages can hide very different weaknesses. Its maintainers explicitly describe false positives, false negatives, and contextual limitations. A dependency-pinning finding and a demonstrated authorization bypass should remain separately actionable. Missing permissions or unavailable scan evidence require an explanation; they must not become an invented score of zero.

## 3. Performance and reliability need a workload and a time window

| Measure | Unit and direction | Evidence to retain | Appropriate interpretation |
| --- | --- | --- | --- |
| p50 / p95 / p99 latency | Usually milliseconds; lower is better for equivalent work | Request distribution, sample count, load, payloads, hardware/runtime, warm-up, successes versus failures | Median and tail experience; a percentile without its workload is insufficient for comparison. |
| Throughput | Completed operations/second; higher is usually better under fixed constraints | Equivalent operation definition, concurrency, error rate, resource limits | More rejected or incorrect requests do not establish increased useful throughput. |
| Memory | Named measure, such as peak RSS in bytes; usually lower is better | Measurement scope, process/container limits, peak sampling, workload | Heap size, resident memory, and allocated bytes per operation are different quantities. |
| Error rate / success SLI | Bad/eligible events or good/eligible events, usually %; lower/higher respectively | Definition of eligible and bad events, counts, evaluation window | Measures the selected user outcome. HTTP success alone may miss incorrect results. |
| SLO attainment | Compare a defined service indicator with its target over its window | Production or other contract-specified observations | A load test cannot establish a month-long production reliability objective. |

[Google SRE's Service Level Objectives chapter](https://sre.google/sre-book/service-level-objectives/) (2016) explains why averages conceal latency tails and why services need different objectives. Record measurement location and request classes: client latency and server processing time differ. A p99 estimate from a small sample is especially unstable. Show sample size and repeated-run variation; do not award points for tiny changes smaller than observed variability. The living [Google Benchmark user guide](https://google.github.io/benchmark/user_guide.html) describes warm-up, repetitions, machine information, and memory reporting for reproducible benchmarks.

An SLO such as 99.9% successful eligible events over 30 days is a **project-selected target**, not a universal industry pass mark. Its allowable bad-event fraction is 0.1%. Define the denominator and outcome first, following [Implementing SLOs](https://sre.google/workbook/implementing-slos/) (2018). Report failed-request latency separately where useful: fast failures can make an aggregate latency result look better. [Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/) (2016) also treats explicitly incorrect responses as possible errors.

For browser experiences, Google's **Core Web Vitals** provide published thresholds:

| Metric | Good | Needs improvement | Poor |
| --- | --- | --- | --- |
| Largest Contentful Paint (LCP) | ≤2,500 ms | >2,500–4,000 ms | >4,000 ms |
| Interaction to Next Paint (INP) | ≤200 ms | >200–500 ms | >500 ms |
| Cumulative Layout Shift (CLS) | ≤0.1 | >0.1–0.25 | >0.25 |

Use the **75th percentile of field measurements, separately for mobile and desktop**. All three must meet good thresholds for an overall good assessment. These are Google recommendations for web experience, not general API budgets. A single lab run does not establish field compliance. See [threshold methodology](https://web.dev/articles/defining-core-web-vitals-thresholds), updated May 7, 2025, and [Web Vitals](https://web.dev/articles/vitals), updated October 31, 2024. INP replaced FID on [March 12, 2024](https://web.dev/blog/inp-cwv-march-12).

## 4. Maintainability scores are structural evidence

| Measure | Unit, range, direction | Tool and evidence | What it misses |
| --- | --- | --- | --- |
| Microsoft Maintainability Index (MI) | Index 0–100; higher is considered easier to maintain | Visual Studio Code Metrics; supported code scope, analyzer version, inputs | Architecture, domain clarity, documentation, and actual change effort. |
| Cyclomatic complexity | Independent control-flow paths; generally integer ≥1 per function, lower means fewer paths | Language-specific analyzer; per-function results and counting rules | Difficulty of expressions, meaningful abstractions, and correctness. |
| Sonar Cognitive Complexity | Integer ≥0; higher indicates more control-flow comprehension difficulty under its rules | Sonar analyzer; language, version, scope, issue locations | The entire human experience of understanding or changing code. |
| Sonar technical-debt ratio / rating | Estimated remediation cost divided by estimated development cost, %; lower is better; configured A–E bands | Sonar rule profile, remediation estimates, LOC assumptions and rating grid | Real calendar estimates and architectural or organizational debt outside detected issues. |

Microsoft's implementation is:

```text
MI = max(0, (171 − 5.2 × ln(Halstead Volume)
                   − 0.23 × Cyclomatic Complexity
                   − 16.2 × ln(Lines of Code)) × 100 / 171)
```

Here `ln` is the natural logarithm. Halstead Volume summarizes token/operator/operand structure. Microsoft's bands are **0–9 low, 10–19 moderate, 20–100 good**. These broad vendor bands are not universal acceptance criteria. Other tools use different MI variants; record the formula and supported language. Empty or undefined inputs need the analyzer's documented handling. See [Microsoft's MI explanation](https://learn.microsoft.com/en-us/visualstudio/code-quality/code-metrics-maintainability-index-range-and-meaning?view=visualstudio), updated October 30, 2025.

Cyclomatic complexity commonly starts at one and adds decision branches, but analyzers differ by language. Microsoft's [CA1502 rule](https://learn.microsoft.com/en-us/dotnet/fundamentals/code-analysis/quality-rules/ca1502) uses a configurable default threshold of 25; that is a tool default, not a law that 24 is maintainable. Sonar's [metric definitions, version 2025.3](https://docs.sonarsource.com/sonarqube-server/2025.3/user-guide/code-metrics/metrics-definition) document language-specific counting and estimated technical debt. Debt estimates depend on rule costs and a configurable development-cost assumption; they are not a delivery forecast. G. Ann Campbell's [Cognitive Complexity explanation](https://www.sonarsource.com/resources/cognitive-complexity/) is dated April 5, 2021.

Use these results to locate code for review. Inspect whether splitting a function improves comprehension or merely relocates complexity. Low complexity cannot establish design fit.

## 5. Test adequacy and debuggability need different evidence

Statement coverage measures executed eligible statements; branch coverage measures exercised eligible branches. Both are percentages, with higher values indicating broader execution under the tool's exclusions. Preserve commands, collected tests, artifact revision, exclusion configuration, and coverage report. [Coverage.py branch documentation](https://coverage.readthedocs.io/en/latest/branch.html), displaying version 7.16.1 when checked, explains how missing branches can survive full statement coverage. Neither result establishes that assertions check the intended behavior.

Mutation testing changes code and asks whether tests detect the changes. [Stryker's metric definition](https://stryker-mutator.io/docs/mutation-testing-elements/mutant-states-and-metrics/) uses:

```text
detected = killed + timeout
valid = detected + survived + no coverage
mutation score = 100 × detected / valid
```

The range is 0–100%, higher meaning more selected mutations detected. Compile/runtime-invalid and ignored mutants are outside this denominator; report them and pending work. Zero valid mutants makes the percentage undefined. Some surviving mutations may be behaviorally equivalent, so inspect survivors before concluding that every one reveals a missing assertion. Mutation testing strengthens evidence about test sensitivity; it does not establish requirement completeness.

Debuggability has no universal scalar established by the sources reviewed here. [OpenTelemetry](https://opentelemetry.io/docs/what-is-opentelemetry/), overview updated April 6, 2026, provides instrumentation and conventions for traces, metrics, and logs. It is not a debuggability rating.

A **proposed v-eval rubric**, explicitly local, can award one point for each demonstrated capability in a controlled failure exercise:

1. An actionable error identifies the failed operation.
2. A correlation identifier connects relevant events across components.
3. Available evidence identifies the failing dependency or operation and its cause.
4. A documented, safe reproduction procedure reproduces the failure.

The total is 0–4, higher meaning more capabilities demonstrated. Define checks before comparison. If one cannot be assessed, keep the total `null` and show known outcomes. These equally weighted checks need validation with maintainers; they do not cover every debugging task.

Recovery-time statistics require operational data and defined incident boundaries. They reflect more than logging quality. DORA's [metrics history](https://dora.dev/insights/dora-metrics-history/), January 2, 2026, explains its 2023 move from MTTR to **failed deployment recovery time**. That narrower deployment-failure measure should not be relabeled as all-incident recovery time.

## 6. Documents and generated context need meaning-based criteria

| Measure | Unit and method | Applicability and limitation |
| --- | --- | --- |
| Source-supported factual claims | Supported assessed atomic claims / assessed atomic claims, %; claim-to-passage evidence | Higher is desirable for factual content. Explicitly report unresolved claims and the full claim count; precision does not establish completeness. |
| Citation support | Defined citation correctness and statement-support measures, with human or model entailment checks | Working links alone prove neither relevance nor support; automated judges can err. |
| Required-context completeness | Supported required information items present / fixed required items, % | Proposed task-specific measure; requires an independently defined checklist. It is not a universal context index. |
| Readability | Named formula and language; audience-dependent direction | Word/sentence structure is a proxy for ease, not truth, usefulness, or faithful compression. |

[FActScore](https://aclanthology.org/2023.emnlp-main.741/) and [ALCE](https://aclanthology.org/2023.emnlp-main.398/), both December 2023, motivate atomic factual support and citation evaluation. The operational measures above must disclose their definitions rather than claim compatibility with a benchmark implementation. A short answer can support every claim while omitting the most important requirement.

Microsoft documents the English Flesch Reading Ease formula as `206.835 − 1.015 × words/sentences − 84.6 × syllables/words`. Higher generally means easier reading; the formula can produce values outside its conventional 0–100 presentation. [Word's readability documentation](https://support.microsoft.com/en-US/Word/get-your-document-s-readability-and-level-statistics-in-microsoft-word) gives audience-oriented guidance, not a technical-document acceptance standard. Technical vocabulary can lower a score while making a specialist explanation more precise. Preserve language and tokenizer/syllable-counting method.

## 7. Read a worked scorecard without inventing certainty

**Everything in this example is fictional. No tools were run and no vulnerability was found.** Imagine an AI-generated export API change, tests, and a handoff document. The contract requires safe authorization, defined load-test budgets, test adequacy, useful diagnostic evidence, and preservation of five specified facts. Its workload is 10,000 eligible requests, a fixed payload mix, concurrency 32, and a declared 2-vCPU/4-GiB environment. Every number would require revision-bound evidence in a real report.

| Dimension and measure | Synthetic result | Threshold origin and decision | Reason and next action |
| --- | --- | --- | --- |
| Security: CVSS-B | 9.3, critical | Project requires no demonstrated critical vulnerability: **FAIL** | Fictional authorization bypass; fix access control and rerun the exploit regression. |
| Supply chain: dependency-pinning check | 7/10 | Advisory OpenSSF heuristic | Inspect unpinned dependencies; this does not dilute the vulnerability finding. |
| Performance: latency | p50 80 / p95 420 / p99 950 ms | Project p95 ≤300 ms: **FAIL** | Investigate tail latency under the same workload. |
| Performance: throughput / peak RSS | 56 requests/s; 180 MiB | Project ≥40 requests/s and ≤256 MiB: **PASS** | Passing resource checks do not erase the latency failure. |
| Reliability: load-test errors | 23/10,000 = 0.23% | Project ≤0.1%: **FAIL** | Inspect failed outcomes; this is not a production SLO measurement. |
| Reliability: 30-day production SLI | `null` | Required 99.9%; **UNKNOWN** | No production-window evidence exists. |
| Maintainability: MI / maximum method complexity | 42/100; 18 paths | MI advisory; project complexity ≤20: **PASS** | MI falls in Microsoft's good band; inspect design fit separately. |
| Tests: branch coverage / mutation score | 90/100 = 90%; 60/80 = 75% | Project ≥85% / ≥80%: **PASS / FAIL** | Exercise uncovered intent and inspect surviving mutations. |
| Debuggability: local four-check rubric | 2/4 | Project ≥3: **FAIL** | Operation and correlation are clear; cause and reproduction are missing. |
| Handoff context: completeness | 4/5 = 80% | Project requires all five: **FAIL** | Restore the omitted compatibility constraint with source evidence. |
| Web Vitals | `null` | **NOT_APPLICABLE** | This contract has no browser interface. |

The fictional CVSS value uses the verified 9.3 example in [FIRST FAQ v1.10, July 17, 2026](https://www.first.org/cvss/v4.0/faq):

```text
CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N
```

The mutation result means 60 detected mutants out of 80 valid mutants, for example 54 killed plus six timeouts. A real report must also show invalid, ignored, and pending counts. The completeness result requires identifying all five expected information items and the missing one.

Overall acceptance is **FAIL**, following the proposed [v-eval policy](design.md): any required applicable failure blocks acceptance. If there were no failures but production evidence remained unknown, the result would be **INCOMPLETE**. Numeric averages cannot override these rules. An operational tool failure is **ERROR**, missing evidence is **UNKNOWN**, and a documented exclusion is **NOT_APPLICABLE**; their numeric values remain `null`, with different reasons.

Retain definition/version, artifact scope, raw value and denominator, unit/range/direction, tool configuration, evidence, threshold origin, and verdict. Keep native scales. Any future composite must disclose normalization and weights, undergo validation against developer decisions, and preserve required-failure vetoes.

To practice, change the workload or audience. Identify incomparable measurements, thresholds needing adjustment, and missing evidence. Have two reviewers independently apply the debuggability and context checklists. Their disagreements reveal unclear anchors.
