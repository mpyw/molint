package flow

import (
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/ssa"
)

// EscapedStoresRead lists the stores into a that come after its address
// escapes, and that a direct read of a sees with nothing on the way that
// may run other code. A direct read is a load of a. It lists none when a is
// followed, since its loads are judged then. The Alloc's own zero is never
// listed: it comes before any escape.
//
// Once the address is out, any code may write a. So a path from the store
// to the read ends at a call, a go or defer statement, the deferred calls
// at a return, a channel operation, a select, or a store through a pointer
// that may be a's address. An instruction this file does not know ends it
// too.
func EscapedStoresRead(a *ssa.Alloc) []*ssa.Store {
	var escapes []ssa.Instruction
	for _, r := range *a.Referrers() {
		if !tracerFollows(a, &[]ssa.Instruction{r}, false) {
			escapes = append(escapes, r)
		}
	}
	if len(escapes) == 0 {
		return nil
	}
	stop := func(in ssa.Instruction) bool { return escapeInterrupts(a, in) }
	var out []*ssa.Store
	for _, r := range *a.Referrers() {
		load, ok := r.(*ssa.UnOp)
		if !ok || escapeSelfCopy(a, load) {
			continue
		}
		b := load.Block()
		for _, d := range storesUntil(a, b, slices.Index(b.Instrs, r), stop) {
			st, ok := d.instr.(*ssa.Store)
			if ok && !slices.Contains(out, st) && escapeBefore(escapes, st) {
				out = append(out, st)
			}
		}
	}
	return out
}

// escapeSelfCopy reports whether load only stores its value back into a.
// A return in a function with a defer does this with a named result: it
// stores the result, runs the deferred calls, and loads it again. The value
// is not used on the way.
func escapeSelfCopy(a *ssa.Alloc, load *ssa.UnOp) bool {
	refs := *load.Referrers()
	for _, r := range refs {
		st, ok := r.(*ssa.Store)
		if !ok || st.Addr != a {
			return false
		}
	}
	return len(refs) > 0
}

// escapeBefore reports whether one of escapes may run before st: earlier in
// its block, or in a block from which st's block can be reached.
func escapeBefore(escapes []ssa.Instruction, st *ssa.Store) bool {
	target := st.Block()
	for _, e := range escapes {
		b := e.Block()
		if b == target && slices.Index(b.Instrs, e) < slices.Index(b.Instrs, ssa.Instruction(st)) {
			return true
		}
		seen := make(map[*ssa.BasicBlock]bool)
		work := slices.Clone(b.Succs)
		for len(work) > 0 {
			n := work[len(work)-1]
			work = work[:len(work)-1]
			if n == target {
				return true
			}
			if !seen[n] {
				seen[n] = true
				work = append(work, n.Succs...)
			}
		}
	}
	return false
}

// escapeInterrupts reports whether in may run code other than the
// function's own, or may write a through another pointer. Anything not
// listed as plain counts.
func escapeInterrupts(a *ssa.Alloc, in ssa.Instruction) bool {
	switch in := in.(type) {
	case *ssa.Call:
		b, ok := in.Call.Value.(*ssa.Builtin)
		return !ok || !escapePlainBuiltin(b.Name())
	case *ssa.UnOp:
		return in.Op == token.ARROW
	case *ssa.Store:
		return escapeMayAlias(a, in.Addr)
	case *ssa.Alloc, *ssa.BinOp, *ssa.ChangeInterface, *ssa.ChangeType,
		*ssa.Convert, *ssa.DebugRef, *ssa.Extract, *ssa.Field, *ssa.FieldAddr,
		*ssa.If, *ssa.Index, *ssa.IndexAddr, *ssa.Jump, *ssa.Lookup,
		*ssa.MakeChan, *ssa.MakeClosure, *ssa.MakeInterface, *ssa.MakeMap,
		*ssa.MakeSlice, *ssa.MapUpdate, *ssa.MultiConvert, *ssa.Next,
		*ssa.Phi, *ssa.Range, *ssa.Slice, *ssa.SliceToArrayPointer,
		*ssa.TypeAssert:
		return false
	}
	return true
}

// escapePlainBuiltin reports whether a builtin never runs other code.
func escapePlainBuiltin(name string) bool {
	switch name {
	case "append", "cap", "complex", "copy", "imag", "len", "max", "min", "real":
		return true
	}
	return false
}

// escapeMayAlias reports whether addr may be a's address, other than a
// itself. A new variable, a global, a field or an element is another
// variable.
func escapeMayAlias(a *ssa.Alloc, addr ssa.Value) bool {
	switch addr.(type) {
	case *ssa.Alloc, *ssa.Global, *ssa.FieldAddr, *ssa.IndexAddr:
		return false
	}
	return types.Identical(addr.Type(), a.Type())
}
