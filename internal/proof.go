package internal

import (
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/nilcheck"
	"github.com/mpyw/molint/internal/typeutil"
)

// proofSite is where a value is proven: the start of a block, or the edge
// into it that a φ reads the value from.
//
//declscope:package
type proofSite struct {
	// block is the block the value is used in.
	block *ssa.BasicBlock
	// from is the predecessor the value comes from, for an edge of a φ.
	from *ssa.BasicBlock
}

// errorChecked reports whether result e of call is checked to be nil on the
// way to the site.
func (s proofSite) errorChecked(call *ssa.Call, e int) bool {
	for _, r := range *call.Referrers() {
		if x, ok := r.(*ssa.Extract); ok && x.Index == e && s.state(x) == nilcheck.Nil {
			return true
		}
	}
	return false
}

// state is what the nil checks on the way to the site say about v.
//
//declscope:package // judge.go asks it about a returned error
func (s proofSite) state(v ssa.Value) nilcheck.State {
	if s.from != nil {
		return nilcheck.OnEdge(v, s.from, s.block)
	}
	return nilcheck.At(v, s.block)
}

// proofFailure is why a value is not proven non-nil: the value the proof
// stopped at, which is where the nil may come from.
//
//declscope:package
type proofFailure struct {
	// value is where the proof stopped.
	value ssa.Value
	// checkedNil means a nil check on the way says the value is nil.
	//
	//declscope:private
	checkedNil bool
	// unchecked means the value is the result of a call whose error result
	// is not checked for nil on the way.
	//
	//declscope:private
	unchecked bool
}

// origin describes where the nil may come from, for a diagnostic.
//
//declscope:package // report.go writes it into the message
func (f *proofFailure) origin(pkg *types.Package) string {
	if f.checkedNil {
		return "a value checked to be nil"
	}
	switch v := f.value.(type) {
	case *ssa.Const:
		return "a nil constant"
	case *ssa.Parameter:
		return "parameter " + v.Name()
	case *ssa.Call:
		return f.callOrigin(v.Common(), pkg)
	case *ssa.Extract:
		// A call's result fails as the call. The value of a comma-ok form is
		// described by the form.
		return (&proofFailure{value: v.Tuple}).origin(pkg)
	case *ssa.UnOp:
		switch x := v.X.(type) {
		// A variable a function literal captures lives on the heap, and is
		// read through its address.
		case *ssa.FreeVar:
			return "captured variable " + x.Name()
		case *ssa.Alloc:
			return "variable " + x.Comment
		case *ssa.Global:
			return "package variable " + x.Name()
		case *ssa.IndexAddr:
			return "an element"
		case *ssa.FieldAddr:
			return "field " + x.X.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Struct).Field(x.Field).Name()
		}
		if v.Op == token.ARROW {
			return "a channel receive"
		}
		return "a pointer load"
	case *ssa.Field:
		return "field " + v.X.Type().Underlying().(*types.Struct).Field(v.Field).Name()
	case *ssa.Lookup:
		return "a map lookup"
	case *ssa.Index:
		return "an element"
	case *ssa.TypeAssert:
		return "a type assertion"
	case *ssa.Convert, *ssa.SliceToArrayPointer:
		return "a conversion"
	}
	return "a value molint does not follow"
}

// callOrigin describes a call whose result is not proven.
func (f *proofFailure) callOrigin(cc *ssa.CallCommon, pkg *types.Package) string {
	what := "a call through an interface or a function value"
	if fn := cc.StaticCallee(); fn != nil {
		what = "a call to " + typeutil.FuncName(fn, pkg)
		if !f.unchecked {
			return what + ", which may return nil"
		}
	}
	if f.unchecked {
		return what + ", whose error is not checked for nil"
	}
	return what
}

// proof proves values non-nil. One lives for one judgement of a function,
// since what it remembers depends on the summaries of the moment.
type proof struct {
	c *checker
	// visiting holds the φs being proven. A cycle back to one is taken as
	// proven: going round a loop adds to a φ only what its other edges bring.
	visiting map[*ssa.Phi]bool
	// assumed counts the cycles taken as proven so far.
	assumed int
	// decided remembers each φ decided, with nil for proven. A failure is
	// final whatever was assumed, since assuming more only proves more. A
	// success is remembered only when nothing was assumed on the way.
	decided map[*ssa.Phi]*proofFailure
}

//declscope:package
func (c *checker) newProof() *proof {
	return &proof{c: c, visiting: make(map[*ssa.Phi]bool), decided: make(map[*ssa.Phi]*proofFailure)}
}

// nonNil proves v non-nil at site, or says where the proof stopped.
//
//declscope:package
func (p *proof) nonNil(v ssa.Value, at proofSite) *proofFailure {
	switch at.state(v) {
	case nilcheck.NonNil:
		return nil
	case nilcheck.Nil:
		return &proofFailure{value: v, checkedNil: true}
	}
	switch v := v.(type) {
	// An address is never nil. A field or an element of a nil pointer panics
	// before it has one. A variable a function literal captures is reached
	// through its address.
	case *ssa.Alloc, *ssa.FieldAddr, *ssa.IndexAddr, *ssa.Global, *ssa.FreeVar:
		return nil
	case *ssa.ChangeType:
		return p.nonNil(v.X, at)
	case *ssa.Phi:
		return p.phi(v)
	case *ssa.Call:
		return p.result(v, 0, at)
	case *ssa.Extract:
		if call, ok := v.Tuple.(*ssa.Call); ok {
			return p.result(call, v.Index, at)
		}
	case *ssa.UnOp:
		if g, ok := v.X.(*ssa.Global); ok && v.Op == token.MUL && p.c.provenGlobal(g) {
			return nil
		}
	}
	return &proofFailure{value: v}
}

// phi proves every edge of a φ, each where it comes from.
func (p *proof) phi(phi *ssa.Phi) *proofFailure {
	if f, ok := p.decided[phi]; ok {
		return f
	}
	if p.visiting[phi] {
		p.assumed++
		return nil
	}
	p.visiting[phi] = true
	before := p.assumed
	var failure *proofFailure
	for i, e := range phi.Edges {
		if failure = p.nonNil(e, proofSite{block: phi.Block(), from: phi.Block().Preds[i]}); failure != nil {
			break
		}
	}
	delete(p.visiting, phi)
	if failure != nil || p.assumed == before {
		p.decided[phi] = failure
	}
	return failure
}

// result proves result i of call. A callee whose last result is an error
// promises its other results only when that error is nil, so the error has to
// be checked on the way to the site.
func (p *proof) result(call *ssa.Call, i int, at proofSite) *proofFailure {
	if f := p.paired(call, i); f != nil {
		return f
	}
	cc := call.Common()
	if e := typeutil.ErrorResult(cc.Signature()); e >= 0 && !at.errorChecked(call, e) {
		return &proofFailure{value: call, unchecked: true}
	}
	return nil
}

// paired proves result i of call wherever the call's own error result is
// nil, which is what its callee promises.
//
// A call through an interface or a function value has no callee to read.
// When it returns an error, it is trusted to follow the Go convention: its
// other results are usable when the error is nil. Without an error result
// there is no convention to trust.
func (p *proof) paired(call *ssa.Call, i int) *proofFailure {
	cc := call.Common()
	fn := cc.StaticCallee()
	if fn == nil {
		if typeutil.ErrorResult(cc.Signature()) >= 0 {
			return nil
		}
		return &proofFailure{value: call}
	}
	if !p.c.provenInSummary(fn, i) {
		return &proofFailure{value: call}
	}
	return nil
}

// together holds a pointer returned beside an error taken as non-nil. Only
// a pair from one call is held, to what its callee promised.
//
//declscope:package
func (p *proof) together(ptr, err ssa.Value) *proofFailure {
	x, ok := ptr.(*ssa.Extract)
	if !ok {
		return nil
	}
	call, ok := x.Tuple.(*ssa.Call)
	if !ok {
		return nil
	}
	if e, ok := err.(*ssa.Extract); !ok || e.Tuple != call {
		return nil
	}
	return p.paired(call, x.Index)
}
