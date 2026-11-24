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
	// instr is the store, or the Alloc for its zero value.
	instr ssa.Instruction
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
	return storesUntil(a, b, end, nil)
}

// storesUntil is storesBefore with a stop. A path that meets an instruction
// stop accepts, before any store, brings nothing. A nil stop accepts none.
//
//declscope:shared // escape.go stops the walk where other code may run
func storesUntil(a *ssa.Alloc, b *ssa.BasicBlock, end int, stop func(ssa.Instruction) bool) []storeDef {
	var out []storeDef
	seen := make(map[*ssa.BasicBlock]bool)
	var walk func(b *ssa.BasicBlock, end int)
	walk = func(b *ssa.BasicBlock, end int) {
		d, ok, stopped := storeScan(a, b, end, stop)
		if ok {
			out = append(out, d)
			return
		}
		if stopped {
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
//declscope:shared // pair.go walks back from a load along one edge
func storeLast(a *ssa.Alloc, b *ssa.BasicBlock, end int) (storeDef, bool) {
	d, ok, _ := storeScan(a, b, end, nil)
	return d, ok
}

// storeScan goes back from end in b. It finds the last store into a, or a
// itself. Or it meets an instruction stop accepts first, and then stopped is
// true.
func storeScan(a *ssa.Alloc, b *ssa.BasicBlock, end int, stop func(ssa.Instruction) bool) (d storeDef, ok, stopped bool) {
	for i := end - 1; i >= 0; i-- {
		in := b.Instrs[i]
		if d, ok := storeInto(a, b, in); ok {
			return d, true, false
		}
		if stop != nil && stop(in) {
			return storeDef{}, false, true
		}
	}
	return storeDef{}, false, false
}

// storeInto reports what in, an instruction of b, stores into a: the stored
// value, or the zero value when in is a itself.
//
//declscope:shared // pair.go walks two variables at once
func storeInto(a *ssa.Alloc, b *ssa.BasicBlock, in ssa.Instruction) (storeDef, bool) {
	switch in := in.(type) {
	case *ssa.Store:
		if in.Addr == a {
			return storeDef{value: in.Val, block: b, instr: in}, true
		}
	case *ssa.Alloc:
		if in == a {
			return storeDef{value: storeZero(a), block: b, instr: in}, true
		}
	}
	return storeDef{}, false
}

// storeZero is the zero value of what a holds.
func storeZero(a *ssa.Alloc) *ssa.Const {
	return ssa.NewConst(nil, a.Type().(*types.Pointer).Elem())
}
