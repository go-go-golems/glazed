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
- config
- set
- get
Flags:
- --config-file
- --global
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

## Setting Configuration

Use the config command to manage settings:

```bash
# Set a value
config set database.url "postgresql://user:pass@localhost/db"

# Get a value
config get database.url

# Use global configuration
config set --global output.format json
```

## Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Local configuration file
4. Global configuration file (lowest priority)
