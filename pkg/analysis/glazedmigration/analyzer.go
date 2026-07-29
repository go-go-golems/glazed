// Package glazedmigration provides source migrations for removed Glazed APIs.
//
// The analyzer rewrites and reports call sites that reference Glazed settings
// APIs removed by the GLZ-OUTPUT-FLAGS-CLEANUP cleanup. It covers:
//
//   - NewGlazedSchema / NewGlazedSection -> NewStructuredOutputSection (R1/R2)
//   - With*SectionOptions wrappers -> unwrapped into the constructor call (R3)
//   - default-map key "output" -> "format" (R4)
//   - GlazedSlug -> StructuredOutputSlug (R5)
//   - Setup* runtime helpers -> report-only (R6/R7/R8)
//   - removed feature sections -> report-only (R9)
//
// It runs despite type errors because the removed symbols make consuming code
// fail to type-check.
package glazedmigration

import (
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	settingsImportPath = "github.com/go-go-golems/glazed/pkg/settings"
	schemaImportPath   = "github.com/go-go-golems/glazed/pkg/cmds/schema"
	cmdsImportPath     = "github.com/go-go-golems/glazed/pkg/cmds"

	oldConstructor  = "NewGlazedSchema"
	altConstructor  = "NewGlazedSection"
	altConstructor2 = "NewOutputSection"
	newConstructor  = "NewStructuredOutputSection"

	oldSlug = "GlazedSlug"
	newSlug = "StructuredOutputSlug"

	glazeCommandMethod = "RunIntoGlazeProcessor"
)

// featureSectionOptionWrappers are the removed settings.With*SectionOptions
// adapters that targeted distinct feature subsections (select, rename, replace,
// template, jq, sort, skip-limit, fields-filters) which no longer exist.
// Their arguments configured those removed subsections, so unwrapping them
// into NewStructuredOutputSection would pass wrong-schema defaults (e.g.
// template-file or select defaults) that fail as unknown fields. They are
// reported (R9) for manual redesign, never auto-unwrapped.
var featureSectionOptionWrappers = map[string]bool{
	"WithSelectSectionOptions":        true,
	"WithTemplateSectionOptions":      true,
	"WithRenameSectionOptions":        true,
	"WithReplaceSectionOptions":       true,
	"WithFieldsFiltersSectionOptions": true,
	"WithJqSectionOptions":            true,
	"WithSortSectionOptions":          true,
	"WithSkipLimitSectionOptions":     true,
}

// outputSectionOptionsWrapper is the only With*SectionOptions adapter whose
// arguments targeted the output subsection (replaced by the structured-output
// section). It is a safe no-op to unwrap into NewStructuredOutputSection.
const outputSectionOptionsWrapper = "WithOutputSectionOptions"

// removedFeatureSectionConstructors built the now-deleted feature sections
// (select, rename, replace, template, jq, sort, skip-limit, fields-filters).
// They have no mechanical migration and must be redesigned by hand.
var removedFeatureSectionConstructors = map[string]bool{
	"NewFieldsFiltersSection": true,
	"NewSelectSection":        true,
	"NewRenameSection":        true,
	"NewReplaceSection":       true,
	"NewTemplateSection":      true,
	"NewJqSection":            true,
	"NewSortSection":          true,
	"NewSkipLimitSection":     true,
}

// setupHelpers are the removed runtime helpers. They are reported (not
// auto-fixed) because their replacements change the return tuple and call
// structure.
var setupHelpers = map[string]string{
	"SetupTableProcessor":        "SetupStructuredOutput",
	"SetupProcessorOutput":       "SetupStructuredOutput",
	"SetupTableOutputFormatter":  "SetupStructuredOutput",
	"SetupRowOutputFormatter":    "SetupStructuredOutput",
	"SetupSimpleTableProcessor":  "SetupStructuredOutput",
	"NewOutputFormatterSettings": "DecodeStructuredOutputSettings",
}

// Analyzer rewrites and reports calls to removed Glazed settings APIs.
var Analyzer = &analysis.Analyzer{
	Name:             "glazedmigration",
	Doc:              "migrate removed Glazed APIs to their structured-output replacements",
	RunDespiteErrors: true,
	Run:              run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		settingsImp := settingsImports(file)
		schemaImp := schemaImports(file)
		if !settingsImp.used() && !schemaImp.used() {
			continue
		}

		applyConstructorRules(pass, file, settingsImp, schemaImp)
		applySlugRules(pass, file, settingsImp)
		applySetupRules(pass, file, settingsImp)
		applyRemovedFeatureRules(pass, file, settingsImp)
		// Wrapper unwrapping (R3) and key rename (R4) operate on the schema
		// package's WithDefaults calls nested inside settings constructor args.
		applyWrapperAndKeyRules(pass, file, settingsImp, schemaImp)
		// Extended R4: rename the "output" field name to "format" in GetField,
		// GetParameter, UpdateExistingValue, map indices, and FromMap-style
		// literals that reference the structured-output section.
		applyFieldNameRules(pass, file, settingsImp)
	}
	return nil, nil
}

// --- import detection ---

type importNames struct {
	qualified map[string]bool
	dot       bool
}

func (i importNames) used() bool {
	return len(i.qualified) > 0 || i.dot
}

func settingsImports(file *ast.File) importNames {
	return importsFor(file, settingsImportPath)
}

func schemaImports(file *ast.File) importNames {
	return importsFor(file, schemaImportPath)
}

func importsFor(file *ast.File, path string) importNames {
	ret := importNames{qualified: map[string]bool{}}
	for _, spec := range file.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil || p != path {
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
		ret.qualified[defaultPkgName(path)] = true
	}
	return ret
}

func defaultPkgName(path string) string {
	switch path {
	case settingsImportPath:
		return "settings"
	case schemaImportPath:
		return "schema"
	case cmdsImportPath:
		return "cmds"
	default:
		// fall back to the last path element
		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == '/' {
				return path[i+1:]
			}
		}
		return path
	}
}

// --- shared selector resolution ---

// selectorMatches reports whether expr is a selector `pkg.Sel` (or a dot-import
// ident `Sel`) whose qualifier resolves to the given package and whose selector
// name is one of the given names. It returns the matched ident (the Sel) when
// true. The importPath is used for type-checker validation.
func selectorMatches(pass *analysis.Pass, expr ast.Expr, imports importNames, importPath string, names ...string) (*ast.Ident, bool) {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	switch fn := expr.(type) {
	case *ast.SelectorExpr:
		qualifier, ok := fn.X.(*ast.Ident)
		if !ok || !nameSet[fn.Sel.Name] || !imports.qualified[qualifier.Name] {
			return nil, false
		}
		// A local declaration can shadow the import in a nested scope. Check the
		// qualifier even when the removed selector itself could not be resolved.
		if obj := pass.TypesInfo.Uses[qualifier]; obj != nil {
			pkgName, ok := obj.(*types.PkgName)
			if !ok || pkgName.Imported() == nil || pkgName.Imported().Path() != importPath {
				return nil, false
			}
		}
		// When type information is available, reject selectors resolved to another package.
		if obj := pass.TypesInfo.Uses[fn.Sel]; obj != nil {
			if typed, ok := obj.(*types.Func); !ok || typed.Pkg() == nil || typed.Pkg().Path() != importPath {
				return nil, false
			}
		}
		return fn.Sel, true
	case *ast.Ident:
		if !imports.dot || !nameSet[fn.Name] {
			return nil, false
		}
		if obj := pass.TypesInfo.Uses[fn]; obj != nil {
			typed, ok := obj.(*types.Func)
			if !ok || typed.Pkg() == nil || typed.Pkg().Path() != importPath {
				return nil, false
			}
		}
		return fn, true
	default:
		return nil, false
	}
}

// dotImportReplacementIsSafe reports whether replacing a dot-imported call with
// newConstructor would still resolve to the Glazed settings package, i.e. no
// local declaration shadows it.
func dotImportReplacementIsSafe(pass *analysis.Pass, file *ast.File, call *ast.CallExpr, replacement string) bool {
	if _, dotCall := call.Fun.(*ast.Ident); !dotCall {
		return true
	}

	// Prefer type-checker scope data.
	var innermost *types.Scope
	for _, scope := range pass.TypesInfo.Scopes {
		if scope.Contains(call.Pos()) && (innermost == nil || scope.Pos() >= innermost.Pos()) {
			innermost = scope
		}
	}
	if innermost != nil {
		_, obj := innermost.LookupParent(replacement, call.Pos())
		if obj != nil {
			fn, ok := obj.(*types.Func)
			return ok && fn.Pkg() != nil && fn.Pkg().Path() == settingsImportPath
		}
	}

	// Conservative fallback: withhold the fix if any declaration with the
	// replacement name is present.
	safe := true
	ast.Inspect(file, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if ok && ident.Name == replacement && ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
			safe = false
			return false
		}
		return safe
	})
	return safe
}

// exprText renders an AST expression back to source text for use in a TextEdit.
// It uses go/format when possible and falls back to an empty string.
func exprText(fset *token.FileSet, expr ast.Expr) string {
	var b strings.Builder
	if err := format.Node(&b, fset, expr); err != nil {
		return ""
	}
	return b.String()
}
