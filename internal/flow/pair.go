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
	return pair(pt, gt, p, []Site{at}, g, []Site{at}, make(map[[2]ssa.Value]bool))
}

func pair(pt, gt *Tracer, p ssa.Value, ps []Site, g ssa.Value, gs []Site, seen map[[2]ssa.Value]bool) bool {
	// A check on the way settles a value for the whole path, before its φ
	// is taken apart and the check is left behind. A value checked to be
	// nil stands for the constant from here on.
	pNil, pNonNil := pt.checked(p, ps...)
	gNil, gNonNil := gt.checked(g, gs...)
	if pNonNil || gNonNil {
		return false
	}
	if pNil {
		p = ssa.NewConst(nil, p.Type())
	}
	if gNil {
		g = ssa.NewConst(nil, g.Type())
	}
	var blk *ssa.BasicBlock
	if phi, ok := g.(*ssa.Phi); ok {
		blk = phi.Block()
	} else if phi, ok := p.(*ssa.Phi); ok {
		blk = phi.Block()
	}
	if blk != nil {
		key := [2]ssa.Value{p, g}
		if seen[key] {
			return false
		}
		seen[key] = true
		for k, pred := range blk.Preds {
			e := siteEdge(pred, blk)
			for _, pa := range pairStep(pt, p, ps, blk, k, e) {
				for _, ga := range pairStep(gt, g, gs, blk, k, e) {
					if pair(pt, gt, pa.v, pa.sites, ga.v, ga.sites, seen) {
						return true
					}
				}
			}
		}
		return false
	}
	if pa, pl := pairFollowedLoad(pt, p); pa != nil {
		if ga, gl := pairFollowedLoad(gt, g); ga != nil {
			return pairStores(pt, gt, pa, pl, ga, gl, seen)
		}
	}
	return pt.Is(p, ps...) && gt.Is(g, gs...)
}

// pairAlt is one value a pair may hold along an edge, with where it is
// judged.
type pairAlt struct {
	v     ssa.Value
	sites []Site
}

// pairStep reads v along edge k into blk. A φ of blk gives its edge. A load
// in blk of a followed variable, with no store before it in blk, gives what
// reaches the end of that predecessor, one alternative per store. Any other
// value is the same on the edge: when it is defined before blk, the checks
// on the edge apply to it too.
func pairStep(t *Tracer, v ssa.Value, sites []Site, blk *ssa.BasicBlock, k int, e Site) []pairAlt {
	if phi, ok := v.(*ssa.Phi); ok && phi.Block() == blk {
		return []pairAlt{{phi.Edges[k], []Site{e}}}
	}
	if a, load := pairFollowedLoad(t, v); a != nil && load.Block() == blk && pairLoadsEntry(a, load) {
		var alts []pairAlt
		for _, d := range storesBefore(a, blk.Preds[k], len(blk.Preds[k].Instrs)) {
			alts = append(alts, pairAlt{d.value, []Site{SiteAt(d.block)}})
		}
		return alts
	}
	if pairDefinedBefore(v, blk) {
		return []pairAlt{{v, append(append([]Site(nil), sites...), e)}}
	}
	return []pairAlt{{v, sites}}
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
func pairStores(pt, gt *Tracer, pa *ssa.Alloc, pl *ssa.UnOp, ga *ssa.Alloc, gl *ssa.UnOp, pairSeen map[[2]ssa.Value]bool) bool {
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
			return pair(pt, gt, pd.value, []Site{SiteAt(pd.block)}, gd.value, []Site{SiteAt(gd.block)}, pairSeen)
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
