---
Title: Creating Custom Parameter Layers
Slug: custom-layer-tutorial
Short: Step-by-step tutorial for creating reusable custom parameter layers in Glazed
Topics:
- tutorial
- layers
- parameters
- custom
- logging
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Creating Custom Parameter Layers: Tutorial

This tutorial demonstrates how to create and use custom parameter layers in Glazed. We'll build a comprehensive logging layer that can be reused across multiple commands, providing consistent logging configuration throughout your application.

## What You'll Build

By the end of this tutorial, you'll have:
- A reusable logging layer with comprehensive configuration options
- A settings struct for type-safe parameter access
- Helper functions for easy integration
- A command that uses the logging layer
- Understanding of layer composition and reuse patterns

## Prerequisites

- Completed the [Build Your First Command](./build-first-command.md) tutorial
- Basic understanding of Go structs and interfaces
- Familiarity with Glazed's command system

## Step 1: Design Your Layer

Before writing code, let's plan what our logging layer should provide:

**Core Features:**
- Log level control (debug, info, warn, error, fatal)
- Output format selection (text, JSON)
- Log file specification
- Caller information toggle
- Verbose mode

**Integration Features:**
- Environment variable support
- Configuration file compatibility
- Validation and defaults
- Helper functions for easy setup

## Step 2: Create the Project Structure

```bash
mkdir glazed-logging-layer
cd glazed-logging-layer
go mod init glazed-logging-layer
go get github.com/go-go-golems/glazed
go get github.com/spf13/cobra
go get github.com/rs/zerolog
```

Create the following structure:
```
glazed-logging-layer/
├── main.go
├── logging/
│   ├── layer.go      # Layer definition
│   ├── settings.go   # Settings struct and helpers  
│   └── init.go       # Logger initialization
└── go.mod
```

## Step 3: Define the Settings Struct

Create `logging/settings.go`:

```go
package logging

import (
    "fmt"
    "io"
    "os"
    "strings"
    "time"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

// LoggingSettings represents all logging configuration options
type LoggingSettings struct {
    // Core logging settings
    Level      string `glazed.parameter:"log-level"`
    Format     string `glazed.parameter:"log-format"`
    File       string `glazed.parameter:"log-file"`
    
    // Advanced settings  
    WithCaller bool   `glazed.parameter:"with-caller"`
    Verbose    bool   `glazed.parameter:"verbose"`
    
    // Logstash integration (optional advanced feature)
    LogstashHost string `glazed.parameter:"logstash-host"`
    LogstashPort int    `glazed.parameter:"logstash-port"`
}

// Validate checks if the logging settings are valid
func (s *LoggingSettings) Validate() error {
    // Validate log level
    validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
    if !contains(validLevels, s.Level) {
        return fmt.Errorf("invalid log level '%s', must be one of: %s", 
            s.Level, strings.Join(validLevels, ", "))
    }
    
    // Validate log format
    validFormats := []string{"text", "json"}
    if !contains(validFormats, s.Format) {
        return fmt.Errorf("invalid log format '%s', must be one of: %s",
            s.Format, strings.Join(validFormats, ", "))
    }
    
    // Validate logstash port if host is specified
    if s.LogstashHost != "" && (s.LogstashPort < 1 || s.LogstashPort > 65535) {
        return fmt.Errorf("logstash port must be between 1 and 65535, got %d", s.LogstashPort)
    }
    
    return nil
}

// GetLogLevel converts string level to zerolog.Level
func (s *LoggingSettings) GetLogLevel() zerolog.Level {
    if s.Verbose {
        return zerolog.DebugLevel
    }
    
    switch strings.ToLower(s.Level) {
    case "debug":
        return zerolog.DebugLevel
    case "info":
        return zerolog.InfoLevel
    case "warn":
        return zerolog.WarnLevel
    case "error":
        return zerolog.ErrorLevel
    case "fatal":
        return zerolog.FatalLevel
    case "panic":
        return zerolog.PanicLevel
    default:
        return zerolog.InfoLevel
    }
}

// GetWriter returns the appropriate writer for log output
func (s *LoggingSettings) GetWriter() (io.Writer, error) {
    if s.File == "" {
        return os.Stderr, nil
    }
    
    file, err := os.OpenFile(s.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return nil, fmt.Errorf("failed to open log file '%s': %w", s.File, err)
    }
    
    return file, nil
}

// SetupLogger configures the global logger with these settings
func (s *LoggingSettings) SetupLogger() error {
    // Validate settings first
    if err := s.Validate(); err != nil {
        return err
    }
    
    // Set log level
    zerolog.SetGlobalLevel(s.GetLogLevel())
    
    // Get writer
    writer, err := s.GetWriter()
    if err != nil {
        return err
    }
    
    // Configure output format
    var output io.Writer = writer
    if s.Format == "text" {
        // Pretty console output for text format
        if s.File == "" { // Only if writing to stderr
            output = zerolog.ConsoleWriter{
                Out:        writer,
                TimeFormat: time.RFC3339,
                NoColor:    false,
            }
        }
    }
    
    // Create logger
    logger := zerolog.New(output).With().Timestamp()
    
    // Add caller information if requested
    if s.WithCaller {
        logger = logger.Caller()
    }
    
    // Set as global logger
    log.Logger = logger.Logger()
    
    return nil
}

// Helper function
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

## Step 4: Create the Parameter Layer

Create `logging/layer.go`:

```go
package logging

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

const (
    // LoggingSlug is the identifier for the logging layer
    LoggingSlug = "logging"
)

// NewLoggingLayer creates a new parameter layer for logging configuration
func NewLoggingLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        LoggingSlug,
        "Logging Configuration",
        layers.WithParameterDefinitions(
            // Core logging parameters
            parameters.NewParameterDefinition(
                "log-level",
                parameters.ParameterTypeChoice,
                parameters.WithHelp("Set the logging level"),
                parameters.WithDefault("info"),
                parameters.WithChoices("debug", "info", "warn", "error", "fatal", "panic"),
                parameters.WithShortFlag("L"),
            ),
            parameters.NewParameterDefinition(
                "log-format",
                parameters.ParameterTypeChoice,
                parameters.WithHelp("Set the log output format"),
                parameters.WithDefault("text"),
                parameters.WithChoices("text", "json"),
            ),
            parameters.NewParameterDefinition(
                "log-file",
                parameters.ParameterTypeString,
                parameters.WithHelp("Log file path (default: stderr)"),
                parameters.WithDefault(""),
            ),
            
            // Advanced parameters
            parameters.NewParameterDefinition(
                "with-caller",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Include caller information in log entries"),
                parameters.WithDefault(false),
            ),
            parameters.NewParameterDefinition(
                "verbose",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Enable verbose logging (sets level to debug)"),
                parameters.WithDefault(false),
                parameters.WithShortFlag("v"),
            ),
            
            // Optional advanced features
            parameters.NewParameterDefinition(
                "logstash-host",
                parameters.ParameterTypeString,
                parameters.WithHelp("Logstash server host for centralized logging"),
                parameters.WithDefault(""),
            ),
            parameters.NewParameterDefinition(
                "logstash-port",
                parameters.ParameterTypeInteger,
                parameters.WithHelp("Logstash server port"),
                parameters.WithDefault(5044),
            ),
        ),
    )
}

// NewLoggingLayerWithOptions creates a logging layer with customization options
func NewLoggingLayerWithOptions(opts ...LoggingLayerOption) (layers.ParameterLayer, error) {
    config := &loggingLayerConfig{
        includeLogstash: false,
        defaultLevel:    "info",
        defaultFormat:   "text",
    }
    
    // Apply options
    for _, opt := range opts {
        opt(config)
    }
    
    layer, err := NewLoggingLayer()
    if err != nil {
        return nil, err
    }
    
    // Modify defaults based on config
    params := layer.GetParameterDefinitions()
    
    if levelParam := params.Get("log-level"); levelParam != nil {
        levelParam.Default = config.defaultLevel
    }
    
    if formatParam := params.Get("log-format"); formatParam != nil {
        formatParam.Default = config.defaultFormat
    }
    
    // Remove logstash parameters if not included
    if !config.includeLogstash {
        layer.RemoveFlag("logstash-host")
        layer.RemoveFlag("logstash-port")
    }
    
    return layer, nil
}

// Configuration options for the logging layer
type loggingLayerConfig struct {
    includeLogstash bool
    defaultLevel    string
    defaultFormat   string
}

type LoggingLayerOption func(*loggingLayerConfig)

// WithLogstash includes Logstash configuration parameters
func WithLogstash() LoggingLayerOption {
    return func(c *loggingLayerConfig) {
        c.includeLogstash = true
    }
}

// WithDefaultLevel sets the default log level
func WithDefaultLevel(level string) LoggingLayerOption {
    return func(c *loggingLayerConfig) {
        c.defaultLevel = level
    }
}

// WithDefaultFormat sets the default log format
func WithDefaultFormat(format string) LoggingLayerOption {
    return func(c *loggingLayerConfig) {
        c.defaultFormat = format
    }
}
```

## Step 5: Create Helper Functions

Create `logging/init.go`:

```go
package logging

import (
    "fmt"

    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/rs/zerolog/log"
)

// GetLoggingSettings extracts logging settings from parsed layers
func GetLoggingSettings(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error) {
    settings := &LoggingSettings{}
    if err := parsedLayers.InitializeStruct(LoggingSlug, settings); err != nil {
        return nil, fmt.Errorf("failed to initialize logging settings: %w", err)
    }
    return settings, nil
}

// InitializeLogging sets up logging from parsed layers
func InitializeLogging(parsedLayers *layers.ParsedLayers) error {
    settings, err := GetLoggingSettings(parsedLayers)
    if err != nil {
        return err
    }
    
    if err := settings.SetupLogger(); err != nil {
        return fmt.Errorf("failed to setup logger: %w", err)
    }
    
    log.Debug().
        Str("level", settings.Level).
        Str("format", settings.Format).
        Str("file", settings.File).
        Bool("with_caller", settings.WithCaller).
        Bool("verbose", settings.Verbose).
        Msg("Logging initialized")
    
    return nil
}

// MustInitializeLogging sets up logging or panics on error
func MustInitializeLogging(parsedLayers *layers.ParsedLayers) {
    if err := InitializeLogging(parsedLayers); err != nil {
        panic(fmt.Sprintf("Failed to initialize logging: %v", err))
    }
}

// SetupDefaultLogging configures logging with default settings (useful for testing)
func SetupDefaultLogging() error {
    settings := &LoggingSettings{
        Level:      "info",
        Format:     "text", 
        File:       "",
        WithCaller: false,
        Verbose:    false,
    }
    return settings.SetupLogger()
}
```

## Step 6: Create a Command Using the Layer

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "glazed-logging-layer/logging"
    
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/go-go-golems/glazed/pkg/types"
    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
)

// ProcessDataCommand demonstrates using the logging layer
type ProcessDataCommand struct {
    *cmds.CommandDescription
}

type ProcessDataSettings struct {
    InputFile  string `glazed.parameter:"input-file"`
    OutputFile string `glazed.parameter:"output-file"`
    Workers    int    `glazed.parameter:"workers"`
    DryRun     bool   `glazed.parameter:"dry-run"`
}

func (c *ProcessDataCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Initialize logging first
    if err := logging.InitializeLogging(parsedLayers); err != nil {
        return fmt.Errorf("failed to initialize logging: %w", err)
    }
    
    log.Info().Msg("Starting data processing command")
    
    // Get command settings
    settings := &ProcessDataSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return err
    }
    
    log.Debug().
        Str("input_file", settings.InputFile).
        Str("output_file", settings.OutputFile).
        Int("workers", settings.Workers).
        Bool("dry_run", settings.DryRun).
        Msg("Command settings parsed")
    
    // Simulate processing
    if settings.DryRun {
        log.Info().Msg("Dry run mode - no actual processing")
    } else {
        log.Info().Msg("Starting actual data processing")
    }
    
    // Simulate some work with progress logging
    for i := 0; i < settings.Workers; i++ {
        log.Info().Int("worker_id", i).Msg("Starting worker")
        
        // Simulate processing time
        time.Sleep(100 * time.Millisecond)
        
        // Create result row
        row := types.NewRow(
            types.MRP("worker_id", i),
            types.MRP("status", "completed"),
            types.MRP("processed_items", (i+1)*10),
            types.MRP("duration_ms", 100),
            types.MRP("timestamp", time.Now().Format(time.RFC3339)),
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            log.Error().Err(err).Int("worker_id", i).Msg("Failed to add result row")
            return err
        }
        
        log.Debug().Int("worker_id", i).Msg("Worker completed")
    }
    
    log.Info().Msg("Data processing completed successfully")
    return nil
}

func NewProcessDataCommand() (*ProcessDataCommand, error) {
    // Create logging layer with custom options
    loggingLayer, err := logging.NewLoggingLayerWithOptions(
        logging.WithDefaultLevel("info"),
        logging.WithDefaultFormat("text"),
        // logging.WithLogstash(), // Uncomment to include Logstash options
    )
    if err != nil {
        return nil, err
    }
    
    // Create glazed layer for output formatting
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }
    
    cmdDesc := cmds.NewCommandDescription(
        "process-data",
        cmds.WithShort("Process data with configurable logging"),
        cmds.WithLong(`
Process data files with comprehensive logging support.
This command demonstrates how to use the custom logging layer.

Examples:
  process-data --input-file data.csv --workers 4
  process-data --input-file data.csv --log-level debug
  process-data --input-file data.csv --log-format json --log-file process.log
  process-data --input-file data.csv --verbose --with-caller
        `),
        
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "input-file",
                parameters.ParameterTypeString,
                parameters.WithHelp("Input file to process"),
                parameters.WithRequired(true),
                parameters.WithShortFlag("i"),
            ),
            parameters.NewParameterDefinition(
                "output-file",
                parameters.ParameterTypeString,
                parameters.WithHelp("Output file path"),
                parameters.WithDefault("output.processed"),
                parameters.WithShortFlag("o"),
            ),
            parameters.NewParameterDefinition(
                "workers",
                parameters.ParameterTypeInteger,
                parameters.WithHelp("Number of worker processes"),
                parameters.WithDefault(2),
                parameters.WithShortFlag("w"),
            ),
            parameters.NewParameterDefinition(
                "dry-run",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Perform a dry run without actual processing"),
                parameters.WithDefault(false),
            ),
        ),
        
        // Add both logging and glazed layers
        cmds.WithLayersList(loggingLayer, glazedLayer),
    )
    
    return &ProcessDataCommand{
        CommandDescription: cmdDesc,
    }, nil
}

var _ cmds.GlazeCommand = &ProcessDataCommand{}

// Second command to demonstrate layer reuse
type AnalyzeDataCommand struct {
    *cmds.CommandDescription
}

type AnalyzeDataSettings struct {
    DataFile   string `glazed.parameter:"data-file"`
    Algorithm  string `glazed.parameter:"algorithm"`
    Iterations int    `glazed.parameter:"iterations"`
}

func (c *AnalyzeDataCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Initialize logging (same layer, reused!)
    if err := logging.InitializeLogging(parsedLayers); err != nil {
        return fmt.Errorf("failed to initialize logging: %w", err)
    }
    
    log.Info().Msg("Starting data analysis command")
    
    settings := &AnalyzeDataSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return err
    }
    
    log.Info().
        Str("data_file", settings.DataFile).
        Str("algorithm", settings.Algorithm).
        Int("iterations", settings.Iterations).
        Msg("Analysis configuration")
    
    // Simulate analysis
    for i := 0; i < settings.Iterations; i++ {
        log.Debug().Int("iteration", i+1).Msg("Running analysis iteration")
        
        // Simulate some analysis work
        time.Sleep(50 * time.Millisecond)
        
        row := types.NewRow(
            types.MRP("iteration", i+1),
            types.MRP("algorithm", settings.Algorithm),
            types.MRP("accuracy", 0.85+float64(i)*0.01),
            types.MRP("processing_time_ms", 50),
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    log.Info().Msg("Analysis completed")
    return nil
}

func NewAnalyzeDataCommand() (*AnalyzeDataCommand, error) {
    // Reuse the same logging layer - this is the power of layers!
    loggingLayer, err := logging.NewLoggingLayer()
    if err != nil {
        return nil, err
    }
    
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }
    
    cmdDesc := cmds.NewCommandDescription(
        "analyze-data",
        cmds.WithShort("Analyze data with configurable logging"),
        cmds.WithLong("Analyze data files using various algorithms with the same logging configuration."),
        
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "data-file",
                parameters.ParameterTypeString,
                parameters.WithHelp("Data file to analyze"),
                parameters.WithRequired(true),
            ),
            parameters.NewParameterDefinition(
                "algorithm",
                parameters.ParameterTypeChoice,
                parameters.WithChoices("linear", "logistic", "random-forest", "neural-net"),
                parameters.WithDefault("linear"),
                parameters.WithHelp("Analysis algorithm to use"),
            ),
            parameters.NewParameterDefinition(
                "iterations",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(3),
                parameters.WithHelp("Number of analysis iterations"),
            ),
        ),
        
        cmds.WithLayersList(loggingLayer, glazedLayer),
    )
    
    return &AnalyzeDataCommand{
        CommandDescription: cmdDesc,
    }, nil
}

var _ cmds.GlazeCommand = &AnalyzeDataCommand{}

func main() {
    rootCmd := &cobra.Command{
        Use:   "data-processor",
        Short: "Data processing application with custom logging layer",
        Long:  "Demonstrates how to create and reuse custom parameter layers in Glazed",
    }
    
    // Create and register process command
    processCmd, err := NewProcessDataCommand()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating process command: %v\n", err)
        os.Exit(1)
    }
    
    cobraProcessCmd, err := cli.BuildCobraCommandFromGlazeCommand(processCmd)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error building process command: %v\n", err)
        os.Exit(1)
    }
    
    // Create and register analyze command  
    analyzeCmd, err := NewAnalyzeDataCommand()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating analyze command: %v\n", err)
        os.Exit(1)
    }
    
    cobraAnalyzeCmd, err := cli.BuildCobraCommandFromGlazeCommand(analyzeCmd)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error building analyze command: %v\n", err)
        os.Exit(1)
    }
    
    rootCmd.AddCommand(cobraProcessCmd, cobraAnalyzeCmd)
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## Step 7: Test Your Custom Layer

Build and test the application:

```bash
go build -o data-processor

# Test basic functionality
./data-processor process-data --help
./data-processor analyze-data --help

# Test different logging configurations
./data-processor process-data --input-file test.csv --workers 3

# Debug logging
./data-processor process-data --input-file test.csv --log-level debug

# JSON logging to file
./data-processor process-data --input-file test.csv --log-format json --log-file process.log

# Verbose mode with caller info
./data-processor process-data --input-file test.csv --verbose --with-caller

# Test analyze command (same logging options!)
./data-processor analyze-data --data-file test.csv --algorithm neural-net --log-level debug

# Combine with Glazed output options
./data-processor process-data --input-file test.csv --output json --fields worker_id,status,processed_items
```

## Step 8: Advanced Features

### Environment Variable Support

Add environment variable support by using middleware when running commands:

```go
// In your main.go, you could add:
import "github.com/go-go-golems/glazed/pkg/cmds/runner"

func runWithEnvironment() {
    cmd, _ := NewProcessDataCommand()
    
    parseOptions := []runner.ParseOption{
        runner.WithEnvMiddleware("DATAPROC_"), // Loads DATAPROC_LOG_LEVEL, etc.
    }
    
    ctx := context.Background()
    err := runner.ParseAndRun(ctx, cmd, parseOptions, nil)
    if err != nil {
        log.Fatal().Err(err).Msg("Command failed")
    }
}
```

### Configuration File Support

```go
parseOptions := []runner.ParseOption{
    runner.WithEnvMiddleware("DATAPROC_"),
    runner.WithViper(), // Loads from config file
}
```

### Layer Composition

Create specialized layers by combining the logging layer with others:

```go
func NewDatabaseLayerWithLogging() ([]layers.ParameterLayer, error) {
    loggingLayer, err := logging.NewLoggingLayer()
    if err != nil {
        return nil, err
    }
    
    dbLayer, err := database.NewDatabaseLayer()
    if err != nil {
        return nil, err
    }
    
    return []layers.ParameterLayer{loggingLayer, dbLayer}, nil
}
```

## What You've Learned

Congratulations! You've successfully created a comprehensive custom parameter layer. Here's what you've accomplished:

### Layer Creation
- **Parameter Definition**: Created parameters with appropriate types, defaults, and validation
- **Settings Struct**: Used struct tags for type-safe parameter access
- **Validation**: Implemented parameter validation with clear error messages
- **Configuration Options**: Added options for customizing layer behavior

### Integration Patterns
- **Helper Functions**: Created functions for easy settings extraction and logger initialization
- **Error Handling**: Implemented robust error handling with context
- **Layer Reuse**: Demonstrated how the same layer can be used across multiple commands
- **Composition**: Showed how layers can be combined for complex applications

### Advanced Features
- **Conditional Parameters**: Used options to include/exclude parameters
- **Default Customization**: Allowed customization of default values
- **Middleware Support**: Prepared the layer for environment and config file integration
- **Logging Integration**: Connected the layer to actual logging functionality

## Best Practices You've Applied

1. **Single Responsibility**: The logging layer only handles logging configuration
2. **Type Safety**: Used struct tags and validation for safe parameter access
3. **Clear Naming**: Used descriptive parameter names with consistent patterns
4. **Good Defaults**: Provided sensible defaults that work out of the box
5. **Documentation**: Added comprehensive help text for all parameters
6. **Error Handling**: Implemented proper error wrapping and context
7. **Reusability**: Created a layer that can be used across different commands

## Next Steps

Now that you understand custom layers, you can:

1. **Create More Layers**: Build layers for database connections, API configurations, etc.
2. **Layer Libraries**: Create packages of reusable layers for your organization
3. **Advanced Validation**: Add more sophisticated validation logic
4. **Dynamic Configuration**: Create layers that adapt based on runtime conditions
5. **Integration Testing**: Write tests for your layers and their interactions

## Common Patterns for Other Layers

### Database Layer
```go
type DatabaseSettings struct {
    Host     string `glazed.parameter:"db-host"`
    Port     int    `glazed.parameter:"db-port"`
    Name     string `glazed.parameter:"db-name"`
    Username string `glazed.parameter:"db-username"`
    Password string `glazed.parameter:"db-password"`
    SSLMode  string `glazed.parameter:"db-ssl-mode"`
}
```

### HTTP Client Layer
```go
type HTTPSettings struct {
    BaseURL    string        `glazed.parameter:"base-url"`
    Timeout    time.Duration `glazed.parameter:"timeout"`
    RetryCount int           `glazed.parameter:"retry-count"`
    UserAgent  string        `glazed.parameter:"user-agent"`
    APIKey     string        `glazed.parameter:"api-key"`
}
```

### File Processing Layer
```go
type FileSettings struct {
    InputDir     string   `glazed.parameter:"input-dir"`
    OutputDir    string   `glazed.parameter:"output-dir"`
    FilePattern  string   `glazed.parameter:"file-pattern"`
    Extensions   []string `glazed.parameter:"extensions"`
    Recursive    bool     `glazed.parameter:"recursive"`
    OverwriteOK  bool     `glazed.parameter:"overwrite"`
}
```

You now have the knowledge to create powerful, reusable parameter layers that make your Glazed applications more maintainable and consistent!
