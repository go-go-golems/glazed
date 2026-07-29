package glazedmigration

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/settings"
	"golang.org/x/tools/go/analysis"
)

// applyWrapperAndKeyRules handles R3 (unwrap With*SectionOptions wrappers) and
// R4 (rename default-map key "output" -> "format").
//
// R3: a settings.With*SectionOptions(args...) call is a no-op adapter from
// []schema.SectionOption to GlazeSectionOption. The replacement constructor
// NewStructuredOutputSection accepts schema.SectionOption directly, so the
// wrapper's arguments should be spliced into the enclosing constructor call.
//
// R4: a schema.WithDefaults(map[string]interface{}{...}) literal whose map
// contains the key "output" must be rewritten to "format", because the new
// section renamed the flag and InitializeDefaultsFromMap errors on unknown
// keys. Unsupported values (sql/template/markdown/excel) are reported only.
//
// Both rules also fire on standalone occurrences (not nested in a constructor),
// so that code which calls With*SectionOptions or WithDefaults directly is
// still flagged.
func applyWrapperAndKeyRules(pass *analysis.Pass, file *ast.File, settingsImp, schemaImp importNames) {
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		// R3: unwrap only the output-section options wrapper. Its arguments
		// targeted the output subsection (replaced by the structured-output
		// section), so they are valid schema.SectionOption values to splice
		// directly into NewStructuredOutputSection.
		if ident, ok := selectorMatches(pass, call.Fun, settingsImp, settingsImportPath, outputSectionOptionsWrapper); ok {
			reportWrapperUnwrap(pass, call, ident)
		}

		// R9: feature-section option wrappers (select/template/rename/etc.)
		// targeted removed subsections. Report them for manual redesign; do not
		// unwrap, because their args configure the wrong schema.
		if ident, ok := selectorMatches(pass, call.Fun, settingsImp, settingsImportPath, featureWrapperNames()...); ok {
			pass.Report(analysis.Diagnostic{
				Pos:     ident.Pos(),
				End:     ident.End(),
				Message: "settings." + ident.Name + " targets a removed feature subsection with no mechanical migration; redesign using application fields or caller-side tools",
			})
		}

		// R4: rename "output" -> "format" in schema.WithDefaults map literals
		// that target the structured-output section. Only fire when the
		// WithDefaults call is an argument to a known structured-output
		// constructor (new or removed), so application-defined sections with a
		// legitimate "output" field are not touched.
		if ident, ok := selectorMatches(pass, call.Fun, schemaImp, schemaImportPath, "WithDefaults"); ok {
			if withDefaultsTargetsStructuredOutput(pass, file, call, settingsImp) {
				reportKeyRename(pass, call, ident)
			}
		}
		return true
	})
}

// reportWrapperUnwrap emits a fix that replaces a With*SectionOptions(args...)
// call with its arguments spliced into the enclosing call's argument list.
//
// The edit spans the entire wrapper call expression and replaces it with the
// comma-joined source of its arguments. When the wrapper is the sole argument
// of an enclosing call, this effectively unwraps it.
func reportWrapperUnwrap(pass *analysis.Pass, call *ast.CallExpr, ident *ast.Ident) {
	if len(call.Args) == 0 {
		// No arguments to splice; the wrapper is a no-op. Replace it with
		// nothing by deleting it (handled by the caller's arg-list edit). Emit
		// a diagnostic without a fix here; the constructor rule will handle
		// the surrounding call.
		pass.Report(analysis.Diagnostic{
			Pos:     ident.Pos(),
			End:     ident.End(),
			Message: "settings." + ident.Name + " is a removed no-op wrapper; remove it and pass its arguments directly to NewStructuredOutputSection",
		})
		return
	}

	// Build the replacement text: the arguments joined by ", ".
	var parts []string
	for _, arg := range call.Args {
		text := exprText(pass.Fset, arg)
		if text == "" {
			// Cannot serialize; report only.
			pass.Report(analysis.Diagnostic{
				Pos:     ident.Pos(),
				End:     ident.End(),
				Message: "settings." + ident.Name + " is a removed wrapper; unwrap it and pass its arguments directly to NewStructuredOutputSection",
			})
			return
		}
		parts = append(parts, text)
	}
	replacement := joinArgs(parts)

	pass.Report(analysis.Diagnostic{
		Pos:     call.Pos(),
		End:     call.End(),
		Message: "unwrap removed settings." + ident.Name + " wrapper; pass its arguments directly to NewStructuredOutputSection",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "unwrap " + ident.Name,
			TextEdits: []analysis.TextEdit{{
				Pos:     call.Pos(),
				End:     call.End(),
				NewText: []byte(replacement),
			}},
		}},
	})
}

// withDefaultsTargetsStructuredOutput reports whether a schema.WithDefaults call
// is an argument to a known structured-output section constructor (new or
// removed), meaning its "output" key refers to the renamed flag rather than an
// application-defined field.
func withDefaultsTargetsStructuredOutput(pass *analysis.Pass, file *ast.File, call *ast.CallExpr, settingsImp importNames) bool {
	parent := findEnclosingCallExpr(file, call.Pos())
	if parent == nil {
		return false
	}
	// Match the new and removed structured-output constructors.
	if _, ok := selectorMatches(pass, parent.Fun, settingsImp, settingsImportPath, newConstructor, oldConstructor, altConstructor, altConstructor2); ok {
		return true
	}
	return false
}

// findEnclosingCallExpr finds the *ast.CallExpr that directly contains pos as
// one of its arguments.
func findEnclosingCallExpr(file *ast.File, pos token.Pos) *ast.CallExpr {
	var found *ast.CallExpr
	ast.Inspect(file, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, arg := range call.Args {
			if arg.Pos() == pos {
				found = call
				return false
			}
		}
		return true
	})
	return found
}

// reportKeyRename inspects a schema.WithDefaults call for a map literal with
// the key "output" and emits a fix renaming it to "format". Values outside the
// supported set are reported only.
func reportKeyRename(pass *analysis.Pass, call *ast.CallExpr, ident *ast.Ident) {
	if len(call.Args) == 0 {
		return
	}
	// The argument is typically a map[string]interface{}{...} composite literal
	// or a struct. We only auto-fix map literals with string keys.
	for _, arg := range call.Args {
		if lit, ok := arg.(*ast.CompositeLit); ok {
			fixKeyInMapLiteral(pass, lit)
		}
	}
}

// fixKeyInMapLiteral scans a composite literal's elements for {key, value}
// pairs where the key is the string "output" and emits a rename fix. It also
// validates the value against the supported format set.
func fixKeyInMapLiteral(pass *analysis.Pass, lit *ast.CompositeLit) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		keyLit, ok := kv.Key.(*ast.BasicLit)
		if !ok || keyLit.Kind != token.STRING {
			continue
		}
		key, err := strconv.Unquote(keyLit.Value)
		if err != nil || key != "output" {
			continue
		}
		// Found the "output" key. Validate the value.
		valueStr := basicLitStringValue(kv.Value)
		if valueStr != "" && !supportedFormat(valueStr) {
			pass.Report(analysis.Diagnostic{
				Pos:     kv.Value.Pos(),
				End:     kv.Value.End(),
				Message: "value " + strconv.Quote(valueStr) + " is not a supported structured-output format; choose table|json|jsonl|csv|tsv|yaml",
			})
			// Still rename the key so the section constructs; the value must be
			// fixed by hand.
		}
		pass.Report(analysis.Diagnostic{
			Pos:     keyLit.Pos(),
			End:     keyLit.End(),
			Message: "rename default-map key \"output\" to \"format\" (the structured-output flag was renamed)",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "rename key to format",
				TextEdits: []analysis.TextEdit{{
					Pos:     keyLit.Pos(),
					End:     keyLit.End(),
					NewText: []byte(strconv.Quote("format")),
				}},
			}},
		})
	}
}

// basicLitStringValue returns the string value of a *ast.BasicLit string
// literal, or "" if the expression is not a string literal.
func basicLitStringValue(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	v, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return v
}

// supportedFormat reports whether s is one of the structured-output formats.
// The allowlist is derived from the settings package's exported accessor to
// avoid drift if formats are added or removed.
func supportedFormat(s string) bool {
	for _, f := range settings.StructuredOutputFormats() {
		if f == s {
			return true
		}
	}
	return false
}

func joinArgs(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
