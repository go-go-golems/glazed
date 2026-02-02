// rename_glazed_api.go - AST-based migration tool for glazed API renames.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type identTarget struct {
	NewPkgPath string
	NewIdent   string
}

type fileReport struct {
	Path           string            `json:"path"`
	ImportsChanged map[string]string `json:"importsChanged,omitempty"`
	IdentsChanged  map[string]string `json:"identsChanged,omitempty"`
	Warnings       []string          `json:"warnings,omitempty"`
	Errors         []string          `json:"errors,omitempty"`
	Skipped        bool              `json:"skipped"`
}

type report struct {
	Root   string       `json:"root"`
	Files  []fileReport `json:"files"`
	Errors []string     `json:"errors,omitempty"`
}

var (
	flagRoot   = flag.String("root", "", "repo root")
	flagWrite  = flag.Bool("write", false, "write changes to files")
	flagReport = flag.String("report", "", "write JSON report to path")
)

var oldToNewPkg = map[string]string{
	"github.com/go-go-golems/glazed/pkg/cmds/layers":                    "github.com/go-go-golems/glazed/pkg/cmds/schema",
	"github.com/go-go-golems/glazed/pkg/cmds/parameters":                "github.com/go-go-golems/glazed/pkg/cmds/fields",
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares":               "github.com/go-go-golems/glazed/pkg/cmds/sources",
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper": "github.com/go-go-golems/glazed/pkg/cmds/sources/patternmapper",
}

var pkgAliasByPath = map[string]string{
	"github.com/go-go-golems/glazed/pkg/cmds/schema":  "schema",
	"github.com/go-go-golems/glazed/pkg/cmds/fields":  "fields",
	"github.com/go-go-golems/glazed/pkg/cmds/values":  "values",
	"github.com/go-go-golems/glazed/pkg/cmds/sources": "sources",
}

var identMap = map[string]map[string]identTarget{
	"github.com/go-go-golems/glazed/pkg/cmds/layers": {
		"ParameterLayer":        {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "Section"},
		"ParameterLayers":       {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "Schema"},
		"ParameterLayerImpl":    {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "SectionImpl"},
		"ParameterLayerOptions": {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "SectionOption"},
		"ParameterLayersOption": {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "SchemaOption"},
		"NewParameterLayer":     {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "NewSection"},
		"NewParameterLayers":    {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "NewSchema"},
		"WithLayers":            {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "WithSections"},
		"WithDefinitions":       {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/layers"], NewIdent: "WithFields"},
		"SectionValues":         {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/values", NewIdent: "SectionValues"},
		"Values":                {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/values", NewIdent: "Values"},
		"SectionValuesOption":   {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/values", NewIdent: "SectionValuesOption"},
		"ValuesOption":          {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/values", NewIdent: "ValuesOption"},
		"NewSectionValues":      {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/values", NewIdent: "NewSectionValues"},
		"NewValues":             {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/values", NewIdent: "New"},
		"WithParameters":        {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/values", NewIdent: "WithParameters"},
		"WithParameterValue":    {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/values", NewIdent: "WithParameterValue"},
	},
	"github.com/go-go-golems/glazed/pkg/cmds/parameters": {
		"ParameterDefinition":       {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/parameters"], NewIdent: "Definition"},
		"ParameterDefinitions":      {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/parameters"], NewIdent: "Definitions"},
		"ParameterDefinitionOption": {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/parameters"], NewIdent: "Option"},
		"NewParameterDefinition":    {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/parameters"], NewIdent: "New"},
		"NewParameterDefinitions":   {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/parameters"], NewIdent: "NewDefinitions"},
		"WithSource":                {NewPkgPath: "github.com/go-go-golems/glazed/pkg/cmds/sources", NewIdent: "WithSource"},
	},
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares": {
		"ExecuteMiddlewares":          {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "Execute"},
		"ParseFromCobraCommand":       {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromCobra"},
		"GatherArguments":             {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromArgs"},
		"UpdateFromEnv":               {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromEnv"},
		"SetFromDefaults":             {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromDefaults"},
		"LoadParametersFromFile":      {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromFile"},
		"LoadParametersFromFiles":     {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromFiles"},
		"UpdateFromMap":               {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromMap"},
		"UpdateFromMapFirst":          {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromMapFirst"},
		"UpdateFromMapAsDefault":      {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromMapAsDefault"},
		"UpdateFromMapAsDefaultFirst": {NewPkgPath: oldToNewPkg["github.com/go-go-golems/glazed/pkg/cmds/middlewares"], NewIdent: "FromMapAsDefaultFirst"},
	},
}

func main() {
	flag.Parse()
	if *flagRoot == "" {
		fmt.Fprintln(os.Stderr, "--root is required")
		os.Exit(2)
	}

	rep := report{Root: *flagRoot}

	_ = filepath.WalkDir(*flagRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			rep.Errors = append(rep.Errors, err.Error())
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		rec := processFile(path)
		rep.Files = append(rep.Files, rec)
		return nil
	})

	if *flagReport != "" {
		data, err := json.MarshalIndent(rep, "", "  ")
		if err == nil {
			_ = os.WriteFile(*flagReport, data, 0o644)
		}
	}
}

func processFile(path string) fileReport {
	fr := fileReport{Path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		fr.Errors = append(fr.Errors, err.Error())
		return fr
	}
	if bytes.HasPrefix(data, []byte("// Code generated")) {
		fr.Skipped = true
		return fr
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
	if err != nil {
		fr.Errors = append(fr.Errors, err.Error())
		return fr
	}

	importsChanged := map[string]string{}
	identsChanged := map[string]string{}

	// Map current import aliases to paths
	aliasToPath := map[string]string{}
	pathToAlias := map[string]string{}
	for _, imp := range file.Imports {
		pathValue := strings.Trim(imp.Path.Value, "\"")
		name := importName(imp, pathValue)
		aliasToPath[name] = pathValue
		pathToAlias[pathValue] = name
	}

	// Rewrite imports from old to new paths
	for _, imp := range file.Imports {
		pathValue := strings.Trim(imp.Path.Value, "\"")
		newPath, ok := oldToNewPkg[pathValue]
		if !ok {
			continue
		}
		oldAlias := importName(imp, pathValue)
		newAlias := pkgAliasByPath[newPath]
		if oldAlias == pathBase(pathValue) {
			if newAlias != oldAlias {
				imp.Name = ast.NewIdent(newAlias)
			}
		}
		imp.Path.Value = fmt.Sprintf("\"%s\"", newPath)
		importsChanged[pathValue] = newPath
		aliasToPath[oldAlias] = newPath
		pathToAlias[newPath] = newAlias
	}

	// Track desired imports created by selector rewrites
	neededAliases := map[string]string{} // alias -> path

	ast.Inspect(file, func(n ast.Node) bool {
		se, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := se.X.(*ast.Ident)
		if !ok {
			return true
		}
		pkgPath, ok := aliasToPath[x.Name]
		if !ok {
			return true
		}
		mapping, hasMapping := identMap[pkgPath]
		var (
			target identTarget
			found  bool
		)
		if hasMapping {
			target, found = mapping[se.Sel.Name]
		}
		// Special handling: ParameterType* => Type*
		if !found && pkgPath == "github.com/go-go-golems/glazed/pkg/cmds/parameters" && strings.HasPrefix(se.Sel.Name, "ParameterType") {
			suffix := strings.TrimPrefix(se.Sel.Name, "ParameterType")
			target = identTarget{NewPkgPath: oldToNewPkg[pkgPath], NewIdent: "Type" + suffix}
			found = true
		}

		if found {
			newAlias := pkgAliasByPath[target.NewPkgPath]
			if newAlias == "" {
				newAlias = pathBase(target.NewPkgPath)
			}
			if x.Name != newAlias {
				x.Name = newAlias
			}
			if se.Sel.Name != target.NewIdent {
				identsChanged[se.Sel.Name] = target.NewIdent
				se.Sel.Name = target.NewIdent
			}
			neededAliases[newAlias] = target.NewPkgPath
		}

		return true
	})

	// Ensure needed imports exist
	for alias, path := range neededAliases {
		if _, ok := pathToAlias[path]; ok {
			continue
		}
		imp := &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", path)}}
		if alias != pathBase(path) {
			imp.Name = ast.NewIdent(alias)
		}
		file.Imports = append(file.Imports, imp)
		importsChanged[path] = path
		pathToAlias[path] = alias
		aliasToPath[alias] = path
	}

	// Remove unused imports
	usedAliases := collectUsedAliases(file)
	newImports := make([]*ast.ImportSpec, 0, len(file.Imports))
	for _, imp := range file.Imports {
		pathValue := strings.Trim(imp.Path.Value, "\"")
		name := importName(imp, pathValue)
		if name == "_" || name == "." || usedAliases[name] {
			newImports = append(newImports, imp)
		}
	}
	file.Imports = newImports

	if len(importsChanged) == 0 && len(identsChanged) == 0 {
		return fr
	}

	fr.ImportsChanged = importsChanged
	fr.IdentsChanged = identsChanged

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		fr.Errors = append(fr.Errors, err.Error())
		return fr
	}
	if *flagWrite {
		_ = os.WriteFile(path, buf.Bytes(), 0o644)
	}
	return fr
}

func collectUsedAliases(file *ast.File) map[string]bool {
	used := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		se, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := se.X.(*ast.Ident)
		if ok {
			used[x.Name] = true
		}
		return true
	})
	return used
}

func importName(imp *ast.ImportSpec, pathValue string) string {
	if imp.Name != nil {
		return imp.Name.Name
	}
	return pathBase(pathValue)
}

func pathBase(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
