package glazedmigration

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// applyConstructorRules handles R1 (redundant-section deletion when the command
// is provably a GlazeCommand) and R2 (rename NewGlazedSchema/NewGlazedSection
// to NewStructuredOutputSection).
//
// The deletion path (R1) is only attempted when the call has zero arguments
// AND the enclosing command is provably a cmds.GlazeCommand. Otherwise the
// rename (R2) is applied, which is always correct and preserves behavior.
func applyConstructorRules(pass *analysis.Pass, file *ast.File, imports importNames, schemaImp importNames) {
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name, ok := selectorMatches(pass, call.Fun, imports, settingsImportPath, oldConstructor, altConstructor)
		if !ok {
			return true
		}

		// Determine whether the call is the legacy constructor.
		isLegacy := name.Name == oldConstructor || name.Name == altConstructor
		if !isLegacy {
			return true
		}

		diagnostic := analysis.Diagnostic{
			Pos:     name.Pos(),
			End:     name.End(),
			Message: "replace settings." + name.Name + " with settings." + newConstructor,
		}

		// R1: attempt deletion when zero-arg and the command is a GlazeCommand.
		if len(call.Args) == 0 && canDeleteRedundantSection(pass, file, call) {
			if del := deletionFix(pass, file, call); del != nil {
				diagnostic.SuggestedFixes = []analysis.SuggestedFix{*del}
				diagnostic.Message += "; the enclosing command implements GlazeCommand, so the explicit section is redundant and can be deleted"
				pass.Report(diagnostic)
				return true
			}
		}

		// R2: rename the constructor identifier.
		canFix := len(call.Args) == 0 || argsAreSchemaSectionOptions(pass, call, imports, schemaImp)
		if canFix && !dotImportReplacementIsSafe(pass, file, call, newConstructor) {
			canFix = false
			diagnostic.Message += "; a local " + newConstructor + " declaration shadows the dot-imported replacement"
		}
		if canFix {
			diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
				Message: "use " + newConstructor,
				TextEdits: []analysis.TextEdit{{
					Pos:     name.Pos(),
					End:     name.End(),
					NewText: []byte(newConstructor),
				}},
			}}
		} else if len(call.Args) > 0 {
			diagnostic.Message += "; legacy GlazeSectionOption arguments require manual migration"
		}
		pass.Report(diagnostic)
		return true
	})
}

// isSchemaPackageCall reports whether expr is a selector `schema.Sel` (or a
// dot-imported ident) whose qualifier resolves to the glazed schema package.
// Any schema-package call is a valid schema.SectionOption, so we accept all
// selector names.
func isSchemaPackageCall(pass *analysis.Pass, expr ast.Expr, schemaImp importNames) bool {
	switch fn := expr.(type) {
	case *ast.SelectorExpr:
		qualifier, ok := fn.X.(*ast.Ident)
		if !ok || !schemaImp.qualified[qualifier.Name] {
			return false
		}
		if obj := pass.TypesInfo.Uses[qualifier]; obj != nil {
			pkgName, ok := obj.(*types.PkgName)
			if !ok || pkgName.Imported() == nil || pkgName.Imported().Path() != schemaImportPath {
				return false
			}
		}
		if obj := pass.TypesInfo.Uses[fn.Sel]; obj != nil {
			if typed, ok := obj.(*types.Func); !ok || typed.Pkg() == nil || typed.Pkg().Path() != schemaImportPath {
				return false
			}
		}
		return true
	case *ast.Ident:
		// dot-imported schema call
		return schemaImp.dot
	default:
		return false
	}
}
// constructor call is a schema.* call (or a settings.With*SectionOptions
// wrapper that R3 will unwrap). When true, the arguments are valid
// schema.SectionOption values and the constructor can be renamed to
// NewStructuredOutputSection, which accepts schema.SectionOption directly.
func argsAreSchemaSectionOptions(pass *analysis.Pass, call *ast.CallExpr, settingsImp, schemaImp importNames) bool {
	if len(call.Args) == 0 {
		return false
	}
	for _, arg := range call.Args {
		sub, ok := arg.(*ast.CallExpr)
		if !ok {
			return false
		}
		// Accept settings.With*SectionOptions wrappers (R3 will unwrap them) and
		// direct schema.* calls (the unwrapped form).
		if _, ok := selectorMatches(pass, sub.Fun, settingsImp, settingsImportPath, wrapperNames()...); ok {
			continue
		}
		if isSchemaPackageCall(pass, sub.Fun, schemaImp) {
			continue
		}
		return false
	}
	return true
}

func wrapperNames() []string {
	out := make([]string, 0, len(withSectionOptionWrappers))
	for n := range withSectionOptionWrappers {
		out = append(out, n)
	}
	return out
}

// canDeleteRedundantSection reports whether the result of the constructor call
// is assigned to a local variable that is passed to cmds.WithSections, and the
// enclosing command type implements GlazeCommand (RunIntoGlazeProcessor).
//
// This is conservative: it requires the assignment to be a simple
// `name, err := settings.NewGlazedSection()` statement and the variable to
// appear as a direct argument to cmds.WithSections in the same function.
func canDeleteRedundantSection(pass *analysis.Pass, file *ast.File, call *ast.CallExpr) bool {
	// Find the enclosing assignment statement.
	assign := findEnclosingAssignStmt(file, call.Pos())
	if assign == nil {
		return false
	}
	// Expect: lhs = [ident, err], rhs = [call]
	if len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
		return false
	}
	ident, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return false
	}
	// Find a cmds.WithSections call in the same function that uses this ident.
	// The deletion is only safe if the variable is used exactly once (in one
	// WithSections call) within the enclosing function, so we do not leave
	// dangling references in other call sites.
	fn := findEnclosingFuncDecl(file, call.Pos())
	if fn == nil {
		return false
	}
	withSections, count := findWithSectionsUsingIdentInFunc(fn, ident.Name)
	if withSections == nil || count != 1 {
		return false
	}
	// Prove the command type implements GlazeCommand.
	return enclosingTypeIsGlazeCommand(pass, file, call.Pos())
}

// findEnclosingAssignStmt walks the file to find the *ast.AssignStmt whose
// RHS contains pos.
func findEnclosingAssignStmt(file *ast.File, pos token.Pos) *ast.AssignStmt {
	var found *ast.AssignStmt
	ast.Inspect(file, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range assign.Rhs {
			if containsPos(rhs, pos) {
				found = assign
				return false
			}
		}
		return true
	})
	return found
}

func containsPos(node ast.Node, pos token.Pos) bool {
	contained := false
	ast.Inspect(node, func(n ast.Node) bool {
		if contained {
			return false
		}
		if n == nil {
			return true
		}
		if n.Pos() <= pos && pos <= n.End() {
			contained = true
			return false
		}
		return true
	})
	return contained
}

// findWithSectionsUsingIdentInFunc finds cmds.WithSections(...) calls within
// the given function that have the given identifier as a direct argument. It
// returns the first match and the total count of matches. The count lets the
// caller withhold deletion when the variable is used in multiple WithSections
// calls (which would leave dangling references).
func findWithSectionsUsingIdentInFunc(fn *ast.FuncDecl, identName string) (*ast.CallExpr, int) {
	var found *ast.CallExpr
	count := 0
	if fn.Body == nil {
		return nil, 0
	}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "WithSections" {
			return true
		}
		for _, arg := range call.Args {
			if id, ok := arg.(*ast.Ident); ok && id.Name == identName {
				count++
				if found == nil {
					found = call
				}
			}
		}
		return true
	})
	return found, count
}

// enclosingTypeIsGlazeCommand resolves the concrete command type returned by
// the function enclosing pos and checks whether its method set contains
// RunIntoGlazeProcessor.
func enclosingTypeIsGlazeCommand(pass *analysis.Pass, file *ast.File, pos token.Pos) bool {
	// Find the function declaration enclosing pos.
	fn := findEnclosingFuncDecl(file, pos)
	if fn == nil {
		return false
	}
	// Resolve the return type of the function.
	if fn.Type == nil || fn.Type.Results == nil {
		return false
	}
	// Look for a single result that is a pointer-to-named type.
	for _, field := range fn.Type.Results.List {
		star, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		named, ok := star.X.(*ast.Ident)
		if !ok {
			continue
		}
		obj := pass.TypesInfo.Uses[named]
		if obj == nil {
			continue
		}
		typeName, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		namedType, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}
		if methodSetHas(namedType, glazeCommandMethod) {
			return true
		}
	}
	return false
}

func findEnclosingFuncDecl(file *ast.File, pos token.Pos) *ast.FuncDecl {
	var found *ast.FuncDecl
	ast.Inspect(file, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		fn, ok := node.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Pos() <= pos && pos <= fn.End() {
			found = fn
			return false
		}
		return true
	})
	return found
}

func methodSetHas(namedType *types.Named, methodName string) bool {
	pointerType := types.NewPointer(namedType)
	ms := types.NewMethodSet(pointerType)
	for i := 0; i < ms.Len(); i++ {
		if ms.At(i).Obj().Name() == methodName {
			return true
		}
	}
	return false
}

// findNextStmt returns the statement immediately following stmt in the file,
// or nil if there is none. It walks the enclosing block to find the statement
// whose Pos() is just after stmt.End().
func findNextStmt(file *ast.File, stmt ast.Node) ast.Stmt {
	var next ast.Stmt
	ast.Inspect(file, func(node ast.Node) bool {
		if next != nil {
			return false
		}
		bl, ok := node.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, s := range bl.List {
			if s == stmt || (s.Pos() == stmt.Pos() && s.End() == stmt.End()) {
				if i+1 < len(bl.List) {
					next = bl.List[i+1]
			}
				return false
			}
		}
		return true
	})
	return next
}

// isOrphanErrCheck reports whether stmt is an `if err != nil { return ... }`
// block that would be orphaned by deleting the assignment that declared err.
func isOrphanErrCheck(stmt ast.Stmt, assign *ast.AssignStmt) bool {
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok {
		return false
	}
	// Condition must be `err != nil`.
	bin, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}
	ident, ok := bin.X.(*ast.Ident)
	if !ok || ident.Name != "err" {
		return false
	}
	// The body must be a return statement (so deleting it is safe).
	bl := ifStmt.Body
	if bl == nil || len(bl.List) == 0 {
		return false
	}
	_, isReturn := bl.List[0].(*ast.ReturnStmt)
	return isReturn
}

// deletionFix builds a SuggestedFix that deletes the redundant constructor
// assignment and removes the variable from the WithSections argument list.
//
// The fix is conservative: it only fires when the assignment is a standalone
// statement (so deleting it is safe) and the variable is a direct argument to
// WithSections (so removing it is a simple positional edit).
func deletionFix(pass *analysis.Pass, file *ast.File, call *ast.CallExpr) *analysis.SuggestedFix {
	assign := findEnclosingAssignStmt(file, call.Pos())
	if assign == nil {
		return nil
	}
	ident, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return nil
	}
	fn := findEnclosingFuncDecl(file, call.Pos())
	if fn == nil {
		return nil
	}
	withSections, count := findWithSectionsUsingIdentInFunc(fn, ident.Name)
	if withSections == nil || count != 1 {
		return nil
	}

	edits := []analysis.TextEdit{}

	// Determine the end of the deletion. If the statement immediately following
	// the assignment is an `if err != nil { return ... }` block that references
	// the err from this assignment, include it in the deletion so we do not
	// leave an orphaned error check.
	deleteEnd := assign.End()
	if next := findNextStmt(file, assign); next != nil {
		if isOrphanErrCheck(next, assign) {
			deleteEnd = next.End()
		}
	}

	// Delete the assignment statement (and the trailing error check if present).
	edits = append(edits, analysis.TextEdit{
		Pos:     assign.Pos(),
		End:     deleteEnd,
		NewText: []byte{},
	})

	// Remove the ident argument from WithSections. We delete the ident plus any
	// trailing comma, or a leading comma if it is the last argument.
	for i, arg := range withSections.Args {
		if id, ok := arg.(*ast.Ident); ok && id.Name == ident.Name {
			if i < len(withSections.Args)-1 {
				// delete ident and the following comma+space
				edits = append(edits, analysis.TextEdit{
					Pos:     arg.Pos(),
					End:     withSections.Args[i+1].Pos(),
					NewText: []byte{},
				})
			} else if i > 0 {
				// delete the leading comma+space and the ident
				edits = append(edits, analysis.TextEdit{
					Pos:     withSections.Args[i-1].End(),
					End:     arg.End(),
					NewText: []byte{},
				})
			} else {
				// sole argument: just delete the ident
				edits = append(edits, analysis.TextEdit{
					Pos:     arg.Pos(),
					End:     arg.End(),
					NewText: []byte{},
				})
			}
			break
		}
	}

	return &analysis.SuggestedFix{
		Message:   "delete redundant GlazeCommand section (auto-injected by BuildCobraCommand)",
		TextEdits: edits,
	}
}
