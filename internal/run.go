// Package internal implements the molint rules.
//
// Each rule has its own file. They share one pass's state, in checker.go:
// the directives, the files that are generated, and the methods that
// implement an interface. The questions they ask of types and of SSA values
// live in the packages below this one.
package internal

import (
	"errors"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

// ErrRunWithoutSSA is returned by Run when the pass carries no buildssa
// result, which means the analyzer was registered without its requirement.
var ErrRunWithoutSSA = errors.New("molint: buildssa result missing")

// Run analyzes one package.
func Run(pass *analysis.Pass, cfg Config) (any, error) {
	info, ok := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	if !ok {
		return nil, ErrRunWithoutSSA
	}
	c := newChecker(pass, cfg)
	for _, p := range c.directives.Problems() {
		c.reportf(p.Pos, p.Message)
	}
	c.checkShapes()
	for _, fn := range runFuncs(info) {
		// Nothing is reported in a generated file, so it is not analyzed.
		if c.inGenerated(fn.Pos()) {
			continue
		}
		c.checkReturnNil(fn)
		c.checkCalls(fn)
		c.checkResultZero(fn)
		c.checkFieldNil(fn)
	}
	// Every function the pass reports on is seen in full, in a package's
	// test variant as in the ordinary one, so an ignore that silenced
	// nothing here silences nothing anywhere.
	for _, pos := range c.directives.Unused(cfg.On) {
		c.reportf(pos, "unused molint:ignore directive")
	}
	c.flush()
	return nil, nil
}

// runFuncs lists the functions of the package with a body: the source
// functions, and the function literals in them.
func runFuncs(info *buildssa.SSA) []*ssa.Function {
	var out []*ssa.Function
	for _, fn := range info.SrcFuncs {
		if fn.Blocks != nil {
			out = append(out, fn)
		}
	}
	return out
}
