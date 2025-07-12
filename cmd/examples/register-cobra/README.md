# Register Cobra Command Examples

This example demonstrates the four different ways to register commands with Cobra in the Glazed framework:

## 1. Bare Command (`bare`)
A simple command that implements only the `BareCommand` interface:
- Outputs text directly to stdout
- Uses the traditional `BuildCobraCommandFromBareCommand` builder

```bash
go run . bare --message "Hello World!"
```

## 2. Writer Command (`writer`)
A command that implements the `WriterCommand` interface:
- Outputs to a specified `io.Writer`
- Uses the traditional `BuildCobraCommandFromWriterCommand` builder

```bash
go run . writer --count 3
```

## 3. Glaze Command (`glaze`)
A command that implements the `GlazeCommand` interface:
- Outputs structured data through the Glaze processor
- Uses the traditional `BuildCobraCommandFromGlazeCommand` builder
- Supports all Glaze output formatting options

```bash
# Table format (default)
go run . glaze --rows 3

# JSON format
go run . glaze --rows 3 --output json

# CSV format
go run . glaze --rows 3 --output csv
```

## 4. Dual Command (`dual`) - NEW!
A command that implements both `BareCommand` and `GlazeCommand` interfaces:
- Can run in both classic and glaze modes
- Uses the new `BuildCobraCommandDualMode` builder
- Mode is controlled by the `--with-glaze-output` flag

```bash
# Classic mode (default)
go run . dual --name "Manuel" --times 2

# Glaze mode with table output
go run . dual --name "Manuel" --times 2 --with-glaze-output

# Glaze mode with JSON output
go run . dual --name "Manuel" --times 2 --with-glaze-output --output json
```

## Key Features of the Dual Command

1. **Single Registration**: One command, registered once, with dual functionality
2. **Flag Management**: Glaze flags are automatically injected when needed
3. **Clean Help**: By default, only shows the toggle flag, keeping help clean
4. **Backward Compatible**: Existing command builders remain unchanged
5. **Flexible**: Supports all existing Glaze output formats when in glaze mode

## Customization Options

The dual command builder supports several options:

```go
cobraDualCmd, err := cli.BuildCobraCommandDualMode(
    dualCmd,
    cli.WithGlazeToggleFlag("custom-flag-name"),     // Rename the toggle flag
    cli.WithHiddenGlazeFlags("output", "format"),    // Keep specific flags hidden
    cli.WithDefaultToGlaze(),                        // Make glaze mode the default
)
```
