// Package molint enforces the use of github.com/samber/mo.
//
// Absence is mo.Option, and failure may be mo.Result, rather than a nil
// pointer, a trailing bool, or a trailing error. An Option or a Result must
// not hold or give a nil it should not. The rules are listed in
// design/rules.md, and each has a flag of its name.
package molint

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"

	"github.com/mpyw/molint/internal"
	"github.com/mpyw/molint/internal/rule"
)

// Analyzer enforces the use of github.com/samber/mo.
var Analyzer = &analysis.Analyzer{
	Name:     "molint",
	Doc:      "enforces the use of github.com/samber/mo",
	URL:      "https://github.com/mpyw/molint",
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
	Run:      runAnalyzer,
}

// analyzerFlagOf holds the flag of each rule.
var analyzerFlagOf = make(map[rule.Name]*bool)

func init() {
	for _, r := range rule.All {
		analyzerFlagOf[r] = Analyzer.Flags.Bool(string(r), rule.OnByDefault(r), "report "+string(r))
	}
}

func runAnalyzer(pass *analysis.Pass) (any, error) {
	return internal.Run(pass, internal.Config{On: func(r rule.Name) bool { return *analyzerFlagOf[r] }})
}
