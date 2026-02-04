//go:build tools
// +build tools

package main

import (
	"encoding/json"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type fileSymbols struct {
	Path   string   `json:"path"`
	Idents []string `json:"idents"`
}

type report struct {
	Root  string        `json:"root"`
	Files []fileSymbols `json:"files"`
}

func main() {
	root := flag.String("root", ".", "root directory")
	out := flag.String("out", "", "output json path")
	flag.Parse()

	identRe := regexp.MustCompile(`(?i)parameter|layer`)
	rep := report{Root: *root}

	_ = filepath.WalkDir(*root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			if strings.Contains(path, string(filepath.Separator)+"ttmp"+string(filepath.Separator)) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil
		}
		idents := map[string]struct{}{}
		ast.Inspect(f, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			if identRe.MatchString(id.Name) {
				idents[id.Name] = struct{}{}
			}
			return true
		})
		if len(idents) == 0 {
			return nil
		}
		list := make([]string, 0, len(idents))
		for name := range idents {
			list = append(list, name)
		}
		sort.Strings(list)
		rep.Files = append(rep.Files, fileSymbols{Path: path, Idents: list})
		return nil
	})

	sort.Slice(rep.Files, func(i, j int) bool { return rep.Files[i].Path < rep.Files[j].Path })
	data, _ := json.MarshalIndent(rep, "", "  ")
	if *out == "" {
		if _, err := os.Stdout.Write(data); err != nil {
			os.Exit(1)
		}
		return
	}
	_ = os.WriteFile(*out, data, 0o644)
}
