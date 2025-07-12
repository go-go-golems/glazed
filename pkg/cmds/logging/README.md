# Clay Logging Layer

This package provides a Glazed parameter layer for configuring logging in Clay applications.

## Documentation

**📖 For API reference and detailed usage**, see: [Logging Layer API Reference](../../doc/reference/logging-layer.md)

**🎓 To learn how to create custom layers**, see: [Custom Layer Tutorial](../../doc/tutorials/custom-layer.md)

## Quick Overview

The logging layer provides:
- **Log Level**: Control verbosity (`debug`, `info`, `warn`, `error`, `fatal`)
- **Log Format**: Choose between text and JSON formats  
- **Log File**: Specify output file (defaults to stderr)
- **Caller Info**: Include caller information in logs
- **Logstash Integration**: Send logs to centralized servers

## Quick Usage

```go
import "github.com/go-go-golems/glazed/pkg/cmds/logging"

// In your command's Run method:
if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
    return err
}

log.Info().Msg("Logging is now configured!")
```

For complete examples and detailed API documentation, see the links above. 