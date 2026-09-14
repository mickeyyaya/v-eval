# How an evaluator can learn from human feedback, locally and safely

Research date and access date: 2026-09-14.

Method: Web-research agent. 24 source rows; six read in full.

Facts, inferences, single-source claims, and unverified items are labeled inside the memo; a label applies to the sentence or row it sits in.

---

Date 2026-09-14. All URLs accessed 2026-09-14. Scope: per-criterion PASS/FAIL/UNKNOWN/ERROR/NOT_APPLICABLE judge, local-first, ollama-class models must work. **F** = fact from a source; **I** = my inference. SINGLE = one source only; UNVERIFIED = number not seen in the primary text.

## 0. Bottom line

- **F** Every shipped "align the judge" product (LangSmith, Braintrust, AlignEval, Husain/Shankar) uses one loop: human binary labels → prompt edits / few-shot examples → TPR/TNR on a held-out split. None fine-tunes.
- **F** 2026 self-improvement papers agree on one rule: the learner may change the grading *function*, never the *locked anchor set*, which selection must never read (Amazon, Regimes, Li).
- **F** 7–8B judges hit 87–89% on closed-ended reference-backed checks but are near random on hard code pairs, default to PASS ("truth bias"), and swing 10–25 pp under ordering/persona. **I** Calibrated abstention (UNKNOWN) is the cheapest accuracy lever on small models.

## 1. Sources

| # | Source | Establishes | Does not / caveat |
|---|---|---|---|
| 1 | Husain & Shankar, *Evals FAQ*, upd. 2026-09-01, <https://hamel.dev/blog/posts/evals-faq/> | 100–200 labels per failure mode; train 10–20% / dev 40–45% / test 40–45%; 30–50 Pass + 30–50 Fail per set; binary not Likert; TPR/TNR on held-out; "start with the most capable models… optimize for cost later"; criteria drift. | No measured gains. |
| 2 | Husain, *validate-evaluator* skill, <https://www.skills.sh/hamelsmu/evals-skills/validate-evaluator> | "~100 traces with binary Pass/Fail labels per failure mode"; "Iterate… until TPR and TNR > 90% on dev"; then test; bias-correct production. | Formula not in fetched page. |
| 3 | LangChain, *Align Evals*, 2025-07-29, <https://www.langchain.com/blog/introducing-align-evals> | Human per-criterion score; manual prompt edits ("add clearer negative criteria"); saved baseline alignment score per prompt version. | No sizes or gains. |
| 4 | Yan, *AlignEval*, 2024-10, <https://eugeneyan.com/writing/aligneval/> | Binary labels; 20 unlock eval, 50 unlock optimize; dev/test; optimizer trials on dev for F1: 0.571→0.722 dev, 0.727 test; warns dev overfit. | One example; optimizer undisclosed. |
| 5 | Braintrust, *Human review → golden datasets*, 2026-05-21, <https://www.braintrust.dev/blog/human-review-golden-datasets> | Signals: pass/fail, categorical, slider, rationale, corrected `expected`; promote reviewed items to golden set; periodic calibration queue. | No sizes/targets. |
| 6 | Shankar et al., *Who Validates the Validators* (UIST 2024), <https://arxiv.org/abs/2404.12272> | EvalGen; 9 practitioners; criteria drift: grading defines the criteria. | Qualitative. |
| 7 | Cursor, *Bugbot learned rules*, 2026-04-08, <https://cursor.com/blog/bugbot-learning> | Signals: downvote, explanatory reply, human comment on a miss; candidate rule → evaluated on incoming PRs → active on accumulated signal; auto-disable on negative signal; per-repo; editable. 44k rules. | "52%→78.13% resolution" is a product metric. SINGLE. |
| 8 | CodeRabbit *Learnings* docs, <https://docs.coderabbit.ai/knowledge-base/learnings> ; Graphite docs, <https://graphite.com/docs/ai-review-customization> | Learnings = text + PR/file/user metadata; scopes local/global/auto; "Path instructions precede learnings"; delete contradictions. Graphite: plain-English rules, up/downvote tracked. | No dedup/limits; Graphite has no documented auto-learning. |
| 9 | Li, *Who Drifted: the System or the Judge?*, 2026-06-13, <https://arxiv.org/abs/2606.15474> | Frozen human anchor set (k=200, re-scored 1 in 5 items), excluded from calibration; only judge drift moves anchors; 60/60 version bumps caught, 110/120 prompt changes attributed; rolling z-test false-alarms 75%. | Single-author preprint; no leak analysis. |
| 10 | *judge-drift-sentinel*, 2026, <https://github.com/homayoun-safarpour/judge-drift-sentinel> | "Label 10-50 representative outputs once"; JSONL anchors; Cohen's kappa baseline vs current → JUDGE_DRIFT / SYSTEM_CHANGE / STABLE; stdlib-only. | Hobby tool. SINGLE. |
| 11 | Zhang et al. (AWS), *Who Grades the Grader?*, 2026-07-14, <https://arxiv.org/html/2607.12790v1> | 10-item anchored dev set is the only supervision; locked test set "never read by any loop"; fail-closed anchoring + validity gate; without them a "vacuous always-pass grader"; "tens of binary tasks carry per-round noise near five points". | 3 tasks, 1 solver. |
| 12 | Nakajima, *Regimes*, 2026-06-08, <https://arxiv.org/pdf/2606.10241> | Ladder: static checks → sandbox → in-sample → held-out; learner edits only named "action seams"; in-sample +0.18 collapsed to held-out +0.04; held-out +0.05–0.10 on 4/5 splits; McNemar; event-sourced audit log. | Single author; one task. |
| 13 | Shao et al., *Your Agent May Misevolve* (ICLR 2026), <https://arxiv.org/abs/2509.26354> | Memory-evolving coding agent lost 45% refusal rate; memory learns "refund ↔ positive feedback". | Agents, not judges. |
| 14 | Sharma et al., *Sycophancy* (ICLR 2024), <https://arxiv.org/abs/2310.13548> | Humans and preference models prefer convincing sycophantic answers "a non-negligible fraction of the time". | 2023 models. |
| 15 | Kim & Khashabi, *Sycophancy Under User Rebuttal* (EMNLP 2025 Findings), <https://arxiv.org/abs/2509.16533> | Follow-up rebuttal vs simultaneous: Llama-3.3-70B flips 86.0% vs 56.5%; GPT-4.1 36.2% vs 26.5%; avg flip: reasoned rebuttal 56.1%, answer-only 24.1%, casual "sure?" 84.5%. | MCQ; no ≤8B models. |
| 16 | Kapetanovic et al., *Anchoring Bias in LLM-as-a-Judge* (CIKM 2026), <https://arxiv.org/html/2608.25869> | Prior score in context blocks 48% of error corrections, flips 10.18% of correct labels; CoT and warnings don't fix. | Code review weakest category. |
| 17 | Lee et al., *How to Correctly Report LLM-as-a-Judge Evaluations*, 2025-11, <https://arxiv.org/abs/2511.21140> | Rogan–Gladen correction from a human calibration set; CI covers test + calibration uncertainty; unbiased under shift; Python plug-in. | Sizes per CI width not extracted. |
| 18 | Badshah et al., *Judge, Retrieve, or Abstain*, 2026-08-18, <https://arxiv.org/html/2608.17994> | Qwen3-4B/8B/14B, Llama-3.1-8B; Clopper–Pearson thresholds bound error among non-abstained ≤ α; Qwen3-8B α=0.20 coverage 59%→100%. | Retrieves web evidence, not past labels. |
| 19 | Jwa et al., *Becoming Experienced Judges*, 2025-12-09, <https://arxiv.org/abs/2512.06751> | Label-free evolving meta-prompt; gpt-4.1 0.629→0.745; "less effective for relatively weak models". | No ≤8B judge. |
| 20 | Laddha et al., *SLMJury*, 2026-06, <https://arxiv.org/html/2606.07810> | 16 SLMs 0.6–14B; closed-ended: Phi-4 89.55, Qwen3-8B 88.96, Llama-3.1-8B 86.79, Qwen2.5-7B 77.45; ensembles +0.06; debate hurts; Llama-3.1-8B −10.2 pp under lenient persona. | No code tasks. |
| 21 | Jiang et al., *CodeJudgeBench*, 2025-07, <https://arxiv.org/abs/2507.10535> | Thinking Qwen3-8B beats Prometheus-14B and Self-Taught-70B; non-thinking judges <60% vs 50% random; pairwise > pointwise; order-sensitive. | Exact Qwen3-8B score UNVERIFIED. |
| 22 | Dussert, *govllm*, 2026-05-23, <https://arxiv.org/abs/2605.24737> | Ollama-tag judges, 49 items: agreement 51.5% (mistral:7b)–69.1% (phi4-mini); 5 few-shot/criterion: gemma3:4b +11.8 pp, phi4-mini +8.3, mistral:7b +5.1, qwen3:1.7b −0.2; position bias to −25 pp; truth bias; reason/verdict dissociation. | n≈44. SINGLE. |
| 23 | Kim et al., *Prometheus 2*, 2024-05, <https://arxiv.org/abs/2405.01535> ; Wang et al., *Self-Taught Evaluators*, 2024-08, <https://arxiv.org/abs/2408.02666> | 7B keeps ≥80% of 8x7B, 0.6–0.7 Pearson w/ GPT-4, 16 GB VRAM; synthetic training lifts Llama3-70B 75.4→88.3 RewardBench. | Weak on code (#21); 70B not laptop-class. |
| 24 | van den Burg et al., 2025-02, <https://arxiv.org/abs/2502.04997> ; Y. Li, *Calibrate, Don't Curate*, 2026-05, <https://arxiv.org/html/2605.09702> | Linear map from judge output to human labels, few calibration examples, +142% agreement, lets small models match large; keep weak judges if calibratable. | Rating tasks; panel setting. |

## 2. Findings by sub-question

**Q1 products.** F: signal is a binary/numeric label plus optional rationale and corrected expected value (#1,#3,#5). F: the judge changes only via prompt edits and few-shot examples from the *train* split (#1,#2). F: ~100–200 labels per failure mode; only AlignEval shows a delta (#4). I: no product persists an override as a rule; it becomes a labeled example.

**Q2 research.** F: label-free self-improvement needs strong models (#19,#23). F: fine-tuned small judges lose to thinking 8B on code (#21). F: calibrated abstention bounds error on small models (#18); a linear post-hoc map aligns a small judge (#24).

**Q3 retrieval.** F: 5 annotated examples per criterion lifted 3 of 4 ollama-class judges 5–12 pp and hurt one (#22). F: a visible prior score anchors the judge (#16). I: retrieve labeled past cases as examples but strip prior verdict/score text, never retrieve from the anchor set, dedupe by content hash plus near-dup similarity. No source measured near-dup control.

**Q4 rules without a model.** F: Bugbot candidate→active ladder with auto-disable (#7); CodeRabbit precedence "path instructions precede learnings" and delete-contradictions (#8). I: v-eval precedence = contract > project rule > learned example, with a contradiction check at write time.

**Q5 guardrails.** F: preference data rewards sycophancy (#14); a follow-up rebuttal flips verdicts far more than simultaneous judging, casual pushback most (#15); memories learn what pleased the user (#13); the learner never touches the locked set (#11,#12). I: a correction must be re-judged *cold* (fresh context, no prior verdict, no user text) before becoming an example.

**Q6 measurement.** F: 100–200 balanced labels per failure mode (#1); tens of binary items ⇒ ±5-point noise (#11); McNemar for paired before/after (#12); Rogan–Gladen CIs (#17); frozen anchors re-scored periodically isolate judge drift (#9,#10). I: Wilson intervals on TPR/TNR at n=50 are ≈±11 pp; treat sub-10-pp changes as unproven.

## 3. (a) Feedback signals v-eval could capture

| Signal | Store as | May safely change | Must not change |
|---|---|---|---|
| Question ("why did you say X?") | Query log | Nothing automatic; flag criterion for clarification | Verdicts, examples |
| Correction with rationale | Labeled case {input, criterion, label, rationale, split} | Few-shot pool (train split only) after cold re-judge; dev/test | Contract, anchors |
| Override, no rationale | Labeled case `override_no_reason` | Dev/test counts only; excluded from examples (#15) | Examples, thresholds |
| Added criterion | Candidate project rule | Active after ≥N agreeing labels and no contradiction | Contract |
| Repeated NOT_APPLICABLE | Path-glob hint | Project rule scoped by path | Contract |
| Disagreement rate per criterion | Calibration stats | Abstention threshold (#18) | Labels |

## 4. (b) Minimal safe loop, v1

1. Freeze the contract (criteria semantics, verdict enum, acceptance rules) as a hashed read-only file. [G: learner cannot edit contract]
2. Bootstrap ≥50 labeled cases per criterion; assign train/dev/test 20/40/40 deterministically by hash at creation. [G: split fixed at birth]
3. Lock the test split plus 10–50 extra items as `anchors.jsonl`; learner and retrieval never read it. [G: locked anchors]
4. Log every correction/override to `feedback.sqlite` with rationale, model id, prompt hash, timestamp.
5. Cold re-judge each correction (fresh context, no prior verdict, no user text); store the disagreement. [G: anti-sycophancy]
6. Promote to the few-shot pool only if train split, has rationale, passes content-hash + near-dup check. [G: no leakage, no clones]
7. Added criteria become candidate rules; need ≥3 agreeing labels and a contradiction check to go active; auto-disable on 2 later downvotes (#7). [G: promotion ladder]
8. At judge time retrieve ≤5 train-split examples per criterion with score text stripped (#16); small models get examples only if step 10 showed a gain for that model.
9. Per criterion, calibrate an abstention threshold on dev so error among non-abstained ≤ α (Clopper–Pearson) (#18); below it emit UNKNOWN. [G: fail-closed]
10. Before any promotion, re-score anchors and test split; accept only if TPR and TNR do not drop and paired McNemar is not adverse; append-only log. [G: held-out gate, auditable]

## 5. (c) Ollama-class models

- **Can:** closed-ended reference-backed correctness at 87–89% (Qwen3-8B 88.96, Llama-3.1-8B 86.79; #20); bounded error with abstention (#18); +5–12 pp from 5 few-shot examples on 3 of 4 models (#22); post-hoc linear calibration (#24).
- **Cannot:** judge hard code-correctness pairs (non-thinking <60% vs 50% random; #21); resist ordering/persona shifts (−10 to −25 pp; #20,#22); self-improve label-free (#19); avoid truth bias and reason/verdict dissociation without forced structured output (#22).
- **I:** use thinking mode, pairwise where possible, forced `{reason, verdict}` JSON, frequent UNKNOWN, and per-model measurement since few-shot can hurt (#22).

## 6. (d) Gaps

- No source measures retrieval of past labeled cases *from the same repo* as judge examples; #22 is closest (n≈44, SINGLE).
- No published TPR/TNR before/after for any product loop; Bugbot's 78% is a resolution metric.
- Exact Qwen3-8B code-judge accuracy (#21) UNVERIFIED here.
- No paper on near-duplicate control for example pools or contradiction detection between learned rules.
- Anchor leakage into prompts is unaddressed by #9/#11 beyond "never read by any loop".
- Rebuttal sycophancy (#15) measured on ≥70B MCQ models; small-model code-review behaviour unmeasured.
