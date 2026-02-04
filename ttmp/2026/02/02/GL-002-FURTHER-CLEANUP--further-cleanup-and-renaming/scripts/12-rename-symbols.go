package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v3"
)

type Mapping struct {
	Pkg string `yaml:"pkg"`
	Old string `yaml:"old"`
	New string `yaml:"new"`
}

type MappingConfig struct {
	Mappings []Mapping `yaml:"mappings"`
}

type renameTarget struct {
	newName string
}

func loadMappings(path string) (map[string]map[string]renameTarget, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg MappingConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	result := make(map[string]map[string]renameTarget)
	for _, m := range cfg.Mappings {
		if m.Pkg == "" || m.Old == "" || m.New == "" {
			return nil, fmt.Errorf("invalid mapping: pkg/old/new required (got pkg=%q old=%q new=%q)", m.Pkg, m.Old, m.New)
		}
		byPkg, ok := result[m.Pkg]
		if !ok {
			byPkg = make(map[string]renameTarget)
			result[m.Pkg] = byPkg
		}
		byPkg[m.Old] = renameTarget{newName: m.New}
	}
	return result, nil
}

func main() {
	var (
		root         string
		mapping      string
		write        bool
		verbose      bool
		patterns     string
		tests        bool
		ignoreErrors bool
	)
	flag.StringVar(&root, "root", ".", "module root to scan")
	flag.StringVar(&mapping, "mapping", "", "path to YAML mapping file")
	flag.BoolVar(&write, "write", false, "write changes to disk")
	flag.BoolVar(&verbose, "verbose", false, "print per-file changes")
	flag.StringVar(&patterns, "patterns", "./...", "comma-separated go/packages patterns")
	flag.BoolVar(&tests, "tests", false, "include _test.go packages")
	flag.BoolVar(&ignoreErrors, "ignore-errors", false, "continue even if packages.Load reports errors")
	flag.Parse()

	if mapping == "" {
		fmt.Fprintln(os.Stderr, "--mapping is required")
		os.Exit(2)
	}

	mappingPath, err := filepath.Abs(mapping)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mappingConfig, err := loadMappings(mappingPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	patternList := []string{"./..."}
	if patterns != "" {
		patternList = []string{}
		for _, p := range splitComma(patterns) {
			if p != "" {
				patternList = append(patternList, p)
			}
		}
	}

	fset := token.NewFileSet()
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule,
		Dir:   absRoot,
		Fset:  fset,
		Tests: tests,
	}

	pkgs, err := packages.Load(cfg, patternList...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 && !ignoreErrors {
		os.Exit(1)
	}

	fileChanges := map[string]int{}
	fileTouched := map[string]bool{}

	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}
		for ident, obj := range pkg.TypesInfo.Defs {
			renameIdent(ident, obj, mappingConfig, fset, fileChanges, fileTouched)
		}
		for ident, obj := range pkg.TypesInfo.Uses {
			renameIdent(ident, obj, mappingConfig, fset, fileChanges, fileTouched)
		}
	}

	if len(fileChanges) == 0 {
		fmt.Println("no matches")
		return
	}

	if !write {
		printSummary(fileChanges, false)
		return
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			filename := fset.Position(file.Pos()).Filename
			if filename == "" || !fileTouched[filename] {
				continue
			}

			var buf bytes.Buffer
			if err := format.Node(&buf, fset, file); err != nil {
				fmt.Fprintf(os.Stderr, "format %s: %v\n", filename, err)
				os.Exit(1)
			}
			if err := os.WriteFile(filename, buf.Bytes(), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", filename, err)
				os.Exit(1)
			}
			if verbose {
				fmt.Printf("updated %s\n", filename)
			}
		}
	}

	printSummary(fileChanges, true)
}

func renameIdent(
	ident *ast.Ident,
	obj types.Object,
	mappingConfig map[string]map[string]renameTarget,
	fset *token.FileSet,
	fileChanges map[string]int,
	fileTouched map[string]bool,
) {
	if ident == nil || obj == nil {
		return
	}
	pkg := obj.Pkg()
	if pkg == nil {
		return
	}
	if _, ok := obj.(*types.PkgName); ok {
		return
	}
	pkgMappings, ok := mappingConfig[pkg.Path()]
	if !ok {
		return
	}
	target, ok := pkgMappings[obj.Name()]
	if !ok {
		return
	}
	if ident.Name == target.newName {
		return
	}
	ident.Name = target.newName
	filename := fset.Position(ident.Pos()).Filename
	fileChanges[filename]++
	fileTouched[filename] = true
}

func splitComma(input string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(input); i++ {
		if input[i] == ',' {
			parts = append(parts, input[start:i])
			start = i + 1
		}
	}
	parts = append(parts, input[start:])
	return parts
}

func printSummary(fileChanges map[string]int, wrote bool) {
	files := make([]string, 0, len(fileChanges))
	for f := range fileChanges {
		files = append(files, f)
	}
	sort.Strings(files)
	if wrote {
		fmt.Println("updated files:")
	} else {
		fmt.Println("planned changes:")
	}
	for _, f := range files {
		fmt.Printf("- %s (%d)\n", f, fileChanges[f])
	}
}
