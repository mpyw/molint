// Command nilproof reports a returned pointer that is not proven non-nil.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/mpyw/nilproof"
)

func main() {
	singlechecker.Main(nilproof.Analyzer)
}
