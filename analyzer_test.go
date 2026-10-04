package molint_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mpyw/molint"
)

// No test in this file may call t.Parallel: the flags live on the global
// molint.Analyzer, and a test that sets one would leak it into the others.

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), molint.Analyzer,
		"returnnil",
		"returnnilpair",
		"returnbool",
		"returnerroroff",
		"fieldniloff",
		"implementing",
		"implementingtest",
		"wrapnil",
		"resultzero",
		"unwrapnil",
		"unwrapdiscard",
		"directives",
		"generated",
	)
}

func TestReturnError(t *testing.T) {
	restore := setFlag(t, "return-error", "true")
	analysistest.Run(t, analysistest.TestData(), molint.Analyzer, "returnerror")
	restore()
	// With the flag back to its default, the same kinds of code report
	// nothing.
	analysistest.Run(t, analysistest.TestData(), molint.Analyzer, "returnerroroff")
}

func TestFieldNil(t *testing.T) {
	restoreStore := setFlag(t, "field-nil-store", "true")
	restoreCompare := setFlag(t, "field-nil-compare", "true")
	analysistest.Run(t, analysistest.TestData(), molint.Analyzer, "fieldnilstore", "fieldnilcompare")
	restoreStore()
	restoreCompare()
	// With the flags back to their defaults, the same kinds of code report
	// nothing.
	analysistest.Run(t, analysistest.TestData(), molint.Analyzer, "fieldniloff")
}

// TestReturnBoolOff runs with the default flags, then with return-bool off,
// then with the default again, to prove that the flags are read on each run
// rather than once.
func TestReturnBoolOff(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), molint.Analyzer, "returnbool")
	restore := setFlag(t, "return-bool", "false")
	analysistest.Run(t, analysistest.TestData(), molint.Analyzer, "returnbooloff")
	restore()
	analysistest.Run(t, analysistest.TestData(), molint.Analyzer, "returnbool")
}

// TestFlags pins the name and the default of each rule's flag.
func TestFlags(t *testing.T) {
	defaults := map[string]string{
		"return-nil":        "true",
		"return-bool":       "true",
		"return-error":      "false",
		"field-nil-store":   "false",
		"field-nil-compare": "false",
		"wrap-nil":          "true",
		"result-zero":       "true",
		"unwrap-nil":        "true",
		"unwrap-discard":    "true",
	}
	for name, want := range defaults {
		f := molint.Analyzer.Flags.Lookup(name)
		if f == nil {
			t.Errorf("flag %s is not defined", name)
			continue
		}
		if f.DefValue != want {
			t.Errorf("flag %s defaults to %s, want %s", name, f.DefValue, want)
		}
	}
}

// setFlag sets a flag of the analyzer. It returns a function that restores
// the default, which also runs when the test ends.
func setFlag(t *testing.T, name, value string) func() {
	t.Helper()
	f := molint.Analyzer.Flags.Lookup(name)
	if f == nil {
		t.Fatalf("flag %s is not defined", name)
	}
	def := f.DefValue
	if err := molint.Analyzer.Flags.Set(name, value); err != nil {
		t.Fatalf("setting %s=%s: %v", name, value, err)
	}
	restore := func() {
		if err := molint.Analyzer.Flags.Set(name, def); err != nil {
			t.Errorf("resetting %s: %v", name, err)
		}
	}
	t.Cleanup(restore)
	return restore
}
