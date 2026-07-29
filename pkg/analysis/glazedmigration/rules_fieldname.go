package glazedmigration

import (
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

// applyFieldNameRules handles the extended R4: renaming the string literal
// "output" to "format" when it references the structured-output section's field
// in contexts other than schema.WithDefaults (which is handled by
// reportKeyRename).
//
// The contexts covered are:
//
//  1. GetField / GetParameter calls where the first arg is the GlazedSlug /
//     StructuredOutputSlug selector and the second arg is the string "output".
//  2. UpdateExistingValue calls on a glazed layer where the first arg is the
//     string "output".
//  3. Map index expressions expr["output"] and map-literal keys "output":
//     these are reported (not auto-fixed) because the analyzer cannot always
//     prove the map is the structured-output section's values. The diagnostic
//     guides the human to rename to "format".
//
// This rule is deliberately conservative: it only auto-fixes contexts where the
// connection to the structured-output section is evident (a slug selector or a
// known glazed-layer method). Custom application fields named "output" (e.g.
// fields.New("output", ...) in a non-structured-output section) are NOT
// touched.
func applyFieldNameRules(pass *analysis.Pass, file *ast.File, settingsImp importNames) {
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.CallExpr:
			handleFieldAccessCall(pass, n, settingsImp)
		case *ast.IndexExpr:
			handleMapIndex(pass, n)
		case *ast.CompositeLit:
			handleMapLiteralKey(pass, n, settingsImp)
		}
		return true
	})
}

// handleFieldAccessCall handles GetField/GetParameter/UpdateExistingValue calls
// that reference the structured-output section's "output" field.
func handleFieldAccessCall(pass *analysis.Pass, call *ast.CallExpr, settingsImp importNames) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	method := sel.Sel.Name
	switch method {
	case "GetField", "GetParameter":
		// Signature: GetField(slug, fieldName). The first arg must be the
		// GlazedSlug or StructuredOutputSlug selector; the second is the field
		// name string literal.
		if len(call.Args) < 2 {
			return
		}
		if !isSlugSelector(pass, call.Args[0], settingsImp) {
			return
		}
		renameStringArg(pass, call.Args[1], method)
	case "UpdateExistingValue":
		// Signature: UpdateExistingValue(fieldName, value, ...). The first arg
		// is the field name. We cannot easily prove the receiver is the glazed
		// layer, but this method is the idiomatic glazed-layer mutation, so we
		// report (with fix) when the field name is "output".
		if len(call.Args) < 1 {
			return
		}
		renameStringArg(pass, call.Args[0], method)
	}
}

// isSlugSelector reports whether expr is a settings.GlazedSlug or
// settings.StructuredOutputSlug selector (or the dot-imported equivalent).
func isSlugSelector(pass *analysis.Pass, expr ast.Expr, settingsImp importNames) bool {
	_, ok := selectorMatches(pass, expr, settingsImp, settingsImportPath, oldSlug, newSlug)
	if ok {
		return true
	}
	// Also accept a dot-imported bare ident.
	if id, ok := expr.(*ast.Ident); ok && settingsImp.dot && (id.Name == oldSlug || id.Name == newSlug) {
		return true
	}
	return false
}

// renameStringArg emits a fix renaming the string literal "output" to "format"
// at the given position, with value validation.
func renameStringArg(pass *analysis.Pass, expr ast.Expr, method string) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil || val != "output" {
		return
	}
	diag := analysis.Diagnostic{
		Pos:     lit.Pos(),
		End:     lit.End(),
		Message: "rename field name \"output\" to \"format\" in " + method + " call (the structured-output flag was renamed)",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "rename field to format",
			TextEdits: []analysis.TextEdit{{
				Pos:     lit.Pos(),
				End:     lit.End(),
				NewText: []byte(strconv.Quote("format")),
			}},
		}},
	}
	pass.Report(diag)
}

// handleMapIndex handles map index expressions of the form expr["output"].
// These are reported (not auto-fixed) because the analyzer cannot reliably
// prove the map is the structured-output section's values map.
func handleMapIndex(pass *analysis.Pass, idx *ast.IndexExpr) {
	lit, ok := idx.Index.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil || val != "output" {
		return
	}
	pass.Report(analysis.Diagnostic{
		Pos:     lit.Pos(),
		End:     lit.End(),
		Message: "map key \"output\" likely refers to the renamed structured-output flag; rename to \"format\" if this map holds structured-output section values",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "rename key to format",
			TextEdits: []analysis.TextEdit{{
				Pos:     lit.Pos(),
				End:     lit.End(),
				NewText: []byte(strconv.Quote("format")),
			}},
		}},
	})
}

// handleMapLiteralKey handles map composite literals whose keys are the string
// "output" when the map is keyed by a GlazedSlug/StructuredOutputSlug (the
// sources.FromMap pattern). When the slug key is present in the same literal,
// the "output" key is auto-fixed; otherwise it is reported only.
func handleMapLiteralKey(pass *analysis.Pass, lit *ast.CompositeLit, settingsImp importNames) {
	// Look for the pattern: map[string]map[string]interface{}{ GlazedSlug: {"output": ...} }
	// or map[...]{ StructuredOutputSlug: {"output": ...} }
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		// Check if the key is a slug selector.
		if !isSlugSelector(pass, kv.Key, settingsImp) {
			continue
		}
		// The value should be a nested map literal.
		inner, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}
		fixKeyInMapLiteral(pass, inner)
	}
}
