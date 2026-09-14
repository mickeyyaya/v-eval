# v-eval core walking skeleton and skill wiring: implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first executable v-eval: a Go core that validates a JSON report, enforces the evidence rule, derives verdicts and counts, renders Markdown and self-contained HTML, exports SARIF, and a `veval` command line the skill calls; with CI on three operating systems and release configuration.

**Architecture:** Ports-and-adapters around one canonical `report.Report` type. `core/report` owns types, validation, and aggregation (pure functions over the struct). `core/render` and `core/export` are adapters behind a `Renderer` port. `cmd/veval` is a thin table-dispatched CLI whose handlers take `(args, stdin, stdout, stderr)` and return an exit code, so every command is testable without a process. The Go struct is the single source of truth; `schema/report.schema.json` is the published contract and a drift test keeps them equal.

**Tech Stack:** Go 1.23 (module `github.com/mickeyyaya/v-eval`), standard library only (`encoding/json`, `text/template`, `html/template`, `crypto/sha256`, `embed`, `flag`, `testing`). No third-party modules. GoReleaser v2 for release builds. GitHub Actions for CI.

**Spec:** `docs/architecture/report-schema.md`, `docs/architecture/pipeline.md` (Verdict computation, Renderers, Trust boundaries), `docs/architecture/packaging-and-portability.md`, `docs/design.md`, decisions `0004`, `0005`, `0006`, `0007`, `0008`, `0013`, `0020`, `0021`, `0022`.

## Context

v-eval is documented and decided but has no executable core. The repository (`/Users/danleemh/ai/claude/v-eval`, first commit `ceec88c` on `main`, pushed to GitHub) holds the requirements, twenty-two decision records, the architecture set, a draft skill, and documentation tooling. The maintainer's direction on 2026-09-14: build the core function and the skill now; the auto review and release loop, in which each version is evaluated by v-eval's own skill against the previous version, comes next and depends on this core.

This plan is Roadmap Stage 1. It produces working, testable software on its own. Adapters, the classifier, forensic detectors, the learning loop, the agent profile, and the plugin manifests are later plans.

## Global Constraints

- **v-eval is an independent, individual project (decision 0001; maintainer's note of 2026-09-14).** No import, vendored code, shared tooling, path reference, build step, or runtime dependency on evolve-loop or any other maintainer project. Conventions below (table-dispatched CLI, golden tests, snake_case JSON, struct-as-source-of-truth) are adopted as v-eval's own and documented here, not referenced from elsewhere. Any later interoperability with evolve-loop happens through the published report and the verdict mapping in `docs/architecture/report-schema.md`, never through code.
- Go `1.23` in `go.mod`; standard library only; no CGO (`CGO_ENABLED=0` in release builds).
- Every script and program runs on macOS, Linux, and Windows (decision 0020); CI matrix covers all three; `.gitattributes` forces LF so golden files match on every checkout.
- Per-criterion results are exactly `PASS`, `FAIL`, `UNKNOWN`, `ERROR`, `NOT_APPLICABLE`; overall is `PASS`, `FAIL`, `INCOMPLETE`, `ADVISORY` (decision 0007; `ADVISORY` is a proposal, see Assumptions).
- A `PASS` requires at least one evidence record with `origin == observed` and a locator that is a file range, a command with exit status and log, or a passage with source and dates (decision 0008). The core rejects anything else.
- No composite score anywhere (decision 0002).
- JSON is canonical; Markdown, HTML, and SARIF add nothing (decision 0006). Section order is fixed by `report-schema.md` and observations precede results (decision 0013).
- HTML is one self-contained file, offline, no external scripts, styles, or fonts, light and dark (decision 0021).
- CLI exit codes: `0` ok, `1` report invalid or check failed, `2` the tool could not run (same contract as `tools/docs`).
- JSON tags snake_case; string enums with `IsValid()`; errors wrapped with `%w` and prefixed by package name; `schema_version` as a semver string.
- Strict TDD: every task writes the failing test first, runs it red, implements the minimum, runs it green, then commits. Functions under 50 lines. Table-driven tests, `t.Parallel()` where safe.
- Review cadence (maintainer rule: simplifier, code reviewer, architecture reviewer): after each task's green tests, run `code-simplifier` on the task's diff, then `go-reviewer`; run `architect` at the four checkpoints marked below and once more before the merge to `main`. Findings are fixed before the task's commit.
- Work happens on branch `feat/core-skeleton`; merge to `main` fast-forward only after the final review passes and CI is green; push to GitHub after merge.
- First execution step copies this plan into the repository as `docs/superpowers/plans/2026-09-14-core-walking-skeleton.md` (documentation rule) and commits it.

## Assumptions to confirm at approval

1. `ADVISORY` overall status exists (maps to evolve-loop `SKIPPED`); `status.advisory == true` selects it.
2. A fourth locator shape, `note`, exists for evidence that is not an opened file, command, or passage (candidate-supplied summaries, retrieved hints). It never qualifies for `PASS`.
3. Standard library only: validation is hand-written in Go and the schema file is kept equal by a drift test. The alternative, a JSON Schema library, is rejected to keep the binary dependency-free and portable.
4. `report_id` and `evidence_digest` are SHA-256 over this encoder's canonical form (struct field order, sorted map keys, two-space indent, LF, trailing newline). RFC 8785 canonicalization is a later minor change if a second implementation needs it.
5. Race-detector tests run on ubuntu and macos; Windows runs the plain suite (race needs a C toolchain there).
6. `ERROR` on a criterion requires an attempted check: at least one `execution` evidence or a `provenance.commands[]` entry; otherwise the assistant must use `UNKNOWN`.

---

## File Structure

```text
go.mod                                    module github.com/mickeyyaya/v-eval; go 1.23
.gitattributes                            * text=auto eol=lf
schema/report.schema.json                 JSON Schema 2020-12, v0.1.0 (the published contract)
schema/schema.go                          package schema: embed, Version = "0.1.0", Report() []byte
schema/schema_test.go
core/report/enums.go                      Result, Overall, Method, Kind, Origin, Isolation, ContractStatus, ClaimStatus, Severity, Disposition, Authority, ObservationOrigin + IsValid()
core/report/types.go                      Report and the 14 section structs, Evidence, Counts, Status
core/report/locator.go                    Locator, LocatorShape, Shape(), Qualifies(); Evidence.SupportsPass()
core/report/json.go                       Decode (DisallowUnknownFields), Encode (canonical), ReportID, EvidenceDigest
core/report/validate.go                   Validate, ValidateForAggregate, Violation, rule ids
core/report/aggregate.go                  Tally, Derive, Aggregate
core/report/schema_drift_test.go          Go struct tags == schema properties/required, enums == schema enums
core/report/testdata/worked-example.json  fixture from examples/code-review (C1-C3 FAIL, C4 PASS, C5 UNKNOWN, overall FAIL)
core/report/*_test.go
core/render/renderer.go                   Renderer interface; ByFormat(format) (Renderer, bool)
core/render/markdown.go                   text/template over templates/report.md.tmpl
core/render/html.go                       html/template over templates/report.html.tmpl, inline CSS, prefers-color-scheme
core/render/funcs.go                      template helpers: badge, joinIDs, supplied
core/render/templates/report.md.tmpl
core/render/templates/report.html.tmpl
core/render/testdata/worked-example.md    golden (regenerate: go test ./core/render -update)
core/render/testdata/worked-example.html  golden
core/render/*_test.go
core/export/sarif_types.go                minimal SARIF 2.1.0 structs used by the mapping
core/export/sarif.go                      ToSARIF(report.Report) (Log, error); Marshal(Log) ([]byte, error)
core/export/testdata/worked-example.sarif.json  golden
core/export/sarif_test.go
internal/version/version.go               Version, BuildDigest (ldflags); String()
cmd/veval/main.go                         os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
cmd/veval/cli.go                          run(), subcommand table, usage, exit codes
cmd/veval/cmd_validate.go  cmd_aggregate.go  cmd_render.go  cmd_export.go  cmd_version.go
cmd/veval/io.go                           readInput(path|"-"), writeOutput(-o|stdout)
cmd/veval/cli_test.go                     table tests over run() with the fixture
cmd/veval/skill_contract_test.go          SKILL.md names the real subcommands; host mapping files exist
.github/workflows/go.yml                  gofmt, vet, test on ubuntu/macos/windows; -race on ubuntu and macos
.goreleaser.yaml                          CGO_ENABLED=0; linux/darwin/windows x amd64/arm64; checksums.txt; veval_{os}_{arch}
skills/evaluate-output/SKILL.md           step 7 and host adaptation wired to veval
skills/evaluate-output/references/hosts/{claude-code,codex,gemini-cli,antigravity,hermes,ollama}.md
docs/superpowers/plans/2026-09-14-core-walking-skeleton.md   this plan, committed
CHANGELOG.md, README.md, docs/architecture/report-schema.md  status lines updated in the last task
```

## Shared definitions (used by every task; implementers see only their task)

### Enums (`core/report/enums.go`)

```go
package report

type Result string

const (
 ResultPass          Result = "PASS"
 ResultFail          Result = "FAIL"
 ResultUnknown       Result = "UNKNOWN"
 ResultError         Result = "ERROR"
 ResultNotApplicable Result = "NOT_APPLICABLE"
)

func (r Result) IsValid() bool {
 switch r {
 case ResultPass, ResultFail, ResultUnknown, ResultError, ResultNotApplicable:
  return true
 }
 return false
}
```

Same pattern, same constant naming `<Type><Value>`, for: `Overall` (`PASS`, `FAIL`, `INCOMPLETE`, `ADVISORY`), `Method` (`execution`, `deterministic_check`, `static_inspection`, `source_verification`, `rubric_judgment`, `human_judgment`), `Kind` (`execution`, `inspection`, `supplied`, `judgment`), `Origin` (`observed`, `candidate_supplied`, `retrieved`), `Isolation` (`none`, `worktree`, `container`, `remote_sandbox`), `ContractStatus` (`user_specified`, `provisional`, `approved`), `ClaimStatus` (`verified`, `contradicted`, `unverified`, `not_checkable`), `Severity` (`observed`, `suspicious`, `confirmed`), `Disposition` (`open`, `explained`, `confirmed`), `Authority` (`formal_standard`, `established_measure`, `vendor_rating`, `project_rubric`), `ObservationOrigin` (`assistant`, `adapter`, `detector`).

### Types (`core/report/types.go`)

```go
type Report struct {
 Identity     Identity          `json:"identity"`
 Contract     Contract          `json:"contract"`
 Routing      Routing           `json:"routing"`
 Observations []Observation     `json:"observations"`
 Claims       []Claim           `json:"claims"`
 Criteria     []CriterionResult `json:"criteria"`
 Forensics    []Finding         `json:"forensics"`
 Dimensions   []Dimension       `json:"dimensions,omitempty"`
 Counts       Counts            `json:"counts"`
 Status       Status            `json:"status"`
 Improvement  []Improvement     `json:"improvement"`
 Limitations  Limitations       `json:"limitations"`
 Provenance   Provenance        `json:"provenance"`
 Learning     *Learning         `json:"learning,omitempty"`
}

type Identity struct {
 ReportID      string   `json:"report_id"`
 SchemaVersion string   `json:"schema_version"`
 VevalVersion  string   `json:"veval_version"`
 SkillRevision string   `json:"skill_revision"`
 Artifact      Artifact `json:"artifact"`
 Task          Task     `json:"task"`
 CreatedAt     string   `json:"created_at"`
 Host          Host     `json:"host"`
}
type Artifact struct {
 Kind         string   `json:"kind"`
 Revision     string   `json:"revision"`
 Paths        []string `json:"paths"`
 BundleDigest string   `json:"bundle_digest"`
}
type Task struct {
 RequestedOutcome string `json:"requested_outcome"`
 IntendedUser     string `json:"intended_user"`
 BriefRef         string `json:"brief_ref"`
}
type Host struct {
 CLI   string `json:"cli"`
 Model string `json:"model,omitempty"`
 OS    string `json:"os"`
 Arch  string `json:"arch"`
}
type Contract struct {
 ContractID      string         `json:"contract_id"`
 ContractVersion string         `json:"contract_version"`
 Status          ContractStatus `json:"status"`
 Conflicts       []Conflict     `json:"conflicts"`
 Criteria        []Criterion    `json:"criteria"`
}
type Conflict struct {
 Between     []string `json:"between"`
 Description string   `json:"description"`
 Resolution  string   `json:"resolution"` // literal "unresolved" blocks acceptance
}
type Criterion struct {
 ID               string   `json:"id"`
 Requirement      string   `json:"requirement"`
 SourceRef        string   `json:"source_ref"`
 Required         bool     `json:"required"`
 Applicability    string   `json:"applicability"`
 MethodsAllowed   []Method `json:"methods_allowed"`
 AcceptanceRule   string   `json:"acceptance_rule"`
 ExpectedEvidence string   `json:"expected_evidence"`
 Provisional      bool     `json:"provisional"`
}
type Routing struct {
 Supplied        []SuppliedSource `json:"supplied"`
 Profiles        []ProfileRef     `json:"profiles"`
 AdaptersRun     []AdapterRun     `json:"adapters_run"`
 AdaptersSkipped []AdapterSkipped `json:"adapters_skipped"`
 Ambiguity       []Ambiguity      `json:"ambiguity"`
 Rationale       string           `json:"rationale"`
}
type SuppliedSource struct {
 SourceType string `json:"source_type"`
 Count      int    `json:"count"`
 Origin     string `json:"origin"`
}
type ProfileRef struct {
 Name    string `json:"name"`
 Version string `json:"version"`
}
type AdapterRun struct {
 Name       string   `json:"name"`
 Version    string   `json:"version"`
 InputsUsed []string `json:"inputs_used"`
}
type AdapterSkipped struct {
 Name         string `json:"name"`
 MissingInput string `json:"missing_input"`
}
type Ambiguity struct {
 Question   string `json:"question"`
 Resolution string `json:"resolution"`
}
type Observation struct {
 ID           string            `json:"id"`
 Text         string            `json:"text"`
 Evidence     []Evidence        `json:"evidence"`
 CriterionIDs []string          `json:"criterion_ids"`
 Origin       ObservationOrigin `json:"origin"`
}
type Claim struct {
 ClaimID      string      `json:"claim_id"`
 Text         string      `json:"text"`
 Location     string      `json:"location"`
 Status       ClaimStatus `json:"status"`
 Verification []Evidence  `json:"verification"`
 Notes        string      `json:"notes"`
}
type CriterionResult struct {
 ID              string     `json:"id"`
 Result          Result     `json:"result"`
 MethodUsed      Method     `json:"method_used"`
 Evidence        []Evidence `json:"evidence"`
 Reasoning       string     `json:"reasoning"`
 NextAction      string     `json:"next_action"`
 SharedCauseWith []string   `json:"shared_cause_with"`
 DimensionRefs   []string   `json:"dimension_refs"`
}
type Finding struct {
 FindingID         string      `json:"finding_id"`
 Detector          string      `json:"detector"`
 DetectorVersion   string      `json:"detector_version"`
 CriterionID       string      `json:"criterion_id"`
 Severity          Severity    `json:"severity"`
 Evidence          []Evidence  `json:"evidence"`
 BenignAlternative string      `json:"benign_alternative"`
 Disposition       Disposition `json:"disposition"`
}
type Dimension struct {
 Dimension       string     `json:"dimension"`
 Metric          string     `json:"metric"`
 MetricVersion   string     `json:"metric_version"`
 AuthorityType   Authority  `json:"authority_type"`
 DefinitionRef   string     `json:"definition_ref"`
 Unit            string     `json:"unit"`
 Range           string     `json:"range"`
 Direction       string     `json:"direction"`
 Value           *float64   `json:"value"` // null when missing; never zero for missing
 Threshold       string     `json:"threshold"`
 ThresholdSource string     `json:"threshold_source"`
 Tool            string     `json:"tool"`
 ToolVersion     string     `json:"tool_version"`
 Workload        string     `json:"workload"`
 Evidence        []Evidence `json:"evidence"`
 Interpretation  string     `json:"interpretation"`
}
type Counts struct {
 Required Tally    `json:"required"`
 Optional Tally    `json:"optional"`
 Coverage Coverage `json:"coverage"`
}
type Tally struct {
 Applicable    int `json:"applicable"`
 Pass          int `json:"pass"`
 Fail          int `json:"fail"`
 Unknown       int `json:"unknown"`
 Error         int `json:"error"`
 NotApplicable int `json:"not_applicable"`
}
type Coverage struct {
 Numerator   int  `json:"numerator"`
 Denominator int  `json:"denominator"`
 Undefined   bool `json:"undefined"`
}
type Status struct {
 Overall     Overall  `json:"overall"`
 RuleApplied string   `json:"rule_applied"`
 BlockedBy   []string `json:"blocked_by"`
 Advisory    bool     `json:"advisory"`
}
type Improvement struct {
 Issue                 string   `json:"issue"`
 Locations             []string `json:"locations"`
 SuggestedChange       string   `json:"suggested_change"`
 ConstraintsToPreserve []string `json:"constraints_to_preserve"`
 VerifyBy              string   `json:"verify_by"`
 LinkedCriteria        []string `json:"linked_criteria"`
}
type Limitations struct {
 NotInspected    []string `json:"not_inspected"`
 NotExecuted     []string `json:"not_executed"`
 UnknownMetadata []string `json:"unknown_metadata"`
 Assumptions     []string `json:"assumptions"`
}
type Provenance struct {
 Tools               []ToolRef       `json:"tools"`
 Commands            []CommandRecord `json:"commands"`
 Environment         Environment     `json:"environment"`
 IsolationLevelsUsed []Isolation     `json:"isolation_levels_used"`
 EvidenceDigest      string          `json:"evidence_digest"`
}
type ToolRef struct {
 Name    string `json:"name"`
 Version string `json:"version"`
}
type CommandRecord struct {
 Command    string    `json:"command"`
 Cwd        string    `json:"cwd"`
 ExitStatus int       `json:"exit_status"`
 StartedAt  string    `json:"started_at"`
 EndedAt    string    `json:"ended_at"`
 LogRef     string    `json:"log_ref"`
 Isolation  Isolation `json:"isolation"`
}
type Environment struct {
 OS              string            `json:"os"`
 Arch            string            `json:"arch"`
 RuntimeVersions map[string]string `json:"runtime_versions"`
}
type Learning struct {
 PrecedentsRetrieved  []PrecedentRef `json:"precedents_retrieved"`
 RewardRecordsCreated []string       `json:"reward_records_created"`
}
type PrecedentRef struct {
 PrecedentID string  `json:"precedent_id"`
 Criterion   string  `json:"criterion"`
 Similarity  float64 `json:"similarity"`
}
type Evidence struct {
 Kind          Kind               `json:"kind"`
 Locator       Locator            `json:"locator"`
 Observation   string             `json:"observation"`
 Provenance    EvidenceProvenance `json:"provenance"`
 Isolation     Isolation          `json:"isolation"`
 Origin        Origin             `json:"origin"`
 RubricVersion string             `json:"rubric_version,omitempty"` // required when Kind == judgment
 Model         string             `json:"model,omitempty"`          // "unknown" is a valid value
}
type EvidenceProvenance struct {
 Tool      string `json:"tool"`
 Version   string `json:"version"`
 Revision  string `json:"revision"`
 Timestamp string `json:"timestamp"`
}
```

### Locator (`core/report/locator.go`)

```go
type Locator struct {
 File      string `json:"file,omitempty"`
 LineStart int    `json:"line_start,omitempty"`
 LineEnd   int    `json:"line_end,omitempty"`

 Command    string `json:"command,omitempty"`
 Cwd        string `json:"cwd,omitempty"`
 ExitStatus *int   `json:"exit_status,omitempty"`
 LogRef     string `json:"log_ref,omitempty"`

 Passage             string `json:"passage,omitempty"`
 SourceRef           string `json:"source_ref,omitempty"`
 SourceDateOrVersion string `json:"source_date_or_version,omitempty"`
 AccessDate          string `json:"access_date,omitempty"`

 Note string `json:"note,omitempty"` // non-qualifying pointer, e.g. "PR description, paragraph 2"
}

type LocatorShape string

const (
 ShapeFile    LocatorShape = "file"
 ShapeCommand LocatorShape = "command"
 ShapePassage LocatorShape = "passage"
 ShapeNote    LocatorShape = "note"
 ShapeNone    LocatorShape = ""
)

// Shape returns the one complete shape; partial or mixed fields yield ShapeNone.
// file: File != "" && LineStart >= 1 && LineEnd >= LineStart, no other group set.
// command: Command, Cwd, LogRef != "" && ExitStatus != nil, no other group set.
// passage: all four passage fields != "", no other group set.
// note: Note != "" and nothing else.
func (l Locator) Shape() LocatorShape

// Qualifies is true for file, command, and passage: what the evaluator opened or ran.
func (l Locator) Qualifies() bool

// SupportsPass is true only for observed evidence with a qualifying locator.
func (e Evidence) SupportsPass() bool { return e.Origin == OriginObserved && e.Locator.Qualifies() }
```

### Validation and aggregation (`core/report/validate.go`, `aggregate.go`)

```go
type Violation struct {
 Path    string // e.g. "criteria[3].evidence"
 Rule    string // stable id, e.g. "evidence.pass_requires_observed_locator"
 Message string
}

func Decode(raw []byte) (Report, error)              // json.Decoder with DisallowUnknownFields
func Encode(rep Report) ([]byte, error)              // canonical: MarshalIndent two spaces + "\n"; maps sorted by encoding/json
func ReportID(rep Report) string                     // "sha256:" + hex(sha256(Encode(rep with Identity.ReportID = "")))
func EvidenceDigest(rep Report) string               // "sha256:" + hex over canonical JSON of [criteria, observations, claims, forensics, dimensions] evidence arrays in document order
func Validate(raw []byte) (Report, []Violation, error)            // all rules; err only for internal failures
func ValidateForAggregate(raw []byte) (Report, []Violation, error) // all rules except counts.match, status.match, identity.report_id, provenance.evidence_digest
func Tally(contract Contract, results []CriterionResult) Counts
func Derive(contract Contract, results []CriterionResult, advisory bool) Status
func Aggregate(rep Report) Report                    // copy with Counts, Status, Provenance.EvidenceDigest, Identity.ReportID filled
```

Amended during execution: the counting function is `TallyCounts`, not `Tally`, because `Tally` is already the name of the per-category count struct that the returned `Counts` holds two of. `Encode` also turns HTML escaping off (`SetEscapeHTML(false)`), so a `<`, `>`, or `&` inside an observation survives the round trip instead of becoming `\u003c`; the shared encoder is exported as `CanonicalJSON`. `EvidenceDigest` hashes each evidence array tagged with its section name rather than a bare list of lists, so moving an evidence record between sections changes the digest.

Rule ids and rules (each is one small function returning `[]Violation`):

- `json`: raw is not valid JSON, or has unknown fields (path `$`).
- `schema_version.supported`: `identity.schema_version == schema.Version`.
- `required.nonempty`: non-empty `identity.{veval_version,skill_revision,artifact.kind,artifact.revision,task.requested_outcome}`, `contract.{contract_id,contract_version}`, `routing.rationale`, every `contract.criteria[].{id,requirement,acceptance_rule}` and `methods_allowed` non-empty, every `criteria[].id`, every `observations[].{id,text}`, every `claims[].{claim_id,text}`, every `forensics[].finding_id`, every evidence `observation`, `status.rule_applied`.
- `enum.valid`: every enum-typed field passes `IsValid()`; `contract.conflicts[].resolution` non-empty.
- `time.rfc3339`: `identity.created_at` parses with `time.RFC3339`.
- `locator.shape`: every evidence locator has `Shape() != ShapeNone`.
- `criteria.contract_link`: every `criteria[].id` exists in the contract; every contract criterion has exactly one result; no duplicates.
- `evidence.pass_requires_observed_locator`: `result == PASS` needs one evidence with `SupportsPass()`.
- `evidence.candidate_supplied_kind`: `origin == candidate_supplied` requires `kind == supplied`.
- `evidence.judgment_metadata`: `kind == judgment` requires non-empty `rubric_version`.
- `criteria.not_applicable_reasoning`: `NOT_APPLICABLE` requires non-empty `reasoning`.
- `criteria.error_is_operational`: `ERROR` requires one evidence with `kind == execution` or one `provenance.commands[]` with non-zero `exit_status`.
- `claims.verification_present`: `verified` or `contradicted` require non-empty `verification`.
- `forensics.criterion_link`: `criterion_id` exists in the contract; `disposition == confirmed` requires `severity == confirmed`.
- `dimensions.no_composite`: no dimension whose lowercased `metric` is `composite`, `overall`, `total`, or `score`.
- `counts.match`, `status.match`, `provenance.evidence_digest`, `identity.report_id`: equal to the recomputed values.

Tally: applicable means `result != NOT_APPLICABLE`; per bucket (required or optional from the contract) count applicable, pass, fail, unknown, error, not_applicable; coverage numerator `= required.pass + required.fail + optional.pass + optional.fail`, denominator `= required.applicable + optional.applicable`, `undefined = denominator == 0`.

Derive, on gating set G = required and applicable and not provisional, in this order:

1. `advisory` → `ADVISORY`, rule `advisory: acceptance not requested; per-criterion results stand`, `blocked_by` = FAIL ids in G (sorted).
2. any FAIL in G → `FAIL`, rule `required applicable criterion failed`, `blocked_by` = those ids.
3. any required applicable criterion with `provisional` → `INCOMPLETE`, rule `provisional required criterion prevents acceptance`, `blocked_by` = those ids.
4. any conflict with `resolution == "unresolved"` → `INCOMPLETE`, rule `unresolved contract conflict`.
5. any UNKNOWN or ERROR in G → `INCOMPLETE`, rule `required applicable criterion unknown or error`, `blocked_by` = those ids.
6. G empty → `INCOMPLETE`, rule `no required applicable criteria`.
7. else `PASS`, rule `all required applicable criteria pass`.

`blocked_by` is always non-nil (`[]string{}` when empty) so canonical JSON is stable.

### CLI contract (`cmd/veval`)

```text
veval version                                       "veval <Version> (<BuildDigest>) schema <schema.Version>"        exit 0
veval validate <report.json|->                      "valid: <name>" exit 0; else "<path>: <rule>: <message>" per line, exit 1
veval aggregate <report.json|-> [-o out]            ValidateForAggregate (exit 1 on violations); Aggregate; canonical JSON to -o or stdout; exit 0
veval render <report.json|-> --format md|html [-o out]   Validate (exit 1); render; bytes to -o or stdout
veval export sarif <report.json|-> [-o out]         Validate (exit 1); SARIF JSON to -o or stdout
```

Amended during execution: `version` prints one line, `veval <Version> (<BuildDigest>) schema <schema.Version>` (`veval 0.0.0-dev (unknown) schema 0.1.0` from an unstamped build). `-` is the stream sentinel in both directions, so `-o -` writes the document to standard output instead of creating a file named `-`, and `--` ends flag parsing, so `veval validate -- -weird.json` reads that file.

Exit `2` with `error: <message>` on stderr for unknown subcommand or flag, missing argument, unreadable input, unwritable output, unknown format, internal failure. `-` reads stdin. Output is LF, no color, no prompts. `run(args []string, stdin io.Reader, stdout, stderr io.Writer) int` is the testable seam; `main` only calls it. Subcommands are a table `[]subcommand{Name, Summary, Run}`; flags use `flag.NewFlagSet(name, flag.ContinueOnError)` with output set to `stderr`; flags may follow positionals (reorder before parsing).

---

## Task 0: Branch and commit the plan

**Files:**

- Create: `docs/superpowers/plans/2026-09-14-core-walking-skeleton.md` (this document)

- [ ] **Step 1:** `git checkout -b feat/core-skeleton`
- [ ] **Step 2:** Copy the plan file into the repository path above.
- [ ] **Step 3:** `python3 tools/docs/vdocs.py check-links .` must print `0 broken`; `npx --yes markdownlint-cli2 "**/*.md" "!node_modules"` must report no errors.
- [ ] **Step 4:** Commit: `docs(plan): core walking skeleton implementation plan`

## Task 1: Module, line endings, embedded schema

**Files:**

- Create: `go.mod`, `.gitattributes`, `schema/report.schema.json`, `schema/schema.go`, `schema/schema_test.go`

**Interfaces:**

- Produces: `schema.Version` (`const string = "0.1.0"`), `schema.Report() []byte`.

- [ ] **Step 1: Write the failing test**

```go
package schema_test

import (
 "encoding/json"
 "testing"

 "github.com/mickeyyaya/v-eval/schema"
)

func TestEmbeddedSchemaParsesAndDeclaresVersion(t *testing.T) {
 t.Parallel()
 var doc map[string]any
 if err := json.Unmarshal(schema.Report(), &doc); err != nil {
  t.Fatalf("schema is not valid JSON: %v", err)
 }
 if got := doc["$id"]; got != "https://github.com/mickeyyaya/v-eval/schema/report/"+schema.Version {
  t.Fatalf("$id = %v, want it to end with %s", got, schema.Version)
 }
 if schema.Version != "0.1.0" {
  t.Fatalf("Version = %q", schema.Version)
 }
}
```

- [ ] **Step 2: Run it red.** `go test ./schema/` → FAIL (package missing).
- [ ] **Step 3: Implement.** `go.mod` with `module github.com/mickeyyaya/v-eval` and `go 1.23`. `.gitattributes`: `* text=auto eol=lf`. `schema/schema.go`:

```go
// Package schema embeds the published report contract. The Go types in core/report are the
// source of truth; core/report/schema_drift_test.go keeps this file equal to them.
package schema

import _ "embed"

// Version is the report schema version every valid report declares.
const Version = "0.1.0"

//go:embed report.schema.json
var raw []byte

// Report returns the JSON Schema document.
func Report() []byte { return raw }
```

`schema/report.schema.json`: the JSON Schema 2020-12 document in the appendix of this plan, verbatim.

- [ ] **Step 4: Run it green.** `go test ./schema/` → PASS. `gofmt -l .` empty.
- [ ] **Step 5: Commit:** `feat(schema): embed report schema v0.1.0`

## Task 2: Enums

**Files:**

- Create: `core/report/enums.go`, `core/report/enums_test.go`

**Interfaces:**

- Produces: all enum types and `IsValid()` methods listed in Shared definitions.

- [ ] **Step 1: Write the failing test**

```go
package report

import "testing"

func TestEnumValidators(t *testing.T) {
 t.Parallel()
 cases := []struct {
  name string
  ok   bool
  fn   func() bool
 }{
  {"PASS", true, Result("PASS").IsValid},
  {"pass", false, Result("pass").IsValid},
  {"INCOMPLETE", true, Overall("INCOMPLETE").IsValid},
  {"ADVISORY", true, Overall("ADVISORY").IsValid},
  {"remote_sandbox", true, Isolation("remote_sandbox").IsValid},
  {"docker", false, Isolation("docker").IsValid},
  {"candidate_supplied", true, Origin("candidate_supplied").IsValid},
  {"rubric_judgment", true, Method("rubric_judgment").IsValid},
  {"not_checkable", true, ClaimStatus("not_checkable").IsValid},
  {"project_rubric", true, Authority("project_rubric").IsValid},
  {"detector", true, ObservationOrigin("detector").IsValid},
  {"", false, Kind("").IsValid},
 }
 for _, c := range cases {
  if c.fn() != c.ok {
   t.Errorf("%s: IsValid = %v, want %v", c.name, !c.ok, c.ok)
  }
 }
}
```

- [ ] **Step 2: Run red.** `go test ./core/report/` → FAIL (undefined types).
- [ ] **Step 3: Implement** every enum exactly as in Shared definitions.
- [ ] **Step 4: Run green.**
- [ ] **Step 5: Commit:** `feat(report): verdict, method, evidence, and isolation enums`

## Task 3: Locator shapes

**Files:**

- Create: `core/report/locator.go`, `core/report/locator_test.go`
- Modify: `core/report/types.go` is created in Task 4; `Evidence.SupportsPass` lives in `locator.go` and needs `Evidence`, so define a minimal `Evidence` struct here now and move nothing later (Task 4 completes the other types in `types.go`).

- [ ] **Step 1: Write the failing test**

```go
package report

import "testing"

func TestLocatorShapes(t *testing.T) {
 t.Parallel()
 zero := 0
 cases := map[string]struct {
  loc  Locator
  want LocatorShape
 }{
  "file":            {Locator{File: "a.go", LineStart: 3, LineEnd: 9}, ShapeFile},
  "file bad range":  {Locator{File: "a.go", LineStart: 9, LineEnd: 3}, ShapeNone},
  "file no lines":   {Locator{File: "a.go"}, ShapeNone},
  "command":         {Locator{Command: "go test ./...", Cwd: "/w", ExitStatus: &zero, LogRef: "logs/1.txt"}, ShapeCommand},
  "command no exit": {Locator{Command: "go test", Cwd: "/w", LogRef: "l"}, ShapeNone},
  "passage":         {Locator{Passage: "Export requires a connection.", SourceRef: "guide.md", SourceDateOrVersion: "v3", AccessDate: "2026-09-14"}, ShapePassage},
  "passage no date": {Locator{Passage: "x", SourceRef: "guide.md"}, ShapeNone},
  "note":            {Locator{Note: "PR description"}, ShapeNote},
  "mixed":           {Locator{File: "a.go", LineStart: 1, LineEnd: 1, Note: "x"}, ShapeNone},
  "empty":           {Locator{}, ShapeNone},
 }
 for name, c := range cases {
  if got := c.loc.Shape(); got != c.want {
   t.Errorf("%s: Shape = %q, want %q", name, got, c.want)
  }
 }
}

func TestSupportsPassNeedsObservedQualifyingLocator(t *testing.T) {
 t.Parallel()
 file := Locator{File: "a.go", LineStart: 1, LineEnd: 1}
 if !(Evidence{Origin: OriginObserved, Locator: file}).SupportsPass() {
  t.Fatal("observed file evidence must support PASS")
 }
 if (Evidence{Origin: OriginCandidateSupplied, Locator: file}).SupportsPass() {
  t.Fatal("candidate-supplied evidence must never support PASS")
 }
 if (Evidence{Origin: OriginObserved, Locator: Locator{Note: "x"}}).SupportsPass() {
  t.Fatal("note locator must never support PASS")
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `Locator`, `LocatorShape`, `Shape()` (one helper per group: `fileSet()`, `commandSet()`, `passageSet()`, `noteSet()` returning whether any field of that group is set and whether the group is complete; `Shape` returns the shape when exactly one group is set and complete), `Qualifies()`, and the `Evidence` struct with `SupportsPass()`.
- [ ] **Step 4: Run green.**
- [ ] **Step 5: Commit:** `feat(report): locator shapes and the PASS-qualifying rule`

## Task 4: Report types, canonical JSON, worked-example fixture

**Files:**

- Create: `core/report/types.go`, `core/report/json.go`, `core/report/testdata/worked-example.json`, `core/report/json_test.go`, `core/report/testhelpers_test.go`

**Interfaces:**

- Produces: all section types; `Decode`, `Encode`; `loadFixture(t) Report` test helper.

Fixture content (`testdata/worked-example.json`), derived from `examples/code-review/input.md` and `report.md`:

- `identity`: `report_id` filled later by Task 8 (write `""` now), `schema_version` `0.1.0`, `veval_version` `0.0.0-dev`, `skill_revision` `draft-2`, artifact `kind` `code_change`, `revision` `content:sha256:` + the hex sha256 of the exact Python snippet in `input.md` (compute with `shasum -a 256`), `paths` `["examples/code-review/input.md"]`, `bundle_digest` `""`; task `requested_outcome` "Deduplicate email-address strings for a contact import…" (copy the Intent paragraph), `intended_user` "developer reviewing an AI-generated change", `brief_ref` `examples/code-review/input.md#intent`; `created_at` `2026-09-14T00:00:00Z`; host `cli` `fixture`, `os` `any`, `arch` `any`.
- `contract`: `contract_id` `code-review-example`, `contract_version` `1`, `status` `user_specified`, `conflicts` `[]`, five criteria copied from the input's table: C1 to C4 `methods_allowed` `["static_inspection","execution"]`, C5 `["execution"]`, all `required: true`, `provisional: false`, `applicability` `"always"`, `source_ref` `examples/code-review/input.md#acceptance-criteria`.
- `routing`: `supplied` `[{"source_type":"brief","count":1,"origin":"user"},{"source_type":"artifact","count":1,"origin":"candidate"},{"source_type":"candidate_claim","count":1,"origin":"candidate"}]`, `profiles` `[{"name":"requirements-and-tests","version":"0.1"}]`, `adapters_run` `[]`, `adapters_skipped` `[{"name":"test-evidence","missing_input":"test definitions and execution record"}]`, `ambiguity` `[]`, `rationale` "Code change with explicit criteria; no tests or execution record supplied, so execution criteria stay UNKNOWN."
- `observations`: O1 "The function sorts its output with sorted(), so first-seen order cannot be preserved." with an observed inspection file-locator evidence on `examples/code-review/input.md` lines of the code block (find with `grep -n "return sorted" examples/code-review/input.md`), `criterion_ids` `["C2"]`, origin `assistant`.
- `claims`: one claim "The code is production-ready and all tests passed." location "input.md, Candidate author's statement", status `unverified`, verification `[]`, notes "No tests, command, environment, or logs accompany the statement."
- `criteria`: C1 FAIL, C2 FAIL, C3 FAIL, each `method_used` `static_inspection`, evidence one observed `inspection` file locator on the code line with `observation` quoting the line, `provenance` `{"tool":"assistant","version":"fixture","revision":"content:sha256:…","timestamp":"2026-09-14T00:00:00Z"}`, `isolation` `none`, reasoning copied from `report.md`, `next_action` copied; C4 PASS with the same file locator and observation "The complete supplied function uses built-ins and a string method only."; C5 UNKNOWN with one `supplied` evidence, origin `candidate_supplied`, locator `{"note":"Candidate author's statement in input.md"}`, observation "Claims tests passed; no tests, command, logs, or verified artifact binding supplied.", reasoning and next_action from `report.md`; every criterion `shared_cause_with` `[]` (C1 and C3 may list each other), `dimension_refs` `[]`.
- `forensics` `[]`, no `dimensions`, `counts` and `status` left as the values Task 8's Aggregate computes (write them by hand now: required applicable 5, pass 1, fail 3, unknown 1; coverage 4/5; overall FAIL, rule "required applicable criterion failed", blocked_by `["C1","C2","C3"]`, advisory false).
- `improvement`: three entries, one per FAIL, taken from `report.md` next actions, `linked_criteria` accordingly.
- `limitations`: `not_inspected` `[]`, `not_executed` `["regression suite"]`, `unknown_metadata` `["test environment"]`, `assumptions` `[]`.
- `provenance`: `tools` `[{"name":"veval","version":"0.0.0-dev"}]`, `commands` `[]`, `environment` `{"os":"any","arch":"any","runtime_versions":{}}`, `isolation_levels_used` `["none"]`, `evidence_digest` `""` (Task 8 fills; the Task 4 test therefore compares canonical form after blanking the two digests).

- [ ] **Step 1: Write the failing test**

```go
package report

import (
 "bytes"
 "os"
 "testing"
)

func loadFixture(t *testing.T) Report {
 t.Helper()
 raw, err := os.ReadFile("testdata/worked-example.json")
 if err != nil {
  t.Fatal(err)
 }
 rep, err := Decode(raw)
 if err != nil {
  t.Fatalf("decode fixture: %v", err)
 }
 return rep
}

func TestFixtureDecodesAndEncodesCanonically(t *testing.T) {
 t.Parallel()
 rep := loadFixture(t)
 if len(rep.Criteria) != 5 || rep.Criteria[3].ID != "C4" || rep.Criteria[3].Result != ResultPass || rep.Criteria[4].Result != ResultUnknown {
  t.Fatalf("unexpected criteria: %+v", rep.Criteria)
 }
 once, err := Encode(rep)
 if err != nil {
  t.Fatal(err)
 }
 again, _ := Encode(rep)
 if !bytes.Equal(once, again) {
  t.Fatal("Encode is not deterministic")
 }
 raw, _ := os.ReadFile("testdata/worked-example.json")
 if !bytes.Equal(once, raw) {
  t.Fatalf("fixture is not in canonical form; write Encode output back to testdata")
 }
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
 t.Parallel()
 if _, err := Decode([]byte(`{"identity":{"unexpected":1}}`)); err == nil {
  t.Fatal("unknown field must be rejected")
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `types.go` (all structs from Shared definitions except `Evidence`, which Task 3 created), `json.go`:

```go
func Decode(raw []byte) (Report, error) {
 dec := json.NewDecoder(bytes.NewReader(raw))
 dec.DisallowUnknownFields()
 var rep Report
 if err := dec.Decode(&rep); err != nil {
  return Report{}, fmt.Errorf("report: decode: %w", err)
 }
 return rep, nil
}

func Encode(rep Report) ([]byte, error) {
 out, err := json.MarshalIndent(rep, "", "  ")
 if err != nil {
  return nil, fmt.Errorf("report: encode: %w", err)
 }
 return append(out, '\n'), nil
}
```

Write the fixture, run `go run ./cmd/…` is not available yet, so produce canonical form with a one-off `go test -run TestFixtureDecodesAndEncodesCanonically` failure message and paste `Encode` output back until the byte comparison passes (or a throwaway `main` under `/tmp`, not committed).

- [ ] **Step 4: Run green.**
- [ ] **Step 5: Commit:** `feat(report): report types, canonical encoding, worked-example fixture`

**Checkpoint A (after Task 4): run `architect` on `core/report` types and fixture.**

## Task 5: Tally and coverage

**Files:**

- Create: `core/report/aggregate.go` (Tally only), `core/report/tally_test.go`

- [ ] **Step 1: Write the failing test**

```go
package report

import "testing"

func TestTallyWorkedExample(t *testing.T) {
 t.Parallel()
 rep := loadFixture(t)
 got := Tally(rep.Contract, rep.Criteria)
 want := Counts{
  Required: Tally{Applicable: 5, Pass: 1, Fail: 3, Unknown: 1},
  Coverage: Coverage{Numerator: 4, Denominator: 5},
 }
 if got != want {
  t.Fatalf("Tally = %+v, want %+v", got, want)
 }
}

func TestTallySeparatesOptionalAndExcludesNotApplicable(t *testing.T) {
 t.Parallel()
 contract := Contract{Criteria: []Criterion{{ID: "R1", Required: true}, {ID: "R2", Required: true}, {ID: "O1"}}}
 results := []CriterionResult{{ID: "R1", Result: ResultPass}, {ID: "R2", Result: ResultNotApplicable}, {ID: "O1", Result: ResultError}}
 got := Tally(contract, results)
 if got.Required != (Tally{Applicable: 1, Pass: 1, NotApplicable: 1}) || got.Optional != (Tally{Applicable: 1, Error: 1}) {
  t.Fatalf("got %+v", got)
 }
 if got.Coverage != (Coverage{Numerator: 1, Denominator: 2}) {
  t.Fatalf("coverage = %+v", got.Coverage)
 }
}

func TestCoverageUndefinedWithZeroApplicable(t *testing.T) {
 t.Parallel()
 contract := Contract{Criteria: []Criterion{{ID: "C1", Required: true}}}
 got := Tally(contract, []CriterionResult{{ID: "C1", Result: ResultNotApplicable}})
 if !got.Coverage.Undefined || got.Coverage.Denominator != 0 || got.Required.NotApplicable != 1 {
  t.Fatalf("coverage = %+v", got.Coverage)
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `Tally` with a `requiredByID` map from the contract and a `bump(*Tally, Result)` helper.
- [ ] **Step 4: Run green.**
- [ ] **Step 5: Commit:** `feat(report): counts and assessment coverage`

## Task 6: Status derivation

**Files:**

- Modify: `core/report/aggregate.go` (add `Derive` and rule constants)
- Create: `core/report/derive_test.go`

- [ ] **Step 1: Write the failing test**

```go
package report

import (
 "reflect"
 "testing"
)

func TestDerive(t *testing.T) {
 t.Parallel()
 req := func(id string, provisional bool) Criterion { return Criterion{ID: id, Required: true, Provisional: provisional} }
 opt := func(id string) Criterion { return Criterion{ID: id} }
 res := func(id string, r Result) CriterionResult { return CriterionResult{ID: id, Result: r} }
 cases := []struct {
  name      string
  contract  Contract
  results   []CriterionResult
  advisory  bool
  want      Overall
  blockedBy []string
 }{
  {"fail dominates", Contract{Criteria: []Criterion{req("A", false), req("B", false)}}, []CriterionResult{res("A", ResultFail), res("B", ResultUnknown)}, false, OverallFail, []string{"A"}},
  {"unknown incomplete", Contract{Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultUnknown)}, false, OverallIncomplete, []string{"A"}},
  {"error incomplete", Contract{Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultError)}, false, OverallIncomplete, []string{"A"}},
  {"all pass", Contract{Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultPass)}, false, OverallPass, []string{}},
  {"provisional blocks", Contract{Criteria: []Criterion{req("A", true)}}, []CriterionResult{res("A", ResultPass)}, false, OverallIncomplete, []string{"A"}},
  {"unresolved conflict", Contract{Conflicts: []Conflict{{Resolution: "unresolved"}}, Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultPass)}, false, OverallIncomplete, []string{}},
  {"zero gating", Contract{Criteria: []Criterion{opt("A")}}, []CriterionResult{res("A", ResultPass)}, false, OverallIncomplete, []string{}},
  {"optional fail ignored", Contract{Criteria: []Criterion{req("A", false), opt("B")}}, []CriterionResult{res("A", ResultPass), res("B", ResultFail)}, false, OverallPass, []string{}},
  {"not applicable excluded", Contract{Criteria: []Criterion{req("A", false), req("B", false)}}, []CriterionResult{res("A", ResultPass), res("B", ResultNotApplicable)}, false, OverallPass, []string{}},
  {"advisory", Contract{Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultFail)}, true, OverallAdvisory, []string{"A"}},
  {"blocked_by sorted", Contract{Criteria: []Criterion{req("B", false), req("A", false)}}, []CriterionResult{res("B", ResultFail), res("A", ResultFail)}, false, OverallFail, []string{"A", "B"}},
 }
 for _, c := range cases {
  got := Derive(c.contract, c.results, c.advisory)
  if got.Overall != c.want || !reflect.DeepEqual(got.BlockedBy, c.blockedBy) || got.RuleApplied == "" || got.Advisory != c.advisory {
   t.Errorf("%s: got %+v, want %s blocked_by %v", c.name, got, c.want, c.blockedBy)
  }
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `Derive` as the seven ordered rules; keep each rule check in its own small function returning `(ids []string, hit bool)`; rule strings as constants `RuleAdvisory`, `RuleFailed`, `RuleProvisional`, `RuleConflict`, `RuleUnknownOrError`, `RuleNoGating`, `RuleAllPass`.
- [ ] **Step 4: Run green.**
- [ ] **Step 5: Commit:** `feat(report): overall status derivation with auditable rule`

## Task 7: Validation rules and schema drift test

**Files:**

- Create: `core/report/validate.go`, `core/report/validate_test.go`, `core/report/schema_drift_test.go`

**Interfaces:**

- Produces: `Violation`, `Validate`, `ValidateForAggregate`, rule id constants (`RuleJSON = "json"` etc. as listed in Shared definitions).

- [ ] **Step 1: Write the failing tests**

```go
package report

import (
 "os"
 "testing"
)

func hasRule(vs []Violation, rule string) bool {
 for _, v := range vs {
  if v.Rule == rule {
   return true
  }
 }
 return false
}

func TestValidateAcceptsFixture(t *testing.T) {
 t.Parallel()
 raw, _ := os.ReadFile("testdata/worked-example.json")
 _, violations, err := Validate(raw)
 if err != nil || len(violations) != 0 {
  t.Fatalf("err=%v violations=%v", err, violations)
 }
}

func TestValidateReportsMalformedJSONAsViolation(t *testing.T) {
 t.Parallel()
 _, violations, err := Validate([]byte("{not json"))
 if err != nil || len(violations) != 1 || violations[0].Rule != "json" || violations[0].Path != "$" {
  t.Fatalf("err=%v violations=%v", err, violations)
 }
}

func TestValidateRejectsPassWithoutObservedLocator(t *testing.T) {
 t.Parallel()
 rep := loadFixture(t)
 rep.Criteria[3].Evidence = []Evidence{{Kind: KindSupplied, Origin: OriginCandidateSupplied, Locator: Locator{Note: "author says so"}, Observation: "tests passed", Isolation: IsolationNone}}
 rep = Aggregate(rep)
 raw, _ := Encode(rep)
 _, violations, _ := Validate(raw)
 if !hasRule(violations, "evidence.pass_requires_observed_locator") {
  t.Fatalf("violations = %v", violations)
 }
}

func TestValidateCatchesInventedStatusAndBrokenLinks(t *testing.T) {
 t.Parallel()
 rep := loadFixture(t)
 rep.Status.Overall = OverallPass
 raw, _ := Encode(rep)
 _, violations, _ := Validate(raw)
 if !hasRule(violations, "status.match") || !hasRule(violations, "identity.report_id") {
  t.Fatalf("violations = %v", violations)
 }
 rep = loadFixture(t)
 rep.Criteria[0].ID = "C9"
 raw, _ = Encode(Aggregate(rep))
 _, violations, _ = Validate(raw)
 if !hasRule(violations, "criteria.contract_link") {
  t.Fatalf("violations = %v", violations)
 }
}

func TestValidateForAggregateSkipsDerivedFields(t *testing.T) {
 t.Parallel()
 rep := loadFixture(t)
 rep.Counts, rep.Status, rep.Identity.ReportID, rep.Provenance.EvidenceDigest = Counts{}, Status{}, "", ""
 raw, _ := Encode(rep)
 _, violations, err := ValidateForAggregate(raw)
 if err != nil || len(violations) != 0 {
  t.Fatalf("err=%v violations=%v", err, violations)
 }
}

func TestValidateErrorRequiresAttemptedCheck(t *testing.T) {
 t.Parallel()
 rep := loadFixture(t)
 rep.Criteria[4].Result = ResultError
 raw, _ := Encode(Aggregate(rep))
 _, violations, _ := Validate(raw)
 if !hasRule(violations, "criteria.error_is_operational") {
  t.Fatalf("violations = %v", violations)
 }
}
```

Note: `Aggregate` is implemented in Task 8; in this task write it minimally (fill `Counts` and `Status` via `Tally` and `Derive`, leave digests empty) so these tests compile, and let Task 8 complete it. `TestValidateAcceptsFixture` and the report_id rule will fail until Task 8; mark those two with `t.Skip("until Task 8")` in this task and remove the skips in Task 8.

Schema drift test (`schema_drift_test.go`): parse `schema.Report()`; for each top-level property and each `$defs` object with `properties`, resolve the Go type by a fixed map (`"identity" → Identity{}`, `"contract" → Contract{}`, … , `"evidence" → Evidence{}`, `"locator" → Locator{}`, `"criterion" → Criterion{}`, `"tally" → Tally{}`) and assert with `reflect` that the set of `json` tag names on the struct equals the set of schema property names, and that every schema `required` name is a tag without `omitempty`; for each enum type assert the schema `enum` list equals the Go constants (build the Go list from a `validValues()` helper per enum). The `locator` `oneOf` is checked by union of its four property sets.

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `validate.go`: `Validate` = `runRules(raw, allRules)`, `ValidateForAggregate` = `runRules(raw, structuralRules)`; each rule a `func(rep Report) []Violation` under 50 lines; JSON decode failure returns the single `json` violation. Add `validValues()` on each enum for the drift test.
- [ ] **Step 4: Run green** (with the two skips).
- [ ] **Step 5: Commit:** `feat(report): validation rules and schema drift test`

## Task 8: Report id, evidence digest, Aggregate

**Files:**

- Modify: `core/report/json.go` (ReportID, EvidenceDigest), `core/report/aggregate.go` (Aggregate), `core/report/validate_test.go` (remove skips), `core/report/testdata/worked-example.json` (fill the two digests from Aggregate output)
- Create: `core/report/aggregate_test.go`

- [ ] **Step 1: Write the failing test**

```go
package report

import (
 "strings"
 "testing"
)

func TestAggregateFillsDerivedFieldsAndIsIdempotent(t *testing.T) {
 t.Parallel()
 rep := loadFixture(t)
 rep.Counts, rep.Status, rep.Identity.ReportID, rep.Provenance.EvidenceDigest = Counts{}, Status{}, "", ""
 first := Aggregate(rep)
 if first.Status.Overall != OverallFail || first.Counts.Required.Fail != 3 {
  t.Fatalf("aggregate: %+v %+v", first.Status, first.Counts)
 }
 if !strings.HasPrefix(first.Identity.ReportID, "sha256:") || !strings.HasPrefix(first.Provenance.EvidenceDigest, "sha256:") {
  t.Fatalf("digests: %q %q", first.Identity.ReportID, first.Provenance.EvidenceDigest)
 }
 if second := Aggregate(first); second.Identity.ReportID != first.Identity.ReportID {
  t.Fatal("Aggregate is not idempotent")
 }
 changed := first
 changed.Criteria = append([]CriterionResult(nil), first.Criteria...)
 changed.Criteria[3].Result = ResultFail
 if Aggregate(changed).Identity.ReportID == first.Identity.ReportID {
  t.Fatal("report_id must change when a result changes")
 }
 if Aggregate(changed).Provenance.EvidenceDigest != first.Provenance.EvidenceDigest {
  t.Fatal("evidence digest must not change when only a result changes")
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `ReportID` (copy, blank `Identity.ReportID`, `Encode`, sha256), `EvidenceDigest` (collect evidence arrays in document order into a slice of slices, `json.Marshal`, sha256), `Aggregate` (copy; `Counts = Tally`; `Status = Derive(..., rep.Status.Advisory)`; `EvidenceDigest`; then `ReportID` last). Fill the fixture's two digests with the computed values; remove the Task 7 skips.
- [ ] **Step 4: Run green** for the whole package with `-race`.
- [ ] **Step 5: Commit:** `feat(report): aggregate counts, status, evidence digest, and report id`

**Checkpoint B (after Task 8): `architect` on `core/report` as a whole.**

## Task 9: Markdown renderer

**Files:**

- Create: `core/render/renderer.go`, `core/render/markdown.go`, `core/render/funcs.go`, `core/render/templates/report.md.tmpl`, `core/render/testdata/worked-example.md`, `core/render/markdown_test.go`, `core/render/testhelpers_test.go`

**Interfaces:**

- Produces: `type Renderer interface { Format() string; Render(report.Report) ([]byte, error) }`, `func ByFormat(format string) (Renderer, bool)`, formats `"md"` and (Task 10) `"html"`.

- [ ] **Step 1: Write the failing test**

```go
package render_test

import (
 "bytes"
 "flag"
 "os"
 "path/filepath"
 "testing"

 "github.com/mickeyyaya/v-eval/core/render"
 "github.com/mickeyyaya/v-eval/core/report"
)

var update = flag.Bool("update", false, "rewrite golden files")

func loadFixture(t *testing.T) report.Report {
 t.Helper()
 raw, err := os.ReadFile(filepath.Join("..", "report", "testdata", "worked-example.json"))
 if err != nil {
  t.Fatal(err)
 }
 rep, err := report.Decode(raw)
 if err != nil {
  t.Fatal(err)
 }
 return rep
}

func assertGolden(t *testing.T, name string, got []byte) {
 t.Helper()
 path := filepath.Join("testdata", name)
 if *update {
  if err := os.WriteFile(path, got, 0o644); err != nil {
   t.Fatal(err)
  }
 }
 want, err := os.ReadFile(path)
 if err != nil {
  t.Fatalf("missing golden %s; run: go test ./core/render -update", path)
 }
 if !bytes.Equal(got, want) {
  t.Fatalf("%s differs from golden; run go test ./core/render -update and review the diff", name)
 }
}

func TestMarkdownGoldenAndSectionOrder(t *testing.T) {
 r, ok := render.ByFormat("md")
 if !ok {
  t.Fatal("md renderer missing")
 }
 out, err := r.Render(loadFixture(t))
 if err != nil {
  t.Fatal(err)
 }
 assertGolden(t, "worked-example.md", out)
 order := []string{"## Status", "## Observations", "## Claims", "## Criteria", "## Forensics", "## Counts", "## Improvement", "## Limitations", "## Routing", "## Provenance"}
 last := -1
 for _, h := range order {
  i := bytes.Index(out, []byte(h))
  if i < 0 || i < last {
   t.Fatalf("section %q missing or out of order", h)
  }
  last = i
 }
 if !bytes.Contains(out, []byte("(supplied)")) {
  t.Fatal("candidate-supplied evidence must be labeled (supplied)")
 }
 if !bytes.Contains(out, []byte("Overall: FAIL")) {
  t.Fatal("status line missing")
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `renderer.go` (interface, registry map), `funcs.go` (`badge(Result) string`, `joinIDs([]string) string`, `supplied(Evidence) string` returning `" (supplied)"` for candidate origin, `locator(Locator) string` printing `file:start-end`, `command (exit N)`, or `passage from source (date)`), `markdown.go` (embedded `report.md.tmpl` via `embed.FS`, `text/template` with the funcs, `bytes.Buffer` output). Template order: title with identity, brief status block ("Overall: FAIL", rule, blocked_by), then Observations, Claims table, Criteria table, Forensics, Dimensions (only if present), Counts with numerator/denominator or "undefined (0 applicable)", Improvement, Limitations, Routing rationale, Provenance, Learning (if present). Run once with `-update` to create the golden; read the golden by eye against `examples/code-review/report.md` before committing.
- [ ] **Step 4: Run green.**
- [ ] **Step 5: Commit:** `feat(render): markdown renderer with fixed section order`

## Task 10: HTML renderer

**Files:**

- Create: `core/render/html.go`, `core/render/templates/report.html.tmpl`, `core/render/testdata/worked-example.html`, `core/render/html_test.go`

- [ ] **Step 1: Write the failing test**

```go
package render_test

import (
 "bytes"
 "testing"

 "github.com/mickeyyaya/v-eval/core/render"
 "github.com/mickeyyaya/v-eval/core/report"
)

func TestHTMLIsSelfContainedThemedOrderedAndEscaped(t *testing.T) {
 rep := loadFixture(t)
 rep.Observations = append(rep.Observations, report.Observation{ID: "O9", Text: "<script>alert(1)</script>", Origin: report.ObservationOriginAssistant, Evidence: []report.Evidence{}, CriterionIDs: []string{}})
 r, _ := render.ByFormat("html")
 out, err := r.Render(rep)
 if err != nil {
  t.Fatal(err)
 }
 for _, forbidden := range []string{"<script src", "<link ", "@import", "url(http", "https://fonts"} {
  if bytes.Contains(out, []byte(forbidden)) {
   t.Fatalf("external resource reference %q", forbidden)
  }
 }
 if !bytes.Contains(out, []byte("prefers-color-scheme: dark")) {
  t.Fatal("must carry a dark theme")
 }
 if bytes.Contains(out, []byte("<script>alert")) || !bytes.Contains(out, []byte("&lt;script&gt;alert")) {
  t.Fatal("observation text must be escaped")
 }
 ids := []string{`id="status"`, `id="observations"`, `id="claims"`, `id="criteria"`, `id="forensics"`, `id="counts"`, `id="improvement"`, `id="limitations"`, `id="routing"`, `id="provenance"`}
 last := -1
 for _, id := range ids {
  i := bytes.Index(out, []byte(id))
  if i < 0 || i < last {
   t.Fatalf("section %s missing or out of order", id)
  }
  last = i
 }
 if !bytes.HasPrefix(out, []byte("<!doctype html>")) {
  t.Fatal("must be a complete document")
 }
}

func TestHTMLGolden(t *testing.T) {
 r, _ := render.ByFormat("html")
 out, err := r.Render(loadFixture(t))
 if err != nil {
  t.Fatal(err)
 }
 assertGolden(t, "worked-example.html", out)
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `html.go` with `html/template` and an embedded template: `<!doctype html>`, `<meta charset>`, `<meta name="viewport">`, one `<style>` with CSS custom properties on `:root`, a `@media (prefers-color-scheme: dark)` block overriding them, tables for claims and criteria, badges per result via class names, evidence lines showing kind and isolation, supplied evidence marked, sections with the ids above in order. No `<script>`. Register `"html"` in the registry.
- [ ] **Step 4: Run green.** Open the golden in a browser once to check it reads cleanly in light and dark.
- [ ] **Step 5: Commit:** `feat(render): self-contained HTML report`

## Task 11: SARIF export

**Files:**

- Create: `core/export/sarif_types.go`, `core/export/sarif.go`, `core/export/testdata/worked-example.sarif.json`, `core/export/sarif_test.go`

**Interfaces:**

- Produces: `func ToSARIF(rep report.Report) (Log, error)`, `func Marshal(log Log) ([]byte, error)`; `Log{Schema, Version string; Runs []Run}`, `Run{Tool Tool; Results []Result; Invocations []Invocation; VersionControlProvenance []VCS; Artifacts []Artifact; Properties map[string]any}`.

- [ ] **Step 1: Write the failing test**

```go
package export_test

import (
 "os"
 "path/filepath"
 "testing"

 "github.com/mickeyyaya/v-eval/core/export"
 "github.com/mickeyyaya/v-eval/core/report"
)

func loadFixture(t *testing.T) report.Report {
 t.Helper()
 raw, err := os.ReadFile(filepath.Join("..", "report", "testdata", "worked-example.json"))
 if err != nil {
  t.Fatal(err)
 }
 rep, err := report.Decode(raw)
 if err != nil {
  t.Fatal(err)
 }
 return rep
}

func TestSARIFMapping(t *testing.T) {
 t.Parallel()
 log, err := export.ToSARIF(loadFixture(t))
 if err != nil {
  t.Fatal(err)
 }
 if log.Version != "2.1.0" || len(log.Runs) != 1 || len(log.Runs[0].Tool.Driver.Rules) != 5 {
  t.Fatalf("shape: %+v", log)
 }
 kinds := map[string]string{}
 for _, r := range log.Runs[0].Results {
  kinds[r.RuleID] = r.Kind
 }
 if kinds["C1"] != "fail" || kinds["C4"] != "pass" || kinds["C5"] != "review" {
  t.Fatalf("kinds = %v", kinds)
 }
 veval := log.Runs[0].Properties["veval"].(map[string]any)
 if veval["overall"] != "FAIL" {
  t.Fatal("run.properties.veval.overall must carry the overall status")
 }
}

func TestSARIFErrorCriterionMarksInvocationFailed(t *testing.T) {
 t.Parallel()
 rep := loadFixture(t)
 rep.Criteria[4].Result = report.ResultError
 rep.Provenance.Commands = []report.CommandRecord{{Command: "go test ./...", Cwd: ".", ExitStatus: 2, StartedAt: "2026-09-14T00:00:00Z", EndedAt: "2026-09-14T00:00:01Z", LogRef: "logs/1.txt", Isolation: report.IsolationNone}}
 log, err := export.ToSARIF(rep)
 if err != nil {
  t.Fatal(err)
 }
 inv := log.Runs[0].Invocations
 if len(inv) != 1 || inv[0].ExecutionSuccessful || len(inv[0].ToolExecutionNotifications) != 1 {
  t.Fatalf("invocations = %+v", inv)
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** the mapping table from `report-schema.md`: rules from contract criteria; results with `kind` (`pass`, `fail`, `notApplicable`, `review` for UNKNOWN, and for ERROR `kind: review` plus `properties.veval_result: ERROR` and the invocation marked failed with one notification per ERROR criterion); `level` `error` for required FAIL, `warning` otherwise; file locators become `physicalLocation` with `artifactLocation.uri` and `region.startLine/endLine`; other locators go to `properties.evidence`; commands become `invocations` (`executionSuccessful = exit_status == 0 && no ERROR criteria`); `versionControlProvenance` from `identity.artifact.revision` when it starts with a git-like id, else `artifacts[]` with `hashes.sha-256` when the revision is `content:sha256:…` (amended during execution: `versionControlProvenance` is not emitted at all. SARIF requires `repositoryUri` on every entry and the report carries no repository URI, so a git-like revision goes to `run.properties.veval.vcs_revision_id` instead; the `artifacts[].hashes` half is unchanged); `run.properties.veval` carries overall status, rule, counts. Golden file via the same `-update` idiom.
- [ ] **Step 4: Run green.**
- [ ] **Step 5: Commit:** `feat(export): SARIF 2.1.0 exporter`

**Checkpoint C (after Task 11): `architect` on `core/render` and `core/export` adapters against the Renderer port.**

## Task 12: The `veval` command line

**Files:**

- Create: `cmd/veval/main.go`, `cli.go`, `io.go`, `cmd_validate.go`, `cmd_aggregate.go`, `cmd_render.go`, `cmd_export.go`, `cli_test.go`

- [ ] **Step 1: Write the failing test**

```go
package main

import (
 "bytes"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func fixturePath() string {
 return filepath.Join("..", "..", "core", "report", "testdata", "worked-example.json")
}

func mustRead(t *testing.T, path string) string {
 t.Helper()
 b, err := os.ReadFile(path)
 if err != nil {
  t.Fatal(err)
 }
 return string(b)
}

func TestRunValidateAndExitCodes(t *testing.T) {
 var out, errb bytes.Buffer
 if code := run([]string{"validate", fixturePath()}, nil, &out, &errb); code != 0 || !strings.HasPrefix(out.String(), "valid:") {
  t.Fatalf("validate: code=%d out=%q err=%q", code, out.String(), errb.String())
 }
 out.Reset()
 broken := strings.Replace(mustRead(t, fixturePath()), `"overall": "FAIL"`, `"overall": "PASS"`, 1)
 if code := run([]string{"validate", "-"}, strings.NewReader(broken), &out, &errb); code != 1 || !strings.Contains(out.String(), "status.match") {
  t.Fatalf("broken: code=%d out=%q", code, out.String())
 }
 if code := run([]string{"frobnicate"}, nil, &out, &errb); code != 2 || !strings.Contains(errb.String(), "error:") {
  t.Fatalf("unknown subcommand code=%d err=%q", code, errb.String())
 }
 if code := run([]string{"render", "--format", "pdf", fixturePath()}, nil, &out, &errb); code != 2 {
  t.Fatalf("unknown format code=%d", code)
 }
 if code := run([]string{"validate", "does-not-exist.json"}, nil, &out, &errb); code != 2 {
  t.Fatalf("missing input code=%d", code)
 }
}

func TestRunAggregateRenderExport(t *testing.T) {
 dir := t.TempDir()
 stripped := strings.Replace(mustRead(t, fixturePath()), `"overall": "FAIL"`, `"overall": ""`, 1)
 outPath := filepath.Join(dir, "agg.json")
 var out, errb bytes.Buffer
 if code := run([]string{"aggregate", "-", "-o", outPath}, strings.NewReader(stripped), &out, &errb); code != 0 {
  t.Fatalf("aggregate: code=%d err=%q", code, errb.String())
 }
 if !strings.Contains(mustRead(t, outPath), `"overall": "FAIL"`) {
  t.Fatal("aggregate must recompute status")
 }
 htmlPath := filepath.Join(dir, "r.html")
 if code := run([]string{"render", outPath, "--format", "html", "-o", htmlPath}, nil, &out, &errb); code != 0 || !strings.HasPrefix(mustRead(t, htmlPath), "<!doctype html>") {
  t.Fatalf("render html: code=%d err=%q", code, errb.String())
 }
 out.Reset()
 if code := run([]string{"render", "--format", "md", outPath}, nil, &out, &errb); code != 0 || !strings.Contains(out.String(), "## Criteria") {
  t.Fatalf("render md: code=%d", code)
 }
 out.Reset()
 if code := run([]string{"export", "sarif", outPath}, nil, &out, &errb); code != 0 || !strings.Contains(out.String(), `"version": "2.1.0"`) {
  t.Fatalf("export: code=%d out=%q", code, out.String())
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `cli.go`:

```go
type subcommand struct {
 Name    string
 Summary string
 Run     func(args []string, stdin io.Reader, stdout, stderr io.Writer) int
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
 if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
  fmt.Fprint(stderr, usage())
  return 2
 }
 for _, cmd := range commands {
  if cmd.Name == args[0] {
   return cmd.Run(args[1:], stdin, stdout, stderr)
  }
 }
 fmt.Fprintf(stderr, "error: unknown command %q\n\n%s", args[0], usage())
 return 2
}
```

`io.go`: `readInput(path string, stdin io.Reader) ([]byte, error)` (`-` → stdin), `writeOutput(path string, stdout io.Writer, data []byte) error`, `splitArgs(flags *flag.FlagSet, args []string) (flagArgs, operands []string)` (amended during execution: it was planned as `reorderArgs` returning one reordered slice; it returns the flag tokens and the operands separately instead, because Go's flag package stops at the first operand, so only a separate flag slice lets the set report `flag needs an argument: -o` for a trailing `-o`. Which flags take a following value comes from the command's own `FlagSet`). Each `cmd_*.go` parses its flag set, reads input, calls the core, prints violations one per line as `<path>: <rule>: <message>` and returns 1, or writes output and returns 0; any operational error prints `error: <msg>` to stderr and returns 2. `main.go` is `os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))`.

- [ ] **Step 4: Run green.** Also `go vet ./...`.
- [ ] **Step 5: Commit:** `feat(veval): validate, aggregate, render, and export commands`

## Task 13: Version, CI matrix, GoReleaser

**Files:**

- Create: `internal/version/version.go`, `cmd/veval/cmd_version.go`, `cmd/veval/version_test.go`, `.github/workflows/go.yml`, `.goreleaser.yaml`

- [ ] **Step 1: Write the failing test**

```go
package main

import (
 "bytes"
 "io"
 "strings"
 "testing"
)

func TestVersionStringIncludesSchema(t *testing.T) {
 var out bytes.Buffer
 if code := run([]string{"version"}, nil, &out, io.Discard); code != 0 || !strings.Contains(out.String(), "schema 0.1.0") || !strings.HasPrefix(out.String(), "veval ") {
  t.Fatalf("code=%d out=%q", code, out.String())
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement** `internal/version`: `var Version = "0.0.0-dev"`, `var BuildDigest = "unknown"`, `func String() string`. `cmd_version.go` prints `veval <Version> (<BuildDigest>) schema <schema.Version>`. Workflow `.github/workflows/go.yml`:

```yaml
name: go
on:
  push:
    branches: [main]
  pull_request:
jobs:
  test:
    name: go (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: gofmt
        shell: bash
        run: test -z "$(gofmt -l .)" || (gofmt -l . && exit 1)
      - name: vet
        run: go vet ./...
      - name: test
        run: go test -count=1 ./...
      - name: test (race)
        if: runner.os != 'Windows'
        run: go test -race -count=1 ./...
      - name: build
        run: go build ./cmd/veval
```

`.goreleaser.yaml`:

```yaml
version: 2
project_name: veval
before:
  hooks:
    - go mod tidy
builds:
  - id: veval
    main: ./cmd/veval
    env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    flags: [-trimpath]
    ldflags:
      - -s -w -X github.com/mickeyyaya/v-eval/internal/version.Version={{.Version}} -X github.com/mickeyyaya/v-eval/internal/version.BuildDigest={{.ShortCommit}}
archives:
  - id: veval
    name_template: "veval_{{ .Os }}_{{ .Arch }}"
    formats: [tar.gz]
    format_overrides:
      - goos: windows
        formats: [zip]
    files: [LICENSE, README.md, CHANGELOG.md]
checksum:
  name_template: checksums.txt
changelog:
  disable: true
```

- [ ] **Step 4: Run green.** `goreleaser check` if installed; otherwise note it in the commit body.
- [ ] **Step 5: Commit:** `ci: Go matrix on three operating systems and GoReleaser config`

## Task 14: Skill wiring and host mappings

**Files:**

- Modify: `skills/evaluate-output/SKILL.md` (step 7 "Report", section "Host adaptation")
- Create: `skills/evaluate-output/references/hosts/claude-code.md`, `codex.md`, `gemini-cli.md`, `antigravity.md`, `hermes.md`, `ollama.md`, `cmd/veval/skill_contract_test.go`

- [ ] **Step 1: Write the failing test**

```go
package main

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestSkillNamesTheRealSubcommandsAndHosts(t *testing.T) {
 text := mustRead(t, filepath.Join("..", "..", "skills", "evaluate-output", "SKILL.md"))
 for _, cmd := range []string{"veval aggregate", "veval validate", "veval render --format html", "veval render --format md", "veval export sarif"} {
  if !strings.Contains(text, cmd) {
   t.Fatalf("SKILL.md does not mention %q", cmd)
  }
 }
 for _, host := range []string{"claude-code", "codex", "gemini-cli", "antigravity", "hermes", "ollama"} {
  path := filepath.Join("..", "..", "skills", "evaluate-output", "references", "hosts", host+".md")
  if _, err := os.Stat(path); err != nil {
   t.Fatalf("missing host mapping %s", host)
  }
  if !strings.Contains(mustRead(t, path), "| Run a command") {
   t.Fatalf("%s must map the run-a-command action", host)
  }
 }
}
```

- [ ] **Step 2: Run red.**
- [ ] **Step 3: Implement.** SKILL.md step 7 becomes: write `report.json` in the schema shape with `counts`, `status`, `report_id`, and `evidence_digest` empty; run `veval aggregate report.json -o report.json`; run `veval validate report.json` and fix every violation before reporting; run `veval render --format html report.json -o report.html` and `veval render --format md report.json -o report.md`; optionally `veval export sarif report.json -o report.sarif.json`; if `veval` is absent, state that in `limitations.unknown_metadata`, compute counts and status by the documented rules, and label the report core-unvalidated. Host adaptation section points at `references/hosts/`. Each host file: a table with rows Read a file, Search file contents, List files, Run a command, Fetch a URL, Dispatch a subagent, Invoke the v-eval core; sources: Claude Code from its tool set (`Read`, `Grep`, `Glob`, `Bash`, `WebFetch`, `Agent`); Codex, Gemini CLI, Antigravity, and Hermes copied from the superpowers plugin's `references/{codex,gemini,antigravity,hermes}-tools.md` (cite the path and 2026-09-14); ollama: "No tools of its own; runs through a harness (Claude Code, Codex, OpenCode) whose mapping applies, or reaches the core over MCP in Stage 5; without either, every execution criterion is UNKNOWN and the report says so." Every host file ends with "Invoke the v-eval core: run `veval …` through the host's command tool; on Windows the binary is `veval.exe`."
- [ ] **Step 4: Run green.** Run `python3 tools/docs/vdocs.py register .` (new links) and `check-links`; `markdownlint` clean.
- [ ] **Step 5: Commit:** `docs(skill): wire draft-2 to the veval core and add host mappings`

## Task 15: Status lines, changelog, final review, merge, push

**Files:**

- Modify: `README.md` (status paragraph: the Go core and `veval` exist; adapters, classifier, detectors, learning do not), `CHANGELOG.md` (Unreleased: Added core, CLI, renderers, SARIF, CI, GoReleaser, host mappings), `docs/architecture/report-schema.md` (status line: schema v0.1.0 exists under `schema/`; field names are no longer proposals), `docs/architecture/packaging-and-portability.md` (tree entries that now exist), `ROADMAP.md` (Stage 1 items done or remaining)

- [ ] **Step 1:** Update the five documents; `vdocs register .`, `check-links`, markdownlint clean.
- [ ] **Step 2:** Full verification: `gofmt -l .` empty, `go vet ./...`, `go test -race -count=1 ./...`, `python3 -m unittest discover -s tools/docs -p "test_*.py"`.
- [ ] **Step 3:** Final review trio on the branch diff: `code-simplifier`, `go-reviewer`, `architect`. Fix findings, re-run Step 2.
- [ ] **Step 4:** Commit: `docs: record the core walking skeleton in status lines and changelog`
- [ ] **Step 5:** `git checkout main && git merge --ff-only feat/core-skeleton && git push origin main`; watch `gh run list` until the `go` and `docs` workflows are green on all three operating systems.

**Checkpoint D (Task 15 Step 3): final `architect` pass before merge.**

---

## Verification (end to end)

1. `go test -race -count=1 ./...` passes on macOS locally; CI shows green `go (ubuntu-latest)`, `go (macos-latest)`, `go (windows-latest)`.
2. `go run ./cmd/veval validate core/report/testdata/worked-example.json` prints `valid: …` and exits 0.
3. Editing the fixture's C4 evidence to a `note` locator and running `validate` prints a line containing `evidence.pass_requires_observed_locator` and exits 1: the core rejects a PASS without opened-or-ran evidence.
4. Changing `"overall": "FAIL"` to `"PASS"` and running `validate` prints `status.match` and exits 1: an invented rollup is caught.
5. `go run ./cmd/veval render --format html core/report/testdata/worked-example.json -o /tmp/r.html` and opening the file offline shows the sections in order, status brief at the top, observations before criteria, in both light and dark system themes.
6. `go run ./cmd/veval export sarif …` output validates structurally (`"version": "2.1.0"`, five rules, results with kinds pass/fail/review).
7. The skill's step 7 commands run as written against the fixture on a machine with the binary on `PATH`, and `cmd/veval/skill_contract_test.go` fails if any command name in SKILL.md drifts from the CLI.

## Appendix: schema/report.schema.json (v0.1.0)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/mickeyyaya/v-eval/schema/report/0.1.0",
  "title": "v-eval report",
  "type": "object",
  "additionalProperties": false,
  "required": ["identity", "contract", "routing", "observations", "claims", "criteria", "forensics", "counts", "status", "improvement", "limitations", "provenance"],
  "properties": {
    "identity": {"type": "object", "additionalProperties": false,
      "required": ["report_id", "schema_version", "veval_version", "skill_revision", "artifact", "task", "created_at", "host"],
      "properties": {
        "report_id": {"type": "string", "pattern": "^(sha256:[0-9a-f]{64})?$"},
        "schema_version": {"const": "0.1.0"},
        "veval_version": {"type": "string", "minLength": 1},
        "skill_revision": {"type": "string", "minLength": 1},
        "artifact": {"type": "object", "additionalProperties": false, "required": ["kind", "revision", "paths", "bundle_digest"],
          "properties": {"kind": {"type": "string", "minLength": 1}, "revision": {"type": "string", "minLength": 1}, "paths": {"type": "array", "items": {"type": "string"}}, "bundle_digest": {"type": "string"}}},
        "task": {"type": "object", "additionalProperties": false, "required": ["requested_outcome", "intended_user", "brief_ref"],
          "properties": {"requested_outcome": {"type": "string", "minLength": 1}, "intended_user": {"type": "string"}, "brief_ref": {"type": "string"}}},
        "created_at": {"type": "string", "format": "date-time"},
        "host": {"type": "object", "additionalProperties": false, "required": ["cli", "os", "arch"],
          "properties": {"cli": {"type": "string"}, "model": {"type": "string"}, "os": {"type": "string"}, "arch": {"type": "string"}}}}},
    "contract": {"type": "object", "additionalProperties": false,
      "required": ["contract_id", "contract_version", "status", "conflicts", "criteria"],
      "properties": {
        "contract_id": {"type": "string", "minLength": 1},
        "contract_version": {"type": "string", "minLength": 1},
        "status": {"enum": ["user_specified", "provisional", "approved"]},
        "conflicts": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["between", "description", "resolution"],
          "properties": {"between": {"type": "array", "items": {"type": "string"}}, "description": {"type": "string"}, "resolution": {"type": "string", "minLength": 1}}}},
        "criteria": {"type": "array", "minItems": 1, "items": {"$ref": "#/$defs/criterion"}}}},
    "routing": {"type": "object", "additionalProperties": false,
      "required": ["supplied", "profiles", "adapters_run", "adapters_skipped", "ambiguity", "rationale"],
      "properties": {
        "supplied": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["source_type", "count", "origin"],
          "properties": {"source_type": {"type": "string"}, "count": {"type": "integer", "minimum": 0}, "origin": {"type": "string"}}}},
        "profiles": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["name", "version"], "properties": {"name": {"type": "string"}, "version": {"type": "string"}}}},
        "adapters_run": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["name", "version", "inputs_used"],
          "properties": {"name": {"type": "string"}, "version": {"type": "string"}, "inputs_used": {"type": "array", "items": {"type": "string"}}}}},
        "adapters_skipped": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["name", "missing_input"], "properties": {"name": {"type": "string"}, "missing_input": {"type": "string"}}}},
        "ambiguity": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["question", "resolution"], "properties": {"question": {"type": "string"}, "resolution": {"type": "string"}}}},
        "rationale": {"type": "string", "minLength": 1}}},
    "observations": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["id", "text", "evidence", "criterion_ids", "origin"],
      "properties": {"id": {"type": "string", "minLength": 1}, "text": {"type": "string", "minLength": 1}, "evidence": {"$ref": "#/$defs/evidence_list"}, "criterion_ids": {"type": "array", "items": {"type": "string"}}, "origin": {"enum": ["assistant", "adapter", "detector"]}}}},
    "claims": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["claim_id", "text", "location", "status", "verification", "notes"],
      "properties": {"claim_id": {"type": "string", "minLength": 1}, "text": {"type": "string", "minLength": 1}, "location": {"type": "string"}, "status": {"enum": ["verified", "contradicted", "unverified", "not_checkable"]}, "verification": {"$ref": "#/$defs/evidence_list"}, "notes": {"type": "string"}}}},
    "criteria": {"type": "array", "minItems": 1, "items": {"type": "object", "additionalProperties": false,
      "required": ["id", "result", "method_used", "evidence", "reasoning", "next_action", "shared_cause_with", "dimension_refs"],
      "properties": {"id": {"type": "string", "minLength": 1}, "result": {"$ref": "#/$defs/result"}, "method_used": {"$ref": "#/$defs/method"}, "evidence": {"$ref": "#/$defs/evidence_list"}, "reasoning": {"type": "string"}, "next_action": {"type": "string"}, "shared_cause_with": {"type": "array", "items": {"type": "string"}}, "dimension_refs": {"type": "array", "items": {"type": "string"}}}}},
    "forensics": {"type": "array", "items": {"type": "object", "additionalProperties": false,
      "required": ["finding_id", "detector", "detector_version", "criterion_id", "severity", "evidence", "benign_alternative", "disposition"],
      "properties": {"finding_id": {"type": "string", "minLength": 1}, "detector": {"type": "string"}, "detector_version": {"type": "string"}, "criterion_id": {"type": "string"}, "severity": {"enum": ["observed", "suspicious", "confirmed"]}, "evidence": {"$ref": "#/$defs/evidence_list"}, "benign_alternative": {"type": "string"}, "disposition": {"enum": ["open", "explained", "confirmed"]}}}},
    "dimensions": {"type": "array", "items": {"type": "object", "additionalProperties": false,
      "required": ["dimension", "metric", "metric_version", "authority_type", "definition_ref", "unit", "range", "direction", "value", "threshold", "threshold_source", "tool", "tool_version", "workload", "evidence", "interpretation"],
      "properties": {"dimension": {"type": "string"}, "metric": {"type": "string"}, "metric_version": {"type": "string"}, "authority_type": {"enum": ["formal_standard", "established_measure", "vendor_rating", "project_rubric"]}, "definition_ref": {"type": "string", "minLength": 1}, "unit": {"type": "string"}, "range": {"type": "string"}, "direction": {"type": "string"}, "value": {"type": ["number", "null"]}, "threshold": {"type": "string"}, "threshold_source": {"type": "string"}, "tool": {"type": "string"}, "tool_version": {"type": "string"}, "workload": {"type": "string"}, "evidence": {"$ref": "#/$defs/evidence_list"}, "interpretation": {"type": "string"}}}},
    "counts": {"type": "object", "additionalProperties": false, "required": ["required", "optional", "coverage"],
      "properties": {"required": {"$ref": "#/$defs/tally"}, "optional": {"$ref": "#/$defs/tally"}, "coverage": {"type": "object", "additionalProperties": false, "required": ["numerator", "denominator", "undefined"], "properties": {"numerator": {"type": "integer", "minimum": 0}, "denominator": {"type": "integer", "minimum": 0}, "undefined": {"type": "boolean"}}}}},
    "status": {"type": "object", "additionalProperties": false, "required": ["overall", "rule_applied", "blocked_by", "advisory"],
      "properties": {"overall": {"enum": ["PASS", "FAIL", "INCOMPLETE", "ADVISORY"]}, "rule_applied": {"type": "string", "minLength": 1}, "blocked_by": {"type": "array", "items": {"type": "string"}}, "advisory": {"type": "boolean"}}},
    "improvement": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["issue", "locations", "suggested_change", "constraints_to_preserve", "verify_by", "linked_criteria"],
      "properties": {"issue": {"type": "string"}, "locations": {"type": "array", "items": {"type": "string"}}, "suggested_change": {"type": "string"}, "constraints_to_preserve": {"type": "array", "items": {"type": "string"}}, "verify_by": {"type": "string"}, "linked_criteria": {"type": "array", "items": {"type": "string"}}}}},
    "limitations": {"type": "object", "additionalProperties": false, "required": ["not_inspected", "not_executed", "unknown_metadata", "assumptions"],
      "properties": {"not_inspected": {"type": "array", "items": {"type": "string"}}, "not_executed": {"type": "array", "items": {"type": "string"}}, "unknown_metadata": {"type": "array", "items": {"type": "string"}}, "assumptions": {"type": "array", "items": {"type": "string"}}}},
    "provenance": {"type": "object", "additionalProperties": false, "required": ["tools", "commands", "environment", "isolation_levels_used", "evidence_digest"],
      "properties": {"tools": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["name", "version"], "properties": {"name": {"type": "string"}, "version": {"type": "string"}}}},
                     "commands": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["command", "cwd", "exit_status", "started_at", "ended_at", "log_ref", "isolation"],
                                  "properties": {"command": {"type": "string"}, "cwd": {"type": "string"}, "exit_status": {"type": "integer"}, "started_at": {"type": "string"}, "ended_at": {"type": "string"}, "log_ref": {"type": "string"}, "isolation": {"$ref": "#/$defs/isolation"}}}},
                     "environment": {"type": "object", "additionalProperties": false, "required": ["os", "arch", "runtime_versions"], "properties": {"os": {"type": "string"}, "arch": {"type": "string"}, "runtime_versions": {"type": "object", "additionalProperties": {"type": "string"}}}},
                     "isolation_levels_used": {"type": "array", "items": {"$ref": "#/$defs/isolation"}},
                     "evidence_digest": {"type": "string", "pattern": "^(sha256:[0-9a-f]{64})?$"}}},
    "learning": {"type": "object", "additionalProperties": false, "required": ["precedents_retrieved", "reward_records_created"],
      "properties": {"precedents_retrieved": {"type": "array", "items": {"type": "object", "additionalProperties": false, "required": ["precedent_id", "criterion", "similarity"], "properties": {"precedent_id": {"type": "string"}, "criterion": {"type": "string"}, "similarity": {"type": "number", "minimum": 0, "maximum": 1}}}},
                     "reward_records_created": {"type": "array", "items": {"type": "string"}}}}
  },
  "$defs": {
    "result": {"enum": ["PASS", "FAIL", "UNKNOWN", "ERROR", "NOT_APPLICABLE"]},
    "method": {"enum": ["execution", "deterministic_check", "static_inspection", "source_verification", "rubric_judgment", "human_judgment"]},
    "isolation": {"enum": ["none", "worktree", "container", "remote_sandbox"]},
    "tally": {"type": "object", "additionalProperties": false, "required": ["applicable", "pass", "fail", "unknown", "error", "not_applicable"],
      "properties": {"applicable": {"type": "integer", "minimum": 0}, "pass": {"type": "integer", "minimum": 0}, "fail": {"type": "integer", "minimum": 0}, "unknown": {"type": "integer", "minimum": 0}, "error": {"type": "integer", "minimum": 0}, "not_applicable": {"type": "integer", "minimum": 0}}},
    "criterion": {"type": "object", "additionalProperties": false,
      "required": ["id", "requirement", "source_ref", "required", "applicability", "methods_allowed", "acceptance_rule", "expected_evidence", "provisional"],
      "properties": {"id": {"type": "string", "minLength": 1}, "requirement": {"type": "string", "minLength": 1}, "source_ref": {"type": "string"}, "required": {"type": "boolean"}, "applicability": {"type": "string"}, "methods_allowed": {"type": "array", "minItems": 1, "items": {"$ref": "#/$defs/method"}}, "acceptance_rule": {"type": "string", "minLength": 1}, "expected_evidence": {"type": "string"}, "provisional": {"type": "boolean"}}},
    "locator": {"type": "object", "oneOf": [
        {"additionalProperties": false, "required": ["file", "line_start", "line_end"], "properties": {"file": {"type": "string", "minLength": 1}, "line_start": {"type": "integer", "minimum": 1}, "line_end": {"type": "integer", "minimum": 1}}},
        {"additionalProperties": false, "required": ["command", "cwd", "exit_status", "log_ref"], "properties": {"command": {"type": "string", "minLength": 1}, "cwd": {"type": "string", "minLength": 1}, "exit_status": {"type": "integer"}, "log_ref": {"type": "string", "minLength": 1}}},
        {"additionalProperties": false, "required": ["passage", "source_ref", "source_date_or_version", "access_date"], "properties": {"passage": {"type": "string", "minLength": 1}, "source_ref": {"type": "string", "minLength": 1}, "source_date_or_version": {"type": "string", "minLength": 1}, "access_date": {"type": "string", "minLength": 1}}},
        {"additionalProperties": false, "required": ["note"], "properties": {"note": {"type": "string", "minLength": 1}}}]},
    "evidence": {"type": "object", "additionalProperties": false, "required": ["kind", "locator", "observation", "provenance", "isolation", "origin"],
      "properties": {"kind": {"enum": ["execution", "inspection", "supplied", "judgment"]}, "locator": {"$ref": "#/$defs/locator"}, "observation": {"type": "string", "minLength": 1},
                     "provenance": {"type": "object", "additionalProperties": false, "required": ["tool", "version", "revision", "timestamp"], "properties": {"tool": {"type": "string"}, "version": {"type": "string"}, "revision": {"type": "string"}, "timestamp": {"type": "string"}}},
                     "isolation": {"$ref": "#/$defs/isolation"}, "origin": {"enum": ["observed", "candidate_supplied", "retrieved"]}, "rubric_version": {"type": "string"}, "model": {"type": "string"}},
      "if": {"properties": {"kind": {"const": "judgment"}}}, "then": {"required": ["rubric_version"]}},
    "evidence_list": {"type": "array", "items": {"$ref": "#/$defs/evidence"}}
  }
}
```

## Self-review

- Spec coverage: report sections 1 to 14 (Task 4 types, Task 7 schema drift), evidence-shape rule (Tasks 3, 7), status derivation and counts (Tasks 5, 6), canonical JSON and digests (Tasks 4, 8), Markdown and HTML with fixed order and theming (Tasks 9, 10), SARIF mapping (Task 11), CLI with exit codes (Task 12), portability CI and release (Task 13), skill wiring and host references (Task 14), status documentation (Task 15). Not covered by design: adapters, classifier, detectors, learning, agent profile, plugin manifests (later plans).
- Placeholder scan: none; every code step carries code, every file is named.
- Type consistency: `Evidence` defined in Task 3, reused unchanged; `Counts`, `Tally`, `Coverage`, `Status` defined in Task 4 and used by Tasks 5 to 8; `Renderer`/`ByFormat` defined in Task 9 and used by Tasks 10 and 12; `run` signature identical in Tasks 12, 13, 14.
