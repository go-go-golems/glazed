---
Title: Logging Layer API Reference
Slug: logging-layer-reference
Short: API reference for the Clay logging layer implementation in Glazed
Topics:
- reference
- logging
- api
- clay
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Reference
---

# Clay Logging Layer API Reference

## Overview

The Clay logging layer provides a Glazed parameter layer for configuring logging in Clay applications. It offers a standard way to configure logging across applications with type-safe settings, Cobra integration, and simplified initialization.

## Package Structure

```
pkg/cmds/logging/
├── layer.go          # Parameter layer definition
├── init.go           # Logger initialization functions  
├── init-logging.go   # Logger setup utilities
└── logstash_writer.go # Logstash integration
```

## Core Types

### LoggingSettings

The main configuration struct for logging settings:

```go
type LoggingSettings struct {
    Level        string `glazed.parameter:"log-level"`
    Format       string `glazed.parameter:"log-format"`
    File         string `glazed.parameter:"log-file"`
    WithCaller   bool   `glazed.parameter:"with-caller"`
    Verbose      bool   `glazed.parameter:"verbose"`
    LogstashHost string `glazed.parameter:"logstash-host"`
    LogstashPort int    `glazed.parameter:"logstash-port"`
}
```

**Fields:**
- `Level`: Log level (debug, info, warn, error, fatal)
- `Format`: Output format (text, json)
- `File`: Log file path (empty = stderr)
- `WithCaller`: Include caller information in logs
- `Verbose`: Enable verbose logging (sets level to debug)
- `LogstashHost`: Logstash server hostname
- `LogstashPort`: Logstash server port

### LoggingParameterLayer

The parameter layer implementation that provides logging configuration flags:

```go
type LoggingParameterLayer struct {
    layers.ParameterLayerImpl
}
```

## Layer Creation Functions

### NewLoggingParameterLayer

Creates a new logging parameter layer with default configuration:

```go
func NewLoggingParameterLayer(options ...layers.ParameterLayerOptions) (*LoggingParameterLayer, error)
```

**Parameters:**
- `options`: Optional layer configuration options

**Returns:**
- `*LoggingParameterLayer`: Configured logging layer
- `error`: Creation error if any

**Example:**
```go
loggingLayer, err := logging.NewLoggingParameterLayer()
if err != nil {
    return nil, fmt.Errorf("failed to create logging layer: %w", err)
}
```

### NewLoggingParameterLayers

Creates logging parameter layers for use with Glazed commands:

```go
func NewLoggingParameterLayers(options ...LoggingParameterLayerOption) (*layers.ParameterLayers, error)
```

**Parameters:**
- `options`: Optional logging-specific layer options

**Returns:**
- `*layers.ParameterLayers`: Parameter layers collection
- `error`: Creation error if any

## Configuration Functions

### GetLoggingSettings

Extracts logging settings from parsed parameter layers:

```go
func GetLoggingSettings(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error)
```

**Parameters:**
- `parsedLayers`: Parsed parameter layers from command execution

**Returns:**
- `*LoggingSettings`: Extracted logging configuration
- `error`: Extraction error if any

**Example:**
```go
settings, err := logging.GetLoggingSettings(parsedLayers)
if err != nil {
    return fmt.Errorf("failed to get logging settings: %w", err)
}
```

### InitializeStructFromLoggingSettings

Initializes a struct with logging settings:

```go
func InitializeStructFromLoggingSettings(settings *LoggingSettings, s interface{}) error
```

**Parameters:**
- `settings`: Source logging settings
- `s`: Target struct to initialize

**Returns:**
- `error`: Initialization error if any

## Logger Initialization

### SetupLogging

Configures the global logger based on logging settings:

```go
func SetupLogging(settings *LoggingSettings) error
```

**Parameters:**
- `settings`: Logging configuration settings

**Returns:**
- `error`: Setup error if any

**Example:**
```go
settings, err := logging.GetLoggingSettings(parsedLayers)
if err != nil {
    return err
}

if err := logging.SetupLogging(settings); err != nil {
    return fmt.Errorf("failed to setup logging: %w", err)
}
```

### SetupLoggingFromParsedLayers

Convenience function that extracts settings and sets up logging in one call:

```go
func SetupLoggingFromParsedLayers(parsedLayers *layers.ParsedLayers) error
```

**Parameters:**
- `parsedLayers`: Parsed parameter layers

**Returns:**
- `error`: Setup error if any

**Example:**
```go
if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
    return fmt.Errorf("logging setup failed: %w", err)
}
```

## Logstash Integration

### LogstashWriter

Writer implementation for sending logs to Logstash:

```go
type LogstashWriter struct {
    conn net.Conn
    host string
    port int
}
```

### NewLogstashWriter

Creates a new Logstash writer:

```go
func NewLogstashWriter(host string, port int) (*LogstashWriter, error)
```

**Parameters:**
- `host`: Logstash server hostname
- `port`: Logstash server port

**Returns:**
- `*LogstashWriter`: Configured Logstash writer
- `error`: Connection error if any

### Write

Implements io.Writer interface for sending data to Logstash:

```go
func (w *LogstashWriter) Write(p []byte) (n int, err error)
```

**Parameters:**
- `p`: Data to write

**Returns:**
- `n`: Number of bytes written
- `err`: Write error if any

### Close

Closes the Logstash connection:

```go
func (w *LogstashWriter) Close() error
```

**Returns:**
- `error`: Close error if any

## Parameter Definitions

The logging layer provides the following command-line parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--log-level` | Choice | `"info"` | Log level (debug, info, warn, error, fatal) |
| `--log-format` | Choice | `"text"` | Log format (text, json) |
| `--log-file` | String | `""` | Log file path (empty = stderr) |
| `--with-caller` | Bool | `false` | Include caller information |
| `--verbose` | Bool | `false` | Enable verbose logging |
| `--logstash-host` | String | `""` | Logstash server host |
| `--logstash-port` | Int | `5044` | Logstash server port |

## Integration Examples

### Basic Usage with Glazed Command

```go
package main

import (
    "context"
    
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/rs/zerolog/log"
)

type MyCommand struct {
    *cmds.CommandDescription
}

func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Initialize logging
    if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
        return err
    }
    
    log.Info().Msg("Command started")
    
    // Command logic here
    
    log.Info().Msg("Command completed")
    return nil
}

func NewMyCommand() (*MyCommand, error) {
    loggingLayer, err := logging.NewLoggingParameterLayer()
    if err != nil {
        return nil, err
    }
    
    cmdDesc := cmds.NewCommandDescription(
        "my-command",
        cmds.WithShort("Example command with logging"),
        cmds.WithLayersList(loggingLayer),
    )
    
    return &MyCommand{
        CommandDescription: cmdDesc,
    }, nil
}
```

### Advanced Usage with Custom Configuration

```go
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Get logging settings for custom handling
    settings, err := logging.GetLoggingSettings(parsedLayers)
    if err != nil {
        return err
    }
    
    // Custom validation
    if settings.LogstashHost != "" && settings.LogstashPort == 0 {
        return fmt.Errorf("logstash port must be specified when host is provided")
    }
    
    // Setup logging
    if err := logging.SetupLogging(settings); err != nil {
        return err
    }
    
    // Use specific logger features
    logger := log.With().
        Str("command", "my-command").
        Str("version", "1.0.0").
        Logger()
    
    logger.Info().Msg("Command started with custom logger")
    
    return nil
}
```

### Environment Variable Integration

The logging layer works with Glazed's middleware system for environment variable support:

```bash
# Environment variables (when using env middleware with prefix "MYAPP_")
export MYAPP_LOG_LEVEL=debug
export MYAPP_LOG_FORMAT=json
export MYAPP_LOG_FILE=/var/log/myapp.log
export MYAPP_WITH_CALLER=true
export MYAPP_VERBOSE=true
```

### Configuration File Integration

Example YAML configuration:

```yaml
logging:
  log-level: debug
  log-format: json
  log-file: /var/log/app.log
  with-caller: true
  verbose: false
  logstash-host: logstash.example.com
  logstash-port: 5044
```

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `failed to create logging layer` | Layer creation failure | Check parameter definitions |
| `failed to get logging settings` | Settings extraction failure | Verify layer is included in command |
| `failed to setup logging` | Logger configuration failure | Check file permissions and paths |
| `invalid log level` | Invalid level parameter | Use valid levels: debug, info, warn, error, fatal |
| `logstash connection failed` | Network connectivity issue | Verify Logstash host and port |

### Error Handling Best Practices

```go
// Always wrap errors with context
if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
    return fmt.Errorf("logging initialization failed: %w", err)
}

// Validate settings before use
settings, err := logging.GetLoggingSettings(parsedLayers)
if err != nil {
    return fmt.Errorf("failed to extract logging settings: %w", err)
}

if settings.LogstashHost != "" && settings.LogstashPort <= 0 {
    return fmt.Errorf("invalid logstash configuration: host=%s port=%d", 
        settings.LogstashHost, settings.LogstashPort)
}

// Use defer for cleanup
if settings.LogstashHost != "" {
    writer, err := logging.NewLogstashWriter(settings.LogstashHost, settings.LogstashPort)
    if err != nil {
        return fmt.Errorf("failed to create logstash writer: %w", err)
    }
    defer writer.Close()
}
```

## Performance Considerations

### File Logging Performance
- Log files are opened in append mode with appropriate permissions
- Consider log rotation for long-running applications
- Use JSON format for better parsing performance in log aggregation systems

### Logstash Integration Performance
- Logstash connections are TCP-based and may have network latency
- Consider buffering for high-throughput applications
- Implement retry logic for connection failures

### Memory Usage
- Text format uses more memory for formatting
- JSON format is more compact and efficient for machine processing
- Caller information adds overhead - use only when debugging

## Thread Safety

The logging layer is thread-safe when used with zerolog:
- Global logger configuration is atomic
- Individual log operations are thread-safe
- Logstash writer uses synchronized writes

## Migration Guide

### From Custom Logging to Clay Logging Layer

1. **Replace custom logging flags** with the logging layer
2. **Update initialization code** to use `SetupLoggingFromParsedLayers`
3. **Migrate configuration** to use standard parameter names
4. **Update log calls** to use zerolog instead of custom loggers

### Example Migration

**Before:**
```go
type MyCommand struct {
    LogLevel string
    LogFile  string
}

func (c *MyCommand) setupLogging() {
    // Custom logging setup
}
```

**After:**
```go
type MyCommand struct {
    *cmds.CommandDescription
}

func (c *MyCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    // Use standard logging layer
    if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
        return err
    }
    // Command logic...
}
```

## See Also

- [Custom Layer Tutorial](../tutorials/custom-layer.md): Learn to create custom layers
- [Layers Guide](../topics/layers-guide.md): Understanding the layer system
- [Commands Reference](../topics/commands-reference.md): Command system documentation
- [Middlewares Guide](../topics/middlewares-guide.md): Parameter loading from various sources
