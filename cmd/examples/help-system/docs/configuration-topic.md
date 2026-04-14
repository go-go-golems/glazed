---
Title: Configuration System
Slug: configuration-topic
Short: Understanding the configuration system and available options
SectionType: GeneralTopic
Topics:
- configuration
- settings
- yaml
Commands:
- help
Flags:
- --config-file
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
Order: 3
---

# Configuration System

The configuration system allows you to customize application behavior through YAML files and command-line flags.

## Configuration File Structure

```yaml
database:
  url: "postgresql://localhost/mydb"
  timeout: 30s

output:
  format: "json"
  pretty: true

logging:
  level: "info"
  file: "/var/log/app.log"
```

## Working With Configuration

Applications built on Glazed typically combine configuration files, environment variables, and command-line flags. A common workflow is:

```bash
# Point a command at an explicit config file
myapp run --config-file ./config.yaml

# Override individual values via flags
myapp run --output-format json
```

## Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Local configuration file
4. Global configuration file (lowest priority)
