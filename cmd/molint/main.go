// Command molint enforces the use of github.com/samber/mo.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/mpyw/molint"
)

func main() {
	singlechecker.Main(molint.Analyzer)
}
