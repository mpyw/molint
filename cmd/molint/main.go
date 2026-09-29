// Command molint reports a returned pointer that is not proven non-nil.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/mpyw/molint"
)

func main() {
	singlechecker.Main(molint.Analyzer)
}
