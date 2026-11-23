package flow

import (
	"go/token"
	"slices"

	"golang.org/x/tools/go/ssa"
)

// Pair reports whether some path to at makes p bad for pt and g bad for gt
// at once. It is how a nil pointer is held to the error or the bool returned
// beside it: only a path that gives both is reported.
//
// Where either value is a φ, each incoming edge is judged on its own, with
// the other value read along the same edge. That repeats into earlier blocks
// while φs remain. A variable that SSA does not lift is followed through the
// stores that reach it, both variables along one path.
func Pair(pt, gt *Tracer, p, g ssa.Value, at Site) bool {
	return pair(pt, gt, pairAlt{v: p, sites: []Site{at}}, pairAlt{v: g, sites: []Site{at}}, make(map[pairKey]bool))
}

// pairKey is a pair already judged, with what the checks on the way said.
type pairKey struct {
	p, g       ssa.Value
	pNil, gNil bool
}

func pair(pt, gt *Tracer, p, g pairAlt, seen map[pairKey]bool) bool {
	// A check on the way settles a value for the whole path, before its φ
	// is taken apart and the check is left behind.
	var pNonNil, gNonNil bool
	p, pNonNil = pairSettle(pt, p)
	g, gNonNil = pairSettle(gt, g)
	if pNonNil || gNonNil {
		return false
	}
	var blk *ssa.BasicBlock
	if phi, ok := g.v.(*ssa.Phi); ok {
		blk = phi.Block()
	} else if phi, ok := p.v.(*ssa.Phi); ok {
		blk = phi.Block()
	}
	if blk != nil {
		key := pairKey{p.v, g.v, p.checkedNil, g.checkedNil}
		if seen[key] {
			return false
		}
		seen[key] = true
		for k, pred := range blk.Preds {
			e := siteEdge(pred, blk)
			for _, pa := range pairStep(pt, p, blk, k, e) {
				for _, ga := range pairStep(gt, g, blk, k, e) {
					if pair(pt, gt, pa, ga, seen) {
						return true
					}
				}
			}
		}
		return false
	}
	if pa, pl := pairFollowedLoad(pt, p.v); pa != nil {
		if ga, gl := pairFollowedLoad(gt, g.v); ga != nil {
			return pairStores(pt, gt, pa, pl, p.checkedNil, ga, gl, g.checkedNil, seen)
		}
	}
	return pairBad(pt, p) && pairBad(gt, g)
}

// pairAlt is one value a pair may hold along an edge, with where it is
// judged. checkedNil records that a φ it was read from was checked to be nil
// on the way.
type pairAlt struct {
	v          ssa.Value
	sites      []Site
	checkedNil bool
}

// pairSettle reads the checks at a's sites. A value checked to be nil that is
// not a φ stands for the constant from here on. A φ checked to be nil is
// still taken apart, so that an edge whose value cannot be nil stays non-nil:
// the check cannot succeed along it.
func pairSettle(t *Tracer, a pairAlt) (pairAlt, bool) {
	isNil, nonNil := t.checked(a.v, a.sites...)
	if isNil {
		if _, ok := a.v.(*ssa.Phi); ok {
			a.checkedNil = true
		} else {
			a.v = ssa.NewConst(nil, a.v.Type())
		}
	}
	return a, nonNil
}

// pairBad reports whether a is bad for t, where neither value of the pair is
// a φ any more. A value read from a φ checked to be nil is nil when it can be.
func pairBad(t *Tracer, a pairAlt) bool {
	return t.Is(a.v, a.sites...) || a.checkedNil && pairMayBeNil(a.v)
}

// pairMayBeNil reports whether v may be nil. An allocation, as in &T{}, and a
// conversion of a non-interface value to an interface never are.
func pairMayBeNil(v ssa.Value) bool {
	switch v.(type) {
	case *ssa.Alloc, *ssa.MakeInterface:
		return false
	}
	return true
}

// pairStep reads a along edge k into blk. A φ of blk gives its edge. A load
// in blk of a followed variable, with no store before it in blk, gives what
// reaches the end of that predecessor, one alternative per store. Any other
// value is the same on the edge: when it is defined before blk, the checks
// on the edge apply to it too.
func pairStep(t *Tracer, a pairAlt, blk *ssa.BasicBlock, k int, e Site) []pairAlt {
	if phi, ok := a.v.(*ssa.Phi); ok && phi.Block() == blk {
		return []pairAlt{{phi.Edges[k], []Site{e}, a.checkedNil}}
	}
	if al, load := pairFollowedLoad(t, a.v); al != nil && load.Block() == blk && pairLoadsEntry(al, load) {
		var alts []pairAlt
		for _, d := range storesBefore(al, blk.Preds[k], len(blk.Preds[k].Instrs)) {
			alts = append(alts, pairAlt{d.value, []Site{SiteAt(d.block)}, a.checkedNil})
		}
		return alts
	}
	if pairDefinedBefore(a.v, blk) {
		a.sites = append(append([]Site(nil), a.sites...), e)
	}
	return []pairAlt{a}
}

// pairLoadsEntry reports whether load reads what a holds on entry to its
// block: nothing in the block stores into a before it.
func pairLoadsEntry(a *ssa.Alloc, load *ssa.UnOp) bool {
	b := load.Block()
	_, stored := storeLast(a, b, slices.Index(b.Instrs, ssa.Instruction(load)))
	return !stored
}

func pairDefinedBefore(v ssa.Value, blk *ssa.BasicBlock) bool {
	in, ok := v.(ssa.Instruction)
	if !ok {
		return true
	}
	b := in.Block()
	return b != blk && b.Dominates(blk)
}

// followedLoad returns the variable v loads, when the tracer follows it.
func pairFollowedLoad(t *Tracer, v ssa.Value) (*ssa.Alloc, *ssa.UnOp) {
	load, ok := v.(*ssa.UnOp)
	if !ok || load.Op != token.MUL {
		return nil, nil
	}
	a, ok := load.X.(*ssa.Alloc)
	if !ok || !t.followed(a) {
		return nil, nil
	}
	return a, load
}

// pairStores walks back from the loads along each path, to the last store
// into each variable on that path, and pairs the two stored values, which
// may be φs themselves.
func pairStores(pt, gt *Tracer, pa *ssa.Alloc, pl *ssa.UnOp, pNil bool, ga *ssa.Alloc, gl *ssa.UnOp, gNil bool, pairSeen map[pairKey]bool) bool {
	type state struct {
		b      *ssa.BasicBlock
		pd, gd *storeDef
	}
	seen := make(map[state]bool)
	var walk func(b *ssa.BasicBlock, end int, pd, gd *storeDef) bool
	walk = func(b *ssa.BasicBlock, end int, pd, gd *storeDef) bool {
		for i := end - 1; i >= 0 && (pd == nil || gd == nil); i-- {
			switch in := b.Instrs[i].(type) {
			case *ssa.Store:
				if pd == nil && in.Addr == pa {
					pd = &storeDef{value: in.Val, block: b}
				}
				if gd == nil && in.Addr == ga {
					gd = &storeDef{value: in.Val, block: b}
				}
			case *ssa.Alloc:
				if pd == nil && in == pa {
					pd = &storeDef{value: storeZero(pa), block: b}
				}
				if gd == nil && in == ga {
					gd = &storeDef{value: storeZero(ga), block: b}
				}
			}
		}
		if pd != nil && gd != nil {
			return pair(pt, gt, pairAlt{pd.value, []Site{SiteAt(pd.block)}, pNil}, pairAlt{gd.value, []Site{SiteAt(gd.block)}, gNil}, pairSeen)
		}
		for _, p := range b.Preds {
			s := state{p, pd, gd}
			if seen[s] {
				continue
			}
			seen[s] = true
			if walk(p, len(p.Instrs), pd, gd) {
				return true
			}
		}
		return false
	}
	// Both loads sit in the return's block, which is where the walk starts.
	b := pl.Block()
	end := len(b.Instrs)
	for i, in := range b.Instrs {
		if in == pl || in == gl {
			end = i
			break
		}
	}
	return walk(b, end, nil, nil)
}
