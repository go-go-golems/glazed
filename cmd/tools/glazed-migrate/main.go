// glazed-migrate applies source migrations for removed Glazed APIs.
//
// Run it with -fix to edit source files in place:
//
//	go run github.com/go-go-golems/glazed/cmd/tools/glazed-migrate@latest -fix ./...
package main

import (
	"github.com/go-go-golems/glazed/pkg/analysis/glazedmigration"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(glazedmigration.Analyzer)
}
