package glazedclilint

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	diagnosticRawEnv             = "use Glazed config/env middleware or an explicit command field instead of os.Getenv in CLI code"
	diagnosticRawFlags           = "define CLI flags with cmds.WithFlags(fields.New(...)) instead of raw Cobra/pflag/flag APIs"
	diagnosticGlazedWithoutRows  = "this command exposes Glazed output flags but does not implement RunIntoGlazeProcessor"
	diagnosticInvalidSuppression = "glazedclilint suppression requires a reason"
	glazedSettingsPkg            = "github.com/go-go-golems/glazed/pkg/settings"
	glazedCmdsPkg                = "github.com/go-go-golems/glazed/pkg/cmds"
	pflagPkg                     = "github.com/spf13/pflag"
	suppressionIgnorePrefix      = "glazedclilint:ignore"
	suppressionFileIgnorePrefix  = "glazedclilint:file-ignore"
)

var (
	allowTestsFlag     bool
	allowGeneratedFlag bool
	allowPathsFlag     string
)

// Analyzer enforces Glazed CLI command policy that the Go compiler cannot express:
// no ad-hoc os.Getenv reads, no raw Cobra/pflag/flag definitions in verbs, and no
// Glazed output section on commands that do not emit structured rows through the
// Glazed processor pipeline.
var Analyzer = &analysis.Analyzer{
	Name:     "glazedclilint",
	Doc:      "enforce Glazed CLI policy: avoid raw env reads, raw flag APIs, and Glazed output flags on non-row commands",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&allowTestsFlag, "allow-tests", true, "skip diagnostics in _test.go files")
	Analyzer.Flags.BoolVar(&allowGeneratedFlag, "allow-generated", true, "skip diagnostics in generated files")
	Analyzer.Flags.StringVar(
		&allowPathsFlag,
		"allow-paths",
		"pkg/analysis/,pkg/cli/,pkg/cmds/fields/,pkg/cmds/logging/,pkg/cmds/sources/,pkg/help/",
		"comma-separated slash-normalized path fragments to skip, for framework bridge code",
	)
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	fileInfo := buildFileInfo(pass)
	reportInvalidSuppressions(pass, fileInfo)
	allowedPaths := splitCSV(allowPathsFlag)

	// Function-local analysis is needed for the Glazed-section rule because it has
	// to connect a local variable initialized from settings.NewGlazedSection to a
	// later cmds.WithSections call and the command type returned by the constructor.
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}
			analyzeFunction(pass, fileInfo, allowedPaths, fn)
			return false
		})
	}

	// Call-local checks do not need function context.
	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if shouldSkip(pass, fileInfo, allowedPaths, call.Pos()) {
			return
		}
		checkRawEnv(pass, fileInfo, call)
		checkRawFlags(pass, fileInfo, call)
	})

	return nil, nil
}

func analyzeFunction(
	pass *analysis.Pass,
	fileInfo map[*ast.File]fileMeta,
	allowedPaths []string,
	fn *ast.FuncDecl,
) {
	commandType := inferredCommandReturnType(pass, fn)
	if commandType == nil || hasRunIntoGlazeProcessor(commandType) {
		return
	}

	glazedVars := map[types.Object]bool{}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			for i, rhs := range node.Rhs {
				if !isGlazedSectionConstructorCall(pass, rhs) || i >= len(node.Lhs) {
					continue
				}
				if ident, ok := node.Lhs[i].(*ast.Ident); ok {
					if obj := pass.TypesInfo.ObjectOf(ident); obj != nil {
						glazedVars[obj] = true
					}
				}
			}
		case *ast.ValueSpec:
			for i, rhs := range node.Values {
				if !isGlazedSectionConstructorCall(pass, rhs) || i >= len(node.Names) {
					continue
				}
				if obj := pass.TypesInfo.ObjectOf(node.Names[i]); obj != nil {
					glazedVars[obj] = true
				}
			}
		case *ast.CallExpr:
			if shouldSkip(pass, fileInfo, allowedPaths, node.Pos()) {
				return true
			}
			if !isCmdsWithSectionsCall(pass, node) {
				return true
			}
			for _, arg := range node.Args {
				if isTrackedGlazedSectionArg(pass, glazedVars, arg) {
					reportDiagnostic(pass, fileInfo, arg.Pos(), diagnosticGlazedWithoutRows)
					break
				}
			}
		}
		return true
	})
}

func checkRawEnv(pass *analysis.Pass, fileInfo map[*ast.File]fileMeta, call *ast.CallExpr) {
	fn := calledFunction(pass, call)
	if fn == nil || fn.Pkg() == nil {
		return
	}
	if fn.Pkg().Path() == "os" && fn.Name() == "Getenv" {
		reportDiagnostic(pass, fileInfo, call.Pos(), diagnosticRawEnv)
	}
}

func checkRawFlags(pass *analysis.Pass, fileInfo map[*ast.File]fileMeta, call *ast.CallExpr) {
	fn := calledFunction(pass, call)
	if fn == nil {
		return
	}

	if fn.Pkg() != nil {
		switch fn.Pkg().Path() {
		case "flag", pflagPkg:
			if isFlagDefinitionName(fn.Name()) {
				reportDiagnostic(pass, fileInfo, call.Pos(), diagnosticRawFlags)
				return
			}
		}
	}

	if !isFlagDefinitionName(fn.Name()) {
		return
	}
	if receiverIsPFlagSet(pass, call) {
		reportDiagnostic(pass, fileInfo, call.Pos(), diagnosticRawFlags)
	}
}

func calledFunction(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	fun := call.Fun
	if idx, ok := fun.(*ast.IndexExpr); ok {
		fun = idx.X
	}
	if idx, ok := fun.(*ast.IndexListExpr); ok {
		fun = idx.X
	}

	switch f := fun.(type) {
	case *ast.SelectorExpr:
		obj := pass.TypesInfo.Uses[f.Sel]
		fn, _ := obj.(*types.Func)
		return fn
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[f]
		fn, _ := obj.(*types.Func)
		return fn
	default:
		return nil
	}
}

func receiverIsPFlagSet(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	tv, ok := pass.TypesInfo.Types[sel.X]
	if !ok || tv.Type == nil {
		return false
	}
	return isPFlagSetType(tv.Type)
}

func isPFlagSetType(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == pflagPkg && named.Obj().Name() == "FlagSet"
}

func isFlagDefinitionName(name string) bool {
	switch name {
	case "Bool", "BoolP", "BoolVar", "BoolVarP",
		"String", "StringP", "StringVar", "StringVarP",
		"Int", "IntP", "IntVar", "IntVarP",
		"Int64", "Int64P", "Int64Var", "Int64VarP",
		"Float64", "Float64P", "Float64Var", "Float64VarP",
		"Duration", "DurationP", "DurationVar", "DurationVarP",
		"StringSlice", "StringSliceP", "StringSliceVar", "StringSliceVarP",
		"StringArray", "StringArrayP", "StringArrayVar", "StringArrayVarP",
		"Var", "VarP", "NewFlagSet", "Parse":
		return true
	default:
		return false
	}
}

func isGlazedSectionConstructorCall(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := unwrapParens(expr).(*ast.CallExpr)
	if !ok {
		return false
	}
	fn := calledFunction(pass, call)
	if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != glazedSettingsPkg {
		return false
	}
	switch fn.Name() {
	case "NewGlazedSection", "NewGlazedSchema":
		return true
	default:
		return false
	}
}

func isCmdsWithSectionsCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	fn := calledFunction(pass, call)
	return fn != nil && fn.Pkg() != nil && fn.Pkg().Path() == glazedCmdsPkg && fn.Name() == "WithSections"
}

func isTrackedGlazedSectionArg(pass *analysis.Pass, glazedVars map[types.Object]bool, expr ast.Expr) bool {
	expr = unwrapParens(expr)
	if isGlazedSectionConstructorCall(pass, expr) {
		return true
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		obj = pass.TypesInfo.Uses[ident]
	}
	return glazedVars[obj]
}

func inferredCommandReturnType(pass *analysis.Pass, fn *ast.FuncDecl) *types.Named {
	if fn.Type.Results == nil {
		return nil
	}
	for _, result := range fn.Type.Results.List {
		t := pass.TypesInfo.TypeOf(result.Type)
		if named := namedFromPointerOrNamed(t); named != nil {
			return named
		}
	}
	return nil
}

func namedFromPointerOrNamed(t types.Type) *types.Named {
	switch tt := t.(type) {
	case *types.Named:
		return tt
	case *types.Pointer:
		if named, ok := tt.Elem().(*types.Named); ok {
			return named
		}
	}
	return nil
}

func hasRunIntoGlazeProcessor(named *types.Named) bool {
	if named == nil {
		return false
	}
	methodSet := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < methodSet.Len(); i++ {
		selection := methodSet.At(i)
		if selection.Obj() == nil || selection.Obj().Name() != "RunIntoGlazeProcessor" {
			continue
		}
		fn, ok := selection.Obj().(*types.Func)
		if !ok {
			continue
		}
		sig, ok := fn.Type().(*types.Signature)
		if !ok {
			continue
		}
		if sig.Params().Len() == 3 && sig.Results().Len() == 1 {
			return true
		}
	}
	return false
}

func unwrapParens(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = paren.X
	}
}

type fileMeta struct {
	filename     string
	generated    bool
	suppressions suppressionSet
}

type suppressionSet struct {
	fileIgnore      bool
	ignoredLines    map[int]bool
	invalidComments []token.Pos
}

func buildFileInfo(pass *analysis.Pass) map[*ast.File]fileMeta {
	ret := map[*ast.File]fileMeta{}
	for _, file := range pass.Files {
		pos := pass.Fset.Position(file.Pos())
		ret[file] = fileMeta{
			filename:     filepath.ToSlash(pos.Filename),
			generated:    isGeneratedFile(file),
			suppressions: parseSuppressions(pass, file),
		}
	}
	return ret
}

func parseSuppressions(pass *analysis.Pass, file *ast.File) suppressionSet {
	set := suppressionSet{ignoredLines: map[int]bool{}}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			text := normalizedCommentText(comment.Text)
			switch {
			case strings.HasPrefix(text, suppressionFileIgnorePrefix):
				if suppressionReason(text, suppressionFileIgnorePrefix) == "" {
					set.invalidComments = append(set.invalidComments, comment.Slash)
					continue
				}
				set.fileIgnore = true
			case strings.HasPrefix(text, suppressionIgnorePrefix):
				if suppressionReason(text, suppressionIgnorePrefix) == "" {
					set.invalidComments = append(set.invalidComments, comment.Slash)
					continue
				}
				start := pass.Fset.Position(comment.Slash)
				end := pass.Fset.Position(comment.End())
				set.ignoredLines[start.Line] = true
				set.ignoredLines[end.Line] = true
				if nextStart, nextEnd := nextNodeRange(pass, file, comment.End()); nextStart > 0 {
					for line := nextStart; line <= nextEnd; line++ {
						set.ignoredLines[line] = true
					}
				}
			}
		}
	}
	return set
}

func normalizedCommentText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "//")
	s = strings.TrimPrefix(s, "/*")
	s = strings.TrimSuffix(s, "*/")
	s = strings.TrimSpace(s)
	// analysistest expectations live in comments as "// want ...". They are
	// not part of the suppression syntax under test.
	if i := strings.Index(s, " // want "); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	return s
}

func suppressionReason(text, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(text, prefix))
}

func nextNodeRange(pass *analysis.Pass, file *ast.File, after token.Pos) (int, int) {
	var best ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		pos := n.Pos()
		if pos <= after {
			return true
		}
		if best == nil || pos < best.Pos() {
			best = n
		}
		return true
	})
	if best == nil {
		return 0, 0
	}
	start := pass.Fset.Position(best.Pos()).Line
	end := pass.Fset.Position(best.End()).Line
	if end < start {
		end = start
	}
	return start, end
}

func reportInvalidSuppressions(pass *analysis.Pass, fileInfo map[*ast.File]fileMeta) {
	for _, meta := range fileInfo {
		for _, pos := range meta.suppressions.invalidComments {
			pass.Reportf(pos, diagnosticInvalidSuppression)
		}
	}
}

func reportDiagnostic(pass *analysis.Pass, fileInfo map[*ast.File]fileMeta, pos token.Pos, message string) {
	if isSuppressed(pass, fileInfo, pos) {
		return
	}
	pass.Report(analysis.Diagnostic{Pos: pos, Message: message})
}

func isSuppressed(pass *analysis.Pass, fileInfo map[*ast.File]fileMeta, pos token.Pos) bool {
	if fileInfo == nil {
		return false
	}
	filename := filepath.ToSlash(pass.Fset.Position(pos).Filename)
	line := pass.Fset.Position(pos).Line
	for _, meta := range fileInfo {
		if meta.filename != filename {
			continue
		}
		if meta.suppressions.fileIgnore {
			return true
		}
		return meta.suppressions.ignoredLines[line]
	}
	return false
}

func shouldSkip(pass *analysis.Pass, fileInfo map[*ast.File]fileMeta, allowedPaths []string, pos token.Pos) bool {
	filename := filepath.ToSlash(pass.Fset.Position(pos).Filename)
	if filename == "" {
		return false
	}
	if allowTestsFlag && strings.HasSuffix(filename, "_test.go") {
		return true
	}
	for _, meta := range fileInfo {
		if meta.filename == filename {
			if allowGeneratedFlag && meta.generated {
				return true
			}
			break
		}
	}
	// Do not apply production path allowlists to analysistest fixtures. Testdata
	// intentionally mirrors real import paths under pkg/analysis/... and still
	// needs diagnostics to fire.
	if !strings.Contains(filename, "/testdata/src/") {
		for _, allowed := range allowedPaths {
			if allowed != "" && strings.Contains(filename, allowed) {
				return true
			}
		}
	}
	return false
}

func isGeneratedFile(file *ast.File) bool {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			text := strings.TrimPrefix(comment.Text, "//")
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSpace(text)
			if strings.Contains(text, "Code generated") && strings.Contains(text, "DO NOT EDIT") {
				return true
			}
		}
	}
	return false
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	ret := make([]string, 0, len(parts))
	for _, part := range parts {
		part = filepath.ToSlash(strings.TrimSpace(part))
		if part != "" {
			ret = append(ret, part)
		}
	}
	return ret
}
