// Command molint enforces the use of github.com/samber/mo.
//
// Usage:
//
//	molint [flags] [packages]   analyze
//	molint skill install        install the authoring skill for an AI agent
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/mpyw/molint"
)

func main() {
	// Before the driver, which reads every non-flag argument as a package
	// pattern.
	skills.Intercept()
	singlechecker.Main(molint.Analyzer)
}
