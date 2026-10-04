package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSkillInstall installs the skills into a directory, as `molint skill
// install --dir` does, and checks that the copy is the skill and is stamped
// with this tool.
func TestSkillInstall(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "skills")
	if err := skills.Run(t.Context(), []string{"install", "--dir", dest}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dest, "molint-authoring", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "# Writing code under molint") {
		t.Errorf("the installed manifest is not the skill:\n%.200s", body)
	}
	if !strings.Contains(string(body), "x-embedded-by: molint") {
		t.Errorf("the installed manifest carries no stamp:\n%.400s", body)
	}
}
