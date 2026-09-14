package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/mickeyyaya/v-eval/core/render"
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

func TestUsageGoesToStderrAndNamesEveryCommand(t *testing.T) {
	for _, args := range [][]string{nil, {"-h"}, {"--help"}} {
		code, stdout, stderr := call(args, "")
		if code != 2 {
			t.Errorf("%v: code=%d, want 2", args, code)
		}
		if stdout != "" {
			t.Errorf("%v: usage wrote %q to stdout, want stderr only", args, stdout)
		}
		for _, cmd := range commands {
			if !strings.Contains(stderr, cmd.Name) {
				t.Errorf("%v: usage does not name %q", args, cmd.Name)
			}
		}
	}
	code, _, stderr := call([]string{"frobnicate"}, "")
	if code != 2 || !strings.Contains(stderr, `error: unknown command "frobnicate"`) {
		t.Errorf("unknown command: code=%d err=%q", code, stderr)
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

func TestReorderArgsMovesFlagsBeforePositionals(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"flags already first", []string{"--format", "md", "r.json"}, []string{"--format", "md", "r.json"}},
		{"flag after positional", []string{"r.json", "--format", "md"}, []string{"--format", "md", "r.json"}},
		{"joined value", []string{"r.json", "--format=md"}, []string{"--format=md", "r.json"}},
		{"two flags around two positionals", []string{"sarif", "-o", "x", "r.json"}, []string{"-o", "x", "sarif", "r.json"}},
		{"dash stays a positional", []string{"-", "-o", "x"}, []string{"-o", "x", "-"}},
		{"trailing flag without a value", []string{"r.json", "-o"}, []string{"-o", "r.json"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := reorderArgs(tc.in); !slices.Equal(got, tc.want) {
				t.Errorf("reorderArgs(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
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
		{"subcommand help", []string{"render", "-h"}},
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
