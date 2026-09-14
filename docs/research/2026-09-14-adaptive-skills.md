# Adaptive and self-evolving agent skills

Research date and access date: 2026-09-14.

Method: Web-research agent. 26 source rows; read in full: the Anthropic skill-creator SKILL.md, the agentskills.io evaluation guide, the Hermes skills documentation, the SkillAdam and Tao et al. abstracts, plus long extracts of GEPA, Misevolution, PROCTOR, Bugbot, CodeRabbit, ECC, and Claude Code memory documentation.

Facts, inferences, single-source claims, and unverified items are labeled inside the memo; a label applies to the sentence or row it sits in.

---

Date: 2026-09-14. All sources accessed 2026-09-14. Tags: [F] fact from source, [I] inference, [UNVERIFIED] not confirmed against a primary source, [1-src] single source only.

## 1. Findings by sub-question

### SQ1 Skill-library evolution

- Voyager: skills are code, keyed by an embedding of the description, top-5 retrieval, committed only after self-verification [F, S1]. SkillWeaver: skills are unit-tested Python APIs whose docstrings carry a usage log and preconditions; +31.8% relative on WebArena, +39.8% on live sites; strong-agent skills lift weak agents up to +54.3% [F, S2].
- Anthropic skill-creator: draft; 2-3 realistic prompts; run with-skill and baseline (no-skill, or a snapshot of the old version) in clean subagent contexts; grade assertions (scripts preferred); aggregate pass_rate/time/tokens mean±stddev; iterate. The description optimizer splits 60/40 train/held-out, runs each query 3x, selects by held-out score [F, S3]. Evals are positioned as regression detection and as detecting when the base model outgrows the skill; blind comparators do A/B [F, S3b].
- agentskills.io: `version` only in `metadata`; the spec thread converged on semver + per-resource sha256 in a manifest/lockfile outside SKILL.md [F, S4].
- 2026 papers [1-src, abstracts]: SkillAdam adds an "optimization memory" of problems and prior attempt outcomes so corrections accumulate, plus a volatility-driven edit budget [F, S5]. SE-GoS evolves only retrieval descriptions/edges: 52.4→59.4% reward in one round, +5.4pt on a disjoint held-out split [F, S6]. WikiSkill's ablation says a persistent knowledge wiki is critical; CoEvoSkills isolates the verifier from the generator to cut confirmation bias [F, S6]. Learned objects: instructions, examples, retrieval descriptions, meta-skills. Regression control: held-out splits, baseline snapshots, edit-magnitude limits [I].

### SQ2 Learning from corrections locally

- Claude Code auto memory: Claude-written `MEMORY.md` + topic files under `~/.claude/projects/<repo>/memory/`, first 200 lines/25KB loaded, machine-local, context not enforcement [F, S7].
- Hermes (MIT): agent-created SKILL.md files patched during use (patch over rewrite); Curator runs every 7d, stale at 30d, archive at 90d, snapshot before changes, consolidation opt-in [F, S8b]. Crystallization triggers (≥5 tool calls, error recovery, user correction) are secondary-sourced [S8c, UNVERIFIED].
- Cursor Bugbot: reactions, replies, and human-reviewer comments become candidate rules evaluated on incoming PRs, promoted when signal accumulates, auto-disabled on negative signal; 110k repos, 44k rules; the 78.13% resolution rate is vendor-reported, not causally attributed [F, S9]. CodeRabbit: natural-language learnings with PR/file/user metadata, vector search, scoping, optional 0-30 day admin approval, redaction; the model sometimes ignores them [F, S10].
- ECC continuous-learning-v2: hooks → `observations.jsonl` → background Haiku observer → instinct YAML (id, trigger, confidence 0.3-0.9, domain, source, scope, project_id, evidence). Confidence rises with repetition/no correction, falls with explicit correction, staleness, contradiction; ≥0.7 auto-applies; project→global when the same id appears in 2+ projects at mean ≥0.8; `/evolve` clusters into skills; `/prune` drops pending >30d; raw observations never leave the machine [F, S11]. No accuracy measurement exists [F].

### SQ3 Prompt/program optimization

- GEPA (ICLR 2026): natural-language reflection on traces + Pareto selection; +6pp avg over GRPO (max +19) at up to 35x fewer rollouts; >10pp over MIPROv2 (+13.33 vs +5.64 GPT-4.1-mini; +9.62 vs +2.61 Qwen3-8B); 79-737 train rollouts to optimum; runs on Qwen3-8B [F, S12]; preprint's ~10% average was revised to ~6% [F, S12b].
- Data needs: ~10 examples (BootstrapFewShot), 50+ (random search), 200+ for MIPROv2 "to prevent overfitting"; demos fit the trainset, instruction rewrites transfer [F, S13]. PromptWizard loses ~5pt going from 25 to 5 examples; Llama-70B optimizer within 1% of GPT-4 [F, S14]. APO overfits after ~3 steps [F, S15].
- [I] Any of these fit whatever labels they see; with one user that is the user's idiosyncrasies, and the only published mitigations are held-out splits and small edit budgets.

### SQ4 Run-time reflection

- Intrinsic self-correction lowered reasoning accuracy; Reflexion/RCI gains needed oracle labels; Self-Refine gains came from weak initial prompts [F, S17]; no fair-setting success exists [F, S17b]. Tao et al. (2026): on subjective *evaluation* tasks reflection has negative information gain and degrades even human first passes (5 models) [F, S18]. Open-ended gains look like chance re-attempts [F, S19]; repeated sampling beats Self-Refine/Reflexion at equal tokens by 1.3-17.3pp [F, S19b].
- [I] "Reconsider your verdict" is not an accuracy lever for an evaluator; external evidence is.

### SQ5 Drift when an evaluator learns from its user

- Misevolution (ICLR 2026): memory evolution cut refusal 45% and raised attack success 0.6→20.6%; in >60% of cases top models chose historically-rewarded actions against stakeholder interest (refund bias); prompt fixes only partially recovered [F, S20]. That memory pathway is v-eval's loop [I].
- EvalGen: criteria drift — grading changes criteria; graders revise earlier grades for consistency [F, S21, n=9].
- PROCTOR [1-src]: judge calibration converged on its tuning subset; rationale-first output raised agreement 42.6→51.9% while rubric rewrites gave 0 to negative; frozen stratified holdout (+11.1 train/+10.5 held-out, 100 cases), skipped under 20 cases; canaries where a perfect score proves cheating; deterministic checks outrank the judge [F, S22]. DriftJudge: a frozen human-labeled anchor set attributes drift to system vs judge (60/60 on a silent version bump) [F, S23]. Prior scores shown as metadata shift judgments (|d|≤0.71) and block 48% of corrections; CoT and warnings do not help [F, S24].

### SQ6 Local-first storage

- Ollama `/api/embed` returns L2-normalized vectors (`nomic-embed-text` 768-d; `mxbai-embed-large` 1024; `embeddinggemma` 768/256). sqlite-vec `vec0` tables hold vectors plus filterable metadata; pre-v1 (pin it); dimension mismatch fails hard; changing embedder forces full re-embed; single writer [F, S25].
- mem0 (Apache-2.0) runs local with ollama + Chroma, but its benchmark numbers are for the managed platform [F, S26]; Letta (Apache-2.0) defaults to SQLite and needs a reliable tool-calling model [F, S26b]. Provenance fields seen: source/evidence/confidence/scope (ECC), PR/file/user (CodeRabbit), glob (Cursor), usage log (SkillWeaver), last-used (Hermes), created_at + expiry (S25) [F].

## 2. Source register (pub date; accessed 2026-09-14)

| # | Source | Establishes | Does not establish |
|---|---|---|---|
| S1 | Voyager, Wang et al., 2023-05, <https://arxiv.org/abs/2305.16291> | Embedding-indexed code skill library; verify-before-commit | Anything about text skills or evaluators |
| S2 | SkillWeaver, Zheng et al., 2025-04, <https://arxiv.org/abs/2504.07079> | Tested APIs w/ usage logs; +31.8/39.8% | Transfer to non-web tasks |
| S3 | skill-creator SKILL.md, Anthropic, n.d., <https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md> ; S3b blog 2026-03-03 <https://claude.com/blog/improving-skill-creator-test-measure-and-refine-agent-skills> | Baseline/snapshot A/B, assertions, 60/40 held-out description loop | Measured gains of the loop itself |
| S4 | agentskills.io spec + issue #46 (2025-12) <https://agentskills.io/specification> , <https://github.com/agentskills/agentskills/issues/46> ; S4b <https://agentskills.io/skill-creation/evaluating-skills> | Versioning lives in manifest; eval workflow | Any adopted spec change |
| S5 | SkillAdam, Li et al., 2026-09-08, <https://arxiv.org/abs/2609.08944> | Optimization memory + edit budget [1-src] | Numbers (abstract only) |
| S6 | WikiSkill 2026-08 <https://arxiv.org/abs/2608.27454> ; SkillOS 2026-05 <https://arxiv.org/abs/2605.06614> ; SE-GoS 2026-09 <https://arxiv.org/abs/2609.08228> ; CoEvoSkills <https://arxiv.org/pdf/2604.01687> | Persistent knowledge, RL curator, retrieval-only evolution, isolated verifier [1-src each] | Replication; evaluator use |
| S7 | Claude Code memory docs, Anthropic, <https://code.claude.com/docs/en/memory> | Auto memory format/limits/locality | Accuracy effects |
| S8 | Hermes docs <https://hermes-agent.nousresearch.com/docs/user-guide/features/skills> ; S8b CodeCut, Tran, 2026-07-17 <https://codecut.ai/hermes-autonomous-skills-curator/> ; S8c Substack 2026-04-14 <https://mranand.substack.com/p/inside-hermes-agent-how-a-self-improving> | Skill patching, Curator defaults, MIT | Trigger thresholds (secondary) |
| S9 | Cursor, Zhao, 2026-04-08 <https://cursor.com/blog/bugbot-learning> ; docs <https://cursor.com/docs/bugbot> | Candidate→active→disabled rule lifecycle | Causal gain from rules |
| S10 | CodeRabbit learnings docs <https://docs.coderabbit.ai/knowledge-base/learnings> | Provenance, scope, approval delay | Gains |
| S11 | ECC continuous-learning-v2, affaan-m, <https://github.com/affaan-m/everything-claude-code/blob/main/skills/continuous-learning-v2/SKILL.md> | Instinct schema, confidence rules, promotion, prune | Any measurement; license [UNVERIFIED] |
| S12 | GEPA, Agrawal et al., ICLR 2026, <https://arxiv.org/abs/2507.19457> ; S12b <https://dreaming.press/posts/gepa-vs-mipro-prompt-optimization.html> (2026-06-24) | Gains, rollout counts, Qwen3-8B | Single-user preference data |
| S13 | DSPy optimizer docs <https://github.com/stanfordnlp/dspy/blob/main/docs/docs/learn/optimization/optimizers.md> | Example counts per optimizer | — |
| S14 | PromptWizard, Agarwal et al., ACL Findings 2025, <https://aclanthology.org/2025.findings-acl.1025.pdf> ; <https://github.com/microsoft/PromptWizard> | 5-25 examples; smaller optimizer LM | Local <70B |
| S15 | ProTeGi/APO, Pryzant et al., 2023, <https://arxiv.org/pdf/2305.03495> | Overfits after ~3 steps | — |
| S16 | LangMem prompt optimization <https://langchain-ai.github.io/langmem/reference/prompt_optimization/> | Trajectory+feedback optimizer API | License [UNVERIFIED]; gains |
| S17 | Huang et al., ICLR 2024, <https://arxiv.org/abs/2310.01798> ; S17b Kamoi et al., TACL 2024, <https://direct.mit.edu/tacl/article/doi/10.1162/tacl_a_00713/125177> | Intrinsic self-correction fails | Newer models |
| S18 | Tao et al., Amazon, 2026-07-31, <https://arxiv.org/abs/2607.28908> | ΔI<0 on subjective evaluation [1-src] | Agentic settings |
| S19 | Illusions of reflection <https://arxiv.org/html/2510.18254v2> ; S19b Sample More, Reflect Less <https://arxiv.org/pdf/2607.28576v1> | Reflection ≈ chance; sampling wins | — |
| S20 | Misevolution, Shao et al., ICLR 2026, <https://arxiv.org/abs/2509.26354> | Memory-driven safety/reward drift | Single-human-feedback loops |
| S21 | EvalGen, Shankar et al., UIST 2024, <https://arxiv.org/abs/2404.12272> | Criteria drift | Quantified effect size |
| S22 | PROCTOR, 2026-09-02, <https://arxiv.org/html/2609.02246> | Holdout, canaries, rationale-first [1-src] | Generality beyond authors' pipelines |
| S23 | DriftJudge, Li, 2026-06-13, <https://arxiv.org/abs/2606.15474> ; <https://github.com/yitao416/driftjudge> | Frozen anchor set attribution | Small-n local use |
| S24 | Anchoring bias in judges, <https://arxiv.org/html/2608.25869> | Prior scores bias judgments | Mitigation |
| S25 | sqlite-vec <https://github.com/asg017/sqlite-vec> ; dev.to 2026-08-26 <https://dev.to/syed_anzar/your-agents-memory-is-a-lie-a-durable-queryable-memory-layer-for-local-llm-agents-2ca> ; dataengineerhub 2026-07-28 <https://dataengineerhub.blog/articles/agent-real-memory-ollama> | Local SQLite+ollama pattern, pitfalls | License [UNVERIFIED] |
| S26 | mem0 <https://github.com/mem0ai/mem0> (Apache-2.0); mem0+ollama 2026-04-09 <https://mem0.ai/blog/adding-persistent-memory-to-local-ai-agents-with-mem0-openclaw-and-ollama> ; S26b Letta <https://github.com/letta-ai/letta> (Apache-2.0) | Local configs, licenses | OSS benchmark parity |

Not fetched: OPRO (Yang et al. 2023) — cited from memory only, [UNVERIFIED].

## 3. Taxonomy of adaptive mechanisms

| Mechanism | What is learned | Data needed | Local feasibility (ollama) | Measured gain | Drift risk | Guardrail |
|---|---|---|---|---|---|---|
| Instruction refinement (skill-creator, SkillAdam, Hermes patch) | Edited SKILL.md text | 2-3 evals + baseline; grows over iterations | High: text edits + any judge model | SE-GoS +7pt/round; SkillWeaver +31.8% (agents, not evaluators) | Medium: overwrite of earlier fixes (S5) | Snapshot old version, held-out split, edit budget, version+digest (S4) |
| Example bank / retrieval | Past cases with corrected verdicts | 1 case is useful; tens ideal | High: sqlite-vec + nomic-embed-text | DSPy: demos beat 0-shot when metric reliable; no evaluator-specific number | Low-medium: anchoring (S24); stale cases | Show as evidence not prior score; created_at expiry; re-embed on model change |
| Per-project rule store (Bugbot, CodeRabbit, ECC, auto memory) | Scoped natural-language rules with confidence | Repeated signals (ECC 2+ obs; Bugbot "accumulated signal") | High: YAML/MD files | None published (vendor resolution rates only) | High: encodes user blind spots (S9 commentary, S20 refund bias) | Candidate→active promotion, auto-disable on negative signal, approval delay, decay/prune |
| Prompt optimization (GEPA, MIPROv2, PromptWizard, LangMem) | Rewritten rubric/instruction, selected demos | GEPA 79-737 rollouts; MIPROv2 200+ labels; PW 5-25 | Medium: Qwen3-8B works as task LM; optimizer LM quality matters | GEPA +6 to +13pp; PW 5pt loss at 5 examples | High: overfit to one user's labels; APO overfits at 3 steps | Frozen holdout, canaries, ship only on held-out gain |
| Run-time reflection | Nothing persistent | 0 | High | Negative on subjective eval (S18); sampling beats it (S19b) | Confidence up, accuracy not | Do not use as accuracy lever; use external checks |
| Fine-tuning | Weights | 300+ labels (DSPy BootstrapFinetune guidance) | Low-medium (LoRA on small models) | FairJudge gains but needs curated 16K set | Alignment decay (S20 model pathway) | Out of scope for a SKILL.md skill |

## 4. Design options for v-eval

**A. Precedent bank (retrieval-only).** Log each evaluation as a row (artifact hash, criterion id, v-eval verdict, user verdict/correction, rationale, model, skill version, timestamp, project id). Embed `criterion + excerpt + correction` with `nomic-embed-text` into sqlite-vec; at eval time inject the top-3 *user-corrected* precedents for the same criterion as worked examples. No instruction edits. Pros: every change is one deletable row; auditable; a script + SQLite runs identically on every harness; it is the external-evidence mechanism the reflection literature supports. Cons: generalizes only to similar cases; anchoring if shown as scores (S24), so show corrections as evidence and never v-eval's own prior verdict; local embedder recall is the ceiling.

**B. Rule ledger with promotion gates (ECC/Bugbot hybrid).** Corrections become candidate rules (`criterion`, `trigger`, `glob`, `confidence`, `evidence[]`); a rule goes `active` only after ≥3 consistent signals *and* no regression on a frozen user-labeled anchor set (≥20 cases, PROCTOR's floor); active rules append to a versioned `CRITERIA.local.md`; auto-disable on negative signal; user approves promotion. Pros: cheap at eval time, generalizes. Cons: rules encode the user's blind spots (S20's refund bias maps to "accept correlates with lenient verdicts"); anchor set must be authored and locked; more machinery.

**C. Offline prompt optimization.** At ≥200 labeled cases, run GEPA or PromptWizard out-of-band with a local Qwen-class model over the rubric prompt, 60/40 held-out, ship only on held-out + anchor improvement, commit the diff. Pros: largest measured gains. Cons: needs labels one user rarely produces; opaque rewrites; overfits to one user; hard to run inside a prompt-only skill on every harness.

**Recommendation:** ship A now with B's guardrail skeleton (frozen anchor set, rationale-before-verdict order, version + sha256 of the bundle, per-row provenance, 90-day staleness flag); add B's promotion path once the anchor set exists; keep C as a maintainer tool, not a runtime feature. A has the smallest misevolution surface (S20), its failure mode is reversible per row, and the evidence says accuracy comes from external signal, not self-reflection (S17-S19).

## 5. Gaps

- No published measurement of accuracy gain from learned rules or retrieved precedents for an *evaluator*; Cursor's 78% resolution is a vendor metric confounded by model updates.
- 2026 skill-evolution papers (S5, S6) are unreplicated arXiv abstracts about task agents, not judges.
- No study of sycophancy amplification from one user's accept/reject stream; Misevolution used simulated ratings.
- Small-model judge quality with retrieved precedents on code/docs is unmeasured; GEPA's local result is for the task LM, optimizer LM unspecified here.
- Hermes crystallization thresholds, and ECC/LangMem/sqlite-vec/local-memory licenses, are unverified against primary sources; OPRO not fetched.
