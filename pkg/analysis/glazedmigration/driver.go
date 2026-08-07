package glazedmigration

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// fileEdit is one suggested-fix text edit resolved to byte offsets at scan
// time, so fix application needs no token.FileSet.
type fileEdit struct {
	start   int
	end     int
	newText []byte
}

// Diagnostic is one analyzer finding rendered in a form suitable for
// structured output and fix application.
type Diagnostic struct {
	// File is the absolute path of the Go source file.
	File string
	// Line and Column are 1-based source positions of the finding.
	Line   int
	Column int
	// Message is the analyzer's human-readable diagnostic text.
	Message string
	// FixCount is the number of suggested-fix text edits attached to the
	// diagnostic.
	FixCount int

	edits []fileEdit
}

// Scan parses every Go source file reachable from paths and runs the
// migration analyzer over each file. paths entries may be directories
// (walked recursively, skipping hidden directories, vendor, node_modules,
// and testdata) or individual .go files. Parsing uses go/parser directly —
// the analyzer is designed to run despite type errors, because consuming
// code that references removed Glazed APIs does not type-check. With no
// type-checker information, the rules fall back to import-aware AST
// matching, which is the same degraded mode the analyzer's own unit tests
// exercise.
func Scan(paths []string) ([]Diagnostic, error) {
	files, err := collectGoFiles(paths)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no Go files found under %s", strings.Join(paths, ", "))
	}

	var diagnostics []Diagnostic
	for _, path := range files {
		fileDiags, err := scanFile(path)
		if err != nil {
			return nil, err
		}
		diagnostics = append(diagnostics, fileDiags...)
	}
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].File != diagnostics[j].File {
			return diagnostics[i].File < diagnostics[j].File
		}
		return diagnostics[i].Line < diagnostics[j].Line
	})
	return diagnostics, nil
}

// collectGoFiles expands paths into a sorted list of .go files.
func collectGoFiles(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("cannot stat %s: %w", p, err)
		}
		if !info.IsDir() {
			if strings.HasSuffix(p, ".go") {
				abs, err := filepath.Abs(p)
				if err != nil {
					return nil, err
				}
				files = append(files, abs)
			}
			continue
		}
		root := p
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				name := d.Name()
				if path != root && (strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == "testdata") {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".go") {
				abs, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				files = append(files, abs)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("cannot walk %s: %w", p, err)
		}
	}
	sort.Strings(files)
	return files, nil
}

// scanFile parses one file and runs the analyzer over it, resolving all
// suggested-fix text edits to byte offsets while the FileSet is available.
func scanFile(path string) ([]Diagnostic, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}

	var diagnostics []Diagnostic
	pass := &analysis.Pass{
		Analyzer: Analyzer,
		Fset:     fset,
		Files:    []*ast.File{file},
		TypesInfo: &types.Info{
			Uses:   map[*ast.Ident]types.Object{},
			Defs:   map[*ast.Ident]types.Object{},
			Scopes: map[ast.Node]*types.Scope{},
		},
		Report: func(d analysis.Diagnostic) {
			pos := fset.Position(d.Pos)
			var edits []fileEdit
			for _, fix := range d.SuggestedFixes {
				for _, textEdit := range fix.TextEdits {
					start := fset.PositionFor(textEdit.Pos, false)
					end := fset.PositionFor(textEdit.End, false)
					edits = append(edits, fileEdit{
						start:   start.Offset,
						end:     end.Offset,
						newText: textEdit.NewText,
					})
				}
			}
			diagnostics = append(diagnostics, Diagnostic{
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Message:  d.Message,
				FixCount: len(edits),
				edits:    edits,
			})
		},
	}
	if _, err := Analyzer.Run(pass); err != nil {
		return nil, fmt.Errorf("analyzer failed on %s: %w", path, err)
	}
	return diagnostics, nil
}

// writeMigratedFile writes migrated content back to path. path always comes
// from collectGoFiles (cleaned, suffixed .go, discovered under the
// user-supplied scan roots) — editing those files in place is the entire
// purpose of the tool, so the G703 file-write-traversal warning does not
// apply here.
func writeMigratedFile(path string, content []byte, mode os.FileMode) error {
	cleaned := filepath.Clean(path)
	if !strings.HasSuffix(cleaned, ".go") {
		return fmt.Errorf("refusing to write non-Go file %s", cleaned)
	}
	return os.WriteFile(cleaned, content, mode) // #nosec G703 -- paths are the scan roots' .go files by construction
}

// ApplyFixes applies every suggested fix carried by diagnostics, grouped per
// file. Edits within a file are applied from the end of the file backwards so
// offsets stay valid. Overlapping or out-of-range edits are skipped (counted,
// not fatal); a file whose edits all conflict is left untouched. Returns the
// number of applied edits per modified file and the total number of skipped
// edits.
func ApplyFixes(diagnostics []Diagnostic) (map[string]int, int, error) {
	skipped := 0
	appliedPerFile := map[string]int{}
	byFile := map[string][]fileEdit{}
	for _, d := range diagnostics {
		byFile[d.File] = append(byFile[d.File], d.edits...)
	}

	files := make([]string, 0, len(byFile))
	for path := range byFile {
		files = append(files, path)
	}
	sort.Strings(files)

	for _, path := range files {
		edits := byFile[path]
		if len(edits) == 0 {
			continue
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return appliedPerFile, skipped, fmt.Errorf("cannot read %s: %w", path, readErr)
		} // Apply from the back so earlier offsets remain valid. lastStart is the
		// start of the nearest already-applied edit behind the current one; an
		// edit ending past it overlaps and is skipped.
		sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
		lastStart := len(content) + 1
		fileApplied := 0
		for _, e := range edits {
			if e.start < 0 || e.end < e.start || e.end > len(content) || e.end > lastStart {
				skipped++
				continue
			}
			replacement := append([]byte{}, e.newText...)
			content = append(content[:e.start], append(replacement, content[e.end:]...)...)
			lastStart = e.start
			fileApplied++
		}
		if fileApplied == 0 {
			continue
		}
		info, statErr := os.Stat(path)
		if statErr != nil {
			return appliedPerFile, skipped, fmt.Errorf("cannot stat %s: %w", path, statErr)
		}
		if writeErr := writeMigratedFile(path, content, info.Mode()); writeErr != nil {
			return appliedPerFile, skipped, fmt.Errorf("cannot write %s: %w", path, writeErr)
		}
		appliedPerFile[path] = fileApplied
	}
	return appliedPerFile, skipped, nil
}
