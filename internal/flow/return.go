package flow

import (
	"go/token"

	"golang.org/x/tools/go/ssa"
)

// Return is what one return statement gives.
type Return struct {
	// Pos is the position of the return statement.
	Pos token.Pos
	// Values holds a value per result.
	Values []ssa.Value
	// Site is where the values are judged.
	Site Site
}

// Returns lists the returns of fn. A return that stores its results into
// their variables and loads them again, as a function with a defer does, is
// read as if it returned the stored values. A return that loads variables it
// did not store, a bare return, keeps the loads, which a Tracer follows.
func Returns(fn *ssa.Function) []Return {
	var out []Return
	for _, b := range fn.Blocks {
		ret, ok := b.Instrs[len(b.Instrs)-1].(*ssa.Return)
		if !ok {
			continue
		}
		values := make([]ssa.Value, len(ret.Results))
		for i, v := range ret.Results {
			values[i] = returnStored(ret, v)
		}
		out = append(out, Return{Pos: ret.Pos(), Values: values, Site: SiteAt(b)})
	}
	return out
}

// returnStored resolves a load of a result variable to the value the return
// statement stored into it. go/ssa gives those stores the position of the
// return statement, which tells them from an assignment just before a bare
// return. Only the instructions the statement emits may sit between the
// store and the return: stores of the other results, the run of the
// defers, and the loads.
func returnStored(ret *ssa.Return, v ssa.Value) ssa.Value {
	load, ok := v.(*ssa.UnOp)
	if !ok || load.Op != token.MUL {
		return v
	}
	a, ok := load.X.(*ssa.Alloc)
	if !ok {
		return v
	}
	// A load of an address is the only unary operation a return statement
	// emits after its stores.
	instrs := ret.Block().Instrs
	for i := len(instrs) - 2; i >= 0; i-- {
		switch in := instrs[i].(type) {
		case *ssa.Store:
			if in.Addr == a {
				if in.Pos() != ret.Pos() {
					return v
				}
				return in.Val
			}
		case *ssa.UnOp, *ssa.RunDefers:
		default:
			return v
		}
	}
	return v
}
