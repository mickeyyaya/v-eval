# Local prior art on the maintainer's machine

Research date and access date: 2026-09-14.

Method: Local reading by the main session: all 26 v-eval files, the evolve-loop runtime evaluator skill and its three reference files, the audit skill, the evaluator isolation profile, the grader best-practices guide, two evolve-loop research dossiers, the evo plugin manifest, four installed ECC and personal skills, and the installed `claude plugin eval` help. 12 local sources, all read in full. Paths in backticks are relative to the evolve-loop repository root unless stated otherwise.

Facts, inferences, single-source claims, and unverified items are labeled inside the memo; a label applies to the sentence or row it sits in.

---

Sources are files on the maintainer's machine, read in full. Paths in backticks are relative to the evolve-loop repository root unless stated otherwise.

## 1. Claude Code has a native eval harness for plugins and skills

Installed Claude Code 2.1.270 ships `claude plugin eval`. Verified from `claude plugin eval --help`:

- Case layout: `<eval dir>/**/case.yaml` or `prompt.md` plus `graders/*.md`; default dir `evals/`.
- Ablation arm: `--ablation with-without` runs a no-plugin baseline and reports the score delta. Graders tagged with-only (for example `tool_used: Skill`) count as a "plugin fired" indicator, not score.
- Grader types: free graders and paid LLM graders; `--judge-model` (default haiku).
- Runs: `--runs` per case (default 3), `--threshold 0..1` exits 1 below threshold, `--max-cost-usd` hard ceiling, `--concurrency 1-8`.
- Output: `--json [path]` full run result with prompts, graders, per-run scores; `--report` self-contained HTML; `aggregate-result.json` under `<eval dir>/results/<timestamp>/`.
- Safety: first-run trust prompt, `--trust-plugin` for CI, OS sandbox on shell tools, `--allow-tools` operator grant for Bash/Write/Edit/WebFetch/mcp, `--scaffold` off by default, MCP mocks by default.
- Authoring: `claude plugin eval init` runs an interview that sources inputs and designs graders; `--bare` writes `prompt.md` + `graders/criteria.md`.

Implication: v-eval's own skill can be regression-tested with a first-party harness that already has baseline ablation, cost ceilings, repeated runs, and JSON output. None of the v-eval docs mention it. This is the cheapest path to "evaluate the evaluator" for the skill layer.

## 2. evolve-loop already has an evaluator, and it is the design v-eval argues against

`runtime/skills/evaluator/SKILL.md` (evo plugin v22.21.0, Apache-2.0, Go 1.23):

- Five layers: SCOPE, GRADE, DETECT, SCORE, DIRECT, plus periodic META-EVAL.
- Six dimensions scored 0.0 to 1.0 with weights (correctness 0.25, security 0.20, maintainability 0.20, architecture 0.15, completeness 0.10, evolution 0.10) and a **weighted composite** mapped to STRONG / ADEQUATE / NEEDS WORK / CRITICAL.
- Anti-gaming: EST perturbation test (re-grade after semantic-preserving reformat; gaming indicator G with 0.1 and 0.3 bands and a fixed 20% penalty), saturation monitor, proxy-true correlation thresholds.
- Isolation is real, not prompt-level: `.evolve/profiles/evaluator.json` gives read-only repo, denied write paths (skills, agents, scripts, evals, state), no network, `max_budget_usd` 0.3, `max_turns` 20, challenge token, different CLI family from builder (codex-tmux default, claude-tmux fallback).
- Output schema: Markdown report plus JSON with `composite`, `verdict`, `gamingCheck`, `dimensions{score, confidence, findings}`, `priorities`, `regressions`.

v-eval's design doc explicitly rejects a default composite and says unlike scales and missing evidence cannot be averaged. So the maintainer has already built the thing the new docs argue against. This is the single most important tension to resolve before any code.

## 3. evolve-loop's audit verdict vocabulary and floor rules

`runtime/skills/audit/SKILL.md`: verdict tokens `PASS | FAIL | WARN | SKIPPED`; ALL-PASS rule across four parallel sub-auditors (eval-replay, lint, regression, build-quality); adversarial mode default on ("require positive evidence for PASS"); auditor model family differs from builder family to break same-model sycophancy; phase gate `gate_audit_to_ship` enforces PASS.

Smart-advisor memory: role archetypes Plan / Build / Evaluate; floor rule "must Evaluate, family(eval) != family(build)"; `core.VerdictReason{Status, Summary, Taxonomy}` exists in Go.

v-eval's proposed vocabulary is PASS / FAIL / UNKNOWN / ERROR / NOT_APPLICABLE per criterion and PASS / FAIL / INCOMPLETE overall. Mapping to evolve-loop: FAIL to FAIL, INCOMPLETE to WARN, NOT_APPLICABLE to SKIPPED is plausible; ERROR has no evolve-loop equivalent.

## 4. evolve-loop's hard-won grader lessons (directly reusable)

`runtime/docs/eval-grader-best-practices.md` (normative, gate-enforced):

- Cycles 102 to 111 were a reward-hacking class rooted in tautological graders; `mutate-eval.sh` enforces a kill rate of at least 0.7 at the gate, 0.8 target.
- Grader level taxonomy: L0 no-op, L1 source-presence grep, L2 output-file check, L3 execution-based jq, L3.5 control-flow structural awk, L4 end-to-end behavioral. Target L3.5 or higher.
- Eval definitions are one-file Markdown with a bash block (`.evolve/evals/<task>.md`).

`runtime/docs/research/execution-grounded-process-supervision-2026.md`: "Stop letting any model report whether the work is done; let the sandbox's exit code be whether the work is done." Cites Skalse et al. 2022 that the only unhackable proxy is a constant, so no auditor-only fix exists.

`runtime/docs/research/verdict-and-gate-proxy-failure-class-2026-06-03.md`: all six phase classifiers derive verdicts by grepping prose headings from LLM reports; headings drifted twice; no golden contract test existed. Lesson for v-eval: a report contract must be machine-validated JSON, never prose-grep.

## 5. evolve-loop extension points v-eval could plug into

- User-defined phases as pure data: `.evolve/phases/<name>/{phase.json, agent.md, profile.json}`, optional-only, `kind: llm`, canonical `verdict_on_pass`, `classify.require_sections`. Merged to main 2026-05-27.
- Evaluator delegation from the Auditor phase is advisory (dimension scores merged into audit-report under "Evaluator Scores", does not override verdict).
- Plugin packaging: `.claude-plugin/plugin.json` and `marketplace.json`, plus `.codex-plugin/plugin.json`. Skills generate slash commands via `evolve skills generate`.

## 6. Other installed prior art

- ecc `eval-harness` skill: eval-driven development, capability and regression evals, code-based and model-based graders, pass@k. Generic; no per-criterion evidence contract.
- ecc `agent-eval` skill: YAML task definitions with `judge: [{type: pytest, command}, {type: grep, pattern}]`, commit pinning, git-worktree isolation, metrics pass rate / cost / time / consistency. Closest local analog to "task plus judge", but it grades agents, not a single artifact.
- ecc `ai-regression-testing` skill: documents the same-model-writes-and-reviews blind spot with a four-fix real incident.
- `agent-self-evaluation-patterns` skill: confidence calibration, LLM-as-judge independence, eval-driven development anti-patterns.

## 7. Gaps in the current v-eval docs revealed by local research

1. No mention of `claude plugin eval` as the skill-testing harness.
2. No mention of evolve-loop's evaluator, its composite scoring, or its isolation profile.
3. No mention of the grader level taxonomy or mutation kill-rate gate, which are the maintainer's own strongest anti-tautology defense.
4. No mention of the prose-grep classifier drift incident, which is the strongest argument for a JSON-first report contract.
5. The requirements say "create a GitHub project"; no repo or remote exists yet, and there are zero commits.

## 8. Maintainer axis stated in session, 2026-09-14

Verbatim: "Don't trust any concise summary, verify and check from the context", "seeing is believing", "Find the hidden trace under the surface".

Mapping to existing material:

- Partly covered by REQ-05 (inspect actual artifacts) and REQ-13 (evidence-backed reasoning), and by the design doc's rule that candidate-supplied logs are "supplied evidence" that cannot establish an execution PASS.
- Not yet stated as its own requirement or as a skill obligation. It deserves a requirement ID and a mandatory policy line.
- Strongly corroborated by evolve-loop's own incident history: tautological graders (cycles 102-111) and prose-heading verdict classifiers that drifted from report content. Both are cases of a gate trusting a surface signal instead of the ground truth beneath it.

Design consequences to test in the questions:

1. Evidence hierarchy as policy: observed execution > direct inspection > supplied logs > candidate summary. Summary alone can never yield PASS.
2. Claim-to-verification table as a required report section: each claim in the candidate's summary is enumerated and either verified (with the action taken) or marked UNVERIFIED.
3. Every PASS cites an artifact the evaluator opened or a command it ran, with location and quoted observation.
4. An explicit "what the summary omits" pass: touched test files, skips, config changes, empty sections under confident headings, unmentioned side effects.
