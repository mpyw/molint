package internal

import (
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/nilproof/internal/typeutil"
)

// globalBook is what the pass has proven about the package's pointer
// variables.
//
// A package-level variable is proven non-nil when every store to it in the
// package is proven, and nothing else takes its address. A store from another
// package, to an exported variable, is not seen. A read before the package is
// initialized is not considered either.
//
//declscope:package
type globalBook struct {
	// globals holds each candidate variable, with the stores to it. A
	// variable whose address is taken, or that is never stored, is not one.
	//
	//declscope:private
	globals map[*ssa.Global][]*ssa.Store
	// provenGlobals holds the candidates proven so far.
	//
	//declscope:private
	provenGlobals map[*ssa.Global]bool
}

// collectGlobals finds the package's pointer variables and the stores to
// them. The initializers of package-level variables are in the package's
// init function, which is not a source function, so it is read too.
//
//declscope:package
func (c *checker) collectGlobals(fns []*ssa.Function, pkg *ssa.Package) {
	c.globals = make(map[*ssa.Global][]*ssa.Store)
	c.provenGlobals = make(map[*ssa.Global]bool)
	for _, m := range pkg.Members {
		if g, ok := m.(*ssa.Global); ok && typeutil.IsPointer(g.Type().(*types.Pointer).Elem()) {
			c.globals[g] = nil
		}
	}
	escaped := make(map[*ssa.Global]bool)
	var ops []*ssa.Value
	for _, fn := range append(fns, pkg.Func("init")) {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				for _, op := range instr.Operands(ops[:0]) {
					g, ok := (*op).(*ssa.Global)
					if _, candidate := c.globals[g]; !ok || !candidate {
						continue
					}
					switch in := instr.(type) {
					case *ssa.Store:
						if in.Addr == g && in.Val != g {
							c.globals[g] = append(c.globals[g], in)
							continue
						}
					case *ssa.UnOp:
						if in.Op == token.MUL {
							continue
						}
					}
					escaped[g] = true
				}
			}
		}
	}
	for g, stores := range c.globals {
		if escaped[g] || len(stores) == 0 {
			delete(c.globals, g)
			continue
		}
		c.provenGlobals[g] = true
	}
}

// refreshGlobals takes back each variable with a store no longer proven, and
// reports whether it took any.
//
//declscope:package
func (c *checker) refreshGlobals() bool {
	changed := false
	for g, stores := range c.globals {
		if !c.provenGlobals[g] {
			continue
		}
		p := c.newProof()
		for _, s := range stores {
			if p.nonNil(s.Val, proofSite{block: s.Block()}) != nil {
				c.provenGlobals[g] = false
				changed = true
				break
			}
		}
	}
	return changed
}

// provenGlobal reports whether a package-level variable is proven non-nil: by
// this pass for one of the package, or by the fact of an imported one.
//
//declscope:package
func (c *checker) provenGlobal(g *ssa.Global) bool {
	if g.Pkg.Pkg == c.pass.Pkg {
		return c.provenGlobals[g]
	}
	obj, ok := g.Object().(*types.Var)
	return ok && c.pass.ImportObjectFact(obj, new(GlobalFact))
}

// exportGlobals publishes a fact for each variable proven.
//
//declscope:package
func (c *checker) exportGlobals() {
	for g, ok := range c.provenGlobals {
		if ok {
			c.pass.ExportObjectFact(g.Object(), new(GlobalFact))
		}
	}
}
