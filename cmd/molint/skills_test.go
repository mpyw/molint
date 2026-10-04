package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
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

func TestSkillsVersion(t *testing.T) {
	info := func(v string, ok bool) func() (*debug.BuildInfo, bool) {
		return func() (*debug.BuildInfo, bool) {
			if !ok {
				return nil, false
			}
			return &debug.BuildInfo{Main: debug.Module{Version: v}}, true
		}
	}
	cases := []struct {
		name  string
		stamp string
		info  func() (*debug.BuildInfo, bool)
		want  string
	}{
		{"a release build", "v0.2.0", info("v0.2.0+dirty", true), "0.2.0"},
		{"go install @version", "", info("v0.2.0", true), "0.2.0"},
		{"a checkout", "", info("(devel)", true), "devel"},
		{"no version", "", info("", true), "devel"},
		{"no build info", "", info("", false), "devel"},
	}
	for _, c := range cases {
		if got := skillsVersion(c.stamp, c.info); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
