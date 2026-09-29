package internal

import (
	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/flow"
	"github.com/mpyw/molint/internal/rule"
	"github.com/mpyw/molint/internal/typeutil"
)

// checkWrapNil reports a nil value given to mo.Some or mo.Ok of a type whose
// nil breaks on use, and a nil error given to mo.Err.
//
//declscope:package // calls.go calls it for each call
func (c *checker) checkWrapNil(call ssa.CallInstruction, callee *ssa.Function, nils *flow.Tracer) {
	cc := call.Common()
	if len(cc.Args) == 0 {
		return
	}
	at := flow.SiteAt(call.Block())
	switch name := typeutil.MoFunc(callee); name {
	case "Some", "Ok":
		if targs := callee.TypeArgs(); len(targs) == 1 && typeutil.NilBreaks(targs[0]) && nils.Is(cc.Args[0], at) {
			msg := "mo.Ok is given nil; pass a non-nil value"
			if name == "Some" {
				msg = "mo.Some is given nil; pass a non-nil value, or use mo.None"
			}
			c.report(call.Pos(), rule.WrapNil, msg)
		}
	case "Err":
		if nils.Is(cc.Args[0], at) {
			c.report(call.Pos(), rule.WrapNil, "mo.Err is given a nil error; pass a non-nil error")
		}
	}
}
