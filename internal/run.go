// Package internal implements the molint analysis.
//
// The analysis does not look for nil. It asks each return of a pointer for a
// proof that the pointer is not nil, and reports the returns that have none.
// A proof is built from what cannot be nil: an address, a nil check on the
// way, or a call to a function that is itself proven. Anything else, a
// parameter, a field, a map lookup, an unknown call, is unproven.
//
// Each function is summarized by the pointer results it proves. Summaries
// cross package boundaries as analysis facts, and inside a package they are
// computed to a fixpoint, so that a constructor or a recursive helper is
// proven wherever it is called.
package internal

import (
	"errors"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

// ErrRunWithoutSSA is returned by Run when the pass carries no buildssa
// result, which means the analyzer was registered without its requirement.
var ErrRunWithoutSSA = errors.New("molint: buildssa result missing")

// Run analyzes one package.
func Run(pass *analysis.Pass) (any, error) {
	info, ok := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	if !ok {
		return nil, ErrRunWithoutSSA
	}
	c := newChecker(pass)
	for _, p := range c.directives.Problems() {
		pass.Report(analysis.Diagnostic{Pos: p.Pos, Message: p.Message})
	}
	c.buildSummaries(info.SrcFuncs, info.Pkg)
	c.exportSummaries()
	c.exportGlobals()
	for _, fn := range info.SrcFuncs {
		c.reportFunc(fn)
	}
	// Every function the pass reports on is seen in full, in a package's
	// test variant as in the ordinary one, so an ignore that silenced
	// nothing here silences nothing anywhere.
	for _, pos := range c.directives.Unused() {
		pass.Reportf(pos, "unused molint:ignore directive")
	}
	return nil, nil
}
