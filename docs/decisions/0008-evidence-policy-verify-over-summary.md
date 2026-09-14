# Evidence policy: never trust a summary; PASS requires an opened-or-ran citation

* Status: accepted
* Deciders: maintainer (mickeyyaya)
* Date: 2026-09-14

Nothing described here is implemented. This record turns a maintainer axis into a core-enforced rule.

## Context and Problem Statement

The maintainer stated a main axis in their own words: "Don't trust any concise summary, verify and check from the context", "seeing is believing", "Find the hidden trace under the surface". The existing requirements partly cover it (REQ-05 inspect actual artifacts; REQ-13 evidence-backed reasoning) and the design downgrades candidate-supplied logs to "supplied evidence", but nothing says a summary can never be the basis for PASS, and nothing requires enumerating a summary's claims and verifying each. evolve-loop's own incidents were gates trusting a surface signal. How strictly should the rule be enforced, and where?

## Decision Drivers

* AI-generated work arrives wrapped in summaries (PR descriptions, "all tests passed", build reports, handoffs) that are candidate claims, not evidence.
* The integrity research shows prompt-only rules fail under pressure; only software-enforced rules hold.
* METR's 2026 investigation found spoofed tool calls and agents treating logs as not authoritative.
* The rule must work on weak local models, so it must live in the core.

## Considered Options

* Hierarchy plus claim table plus PASS requires an opened-or-ran citation, core-enforced
* Hierarchy and claim table, prompt-level only
* Hierarchy only

## Decision Outcome

Chosen option: "Hierarchy plus claim table plus PASS requires an opened-or-ran citation, core-enforced", because it is the only option that survives a weak model or a persuasive candidate. The policy:

1. Evidence hierarchy: observed execution, then direct inspection, then supplied logs, then candidate summary. Summary text alone can never yield PASS.
2. Claim-to-verification table: every claim in the candidate's summary is listed and either verified with the action taken (file opened, command run, passage located) or marked UNVERIFIED.
3. Every PASS cites something the evaluator opened or ran: a file and line range, a command with exit status and log location, or a quoted source passage. The core rejects a PASS whose evidence field lacks one.
4. A "what the summary omits" pass: touched test files, skips, configuration changes, empty sections under confident headings, unmentioned side effects.

### Consequences

* Good, because sycophantic or fabricated summaries cannot produce a PASS.
* Good, because the claim table is the natural home for the forensic findings in [0012](0012-forensic-detectors-v1.md).
* Bad, because reports get longer; the HTML renderer ([0021](0021-html-report-every-evaluation.md)) must keep the table readable.
* Bad, because criteria without inspectable evidence will often land on UNKNOWN, which the pilot must measure against over-abstention.

## Confirmation

Schema: the evidence object for a PASS has a required `citation` with one of `{file, lines}`, `{command, exit_status, log}`, or `{source, passage}`; validation fails otherwise. A core test feeds a report whose only evidence is a summary and expects rejection. The skill's draft-2 text carries the four rules verbatim. None exists yet.

## Pros and Cons of the Options

### Core-enforced hierarchy, claim table, citation rule

* Good, because it holds on every CLI and model.
* Good, because it makes the "hidden trace" pass mandatory.
* Bad, because more schema and more report length.

### Prompt-level only

* Good, because cheaper now.
* Bad, because the integrity research says prompt-only rules fail under pressure.

### Hierarchy only

* Good, because lighter reports.
* Bad, because the omissions pass is not forced.

## More Information

* Related requirements: REQ-05, REQ-13, and REQ-28 in [requirements](../requirements.md). Informed by [integrity and RSI research](../research/2026-09-14-integrity-and-rsi.md) and [local prior art](../research/2026-09-14-local-prior-art.md).
* METR, "OpenAI/Hugging Face incident investigation", spoofed tool calls; logs not treated as source of truth: <https://metr.org/blog/2026-08-26-openai-hugging-face-incident-investigation/> (2026-08-26; accessed 2026-09-14).
* METR, "Recent Frontier Models Are Reward Hacking": <https://metr.org/blog/2025-06-05-recent-reward-hacking/> (2025-06-05; accessed 2026-09-14).
* Nasr, Carlini et al., "The Attacker Moves Second", in-band defenses bypassed under adaptive attack: <https://arxiv.org/abs/2510.09023> (2025-10-10; accessed 2026-09-14).
* Narisetty et al., out-of-band deterministic enforcement: <https://arxiv.org/abs/2606.26479> (2026-06-25; accessed 2026-09-14; single paper).
* Zeng et al., LLMBar, judges reward polished responses that miss the instruction: <https://arxiv.org/abs/2310.07641> (accessed 2026-09-14).
* evolve-loop EGPS dossier, `runtime/docs/research/execution-grounded-process-supervision-2026.md`, read locally 2026-09-14.
* Revisit when pilot data shows the citation rule produces excessive UNKNOWN on legitimate cases.
