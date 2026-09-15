package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
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
	code, out, errOut := call([]string{"validate", fixturePath()}, "")
	if code != 0 || !strings.HasPrefix(out, "valid:") {
		t.Fatalf("validate: code=%d out=%q err=%q", code, out, errOut)
	}
	broken := strings.Replace(mustRead(t, fixturePath()), `"overall": "FAIL"`, `"overall": "PASS"`, 1)
	code, out, _ = call([]string{"validate", "-"}, broken)
	if code != 1 || !strings.Contains(out, "status.match") {
		t.Fatalf("broken: code=%d out=%q", code, out)
	}
	code, _, errOut = call([]string{"frobnicate"}, "")
	if code != 2 || !strings.Contains(errOut, "error:") {
		t.Fatalf("unknown subcommand code=%d err=%q", code, errOut)
	}
	if code, _, _ := call([]string{"render", "--format", "pdf", fixturePath()}, ""); code != 2 {
		t.Fatalf("unknown format code=%d", code)
	}
	if code, _, _ := call([]string{"validate", "does-not-exist.json"}, ""); code != 2 {
		t.Fatalf("missing input code=%d", code)
	}
}

func TestRunAggregateRenderExport(t *testing.T) {
	dir := t.TempDir()
	stripped := strings.Replace(mustRead(t, fixturePath()), `"overall": "FAIL"`, `"overall": ""`, 1)
	outPath := filepath.Join(dir, "agg.json")
	code, _, errOut := call([]string{"aggregate", "-", "-o", outPath}, stripped)
	if code != 0 {
		t.Fatalf("aggregate: code=%d err=%q", code, errOut)
	}
	if !strings.Contains(mustRead(t, outPath), `"overall": "FAIL"`) {
		t.Fatal("aggregate must recompute status")
	}
	htmlPath := filepath.Join(dir, "r.html")
	code, _, errOut = call([]string{"render", outPath, "--format", "html", "-o", htmlPath}, "")
	if code != 0 || !strings.HasPrefix(mustRead(t, htmlPath), "<!doctype html>") {
		t.Fatalf("render html: code=%d err=%q", code, errOut)
	}
	code, out, _ := call([]string{"render", "--format", "md", outPath}, "")
	if code != 0 || !strings.Contains(out, "## Criteria") {
		t.Fatalf("render md: code=%d", code)
	}
	code, out, _ = call([]string{"export", "sarif", outPath}, "")
	if code != 0 || !strings.Contains(out, `"version": "2.1.0"`) {
		t.Fatalf("export: code=%d out=%q", code, out)
	}
}

// call runs the CLI on its own buffers, so no two tests share output.
func call(args []string, stdin string) (code int, stdout, stderr string) {
	var out, errb bytes.Buffer
	code = run(args, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

// aggregated writes the fixture through aggregate into a temp file, which is
// the only form the full validation the CLI gates on accepts.
func aggregated(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "agg.json")
	if code, _, errOut := call([]string{"aggregate", fixturePath(), "-o", path}, ""); code != 0 {
		t.Fatalf("aggregate fixture: code=%d err=%q", code, errOut)
	}
	return path
}

// wantUsageLines is the usage the command line promises: one line per
// command, naming that command's own flags and operands.
var wantUsageLines = []string{
	"veval validate <report.json|->",
	"veval aggregate [-o out] <report.json|->",
	"veval render [--format md|html] [-o out] <report.json|->",
	"veval export sarif [-o out] <report.json|->",
	"veval version",
}

// assertUsage requires text to be the whole usage: one line per command,
// naming that command's own flags and operands, and no block of global flags.
func assertUsage(t *testing.T, text string) {
	t.Helper()
	for _, cmd := range commands() {
		if !strings.Contains(text, cmd.Name) {
			t.Errorf("usage does not name %q", cmd.Name)
		}
		if !strings.Contains(text, cmd.Usage) {
			t.Errorf("usage does not carry %q", cmd.Usage)
		}
	}
	for _, line := range wantUsageLines {
		if !strings.Contains(text, line) {
			t.Errorf("usage does not carry %q", line)
		}
	}
	if strings.Contains(text, "write to this file instead") {
		t.Error("usage still carries the global flags block")
	}
}

func TestUsageGoesToStderrWhenNoCommandIsNamed(t *testing.T) {
	code, stdout, stderr := call(nil, "")
	if code != 2 {
		t.Errorf("code=%d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("usage wrote %q to stdout, want stderr only", stdout)
	}
	assertUsage(t, stderr)

	code, _, stderr = call([]string{"frobnicate"}, "")
	if code != 2 || !strings.Contains(stderr, `error: unknown command "frobnicate"`) {
		t.Errorf("unknown command: code=%d err=%q", code, stderr)
	}
}

// TestHelpIsAnAnswerNotAFailure separates the two cases the usage serves. A
// caller who typed nothing was told what to type, which is a failed
// invocation; a caller who asked for the usage was given what they asked for,
// so it goes to standard output and the exit code says the command succeeded.
func TestHelpIsAnAnswerNotAFailure(t *testing.T) {
	for _, args := range [][]string{{"-h"}, {"--help"}} {
		code, stdout, stderr := call(args, "")
		if code != 0 {
			t.Errorf("%v: code=%d, want 0", args, code)
		}
		if stderr != "" {
			t.Errorf("%v: help wrote %q to stderr, want stdout only", args, stderr)
		}
		assertUsage(t, stdout)
	}
}

// TestSubcommandHelpIsAnAnswerNotAFailure holds a command's own help to the
// top level's rule: a caller who asked for the usage was given what they
// asked for, on standard output, with an exit code that says so. The answer
// is the command's table line and its flag defaults; standard error stays
// silent.
func TestSubcommandHelpIsAnAnswerNotAFailure(t *testing.T) {
	for _, args := range [][]string{{"render", "-h"}, {"render", "--help"}, {"render", fixturePath(), "-h"}} {
		code, stdout, stderr := call(args, "")
		if code != exitOK {
			t.Errorf("%v: code=%d, want 0", args, code)
		}
		if stderr != "" {
			t.Errorf("%v: help wrote %q to stderr, want stdout only", args, stderr)
		}
		if want := "usage: " + usageLine("render") + "\n"; !strings.HasPrefix(stdout, want) {
			t.Errorf("%v: stdout %q does not start with %q", args, stdout, want)
		}
		for _, flagName := range []string{"-format", "-o"} {
			if !strings.Contains(stdout, "  "+flagName+" ") {
				t.Errorf("%v: help does not list the %s flag: %q", args, flagName, stdout)
			}
		}
	}
}

func TestUnknownFormatNamesTheAcceptedFormats(t *testing.T) {
	code, _, stderr := call([]string{"render", "--format", "pdf", fixturePath()}, "")
	want := `error: unknown format "pdf" (accepted: html, md)`
	if code != 2 || !strings.Contains(stderr, want) {
		t.Errorf("code=%d err=%q, want %q", code, stderr, want)
	}
	for _, format := range render.Formats() {
		if !strings.Contains(stderr, format) {
			t.Errorf("the message does not name the registered format %q: %q", format, stderr)
		}
	}
	code, _, stderr = call([]string{"export", "pdf", fixturePath()}, "")
	if code != 2 || !strings.Contains(stderr, `error: unknown format "pdf" (accepted: sarif)`) {
		t.Errorf("export format: code=%d err=%q", code, stderr)
	}
}

func TestFlagsMayFollowPositionals(t *testing.T) {
	path := aggregated(t)
	before, after := []string{"render", "--format", "md", path}, []string{"render", path, "--format", "md"}
	codeBefore, outBefore, _ := call(before, "")
	codeAfter, outAfter, _ := call(after, "")
	if codeBefore != 0 || codeAfter != 0 {
		t.Fatalf("codes: flags first %d, flags last %d", codeBefore, codeAfter)
	}
	if outBefore != outAfter {
		t.Error("flag order changed the rendering")
	}
}

// testFlags is a flag set shaped like the commands' own -- a string flag
// whose value is a separate token, a joined form, and a boolean flag that
// takes none -- so splitArgs is exercised over every kind of flag a
// command defines.
func testFlags() *flag.FlagSet {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.String("format", "md", "")
	flags.String("o", "", "")
	flags.Bool("strict", false, "")
	return flags
}

func TestSplitArgsSplitsFlagsFromOperands(t *testing.T) {
	cases := []struct {
		name         string
		in           []string
		wantFlags    []string
		wantOperands []string
	}{
		{"flags already first", []string{"--format", "md", "r.json"}, []string{"--format", "md"}, []string{"r.json"}},
		{"flag after operand", []string{"r.json", "--format", "md"}, []string{"--format", "md"}, []string{"r.json"}},
		{"joined value", []string{"r.json", "--format=md"}, []string{"--format=md"}, []string{"r.json"}},
		{"two flags around two operands", []string{"sarif", "-o", "x", "r.json"}, []string{"-o", "x"}, []string{"sarif", "r.json"}},
		{"dash stays an operand", []string{"-", "-o", "x"}, []string{"-o", "x"}, []string{"-"}},
		{"trailing flag keeps its missing value", []string{"r.json", "-o"}, []string{"-o"}, []string{"r.json"}},
		{"a boolean flag swallows nothing", []string{"-strict", "r.json"}, []string{"-strict"}, []string{"r.json"}},
		{"an undefined flag swallows nothing", []string{"r.json", "--nope", "x"}, []string{"--nope"}, []string{"r.json", "x"}},
		{"terminator hands the rest over", []string{"--", "-weird.json"}, []string{}, []string{"-weird.json"}},
		{"terminator ends flag parsing", []string{"-o", "x", "--", "-weird.json", "--format"}, []string{"-o", "x"}, []string{"-weird.json", "--format"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotFlags, gotOperands := splitArgs(testFlags(), tc.in)
			if !slices.Equal(gotFlags, tc.wantFlags) || !slices.Equal(gotOperands, tc.wantOperands) {
				t.Errorf("splitArgs(%q) = %q, %q, want %q, %q",
					tc.in, gotFlags, gotOperands, tc.wantFlags, tc.wantOperands)
			}
		})
	}
}

func TestATrailingFlagWithoutAValueIsReported(t *testing.T) {
	code, stdout, stderr := call([]string{"render", aggregated(t), "-o"}, "")
	if code != 2 || !strings.Contains(stderr, "flag needs an argument: -o") {
		t.Errorf("code=%d err=%q, want 2 and `flag needs an argument: -o`", code, stderr)
	}
	if stdout != "" {
		t.Errorf("wrote %q to stdout, want stderr only", stdout)
	}
}

func TestOutputFileIsLFOnlyAndMode644(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.md")
	source := aggregated(t)
	code, _, stderr := call([]string{"render", "--format", "md", source, "-o", path}, "")
	if code != 0 {
		t.Fatalf("render: code=%d err=%q", code, stderr)
	}
	written := mustRead(t, path)
	if strings.Contains(written, "\r") {
		t.Error("the written file carries a carriage return")
	}
	_, onStdout, _ := call([]string{"render", "--format", "md", source}, "")
	if written != onStdout {
		t.Error("the file and stdout renderings differ")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o644 {
		t.Errorf("mode %v, want -rw-r--r--", info.Mode().Perm())
	}
}

func TestStdinIsReadForEverySubcommand(t *testing.T) {
	raw := mustRead(t, aggregated(t))
	cases := [][]string{{"validate", "-"}, {"aggregate", "-"}, {"render", "--format", "md", "-"}, {"export", "sarif", "-"}}
	for _, args := range cases {
		t.Run(args[0], func(t *testing.T) {
			code, stdout, stderr := call(args, raw)
			if code != 0 {
				t.Fatalf("code=%d err=%q", code, stderr)
			}
			if stdout == "" {
				t.Error("nothing written to stdout")
			}
		})
	}
}

func TestEverySubcommandExitsOneOnViolations(t *testing.T) {
	full := strings.Replace(mustRead(t, fixturePath()), `"overall": "FAIL"`, `"overall": "PASS"`, 1)
	structural := strings.Replace(mustRead(t, fixturePath()), `"schema_version": "0.1.0"`, `"schema_version": "0.0.9"`, 1)
	cases := []struct {
		name  string
		args  []string
		stdin string
	}{
		{"validate", []string{"validate", "-"}, full},
		{"render", []string{"render", "--format", "md", "-"}, full},
		{"export", []string{"export", "sarif", "-"}, full},
		{"aggregate", []string{"aggregate", "-"}, structural},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, stdout, stderr := call(tc.args, tc.stdin)
			if code != 1 {
				t.Fatalf("code=%d out=%q err=%q", code, stdout, stderr)
			}
			if stderr != "" {
				t.Errorf("violations wrote %q to stderr, want stdout only", stderr)
			}
			for _, line := range strings.Split(strings.TrimSuffix(stdout, "\n"), "\n") {
				if strings.Count(line, ": ") < 2 {
					t.Errorf("violation line %q is not <path>: <rule>: <message>", line)
				}
			}
		})
	}
}

func TestViolationLineNamesPathRuleAndMessage(t *testing.T) {
	broken := strings.Replace(mustRead(t, fixturePath()), `"overall": "FAIL"`, `"overall": "PASS"`, 1)
	_, stdout, _ := call([]string{"validate", "-"}, broken)
	if !strings.Contains(stdout, "status: status.match: stated ") {
		t.Errorf("want a line starting `status: status.match: stated `, got %q", stdout)
	}
}

func TestOperationalFailuresGoToStderrWithCodeTwo(t *testing.T) {
	source := aggregated(t)
	missingDir := filepath.Join(t.TempDir(), "no-such-dir", "out.md")
	cases := []struct {
		name string
		args []string
	}{
		{"missing input", []string{"validate", "does-not-exist.json"}},
		{"no input path", []string{"validate"}},
		{"unreadable output path", []string{"render", "--format", "md", source, "-o", missingDir}},
		{"unknown flag", []string{"validate", "--nope", source}},
		{"too many positionals", []string{"validate", source, source}},
		{"export without a format", []string{"export", source}},
		{"help after the terminator is a path", []string{"validate", "--", "-h"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, stdout, stderr := call(tc.args, "")
			if code != 2 {
				t.Fatalf("code=%d out=%q err=%q", code, stdout, stderr)
			}
			if stderr == "" {
				t.Error("nothing written to stderr")
			}
			if stdout != "" {
				t.Errorf("wrote %q to stdout, want stderr only", stdout)
			}
		})
	}
}

func TestRenderWritesEveryRegisteredFormat(t *testing.T) {
	source := aggregated(t)
	for _, format := range render.Formats() {
		t.Run(format, func(t *testing.T) {
			code, stdout, stderr := call([]string{"render", "--format", format, source}, "")
			if code != 0 || stdout == "" {
				t.Fatalf("code=%d err=%q", code, stderr)
			}
		})
	}
}

func TestAggregateIsIdempotentThroughTheCLI(t *testing.T) {
	source := aggregated(t)
	code, stdout, stderr := call([]string{"aggregate", source}, "")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, stderr)
	}
	if stdout != mustRead(t, source) {
		t.Error("aggregating an aggregated report changed it")
	}
}

func TestFormatIsCheckedBeforeTheInputIsRead(t *testing.T) {
	code, _, stderr := call([]string{"render", "--format", "pdf", "does-not-exist.json"}, "")
	if code != 2 || !strings.Contains(stderr, "unknown format") {
		t.Errorf("a bad format must be reported as a bad format: code=%d err=%q", code, stderr)
	}
}

func TestNoDocumentIsWrittenForAReportThatBreaksARule(t *testing.T) {
	broken := strings.Replace(mustRead(t, fixturePath()), `"overall": "FAIL"`, `"overall": "PASS"`, 1)
	path := filepath.Join(t.TempDir(), "out.md")
	code, stdout, _ := call([]string{"render", "--format", "md", "-", "-o", path}, broken)
	if code != 1 || !strings.Contains(stdout, "status.match") {
		t.Fatalf("code=%d out=%q", code, stdout)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("render wrote a document from a report that breaks a rule")
	}
}

// inTempDir runs the rest of the test in a new empty directory, so that a
// relative path the command writes or reads can be checked without touching
// the package directory.
func inTempDir(t *testing.T) string {
	t.Helper()
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(here); err != nil {
			t.Fatal(err)
		}
	})
	return dir
}

func TestDashOutputWritesToStandardOutput(t *testing.T) {
	source := aggregated(t)
	inTempDir(t)
	code, stdout, stderr := call([]string{"render", "--format", "md", source, "-o", "-"}, "")
	if code != 0 || !strings.Contains(stdout, "## Criteria") {
		t.Fatalf("code=%d out=%q err=%q", code, stdout, stderr)
	}
	if _, err := os.Stat("-"); !os.IsNotExist(err) {
		t.Error(`-o - created a file named "-" instead of writing to standard output`)
	}
}

func TestTerminatorHandsEveryLaterTokenOver(t *testing.T) {
	raw := mustRead(t, aggregated(t))
	dir := inTempDir(t)
	awkward := filepath.Join(dir, "-weird.json")
	if err := os.WriteFile(awkward, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := call([]string{"validate", "--", "-weird.json"}, "")
	if code != 0 || !strings.HasPrefix(stdout, "valid:") {
		t.Fatalf("code=%d out=%q err=%q", code, stdout, stderr)
	}
}

func TestNilStandardInputIsReportedNotPanicked(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"validate", streamPath}, nil, &out, &errb)
	if code != 2 || !strings.HasPrefix(errb.String(), "error:") {
		t.Errorf("code=%d err=%q, want 2 and an error: line", code, errb.String())
	}
}

// TestALeadingByteOrderMarkIsNotAParseError covers the mark some editors and
// shell redirections write at the head of a UTF-8 file. It is not part of the
// JSON, so a reader handed one would otherwise be told their report is
// malformed, with nothing on screen to show what is wrong with it.
func TestALeadingByteOrderMarkIsNotAParseError(t *testing.T) {
	withBOM := "\ufeff" + mustRead(t, fixturePath())
	code, stdout, stderr := call([]string{"validate", "-"}, withBOM)
	if code != 0 || !strings.HasPrefix(stdout, "valid:") {
		t.Fatalf("stdin: code=%d out=%q err=%q", code, stdout, stderr)
	}
	path := filepath.Join(t.TempDir(), "bom.json")
	if err := os.WriteFile(path, []byte(withBOM), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = call([]string{"validate", path}, "")
	if code != 0 || !strings.HasPrefix(stdout, "valid:") {
		t.Fatalf("file: code=%d out=%q err=%q", code, stdout, stderr)
	}
}

// TestAggregateReachesAnAdvisoryVerdict covers the one status a report asks
// for rather than earns: advisory is stated on the way in, and everything
// else about the status -- the verdict, the rule, the counts, the digests --
// is left empty for aggregate to compute. The result must be a report that
// validates, so the advisory path is closed end to end rather than only
// inside Derive.
func TestAggregateReachesAnAdvisoryVerdict(t *testing.T) {
	rep, err := report.Decode([]byte(mustRead(t, fixturePath())))
	if err != nil {
		t.Fatal(err)
	}
	rep.Counts, rep.Status = report.Counts{}, report.Status{Advisory: true}
	rep.Identity.ReportID, rep.Provenance.EvidenceDigest = "", ""
	raw, err := report.Encode(rep)
	if err != nil {
		t.Fatal(err)
	}
	code, aggregated, stderr := call([]string{"aggregate", "-"}, string(raw))
	if code != 0 {
		t.Fatalf("aggregate: code=%d err=%q", code, stderr)
	}
	if !strings.Contains(aggregated, `"overall": "ADVISORY"`) {
		t.Fatalf("an advisory report did not aggregate to ADVISORY:\n%s", aggregated)
	}
	code, stdout, stderr := call([]string{"validate", "-"}, aggregated)
	if code != 0 {
		t.Fatalf("the aggregated advisory report does not validate: code=%d out=%q err=%q", code, stdout, stderr)
	}
}
