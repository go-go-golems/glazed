package main

import (
	"github.com/go-go-golems/glazed/pkg/analysis/glazedclilint"
	"golang.org/x/tools/go/analysis/multichecker"
)

// glazed-lint bundles Glazed's custom go/analysis analyzers into a single vettool.
//
// Build it with:
//
//	go build -o /tmp/glazed-lint ./cmd/tools/glazed-lint
//
// Then run it with:
//
//	go vet -vettool=/tmp/glazed-lint ./...
func main() {
	multichecker.Main(
		glazedclilint.Analyzer,
	)
}
