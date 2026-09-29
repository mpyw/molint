package flow

import (
	"go/token"
	"slices"

	"golang.org/x/tools/go/ssa"
)

// rangeSynthetic is how go/ssa marks the body of a range-over-func loop.
const rangeSynthetic = "range-over-func yield"

// IsRangeBody reports whether fn is the body of a range-over-func loop.
func IsRangeBody(fn *ssa.Function) bool {
	return fn.Synthetic == rangeSynthetic
}

// RangeOwner is the function whose results a range-over-func body returns:
// the nearest enclosing function that is not such a body.
func RangeOwner(fn *ssa.Function) *ssa.Function {
	for IsRangeBody(fn) && fn.Parent() != nil {
		fn = fn.Parent()
	}
	return fn
}

// RangeReturns lists the returns inside the body of a range-over-func loop,
// as returns of the function that encloses the loop. The body is a
// synthetic function literal. A return in it stores every result into
// variables of the enclosing function, then ends the iteration. The body
// also ends at a break, a continue, or the end of an iteration, which store
// no result: go/ssa gives the stores of a return statement its position,
// which tells a return from those.
func RangeReturns(yield *ssa.Function) []Return {
	index := rangeResultIndex(yield)
	n := RangeOwner(yield).Signature.Results().Len()
	var out []Return
	for _, b := range yield.Blocks {
		ret, ok := b.Instrs[len(b.Instrs)-1].(*ssa.Return)
		if !ok {
			continue
		}
		values := make([]ssa.Value, n)
		stored := 0
		for _, in := range b.Instrs {
			st, ok := in.(*ssa.Store)
			if !ok || st.Pos() != ret.Pos() {
				continue
			}
			if fv, ok := st.Addr.(*ssa.FreeVar); ok {
				if i, ok := index[fv]; ok && values[i] == nil {
					values[i] = st.Val
					stored++
				}
			}
		}
		if n > 0 && stored == n {
			out = append(out, Return{Pos: ret.Pos(), Values: values, Site: SiteAt(b)})
		}
	}
	return out
}

// IsRangeReturnStore reports whether st is a store a return statement in a
// range-over-func body makes into a result of the enclosing function.
func IsRangeReturnStore(st *ssa.Store) bool {
	fv, ok := st.Addr.(*ssa.FreeVar)
	if !ok || !IsRangeBody(st.Parent()) {
		return false
	}
	ret, ok := st.Block().Instrs[len(st.Block().Instrs)-1].(*ssa.Return)
	if !ok || st.Pos() != ret.Pos() {
		return false
	}
	_, ok = rangeResultIndex(st.Parent())[fv]
	return ok
}

// rangeResultIndex maps the captured variables of a range-over-func body
// that hold results of the enclosing function to the index of their result.
//
// The body always captures the variable that records how the iteration
// ended, so the enclosing function makes it with a closure.
func rangeResultIndex(yield *ssa.Function) map[*ssa.FreeVar]int {
	parent := yield.Parent()
	mc := rangeClosure(parent, yield)
	out := make(map[*ssa.FreeVar]int)
	var outer map[*ssa.FreeVar]int
	if IsRangeBody(parent) {
		outer = rangeResultIndex(parent)
	}
	for i, b := range mc.Bindings {
		switch b := b.(type) {
		case *ssa.Alloc:
			if r, ok := rangeLoadedByReturn(parent, b); ok {
				out[yield.FreeVars[i]] = r
			}
		case *ssa.FreeVar:
			if r, ok := outer[b]; ok {
				out[yield.FreeVars[i]] = r
			}
		}
	}
	return out
}

// rangeBinding finds the variable of an enclosing function that a captured
// variable of a range-over-func body stands for, and the closure that
// captures it. Nested bodies are followed out to the variable. It is nil
// for any other captured variable.
//
//declscope:package // tracer.go follows a load in a body to it
func rangeBinding(fv *ssa.FreeVar) (*ssa.Alloc, *ssa.MakeClosure) {
	fn := fv.Parent()
	if !IsRangeBody(fn) {
		return nil, nil
	}
	// A captured variable is bound to its address: an Alloc of the
	// enclosing function, or a captured variable of an enclosing body.
	mc := rangeClosure(fn.Parent(), fn)
	switch b := mc.Bindings[slices.Index(fn.FreeVars, fv)].(type) {
	case *ssa.FreeVar:
		return rangeBinding(b)
	default:
		return b.(*ssa.Alloc), mc
	}
}

// rangeClosure finds where parent makes the closure of a range-over-func
// body.
func rangeClosure(parent, fn *ssa.Function) *ssa.MakeClosure {
	var found *ssa.MakeClosure
	for _, b := range parent.Blocks {
		for _, in := range b.Instrs {
			if mc, ok := in.(*ssa.MakeClosure); ok && mc.Fn == fn {
				found = mc
			}
		}
	}
	return found
}

// rangeLoadedByReturn reports which result of fn a holds, when a return of
// fn loads a.
func rangeLoadedByReturn(fn *ssa.Function, a *ssa.Alloc) (int, bool) {
	for _, b := range fn.Blocks {
		ret, ok := b.Instrs[len(b.Instrs)-1].(*ssa.Return)
		if !ok {
			continue
		}
		for i, v := range ret.Results {
			if load, ok := v.(*ssa.UnOp); ok && load.Op == token.MUL && load.X == a {
				return i, true
			}
		}
	}
	return 0, false
}
