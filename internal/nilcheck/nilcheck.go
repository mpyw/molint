// Package nilcheck reads what the branches above a point say about a value
// being nil.
//
// A value compared with nil by an if is known on each edge the if leaves by.
// A struct compared with its zero value counts the same way, so that "nil"
// here means the zero constant of the value's type.
// It stays known in every block that edge dominates, since an SSA value never
// changes: a use of the value is dominated by its definition, so no path
// defines it again between the check and the use.
package nilcheck

import (
	"go/token"

	"golang.org/x/tools/go/ssa"
)

// State is what is known about a value being nil.
type State int

const (
	// Unknown means no check on the way says anything. It is the zero value,
	// so a caller that switches on Nil and NonNil falls through to it.
	Unknown State = iota
	// Nil means the value was checked to be nil.
	Nil
	// NonNil means the value was checked not to be nil.
	NonNil
)

// At is what the checks on every path to the start of b say about v.
func At(v ssa.Value, b *ssa.BasicBlock) State {
	// A block with one predecessor is entered by one edge, so a block that
	// dominates b and has one predecessor puts that edge on every path to b.
	for s := b; s != nil; s = s.Idom() {
		if len(s.Preds) == 1 {
			if st := branch(v, s.Preds[0], s); st != Unknown {
				return st
			}
		}
	}
	return Unknown
}

// OnEdge is what is known about v on the edge from `from` to `to`: what the
// branch ending `from` says, or else what holds at the start of `from`.
func OnEdge(v ssa.Value, from, to *ssa.BasicBlock) State {
	if st := branch(v, from, to); st != Unknown {
		return st
	}
	return At(v, from)
}

// branch is what the if ending `from` says about v when it leaves to `to`.
func branch(v ssa.Value, from, to *ssa.BasicBlock) State {
	br, ok := from.Instrs[len(from.Instrs)-1].(*ssa.If)
	// An if whose edges both lead to one block says nothing about which was taken.
	if !ok || from.Succs[0] == from.Succs[1] {
		return Unknown
	}
	cmp, ok := br.Cond.(*ssa.BinOp)
	if !ok || (cmp.Op != token.EQL && cmp.Op != token.NEQ) {
		return Unknown
	}
	if (cmp.X != v || !isNil(cmp.Y)) && (cmp.Y != v || !isNil(cmp.X)) {
		return Unknown
	}
	if (cmp.Op == token.EQL) == (to == from.Succs[0]) {
		return Nil
	}
	return NonNil
}

// isNil reports whether v is the zero constant: nil, or the zero value of a
// struct.
func isNil(v ssa.Value) bool {
	c, ok := v.(*ssa.Const)
	return ok && c.Value == nil
}
