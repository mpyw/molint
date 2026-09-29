package internal

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/flow"
	"github.com/mpyw/molint/internal/rule"
	"github.com/mpyw/molint/internal/typeutil"
)

// checkResultZero reports each use of a zero mo.Result: a return, an
// argument, a store, a send, a map update, and a conversion to an
// interface. A comparison is not a use. Neither is a store into a local
// variable of the same function, whose loads are judged instead, nor the
// store a return makes into its result variable, since the return is judged
// once. A function literal that stores into a variable it captures makes a
// use: the loads of that variable are no longer followed.
//
//declscope:package // run.go calls it
func (c *checker) checkResultZero(fn *ssa.Function) {
	zeros := flow.CheckingTracer(isResultZero)
	check := func(v ssa.Value, pos token.Pos, at flow.Site) {
		if m, _ := typeutil.MoOf(v.Type()); m == typeutil.Result && zeros.Is(v, at) {
			c.report(pos, rule.ResultZero, fmt.Sprintf("a zero %s is Ok with a zero value; build it with mo.Ok or mo.Err", typeutil.TypeString(v.Type(), c.pass.Pkg)))
		}
	}
	rets := flow.Returns(fn)
	if flow.IsRangeBody(fn) {
		rets = append(rets, flow.RangeReturns(fn)...)
	}
	for _, ret := range rets {
		for _, v := range ret.Values {
			check(v, ret.Pos, ret.Site)
		}
	}
	for _, b := range fn.Blocks {
		at := flow.SiteAt(b)
		for _, in := range b.Instrs {
			switch in := in.(type) {
			case ssa.CallInstruction:
				for _, a := range in.Common().Args {
					check(a, in.Pos(), at)
				}
			case *ssa.Store:
				switch in.Addr.(type) {
				case *ssa.Alloc:
				case *ssa.FreeVar:
					if !flow.IsRangeReturnStore(in) {
						check(in.Val, in.Pos(), at)
					}
				default:
					check(in.Val, in.Pos(), at)
				}
			case *ssa.MapUpdate:
				check(in.Key, in.Pos(), at)
				check(in.Value, in.Pos(), at)
			case *ssa.Send:
				check(in.X, in.Pos(), at)
			case *ssa.MakeInterface:
				check(in.X, resultZeroPos(in), at)
			}
		}
	}
}

// resultZeroPos is where a conversion to an interface is reported. An
// implicit conversion has no position of its own, so it takes the position
// of the instruction that uses it, looking through the φs it flows into.
func resultZeroPos(mi *ssa.MakeInterface) token.Pos {
	if mi.Pos().IsValid() {
		return mi.Pos()
	}
	seen := make(map[*ssa.Phi]bool)
	var find func(v ssa.Value) token.Pos
	find = func(v ssa.Value) token.Pos {
		for _, r := range *v.Referrers() {
			pos := r.Pos()
			if phi, ok := r.(*ssa.Phi); ok {
				pos = token.NoPos
				if !seen[phi] {
					seen[phi] = true
					pos = find(phi)
				}
			}
			if pos.IsValid() {
				return pos
			}
		}
		return token.NoPos
	}
	return find(mi)
}

// isResultZero reports whether c is the zero value of a mo.Result.
func isResultZero(c *ssa.Const) bool {
	m, _ := typeutil.MoOf(c.Type())
	return m == typeutil.Result && c.Value == nil
}
