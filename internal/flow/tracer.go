package flow

import (
	"go/constant"
	"go/token"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/nilcheck"
)

// Tracer tells whether values are one of the constants a rule looks for. It
// remembers what it has worked out about a function, so one tracer serves
// one function.
type Tracer struct {
	// bad reports whether a constant is one the rule looks for.
	bad func(*ssa.Const) bool
	// checks reports whether nil checks on the way count: a value checked
	// to be nil is bad, and one checked not to be nil is not.
	checks bool

	// visiting holds the φs being judged. A cycle back to one is taken as
	// not bad: going round a loop brings only what the other edges bring.
	visiting map[*ssa.Phi]bool
	// assumed counts the cycles taken as not bad so far.
	assumed int
	// decided remembers each φ judged. A bad φ is bad whatever was assumed.
	// One found not bad is remembered only when no cycle was assumed on the
	// way, since the assumption may not hold from another start.
	decided map[*ssa.Phi]bool
	follow  map[*ssa.Alloc]bool
}

// NilTracer looks for nil, and counts nil checks.
func NilTracer() *Tracer {
	return CheckingTracer(func(c *ssa.Const) bool { return c.IsNil() })
}

// CheckingTracer looks for the constants bad accepts, and counts checks
// against the zero constant on the way.
func CheckingTracer(bad func(*ssa.Const) bool) *Tracer {
	return &Tracer{bad: bad, checks: true}
}

// TrueTracer looks for the constant true.
func TrueTracer() *Tracer {
	return &Tracer{bad: func(c *ssa.Const) bool {
		return c.Value != nil && c.Value.Kind() == constant.Bool && constant.BoolVal(c.Value)
	}}
}

// checkedNonNil reports whether a nil check at one of sites says v is not
// nil. It is false when the tracer does not count checks.
//
//declscope:package // pair.go settles a value before taking its φ apart
func (t *Tracer) checkedNonNil(v ssa.Value, sites ...Site) bool {
	if !t.checks {
		return false
	}
	for _, s := range sites {
		if s.state(v) == nilcheck.NonNil {
			return true
		}
	}
	return false
}

// Is reports whether v may be a bad constant on some path to one of sites.
// Several sites stand for one path seen at several points: a check at any of
// them counts.
func (t *Tracer) Is(v ssa.Value, sites ...Site) bool {
	if t.checks {
		nilSeen := false
		for _, s := range sites {
			switch s.state(v) {
			case nilcheck.NonNil:
				return false
			case nilcheck.Nil:
				nilSeen = true
			}
		}
		if nilSeen {
			return true
		}
	}
	switch v := v.(type) {
	case *ssa.Const:
		return t.bad(v)
	case *ssa.ChangeType:
		return t.Is(v.X, sites...)
	case *ssa.Phi:
		return t.phi(v)
	case *ssa.UnOp:
		if v.Op != token.MUL {
			return false
		}
		switch x := v.X.(type) {
		case *ssa.Alloc:
			if t.followed(x) {
				return t.load(x, v)
			}
		case *ssa.FreeVar:
			// The body of a range-over-func loop runs while the loop's
			// call runs, so a variable of the enclosing function holds
			// what reached that call.
			if a, mc := rangeBinding(x); a != nil && t.followed(a) {
				return t.load(a, mc)
			}
		}
	}
	return false
}

// phi judges each edge of a φ where it comes from. The answer does not
// depend on where the φ is used, so it is remembered for the function.
func (t *Tracer) phi(phi *ssa.Phi) bool {
	if t.visiting == nil {
		t.visiting = make(map[*ssa.Phi]bool)
		t.decided = make(map[*ssa.Phi]bool)
	}
	if bad, ok := t.decided[phi]; ok {
		return bad
	}
	if t.visiting[phi] {
		t.assumed++
		return false
	}
	t.visiting[phi] = true
	before := t.assumed
	bad := false
	b := phi.Block()
	for i, e := range phi.Edges {
		if t.Is(e, siteEdge(b.Preds[i], b)) {
			bad = true
			break
		}
	}
	delete(t.visiting, phi)
	if bad || t.assumed == before {
		t.decided[phi] = bad
	}
	return bad
}

// load reports whether a store that reaches the load brings a bad constant.
func (t *Tracer) load(a *ssa.Alloc, load ssa.Instruction) bool {
	for _, d := range storesReaching(a, load) {
		if t.Is(d.value, SiteAt(d.block)) {
			return true
		}
	}
	return false
}

// followed reports whether the loads of a are followed: nothing but its own
// loads and stores uses its address, apart from function literals that only
// load it.
//
//declscope:package // pair.go follows two loads along one path
func (t *Tracer) followed(a *ssa.Alloc) bool {
	if t.follow == nil {
		t.follow = make(map[*ssa.Alloc]bool)
	}
	f, ok := t.follow[a]
	if !ok {
		f = tracerFollows(a, a.Referrers(), false)
		t.follow[a] = f
	}
	return f
}

// tracerFollows reports whether every referrer of addr loads it, binds it
// into a function literal that only loads it, or, unless onlyLoads, stores
// into it.
//
// An address has no unary operation but a load, so any UnOp loads it.
func tracerFollows(addr ssa.Value, refs *[]ssa.Instruction, onlyLoads bool) bool {
	for _, r := range *refs {
		switch r := r.(type) {
		case *ssa.UnOp:
		case *ssa.Store:
			if onlyLoads || r.Val == addr {
				return false
			}
		case *ssa.MakeClosure:
			fn := r.Fn.(*ssa.Function)
			for i, b := range r.Bindings {
				if b == addr && !tracerFollows(fn.FreeVars[i], fn.FreeVars[i].Referrers(), true) {
					return false
				}
			}
		default:
			return false
		}
	}
	return true
}
