package internal

import (
	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/flow"
)

// checkCalls runs the rules about calls into samber/mo on every static call
// in fn: wrap-nil, unwrap-nil and unwrap-discard.
//
//declscope:package // run.go calls it
func (c *checker) checkCalls(fn *ssa.Function) {
	nils := flow.NilTracer()
	for _, b := range fn.Blocks {
		for _, in := range b.Instrs {
			call, ok := in.(ssa.CallInstruction)
			if !ok {
				continue
			}
			callee := call.Common().StaticCallee()
			if callee == nil {
				continue
			}
			c.checkWrapNil(call, callee, nils)
			c.checkUnwrapNil(call, callee, nils)
			if v, ok := call.(*ssa.Call); ok {
				c.checkUnwrapDiscard(v, callee)
			}
		}
	}
}
