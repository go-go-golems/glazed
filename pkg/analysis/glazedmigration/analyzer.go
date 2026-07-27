// Package glazedmigration provides source migrations for removed Glazed APIs.
package glazedmigration

import (
	"go/ast"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

const (
	settingsImportPath = "github.com/go-go-golems/glazed/pkg/settings"
	oldConstructor     = "NewGlazedSchema"
	newConstructor     = "NewStructuredOutputSection"
)

// Analyzer rewrites calls to settings.NewGlazedSchema() to
// settings.NewStructuredOutputSection(). It runs despite type errors because the old
// constructor is absent from current Glazed versions.
var Analyzer = &analysis.Analyzer{
	Name:             "glazedmigration",
	Doc:              "migrate removed Glazed APIs to their structured-output replacements",
	RunDespiteErrors: true,
	Run:              run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		imports := settingsImports(file)
		if len(imports.qualified) == 0 && !imports.dot {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := legacyConstructorName(pass, call.Fun, imports)
			if !ok {
				return true
			}

			diagnostic := analysis.Diagnostic{
				Pos:     name.Pos(),
				End:     name.End(),
				Message: "replace settings.NewGlazedSchema with settings.NewStructuredOutputSection",
			}
			if len(call.Args) == 0 {
				diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
					Message: "use NewStructuredOutputSection",
					TextEdits: []analysis.TextEdit{{
						Pos:     name.Pos(),
						End:     name.End(),
						NewText: []byte(newConstructor),
					}},
				}}
			} else {
				diagnostic.Message += "; legacy GlazeSectionOption arguments require manual migration"
			}
			pass.Report(diagnostic)
			return true
		})
	}
	return nil, nil
}

type importNames struct {
	qualified map[string]bool
	dot       bool
}

func settingsImports(file *ast.File) importNames {
	ret := importNames{qualified: map[string]bool{}}
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != settingsImportPath {
			continue
		}
		if spec.Name != nil {
			switch spec.Name.Name {
			case "_":
				continue
			case ".":
				ret.dot = true
			default:
				ret.qualified[spec.Name.Name] = true
			}
			continue
		}
		ret.qualified["settings"] = true
	}
	return ret
}

func legacyConstructorName(pass *analysis.Pass, expr ast.Expr, imports importNames) (*ast.Ident, bool) {
	switch fn := expr.(type) {
	case *ast.SelectorExpr:
		qualifier, ok := fn.X.(*ast.Ident)
		if !ok || fn.Sel.Name != oldConstructor || !imports.qualified[qualifier.Name] {
			return nil, false
		}
		// A local declaration can shadow the import in a nested scope. Check the
		// qualifier even when the removed selector itself could not be resolved.
		if obj := pass.TypesInfo.Uses[qualifier]; obj != nil {
			pkgName, ok := obj.(*types.PkgName)
			if !ok || pkgName.Imported() == nil || pkgName.Imported().Path() != settingsImportPath {
				return nil, false
			}
		}
		// When type information is available, reject selectors resolved to another package.
		if obj := pass.TypesInfo.Uses[fn.Sel]; obj != nil {
			if typed, ok := obj.(*types.Func); !ok || typed.Pkg() == nil || typed.Pkg().Path() != settingsImportPath {
				return nil, false
			}
		}
		return fn.Sel, true
	case *ast.Ident:
		if !imports.dot || fn.Name != oldConstructor {
			return nil, false
		}
		if obj := pass.TypesInfo.Uses[fn]; obj != nil {
			typed, ok := obj.(*types.Func)
			if !ok || typed.Pkg() == nil || typed.Pkg().Path() != settingsImportPath {
				return nil, false
			}
		}
		return fn, true
	default:
		return nil, false
	}
}
