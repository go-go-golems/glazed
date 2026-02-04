# Clay Logging Section

This package provides a Glazed field section for configuring logging in Clay applications.

## Documentation

**📖 For API reference and detailed usage**, see: [Logging Section API Reference](../../doc/reference/logging-section.md)

**🎓 To learn how to create custom sections**, see: [Custom Section Tutorial](../../doc/tutorials/custom-section.md)

## Quick Overview

The logging section provides:
- **Log Level**: Control verbosity (`debug`, `info`, `warn`, `error`, `fatal`)
- **Log Format**: Choose between text and JSON formats  
- **Log File**: Specify output file (defaults to stderr)
- **Caller Info**: Include caller information in logs
- **Logstash Integration**: Send logs to centralized servers

## Quick Usage

```go
import "github.com/go-go-golems/glazed/pkg/cmds/logging"

// In your command's Run method:
if err := logging.SetupLoggingFromValues(parsedSections); err != nil {
    return err
}

log.Info().Msg("Logging is now configured!")
```

For complete examples and detailed API documentation, see the links above. 