# Deep Dive Analysis: gopls CLI Tools Architecture and Implementation

**Author**: Manus AI  
**Date**: February 3, 2026

---

## Executive Summary

This document provides an exhaustive analysis of the `gopls` command-line interface (CLI) tools, focusing on the architectural patterns, implementation details, and underlying libraries that power commands such as `prepare_rename`, `call_hierarchy`, `codeaction`, and `codelens`. The analysis is designed to enable developers to build custom tooling that either invokes the `gopls` CLI directly or leverages the same packages and libraries for deeper integration.

The `gopls` project is the official Go language server, implementing the Language Server Protocol (LSP) to provide IDE-like features for Go development. While primarily designed as a server that editors connect to, `gopls` also exposes a rich CLI that allows direct invocation of LSP features from the command line. This dual nature makes it an excellent foundation for building custom Go development tools.

---

## Table of Contents

1. [Introduction and Background](#1-introduction-and-background)
2. [Project Structure and Organization](#2-project-structure-and-organization)
3. [CLI Architecture and Command Framework](#3-cli-architecture-and-command-framework)
4. [Position and Span Handling](#4-position-and-span-handling)
5. [Connection and Server Initialization](#5-connection-and-server-initialization)
6. [Deep Dive: prepare_rename Command](#6-deep-dive-prepare_rename-command)
7. [Deep Dive: call_hierarchy Command](#7-deep-dive-call_hierarchy-command)
8. [Deep Dive: codeaction Command](#8-deep-dive-codeaction-command)
9. [Deep Dive: codelens Command](#9-deep-dive-codelens-command)
10. [Server-Side Implementation](#10-server-side-implementation)
11. [Language-Specific Logic (golang Package)](#11-language-specific-logic-golang-package)
12. [Protocol Layer and LSP Types](#12-protocol-layer-and-lsp-types)
13. [Cache and Session Management](#13-cache-and-session-management)
14. [Building Custom Tooling: CLI Invocation](#14-building-custom-tooling-cli-invocation)
15. [Building Custom Tooling: Library Usage](#15-building-custom-tooling-library-usage)
16. [Best Practices and Recommendations](#16-best-practices-and-recommendations)
17. [Conclusion](#17-conclusion)

---

## 1. Introduction and Background

The Go programming language has seen tremendous growth in adoption, and with it, the demand for sophisticated development tools has increased. The `gopls` project emerged as the official Go language server, providing a standardized way for editors and IDEs to offer features like code completion, navigation, refactoring, and diagnostics.

The Language Server Protocol (LSP), developed by Microsoft, defines a standard protocol for communication between development tools and language servers. By implementing LSP, `gopls` can work with any editor that supports the protocol, including Visual Studio Code, Vim, Emacs, and many others.

While the primary use case for `gopls` is as a long-running server process that editors communicate with over JSON-RPC, the project also provides a command-line interface that exposes many of these features. This CLI is valuable for:

- **Automation**: Integrating Go language features into build scripts, CI/CD pipelines, and other automated workflows.
- **Testing**: Validating code transformations and refactorings in a scriptable manner.
- **Custom Tooling**: Building specialized tools that leverage Go's type system and semantic analysis.
- **Debugging**: Investigating how `gopls` interprets code and what information it can provide.

This document focuses on understanding how the CLI works, both from an invocation perspective and from the standpoint of the underlying implementation. By understanding these internals, developers can make informed decisions about whether to use the CLI directly or to integrate with the `gopls` libraries.

---

## 2. Project Structure and Organization

The `gopls` project is part of the larger `golang.org/x/tools` repository. The relevant directory structure is as follows:

```
tools/
├── gopls/
│   ├── main.go                    # Entry point for the gopls binary
│   ├── internal/
│   │   ├── cmd/                   # CLI command implementations
│   │   │   ├── cmd.go            # Main application and command dispatch
│   │   │   ├── prepare_rename.go # prepare_rename command
│   │   │   ├── call_hierarchy.go # call_hierarchy command
│   │   │   ├── codeaction.go     # codeaction command
│   │   │   ├── codelens.go       # codelens command
│   │   │   ├── definition.go     # definition command
│   │   │   ├── references.go     # references command
│   │   │   ├── rename.go         # rename command
│   │   │   ├── parsespan.go      # Position parsing utilities
│   │   │   ├── span.go           # Span type definition
│   │   │   └── ...               # Other commands
│   │   ├── server/               # LSP server implementation
│   │   │   ├── call_hierarchy.go
│   │   │   ├── code_action.go
│   │   │   ├── code_lens.go
│   │   │   ├── general.go
│   │   │   └── ...
│   │   ├── golang/               # Go-specific language logic
│   │   │   ├── call_hierarchy.go
│   │   │   ├── code_lens.go
│   │   │   ├── rename.go
│   │   │   └── ...
│   │   ├── protocol/             # LSP protocol types and interfaces
│   │   │   ├── protocol.go
│   │   │   ├── tsprotocol.go    # Generated LSP types
│   │   │   ├── tsserver.go      # Generated server interface
│   │   │   ├── tsclient.go      # Generated client interface
│   │   │   ├── mapper.go        # UTF-8/UTF-16 conversion
│   │   │   └── ...
│   │   ├── cache/                # Caching and session management
│   │   │   ├── session.go
│   │   │   ├── snapshot.go
│   │   │   ├── check.go
│   │   │   └── ...
│   │   ├── lsprpc/               # RPC connection management
│   │   │   ├── lsprpc.go
│   │   │   └── ...
│   │   └── ...
│   └── ...
└── ...
```

The organization follows a clear separation of concerns:

- **`cmd/`**: Contains the CLI command implementations. Each command is typically in its own file.
- **`server/`**: Implements the LSP server interface, handling incoming requests and delegating to language-specific logic.
- **`golang/`**: Contains Go-specific implementations of language features like call hierarchy, code actions, and refactorings.
- **`protocol/`**: Defines the LSP protocol types and provides utilities for working with them.
- **`cache/`**: Manages sessions, snapshots, and caching of parsed files and type information.
- **`lsprpc/`**: Handles the JSON-RPC communication layer.

This modular structure makes it relatively straightforward to understand how a CLI command flows through the system, from parsing command-line arguments to invoking server methods to executing language-specific logic.

---

## 3. CLI Architecture and Command Framework

The `gopls` CLI is built on top of the `golang.org/x/tools/internal/tool` package, which provides a lightweight framework for creating command-line applications with subcommands. This framework is similar in spirit to tools like `git` or `go`, where a single binary provides multiple subcommands.

### 3.1. Main Entry Point

The entry point for `gopls` is in `gopls/main.go`:

```go
package main

import (
	"context"
	"log"
	"os"

	"golang.org/x/telemetry"
	"golang.org/x/telemetry/counter"
	"golang.org/x/tools/gopls/internal/cmd"
	"golang.org/x/tools/gopls/internal/filecache"
	versionpkg "golang.org/x/tools/gopls/internal/version"
	"golang.org/x/tools/internal/tool"
)

var version = "" // if set by the linker, overrides the gopls version

func main() {
	versionpkg.VersionOverride = version

	telemetry.Start(telemetry.Config{
		ReportCrashes: true,
		Upload:        true,
	})

	// Force early creation of the filecache and refuse to start
	// if there were unexpected errors such as ENOSPC.
	if _, err := filecache.Get("nonesuch", [32]byte{}); err != nil && err != filecache.ErrNotFound {
		counter.Inc("gopls/nocache")
		log.Fatalf("gopls cannot access its persistent index (disk full?): %v", err)
	}

	ctx := context.Background()
	tool.Main(ctx, cmd.New(), os.Args[1:])
}
```

The `main` function performs some initialization (telemetry, file cache) and then calls `tool.Main`, passing it a new `Application` instance created by `cmd.New()` and the command-line arguments.

### 3.2. Application Structure

The `Application` struct is defined in `gopls/internal/cmd/cmd.go`:

```go
type Application struct {
	// Core application flags
	tool.Profile

	// We include the server configuration directly for now
	Serve Serve

	// the options configuring function to invoke when building a server
	options func(*settings.Options)

	// Support for remote LSP server.
	Remote string `flag:"remote" help:"forward all commands to a remote lsp..."`

	// Verbose enables verbose logging.
	Verbose bool `flag:"v,verbose" help:"verbose output"`

	// VeryVerbose enables a higher level of verbosity in logging output.
	VeryVerbose bool `flag:"vv,veryverbose" help:"very verbose output"`

	// OTel specifies the OpenTelemetry collector endpoint
	OTel string `flag:"otel" help:"export telemetry to specified OpenTelemetry collector address"`

	// PrepareOptions is called to update the options when a new view is built.
	PrepareOptions func(*settings.Options)

	// editFlags holds flags that control how file edit operations are applied
	editFlags *EditFlags
}

type EditFlags struct {
	Write    bool `flag:"w,write" help:"write edited content to source files"`
	Preserve bool `flag:"preserve" help:"with -write, make copies of original files"`
	Diff     bool `flag:"d,diff" help:"display diffs instead of edited file content"`
	List     bool `flag:"l,list" help:"display names of edited files"`
}
```

The `Application` struct embeds `tool.Profile` for profiling support and includes various flags that control the behavior of the CLI. The `editFlags` field is used by commands that modify files (like `rename` and `codeaction`) to control how edits are applied.

### 3.3. Command Registration

Commands are registered in the `Commands()` method:

```go
func (app *Application) Commands() []tool.Application {
	var commands []tool.Application
	commands = append(commands, app.mainCommands()...)
	commands = append(commands, app.featureCommands()...)
	commands = append(commands, app.internalCommands()...)
	return commands
}

func (app *Application) mainCommands() []tool.Application {
	return []tool.Application{
		&app.Serve,
		&version{app: app},
		&bug{app: app},
		&help{app: app},
		&apiJSON{app: app},
		&licenses{app: app},
	}
}

func (app *Application) featureCommands() []tool.Application {
	return []tool.Application{
		&callHierarchy{app: app},
		&check{app: app, Severity: "warning"},
		&codeaction{app: app},
		&codelens{app: app},
		&definition{app: app},
		&execute{app: app},
		&fix{app: app},
		&foldingRanges{app: app},
		&format{app: app},
		&headlessMCP{app: app},
		&highlight{app: app},
		&implementation{app: app},
		&imports{app: app},
		newRemote(app, ""),
		newRemote(app, "inspect"),
		&links{app: app},
		&prepareRename{app: app},
		&references{app: app},
		&rename{app: app},
		&semanticToken{app: app},
		&signature{app: app},
		&stats{app: app},
		&symbols{app: app},
		&workspaceSymbol{app: app},
	}
}
```

Commands are categorized into:

- **Main commands**: Core functionality like `serve`, `version`, and `help`.
- **Feature commands**: LSP features exposed as CLI commands.
- **Internal commands**: Commands for internal use (e.g., `vulncheck`).

### 3.4. Command Dispatch

The `Run` method handles command dispatch:

```go
func (app *Application) Run(ctx context.Context, args ...string) error {
	filecache.Start()

	ctx = debug.WithInstance(ctx, app.OTel)
	if len(args) == 0 {
		s := flag.NewFlagSet(app.Name(), flag.ExitOnError)
		return tool.Run(ctx, s, &app.Serve, args)
	}
	command, args := args[0], args[1:]
	for _, c := range app.Commands() {
		if c.Name() == command {
			s := flag.NewFlagSet(app.Name(), flag.ExitOnError)
			return tool.Run(ctx, s, c, args)
		}
	}
	return tool.CommandLineErrorf("Unknown command %v", command)
}
```

If no command is specified, it defaults to the `serve` command. Otherwise, it searches for a command with a matching name and calls `tool.Run` to execute it.

### 3.5. Command Interface

Each command implements the `tool.Application` interface:

```go
type Application interface {
	Name() string
	Usage() string
	ShortHelp() string
	DetailedHelp(f *flag.FlagSet)
	Run(ctx context.Context, args ...string) error
}
```

This interface provides a consistent way to define and execute commands.

---

## 4. Position and Span Handling

One of the key challenges in implementing CLI tools for code analysis is specifying positions within source files. The `gopls` CLI uses a flexible syntax for specifying positions, and the `parsespan.go` and `span.go` files provide the implementation.

### 4.1. Position Syntax

The CLI accepts positions in the following formats:

- `file.go`: Just the file, no specific position.
- `file.go:line`: Line number (1-indexed).
- `file.go:line:col`: Line and column (both 1-indexed).
- `file.go:#offset`: Byte offset (0-indexed).
- `file.go:line:col#offset`: Both line:col and offset.
- `file.go:line:col-line2:col2`: A range from one position to another.

### 4.2. The `span` Type

The `span` type represents a range of text within a file:

```go
type span struct {
	v _span
}

type _span struct {
	URI   protocol.DocumentURI `json:"uri"`
	Start _point               `json:"start"`
	End   _point               `json:"end"`
}

type _point struct {
	Line   int `json:"line"`   // 1-based line number
	Column int `json:"column"` // 1-based, UTF-8 codes (bytes)
	Offset int `json:"offset"` // 0-based byte offset
}
```

A `span` consists of a URI (file path) and a start and end point. Each point can have either a line/column pair, a byte offset, or both. This flexibility allows the CLI to work with different input formats and convert between them as needed.

### 4.3. Parsing Positions

The `parseSpan` function in `parsespan.go` parses the position string:

```go
func parseSpan(input string) span {
	uri := protocol.URIFromPath

	// Parse the input string to extract file, line, column, and offset
	// The implementation uses a suffix-stripping approach to handle
	// the various formats
	// ...

	return newSpan(uri(filename), start, end)
}
```

The parsing logic is somewhat complex due to the variety of supported formats, but the key idea is to work backwards from the end of the string, extracting numbers and separators (`:`, `#`, `-`) to build up the position information.

### 4.4. Converting to Protocol Types

Once a `span` is parsed, it needs to be converted to LSP protocol types. The `cmdFile` type provides methods for this:

```go
func (f *cmdFile) spanLocation(s span) (protocol.Location, error) {
	// Convert span to protocol.Location
	rng, err := f.spanRange(s)
	if err != nil {
		return protocol.Location{}, err
	}
	return protocol.Location{URI: f.uri, Range: rng}, nil
}

func (f *cmdFile) spanRange(s span) (protocol.Range, error) {
	// Convert span to protocol.Range
	// This involves using the Mapper to convert between UTF-8 and UTF-16
	// ...
}
```

The conversion requires access to the file's content because LSP uses UTF-16 code units for positions, while Go uses UTF-8. The `protocol.Mapper` type handles this conversion.

---

## 5. Connection and Server Initialization

Before executing any LSP command, the CLI needs to establish a connection to the language server. The `connect` method in `cmd.go` handles this:

```go
func (app *Application) connect(ctx context.Context) (*client, *cache.Session, error) {
	c := cache.New()
	session := cache.NewSession(ctx, c)

	options := settings.DefaultOptions(app.options)
	cli := &client{
		app:     app,
		cache:   c,
		session: session,
		files:   make(map[protocol.DocumentURI]*cmdFile),
		iwlDone: make(chan struct{}),
	}

	svr := server.New(session, cli, options)
	cli.server = svr

	// Initialize the server
	params := &protocol.ParamInitialize{}
	params.RootURI = protocol.URIFromPath(app.wd)
	// ... set other initialization parameters ...

	if _, err := cli.server.Initialize(ctx, params); err != nil {
		return nil, nil, err
	}

	if err := cli.server.Initialized(ctx, &protocol.InitializedParams{}); err != nil {
		return nil, nil, err
	}

	return cli, session, nil
}
```

This method:

1. Creates a new `Cache` and `Session`.
2. Creates a `client` instance that implements the LSP client interface.
3. Creates a `server.Server` instance.
4. Sends the `initialize` and `initialized` requests to the server.

The key insight here is that the CLI creates an in-process LSP server rather than connecting to a remote server. This is more efficient for one-off commands and simplifies the implementation.

### 5.1. The `client` Type

The `client` type implements the LSP client interface:

```go
type client struct {
	app     *Application
	cache   *cache.Cache
	session *cache.Session
	server  protocol.Server

	filesMu sync.Mutex
	files   map[protocol.DocumentURI]*cmdFile

	diagnosticsMu sync.Mutex
	diagnostics   map[protocol.DocumentURI][]*protocol.Diagnostic

	progressMu sync.Mutex
	iwlToken   protocol.ProgressToken
	iwlDone    chan struct{}
}
```

The client maintains a map of open files and their diagnostics. It also implements various LSP client methods like `PublishDiagnostics`, `ApplyEdit`, and `Progress`.

### 5.2. File Handling

The `openFile` method opens a file and notifies the server:

```go
func (cli *client) openFile(ctx context.Context, uri protocol.DocumentURI) (*cmdFile, error) {
	file := cli.getFile(uri)
	if file.err != nil {
		return nil, file.err
	}

	// Choose language ID from file extension
	var langID protocol.LanguageKind
	switch filepath.Ext(uri.Path()) {
	case ".go":
		langID = "go"
	case ".mod":
		langID = "go.mod"
	// ... other cases ...
	}

	p := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: langID,
			Version:    1,
			Text:       string(file.mapper.Content),
		},
	}
	if err := cli.server.DidOpen(ctx, p); err != nil {
		file.err = fmt.Errorf("%v: %v", uri, err)
		return nil, file.err
	}
	return file, nil
}
```

This method reads the file from disk, creates a `protocol.Mapper` for UTF-8/UTF-16 conversion, and sends a `textDocument/didOpen` notification to the server.

---

## 6. Deep Dive: prepare_rename Command

The `prepare_rename` command checks whether a rename operation is valid at a given position. This is useful for validating that a rename will work before actually performing it.

### 6.1. Command Structure

The command is defined in `prepare_rename.go`:

```go
type prepareRename struct {
	app *Application
}

func (r *prepareRename) Name() string      { return "prepare_rename" }
func (r *prepareRename) Parent() string    { return r.app.Name() }
func (r *prepareRename) Usage() string     { return "<position>" }
func (r *prepareRename) ShortHelp() string { return "test validity of a rename operation at location" }
```

### 6.2. Implementation

The `Run` method implements the command logic:

```go
func (r *prepareRename) Run(ctx context.Context, args ...string) error {
	if len(args) != 1 {
		return tool.CommandLineErrorf("prepare_rename expects 1 argument (file)")
	}

	cli, _, err := r.app.connect(ctx)
	if err != nil {
		return err
	}
	defer cli.terminate(ctx)

	from := parseSpan(args[0])
	file, err := cli.openFile(ctx, from.URI())
	if err != nil {
		return err
	}
	loc, err := file.spanLocation(from)
	if err != nil {
		return err
	}
	p := protocol.PrepareRenameParams{
		TextDocumentPositionParams: protocol.LocationTextDocumentPositionParams(loc),
	}
	result, err := cli.server.PrepareRename(ctx, &p)
	if err != nil {
		return fmt.Errorf("prepare_rename failed: %w", err)
	}
	if result == nil {
		return ErrInvalidRenamePosition
	}

	s, err := file.rangeSpan(result.Range)
	if err != nil {
		return err
	}

	fmt.Println(s)
	return nil
}
```

The flow is:

1. Parse the position argument.
2. Connect to the server.
3. Open the file.
4. Convert the position to a `protocol.Location`.
5. Call `cli.server.PrepareRename`.
6. Print the result (the range that would be renamed).

### 6.3. Server-Side Implementation

On the server side, the `PrepareRename` method is implemented in `server/rename.go`:

```go
func (s *server) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.PrepareRenameResult, error) {
	ctx, done := event.Start(ctx, "server.PrepareRename")
	defer done()

	fh, snapshot, release, err := s.session.FileOf(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	defer release()

	// Delegate to the golang package
	return golang.PrepareRename(ctx, snapshot, fh, params.Position)
}
```

The server delegates to the `golang.PrepareRename` function, which performs the actual analysis.

### 6.4. Language-Specific Implementation

The `golang.PrepareRename` function in `golang/rename.go` performs the rename validation:

```go
func PrepareRename(ctx context.Context, snapshot *cache.Snapshot, fh file.Handle, pp protocol.Position) (*protocol.PrepareRenameResult, error) {
	// Parse the file and find the identifier at the position
	pkg, pgf, err := NarrowestPackageForFile(ctx, snapshot, fh.URI())
	if err != nil {
		return nil, err
	}

	pos, err := pgf.PositionPos(pp)
	if err != nil {
		return nil, err
	}

	// Find the identifier at the position
	path, _ := astutil.PathEnclosingInterval(pgf.File, pos, pos)
	if len(path) == 0 {
		return nil, errNoIdentFound
	}

	id, ok := path[0].(*ast.Ident)
	if !ok {
		return nil, errNoIdentFound
	}

	// Check if the identifier can be renamed
	obj := pkg.TypesInfo().ObjectOf(id)
	if obj == nil {
		return nil, errNoObjectFound
	}

	// Perform various checks to ensure the rename is valid
	// ...

	// Return the range of the identifier
	rng, err := pgf.NodeRange(id)
	if err != nil {
		return nil, err
	}

	return &protocol.PrepareRenameResult{
		Range:       rng,
		Placeholder: id.Name,
	}, nil
}
```

This function:

1. Parses the file and builds the AST.
2. Finds the identifier at the specified position.
3. Looks up the identifier's object in the type information.
4. Performs various checks to ensure the rename is valid (e.g., not a built-in, not in a different package).
5. Returns the range of the identifier.

### 6.5. Example Usage

```bash
$ gopls prepare_rename myfile.go:10:5
myfile.go:10:5-10:15
```

This indicates that the identifier at line 10, column 5 can be renamed, and the rename would affect the range from column 5 to column 15.

---

## 7. Deep Dive: call_hierarchy Command

The `call_hierarchy` command displays the call hierarchy for a function, showing both incoming calls (callers) and outgoing calls (callees).

### 7.1. Command Structure

The command is defined in `call_hierarchy.go`:

```go
type callHierarchy struct {
	app *Application
}

func (c *callHierarchy) Name() string      { return "call_hierarchy" }
func (c *callHierarchy) Parent() string    { return c.app.Name() }
func (c *callHierarchy) Usage() string     { return "<position>" }
func (c *callHierarchy) ShortHelp() string { return "display selected identifier's call hierarchy" }
```

### 7.2. Implementation

The `Run` method:

```go
func (c *callHierarchy) Run(ctx context.Context, args ...string) error {
	if len(args) != 1 {
		return tool.CommandLineErrorf("call_hierarchy expects 1 argument (position)")
	}

	cli, _, err := c.app.connect(ctx)
	if err != nil {
		return err
	}
	defer cli.terminate(ctx)

	from := parseSpan(args[0])
	file, err := cli.openFile(ctx, from.URI())
	if err != nil {
		return err
	}

	loc, err := file.spanLocation(from)
	if err != nil {
		return err
	}

	p := protocol.CallHierarchyPrepareParams{
		TextDocumentPositionParams: protocol.LocationTextDocumentPositionParams(loc),
	}

	callItems, err := cli.server.PrepareCallHierarchy(ctx, &p)
	if err != nil {
		return err
	}
	if len(callItems) == 0 {
		return fmt.Errorf("function declaration identifier not found at %v", args[0])
	}

	for _, item := range callItems {
		incomingCalls, err := cli.server.IncomingCalls(ctx, &protocol.CallHierarchyIncomingCallsParams{Item: item})
		if err != nil {
			return err
		}
		for i, call := range incomingCalls {
			printString, err := callItemPrintString(ctx, cli, call.From, call.From.URI, call.FromRanges)
			if err != nil {
				return err
			}
			fmt.Printf("caller[%d]: %s\n", i, printString)
		}

		printString, err := callItemPrintString(ctx, cli, item, "", nil)
		if err != nil {
			return err
		}
		fmt.Printf("identifier: %s\n", printString)

		outgoingCalls, err := cli.server.OutgoingCalls(ctx, &protocol.CallHierarchyOutgoingCallsParams{Item: item})
		if err != nil {
			return err
		}
		for i, call := range outgoingCalls {
			printString, err := callItemPrintString(ctx, cli, call.To, item.URI, call.FromRanges)
			if err != nil {
				return err
			}
			fmt.Printf("callee[%d]: %s\n", i, printString)
		}
	}

	return nil
}
```

The flow is:

1. Parse the position.
2. Connect to the server and open the file.
3. Call `PrepareCallHierarchy` to get the initial `CallHierarchyItem`.
4. Call `IncomingCalls` to get the callers.
5. Call `OutgoingCalls` to get the callees.
6. Print the results.

### 7.3. Server-Side Implementation

The server methods in `server/call_hierarchy.go` delegate to the `golang` package:

```go
func (s *server) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	ctx, done := event.Start(ctx, "server.PrepareCallHierarchy")
	defer done()

	fh, snapshot, release, err := s.session.FileOf(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	defer release()
	switch snapshot.FileKind(fh) {
	case file.Go:
		return golang.PrepareCallHierarchy(ctx, snapshot, fh, params.Range)
	}
	return nil, nil
}

func (s *server) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	// Similar delegation to golang.IncomingCalls
}

func (s *server) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	// Similar delegation to golang.OutgoingCalls
}
```

### 7.4. Language-Specific Implementation

The `golang.PrepareCallHierarchy` function in `golang/call_hierarchy.go`:

```go
func PrepareCallHierarchy(ctx context.Context, snapshot *cache.Snapshot, fh file.Handle, rng protocol.Range) ([]protocol.CallHierarchyItem, error) {
	ctx, done := event.Start(ctx, "golang.PrepareCallHierarchy")
	defer done()

	pkg, pgf, err := NarrowestPackageForFile(ctx, snapshot, fh.URI())
	if err != nil {
		return nil, err
	}
	start, end, err := pgf.RangePos(rng)
	if err != nil {
		return nil, err
	}
	obj, err := callHierarchyFuncAtRange(pkg.TypesInfo(), pgf, astutil.RangeOf(start, end))
	if err != nil {
		return nil, err
	}
	declLoc, err := ObjectLocation(ctx, pkg.FileSet(), snapshot, obj)
	if err != nil {
		return nil, err
	}

	return []protocol.CallHierarchyItem{{
		Name:           obj.Name(),
		Kind:           protocol.Function,
		Tags:           []protocol.SymbolTag{},
		Detail:         callHierarchyItemDetail(obj, declLoc),
		URI:            declLoc.URI,
		Range:          declLoc.Range,
		SelectionRange: declLoc.Range,
	}}, nil
}
```

The `IncomingCalls` function finds all references to the function and groups them by their enclosing function:

```go
func IncomingCalls(ctx context.Context, snapshot *cache.Snapshot, fh file.Handle, rng protocol.Range) ([]protocol.CallHierarchyIncomingCall, error) {
	ctx, done := event.Start(ctx, "golang.IncomingCalls")
	defer done()

	refs, err := references(ctx, snapshot, fh, rng, false)
	if err != nil {
		return nil, err
	}

	// Group references by their enclosing function declaration
	incomingCalls := make(map[protocol.Location]*protocol.CallHierarchyIncomingCall)
	for _, ref := range refs {
		callItem, err := enclosingNodeCallItem(ctx, snapshot, ref.pkgPath, ref.location)
		if err != nil {
			continue
		}
		loc := callItem.URI.Location(callItem.Range)
		call, ok := incomingCalls[loc]
		if !ok {
			call = &protocol.CallHierarchyIncomingCall{From: callItem}
			incomingCalls[loc] = call
		}
		call.FromRanges = append(call.FromRanges, ref.location.Range)
	}

	// Convert map to slice
	incomingCallItems := make([]protocol.CallHierarchyIncomingCall, 0, len(incomingCalls))
	for _, callItem := range moremaps.SortedFunc(incomingCalls, protocol.CompareLocation) {
		incomingCallItems = append(incomingCallItems, *callItem)
	}
	return incomingCallItems, nil
}
```

The `OutgoingCalls` function finds all function calls within the function:

```go
func OutgoingCalls(ctx context.Context, snapshot *cache.Snapshot, fh file.Handle, pp protocol.Position) ([]protocol.CallHierarchyOutgoingCall, error) {
	// ... parse and find the function declaration ...

	// Find calls to known functions/methods
	var callRanges []astutil.Range
	for n := range ast.Preorder(declNode) {
		if call, ok := n.(*ast.CallExpr); ok {
			callee := typeutil.Callee(pkg.TypesInfo(), call)
			switch callee.(type) {
			case *types.Func, *types.Builtin:
				if callee.Pkg() == nil {
					continue // Skip trivial builtins
				}
				id := typesinternal.UsedIdent(pkg.TypesInfo(), call.Fun)
				callRanges = append(callRanges, astutil.RangeOf(id.Pos(), id.End()))
			}
		}
	}

	// Group by callee
	outgoingCalls := make(map[protocol.Location]*protocol.CallHierarchyOutgoingCall)
	for _, callRange := range callRanges {
		obj, err := callHierarchyFuncAtRange(declPkg.TypesInfo(), declPGF, callRange)
		if err != nil {
			continue
		}

		loc, err := ObjectLocation(ctx, declPkg.FileSet(), snapshot, obj)
		if err != nil {
			return nil, err
		}

		outgoingCall, ok := outgoingCalls[loc]
		if !ok {
			outgoingCall = &protocol.CallHierarchyOutgoingCall{
				To: protocol.CallHierarchyItem{
					Name:           obj.Name(),
					Kind:           protocol.Function,
					Tags:           []protocol.SymbolTag{},
					Detail:         callHierarchyItemDetail(obj, loc),
					URI:            loc.URI,
					Range:          loc.Range,
					SelectionRange: loc.Range,
				},
			}
			outgoingCalls[loc] = outgoingCall
		}

		rng, err := declPGF.PosRange(callRange.Pos(), callRange.End())
		if err != nil {
			return nil, err
		}
		outgoingCall.FromRanges = append(outgoingCall.FromRanges, rng)
	}

	// Convert map to slice
	outgoingCallItems := make([]protocol.CallHierarchyOutgoingCall, 0, len(outgoingCalls))
	for _, callItem := range moremaps.SortedFunc(outgoingCalls, protocol.CompareLocation) {
		outgoingCallItems = append(outgoingCallItems, *callItem)
	}
	return outgoingCallItems, nil
}
```

### 7.5. Example Usage

```bash
$ gopls call_hierarchy myfile.go:10:5
caller[0]: ranges 15:10-15:15 in /path/to/caller.go from/to function Caller in myfile.go:15:1-15:20
identifier: function MyFunc in myfile.go:10:1-10:15
callee[0]: ranges 12:5-12:10 in /path/to/myfile.go from/to function Helper in helper.go:5:1-5:15
```

---

## 8. Deep Dive: codeaction Command

The `codeaction` command lists or executes code actions for a given file or range. Code actions include quick fixes, refactorings, and other transformations.

### 8.1. Command Structure

The command is defined in `codeaction.go`:

```go
type codeaction struct {
	EditFlags
	Kind  string `flag:"kind" help:"comma-separated list of code action kinds to filter"`
	Title string `flag:"title" help:"regular expression to match title"`
	Exec  bool   `flag:"exec" help:"execute the first matching code action"`

	app *Application
}
```

The command supports filtering by kind and title, and can either list or execute actions.

### 8.2. Implementation

The `Run` method:

```go
func (cmd *codeaction) Run(ctx context.Context, args ...string) error {
	if len(args) < 1 {
		return tool.CommandLineErrorf("codeaction expects at least 1 argument")
	}
	cmd.app.editFlags = &cmd.EditFlags
	cli, _, err := cmd.app.connect(ctx)
	if err != nil {
		return err
	}
	defer cli.terminate(ctx)

	from := parseSpan(args[0])
	uri := from.URI()
	file, err := cli.openFile(ctx, uri)
	if err != nil {
		return err
	}
	rng, err := file.spanRange(from)
	if err != nil {
		return err
	}

	titleRE, err := regexp.Compile(cmd.Title)
	if err != nil {
		return err
	}

	// Get diagnostics
	if err := diagnoseFiles(ctx, cli.server, []protocol.DocumentURI{uri}); err != nil {
		return err
	}
	file.diagnosticsMu.Lock()
	diagnostics := slices.Clone(file.diagnostics)
	file.diagnosticsMu.Unlock()

	// Request code actions
	var kinds []protocol.CodeActionKind
	if cmd.Kind != "" {
		for kind := range strings.SplitSeq(cmd.Kind, ",") {
			kinds = append(kinds, protocol.CodeActionKind(kind))
		}
	} else {
		kinds = append(kinds, protocol.Empty)
	}
	actions, err := cli.server.CodeAction(ctx, &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
		Range:        rng,
		Context: protocol.CodeActionContext{
			Only:        kinds,
			Diagnostics: diagnostics,
		},
	})
	if err != nil {
		return fmt.Errorf("%v: %v", from, err)
	}

	// Process actions
	for _, act := range actions {
		if act.Disabled != nil {
			continue
		}
		if !titleRE.MatchString(act.Title) {
			continue
		}

		// Check if action matches the position
		if from.HasPosition() && len(act.Diagnostics) > 0 &&
			!slices.ContainsFunc(act.Diagnostics, func(diag protocol.Diagnostic) bool {
				return diag.Range.Start == rng.Start
			}) {
			continue
		}

		if cmd.Exec {
			// Execute the action
			if act.Command != nil {
				if _, err := executeCommand(ctx, cli.server, act.Command); err != nil {
					return err
				}
			} else {
				// Apply edits
				for _, c := range act.Edit.DocumentChanges {
					tde := c.TextDocumentEdit
					if tde != nil && tde.TextDocument.URI == uri {
						edits := protocol.AsTextEdits(tde.Edits)
						return applyTextEdits(file.mapper, edits, cmd.app.editFlags)
					}
				}
			}
			return nil
		} else {
			// List the action
			action := "edit"
			if act.Command != nil {
				action = "command"
			}
			fmt.Printf("%s\t%q [%s]\n", action, act.Title, act.Kind)
		}
	}

	if cmd.Exec {
		return fmt.Errorf("no matching code action at %s", from)
	}
	return nil
}
```

### 8.3. Server-Side Implementation

The server's `CodeAction` method in `server/code_action.go`:

```go
func (s *server) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	ctx, done := event.Start(ctx, "server.CodeAction")
	defer done()

	fh, snapshot, release, err := s.session.FileOf(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	defer release()
	uri := fh.URI()
	kind := snapshot.FileKind(fh)

	// Determine enabled code action kinds
	codeActionKinds := make(map[protocol.CodeActionKind]bool)
	if len(params.Context.Only) > 0 {
		for _, kind := range params.Context.Only {
			codeActionKinds[kind] = true
		}
	} else {
		// Heuristic based on trigger kind
		if triggerKind(params) == protocol.CodeActionAutomatic {
			codeActionKinds[protocol.QuickFix] = true
		} else {
			codeActionKinds[protocol.Empty] = true
		}
	}

	enabled := func(kind protocol.CodeActionKind) bool {
		// Check if kind is enabled (hierarchical matching)
		for {
			if v, ok := codeActionKinds[kind]; ok {
				return v
			}
			if kind == "" {
				return false
			}
			// Special case for source.test
			if kind == settings.GoTest {
				return false
			}
			// Try parent
			if dot := strings.LastIndexByte(string(kind), '.'); dot >= 0 {
				kind = kind[:dot]
			} else {
				kind = ""
			}
		}
	}

	switch kind {
	case file.Go:
		// Get diagnostic-bundled code actions
		actions, err := s.codeActionsMatchingDiagnostics(ctx, uri, snapshot, params.Context.Diagnostics, enabled)
		if err != nil {
			return nil, err
		}

		// Get computed code actions
		moreActions, err := golang.CodeActions(ctx, snapshot, fh, params.Range, params.Context.Diagnostics, enabled, triggerKind(params))
		if err != nil {
			return nil, err
		}
		actions = append(actions, moreActions...)

		// Filter for generated files
		if golang.IsGenerated(ctx, snapshot, uri) {
			actions = slices.DeleteFunc(actions, func(a protocol.CodeAction) bool {
				switch a.Kind {
				case settings.GoTest, settings.GoDoc, settings.GoFreeSymbols,
					settings.GoSplitPackage, settings.GoAssembly,
					settings.GoplsDocFeatures, settings.GoToggleCompilerOptDetails:
					return false // read-only query
				case settings.OrganizeImports:
					return false // allowed in generated files
				}
				return true // potential write operation
			})
		}

		return actions, nil

	case file.Mod:
		// Handle go.mod files
		// ...
	}

	return nil, nil
}
```

### 8.4. Language-Specific Implementation

The `golang.CodeActions` function in `golang/code_action.go` computes available code actions:

```go
func CodeActions(ctx context.Context, snapshot *cache.Snapshot, fh file.Handle, rng protocol.Range, diagnostics []protocol.Diagnostic, enabled func(protocol.CodeActionKind) bool, trigger protocol.CodeActionTriggerKind) ([]protocol.CodeAction, error) {
	// Parse the file
	pkg, pgf, err := NarrowestPackageForFile(ctx, snapshot, fh.URI())
	if err != nil {
		return nil, err
	}

	var actions []protocol.CodeAction

	// Add various code actions based on context
	// - Extract function/variable/constant
	// - Inline call
	// - Change quote style
	// - Fill struct
	// - Organize imports
	// - etc.

	// Each action is added if enabled(action.Kind) returns true

	return actions, nil
}
```

### 8.5. Example Usage

List all code actions:

```bash
$ gopls codeaction myfile.go:10:5
edit	"Organize Imports" [source.organizeImports]
command	"Run test" [source.test]
edit	"Extract function" [refactor.extract.function]
```

Execute a specific code action:

```bash
$ gopls codeaction -kind=refactor.extract.function -exec -diff myfile.go:10:5-10:20
--- myfile.go.orig
+++ myfile.go
@@ -8,7 +8,11 @@
 func main() {
-	result := x + y
+	result := add(x, y)
 }
+
+func add(x, y int) int {
+	return x + y
+}
```

---

## 9. Deep Dive: codelens Command

The `codelens` command lists or executes code lenses for a file. Code lenses are commands that appear inline in the editor, such as "run test" or "debug test" above test functions.

### 9.1. Command Structure

The command is defined in `codelens.go`:

```go
type codelens struct {
	EditFlags
	app *Application

	Exec bool `flag:"exec" help:"execute the first matching code lens"`
}
```

### 9.2. Implementation

The `Run` method:

```go
func (r *codelens) Run(ctx context.Context, args ...string) error {
	var filename, title string
	switch len(args) {
	case 0:
		return tool.CommandLineErrorf("codelens requires a file name")
	case 2:
		title = args[1]
		fallthrough
	case 1:
		filename = args[0]
	default:
		return tool.CommandLineErrorf("codelens expects at most two arguments")
	}

	r.app.editFlags = &r.EditFlags

	// Override codelens settings
	origOptions := r.app.options
	r.app.options = func(opts *settings.Options) {
		if origOptions != nil {
			origOptions(opts)
		}
		if opts.Codelenses == nil {
			opts.Codelenses = make(map[settings.CodeLensSource]bool)
		}
		opts.Codelenses[settings.CodeLensTest] = true
	}

	cli, _, err := r.app.connect(ctx)
	if err != nil {
		return err
	}
	defer cli.terminate(ctx)

	filespan := parseSpan(filename)
	file, err := cli.openFile(ctx, filespan.URI())
	if err != nil {
		return err
	}
	loc, err := file.spanLocation(filespan)
	if err != nil {
		return err
	}

	p := protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: loc.URI},
	}
	lenses, err := cli.server.CodeLens(ctx, &p)
	if err != nil {
		return err
	}

	for _, lens := range lenses {
		sp, err := file.rangeSpan(lens.Range)
		if err != nil {
			return nil
		}

		if title != "" && lens.Command.Title != title {
			continue
		}
		if filespan.HasPosition() && !protocol.Intersect(loc.Range, lens.Range) {
			continue
		}

		if r.Exec {
			_, err := executeCommand(ctx, cli.server, lens.Command)
			return err
		}

		fmt.Printf("%v: %q [%s]\n", sp, lens.Command.Title, lens.Command.Command)
	}

	if r.Exec {
		return fmt.Errorf("no code lens at %s with title %q", filespan, title)
	}
	return nil
}
```

### 9.3. Server-Side Implementation

The server's `CodeLens` method in `server/code_lens.go`:

```go
func (s *server) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	ctx, done := event.Start(ctx, "server.CodeLens", label.URI.Of(params.TextDocument.URI))
	defer done()

	fh, snapshot, release, err := s.session.FileOf(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	defer release()

	var lensFuncs map[settings.CodeLensSource]cache.CodeLensSourceFunc
	switch snapshot.FileKind(fh) {
	case file.Mod:
		lensFuncs = mod.CodeLensSources()
	case file.Go:
		lensFuncs = golang.CodeLensSources()
	default:
		return nil, nil
	}

	var lenses []protocol.CodeLens
	for kind, lensFunc := range lensFuncs {
		if !snapshot.Options().Codelenses[kind] {
			continue
		}
		added, err := lensFunc(ctx, snapshot, fh)
		if err != nil {
			event.Error(ctx, fmt.Sprintf("code lens %s failed", kind), err)
			continue
		}
		lenses = append(lenses, added...)
	}

	// Add metadata to commands
	for i := range lenses {
		if lenses[i].Command == nil {
			continue
		}
		meta := codeLensMetadata{Source: "codelens"}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			continue
		}
		lenses[i].Command.Arguments = append(lenses[i].Command.Arguments, json.RawMessage(metaBytes))
	}

	sort.Slice(lenses, func(i, j int) bool {
		a, b := lenses[i], lenses[j]
		if cmp := protocol.CompareRange(a.Range, b.Range); cmp != 0 {
			return cmp < 0
		}
		return a.Command.Command < b.Command.Command
	})
	return lenses, nil
}
```

### 9.4. Language-Specific Implementation

The `golang.CodeLensSources` function returns a map of code lens providers:

```go
func CodeLensSources() map[settings.CodeLensSource]cache.CodeLensSourceFunc {
	return map[settings.CodeLensSource]cache.CodeLensSourceFunc{
		settings.CodeLensTest:         testCodeLens,
		settings.CodeLensReferences:   referencesCodeLens,
		settings.CodeLensImplementations: implementationsCodeLens,
		// ... other code lens sources
	}
}
```

Each provider is a function that computes code lenses for a file. For example, `testCodeLens` finds test functions and adds "run test" and "debug test" lenses:

```go
func testCodeLens(ctx context.Context, snapshot *cache.Snapshot, fh file.Handle) ([]protocol.CodeLens, error) {
	pkg, pgf, err := NarrowestPackageForFile(ctx, snapshot, fh.URI())
	if err != nil {
		return nil, err
	}

	// Only add code lenses in test files
	if !strings.HasSuffix(fh.URI().Path(), "_test.go") {
		return nil, nil
	}

	var lenses []protocol.CodeLens
	for _, decl := range pgf.File.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if it's a test function
		if !isTestFunc(fn) {
			continue
		}

		rng, err := pgf.NodeRange(fn)
		if err != nil {
			return nil, err
		}

		// Add "run test" lens
		lenses = append(lenses, protocol.CodeLens{
			Range: rng,
			Command: &protocol.Command{
				Title:     "run test",
				Command:   command.Test.String(),
				Arguments: []json.RawMessage{/* ... */},
			},
		})

		// Add "debug test" lens
		lenses = append(lenses, protocol.CodeLens{
			Range: rng,
			Command: &protocol.Command{
				Title:     "debug test",
				Command:   command.Test.String(),
				Arguments: []json.RawMessage{/* ... */},
			},
		})
	}

	return lenses, nil
}
```

### 9.5. Example Usage

List code lenses:

```bash
$ gopls codelens myfile_test.go
myfile_test.go:10:1-10:15: "run test" [gopls.test]
myfile_test.go:10:1-10:15: "debug test" [gopls.test]
myfile_test.go:20:1-20:20: "run test" [gopls.test]
```

Execute a code lens:

```bash
$ gopls codelens -exec myfile_test.go:10:1 "run test"
=== RUN   TestMyFunction
--- PASS: TestMyFunction (0.00s)
PASS
```

---

## 10. Server-Side Implementation

The server-side implementation is in the `gopls/internal/server` package. The `server` struct implements the LSP `Server` interface, which defines methods for all LSP requests.

### 10.1. Server Structure

```go
type server struct {
	session *cache.Session
	client  protocol.Client
	options *settings.Options

	// ... other fields
}
```

The server maintains a reference to the session (which manages the workspace state), the client (for sending notifications and requests to the editor), and the options (user preferences).

### 10.2. Request Handling Pattern

Each LSP method follows a similar pattern:

1. Extract the file handle and snapshot from the session.
2. Delegate to a language-specific function (e.g., in the `golang` package).
3. Return the result.

For example, the `Definition` method:

```go
func (s *server) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	ctx, done := event.Start(ctx, "server.Definition")
	defer done()

	fh, snapshot, release, err := s.session.FileOf(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	defer release()

	switch snapshot.FileKind(fh) {
	case file.Go:
		return golang.Definition(ctx, snapshot, fh, params.Position)
	case file.Mod:
		return mod.Definition(ctx, snapshot, fh, params.Position)
	}
	return nil, nil
}
```

This pattern keeps the server layer thin and delegates the actual work to language-specific packages.

### 10.3. File Kind Dispatch

The server dispatches to different implementations based on the file kind:

- `file.Go`: Go source files
- `file.Mod`: `go.mod` files
- `file.Sum`: `go.sum` files
- `file.Work`: `go.work` files
- `file.Tmpl`: Template files

This allows `gopls` to support multiple file types with different semantics.

---

## 11. Language-Specific Logic (golang Package)

The `gopls/internal/golang` package contains the Go-specific implementations of language features. This is where the actual analysis happens.

### 11.1. Package Structure

The `golang` package is organized by feature:

- `call_hierarchy.go`: Call hierarchy
- `code_lens.go`: Code lenses
- `completion.go`: Code completion
- `definition.go`: Go to definition
- `highlight.go`: Document highlights
- `hover.go`: Hover information
- `references.go`: Find references
- `rename.go`: Rename refactoring
- `signature.go`: Signature help
- etc.

### 11.2. Common Patterns

Most functions in the `golang` package follow this pattern:

1. Parse the file and build the AST.
2. Type-check the package.
3. Find the relevant AST node(s).
4. Perform the analysis.
5. Convert the results to LSP protocol types.

For example, the `Definition` function:

```go
func Definition(ctx context.Context, snapshot *cache.Snapshot, fh file.Handle, position protocol.Position) ([]protocol.Location, error) {
	// Parse the file
	pkg, pgf, err := NarrowestPackageForFile(ctx, snapshot, fh.URI())
	if err != nil {
		return nil, err
	}

	// Convert position to token.Pos
	pos, err := pgf.PositionPos(position)
	if err != nil {
		return nil, err
	}

	// Find the identifier at the position
	path, _ := astutil.PathEnclosingInterval(pgf.File, pos, pos)
	if len(path) == 0 {
		return nil, nil
	}

	id, ok := path[0].(*ast.Ident)
	if !ok {
		return nil, nil
	}

	// Look up the object
	obj := pkg.TypesInfo().ObjectOf(id)
	if obj == nil {
		return nil, nil
	}

	// Find the declaration
	declLoc, err := ObjectLocation(ctx, pkg.FileSet(), snapshot, obj)
	if err != nil {
		return nil, err
	}

	return []protocol.Location{declLoc}, nil
}
```

### 11.3. Type Checking and Analysis

The `golang` package relies heavily on the `go/types` package for type checking and the `go/ast` package for AST manipulation. The `cache` package provides parsed files and type information, which are cached for performance.

### 11.4. Refactorings

Refactorings like extract function, inline call, and rename are more complex. They involve:

1. Analyzing the code to determine if the refactoring is valid.
2. Computing the necessary edits.
3. Returning a `WorkspaceEdit` that describes the changes.

For example, the `Rename` function:

```go
func Rename(ctx context.Context, snapshot *cache.Snapshot, fh file.Handle, position protocol.Position, newName string) (*protocol.WorkspaceEdit, error) {
	// Validate the rename
	// Find all references
	// Compute edits for each reference
	// Return a WorkspaceEdit

	// ... (implementation details)
}
```

---

## 12. Protocol Layer and LSP Types

The `gopls/internal/protocol` package defines the LSP protocol types and provides utilities for working with them.

### 12.1. Generated Types

Most of the protocol types are generated from the LSP specification. The files `tsprotocol.go`, `tsserver.go`, and `tsclient.go` contain these generated types and interfaces.

For example, the `TextDocumentPositionParams` type:

```go
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type TextDocumentIdentifier struct {
	URI DocumentURI `json:"uri"`
}

type Position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}
```

### 12.2. Server and Client Interfaces

The `Server` and `Client` interfaces define all LSP methods:

```go
type Server interface {
	Initialize(context.Context, *ParamInitialize) (*InitializeResult, error)
	Initialized(context.Context, *InitializedParams) error
	Shutdown(context.Context) error
	Exit(context.Context) error

	// Document synchronization
	DidOpen(context.Context, *DidOpenTextDocumentParams) error
	DidChange(context.Context, *DidChangeTextDocumentParams) error
	DidClose(context.Context, *DidCloseTextDocumentParams) error
	DidSave(context.Context, *DidSaveTextDocumentParams) error

	// Language features
	Completion(context.Context, *CompletionParams) (*CompletionList, error)
	Hover(context.Context, *HoverParams) (*Hover, error)
	SignatureHelp(context.Context, *SignatureHelpParams) (*SignatureHelp, error)
	Definition(context.Context, *DefinitionParams) ([]Location, error)
	References(context.Context, *ReferenceParams) ([]Location, error)
	DocumentHighlight(context.Context, *DocumentHighlightParams) ([]DocumentHighlight, error)
	DocumentSymbol(context.Context, *DocumentSymbolParams) ([]any, error)
	CodeAction(context.Context, *CodeActionParams) ([]CodeAction, error)
	CodeLens(context.Context, *CodeLensParams) ([]CodeLens, error)
	Formatting(context.Context, *DocumentFormattingParams) ([]TextEdit, error)
	RangeFormatting(context.Context, *DocumentRangeFormattingParams) ([]TextEdit, error)
	Rename(context.Context, *RenameParams) (*WorkspaceEdit, error)
	PrepareRename(context.Context, *PrepareRenameParams) (*PrepareRenameResult, error)
	PrepareCallHierarchy(context.Context, *CallHierarchyPrepareParams) ([]CallHierarchyItem, error)
	IncomingCalls(context.Context, *CallHierarchyIncomingCallsParams) ([]CallHierarchyIncomingCall, error)
	OutgoingCalls(context.Context, *CallHierarchyOutgoingCallsParams) ([]CallHierarchyOutgoingCall, error)

	// ... many more methods
}
```

### 12.3. Mapper: UTF-8/UTF-16 Conversion

One of the key challenges in implementing LSP is handling the difference between UTF-8 (used by Go) and UTF-16 (used by LSP). The `Mapper` type provides utilities for converting between these encodings:

```go
type Mapper struct {
	URI     DocumentURI
	Content []byte
	// ... internal fields for conversion
}

func NewMapper(uri DocumentURI, content []byte) *Mapper {
	// ... create and initialize the mapper
}

func (m *Mapper) RangeOffsets(r Range) (int, int, error) {
	// Convert a protocol.Range (UTF-16) to byte offsets (UTF-8)
}

func (m *Mapper) PosRange(start, end token.Pos) (Range, error) {
	// Convert token.Pos (UTF-8) to protocol.Range (UTF-16)
}
```

The mapper is essential for correctly handling positions in files with non-ASCII characters.

---

## 13. Cache and Session Management

The `gopls/internal/cache` package provides caching and session management. This is crucial for performance, as parsing and type-checking Go code can be expensive.

### 13.1. Cache

The `Cache` is a global cache shared across all sessions:

```go
type Cache struct {
	// ... internal fields
}

func New() *Cache {
	return &Cache{
		// ... initialization
	}
}
```

The cache stores parsed files, type information, and other computed results.

### 13.2. Session

A `Session` represents a user's workspace:

```go
type Session struct {
	id          string
	cache       *Cache
	gocmdRunner *gocommand.Runner

	viewMu  sync.Mutex
	views   []*View
	viewMap map[protocol.DocumentURI]*View

	snapshotWG sync.WaitGroup
	parseCache *parseCache

	*overlayFS
}
```

A session can have multiple views (one per workspace folder) and maintains an overlay filesystem for unsaved file changes.

### 13.3. Snapshot

A `Snapshot` represents the state of a workspace at a particular point in time:

```go
type Snapshot struct {
	sequenceID uint64
	globalID   SnapshotID

	view *View

	// ... internal fields
}
```

Snapshots are immutable and can be used concurrently. When a file changes, a new snapshot is created.

### 13.4. View

A `View` represents a workspace folder:

```go
type View struct {
	id     string
	folder *Folder

	// ... internal fields
}
```

A view manages the Go modules and packages within a folder.

### 13.5. Caching Strategy

The caching strategy is designed to minimize redundant work:

1. Parsed files are cached and reused across snapshots.
2. Type information is cached per package.
3. Analysis results are cached and invalidated when dependencies change.
4. The cache uses a least-recently-used (LRU) eviction policy to limit memory usage.

---

## 14. Building Custom Tooling: CLI Invocation

The simplest way to build custom tooling that leverages `gopls` is to invoke the CLI commands directly. This approach is suitable for many use cases and requires no knowledge of the `gopls` internals.

### 14.1. Basic Invocation

You can invoke `gopls` commands using the `os/exec` package in Go:

```go
package main

import (
	"fmt"
	"os/exec"
)

func main() {
	// Get the definition of a symbol
	out, err := exec.Command("gopls", "definition", "myfile.go:10:5").Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Definition:", string(out))
}
```

### 14.2. Parsing Output

The output format varies by command. Some commands produce human-readable output, while others support JSON output (via the `-json` flag).

For example, the `definition` command supports `-json`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Definition struct {
	Span        string `json:"span"`
	Description string `json:"description"`
}

func main() {
	out, err := exec.Command("gopls", "definition", "-json", "myfile.go:10:5").Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var def Definition
	if err := json.Unmarshal(out, &def); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	fmt.Printf("Definition at %s: %s\n", def.Span, def.Description)
}
```

### 14.3. Handling Errors

The CLI returns non-zero exit codes on errors. You can check the error type to determine what went wrong:

```go
out, err := exec.Command("gopls", "definition", "myfile.go:10:5").CombinedOutput()
if err != nil {
	if exitErr, ok := err.(*exec.ExitError); ok {
		fmt.Printf("gopls exited with code %d\n", exitErr.ExitCode())
		fmt.Printf("Output: %s\n", string(out))
	} else {
		fmt.Printf("Failed to run gopls: %v\n", err)
	}
	return
}
```

### 14.4. Advantages and Limitations

**Advantages:**

- Simple to implement.
- No need to understand `gopls` internals.
- Stable interface (the CLI is more stable than internal packages).

**Limitations:**

- Performance overhead of starting a new process for each command.
- Limited control over the analysis (e.g., can't customize options easily).
- Output parsing can be fragile if the format changes.

---

## 15. Building Custom Tooling: Library Usage

For more advanced integrations, you can use the `gopls` libraries directly in your Go code. This gives you more control and better performance, but requires understanding the internal APIs.

### 15.1. Creating an In-Process Server

You can create an in-process LSP server similar to how the CLI does it:

```go
package main

import (
	"context"
	"fmt"

	"golang.org/x/tools/gopls/internal/cache"
	"golang.org/x/tools/gopls/internal/protocol"
	"golang.org/x/tools/gopls/internal/server"
	"golang.org/x/tools/gopls/internal/settings"
)

type myClient struct {
	protocol.Client
}

func (c *myClient) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) error {
	fmt.Printf("Server message: %s\n", params.Message)
	return nil
}

// Implement other required client methods...

func main() {
	ctx := context.Background()

	// Create cache and session
	c := cache.New()
	defer c.Close()
	session := cache.NewSession(ctx, c)
	defer session.Shutdown(ctx)

	// Create client and server
	client := &myClient{}
	options := settings.DefaultOptions(nil)
	svr := server.New(session, client, options)

	// Initialize the server
	params := &protocol.ParamInitialize{
		RootURI: protocol.URIFromPath("/path/to/workspace"),
		Capabilities: protocol.ClientCapabilities{
			// ... set capabilities
		},
	}
	_, err := svr.Initialize(ctx, params)
	if err != nil {
		fmt.Println("Initialize error:", err)
		return
	}

	err = svr.Initialized(ctx, &protocol.InitializedParams{})
	if err != nil {
		fmt.Println("Initialized error:", err)
		return
	}

	// Now you can call server methods
	// For example, get the definition of a symbol
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.URIFromPath("/path/to/myfile.go"),
			},
			Position: protocol.Position{Line: 9, Character: 4}, // 0-indexed
		},
	}
	locs, err := svr.Definition(ctx, defParams)
	if err != nil {
		fmt.Println("Definition error:", err)
		return
	}

	for _, loc := range locs {
		fmt.Printf("Definition: %s:%d:%d\n", loc.URI.Path(), loc.Range.Start.Line+1, loc.Range.Start.Character+1)
	}
}
```

### 15.2. Using Language-Specific Functions

You can also call the language-specific functions directly, bypassing the server layer:

```go
package main

import (
	"context"
	"fmt"

	"golang.org/x/tools/gopls/internal/cache"
	"golang.org/x/tools/gopls/internal/golang"
	"golang.org/x/tools/gopls/internal/protocol"
)

func main() {
	ctx := context.Background()

	// Create cache and session
	c := cache.New()
	defer c.Close()
	session := cache.NewSession(ctx, c)
	defer session.Shutdown(ctx)

	// Create a view for your workspace
	folder := &cache.Folder{
		Dir: protocol.URIFromPath("/path/to/workspace"),
		// ... set other fields
	}
	view, snapshot, release, err := session.NewView(ctx, folder)
	if err != nil {
		fmt.Println("NewView error:", err)
		return
	}
	defer release()

	// Get a file handle
	uri := protocol.URIFromPath("/path/to/myfile.go")
	fh, err := snapshot.ReadFile(ctx, uri)
	if err != nil {
		fmt.Println("ReadFile error:", err)
		return
	}

	// Call a language-specific function
	position := protocol.Position{Line: 9, Character: 4}
	locs, err := golang.Definition(ctx, snapshot, fh, position)
	if err != nil {
		fmt.Println("Definition error:", err)
		return
	}

	for _, loc := range locs {
		fmt.Printf("Definition: %s:%d:%d\n", loc.URI.Path(), loc.Range.Start.Line+1, loc.Range.Start.Character+1)
	}
}
```

### 15.3. Handling File Changes

If you're building a long-running tool, you'll need to handle file changes by creating new snapshots:

```go
// When a file changes
uri := protocol.URIFromPath("/path/to/myfile.go")
content := []byte("package main\n\nfunc main() {}\n")

// Update the overlay
session.DidModifyFiles(ctx, []cache.FileModification{
	{
		URI:        uri,
		Action:     cache.Change,
		Version:    1,
		Text:       content,
		LanguageID: "go",
	},
})

// Get a new snapshot
view, snapshot, release, err := session.ViewOf(uri)
if err != nil {
	// handle error
}
defer release()

// Use the new snapshot for analysis
```

### 15.4. Advantages and Limitations

**Advantages:**

- Better performance (no process startup overhead).
- More control over the analysis.
- Can maintain state across multiple operations.
- Can customize options and behavior.

**Limitations:**

- More complex to implement.
- Requires understanding of internal APIs.
- Internal packages are not guaranteed to be stable.
- Need to handle file changes and cache invalidation.

### 15.5. API Stability Warning

**Important**: The internal packages of `gopls` are not guaranteed to have a stable API. They may change without notice in future versions. If you use these packages directly, you should:

1. Pin to a specific version of `gopls`.
2. Be prepared to update your code when upgrading.
3. Consider vendoring the `gopls` code if stability is critical.

---

## 16. Best Practices and Recommendations

Based on the analysis of the `gopls` codebase, here are some best practices and recommendations for building custom tooling:

### 16.1. When to Use CLI vs. Libraries

**Use the CLI when:**

- You need a simple, one-off operation.
- You want a stable interface.
- You're building a script or tool in a language other than Go.
- Performance is not critical.

**Use the libraries when:**

- You need high performance (e.g., analyzing many files).
- You want more control over the analysis.
- You're building a long-running tool (e.g., a custom language server).
- You need to customize options or behavior.

### 16.2. Position Handling

When working with positions:

- Remember that LSP uses 0-indexed lines and UTF-16 code units.
- Go uses 1-indexed lines and UTF-8 bytes.
- Always use `protocol.Mapper` for conversions.
- Be careful with files containing non-ASCII characters.

### 16.3. Error Handling

- Always check for errors from `gopls` functions.
- Be prepared to handle cases where analysis fails (e.g., parse errors, type errors).
- Provide meaningful error messages to users.

### 16.4. Performance Considerations

- Reuse sessions and snapshots when possible.
- Avoid opening the same file multiple times.
- Use the cache to avoid redundant parsing and type-checking.
- Consider using goroutines for parallel analysis.

### 16.5. Testing

- Write tests for your tooling using the `gopls` test infrastructure.
- Use the `gopls/internal/test/integration` package for integration tests.
- Test with various Go versions and module configurations.

### 16.6. Documentation

- Document the `gopls` version your tool is compatible with.
- Provide examples of how to use your tool.
- Explain any limitations or known issues.

---

## 17. Conclusion

This document has provided an exhaustive analysis of the `gopls` CLI tools, covering the architecture, implementation details, and underlying libraries. We have examined the command framework, position handling, server initialization, and the implementation of specific commands like `prepare_rename`, `call_hierarchy`, `codeaction`, and `codelens`.

The key takeaways are:

1. **Modular Architecture**: The `gopls` codebase is well-organized, with clear separation between the CLI layer, server layer, and language-specific logic.

2. **LSP Protocol**: Understanding the LSP protocol is essential for working with `gopls`, especially the difference between UTF-8 and UTF-16 encodings.

3. **Two Approaches**: You can build custom tooling either by invoking the CLI or by using the libraries directly. The CLI is simpler but less flexible; the libraries offer more control but require deeper understanding.

4. **Caching and Performance**: The cache and session management are crucial for performance. Reusing sessions and snapshots can significantly improve the speed of your tool.

5. **API Stability**: The internal packages are not guaranteed to be stable. If you use them, be prepared for changes in future versions.

By understanding how `gopls` works internally, you can build powerful custom tools that leverage Go's type system and semantic analysis. Whether you're building a code quality checker, a refactoring tool, or a custom language server, the patterns and techniques described in this document will serve as a solid foundation.

The `gopls` project is actively developed and continuously improving. As you build your tooling, consider contributing back to the project by reporting issues, suggesting improvements, or submitting patches. The Go community will benefit from your contributions.

---

**End of Document**
