package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillNamesTheRealSubcommandsAndHosts(t *testing.T) {
	text := mustRead(t, filepath.Join("..", "..", "skills", "evaluate-output", "SKILL.md"))
	// Every name comes from the table, so a command added to the build is a
	// command the skill has to name. There is no exemption list: a verb the
	// workflow never reaches is a verb the workflow has not accounted for.
	for _, cmd := range commands() {
		if !strings.Contains(text, "veval "+cmd.Name) {
			t.Fatalf("SKILL.md does not mention %q", "veval "+cmd.Name)
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
