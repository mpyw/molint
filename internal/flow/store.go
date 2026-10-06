package flow

import (
	"go/types"
	"slices"

	"golang.org/x/tools/go/ssa"
)

// storeDef is a value stored into a variable, and the block of the store.
//
//declscope:shared // tracer.go and pair.go judge the stored values
type storeDef struct {
	value ssa.Value
	block *ssa.BasicBlock
}

// storesReaching lists the values that the stores reaching load bring. The
// allocation itself counts as a store of the zero value, since SSA emits no
// store for a variable declared without a value.
//
//declscope:shared // tracer.go follows a load through it
func storesReaching(a *ssa.Alloc, load ssa.Instruction) []storeDef {
	b := load.Block()
	return storesBefore(a, b, slices.Index(b.Instrs, load))
}

// storesBefore lists the values that the stores reaching instruction end of
// b bring, walking back through the predecessors.
//
//declscope:shared // pair.go starts the walk at the end of a predecessor
func storesBefore(a *ssa.Alloc, b *ssa.BasicBlock, end int) []storeDef {
	var out []storeDef
	seen := make(map[*ssa.BasicBlock]bool)
	var walk func(b *ssa.BasicBlock, end int)
	walk = func(b *ssa.BasicBlock, end int) {
		if d, ok := storeLast(a, b, end); ok {
			out = append(out, d)
			return
		}
		for _, p := range b.Preds {
			if !seen[p] {
				seen[p] = true
				walk(p, len(p.Instrs))
			}
		}
	}
	walk(b, end)
	return out
}

// storeLast finds the last store into a in b before end, or a itself.
//
//declscope:shared // pair.go asks whether a block stores before a load
func storeLast(a *ssa.Alloc, b *ssa.BasicBlock, end int) (storeDef, bool) {
	for i := end - 1; i >= 0; i-- {
		switch in := b.Instrs[i].(type) {
		case *ssa.Store:
			if in.Addr == a {
				return storeDef{value: in.Val, block: b}, true
			}
		case *ssa.Alloc:
			if in == a {
				return storeDef{value: storeZero(a), block: b}, true
			}
		}
	}
	return storeDef{}, false
}

// storeZero is the zero value of what a holds.
//
//declscope:shared // pair.go walks the stores itself
func storeZero(a *ssa.Alloc) *ssa.Const {
	return ssa.NewConst(nil, a.Type().(*types.Pointer).Elem())
}
