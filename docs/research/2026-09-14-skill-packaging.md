# Prior art for skill-packaged AI evaluation

Research date and access date: 2026-09-14.

Method: Web-research agent. About 25 sources cited inline; six primary sources read in full (the Husain and Shankar post and repository, the smevals README and announcement, the official plugin-evals and skills documentation, the agentskills.io specification and evaluating-skills page, and the superpowers skills).

Facts, inferences, single-source claims, and unverified items are labeled inside the memo; a label applies to the sentence or row it sits in.

---

Prepared 2026-09-14 for the v-eval packaging decision. All URLs accessed 2026-09-14. **F** = fact from a read source; **I** = my inference; **UNVERIFIED** = no primary confirmation; **[single-source]** = one secondary source.

## 1. Hamel Husain & Shreya Shankar — "Evals Skills for Coding Agents"

- **Source:** hamel.dev/blog/posts/evals-skills/ — pub. 2026-03-02, updated 2026-08-31 (F). Repo github.com/ai-evals-course/evals-skills (587 stars, v0.3.1, last push 2026-08-31). Read in full: post, README, `.claude-plugin/plugin.json`, `skills/eval-audit/SKILL.md`.
- **Skills (F):** `evals-start` (router), `eval-audit`, `error-discovery`, `generate-synthetic-data`, `write-judge-prompt`, `validate-evaluator`, `evaluate-rag`, `build-review-interface`. Scope is *product evals*: trace error analysis, LLM-judge design, judge calibration (TPR/TNR), annotation UIs.
- **Packaging (F):** `npx skills add https://github.com/ai-evals-course/evals-skills` (vercel-labs/skills). Repo carries `.claude-plugin/{plugin.json,marketplace.json}`, `.codex-plugin/plugin.json` and `.agents/plugins` — one skill tree published as both a Claude Code and a Codex plugin. Invoked by prompt; `eval-audit` fans diagnostics to parallel subagents.
- **Explicit non-goals (F):** README: skills cover only what "generalize[s] across projects… not production monitoring, CI/CD regression suites, and cost optimization." `eval-audit`: "Do NOT use when the goal is to build a new evaluator from scratch." Post: "only a starting point."
- **License:** no LICENSE file; GitHub reports `null` (F). Reuse rights UNVERIFIED.
- **Contract-based acceptance?** No (F). None takes artifact + criteria and emits per-criterion verdicts; `eval-audit` emits a prioritized findings list. It does insist on binary pass/fail evaluators over Likert scales — a principle v-eval shares (I).

## 2. Simon Willison — smevals

- **Source:** simonwillison.net/2026/Jul/31/smevals/ (2026-07-31) + README at github.com/prime-radiant-inc/smevals (MIT; F). PyPI 0.1.0 (07-17), 0.2.0 (07-18). Activity: 265 stars, sole contributor simonw, last push 2026-08-11, 1 open issue (F). "My third iteration"; built with Jesse Vincent's Prime Radiant.
- **Layout (F):** `eval.yaml`, `tasks/*.yaml`, `configs/*.yaml`, `graders/*.yaml`, `checkers/*`, a runner executable, tool-owned immutable `runs/<task>/<config>/<model>/<ts>/{run.yaml,output.txt,grades/<grader>/{grade.yaml,grader.yaml}}`.
- **Execution vs grading (F):** `smevals run` executes a Runner (any executable; env-var contract `SMEVALS_MODEL/TASK/PROMPT/RUN_DIR`; stdout → `output.txt`; non-zero exit = *harness failure, never graded*). `smevals grade` applies a Grader later; multiple graders coexist; `--regrade` without re-running; each Grade snapshots its Grader.
- **Grader types (F):** Grader = ordered Checks with `required`, `creates`, `scoring.pass_threshold`. Built-ins only `contains`, `xml-valid`; everything else is a Checker executable returning exit code + JSON `{score, metrics, tags, notes, details}`. Outcome `fail` if any check failed, else `pass` if score ≥ threshold; a crashed check leaves score null.
- **Reports (F):** `report` (markdown, `--json`), `serve` (live UI), `build` (static HTML); `docs` prints the README for the agent.
- **Relevance:** README names "evaluating whether an implementation satisfies a provided specification" as an example Eval, but ships no such checker (F). Its harness-error ≠ grading-failure rule maps onto v-eval's ERROR vs FAIL (I).

## 3. Claude Code native: `claude plugin eval` and `/skill-doctor`

- **Sources (official):** code.claude.com/docs/en/plugin-evals (read in full), …/skills, …/plugins-reference, anthropics/claude-code CHANGELOG 2.1.269 ("Added `claude plugin eval`… JSON + HTML report"). Secondary: matthewswong.com (09-06), samuellawrentz.com (09-12), byteiota (09-12), goldie.agency (09-13).
- **What it evaluates (F):** *plugin behaviour*, not artifacts. Case = `evals/<case>/prompt.md` (+ `case.yaml`) + `graders/*.md`. Six grader types: free `regex`, `tool_used`, `tool_order`, `file_exists`; judge-billed `llm` (prose PASS/FAIL rubric), `baseline`. "There are no custom-code graders." Each case runs 3× with and 3× without the plugin → `WITH`, `W/OUT`, `Δ`; `tool_used: Skill` is indicator-only under ablation.
- **Sandbox (F):** throwaway HOME/cwd/config per run via `claude -p`; no user settings, CLAUDE.md, MCP, memory or other plugins; only `EVAL_*` env; agent cannot read the eval dir; Artifact off; never prompts — `Bash/Write/Edit/WebFetch/WebSearch` removed unless `--allow-tools`; granted Bash runs in the OS sandbox (bubblewrap/socat; native Windows refused). `scaffold_script` runs as the operator, only with `--scaffold`.
- **Report/CI (F):** `results/<ts>/{aggregate-result.json,report.html}`; `--json` → `schemaVersion: 1` with `aggregates.{overallScore,casesPassed,casesTotal,meanDelta}`, `cases[].aggregates.{score,delta}`, `partial`. Exit 0/1/2(partial)/130/143. Documented CI line: `--trust-plugin --json results.json --threshold 0.8 --model … --judge-model … --no-publish --max-cost-usd 20`. Needs v2.1.269+.
- **Availability conflict:** three practitioners report the command is early-access gated (exits 1); the official page says nothing. UNVERIFIED for arbitrary accounts.
- **`/skill-doctor` (F, official skills doc):** shipped v2.1.261, 2026-09-04 [single-source for version]. Shows each skill's context cost and invocation frequency, flags never-invoked skills, lists unused plugins; excludes bundled/enterprise skills; `/plugin` Stats tab or text under `-p`; needs v2.1.252+; not over Remote Control. Context-hygiene tool, not an eval; no published schema.
- **Official guide to skill evals (F):** two. (a) the plugin-evals page; (b) `skill-creator` workflow at agentskills.io/skill-creation/evaluating-skills — `evals/evals.json` (prompt, expected_output, files, assertions) → per-assertion `grading.json` with `passed` + `evidence` → `benchmark.json` with/without deltas → blind A/B. Docs state the two case formats are separate.
- **Would v-eval be testable? (F→I)** Yes *if* shipped with `.claude-plugin/plugin.json` (bare SKILL.md folders are not targets; "skills-directory plugins" also need the manifest). Cases would check that v-eval fired (`tool_used: Skill`) and that the written report matches (`regex`/`file_exists`). Caveats: fixtures must come via `scaffold_script`/`add_dirs`; grading v-eval's own judgement with an `llm` grader is judge-of-a-judge noise the docs warn about.

## 4. Agent Skills spec, marketplaces, neighbouring skills

- **Spec (F):** agentskills.io/specification — required `name` (≤64, kebab, = dir) and `description` (≤1024); optional `license`, `compatibility`, `metadata`, `allowed-tools` (experimental); `scripts/`, `references/`, `assets/`; body <5k tokens / 500 lines; `skills-ref validate`.
- **Cross-vendor (F):** Codex docs: "Skills build on the open agent skills standard… Plugins are the installable distribution unit"; `.agents/skills`, `$skill`, `agents/openai.yaml`; bundled `skill-creator`, `review-agent`. Gemini CLI: agentskills.io-based, `activate_skill` with consent prompt, `gemini skills install <git>`. openai/plugins mirrors obra/superpowers. "Codex has an official skill evals framework" — UNVERIFIED [single-source: thinkingtokens.ai].
- **Anthropic catalogs (F):** anthropics/skills (19) and claude-plugins-official (39) contain no acceptance-against-spec skill. `code-review` = 4 agents checking CLAUDE.md compliance + bugs, confidence ≥80. `pr-review-toolkit` = 6 aspect agents. Bundled `/verify` builds and drives the app, recording a recipe to `.claude/skills/verify/SKILL.md`.
- **obra/superpowers (MIT; F):** `verification-before-completion` is a discipline skill ("NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE"; includes "re-read plan → checklist → verify each → report gaps") with no output schema. `requesting-code-review` dispatches `code-reviewer.md`: has a "Plan alignment" section but outputs Strengths / Issues by severity / "Ready to merge? Yes|No|With fixes".
- **Community nearest matches (F):** Microck `reviewing-code` — reads user-story acceptance criteria, "Spec Alignment /20", verdict APPROVE/NEEDS REVISION/MAJOR REWORK. hirogakatageri `develop:product-reviewer` — Implemented/Partial/Missing with evidence. cskwork/skill-ab-eval — with/without × harness leaderboard CLI.
- **Gap (I):** nothing emits five-state PASS/FAIL/UNKNOWN/ERROR/NOT_APPLICABLE per criterion with evidence. `UNKNOWN`/`NOT_APPLICABLE` have no prior art here; smevals' harness-failure rule is the only `ERROR` analogue.

## 5. 2026 writing on skill vs CLI vs adapter packaging

- Chame, "MCP vs CLI vs CLI+Skills" (Medium, 2026-05-03): "use the lightest interface that still gives you enough reliability, safety, and reuse"; CLI+Skills = simple execution + guardrails.
- Watt, "The Reuse Ladder" (2026-07-09): instruction → skill → plugin; Claude/Codex manifests differ, so packaging "does not always transfer byte-for-byte".
- Pexo (2026-09-10): "the layer is not the transport"; things get "wrapped as a Claude Code skill… because the skill container is the only agent-native distribution channel".
- Vaughan on supabase/evals (2026-08-02): skill atop MCP lifted pass rates (Sonnet 46%→71%); activation went ~10%→60% by rewriting the description — discoverability is a first-class risk.
- RockB (2026-07-09): keep evals in versioned files outside any framework.
- Willison's pattern: CLI as execution, README-as-skill via `smevals docs`.
- No source directly weighs "eval workflow as skill vs adapter"; Hamel/Shreya's "eval tools… nudge you toward generic metrics" is the strongest pro-skill argument (I: valid for judgement steps, not for deterministic verdict aggregation).

## Implications for the v-eval packaging decision

1. Prior art splits: skills carry *judgement/procedure* (evals-skills, superpowers); CLIs carry *reproducible execution + grading* (smevals, plugin eval). No one ships contract acceptance with structured verdicts — v-eval is differentiated either way.
2. **A — pure SKILL.md (+ `scripts/`), spec-portable.** Pros: `npx skills add` on Claude/Codex/Gemini, no runtime deps, matches evals-skills. Cons: verdict rollup and evidence formatting re-derived by the model each run (drift); not `claude plugin eval`-testable without a manifest; lives or dies by `description`.
3. **B — skill + thin deterministic script (`scripts/veval-report`) + `.claude-plugin` and `.codex-plugin` manifests (the evals-skills layout).** Model judges per criterion; script validates the five-state schema, computes rollups, emits JSON/markdown. Testable via `claude plugin eval` (regex/file_exists on the report) and skill-creator `evals.json`. Cost: one more artifact; Bash must be granted in evals.
4. **C — smevals adapter (v-eval as Runner + Checker).** Free immutable runs, re-grading, HTML leaderboards; but frames v-eval as model comparison, not a one-shot gate, on a single-maintainer dependency.
5. Recommendation (I): B, with A as the guaranteed portable subset and C as an optional export. Grade report files deterministically; reserve `llm` graders for short summaries.

## Gaps not resolved

- Whether `claude plugin eval` is generally available or early-access gated (docs vs practitioners).
- `/skill-doctor` exact output format; no schema published.
- Verbatim grader-option table (page truncated there; the six types are confirmed).
- evals-skills license (no LICENSE file).
- Codex official skill-eval framework beyond `skill-creator` (single secondary source).
- Any Gemini CLI skill-testing tooling (no primary source found).
