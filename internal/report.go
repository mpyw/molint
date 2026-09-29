package internal

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/nilproof/internal/typeutil"
)

// reportFunc reports each return of fn that does not prove a pointer result.
//
//declscope:package
func (c *checker) reportFunc(fn *ssa.Function) {
	for _, f := range c.judge(fn).failures {
		pos := f.ret.Pos()
		if !pos.IsValid() {
			pos = fn.Pos()
		}
		if c.generated[c.pass.Fset.Position(pos).Filename] || c.directives.Ignored(pos) {
			continue
		}
		d := analysis.Diagnostic{Pos: pos, Message: c.reportMessage(fn, f)}
		if at := f.failure.value.Pos(); at.IsValid() && c.pass.Fset.Position(at).Line != c.pass.Fset.Position(pos).Line {
			d.Related = []analysis.RelatedInformation{{Pos: at, Message: "the value that may be nil"}}
		}
		c.pass.Report(d)
	}
}

// reportMessage says which result may be nil, where the nil may come from,
// and what to return instead. It is written to be acted on without reading
// the analyzer: by a developer, or by a coding agent.
func (c *checker) reportMessage(fn *ssa.Function, f judgedFailure) string {
	sig := fn.Signature
	t := types.TypeString(sig.Results().At(f.result).Type(), typeutil.Qualifier(c.pass.Pkg))
	var b strings.Builder
	fmt.Fprintf(&b, "%s may return a nil %s", reportName(fn, c.pass.Pkg), t)
	if len(typeutil.PointerResults(sig)) > 1 {
		fmt.Fprintf(&b, " as result %d", f.result)
	}
	if f.nilError {
		b.WriteString(" with a nil error")
	}
	fmt.Fprintf(&b, ", from %s; ", f.failure.origin(c.pass.Pkg))
	if f.nilError {
		fmt.Fprintf(&b, "return an error, or mo.Option[%s], where absence is expected", t)
	} else {
		fmt.Fprintf(&b, "return mo.Option[%s] where absence is expected", t)
	}
	return b.String()
}

// reportName names fn as a diagnostic spells it.
func reportName(fn *ssa.Function, pkg *types.Package) string {
	if parent := fn.Parent(); parent != nil {
		return "the function literal in " + reportName(parent, pkg)
	}
	return typeutil.FuncName(fn, pkg)
}
