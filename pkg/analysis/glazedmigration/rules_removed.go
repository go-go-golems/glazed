package glazedmigration

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// applyRemovedFeatureRules handles R9: report-only diagnostics for the removed
// feature sections (select, rename, replace, template, jq, sort, skip-limit,
// fields-filters). These implemented runtime middlewares that were
// intentionally removed and have no mechanical migration; consuming code must
// be redesigned using application fields or caller-side tooling.
func applyRemovedFeatureRules(pass *analysis.Pass, file *ast.File, imports importNames) {
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := selectorMatches(pass, call.Fun, imports, settingsImportPath, removedFeatureNames()...)
		if !ok {
			return true
		}
		pass.Report(analysis.Diagnostic{
			Pos:     ident.Pos(),
			End:     ident.End(),
			Message: "settings." + ident.Name + " implements a removed feature section with no mechanical migration; redesign using application fields or caller-side tools (e.g. jq)",
		})
		return true
	})
}

func removedFeatureNames() []string {
	out := make([]string, 0, len(removedFeatureSectionConstructors))
	for n := range removedFeatureSectionConstructors {
		out = append(out, n)
	}
	return out
}
