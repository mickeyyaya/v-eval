package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/mickeyyaya/v-eval/internal/version"
	"github.com/mickeyyaya/v-eval/schema"
)

func TestVersionStringIncludesSchema(t *testing.T) {
	var out bytes.Buffer
	if code := run([]string{"version"}, nil, &out, io.Discard); code != 0 || !strings.Contains(out.String(), "schema 0.1.0") || !strings.HasPrefix(out.String(), "veval ") {
		t.Fatalf("code=%d out=%q", code, out.String())
	}
}

// TestVersionNamesTheBuildAndTheSchema pins the whole line: what a build was
// stamped with, and which schema version that build reads. A release note or
// a bug report quotes this line, so its shape is part of the contract.
func TestVersionNamesTheBuildAndTheSchema(t *testing.T) {
	code, stdout, stderr := call([]string{"version"}, "")
	want := "veval " + version.String() + " schema " + schema.Version + "\n"
	if code != exitOK || stdout != want {
		t.Fatalf("code=%d out=%q, want 0 and %q", code, stdout, want)
	}
	if stderr != "" {
		t.Errorf("version wrote %q to stderr, want stdout only", stderr)
	}
	if !strings.Contains(stdout, version.Version) || !strings.Contains(stdout, "("+version.BuildDigest+")") {
		t.Errorf("out=%q carries neither the stamped version %q nor the digest %q",
			stdout, version.Version, version.BuildDigest)
	}
}

// TestVersionTakesNoArguments: the answer does not depend on anything the
// caller names, so naming something is a mistake worth reporting rather than
// a word to ignore.
func TestVersionTakesNoArguments(t *testing.T) {
	code, stdout, stderr := call([]string{"version", "report.json"}, "")
	if code != exitError || !strings.Contains(stderr, "error:") {
		t.Fatalf("code=%d err=%q, want 2 and an `error:` line", code, stderr)
	}
	if stdout != "" {
		t.Errorf("wrote %q to stdout, want stderr only", stdout)
	}
}

// TestCommandHelpCarriesTheTableUsageLine holds the two places a usage line
// could be spelled to one spelling: `veval render -h` must answer with the
// same line `veval -h` lists for render -- and answer it the same way, on
// standard output with exit code 0, because help asked for is help given.
func TestCommandHelpCarriesTheTableUsageLine(t *testing.T) {
	for _, cmd := range commands() {
		t.Run(cmd.Name, func(t *testing.T) {
			code, stdout, stderr := call([]string{cmd.Name, "-h"}, "")
			if code != exitOK {
				t.Errorf("code=%d, want 0", code)
			}
			if !strings.Contains(stdout, cmd.Usage) {
				t.Errorf("help for %s does not carry %q: %q", cmd.Name, cmd.Usage, stdout)
			}
			if stderr != "" {
				t.Errorf("help wrote %q to stderr, want stdout only", stderr)
			}
		})
	}
}
