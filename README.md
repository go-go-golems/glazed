# glazed - A framework for building powerful CLI applications

![](https://img.shields.io/github/license/go-go-golems/glazed)
![](https://img.shields.io/github/actions/workflow/status/go-go-golems/glazed/push.yml?branch=main)

> Add the icing to your structured data!

Glazed is a comprehensive Go framework for building command-line applications that handle structured data elegantly. It provides a rich command system, flexible field management, multiple output formats, and an integrated help system.

The framework implements ideas from [14 great tips to make amazing command line applications](https://dev.to/wesen/14-great-tips-to-make-amazing-cli-applications-3gp3) and focuses on making CLI development both powerful and developer-friendly.

![Command line recording of the functionality described in "Features"](https://imgur.com/ZEtdLes.gif)

## Core Features

### Rich Output Formats
Output structured data in multiple formats with automatic field flattening, filtering, and transformation:

**Tables**: Human-readable ASCII tables and Markdown format
```
❯ glaze json misc/test-data/*.json
+-----+-----+------------+-----+-----+
| a   | b   | c          | d.e | d.f |
+-----+-----+------------+-----+-----+
| 1   | 2   | [3 4 5]    | 6   | 7   |
| 10  | 20  | [30 40 50] | 60  | 70  |
| 100 | 200 | [300]      |     |     |
+-----+-----+------------+-----+-----+
```

**JSON/YAML**: Structured data with optional flattening
```
❯ glaze json misc/test-data/2.json --output json
[
  {
    "a": 10,
    "b": 20,
    "c": [30, 40, 50],
    "d": {
      "e": 60,
      "f": 70
    }
  }
]
```

**CSV/TSV**: Spreadsheet-compatible output
```
❯ glaze json misc/test-data/*.json --output csv
a,b,c,d.e,d.f
1,2,[3 4 5],6,7
10,20,[30 40 50],60,70
100,200,[300],,
```

**Templates**: Go template support for custom formatting
```
❯ glaze json misc/test-data/*.json --template '{{.a}}-{{.b}}: {{.d.f}}'
+---------------------+
| _0                  |
+---------------------+
| 1-2: 7              |
| 10-20: 70           |
| 100-200: <no value> |
+---------------------+
```

### Flexible Command System
Build CLI applications with multiple command types and output modes:

- **BareCommand**: Simple commands with custom output handling
- **WriterCommand**: Commands that write to any io.Writer
- **GlazeCommand**: Commands producing structured data
- **Dual Commands**: Support both classic and structured output modes

### Field Section System
Organize command fields into reusable, composable sections:

- Logical grouping (database, logging, output, etc.)
- Multiple configuration sources (CLI, files, environment)
- Type-safe field extraction
- Built-in validation and help generation

### Integrated Help System
Rich, searchable documentation system with Markdown support:

- Topic-based organization
- Interactive help browsing
- Embedded documentation
- Context-sensitive help

## Building Commands

Glazed provides three main command interfaces for different use cases:

### BareCommand
Simple commands that handle their own output:
```go
type MyCommand struct {
    *cmds.CommandDescription
}

func (c *MyCommand) Run(ctx context.Context, parsedSections *values.Values) error {
    fmt.Println("Hello, World!")
    return nil
}
```

### GlazeCommand
Commands that produce structured data output:
```go
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedSections *values.Values,
    gp middlewares.Processor,
) error {
    row := types.NewRow(
        types.MRP("name", "John"),
        types.MRP("age", 30),
    )
    return gp.AddRow(ctx, row)
}
```

### Dual Commands
Commands supporting both classic and structured output modes:
```go
// Use the dual command builder for Cobra integration
cobraCmd, err := cli.BuildCobraCommandDualMode(
    myDualCommand,
    cli.WithGlazeToggleFlag("with-glaze-output"),
)
```

## Field Sections

Organize command fields into logical, reusable groups:

```go
// Define sections for different concerns
func NewDatabaseSection() *schema.Section {
    return schema.NewSection("database", "Database configuration",
        fields.New("host", fields.TypeString,
            fields.WithDefault("localhost")),
        fields.New("port", fields.TypeInteger,
            fields.WithDefault(5432)),
    )
}

// Use sections in command definitions
cmd := cmds.NewCommandDescription("mycommand",
    cmds.WithSectionsList(
        databaseSection,
        loggingSection, 
        glazedSection,
    ),
)
```

**Benefits:**
- Reuse common field sets across commands
- Avoid field naming conflicts with prefixes
- Type-safe field extraction with structs
- Built-in validation and help generation

## Help System

Create rich, searchable documentation with Markdown files:

```go
// Embed documentation
//go:embed doc/*
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
    return helpSystem.LoadSectionsFromFS(docFS, "doc")
}
```

**Markdown structure:**
```yaml
---
Title: Command Usage Guide
Slug: command-usage
Short: Learn how to use commands effectively
Topics: [commands, usage]
SectionType: Tutorial
---

# Command Usage Content
Your documentation content here...
```

## Installation

### Installing the Framework
To use Glazed as a library in your Go project:
```bash
go get github.com/go-go-golems/glazed
```

### Installing the `glaze` CLI Tool

**Using Homebrew:**
```bash
brew tap go-go-golems/go-go-go
brew install go-go-golems/go-go-go/glazed
```

**Using apt-get:**
```bash
echo "deb [trusted=yes] https://apt.fury.io/go-go-golems/ /" >> /etc/apt/sources.list.d/fury.list
apt-get update
apt-get install glazed
```

**Using yum:**
```bash
echo "
[fury]
name=Gemfury Private Repo
baseurl=https://yum.fury.io/go-go-golems/
enabled=1
gpgcheck=0
" >> /etc/yum.repos.d/fury.repo
yum install glazed
```

**Using go install:**
```bash
go install github.com/go-go-golems/glazed/cmd/glaze@latest
```

**Download binaries from [GitHub Releases](https://github.com/go-go-golems/glazed/releases)**

**Or run from source:**
```bash
go run ./cmd/glaze
```

## Quick Start

1. **Create a command:**
```go
cmd, err := NewMyGlazeCommand()
cobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(cmd)
```

2. **Add to your CLI:**
```go
rootCmd.AddCommand(cobraCmd)
```

3. **Run with multiple output formats:**
```bash
myapp command --output json
myapp command --output table --fields name,status
myapp command --output csv > data.csv
```

## Documentation

For comprehensive guides and API references, see:
- [Commands Reference](pkg/doc/topics/commands-reference.md) - Complete command system guide
- [Field Sections Guide](pkg/doc/topics/sections-guide.md) - Section system with examples
- [Writing Help Entries](pkg/doc/topics/14-writing-help-entries.md) - Help system documentation

## Examples

See the [`cmd/examples`](cmd/examples/) directory for working examples including:
- Basic command types
- Dual command implementations  
- Field section usage
- Help system integration

