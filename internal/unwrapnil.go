package internal

import (
	"fmt"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/flow"
	"github.com/mpyw/molint/internal/rule"
	"github.com/mpyw/molint/internal/typeutil"
)

// checkUnwrapNil reports OrEmpty, and OrElse with a nil value, on an Option
// or a Result of a pointer.
//
//declscope:package // calls.go calls it for each call
func (c *checker) checkUnwrapNil(call ssa.CallInstruction, callee *ssa.Function, nils *flow.Tracer) {
	m, name, arg := typeutil.MoMethod(callee)
	if m == typeutil.NotMo || !typeutil.IsPointer(arg) {
		return
	}
	recv := typeutil.RecvString(callee, c.pass.Pkg)
	when, fix := "it is empty", "use Get and check ok, or OrElse with a non-nil value"
	if m == typeutil.Result {
		when, fix = "it is an error", "use Get and check the error, or OrElse with a non-nil value"
	}
	cc := call.Common()
	switch name {
	case "OrEmpty":
		c.report(call.Pos(), rule.UnwrapNil, fmt.Sprintf("OrEmpty on %s gives nil when %s; %s", recv, when, fix))
	case "OrElse":
		if len(cc.Args) == 2 && nils.Is(cc.Args[1], flow.SiteAt(call.Block())) {
			c.report(call.Pos(), rule.UnwrapNil, fmt.Sprintf("OrElse(nil) on %s gives nil when %s; pass a non-nil value", recv, when))
		}
	}
}
