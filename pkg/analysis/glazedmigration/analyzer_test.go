package glazedmigration

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestAnalyzerSuggestedFix(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		wantDiagnostics int
		wantFixes       int
	}{
		{
			name:            "default import",
			source:          `package p; import "github.com/go-go-golems/glazed/pkg/settings"; func f() { settings.NewGlazedSchema() }`,
			wantDiagnostics: 1,
			wantFixes:       1,
		},
		{
			name:            "alias import",
			source:          `package p; import output "github.com/go-go-golems/glazed/pkg/settings"; func f() { output.NewGlazedSchema() }`,
			wantDiagnostics: 1,
			wantFixes:       1,
		},
		{
			name:            "dot import",
			source:          `package p; import . "github.com/go-go-golems/glazed/pkg/settings"; func f() { NewGlazedSchema() }`,
			wantDiagnostics: 1,
			wantFixes:       1,
		},
		{
			name:            "legacy options require manual migration",
			source:          `package p; import "github.com/go-go-golems/glazed/pkg/settings"; func f() { settings.NewGlazedSchema(option) }`,
			wantDiagnostics: 1,
			wantFixes:       0,
		},
		{
			name:            "other package",
			source:          `package p; import other "example.com/settings"; func f() { other.NewGlazedSchema() }`,
			wantDiagnostics: 0,
			wantFixes:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "input.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			var diagnostics []analysis.Diagnostic
			pass := &analysis.Pass{
				Analyzer: Analyzer,
				Fset:     fset,
				Files:    []*ast.File{file},
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{},
				},
				Report: func(d analysis.Diagnostic) { diagnostics = append(diagnostics, d) },
			}
			if _, err := run(pass); err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != tt.wantDiagnostics {
				t.Fatalf("got %d diagnostics, want %d", len(diagnostics), tt.wantDiagnostics)
			}
			fixes := 0
			for _, diagnostic := range diagnostics {
				fixes += len(diagnostic.SuggestedFixes)
				for _, fix := range diagnostic.SuggestedFixes {
					if got := string(fix.TextEdits[0].NewText); got != newConstructor {
						t.Fatalf("replacement is %q, want %q", got, newConstructor)
					}
				}
			}
			if fixes != tt.wantFixes {
				t.Fatalf("got %d fixes, want %d", fixes, tt.wantFixes)
			}
		})
	}
}
