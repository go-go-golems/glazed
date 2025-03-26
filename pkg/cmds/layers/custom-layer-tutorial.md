# Tutorial: Creating and Using a Custom Parameters Layer in Glazed

This tutorial demonstrates how to create and use a custom parameters layer in Glazed, using the Clay logging layer as a real-world example.

## 1. Creating a New Parameters Layer

A parameters layer in Glazed allows you to group related parameters together. Here's how to create a new one:

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func NewLoggingLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "logging",
        "Logging configuration options",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "with-caller",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Log caller information"),
                parameters.WithDefault(false),
            ),
            parameters.NewParameterDefinition(
                "log-level",
                parameters.ParameterTypeChoice,
                parameters.WithHelp("Log level (debug, info, warn, error, fatal)"),
                parameters.WithDefault("info"),
                parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
            ),
            parameters.NewParameterDefinition(
                "log-format",
                parameters.ParameterTypeChoice,
                parameters.WithHelp("Log format (json, text)"),
                parameters.WithDefault("text"),
                parameters.WithChoices("json", "text"),
            ),
            parameters.NewParameterDefinition(
                "log-file",
                parameters.ParameterTypeString,
                parameters.WithHelp("Log file (default: stderr)"),
                parameters.WithDefault(""),
            ),
            parameters.NewParameterDefinition(
                "verbose",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Verbose output"),
                parameters.WithDefault(false),
            ),
        ),
    )
}
```

This creates a new parameter layer with various logging-related parameters. Note the use of `ParameterTypeChoice` for parameters with a fixed set of valid values.

## 2. Creating a Settings Struct

To easily access the parsed values of your custom layer, create a struct that maps to these parameters:

```go
// LoggingSettings holds the logging configuration parameters
type LoggingSettings struct {
    WithCaller bool   `glazed.parameter:"with-caller"`
    LogLevel   string `glazed.parameter:"log-level"`
    LogFormat  string `glazed.parameter:"log-format"`
    LogFile    string `glazed.parameter:"log-file"`
    Verbose    bool   `glazed.parameter:"verbose"`
}
```

Note that the struct tags must exactly match the parameter names defined in your layer.

## 3. Parsing the Layer into the Settings Struct

Create a helper function to extract the settings from a parsed layers object:

```go
// GetLoggingSettingsFromParsedLayers extracts logging settings from parsed layers
func GetLoggingSettingsFromParsedLayers(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error) {
    s := &LoggingSettings{}
    if err := parsedLayers.InitializeStruct("logging", s); err != nil {
        return nil, err
    }
    return s, nil
}
```

Then, in your command's `Run` method, you can use this helper function:

```go
func (c *MyCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Get the logging settings from the parsed layers
    loggingSettings, err := GetLoggingSettingsFromParsedLayers(parsedLayers)
    if err != nil {
        return err
    }

    // Now you can use loggingSettings.LogLevel, loggingSettings.LogFormat, etc.
    // ...

    return nil
}
```

## 4. Using the Settings to Initialize a System

In Clay, the logging settings are used to initialize the zerolog logger. Here's an example of how to use the settings:

```go
// InitLoggerFromSettings initializes the logger based on the provided settings
func InitLoggerFromSettings(settings *LoggingSettings) error {
    if settings.WithCaller {
        log.Logger = log.With().Caller().Logger()
    }

    // Set timestamp format to include milliseconds
    zerolog.TimeFieldFormat = time.RFC3339Nano

    // default is json
    var logWriter io.Writer
    if settings.LogFormat == "text" {
        logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
    } else {
        logWriter = os.Stderr
    }

    if settings.LogFile != "" {
        fileLogger := &lumberjack.Logger{
            Filename:   settings.LogFile,
            MaxSize:    10, // megabytes
            MaxBackups: 3,
            MaxAge:     28,    //days
            Compress:   false, // disabled by default
        }
        var writer io.Writer
        writer = fileLogger
        if settings.LogFormat == "text" {
            log.Info().Str("file", settings.LogFile).Msg("Logging to file")
            writer = zerolog.ConsoleWriter{
                NoColor:    true,
                Out:        fileLogger,
                TimeFormat: time.RFC3339Nano,
            }
        }
        logWriter = writer
    }

    log.Logger = log.Output(logWriter)

    logLevel := strings.ToLower(settings.LogLevel)
    if settings.Verbose && logLevel != "trace" {
        logLevel = "debug"
    }

    switch logLevel {
    case "debug":
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    case "info":
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    case "warn":
        zerolog.SetGlobalLevel(zerolog.WarnLevel)
    case "error":
        zerolog.SetGlobalLevel(zerolog.ErrorLevel)
    case "fatal":
        zerolog.SetGlobalLevel(zerolog.FatalLevel)
    }

    return nil
}
```

## 5. Adding the Custom Layer to a Command

To add your custom layer to a Glazed command, you have several options:

### 5.1. Adding the Layer Directly in the Command Constructor

```go
func NewMyCommand() (*MyCommand, error) {
    loggingLayer, err := NewLoggingLayer()
    if err != nil {
        return nil, err
    }

    return &MyCommand{
        CommandDescription: cmds.NewCommandDescription(
            "my-command",
            cmds.WithShort("A command example"),
            cmds.WithLayers(loggingLayer),
        ),
    }, nil
}
```

### 5.2. Creating a Helper Function to Add the Layer

```go
// AddLoggingLayerToCommand adds the logging layer to a Glazed command
func AddLoggingLayerToCommand(cmd cmds.Command) (cmds.Command, error) {
    loggingLayer, err := NewLoggingLayer()
    if err != nil {
        return nil, err
    }

    cmd.Description().SetLayers(loggingLayer)
    return cmd, nil
}
```

## 6. Integrating with Cobra

To use your custom layer with Cobra, you can create a helper function to build a Cobra command with your layer:

```go
// BuildCobraCommandWithLogging builds a Cobra command with the logging layer
func BuildCobraCommandWithLogging(
    cmd cmds.Command,
    options ...cli.CobraParserOption,
) (*cobra.Command, error) {
    cmd, err := AddLoggingLayerToCommand(cmd)
    if err != nil {
        return nil, err
    }

    // Add an option to show logging parameters in the short help
    options = append(options, cli.WithCobraShortHelpLayers("logging"))

    return cli.BuildCobraCommandFromCommand(cmd, options...)
}
```

Then in your main application:

```go
import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/spf13/cobra"
)

func main() {
    // Create the root command
    rootCmd := &cobra.Command{
        Use:   "my-app",
        Short: "My application",
    }

    // Create your command
    myCmd, err := NewMyCommand()
    if err != nil {
        fmt.Printf("Error creating command: %v\n", err)
        os.Exit(1)
    }

    // Build a Cobra command with the logging layer
    cobraCmd, err := BuildCobraCommandFromCommand(myCmd)
    if err != nil {
        fmt.Printf("Error building cobra command: %v\n", err)
        os.Exit(1)
    }

    rootCmd.AddCommand(cobraCmd)

    // Initialize Viper and logging
    err = pkg.InitViper("my-app", rootCmd)
    if err != nil {
        fmt.Printf("Error initializing viper: %v\n", err)
        os.Exit(1)
    }

    // Execute the command
    if err := rootCmd.Execute(); err != nil {
        fmt.Printf("Error executing command: %v\n", err)
        os.Exit(1)
    }
}
```

## 8. Complete Example: Logging Layer in Clay

Below is a complete example showing how to use the logging layer in a command:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/go-go-golems/clay/pkg"
    "github.com/go-go-golems/clay/pkg/logging"
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
)

// ExampleCommand implements a simple command that uses the logging layer
type ExampleCommand struct {
    *cmds.CommandDescription
}

// Ensure ExampleCommand implements the Command interface
var _ cmds.Command = (*ExampleCommand)(nil)

func (c *ExampleCommand) Description() *cmds.CommandDescription {
    return c.CommandDescription
}

func (c *ExampleCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Get the logging settings from the parsed layers
    loggingSettings, err := logging.GetLoggingSettingsFromParsedLayers(parsedLayers)
    if err != nil {
        return err
    }

    // Log some information to show different log levels
    log.Debug().Interface("settings", loggingSettings).Msg("Logging settings")
    log.Info().Msg("Running example command")
    log.Warn().Msg("This is a warning message")
    log.Error().Msg("This is an error message")

    return nil
}

// NewExampleCommand creates a new example command
func NewExampleCommand() (*ExampleCommand, error) {
    cmd := &ExampleCommand{
        CommandDescription: cmds.NewCommandDescription(
            "example",
            cmds.WithShort("Example command showing logging layer usage"),
            cmds.WithLong("This command demonstrates how to use the logging layer in a Glazed command."),
        ),
    }

    return cmd, nil
}

func main() {
    // Create the root command
    rootCmd := &cobra.Command{
        Use:   "logging-example",
        Short: "Example application with logging layer",
    }

    // Create the example command
    exampleCmd, err := NewExampleCommand()
    if err != nil {
        fmt.Printf("Error creating example command: %v\n", err)
        os.Exit(1)
    }

    // Method 1: Build a Cobra command with the logging layer
    cobraCmd, err := pkg.BuildCobraCommandWithLogging(exampleCmd)
    if err != nil {
        fmt.Printf("Error building cobra command: %v\n", err)
        os.Exit(1)
    }

    rootCmd.AddCommand(cobraCmd)

    // Method 2: Add logging layer to the command directly and then build a Cobra command
    // This is useful if you need to customize the command further
    anotherExampleCmd, err := NewExampleCommand()
    if err != nil {
        fmt.Printf("Error creating another example command: %v\n", err)
        os.Exit(1)
    }

    // Add the logging layer to the command and handle the returned Command interface
    cmdWithLogging, err := pkg.AddLoggingLayerToCommand(anotherExampleCmd)
    if err != nil {
        fmt.Printf("Error adding logging layer: %v\n", err)
        os.Exit(1)
    }

    // Build the Cobra command with short help for the logging layer
    anotherCobraCmd, err := cli.BuildCobraCommandFromCommand(
        cmdWithLogging,
        cli.WithCobraShortHelpLayers("logging"),
    )
    if err != nil {
        fmt.Printf("Error building another cobra command: %v\n", err)
        os.Exit(1)
    }

    // Give it a different name
    anotherCobraCmd.Use = "another-example"
    rootCmd.AddCommand(anotherCobraCmd)

    // Set up Viper and initialize logging
    err = pkg.InitViper("logging-example", rootCmd)
    if err != nil {
        fmt.Printf("Error initializing viper: %v\n", err)
        os.Exit(1)
    }

    // Execute the command
    if err := rootCmd.Execute(); err != nil {
        fmt.Printf("Error executing command: %v\n", err)
        os.Exit(1)
    }
}
```

## Conclusion

This tutorial showed how to create and use a custom parameters layer in Glazed, using the Clay logging layer as a real-world example. By creating a custom layer, you can:

1. Group related parameters together
2. Provide structured, type-safe access to parameter values
3. Reuse the layer across multiple commands
4. Integrate with Cobra and Viper for a complete CLI solution

The logging layer example demonstrates best practices for creating reusable parameter layers in Glazed applications. 