package glazedmigration

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// applySlugRules handles R5: rename settings.GlazedSlug to
// settings.StructuredOutputSlug. This is a deterministic identifier rename on
// a package-qualified selector (or dot-imported ident).
//
// GlazedSlug is a constant, not a function, so it appears as a SelectorExpr
// (qualified/aliased) or a bare Ident (dot-imported) in expression position,
// not as a CallExpr. We inspect all identifiers and selectors.
func applySlugRules(pass *analysis.Pass, file *ast.File, imports importNames) {
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.SelectorExpr:
			ident, ok := selectorMatches(pass, n, imports, settingsImportPath, oldSlug)
			if !ok {
				return true
			}
			reportSlugRename(pass, ident)
		case *ast.Ident:
			// Dot-imported bare ident.
			if !imports.dot || n.Name != oldSlug {
				return true
			}
			// Confirm it resolves to the settings package when type info is
			// available; otherwise (no type info) accept the dot-import match.
			if obj := pass.TypesInfo.Uses[n]; obj != nil {
				// GlazedSlug is a const, so it's a *types.Const, not *types.Func.
				// Accept any object whose package is the settings package.
				pkg := obj.Pkg()
				if pkg != nil && pkg.Path() != settingsImportPath {
					return true
				}
			}
			reportSlugRename(pass, n)
		}
		return true
	})
}

func reportSlugRename(pass *analysis.Pass, ident *ast.Ident) {
	pass.Report(analysis.Diagnostic{
		Pos:     ident.Pos(),
		End:     ident.End(),
		Message: "replace settings." + oldSlug + " with settings." + newSlug,
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "use " + newSlug,
			TextEdits: []analysis.TextEdit{{
				Pos:     ident.Pos(),
				End:     ident.End(),
				NewText: []byte(newSlug),
			}},
		}},
	})
}
