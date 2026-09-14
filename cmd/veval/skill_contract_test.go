package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// skillSilentCommands are the subcommands SKILL.md deliberately does not name.
// version answers about this build rather than about a report, so no step of
// the skill's workflow runs it. The test below requires every name here to
// still be a command, so an exemption cannot outlive the command it exempts.
var skillSilentCommands = []string{"version"}

func TestSkillNamesTheRealSubcommandsAndHosts(t *testing.T) {
	text := mustRead(t, filepath.Join("..", "..", "skills", "evaluate-output", "SKILL.md"))
	for _, cmd := range commands() {
		if slices.Contains(skillSilentCommands, cmd.Name) {
			continue
		}
		if !strings.Contains(text, "veval "+cmd.Name) {
			t.Fatalf("SKILL.md does not mention %q", "veval "+cmd.Name)
		}
	}
	for _, name := range skillSilentCommands {
		if !slices.ContainsFunc(commands(), func(cmd subcommand) bool { return cmd.Name == name }) {
			t.Fatalf("%q is exempted from SKILL.md but is no longer a subcommand", name)
		}
	}
	// The full invocations pin the flags as well as the names, which the
	// table alone cannot: a flag renamed leaves every command name intact.
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
