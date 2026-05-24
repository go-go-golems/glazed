package main

import (
	"github.com/go-go-golems/glazed/pkg/analysis/glazedclilint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(glazedclilint.Analyzer)
}
