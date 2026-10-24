package molint_test

import (
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mpyw/molint"
)

// TestNew runs analyzers that New builds, side by side with different
// settings, to prove that they share nothing with each other or with the
// flags of molint.Analyzer.
func TestNew(t *testing.T) {
	t.Parallel()
	on := mustNew(t, molint.Config{Rules: map[string]bool{"return-error": true}})
	off := mustNew(t, molint.Config{Rules: map[string]bool{"return-bool": false}})
	def := mustNew(t, molint.Config{})
	if def.Flags.Lookup("return-error") != nil {
		t.Error("an analyzer New built has flags, want none")
	}
	analysistest.Run(t, analysistest.TestData(), on, "returnerror")
	analysistest.Run(t, analysistest.TestData(), off, "returnbooloff")
	analysistest.Run(t, analysistest.TestData(), def, "returnbool", "returnerroroff")
}

func TestNewRejectsAnUnknownRule(t *testing.T) {
	t.Parallel()
	_, err := molint.New(molint.Config{Rules: map[string]bool{"field-nil": true}})
	if err == nil || !strings.Contains(err.Error(), `unknown rule "field-nil"`) {
		t.Errorf("got error %v, want one naming the rule", err)
	}
}

func mustNew(t *testing.T, cfg molint.Config) *analysis.Analyzer {
	t.Helper()
	a, err := molint.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
