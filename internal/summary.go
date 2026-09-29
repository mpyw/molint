package internal

import (
	"go/types"
	"slices"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/typeutil"
)

// summaryBook is what the pass has proven about the package's functions.
//
//declscope:package
type summaryBook struct {
	// funcs holds, for each function with a pointer result, the bit set of
	// its pointer results proven non-nil.
	//
	//declscope:private
	funcs map[*ssa.Function]uint64
	// declared finds the function of a declared object of the package, for a
	// call that reaches it through a wrapper.
	//
	//declscope:private
	declared map[*types.Func]*ssa.Function
}

// buildSummaries proves the functions of the package to a fixpoint.
//
// Every function starts out with all its pointer results proven, and every
// round takes back what no longer holds. A function that calls another of
// the package, itself included, is judged against the other's summary of the
// moment. So recursion is proven when the base cases are, and a round that
// changes nothing is the fixpoint. Rounds only take proofs back, so they end.
//
//declscope:package
func (c *checker) buildSummaries(fns []*ssa.Function, pkg *ssa.Package) {
	c.funcs = make(map[*ssa.Function]uint64)
	c.declared = make(map[*types.Func]*ssa.Function)
	for _, fn := range fns {
		if obj := typeutil.Func(fn); obj != nil {
			c.declared[obj] = fn
		}
		// A result past the 64th shifts its bit out, so it is never proven.
		var bits uint64
		for _, i := range typeutil.PointerResults(fn.Signature) {
			bits |= 1 << i
		}
		if bits != 0 {
			c.funcs[fn] = bits
		}
	}
	c.collectGlobals(fns, pkg)
	for changed := true; changed; {
		changed = false
		for _, fn := range fns {
			old, ok := c.funcs[fn]
			if !ok {
				continue
			}
			if got := c.judge(fn).proven & old; got != old {
				c.funcs[fn] = got
				changed = true
			}
		}
		if c.refreshGlobals() {
			changed = true
		}
	}
}

// provenInSummary reports whether result i of fn is proven non-nil: by the
// summary of a function of the package, or by the fact of an imported one.
//
//declscope:package
func (c *checker) provenInSummary(fn *ssa.Function, i int) bool {
	if o := fn.Origin(); o != nil {
		fn = o
	}
	if bits, ok := c.funcs[fn]; ok {
		return bits&(1<<i) != 0
	}
	obj := typeutil.Func(fn)
	if obj != nil && obj.Pkg() == c.pass.Pkg {
		bits, ok := c.funcs[c.declared[obj]]
		return ok && bits&(1<<i) != 0
	}
	var f Fact
	return obj != nil && c.pass.ImportObjectFact(obj, &f) && slices.Contains(f.NonNil, i)
}

// exportSummaries publishes a fact for each declared function with a proven
// pointer result.
//
//declscope:package
func (c *checker) exportSummaries() {
	for fn, bits := range c.funcs {
		obj := typeutil.Func(fn)
		if obj == nil || bits == 0 {
			continue
		}
		f := new(Fact)
		for i := range 64 {
			if bits&(1<<i) != 0 {
				f.NonNil = append(f.NonNil, i)
			}
		}
		c.pass.ExportObjectFact(obj, f)
	}
}
