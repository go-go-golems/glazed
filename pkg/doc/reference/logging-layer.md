---
Title: Logging Layer API Reference
Slug: logging-layer-reference
Short: Comprehensive logging configuration for CLI applications
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

# Logging Layer API Reference

## Overview

The Glazed logging layer provides comprehensive logging configuration for CLI applications through command-line parameters, environment variables, and configuration files. The layer handles setup for console output, file logging, and centralized log aggregation while supporting multiple output formats and verbosity levels.

### Key Capabilities

- **Multiple output targets**: Console, file, and Logstash integration
- **Structured logging**: JSON and text formats with contextual fields
- **Automatic configuration**: Single function call for complete setup
- **Production features**: Log rotation, centralized collection, performance optimization

## Architecture

```mermaid
graph TD
    A[CLI Parameters] --> B[Logging Layer]
    B --> C[Console Output]
    B --> D[File Output]
    B --> E[Logstash/ELK]
    
    C --> F[Human-readable text]
    D --> G[Rotating log files]
    E --> H[Centralized monitoring]
```

The logging layer transforms command-line parameters into configured log outputs, supporting development, testing, and production deployment scenarios.

## Implementation

### Basic Integration

Add the logging layer to any Glazed command:

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    "github.com/rs/zerolog/log"
)

func NewMyCommand() (*MyCommand, error) {
    loggingLayer, err := logging.NewLoggingLayer()
    if err != nil {
        return nil, fmt.Errorf("failed to create logging layer: %w", err)
    }
    
    cmdDesc := cmds.NewCommandDescription(
        "my-command",
        cmds.WithShort("Command with logging support"),
        cmds.WithLayersList(loggingLayer),
    )
    
    return &MyCommand{CommandDescription: cmdDesc}, nil
}

func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
        return err
    }
    
    log.Info().Msg("Processing started")
    // Command implementation
    return nil
}
```

### Structured Logging Patterns

Use structured fields for effective logging:

```go
func processFile(fileName string) error {
    start := time.Now()
    
    log.Debug().
        Str("file", fileName).
        Msg("Starting file processing")
    
    data, err := os.ReadFile(fileName)
    if err != nil {
        log.Error().
            Str("file", fileName).
            Err(err).
            Msg("Failed to read file")
        return fmt.Errorf("reading file %s: %w", fileName, err)
    }
    
    log.Info().
        Str("file", fileName).
        Int("bytes_processed", len(data)).
        Dur("duration", time.Since(start)).
        Msg("File processed successfully")
    
    return nil
}
```

### Contextual Loggers

Create loggers with persistent context for complex operations:

```go
func processUser(userID string) error {
    userLogger := log.With().
        Str("user_id", userID).
        Str("operation", "user_processing").
        Logger()
    
    userLogger.Info().Msg("Starting user processing")
    
    if err := validateUser(userID); err != nil {
        userLogger.Error().
            Err(err).
            Msg("User validation failed")
        return err
    }
    
    userLogger.Info().Msg("User processing completed")
    return nil
}
```

## Configuration Parameters

### Command-Line Flags

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--log-level` | choice | `info` | Verbosity level (trace, debug, info, warn, error, fatal) |
| `--log-format` | choice | `text` | Output format (text, json) |
| `--log-file` | string | `""` | Output file path with automatic rotation |
| `--with-caller` | bool | `false` | Include source file and line number |
| `--log-to-stdout` | bool | `false` | Force output to stdout regardless of other settings |
| `--logstash-enabled` | bool | `false` | Enable Logstash output |
| `--logstash-host` | string | `""` | Logstash server hostname |
| `--logstash-port` | int | `5044` | Logstash server port |
| `--logstash-protocol` | string | `tcp` | Connection protocol |
| `--logstash-app-name` | string | `""` | Application identifier in logs |
| `--logstash-environment` | string | `""` | Environment tag (dev, staging, prod) |

### Log Levels

| Level | Purpose | Usage |
|-------|---------|-------|
| `trace` | Extremely detailed execution flow | Performance debugging |
| `debug` | Detailed diagnostic information | Development troubleshooting |
| `info` | General application progress | Default production level |
| `warn` | Unexpected conditions that don't halt execution | Monitoring degraded performance |
| `error` | Error conditions with continued execution | Problem tracking |
| `fatal` | Critical errors requiring application termination | System failures |

### Output Formats

**Text format** (human-readable):
```
INFO  2023-12-07T10:30:00Z Processing started
DEBUG 2023-12-07T10:30:01Z File loaded file=data.csv rows=1000
```

**JSON format** (machine-readable):
```json
{"level":"info","time":"2023-12-07T10:30:00Z","message":"Processing started"}
{"level":"debug","time":"2023-12-07T10:30:01Z","message":"File loaded","file":"data.csv","rows":1000}
```

## Production Deployment

### File Logging

Configure automatic log rotation for production:

```bash
myapp process-data \
    --log-level info \
    --log-format json \
    --log-file /var/log/myapp/application.log
```

File logging features:
- **Rotation**: 10MB maximum file size
- **Retention**: 3 backup files, 28-day retention
- **Thread safety**: Atomic writes for concurrent operations

### Centralized Logging

Send logs directly to ELK stack:

```bash
myapp process-data \
    --log-level info \
    --log-format json \
    --logstash-host logs.company.com \
    --logstash-port 5044 \
    --logstash-app-name myapp \
    --logstash-environment production
```

### Environment Variables

Configure logging through environment variables:

```bash
export MYAPP_LOG_LEVEL=info
export MYAPP_LOG_FORMAT=json
export MYAPP_LOG_FILE=/var/log/myapp.log
export MYAPP_LOGSTASH_HOST=logs.company.com
export MYAPP_LOGSTASH_PORT=5044
```

## API Reference

### Core Types

#### LoggingSettings

```go
type LoggingSettings struct {
    WithCaller          bool   `glazed.parameter:"with-caller"`
    LogLevel            string `glazed.parameter:"log-level"`
    LogFormat           string `glazed.parameter:"log-format"`
    LogFile             string `glazed.parameter:"log-file"`
    LogToStdout         bool   `glazed.parameter:"log-to-stdout"`
    LogstashEnabled     bool   `glazed.parameter:"logstash-enabled"`
    LogstashHost        string `glazed.parameter:"logstash-host"`
    LogstashPort        int    `glazed.parameter:"logstash-port"`
    LogstashProtocol    string `glazed.parameter:"logstash-protocol"`
    LogstashAppName     string `glazed.parameter:"logstash-app-name"`
    LogstashEnvironment string `glazed.parameter:"logstash-environment"`
}
```

### Functions

#### SetupLoggingFromParsedLayers

```go
func SetupLoggingFromParsedLayers(parsedLayers *layers.ParsedLayers) error
```

Configures global logger from command-line parameters. Call early in command execution.

**Usage**:
```go
if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
    return fmt.Errorf("failed to setup logging: %w", err)
}
```

#### GetLoggingSettings

```go
func GetLoggingSettings(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error)
```

Extracts logging configuration for custom validation or setup.

**Usage**:
```go
settings, err := logging.GetLoggingSettings(parsedLayers)
if err != nil {
    return err
}

// Custom validation
if settings.LogstashHost != "" && settings.LogstashPort <= 0 {
    return fmt.Errorf("invalid logstash configuration")
}

return logging.SetupLogging(settings)
```

#### NewLoggingLayer

```go
func NewLoggingLayer() (layers.ParameterLayer, error)
```

Creates parameter layer for command definitions.

**Usage**:
```go
loggingLayer, err := logging.NewLoggingLayer()
if err != nil {
    return nil, err
}

cmdDesc := cmds.NewCommandDescription(
    "my-command",
    cmds.WithLayersList(loggingLayer),
)
```

## Performance Optimization

### Conditional Expensive Operations

Avoid performance penalties for debug logging:

```go
// Efficient: Check if debug is enabled before expensive operations
if log.Debug().Enabled() {
    expensiveData := calculateComplexDebuggingInfo()
    log.Debug().
        Interface("debug_data", expensiveData).
        Msg("Detailed debug information")
}

// Always efficient: Simple field logging
log.Info().
    Str("user", userID).
    Int("count", itemCount).
    Dur("elapsed", duration).
    Msg("Operation completed")
```

## Configuration Examples

### Development Configuration

```yaml
logging:
  log-level: debug
  log-format: text
  with-caller: true
```

### Production Configuration

```yaml
logging:
  log-level: info
  log-format: json
  log-file: /var/log/application.log
  logstash-enabled: true
  logstash-host: logs.company.com
  logstash-port: 5044
  logstash-app-name: myapp
  logstash-environment: production
```

## Common Issues

### No Log Output

**Symptoms**: Application runs but produces no log output

**Solutions**:
1. Verify `SetupLoggingFromParsedLayers` is called before logging
2. Check log level filtering: use `--log-level debug`
3. Verify file permissions for `--log-file` destinations

### Incorrect Format

**Symptoms**: Unexpected output format (JSON instead of text or vice versa)

**Solutions**:
1. Explicitly set format: `--log-format json` or `--log-format text`
2. Check environment variables: `LOG_FORMAT` may override settings
3. Verify configuration file format settings

### Logstash Connection Failures

**Symptoms**: Cannot connect to Logstash server

**Solutions**:
1. Test network connectivity: `telnet logstash-host 5044`
2. Verify firewall rules and network policies
3. Ensure Logstash TCP input configuration
4. Check DNS resolution for hostnames

## See Also

- [Custom Layer Tutorial](../tutorials/custom-layer.md): Creating custom parameter layers
- [Layers Guide](../topics/layers-guide.md): Parameter layer system overview
- [Commands Guide](../topics/commands-guide.md): Building CLI commands
- [Configuration Guide](../topics/configuration-guide.md): Advanced configuration patterns
