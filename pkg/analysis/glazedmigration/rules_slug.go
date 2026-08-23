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
			// Dot-imported bare ident. Require it to be a *use* of the settings
			// package symbol, not a local declaration that merely shares the name.
			// Without this guard, a locally declared `GlazedSlug` (which has no
			// Uses entry) would be wrongly renamed.
			if !imports.dot || n.Name != oldSlug {
				return true
			}
			// Skip declarations (defs) entirely.
			if _, isDef := pass.TypesInfo.Defs[n]; isDef {
				return true
			}
			obj := pass.TypesInfo.Uses[n]
			if obj == nil {
				// The standalone migration driver intentionally runs without
				// type-checking because removed APIs make target packages fail to
				// compile. go/parser still resolves same-file declarations through
				// Ident.Obj: a nil Obj on a bare identifier under a settings dot
				// import is therefore the imported symbol, while local declarations
				// and their uses remain protected from rewriting.
				if n.Obj == nil {
					reportSlugRename(pass, n)
				}
				return true
			}
			pkg := obj.Pkg()
			if pkg == nil || pkg.Path() != settingsImportPath {
				return true
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
