// Command molint enforces the use of github.com/samber/mo.
//
// Usage:
//
//	molint [flags] [packages]   analyze
//	molint skill install        install the authoring skill for an AI agent
//	molint -V=full              print the release, as go vet -vettool asks
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/mpyw/molint"
)

// init runs both before the driver. The driver reads every non-flag argument
// as a package pattern, and it registers its own -V only when none is.
//
// They run here rather than in main so that the test binary runs them too.
// Intercept reads only the first argument, and a test binary's first
// argument is a -test flag, so it hands over at once.
func init() {
	skills.Intercept()
	registerVersionFlag()
}

func main() {
	singlechecker.Main(molint.Analyzer)
}
