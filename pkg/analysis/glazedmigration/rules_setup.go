package glazedmigration

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// applySetupRules handles R6/R7/R8: report-only diagnostics for the removed
// runtime helpers. These are not auto-fixed because their replacements change
// the return tuple and call structure.
//
//   - SetupTableProcessor + SetupProcessorOutput collapse into a single
//     SetupStructuredOutput(values, writer) call returning
//     (*TableProcessor, OutputFormatter, error).
//   - SetupTableOutputFormatter / SetupRowOutputFormatter / SetupSimpleTableProcessor
//     are replaced by SetupStructuredOutput.
//   - NewOutputFormatterSettings is replaced by DecodeStructuredOutputSettings;
//     the old OutputFormatterSettings type is gone, replaced by
//     StructuredOutputSettings, and field .Output becomes .Format.
func applySetupRules(pass *analysis.Pass, file *ast.File, imports importNames) {
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := selectorMatches(pass, call.Fun, imports, settingsImportPath, setupHelperNames()...)
		if !ok {
			return true
		}
		replacement := setupHelpers[ident.Name]
		msg := "replace settings." + ident.Name + " with settings." + replacement
		switch ident.Name {
		case "SetupTableProcessor":
			// If the same function also calls SetupProcessorOutput, the pair
			// collapses into SetupStructuredOutput (which attaches a formatter
			// and needs a writer). Otherwise the standalone processor contract
			// is preserved by SetupStructuredProcessor (no writer, no formatter).
			if functionContainsCall(file, call.Pos(), "SetupProcessorOutput") {
				msg += "; collapse SetupTableProcessor + SetupProcessorOutput into a single SetupStructuredOutput(values, writer) call; note the return tuple changes to (*TableProcessor, OutputFormatter, error)"
			} else {
				msg = "replace settings.SetupTableProcessor with settings.SetupStructuredProcessor"
				msg += "; the standalone processor contract is preserved by SetupStructuredProcessor (no writer, no formatter); note the return tuple changes to (*TableProcessor, *StructuredOutputSettings, error)"
			}
		case "SetupProcessorOutput":
			msg += "; collapse SetupTableProcessor + SetupProcessorOutput into a single SetupStructuredOutput(values, writer) call; note the return tuple changes to (*TableProcessor, OutputFormatter, error)"
		case "SetupTableOutputFormatter", "SetupRowOutputFormatter", "SetupSimpleTableProcessor":
			msg += "; restructure to use SetupStructuredOutput(values, writer), which builds the processor and attaches the formatter"
		case "NewOutputFormatterSettings":
			msg += "; the old OutputFormatterSettings type is removed, replaced by StructuredOutputSettings; field .Output becomes .Format"
		}
		pass.Report(analysis.Diagnostic{
			Pos:     ident.Pos(),
			End:     ident.End(),
			Message: msg,
		})
		return true
	})
}

func setupHelperNames() []string {
	out := make([]string, 0, len(setupHelpers))
	for n := range setupHelpers {
		out = append(out, n)
	}
	return out
}

// functionContainsCall reports whether the function enclosing pos contains a
// call to a selector named name (e.g. "SetupProcessorOutput"). It walks only
// the enclosing function body.
func functionContainsCall(file *ast.File, pos token.Pos, name string) bool {
	fn := findEnclosingFuncDecl(file, pos)
	if fn == nil || fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if found {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && sel.Sel.Name == name {
			found = true
			return false
		}
		return true
	})
	return found
}
