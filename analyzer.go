// Package molint enforces the use of github.com/samber/mo.
//
// Absence is mo.Option, and failure may be mo.Result, rather than a nil
// pointer, a trailing bool, or a trailing error. An Option or a Result must
// not hold or give a nil it should not. The rules are listed in
// design/rules.md, and each has a flag of its name.
package molint

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"

	"github.com/mpyw/molint/internal"
	"github.com/mpyw/molint/internal/rule"
)

// Analyzer enforces the use of github.com/samber/mo. It takes its settings
// from flags, as the molint command and go vet give them. A program that
// holds its own settings builds an analyzer with New.
var Analyzer = analyzerWithFlags()

// Config holds the settings of an analyzer that New builds. The zero Config
// keeps every rule at its default.
type Config struct {
	// Rules turns a rule on or off by its name, such as "return-error". A
	// rule left out keeps its default.
	Rules map[string]bool
}

// New builds an analyzer with the settings of cfg, and no flags. Each call
// builds one of its own, so two of them share nothing. It fails on a rule
// it does not know.
func New(cfg Config) (*analysis.Analyzer, error) {
	on := make(map[rule.Name]bool, len(cfg.Rules))
	for name, v := range cfg.Rules {
		if !rule.Known(name) {
			return nil, fmt.Errorf("unknown rule %q", name)
		}
		on[rule.Name(name)] = v
	}
	c := internal.Config{On: func(r rule.Name) bool {
		if v, ok := on[r]; ok {
			return v
		}
		return rule.OnByDefault(r)
	}}
	return analyzerBuild(c), nil
}

// analyzerWithFlags builds the analyzer with a flag for each rule. The flags
// live in the closure of Run, so nothing is shared but the analyzer itself.
func analyzerWithFlags() *analysis.Analyzer {
	on := make(map[rule.Name]*bool, len(rule.All))
	a := analyzerBuild(internal.Config{On: func(r rule.Name) bool { return *on[r] }})
	for _, r := range rule.All {
		on[r] = a.Flags.Bool(string(r), rule.OnByDefault(r), "report "+string(r))
	}
	return a
}

// analyzerBuild builds an analyzer that runs with the settings of c.
func analyzerBuild(c internal.Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "molint",
		Doc:      "enforces the use of github.com/samber/mo",
		URL:      "https://github.com/mpyw/molint",
		Requires: []*analysis.Analyzer{buildssa.Analyzer},
		Run: func(pass *analysis.Pass) (any, error) {
			return internal.Run(pass, c)
		},
	}
}
