package internal

import (
	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/nilcheck"
	"github.com/mpyw/molint/internal/typeutil"
)

// judgement is what judging one function found.
//
//declscope:package
type judgement struct {
	// proven is the bit set of the pointer results proven on every return.
	proven uint64
	// failures lists each return and pointer result not proven, once.
	failures []judgedFailure
}

// judgedFailure is a pointer result that one return does not prove.
//
//declscope:package
type judgedFailure struct {
	ret    *ssa.Return
	result int
	// nilError means the return's error result may be nil there.
	nilError bool
	failure  *proofFailure
}

// judgedPair is one way a return is reached, with the values it returns
// that way.
type judgedPair struct {
	site    proofSite
	results []ssa.Value
}

// judge proves the pointer results of fn on every return.
//
// A function without an error result must return non-nil pointers on every
// return. A function whose last result is an error must do so wherever that
// error may be nil: the constant nil, a value checked to be nil, or a φ that
// may bring either. Any other error is taken as non-nil, since the rule is
// there to catch return nil, nil and not to prove errors. One exception
// holds: a pointer and an error returned together from one call stand for
// what the callee promised, so the callee has to be proven.
//
//declscope:package
func (c *checker) judge(fn *ssa.Function) judgement {
	ptrs := typeutil.PointerResults(fn.Signature)
	errIdx := typeutil.ErrorResult(fn.Signature)
	var j judgement
	if len(ptrs) == 0 {
		return j
	}
	for _, i := range ptrs {
		j.proven |= 1 << i
	}
	p := c.newProof()
	for _, b := range fn.Blocks {
		ret, ok := b.Instrs[len(b.Instrs)-1].(*ssa.Return)
		if !ok {
			continue
		}
		failed := make(map[int]bool)
		for _, pair := range judgedPairs(ret, errIdx) {
			for _, i := range ptrs {
				if failed[i] {
					continue
				}
				f := judgedFailure{ret: ret, result: i}
				switch {
				case errIdx < 0:
					f.failure = p.nonNil(pair.results[i], pair.site)
				case judgedMayBeNil(pair.results[errIdx], pair.site):
					f.nilError = true
					f.failure = p.nonNil(pair.results[i], pair.site)
				default:
					f.failure = p.together(pair.results[i], pair.results[errIdx])
				}
				if f.failure != nil {
					failed[i] = true
					j.proven &^= 1 << i
					j.failures = append(j.failures, f)
				}
			}
		}
	}
	return j
}

// judgedPairs lists the ways ret is reached. A return whose error is a φ of
// its own block is judged along each edge, with every other φ of the block
// read along the same edge, so that a pointer is held only to the error it
// is returned with.
func judgedPairs(ret *ssa.Return, errIdx int) []judgedPair {
	b := ret.Block()
	if errIdx >= 0 {
		if phi, ok := ret.Results[errIdx].(*ssa.Phi); ok && phi.Block() == b {
			pairs := make([]judgedPair, len(b.Preds))
			for k, pred := range b.Preds {
				results := make([]ssa.Value, len(ret.Results))
				for i, r := range ret.Results {
					if rp, ok := r.(*ssa.Phi); ok && rp.Block() == b {
						r = rp.Edges[k]
					}
					results[i] = r
				}
				pairs[k] = judgedPair{site: proofSite{block: b, from: pred}, results: results}
			}
			return pairs
		}
	}
	return []judgedPair{{site: proofSite{block: b}, results: ret.Results}}
}

// judgedMayBeNil reports whether an error may be nil at site: the constant
// nil, a value checked to be nil, or a φ with either among its edges.
func judgedMayBeNil(e ssa.Value, at proofSite) bool {
	switch at.state(e) {
	case nilcheck.Nil:
		return true
	case nilcheck.NonNil:
		return false
	}
	return judgedHoldsNil(e, make(map[*ssa.Phi]bool))
}

// judgedHoldsNil reports whether v is the constant nil, or a φ that may bring it.
func judgedHoldsNil(v ssa.Value, seen map[*ssa.Phi]bool) bool {
	switch v := v.(type) {
	case *ssa.Const:
		return v.IsNil()
	case *ssa.Phi:
		if seen[v] {
			return false
		}
		seen[v] = true
		for _, e := range v.Edges {
			if judgedHoldsNil(e, seen) {
				return true
			}
		}
	}
	return false
}
