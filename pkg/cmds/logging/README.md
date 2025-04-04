# Clay Logging Layer

This package provides a Glazed parameter layer for configuring logging in Clay applications.

## Overview

The logging layer provides a standard way to configure logging in Clay applications, using Glazed parameter layers.
It offers the following advantages:

- Consistent logging configuration across applications
- Type-safe settings through a structured settings object
- Easy integration with Cobra commands
- Simplified initialization process

## Features

The logging layer supports the following configuration options:

- **Log Level**: Control the verbosity of logging (`debug`, `info`, `warn`, `error`, `fatal`)
- **Log Format**: Choose between text (human-readable) and JSON formats
- **Log File**: Specify a file to write logs to (defaults to stderr)
- **With Caller**: Include caller information in log entries
- **Verbose**: Enable verbose logging (sets log level to debug)
- **Logstash Integration**: Send logs to a Logstash server for centralized logging

## Usage

### Basic Usage with Glazed and Cobra

```go
package main

import (
    "github.com/go-go-golems/clay/pkg"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/spf13/cobra"
)

func main() {
    // Create your Glazed command
    myCmd := cmds.NewCommandDescription(
        "my-command",
        cmds.WithShort("My example command"),
    )
    
    // Build a Cobra command with the logging layer
    cobraCmd, err := pkg.BuildCobraCommandWithLogging(myCmd)
    if err != nil {
        // handle error
    }
    
    // Add to root command
    rootCmd := &cobra.Command{
        Use:   "my-app",
        Short: "My application",
    }
    rootCmd.AddCommand(cobraCmd)
    
    // Initialize Viper and logging
    err = pkg.InitViper("my-app", rootCmd)
    if err != nil {
        // handle error
    }
    
    // Execute command
    rootCmd.Execute()
}
```

### Accessing Logging Settings in a Command

To access the logging settings in your command implementation:

```go
import (
    "context"
    "github.com/go-go-golems/clay/pkg/logging"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/rs/zerolog/log"
)

func (c *MyCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Get logging settings from parsed layers
    loggingSettings, err := logging.GetLoggingSettingsFromParsedLayers(parsedLayers)
    if err != nil {
        return err
    }
    
    // Now you can use loggingSettings or just use the zerolog logger
    log.Debug().Interface("settings", loggingSettings).Msg("Debug message")
    log.Info().Msg("Info message")
    
    return nil
}
```

### Manual Configuration

If you need to configure logging manually:

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
)

func configureLogging() error {
    settings := &logging.LoggingSettings{
        LogLevel:         "debug",
        LogFormat:        "text",
        LogFile:          "/var/log/myapp.log",
        WithCaller:       true,
        LogstashEnabled:  true,
        LogstashHost:     "logstash.example.com",
        LogstashPort:     5044,
        LogstashProtocol: "tcp",
        AppName:          "my-application",
        Environment:      "production",
    }
    
    return logging.InitLoggerFromSettings(settings)
}
```

### Configuring Logstash Integration via CLI

The logging layer automatically adds Logstash-related flags to your application, which can be used as follows:

```bash
# Enable Logstash integration
$ ./myapp --logstash-enabled --logstash-host logstash.example.com --logstash-port 5044 --app-name "my-application" --environment production

# Other Logstash options (with their defaults)
# --logstash-protocol tcp     # Protocol to use (tcp or udp)
# --app-name ""               # Application name in Logstash logs
# --environment development   # Environment name in Logstash logs
```

When Logstash integration is enabled, logs are sent to both the standard output and the Logstash server. The logs sent to Logstash include additional metadata like application name, environment, and timestamp.

## Complete Example

See the `examples/logging_layer_example.go` file for a complete example of using the logging layer in a Glazed command. 

For a dedicated example of Logstash integration, see `examples/logstash_example.go`. 