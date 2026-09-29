package internal

import (
	"fmt"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/rule"
	"github.com/mpyw/molint/internal/typeutil"
)

// checkUnwrapDiscard reports Get on an Option or a Result whose value is
// used while its second result is discarded. A result is discarded when no
// Extract of it exists, or its Extract has no referrer.
//
//declscope:package // calls.go calls it for each call
func (c *checker) checkUnwrapDiscard(call *ssa.Call, callee *ssa.Function) {
	m, name, _ := typeutil.MoMethod(callee)
	if m == typeutil.NotMo || name != "Get" {
		return
	}
	used := [2]bool{}
	for _, r := range *call.Referrers() {
		if x, ok := r.(*ssa.Extract); ok && x.Index < 2 && len(*x.Referrers()) > 0 {
			used[x.Index] = true
		}
	}
	if !used[0] || used[1] {
		return
	}
	recv := typeutil.RecvString(callee, c.pass.Pkg)
	msg := fmt.Sprintf("Get on %s discards ok; check it, or use OrElse", recv)
	if m == typeutil.Result {
		msg = fmt.Sprintf("Get on %s discards the error; check it", recv)
	}
	c.report(call.Pos(), rule.UnwrapDiscard, msg)
}
