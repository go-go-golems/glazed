// glazed-migrate-analyzer applies source migrations for removed Glazed APIs
// as a go/analysis singlechecker (vettool).
//
// Run it with -fix to edit source files in place:
//
//	go run github.com/go-go-golems/glazed/cmd/tools/glazed-migrate-analyzer@latest -fix ./...
//
// For structured output, embedded help, and operation on code that no longer
// compiles, use the glazed-migrate command instead (glazed-migrate help
// glazed-migrate-guide).
package main

import (
	"github.com/go-go-golems/glazed/pkg/analysis/glazedmigration"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(glazedmigration.Analyzer)
}
