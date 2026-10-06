package internal

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/flow"
	"github.com/mpyw/molint/internal/rule"
	"github.com/mpyw/molint/internal/typeutil"
)

// checkReturnNil reports each return that gives a nil pointer result. A
// function literal is exempt, but the body of a range-over-func loop is
// not: its returns are returns of the function around the loop.
//
//declscope:shared // run.go calls it
func (c *checker) checkReturnNil(fn *ssa.Function) {
	owner, rets := fn, []flow.Return(nil)
	switch {
	case flow.IsRangeBody(fn):
		owner = flow.RangeOwner(fn)
		rets = flow.RangeReturns(fn)
	case fn.Parent() != nil:
		return
	default:
		rets = flow.Returns(fn)
	}
	obj, ok := owner.Object().(*types.Func)
	if !ok || c.exempt[obj] {
		return
	}
	sig := owner.Signature
	trailing := typeutil.TrailingOf(sig)
	guard := sig.Results().Len() - 1
	ptr := flow.NilTracer()
	var guardTracer *flow.Tracer
	switch trailing {
	case typeutil.TrailingError:
		guardTracer = flow.NilTracer()
	case typeutil.TrailingBool:
		guardTracer = flow.TrueTracer()
	}
	name := typeutil.FuncName(obj, c.pass.Pkg)
	for _, ret := range rets {
		for i := range sig.Results().Len() {
			t := sig.Results().At(i).Type()
			v := ret.Values[i]
			if !typeutil.IsPointer(t) {
				continue
			}
			var bad bool
			if guardTracer == nil {
				bad = ptr.Is(v, ret.Site)
			} else {
				bad = flow.Pair(ptr, guardTracer, v, ret.Values[guard], ret.Site)
			}
			if bad {
				c.report(ret.Pos, rule.ReturnNil, returnNilMessage(name, typeutil.TypeString(t, c.pass.Pkg), trailing))
			}
		}
	}
}

func returnNilMessage(name, t string, trailing typeutil.Trailing) string {
	switch trailing {
	case typeutil.TrailingError:
		return fmt.Sprintf("%s returns a nil %s with a nil error; return an error, or mo.Option[%s]", name, t, t)
	case typeutil.TrailingBool:
		return fmt.Sprintf("%s returns a nil %s with ok true; return mo.Option[%s]", name, t, t)
	}
	return fmt.Sprintf("%s returns a nil %s; return mo.Option[%s] instead", name, t, t)
}
