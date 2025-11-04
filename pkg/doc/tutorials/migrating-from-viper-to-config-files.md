---
Title: Migrating from Viper to Config Files
Slug: migrating-from-viper-to-config-files
Short: Step-by-step guide to migrate existing Glazed applications from Viper-based configuration to the new config file middleware system
Topics:
- tutorial
- migration
- configuration
- middlewares
- viper
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Migrating from Viper to Config Files

Glazed has moved away from Viper-based configuration parsing to a more explicit, traceable config file middleware system. This migration guide walks you through updating existing applications to use the new approach, which provides better observability, deterministic precedence, and cleaner separation between config sources.

The new system replaces Viper's automatic config discovery and merging with explicit file loading middlewares that record each parse step. This makes it clear where each parameter value originated and enables better debugging with `--print-parsed-parameters`.

## Overview of Changes

The migration involves three main areas:

1. **Config File Loading**: Replace `GatherFlagsFromViper()` and `GatherFlagsFromCustomViper()` with `LoadParametersFromFile()` or `LoadParametersFromFiles()`
2. **Logging Initialization**: Move from `InitLoggerFromViper()` to `InitLoggerFromCobra()` or `SetupLoggingFromParsedLayers()`
3. **Cobra Integration**: Use `CobraParserConfig` to wire config discovery, environment variables, and file loading into your commands

## ⚠️ Critical: Config File Changes Required

**Two breaking changes require immediate attention:**

### 1. Config File Discovery No Longer Automatic

**Before (Viper):** Automatic discovery in standard paths:
```go
viper.AddConfigPath("$HOME/.myapp")
viper.AddConfigPath("/etc/myapp")
viper.ReadInConfig()  // Searches automatically
```

**After:** Explicit discovery required:
```go
// Option A: Use ResolveAppConfigPath helper
configPath, err := appconfig.ResolveAppConfigPath("myapp", "")
// Searches: $XDG_CONFIG_HOME/myapp, $HOME/.myapp, /etc/myapp

// Option B: Use ConfigFilesFunc in CobraParserConfig (recommended)
cli.WithParserConfig(cli.CobraParserConfig{
    AppName: "myapp",  // Enables automatic discovery
    ConfigFilesFunc: resolver,
})
```

**Action required:** Every application using Viper config discovery must add explicit config file resolution (see Step 4).

### 2. Config File Format Must Match Layer Structure

**Before (Viper):** Config structure was flexible - Viper read any keys:
```yaml
# Flat structure that Viper handled
api-key: "secret"
threshold: 42
log-level: "debug"
```

**After:** Config must match layer names and parameters:
```yaml
# Layer names as top-level keys
demo:
  api-key: "secret"
  threshold: 42
logging:
  log-level: "debug"
```

**If your config doesn't match this structure:**

**Option A: Restructure your config files** (simplest)
- Group parameters under layer names
- Update parameter names to match definitions

**Option B: Use pattern-based mapping** (for legacy configs)
```go
mapper, _ := patternmapper.NewConfigMapper(layers,
    patternmapper.MappingRule{
        Source:          "api-key",  // Flat config
        TargetLayer:     "demo",
        TargetParameter: "api-key",
    },
)
middlewares.LoadParametersFromFile("config.yaml",
    middlewares.WithConfigMapper(mapper))
```

**Option C: Use custom mapper function** (for complex transformations)
```go
mapper := func(raw interface{}) (map[string]map[string]interface{}, error) {
    // Transform your config to layer format
    return map[string]map[string]interface{}{
        "demo": {"api-key": raw["api-key"]},
    }, nil
}
middlewares.LoadParametersFromFile("config.yaml",
    middlewares.WithConfigFileMapper(mapper))
```

**Action required:** Audit your config files and either restructure them or add a mapper (see Step 5).

## Step 1: Replace Viper Config Middleware

The primary change is replacing Viper-based middleware with explicit config file middlewares. The old approach relied on Viper's automatic config discovery and merging, while the new approach gives you explicit control over which files are loaded and in what order.

### Before: Using GatherFlagsFromViper

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func GetCommandMiddlewares(cmd *cobra.Command) []middlewares.Middleware {
    return []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.GatherFlagsFromViper(
            parameters.WithParseStepSource("viper"),
        ),
        middlewares.SetFromDefaults(),
    }
}
```

Viper would automatically:
- Search standard config paths (`$HOME/.app`, `/etc/app`, etc.)
- Load config files based on app name
- Merge environment variables
- Bind command flags

### After: Using LoadParametersFromFile

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func GetCommandMiddlewares(cmd *cobra.Command) []middlewares.Middleware {
    return []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.UpdateFromEnv("APP"),  // Explicit env prefix
        middlewares.LoadParametersFromFile("config.yaml",
            middlewares.WithParseOptions(
                parameters.WithParseStepSource("config"),
            ),
        ),
        middlewares.SetFromDefaults(),
    }
}
```

The new approach:
- Requires explicit file paths (no magic discovery)
- Separates environment variable handling from file loading
- Records each config file as a distinct parse step
- Applies files in the order you specify (low → high precedence)

### Single Config File

For applications with a single config file, use `LoadParametersFromFile`:

```go
middlewares.LoadParametersFromFile("/etc/myapp/config.yaml",
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)
```

The config file must match the default structure (layer names as top-level keys):

```yaml
demo:
  api-key: "secret123"
  threshold: 42
```

### Multiple Config Files (Overlays)

For applications that compose configuration from multiple files, use `LoadParametersFromFiles`:

```go
middlewares.LoadParametersFromFiles([]string{
    "base.yaml",
    "env.yaml", 
    "local.yaml",
}, middlewares.WithParseOptions(
    parameters.WithParseStepSource("config"),
))
```

Files are applied in order (low → high precedence), and each file is recorded as a separate parse step. The last file's values win.

## Step 2: Replace Custom Viper Instances

If you were using `GatherFlagsFromCustomViper` to load configuration from specific files or other applications, replace it with explicit file loading middlewares.

### Before: Using GatherFlagsFromCustomViper

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
)

func GetAdvancedMiddlewares(commandSettings *cli.GlazedCommandSettings) []middlewares.Middleware {
    return []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        
        // Profile-specific override file
        middlewares.GatherFlagsFromCustomViper(
            middlewares.WithConfigFile(
                fmt.Sprintf("/etc/myapp/%s.yaml", commandSettings.Profile),
            ),
            middlewares.WithParseOptions(
                parameters.WithParseStepSource("profile-overrides"),
            ),
        ),
        
        // Shared configuration from another app
        middlewares.GatherFlagsFromCustomViper(
            middlewares.WithAppName("shared-config"),
            middlewares.WithParseOptions(
                parameters.WithParseStepSource("shared"),
            ),
        ),
        
        middlewares.SetFromDefaults(),
    }
}
```

### After: Using LoadParametersFromFiles

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func GetAdvancedMiddlewares(commandSettings *cli.GlazedCommandSettings) []middlewares.Middleware {
    files := []string{}
    
    // Profile-specific override file
    if commandSettings.Profile != "" {
        files = append(files, fmt.Sprintf("/etc/myapp/%s.yaml", commandSettings.Profile))
    }
    
    // Shared configuration (if you have explicit file paths)
    // Note: Cross-app config sharing now requires explicit file paths
    if sharedConfigPath := os.Getenv("SHARED_CONFIG_PATH"); sharedConfigPath != "" {
        files = append(files, sharedConfigPath)
    }
    
    return []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.LoadParametersFromFiles(files,
            middlewares.WithParseOptions(
                parameters.WithParseStepSource("config"),
            ),
        ),
        middlewares.SetFromDefaults(),
    }
}
```

**Key differences:**
- No more automatic discovery based on app names
- Explicit file paths required
- You control the exact order and sources
- Cross-app config sharing requires explicit file paths (no `WithAppName`)

## Step 3: Update Logging Initialization

Logging initialization has been simplified to work directly with Cobra flags instead of requiring Viper binding.

### Before: Viper-Based Logging

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
    Use: "myapp",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // Initialize logger after Cobra parsed flags and Viper loaded config
        err := logging.InitLoggerFromViper()
        cobra.CheckErr(err)
    },
}

func main() {
    // Add logging layer
    err := logging.AddLoggingLayerToRootCommand(rootCmd, "myapp")
    cobra.CheckErr(err)
    
    // Bind flags to Viper before initializing logger
    err = viper.BindPFlags(rootCmd.PersistentFlags())
    cobra.CheckErr(err)
    
    // Initialize logger early
    err = logging.InitLoggerFromViper()
    cobra.CheckErr(err)
    
    _ = rootCmd.Execute()
}
```

### After: Cobra-Based Logging

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use: "myapp",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Initialize logger after Cobra parsed flags
        return logging.InitLoggerFromCobra(cmd)
    },
}

func main() {
    // Add logging flags to root command
    _ = logging.AddLoggingLayerToRootCommand(rootCmd, "myapp")
    
    // ... register commands, help system, etc.
    _ = rootCmd.Execute()
}
```

**Key changes:**
- No Viper binding required
- Single initialization point in `PersistentPreRunE`
- Logging reads directly from Cobra flags
- Simpler setup with fewer moving parts

### Alternative: Initialize from Parsed Layers

If you're using Glazed's middleware system and want logging to respect config file values, initialize from parsed layers instead:

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
)

func runCommand(cmd *cobra.Command, args []string) error {
    // ... setup layers and parse ...
    
    err := middlewares.ExecuteMiddlewares(layers_, parsed,
        middlewares.LoadParametersFromFile("config.yaml"),
        middlewares.UpdateFromEnv("APP"),
        middlewares.ParseFromCobraCommand(cmd),
    )
    if err != nil {
        return err
    }
    
    // Initialize logging from parsed layers (includes config file values)
    err = logging.SetupLoggingFromParsedLayers(parsed)
    if err != nil {
        return err
    }
    
    // ... rest of command logic ...
}
```

## Step 4: Update Cobra Command Setup

For CLI applications, use `CobraParserConfig` to wire config discovery, environment variables, and file loading. This replaces manual Viper setup and middleware chaining.

### Before: Manual Viper Setup

```go
import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/spf13/viper"
)

func buildCommand() (*cobra.Command, error) {
    // ... create command description ...
    
    cobraCmd := cli.NewCobraCommandFromCommandDescription(desc)
    
    // Manual Viper setup
    viper.SetEnvPrefix("MYAPP")
    viper.AddConfigPath("$HOME/.myapp")
    viper.AddConfigPath("/etc/myapp")
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.ReadInConfig()
    viper.BindPFlags(cobraCmd.Flags())
    
    return cobraCmd, nil
}
```

### After: Using CobraParserConfig

```go
import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    appconfig "github.com/go-go-golems/glazed/pkg/config"
)

func buildCommand() (*cobra.Command, error) {
    // ... create command description ...
    
    cobraCmd, err := cli.BuildCobraCommandFromCommand(command,
        cli.WithParserConfig(cli.CobraParserConfig{
            AppName: "myapp",  // Enables env prefix MYAPP_ and config discovery
            ConfigFilesFunc: func(parsed *layers.ParsedLayers, cmd *cobra.Command, args []string) ([]string, error) {
                // Use explicit path if provided via --config-file
                cs := &cli.CommandSettings{}
                _ = parsed.InitializeStruct(cli.CommandSettingsSlug, cs)
                if cs.ConfigFile != "" {
                    return []string{cs.ConfigFile}, nil
                }
                
                // Otherwise discover config file
                configPath, err := appconfig.ResolveAppConfigPath("myapp", "")
                if err != nil {
                    return nil, nil  // No config file found, continue without
                }
                return []string{configPath}, nil
            },
        }),
    )
    
    return cobraCmd, err
}
```

**Benefits:**
- `AppName` automatically enables environment variable prefix (`MYAPP_`)
- `ConfigFilesFunc` gives you full control over file discovery
- `--config-file` flag is automatically available via `command-settings` layer
- Config files integrate cleanly with env vars and flags

### Simple Config Path

For simpler cases, you can specify a single config path directly:

```go
cobraCmd, err := cli.BuildCobraCommandFromCommand(command,
    cli.WithParserConfig(cli.CobraParserConfig{
        AppName:    "myapp",
        ConfigPath: "/etc/myapp/config.yaml",  // Explicit path
    }),
)
```

## Step 5: Handle Custom Config Structures

If your config files don't match the default layer structure, you have two options: pattern-based mapping (declarative) or custom mapper functions (programmatic).

### Pattern-Based Mapping (Recommended)

Use pattern-based mapping when you can describe your config structure with patterns. This keeps mapping logic declarative and testable:

```go
import (
    pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
)

// Define mapping rules
mapper, err := pm.NewConfigMapper(layers_,
    pm.MappingRule{
        Source:          "app.settings.api_key",
        TargetLayer:     "demo",
        TargetParameter: "api-key",
    },
    pm.MappingRule{
        Source:          "app.{env}.api_key",
        TargetLayer:     "demo",
        TargetParameter: "{env}-api-key",
    },
)

// Use with LoadParametersFromFile
middleware := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigMapper(mapper),
)
```

**When to use:**
- Config structure is predictable and can be described with patterns
- You want declarative, testable mapping rules
- Multiple environments or tenants share similar structures

### Custom Mapper Functions

Use custom mapper functions when you need full control over config transformation:

```go
mapper := func(rawConfig interface{}) (map[string]map[string]interface{}, error) {
    configMap := rawConfig.(map[string]interface{})
    result := map[string]map[string]interface{}{
        "demo": make(map[string]interface{}),
    }
    
    // Transform config structure to layer format
    if apiKey, ok := configMap["api_key"]; ok {
        result["demo"]["api-key"] = apiKey
    }
    
    // Handle nested structures, arrays, validation, etc.
    // ...
    
    return result, nil
}

middleware := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigFileMapper(mapper),
)
```

**When to use:**
- Complex transformations that patterns can't express
- Conditional logic based on config values
- Cross-field validation or derivation
- Integration with external config formats

## Step 6: Update Middleware Execution Order

The precedence order remains the same, but the way you express it changes. Remember: middlewares execute in reverse order (last middleware runs first).

### Correct Precedence Order

```go
middlewares.ExecuteMiddlewares(layers_, parsed,
    middlewares.SetFromDefaults(),                              // Lowest priority
    middlewares.LoadParametersFromFiles([]string{               // Config files (low → high)
        "base.yaml",
        "env.yaml", 
        "local.yaml",
    }),
    middlewares.UpdateFromEnv("APP"),                           // Environment variables
    middlewares.GatherArguments(args),                          // Positional arguments
    middlewares.ParseFromCobraCommand(cmd),                     // Flags (highest priority)
)
```

**Precedence:** Defaults < Config Files (low→high) < Env < Args < Flags

Each source overrides the previous ones, and each config file in the list overrides earlier files.

## Step 7: Remove Viper Dependencies

After migrating, you can remove Viper-related code:

1. **Remove Viper imports:**
   ```go
   // Remove these
   import "github.com/spf13/viper"
   ```

2. **Remove Viper initialization:**
   ```go
   // Remove calls like:
   viper.SetEnvPrefix("APP")
   viper.AddConfigPath("...")
   viper.ReadInConfig()
   viper.BindPFlags(...)
   ```

3. **Remove deprecated middleware calls:**
   ```go
   // Remove:
   middlewares.GatherFlagsFromViper(...)
   middlewares.GatherFlagsFromCustomViper(...)
   ```

4. **Update logging initialization:**
   ```go
   // Replace:
   logging.InitLoggerFromViper()
   
   // With:
   logging.InitLoggerFromCobra(cmd)
   // or
   logging.SetupLoggingFromParsedLayers(parsed)
   ```

## Common Migration Patterns

### Pattern 1: Single Config File with Discovery

**Before:**
```go
viper.SetConfigName("config")
viper.AddConfigPath("$HOME/.myapp")
viper.AddConfigPath("/etc/myapp")
viper.ReadInConfig()
```

**After:**
```go
import appconfig "github.com/go-go-golems/glazed/pkg/config"

configPath, err := appconfig.ResolveAppConfigPath("myapp", "")
if err == nil {
    middlewares.LoadParametersFromFile(configPath)
}
```

### Pattern 2: Profile-Based Config Files

**Before:**
```go
middlewares.GatherFlagsFromCustomViper(
    middlewares.WithConfigFile(
        fmt.Sprintf("/etc/myapp/%s.yaml", profile),
    ),
)
```

**After:**
```go
files := []string{}
if profile != "" {
    files = append(files, fmt.Sprintf("/etc/myapp/%s.yaml", profile))
}
middlewares.LoadParametersFromFiles(files)
```

### Pattern 3: Environment Variable Overrides

**Before:**
Viper automatically merged environment variables based on prefix and key naming.

**After:**
```go
// Explicit env prefix
middlewares.UpdateFromEnv("APP")  // Reads APP_* variables
```

Environment variable names follow the pattern: `{PREFIX}_{LAYER}_{PARAMETER}` (e.g., `APP_DEMO_API_KEY` for `demo.api-key`).

### Pattern 4: Config File Override Pattern

**Before:**
Manual Viper config file merging with custom precedence.

**After:**
```go
resolver := func(parsed *layers.ParsedLayers, _ *cobra.Command, _ []string) ([]string, error) {
    cs := &cli.CommandSettings{}
    _ = parsed.InitializeStruct(cli.CommandSettingsSlug, cs)
    files := []string{}
    
    if cs.ConfigFile != "" {
        files = append(files, cs.ConfigFile)
        
        // Add override file if it exists
        dir := filepath.Dir(cs.ConfigFile)
        base := filepath.Base(cs.ConfigFile)
        stem := strings.TrimSuffix(base, filepath.Ext(base))
        override := filepath.Join(dir, fmt.Sprintf("%s.override.yaml", stem))
        if _, err := os.Stat(override); err == nil {
            files = append(files, override)
        }
    }
    
    return files, nil
}
```

## Debugging and Validation

### Inspect Parse Steps

Use `--print-parsed-parameters` to see exactly where each parameter value came from:

```bash
myapp command --print-parsed-parameters
```

This shows the full parse history for each parameter, including which config file set each value.

### Validate Config Files

Before applying config files, validate them against your layer definitions:

```go
import (
    "os"
    "gopkg.in/yaml.v3"
)

func validateConfigFile(layers_ *layers.ParameterLayers, path string) error {
    b, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    var raw map[string]interface{}
    if err := yaml.Unmarshal(b, &raw); err != nil {
        return err
    }
    
    // Check each layer and parameter
    for layerSlug, v := range raw {
        layer, ok := layers_.Get(layerSlug)
        if !ok {
            return fmt.Errorf("unknown layer: %s", layerSlug)
        }
        
        kv, ok := v.(map[string]interface{})
        if !ok {
            return fmt.Errorf("layer %s must be an object", layerSlug)
        }
        
        pds := layer.GetParameterDefinitions()
        for key, val := range kv {
            pd, ok := pds.Get(key)
            if !ok {
                return fmt.Errorf("unknown parameter %s.%s", layerSlug, key)
            }
            
            if _, err := pd.CheckValueValidity(val); err != nil {
                return fmt.Errorf("invalid value for %s.%s: %v", layerSlug, key, err)
            }
        }
    }
    
    return nil
}
```

## Troubleshooting

### Config File Not Found

If your config file isn't being loaded, check:
1. File path is correct and file exists
2. `ConfigFilesFunc` returns the path (or `ConfigPath` is set)
3. File has correct permissions

### Environment Variables Not Working

If environment variables aren't being read:
1. Check the prefix matches your `AppName` (e.g., `MYAPP_` for `AppName: "myapp"`)
2. Variable names follow `{PREFIX}_{LAYER}_{PARAMETER}` format
3. `UpdateFromEnv` middleware is included in your middleware chain

### Precedence Issues

If values aren't overriding as expected:
1. Verify middleware order (last middleware runs first)
2. Check config file order in `LoadParametersFromFiles` (low → high)
3. Use `--print-parsed-parameters` to see actual precedence

### Legacy Config Format

If you have legacy config files that don't match the layer structure:
1. Use pattern-based mapping for structured transformations
2. Use custom mapper functions for complex transformations
3. Consider migrating config files to the new format over time

## Complete Example

Here's a complete example showing a before and after migration:

### Before

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
    Use: "myapp",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        err := logging.InitLoggerFromViper()
        cobra.CheckErr(err)
    },
}

func main() {
    err := logging.AddLoggingLayerToRootCommand(rootCmd, "myapp")
    cobra.CheckErr(err)
    
    viper.SetEnvPrefix("MYAPP")
    viper.AddConfigPath("$HOME/.myapp")
    viper.AddConfigPath("/etc/myapp")
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.ReadInConfig()
    viper.BindPFlags(rootCmd.PersistentFlags())
    
    err = logging.InitLoggerFromViper()
    cobra.CheckErr(err)
    
    _ = rootCmd.Execute()
}

func GetMiddlewares(cmd *cobra.Command) []middlewares.Middleware {
    return []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.GatherFlagsFromViper(),
        middlewares.SetFromDefaults(),
    }
}
```

### After

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    appconfig "github.com/go-go-golems/glazed/pkg/config"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use: "myapp",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return logging.InitLoggerFromCobra(cmd)
    },
}

func main() {
    _ = logging.AddLoggingLayerToRootCommand(rootCmd, "myapp")
    
    // ... register commands with BuildCobraCommandFromCommand ...
    
    _ = rootCmd.Execute()
}

func buildCommand() (*cobra.Command, error) {
    // ... create command description ...
    
    return cli.BuildCobraCommandFromCommand(command,
        cli.WithParserConfig(cli.CobraParserConfig{
            AppName: "myapp",
            ConfigFilesFunc: func(parsed *layers.ParsedLayers, cmd *cobra.Command, args []string) ([]string, error) {
                cs := &cli.CommandSettings{}
                _ = parsed.InitializeStruct(cli.CommandSettingsSlug, cs)
                if cs.ConfigFile != "" {
                    return []string{cs.ConfigFile}, nil
                }
                
                configPath, err := appconfig.ResolveAppConfigPath("myapp", "")
                if err != nil {
                    return nil, nil  // No config file, continue without
                }
                return []string{configPath}, nil
            },
        }),
    )
}
```

## Next Steps

After completing the migration:

1. **Test thoroughly**: Verify all config sources work correctly
2. **Use `--print-parsed-parameters`**: Confirm precedence is as expected
3. **Update documentation**: Document your config file locations and formats
4. **Remove Viper**: Clean up any remaining Viper dependencies
5. **Consider validation**: Add config file validation to catch errors early

For more details on the new config system, see:
- `glaze help config-files` - Config files and overlays guide
- `glaze help pattern-based-config-mapping` - Pattern-based mapping guide
- `glaze help cmds-middlewares` - Middleware system reference

