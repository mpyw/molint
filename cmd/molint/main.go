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

func main() {
	// Both run before the driver. It reads every non-flag argument as a
	// package pattern, and it registers its own -V only when none is.
	skills.Intercept()
	registerVersionFlag()
	singlechecker.Main(molint.Analyzer)
}
