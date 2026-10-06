// Package flow follows a value back through one function, to tell whether it
// is a constant that a rule looks for: a nil pointer, a nil error, the
// constant true, or a zero mo.Result.
//
// It follows φ-nodes, conversions between named types, and loads of local
// variables that SSA does not lift. Every step is judged where it happens: a
// nil check on the way to a φ edge, or to a store, stops the value there. It
// follows nothing out of the function: a parameter, a field, or a call's
// result is not one of the constants.
package flow

import (
	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/nilcheck"
)

// Site is where a value is judged: the start of a block, or the edge into it
// that a φ reads the value from.
type Site struct {
	block *ssa.BasicBlock
	from  *ssa.BasicBlock
}

// SiteAt is the site of a use in block b.
func SiteAt(b *ssa.BasicBlock) Site { return Site{block: b} }

// siteEdge is the site of a φ edge from `from` into `to`.
//
//declscope:shared // tracer.go and pair.go judge φ edges
func siteEdge(from, to *ssa.BasicBlock) Site { return Site{block: to, from: from} }

// state is what the nil checks on the way to the site say about v.
//
//declscope:shared // tracer.go reads the checks
func (s Site) state(v ssa.Value) nilcheck.State {
	if s.from != nil {
		return nilcheck.OnEdge(v, s.from, s.block)
	}
	return nilcheck.At(v, s.block)
}
