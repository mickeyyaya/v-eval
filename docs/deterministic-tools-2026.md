# Deterministic tools for evaluating AI output in 2026

As of **September 14, 2026**. This is a focused selection of twelve existing tools and recent research, not a comparative benchmark. Official documentation and representative releases were checked online. No tools were installed or adopted, and v-eval still has no production evaluation runner.

Useful deterministic checks already exist. The opportunity for v-eval is to connect their narrowly defined results to the user's requirements and preserve missing evidence. An exact checker for the wrong requirement remains a bad evaluation.

## What “deterministic” should mean here

A deterministic check produces the same result for the same artifact, rule, and relevant environment. Keep three properties separate:

- **Deterministic scoring:** a fixed predicate evaluates a fixed input.
- **Reproducible execution:** the inputs, dependencies, configuration, and environment can be reconstructed.
- **Valid evaluation:** the predicate actually represents the intended requirement.

The first does not establish the other two. A regular expression can consistently accept a misleading claim. Tests can depend on time, network responses, ordering, or random inputs. A script is not automatically deterministic merely because it is code.

Setting a model's temperature to zero does not guarantee repeatable API output; Anthropic explicitly documents that limitation. A model that extracts claims followed by a deterministic checker is a **hybrid evaluation**: the checker can validate the extracted representation while the extraction is wrong or incomplete. Likewise, a fixed DAG of evaluation steps controls execution order, but its LLM nodes remain model judgments. [Claude documentation on temperature](https://platform.claude.com/docs/en/about-claude/glossary).

## Existing tools, grouped by the evidence they supply

### Structured output: Ajv and python-jsonschema

**Ajv** validates JSON against a configured schema, including types, required properties, ranges, and structural constraints. Strict mode helps expose ambiguous or silently ignored schema constructs. Record the schema dialect and options. Avoid silently coercing or deleting candidate data during acceptance validation; if normalization is intentional, retain both raw and normalized artifacts. [Ajv strict mode](https://ajv.js.org/strict-mode), [data-modifying options](https://ajv.js.org/guide/modifying-data.html).

**python-jsonschema** supplies validators for different JSON Schema drafts. Its documentation makes an easily missed point: `format` is not enforced unless a format checker is supplied; some checks also need optional dependencies. Unknown formats are not automatically rejected. Validate the schema and exercise a known-invalid example before claiming that a format requirement is covered. [Schema validation documentation](https://python-jsonschema.readthedocs.io/en/stable/validate/).

These tools can establish “every item has an integer quantity greater than zero.” They cannot establish that the quantities are correct, the described products exist, or the chosen fields satisfy the user's business need. A syntactically valid citation URL is similarly only a string conforming to the schema.

### Documentation rules: Vale and markdownlint

**Vale** applies configurable prose rules, such as terminology substitutions and rules scoped to headings or other document regions. The project explicitly distinguishes style consistency from general writing correctness. Use it for an agreed style guide, terminology policy, or local prohibition; retain rule IDs and configuration. It cannot establish factual support or usefulness for the intended reader. [Vale introduction](https://docs.vale.sh/).

**markdownlint** checks Markdown/CommonMark conventions, including configurable heading, list, whitespace, and related rules. Its custom-rule capability can support additional concrete constraints. Select the appropriate parser and rules for the repository; extensions and deliberate formatting choices can affect applicability. A clean lint result does not establish that a document contains the required argument, accurate instructions, or valid evidence. [Official markdownlint project and rules](https://github.com/DavidAnson/markdownlint).

For v-eval, these are separate criteria such as “uses the project's preferred terminology” and “passes the configured Markdown rules.” Do not relabel either result “document quality.” A link checker, if later added, would establish reachability at a particular time, not that the destination supports the associated claim.

### Generated code: pytest, Ruff, Pyright, and Semgrep

**pytest** runs explicit assertions about behavior. Existing project tests are often the strongest starting point because they reflect the project's interfaces and environment. Preserve collection counts, skips, failures, and the tested revision. Its exit-code specification distinguishes successful execution from failures, usage errors, and no tests collected; a wrapper must not flatten these into one acceptance result. [pytest exit codes](https://docs.pytest.org/en/stable/reference/exit-codes.html).

**Ruff** checks selected Python lint rules and provides formatting tools. It can identify concrete violations such as unused imports or configured coding patterns. The enabled rules and exclusions define its scope. A Ruff pass is neither a runtime test nor a type-correctness result. Run check modes during evaluation; modifying the candidate would create a different artifact. [Official Ruff documentation](https://docs.astral.sh/ruff/).

**Pyright** performs static type checking for Python. It contributes evidence about consistency with annotations and modeled language behavior. Untyped areas, permissive configuration, stubs, ignored diagnostics, and dynamic code limit the conclusion. Type compatibility cannot prove that the implementation performs the requested calculation or handles the right domain cases. [Official Pyright project](https://github.com/microsoft/pyright).

**Semgrep** applies code-pattern and data-flow rules, useful for project-specific prohibited constructs and selected security checks. Use explicit rule versions and record scan scope and exclusions. Keep rule-engine findings separate from any optional model-assisted analysis. No findings means no matches were found under the scan's configuration; it does not prove the absence of vulnerabilities. Edition and engine capabilities differ, so an adapter must record what ran. [Official rule-writing documentation](https://docs.semgrep.dev/writing-rules/overview).

The four tools answer different questions. A patch may pass formatting, typing, and every supplied test while violating an untested design constraint. Conversely, a valid alternative implementation should not fail simply because a checker encodes an unstated preference for one solution.

### Stronger test evidence: Hypothesis and Stryker

**Hypothesis** generates examples to exercise properties, such as round-trip preservation or invariants under input transformations. This expands the cases checked, but the user must still define a meaningful property. Its `derandomize` option stabilizes generated cases until relevant code or versions change; it is not a promise of identical behavior across every environment. Record settings, seed or replay information when applicable, dependency versions, and retained counterexamples. [Hypothesis settings and API](https://hypothesis.readthedocs.io/en/latest/reference/api.html).

**Stryker** mutates code and checks whether tests detect the changes. Its documented mutation score is detected mutants divided by valid mutants; detected includes killed and timeout cases, while invalid mutants are excluded. This gives evidence about test sensitivity under chosen mutation operators. It does not prove full correctness, and equivalent mutants or unstable timing complicate interpretation. Keep mutation scope, exclusions, timeouts, and the underlying test run visible. [Stryker metrics](https://stryker-mutator.io/docs/mutation-testing-elements/mutant-states-and-metrics/), [equivalent mutants](https://stryker-mutator.io/docs/mutation-testing-elements/equivalent-mutants/).

Both approaches can fit a reproducible evaluation process, but neither entire run should be advertised as unconditionally deterministic. First stabilize the existing tests; use these tools when a concrete adequacy question justifies them.

### Evaluation frameworks: promptfoo and Inspect

**promptfoo** supports literal, regular-expression, JSON, and custom JavaScript/Python assertions. Its documentation distinguishes model-dependent checks from deterministic assertions. A custom function can enforce a domain-specific requirement, but may itself call a model, webhook, or unstable service; inspect the implementation before labeling it deterministic. Prefer raw criterion results over an aggregate that could hide a required failure. [Deterministic metrics](https://www.promptfoo.dev/docs/configuration/expected-outputs/deterministic/), [custom Python assertions](https://www.promptfoo.dev/docs/configuration/expected-outputs/python/).

**Inspect** supplies standard and custom scorers alongside model grading. It can organize scoring and preserve evaluation logs; a custom scorer can use ordinary deterministic code. Choosing the framework does not choose the evidence method. Mark model grading as such, and keep the artifact, target, scorer version, and scoring policy together. [Inspect scorers](https://inspect.aisi.org.uk/scorers.html).

Either is a plausible future integration point. Neither eliminates the need to define acceptance criteria, validate the grader, or separate evidence supplied by the candidate from independently observed results.

## Verified release snapshot

The following versions were returned by each repository's official GitHub `releases/latest` API on the research date; every returned record was non-draft and non-prerelease. Dates below are the API's **publication dates in UTC**, not webpage crawl dates. Matching release pages were also opened. “Latest” here means the repository's designated latest release at lookup time, not a claim about every distribution channel.

| Tool | Verified release | Publication date | Primary release record |
| --- | --- | --- | --- |
| Ajv | v8.20.0 | 2026-04-24 | [GitHub release](https://github.com/ajv-validator/ajv/releases/tag/v8.20.0) |
| python-jsonschema | v4.26.0 | 2026-01-07 | [GitHub release](https://github.com/python-jsonschema/jsonschema/releases/tag/v4.26.0) |
| Vale | v3.21.0 | 2026-09-09 | [GitHub release](https://github.com/vale-cli/vale/releases/tag/v3.21.0) |
| Ruff | 0.16.7 | 2026-09-10 | [GitHub release](https://github.com/astral-sh/ruff/releases/tag/0.16.7) |
| Semgrep | v1.177.0 | 2026-09-10 | [GitHub release](https://github.com/semgrep/semgrep/releases/tag/v1.177.0) |
| promptfoo | 0.123.0 | 2026-09-10 | [GitHub release](https://github.com/promptfoo/promptfoo/releases/tag/0.123.0) |

No current release version is asserted here for the other six tools. Inspect's `releases/latest` lookup returned HTTP 404; its official documentation was available. That response does not establish that Inspect is unavailable or unmaintained. Versions should be verified again when selecting dependencies.

## What is actually new in 2026?

The established tools above predate 2026. Their release dates indicate current published software, not a newly discovered way to measure usefulness. Recent papers extend the kinds of requirements that can be checked programmatically:

- **IndicIFEval, February 25, 2026:** evaluates automatically verifiable instructions across fourteen Indic languages and provides an evaluation repository. It broadens the language coverage of constrained-generation evaluation. Language-specific tokenization and lexical rules need inspection before reuse; this is not a general factuality or translation-quality checker. [Paper](https://arxiv.org/abs/2602.22125), [official project](https://github.com/AI4Bharat/IndicIFEval).
- **IFHierBench, July 30, 2026:** describes 600 prompts with constraints at different structural depths and deterministic checkers scoped to those levels. Its direct design lesson is to record whether a rule applies to the entire artifact, one section, or a nested field. A checker remains limited to its formalized constraints. [Paper](https://arxiv.org/abs/2607.27912).
- **Constraint Saturation Evaluation, August 12, 2026:** studies simultaneous constraints using rule-based verifiers without LLM judging. Its reported degradation in satisfying whole sets of constraints motivates distinguishing individual pass rates from “all required criteria pass.” Its model results apply to that generated benchmark and protocol, not production acceptance rates. [Paper](https://arxiv.org/abs/2608.12426).

These are 2026 first submissions, not independently replicated results from this project. The papers were inspected for scope and methodology; their suites were not run. By comparison, **IFBench / Generalizing Verifiable Instruction Following** first appeared in **July 2025**. It remains relevant to verifiable constraint generalization, but should not be labeled a new 2026 publication. [IFBench paper](https://arxiv.org/abs/2507.02833).

## Recommendation for v-eval

Start with the repository's existing tests and standard-library checks for simple requirements: parse structured data, compare expected values, count explicitly defined items, and verify local references. Use an existing schema validator when a real schema contract warrants it; do not build a partial JSON Schema implementation. No dependency decision is made by this report.

For each requirement, record the smallest adequate check, its scope, and what remains unresolved. A document's required sections can be checked structurally; whether those sections answer the reader's question usually needs a separate rubric or human judgment. If a model extracts claims, preserve the original passages and evaluate extraction completeness separately from arithmetic or schema validity.

A future adapter should emit criterion ID, artifact identity, checker/rule version, command or function, environment, raw result, and evidence location. Map failures narrowly: an observed rule violation is FAIL; an operational crash is ERROR; an unavailable required check is UNKNOWN. A required semantic judgment cannot disappear because all mechanical checks passed. Compute overall acceptance from the documented policy rather than asking a model to invent it.

Before adoption, try one valid case, one known violation, one valid alternative, and one missing-evidence case. This checks whether an adapter supports the actual requirement and rejects obvious false evidence. It does not establish broad accuracy. Costs, language fit, contributor familiarity, and measured usefulness on v-eval's pilot should determine which integrations follow.

## Source inventory and research limits

All linked tool documentation is first-party, accessed 2026-09-14; absent an explicit release record, no publication date is inferred. The release table supplies the dated inventory for six representative tools. The twelve tool descriptions link to their official documentation or repositories.

| Research author | First submission | Title / primary source |
| --- | --- | --- |
| Jayakumar et al. | 2026-02-25 | [IndicIFEval: A Benchmark for Verifiable Instruction-Following Evaluation in 14 Indic Languages](https://arxiv.org/abs/2602.22125) |
| Mao and Chen | 2026-07-30 | [IFHierBench: Hierarchical Instruction Following for Large Language Models](https://arxiv.org/abs/2607.27912) |
| Vasileva | 2026-08-12 | [Large Language Models Can Follow Instructions, But Not Many at Once: Phase Transitions in Compositional Constraint Satisfaction](https://arxiv.org/abs/2608.12426) |
| Pyatkin et al. | 2025-07-03 | [Generalizing Verifiable Instruction Following](https://arxiv.org/abs/2507.02833) |

This review verifies published capabilities and selected release metadata. It contains no local performance comparison, blanket security assurance, or evidence that any tool evaluates “meaningfulness” across arbitrary artifacts. Those gaps require task-specific examples and human-reviewed evaluation, as described in the [design](design.md) and [gaming analysis](gaming-and-defenses.md).
