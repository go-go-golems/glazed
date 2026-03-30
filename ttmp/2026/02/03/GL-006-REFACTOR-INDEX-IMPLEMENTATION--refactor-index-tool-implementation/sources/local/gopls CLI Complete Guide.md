# gopls CLI Complete Guide

This document provides hands-on examples of every gopls CLI feature with real output from our demo project.

## Table of Contents

1. [Code Navigation](#code-navigation)
2. [Call Hierarchy](#call-hierarchy)
3. [Rename Operations](#rename-operations)
4. [Code Actions](#code-actions)
5. [Code Lenses](#code-lenses)
6. [Diagnostics](#diagnostics)
7. [Symbols](#symbols)
8. [Advanced Features](#advanced-features)

---

## Code Navigation

### 1. definition - Find where a symbol is defined

**Command:**
```bash
gopls definition <file>:<line>:<column>
```

**Example:**
```bash
$ gopls definition calculator/calculator.go:19:10
```

**Output:**
```
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:9:6-16: defined here as type Calculator struct {
	memory float64
}
Calculator provides basic arithmetic operations
func (c *Calculator) Add(a float64, b float64) float64
func (c *Calculator) AddWithValidation(a float64, b float64) (float64, error)
func (c *Calculator) ClearMemory()
func (c *Calculator) Divide(a float64, b float64) (float64, error)
func (c *Calculator) GetMemory() float64
func (c *Calculator) Multiply(a float64, b float64) float64
func (c *Calculator) Power(a float64, b float64) float64
func (c *Calculator) SquareRoot(a float64) (float64, error)
func (c *Calculator) Subtract(a float64, b float64) float64
```

**What it shows:**
- The exact location of the symbol's definition
- The full type signature
- All methods available on the type

**Use cases:**
- Jump to definition in scripts or automation
- Understand where a function/type is declared
- Quick documentation lookup

---

### 2. references - Find all uses of a symbol

**Command:**
```bash
gopls references <file>:<line>:<column>
```

**Example:**
```bash
$ gopls references calculator/calculator.go:19:10
```

**Output:**
```
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:14:13-23
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:15:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:19:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:26:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:33:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:40:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:50:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:57:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:67:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:72:10-20
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:85:10-20
/home/ubuntu/gopls-cli-demo/server/server.go:15:19-29
```

**What it shows:**
- Every location where the symbol is referenced
- Cross-package references (note server/server.go)
- Line and column ranges for each reference

**Use cases:**
- Find all usages before refactoring
- Understand the impact of changes
- Generate dependency graphs

---

### 3. implementation - Find implementations of an interface

**Command:**
```bash
gopls implementation <file>:<line>:<column>
```

**Use cases:**
- Find all types that implement an interface
- Discover concrete implementations
- Understand polymorphic code

---

## Call Hierarchy

### 4. call_hierarchy - Show who calls a function and what it calls

**Command:**
```bash
gopls call_hierarchy <file>:<line>:<column>
```

**Example:**
```bash
$ gopls call_hierarchy calculator/calculator.go:19:23
```

**Output:**
```
caller[0]: ranges 92:11-14 in /home/ubuntu/gopls-cli-demo/calculator/calculator.go from/to function AddWithValidation in /home/ubuntu/gopls-cli-demo/calculator/calculator.go:85:22-39
caller[1]: ranges 10:17-20 in /home/ubuntu/gopls-cli-demo/calculator/calculator_test.go from/to function TestAdd in /home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:8:6-13
caller[2]: ranges 88:7-10 in /home/ubuntu/gopls-cli-demo/calculator/calculator_test.go from/to function TestMemoryOperations in /home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:85:6-26
caller[3]: ranges 135:8-11 in /home/ubuntu/gopls-cli-demo/calculator/calculator_test.go from/to function BenchmarkAdd in /home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:132:6-18
caller[4]: ranges 45:18-21 in /home/ubuntu/gopls-cli-demo/main.go from/to function runCLI in /home/ubuntu/gopls-cli-demo/main.go:40:6-12
caller[5]: ranges 66:19-22 in /home/ubuntu/gopls-cli-demo/server/server.go from/to function handleAdd in /home/ubuntu/gopls-cli-demo/server/server.go:54:18-27
identifier: function Add in /home/ubuntu/gopls-cli-demo/calculator/calculator.go:19:22-25
```

**What it shows:**
- **Incoming calls (callers)**: All functions that call this function
- **Outgoing calls (callees)**: All functions this function calls (if any)
- Exact line numbers and ranges for each call site
- Cross-package call relationships

**Use cases:**
- Understand function dependencies
- Find dead code (functions with no callers)
- Trace execution flow
- Impact analysis before refactoring

---

## Rename Operations

### 5. prepare_rename - Check if a rename is safe

**Command:**
```bash
gopls prepare_rename <file>:<line>:<column>
```

**Example:**
```bash
$ gopls prepare_rename calculator/calculator.go:19:23
```

**Output:**
```
/home/ubuntu/gopls-cli-demo/calculator/calculator.go:19:22-25
```

**What it shows:**
- The exact range that will be renamed
- Validates that the position is a renameable identifier
- Returns an error if rename is not possible

**Use cases:**
- Validate rename before executing
- Check if a symbol can be safely renamed
- Preview the scope of a rename

---

### 6. rename - Perform a rename refactoring

**Command:**
```bash
gopls rename <file>:<line>:<column> <new-name>
```

**Flags:**
- `-w`: Write changes to files
- `-d`: Show diff instead of applying changes
- `-preserve`: Make backup copies before modifying

**Example (dry run):**
```bash
$ gopls rename -d calculator/calculator.go:19:23 Sum
```

**What it does:**
- Renames the symbol at the specified position
- Updates all references across all files
- Preserves code correctness (type-safe)
- Can show diff or apply changes directly

**Use cases:**
- Safe refactoring across entire codebase
- Rename functions, variables, types, packages
- Automated code modernization

---

## Code Actions

### 7. codeaction - List or execute quick fixes and refactorings

**Command:**
```bash
gopls codeaction <file>:<line>:<column>
```

**Flags:**
- `-kind`: Filter by action kind (e.g., `refactor.extract`, `quickfix`)
- `-title`: Filter by title (regex)
- `-exec`: Execute the first matching action

**Example:**
```bash
$ gopls codeaction calculator/calculator.go:1:1
```

**Output:**
```
command	"Browse documentation for package calculator" [source.doc]
command	"Split package \"calculator\"" [source.splitPackage]
command	"Show compiler optimization details for \"calculator\"" [source.toggleCompilerOptDetails]
command	"Browse gopls feature documentation" [gopls.doc.features]
```

**Available Code Action Kinds:**
- `source.organizeImports` - Sort and remove unused imports
- `refactor.extract.function` - Extract selected code into a function
- `refactor.extract.variable` - Extract expression into a variable
- `refactor.inline` - Inline function or variable
- `quickfix` - Fix diagnostics (errors/warnings)
- `source.doc` - Browse documentation
- `source.splitPackage` - Split package into multiple files

**Use cases:**
- Automated code cleanup
- Extract complex code into functions
- Organize imports
- Apply quick fixes to errors

---

## Code Lenses

### 8. codelens - List or execute inline commands

**Command:**
```bash
gopls codelens <file>
```

**Flags:**
- `-exec`: Execute a code lens by title

**Example:**
```bash
$ gopls codelens calculator/calculator_test.go
```

**Output:**
```
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go: "run file benchmarks" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:8: "run test" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:19: "run test" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:27: "run test" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:35: "run test" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:56: "run test" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:64: "run test" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:85: "run test" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:104: "run test" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:132: "run benchmark" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:139: "run benchmark" [gopls.run_tests]
/home/ubuntu/gopls-cli-demo/calculator/calculator_test.go:146: "run benchmark" [gopls.run_tests]
```

**What it shows:**
- Inline commands available in the file
- Test/benchmark runners
- Documentation links
- Regenerate commands

**Execute a specific test:**
```bash
$ gopls codelens -exec calculator/calculator_test.go:8 "run test"
```

**Use cases:**
- Run individual tests from command line
- Execute benchmarks
- Trigger code generation
- Quick access to documentation

---

## Diagnostics

### 9. check - Show errors and warnings

**Command:**
```bash
gopls check <file>
```

**Example:**
```bash
$ gopls check calculator/calculator.go
```

**Output:**
```
(no output means no errors)
```

**What it shows:**
- Compilation errors
- Type errors
- Lint warnings
- Unused variables/imports

**Use cases:**
- CI/CD validation
- Pre-commit checks
- Quick syntax validation

---

## Symbols

### 10. symbols - List all symbols in a file

**Command:**
```bash
gopls symbols <file>
```

**What it shows:**
- All functions, types, variables, constants
- Symbol kind (function, struct, method, etc.)
- Location of each symbol

**Use cases:**
- Generate documentation
- Code navigation
- Symbol search

---

### 11. workspace_symbol - Search symbols across workspace

**Command:**
```bash
gopls workspace_symbol <query>
```

**What it does:**
- Searches all symbols in the workspace
- Fuzzy matching
- Returns locations

**Use cases:**
- Find symbols by name
- Workspace-wide search
- Code exploration

---

## Advanced Features

### 12. format - Format code according to gofmt

**Command:**
```bash
gopls format <file>
```

**Flags:**
- `-w`: Write formatted code back to file
- `-d`: Show diff

**Use cases:**
- Automated code formatting
- CI/CD formatting checks

---

### 13. imports - Organize imports

**Command:**
```bash
gopls imports <file>
```

**What it does:**
- Adds missing imports
- Removes unused imports
- Sorts imports

**Use cases:**
- Automated import management
- Code cleanup

---

### 14. signature - Show function signature help

**Command:**
```bash
gopls signature <file>:<line>:<column>
```

**What it shows:**
- Function signature at cursor
- Parameter information
- Return types

**Use cases:**
- Quick function documentation
- Parameter hints

---

### 15. highlight - Show related identifiers

**Command:**
```bash
gopls highlight <file>:<line>:<column>
```

**What it shows:**
- All occurrences of the identifier in the file
- Read/write references

**Use cases:**
- Visual code navigation
- Find local usages

---

### 16. folding_ranges - Show code folding regions

**Command:**
```bash
gopls folding_ranges <file>
```

**What it shows:**
- Regions that can be folded (functions, structs, imports)
- Start and end positions

**Use cases:**
- Editor integration
- Code structure analysis

---

### 17. links - Show links in a file

**Command:**
```bash
gopls links <file>
```

**What it shows:**
- Import paths
- Documentation links
- URLs in comments

**Use cases:**
- Extract dependencies
- Find documentation references

---

## Position Syntax

gopls accepts positions in multiple formats:

- `file.go:line:col` - Line and column (1-indexed)
- `file.go:line` - Just line number
- `file.go:#offset` - Byte offset (0-indexed)
- `file.go:line:col-line2:col2` - Range

**Examples:**
```bash
gopls definition calculator/calculator.go:19:23
gopls definition calculator/calculator.go:19
gopls definition calculator/calculator.go:#450
gopls definition calculator/calculator.go:19:1-19:25
```

---

## Tips and Tricks

### 1. Combine with other tools

```bash
# Find all callers and count them
gopls call_hierarchy calculator/calculator.go:19:23 | grep "caller\[" | wc -l

# Extract just the file paths from references
gopls references calculator/calculator.go:19:10 | cut -d: -f1 | sort -u

# Check multiple files
find . -name "*.go" -exec gopls check {} \;
```

### 2. Use in scripts

```bash
#!/bin/bash
# Check if a function has any callers
CALLERS=$(gopls call_hierarchy "$1" | grep -c "caller\[")
if [ "$CALLERS" -eq 0 ]; then
    echo "Warning: Function has no callers (dead code?)"
fi
```

### 3. Integration with editors

Most editors use gopls as a language server, but you can also use the CLI:

```bash
# Get definition and open in editor
FILE=$(gopls definition calculator/calculator.go:19:10 | head -1 | cut -d: -f1)
vim "$FILE"
```

---

## Common Workflows

### Workflow 1: Safe Refactoring

```bash
# 1. Check if rename is safe
gopls prepare_rename calculator/calculator.go:19:23

# 2. Preview the changes
gopls rename -d calculator/calculator.go:19:23 Sum

# 3. Apply the rename
gopls rename -w calculator/calculator.go:19:23 Sum

# 4. Verify no errors
gopls check calculator/calculator.go
```

### Workflow 2: Understanding Code

```bash
# 1. Find definition
gopls definition main.go:45:18

# 2. See all references
gopls references calculator/calculator.go:19:23

# 3. Check call hierarchy
gopls call_hierarchy calculator/calculator.go:19:23

# 4. Look at implementations
gopls implementation calculator/calculator.go:9:10
```

### Workflow 3: Code Cleanup

```bash
# 1. Organize imports
gopls imports -w calculator/calculator.go

# 2. Format code
gopls format -w calculator/calculator.go

# 3. Check for errors
gopls check calculator/calculator.go

# 4. Apply code actions
gopls codeaction -kind source.organizeImports -exec calculator/calculator.go:1:1
```

---

## Conclusion

The gopls CLI provides powerful tools for:
- **Code navigation** (definition, references, implementation)
- **Refactoring** (rename, extract, inline)
- **Analysis** (call hierarchy, diagnostics)
- **Automation** (code actions, formatting, imports)

All these features work across your entire workspace and understand Go's type system, making them safe and reliable for production use.
