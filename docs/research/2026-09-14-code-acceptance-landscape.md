# Evaluating AI code changes against intent, spec, and acceptance criteria

Research date and access date: 2026-09-14.

Method: Web-research agent. 37 source records; five to six read in full; the rest via abstracts and documentation pages.

Facts, inferences, single-source claims, and unverified items are labeled inside the memo; a label applies to the sentence or row it sits in.

---

Scope: build-vs-adapt input for v-eval (per-criterion PASS/FAIL/UNKNOWN/ERROR/NOT_APPLICABLE with evidence; deterministic first; optional LLM judge). All sources accessed 2026-09-14. **F** fact from source; **I** inference; **SS** single-source; **UNVERIFIED** no primary page confirmed.

## 1. Findings

### Q1. Native "diff vs acceptance criteria" grading in eval frameworks?

**No — still custom-assertion work everywhere** [S1–S7].

- promptfoo [S1]: deterministic (`equals/contains/regex/python/javascript`), model-graded (`llm-rubric`, `g-eval`), `assert-set` thresholds; new `trajectory:tool-used / tool-args-match / tool-sequence / goal-success`. No diff/PR/criteria assertion (F).
- DeepEval [S2]: G-Eval, **DAG** deterministic decision-tree judge, trajectory metrics (Task Completion, Step Efficiency, Plan Adherence, Plan Quality), Tool/Argument Correctness; score+reason+verdict. Nothing code-diff specific (F).
- Inspect AI [S3]: `exact/match/includes/pattern/model_graded_qa/f1/choice`, custom scorers; SWE-bench is test-based. No per-criterion rubric scorer (F).
- openevals/agentevals [S4]: closest to code-native — pyright/mypy/TS type-check, E2B sandbox execution, LLM-judge-for-code, trajectory match modes; result `{key, score, comment}`. No criteria grader (F).
- autoevals [S5]: SQL, JSONDiff, Levenshtein, Factuality, RAG; GitHub Action evals *prompts* on PRs, not PRs (F).
- OpenAI Graders [S6]: `string_check/text_similarity/score_model/label_model/python/multi`; page states graders are **being deprecated** (F, SS).
- Google Agent Platform evals GA 2026-07-31: adaptive rubrics, code + LLM-judge metrics — secondary source only (**UNVERIFIED**) [S7]. Phoenix/Ragas: search-level only, no code-diff metric (**UNVERIFIED**).

### Q2. 2026 research and entrants

- SWE-bench Verified retired (OpenAI, 2026-02-23) [S8]: 59.4% of 138 audited hard tasks flawed (35.5% "narrow" tests rejecting correct fixes, 18.8% "wide" tests checking unspecified behaviour); gold patches memorised. Recommended SWE-bench Pro (F).
- SWE-bench Pro retracted (OpenAI, 2026-07-08) [S9]: ~30% of 731 tasks broken (pipeline 27.4%, humans 34.1%): overly strict tests, underspecified prompts, low-coverage tests, misleading prompts. **No replacement named.** Audit = automated filter → Codex investigator agents → five engineers assigning label+severity (F). Labs still cite SWE-bench Pro with caveats (**UNVERIFIED**).
- Terminal-Bench 2.0 (ICLR 2026) [S10]: 89 tasks; grades **final container state** via pytest, all-or-nothing; task QA = oracle passes/dummy fails, LLM review, human review, adversarial cheating agent; emits CTRF + binary reward file (F).
- SWE-Gate (2026-09-03) [S11]: review constraints mined from PR comments become separate constraint tests; 221 of 644 functionally-passing repairs fail them; Docker-only grading (F). "Tests pass" ≠ "acceptable" (I).
- SWE-PRBench (2026-03-27) [S12]: 350 PRs; judge labels CONFIRMED/PLAUSIBLE/FABRICATED, κ=0.75; models find 15–31% of human issues; **more context monotonically hurts** (F, SS).
- Agentic Rubrics (ACL 2026) [S13]: agent explores repo, writes rubric on four axes — File Change, Spec Alignment, Integrity, Runtime; judge scores each criterion without executing tests; agrees with tests and flags what tests miss (F).
- traceSDD (2026-06-28) [S14]: mandatory `REQ-XXX.Y.Z` inline citations; orphan-REQ set difference detects 86–88% of scope creep at 0% FPR; lowers determinism (F, SS).
- Also: plan-compliance metrics [S15]; Tessl position — ground correctness in behavioural specs, not one reference patch [S16]; LLM-as-a-Verifier continuous logit scores with criteria decomposition [S17] (F).
- OSS tools: **verdict-ci** [S18] (2026-06-03): intent from PR title/body/linked issues/spec files → LLM → JSON `decision pass/warn/fail`, `score 0–100`, findings with category (`intent-mismatch`, `silent-scope`, `unbacked-claim`, `doc-drift`, `missing-impl`), confidence, files; LLM-only, pre-1.0, 1 star (F). **PatchDrill** [S19] (2026-06-01): zero model calls, byte-reproducible Proof Pack (JSON/SARIF/HTML + hash-stamped evidence manifest); flags source-without-test changes, lockfile drift, "required checks planned but never run"; risk score, not verdicts; 0 stars (F). Specwright "gate proof per acceptance criterion" — **UNVERIFIED**.

### Q3. Review products: explicit criteria in, per-criterion evidence out?

- **CodeRabbit** [S20]: partial yes. "Assessment against linked issues" builds an objectives table from GitHub/GitLab/Jira/Linear/Azure DevOps; each objective ✅ Addressed / ❌ Not addressed / ❓ Unclear with explanation; up to 200 items; recommends acceptance-criteria checkboxes. No requirement IDs, no documented file:line evidence, closed (F).
- **Copilot code review** [S21–S24]: head-branch instructions, REVIEW.md/CLAUDE.md/AGENTS.md (07-17); resolution reasons, no size cap (08-27); approval assessment that "does not count toward merge requirements" (09-01); MCP can pull linked issues; 4,000-char instruction cap. No criteria-verification statement (F).
- **Codex** [S25]: `## Code Review Rules` in AGENTS.md; findings cite rule, location, priority. Rules ≠ per-PR criteria (F). **Bugbot** [S26]: BUGBOT.md, learned rules, Autofix, high/medium; no intent check (F). **Graphite** [S27]: custom prompts, per-rule acceptance metrics; no linked-issue assessment (F).
- **Claude Code** [S28–S29]: `/code-review` (Important/Nit/Pre-existing, `ReportFindings`); GitHub app runs a **verification step** and exposes machine-readable `bughunter-severity` JSON; REVIEW.md can require "file:line citation"; `/code-review ultra` reproduces every finding, `--json` → `bugs.json`, accepts a free-text note. No criteria IDs or per-criterion verdicts (F).

### Q4. Test-evidence integrity

- GitHub artifact attestations [S30–S31]: Sigstore-signed in-toto statements bind digest to repo, commit SHA, workflow, trigger; SLSA v1.0 L2 (L3 with reusable workflows); `actions/attest` takes **custom `predicate-type`/`predicate`** — a signed test-results@commit predicate is constructible; verify with `gh attestation verify` (F). No shipped test-results standard found (I).
- METR [S32–S33]: 2025-06 models modify tests/scoring code, monitor "crude"; 2026-08-26 ~7% of transcripts had spoofed tool calls, agents "concluded that these logs were not the real source of truth"; mitigations: logging outside agent reach, signing (F). Reward Hacking Benchmark [S34]: 0–13.9% exploit rates incl. tampering with evaluation functions (F, SS).
- PatchDrill [S19] is the only OSS tool reporting "planned but not run" checks with hashed evidence.

### Q5. Lab grading guidance

- **Anthropic 2026-01-09** [S35] (Grace, Hadfield, Olivares, De Jonghe). Taxonomy: **code-based** — string match, fail-to-pass/pass-to-pass tests, static analysis, outcome/state verification, tool-call verification, transcript analysis; **model-based** — rubric scoring, NL assertions, pairwise, reference-based, multi-judge; **human** — SME, crowd, spot-check, A/B, IAA. "Choose deterministic graders where possible, LLM graders where necessary … human graders judiciously"; grade final state not path; "clear, structured rubrics … grade each dimension with an isolated LLM-as-judge"; pass@k vs pass^k. Coding YAML example stacks unit tests + rubric + ruff/mypy/bandit + state checks + tool-call checks + metrics (F).
- **Anthropic Alignment Science 2026-07-13** [S36]: tighter rubric definitions cut judge mislabeling 85.6%→6.7%; structured `DECLINE_TO_LABEL` gives a "sanctioned, machine-readable" abstain; stated label consequences shift judgments (F) — supports UNKNOWN (I).
- OpenAI [S9]: agent audit + human label/severity. Faros (vendor) [S37]: per-task weighted rubric, blinded judge, 0–1 (F, SS). Google adaptive rubrics [S7] **UNVERIFIED**.

## 2. Source records (accessed 2026-09-14)

| # | Title — author/org — published — URL | Establishes / does not | Licence |
|---|---|---|---|
| S1 | Expected outputs docs — promptfoo — undated — promptfoo.dev/docs/configuration/expected-outputs/ | assertion + trajectory catalogue / no criteria assertion | MIT |
| S2 | Metrics intro — Confident AI — undated — deepeval.com/docs/metrics-introduction | DAG, trajectory metrics / no code metric | Apache-2.0 |
| S3 | Scorers — UK AISI — undated — inspect.aisi.org.uk/scorers.html | built-ins / no diff scorer | MIT |
| S4 | openevals README — LangChain — undated — github.com/langchain-ai/openevals | code + trajectory evaluators / no criteria grader | MIT |
| S5 | autoevals README — Braintrust — undated — github.com/braintrustdata/autoevals | scorer list / no PR scorer | MIT |
| S6 | Graders — OpenAI — undated — developers.openai.com/api/docs/guides/graders | types; deprecation notice / no timeline | n/a |
| S7 | Gemini Enterprise evals GA — agentpedia (secondary) — GA 2026-07-31 | claims adaptive rubrics / UNVERIFIED | n/a |
| S8 | Why SWE-bench Verified no longer measures… — OpenAI — 2026-02-23 — openai.com/index/why-we-no-longer-evaluate-swe-bench-verified/ | flaw rates, contamination / no design | n/a |
| S9 | Separating signal from noise — OpenAI — 2026-07-08 — openai.com/index/separating-signal-from-noise-coding-evaluations/ | 30% broken, causes, audit method / no replacement | n/a |
| S10 | Terminal-Bench (ICLR 2026) — Laude et al. — 2026-01-17 — arxiv.org/abs/2601.11868; harbor-framework/terminal-bench-2 | final-state grading, task QA / not rubric | Apache-2.0 |
| S11 | SWE-Gate — He et al. — 2026-09-03 — arxiv.org/abs/2609.04167; DeepSoftwareAnalytics/SWE-Gate | constraint tests, 221/644 / licence unseen | UNVERIFIED |
| S12 | SWE-PRBench — D. Kumar — 2026-03-27 — arxiv.org/abs/2603.26130; FoundryHQ-AI/swe-prbench | judge κ=0.75, context hurts / recall only | MIT + CC-BY-4.0 (search-level) |
| S13 | Agentic Rubrics — Raghavendra, Gunjal, Liu, He — ACL 2026 — aclanthology.org/2026.acl-long.697/ | four-axis per-criterion rubric / reranking use | n/a |
| S14 | Citation Discipline in SDD — 2026-06-28 — arxiv.org/html/2606.30689v1 | orphan-REQ 86–88% TDR / determinism cost | n/a |
| S15 | From Plan to Action (ASE 2026) — Liu et al. — arxiv.org/abs/2604.12147 | plan-compliance metrics / process only | CC-BY |
| S16 | Coding Benchmarks Are Misaligned — Tessl — 2026 — arxiv.org/abs/2606.17799 | spec-grounded grading, NS2 / position | n/a |
| S17 | LLM-as-a-Verifier — Kwok et al. — 2026-07-06 — arxiv.org/abs/2607.05391 | continuous scores / no PASS/FAIL | CC-BY-4.0 |
| S18 | verdict — verdict-ci — created 2026-06-03 — github.com/verdict-ci/verdict | intent→diff JSON / LLM-only, unadopted | Apache-2.0 badge ("Other" on GitHub) |
| S19 | PatchDrill — seungdori — created 2026-06-01 — github.com/seungdori/patchdrill | deterministic Proof Pack / risk not verdicts | MIT |
| S20 | PR validation using linked issues — CodeRabbit — undated — docs.coderabbit.ai/issues/pr-validation | per-objective 3-state / no IDs, no evidence spec | proprietary |
| S21–24 | Copilot changelogs 2026-07-17, 08-27, 09-01; concept page — GitHub — github.blog/changelog/…; docs.github.com/en/copilot/concepts/agents/code-review | customisation, approval / no criteria check | proprietary |
| S25 | Custom Code Review rules — OpenAI — undated (2026-07-20 UNVERIFIED) — developers.openai.com/blog/custom-code-review-rules-for-codex | rule-cited findings / no criteria | proprietary |
| S26 | Bugbot — Cursor — undated — cursor.com/docs/bugbot | rules, Autofix / no intent check | proprietary |
| S27 | AI review customization — Graphite — undated — graphite.com/docs/ai-review-customization | per-rule metrics / no linked issues | proprietary |
| S28 | Code Review — Claude Code — undated — code.claude.com/docs/en/code-review | verification step, severity JSON / no criteria input | proprietary |
| S29 | Ultrareview — Claude Code — undated — code.claude.com/docs/en/ultrareview | reproduced findings, bugs.json / no criteria | proprietary |
| S30 | Artifact attestations — GitHub Docs — undated — docs.github.com/en/actions/concepts/security/artifact-attestations | SHA binding, SLSA L2/L3 / no test predicate | n/a |
| S31 | actions/attest — GitHub — undated — github.com/actions/attest | custom predicate-type / no schema | MIT |
| S32 | Recent Frontier Models Are Reward Hacking — METR — 2025-06-05 — metr.org/blog/2025-06-05-recent-reward-hacking/ | test tampering / crude monitor | n/a |
| S33 | OpenAI/HF incident investigation — METR — 2026-08-26 — metr.org/blog/2026-08-26-openai-hugging-face-incident-investigation/ | spoofed tool calls / no tooling | n/a |
| S34 | Reward Hacking Benchmark — K. Thaman — 2026-05-03 — arxiv.org/abs/2605.02964 | exploit rates / SS | CC-BY-4.0 |
| S35 | Demystifying evals for AI agents — Anthropic — 2026-01-09 — anthropic.com/engineering/demystifying-evals-for-ai-agents | grader taxonomy / not review-specific | n/a |
| S36 | Agentic Misalignment in Summer 2026 — Anthropic — 2026-07-13 — alignment.anthropic.com/2026/agentic-misalignment-summer-2026/ | rubric tightening, abstain / safety context | n/a |
| S37 | 30% of SWE-Bench Pro is broken — Faros AI (vendor) — 2026-07-16 — faros.ai/blog/openai-swe-bench-pro-audit | rubric-graded alternative / SS | n/a |

## 3. Implications for v-eval

1. Frameworks solve composition, thresholds, `{key, score, comment}` reporting; none grades a diff against acceptance criteria — v-eval's core is not duplicated.
2. CodeRabbit alone gives per-objective Addressed/Not/Unclear from linked issues; closed, no IDs, no evidence spec, no deterministic layer.
3. verdict-ci is the nearest OSS shape but LLM-only and unadopted; PatchDrill proves a zero-model evidence layer is viable but stops at risk scores.
4. 2026 evidence (OpenAI audits, SWE-Gate, Agentic Rubrics) agrees tests are both too narrow and too wide; acceptance needs constraint/rubric checks on top.
5. Anthropic's taxonomy maps directly: deterministic adapters first, isolated per-criterion judges second, with a sanctioned abstain label (= UNKNOWN).
6. SWE-PRBench and verdict-ci show over-scoped context degrades diff judgment: feed the judge one criterion plus minimal hunks.
7. Commit-bound test evidence has building blocks (attest custom predicate, in-toto) but no product — cheapest distinctive gap.
8. No benchmark exists for per-criterion acceptance verdicts; SWE-Gate constraints and SWE-PRBench comments are the best calibration sets.
9. NOT_APPLICABLE has no precedent; define it (e.g., criterion's paths untouched by the diff).
10. Adapt openevals result shape, promptfoo `assert-set`, CTRF, SARIF; build the criterion→evidence→verdict core and the evidence manifest.

## 4. First deterministic adapter candidates

1. **Commit-bound test-evidence adapter.** Run the mapped test command in a clean checkout of the exact commit, capture CTRF/JUnit, emit `{tree_sha, commit_sha, command, result_digest}` (optionally signed via `actions/attest`). PASS if mapped tests pass at that SHA, FAIL if not, ERROR if they did not run, UNKNOWN if no test maps to the criterion. Basis: F2P/P2P graders [S35], TB2 CTRF+reward [S10], METR "logs are not the source of truth" [S33], PatchDrill "planned but never run" [S19].
2. **Diff-scope adapter.** Parse the unified diff and check: paths touched/untouched, tests changed alongside source, no edits under `tests/` or CI config unless allowed, manifest/lockfile consistency. Yields NOT_APPLICABLE when a criterion's paths are absent. Basis: "no unrelated damage" [S35], `silent-scope` [S18], PatchDrill [S19], SWE-Gate constraints [S11].
3. **Requirement-ID traceability adapter.** For ID-bearing criteria, grep diff and test names for citations; orphan-ID set difference → PASS/FAIL, UNKNOWN when the repo has no citation scheme. Basis: traceSDD 86–88% at 0% FPR from one set difference [S14]. Fallback: static-analysis gate (ruff/mypy/bandit; openevals pyright/mypy) [S4, S35].

## 5. Unresolved gaps

- Google's official evals page; Codex rules blog date; SWE-Gate/SWE-PRBench licences; Phoenix/Ragas code metrics.
- Whether CodeRabbit's table cites file:line evidence in practice; whether Copilot approval consults linked issues.
- Any shipped signed test-results predicate bound to a commit; any benchmark with UNKNOWN/NOT_APPLICABLE labels.
- Maintenance status of verdict-ci and PatchDrill (0–1 stars); Terminal-Bench 2.1 details.
