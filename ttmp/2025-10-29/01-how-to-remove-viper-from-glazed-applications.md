### How to remove Viper from Glazed applications (and replace with middlewares)

This document explains how Viper is currently integrated in Glazed- and Clay-based CLIs, and proposes a concrete design to replace Viper-driven flag/env/config parsing with the standard Glazed middleware system described in `glazed/pkg/doc/topics/21-cmds-middlewares.md`, while preserving the same environment-variable naming semantics.

## 1) Current system overview

### 1.1 Where Viper is wired into applications

- Clay provides `InitViper(...)` helpers used by apps like Pinocchio to initialize logging flags on the root `cobra.Command`, bind flags to Viper, read config files, and enable env var parsing with a hyphen→underscore replacer and an app prefix.

```14:44:clay/pkg/init.go
func InitViperWithAppName(appName string, configFile string) error {
    viper.SetEnvPrefix(appName)

    if configFile != "" {
        viper.SetConfigFile(configFile)
        viper.SetConfigType("yaml")
    } else {
        viper.SetConfigType("yaml")
        viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", appName))
        viper.AddConfigPath(fmt.Sprintf("/etc/%s", appName))

        xdgConfigPath, err := os.UserConfigDir()
        if err == nil {
            viper.AddConfigPath(fmt.Sprintf("%s/%s", xdgConfigPath, appName))
        }
    }

    // Read the configuration file into Viper
    err := viper.ReadInConfig()
    // if the file does not exist, continue normally
    if _, ok := err.(viper.ConfigFileNotFoundError); ok {
        // Config file not found; ignore error
    } else if err != nil {
        // Config file was found but another error was produced
        return err
    }
    viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
    viper.AutomaticEnv()

    return nil
}
```

```46:79:clay/pkg/init.go
func InitViper(appName string, rootCmd *cobra.Command) error {
    err := logging.AddLoggingLayerToRootCommand(rootCmd, appName)
    if err != nil {
        return err
    }

    // parse the flags one time just to catch --config
    configFile := ""
    for idx, arg := range os.Args {
        if arg == "--config" {
            if len(os.Args) > idx+1 {
                configFile = os.Args[idx+1]
            }
        }
    }

    err = InitViperWithAppName(appName, configFile)
    if err != nil {
        return err
    }

    // Bind the variables to the command-line flags
    err = viper.BindPFlags(rootCmd.PersistentFlags())
    if err != nil {
        return err
    }

    err = logging.InitLoggerFromViper()
    if err != nil {
        return err
    }

    return nil
}
```

- Example application (Pinocchio) uses both Viper-backed logging and Clay’s `InitViper` during boot, and reads app settings directly from Viper:

```47:58:pinocchio/cmd/pinocchio/main.go
var rootCmd = &cobra.Command{
    Use:     "pinocchio",
    Short:   "pinocchio is a tool to run LLM applications",
    Version: version,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        err := logging.InitLoggerFromViper()
        if err != nil {
            return err
        }
        return nil
    },
}
```

```130:136:pinocchio/cmd/pinocchio/main.go
err = clay.InitViper("pinocchio", rootCmd)
cobra.CheckErr(err)

rootCmd.AddCommand(runCommandCmd)
rootCmd.AddCommand(pinocchio_cmds.NewCodegenCommand())
return helpSystem, nil
```

```138:148:pinocchio/cmd/pinocchio/main.go
repositoryPaths := viper.GetStringSlice("repositories")

defaultDirectory := "$HOME/.pinocchio/prompts"
repositoryPaths = append(repositoryPaths, defaultDirectory)

loader := &cmds.PinocchioCommandLoader{}

directories := []repositories.Directory{
    {
        FS:               promptsFS,
        RootDirectory:    "prompts",
        RootDocDirectory: "prompts/doc",
        Name:             "pinocchio",
        SourcePrefix:     "embed",
    }}
```

### 1.2 Where Viper populates Glazed parameters

- The middleware `GatherFlagsFromViper` bridges Viper into Glazed’s `ParsedLayers` by reading parameter definitions per layer (respecting each layer’s flag prefix) and updating parsed values from Viper’s keyspace.

```95:146:glazed/pkg/cmds/middlewares/cobra.go
// GatherFlagsFromViper creates a middleware that loads parameter values from Viper configuration.
// This middleware is useful for integrating Viper-based configuration management with Glazed commands.
//
// It iterates through each layer, gathering flags from Viper for all parameters in that layer.
//
// Usage:
//
//  middleware := middlewares.GatherFlagsFromViper(sources.WithSource("viper"))
func GatherFlagsFromViper(options ...parameters.ParseStepOption) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(layers_ *schema.Schema, parsedLayers *values.Values) error {

            err := next(layers_, parsedLayers)
            if err != nil {
                return err
            }
            err = layers_.ForEachE(func(key string, l schema.Section) error {
                options_ := append([]parameters.ParseStepOption{
                    sources.WithSource("viper"),
                    sources.WithMetadata(map[string]interface{}{
                        "layer":          l.GetName(),
                        "layer_slug":     l.GetSlug(),
                        "layer_prefix":   l.GetPrefix(),
                        "registered_key": key,
                    }),
                }, options...)

                parsedLayer := parsedLayers.GetOrCreate(l)
                parameterDefinitions := l.GetParameterDefinitions()
                prefix := l.GetPrefix()

                ps, err := parameterDefinitions.GatherFlagsFromViper(true, prefix, options_...)
                if err != nil {
                    return err
                }

                _, err = parsedLayer.Parameters.Merge(ps)
                if err != nil {
                    return err
                }

                return nil
            })

            if err != nil {
                return err
            }

            return nil
        }
    }
}
```

- The underlying parameter-level Viper reader encodes the de-facto env/flag mapping semantics: key = `layerPrefix + flag-name`, and an env key shape comment is added to parsed metadata using uppercased, hyphen→underscore conversion.

```9:29:glazed/pkg/cmds/parameters/viper.go
func (pds *ParameterDefinitions) GatherFlagsFromViper(
    onlyProvided bool,
    prefix string,
    options ...ParseStepOption,
) (*ParsedParameters, error) {
    ret := NewParsedParameters()

    for v := pds.Oldest(); v != nil; v = v.Next() {
        p := v.Value

        parsed := &ParsedParameter{
            ParameterDefinition: p,
        }

        flagName := prefix + p.Name
        if onlyProvided && !viper.IsSet(flagName) {
            continue
        }
```

```40:49:glazed/pkg/cmds/parameters/viper.go
// Add metadata about the viper key and derived env key shape
upperKey := strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
meta := map[string]interface{}{
    "flag":    flagName,
    "env_key": upperKey,
}
options := append([]ParseStepOption{
    WithParseStepMetadata(meta),
    WithParseStepSource("viper"),
}, options...)
```

- Cobra parser injects Viper-based gathering both for command settings and (optionally) for full commands via custom runners:

```31:61:glazed/pkg/cli/cobra-parser.go
func CobraCommandDefaultMiddlewares(
    parsedCommandLayers *values.Values,
    cmd *cobra.Command,
    args []string,
) ([]cmd_middlewares.Middleware, error) {
    commandSettings := &CommandSettings{}
    err := parsedCommandLayers.InitializeStruct(CommandSettingsSlug, commandSettings)
    if err != nil {
        return nil, err
    }

    middlewares_ := []cmd_middlewares.Middleware{
        cmd_sources.FromCobra(cmd,
            sources.WithSource("cobra"),
        ),
        cmd_sources.FromArgs(args,
            sources.WithSource("arguments"),
        ),
    }

    if commandSettings.LoadParametersFromFile != "" {
        middlewares_ = append(middlewares_,
            cmd_sources.FromFile(commandSettings.LoadParametersFromFile))
    }

    middlewares_ = append(middlewares_,
        cmd_sources.FromDefaults(sources.WithSource(sources.SourceDefaults)),
    )

    return middlewares_, nil
}
```

```228:269:glazed/pkg/cli/cobra-parser.go
// ParseGlazedCommandLayer parses the global glazed settings from the given cobra.Command, if not nil,
// and from the configured viper config file.
func ParseCommandSettingsLayer(cmd *cobra.Command) (*values.Values, error) {
    parsedLayers := values.New()
    commandSettingsLayer, err := NewCommandSettingsLayer()
    if err != nil {
        return nil, err
    }
    
    profileSettingsLayer, err := NewProfileSettingsLayer()
    if err != nil {
        return nil, err
    }
    
    createCommandSettingsLayer, err := NewCreateCommandSettingsLayer()
    if err != nil {
        return nil, err
    }
    
    commandSettingsLayers := schema.NewSchema(
        layers.WithLayers(
            commandSettingsLayer,
            profileSettingsLayer,
            createCommandSettingsLayer,
        ),
    )
    
    // Parse the glazed command settings from the cobra command and config file
    middlewares_ := []cmd_middlewares.Middleware{}
    
    if cmd != nil {
        middlewares_ = append(middlewares_, cmd_sources.FromCobra(cmd, sources.WithSource("cobra")))
    }
    
    middlewares_ = append(middlewares_, cmd_middlewares.GatherFlagsFromViper(sources.WithSource("viper")))
    
    err = cmd_sources.Execute(commandSettingsLayers, parsedLayers, middlewares_...)
    if err != nil {
        return nil, err
    }
    
    return parsedLayers, nil
}
```

### 1.3 Logging is initialized from Viper today

Glazed exposes `InitLoggerFromViper()` which reads required logging keys from the Viper keyspace; apps invoke it in `PersistentPreRun(E)` and often once at startup.

```125:141:glazed/pkg/cmds/logging/init.go
// InitLoggerFromViper initializes the logger using settings from Viper
func InitLoggerFromViper() error {
    settings := &LoggingSettings{
        LogLevel:            viper.GetString("log-level"),
        LogFile:             viper.GetString("log-file"),
        LogFormat:           viper.GetString("log-format"),
        WithCaller:          viper.GetBool("with-caller"),
        LogToStdout:         viper.GetBool("log-to-stdout"),
        LogstashEnabled:     viper.GetBool("logstash-enabled"),
        LogstashHost:        viper.GetString("logstash-host"),
        LogstashPort:        viper.GetInt("logstash-port"),
        LogstashProtocol:    viper.GetString("logstash-protocol"),
        LogstashAppName:     viper.GetString("logstash-app-name"),
        LogstashEnvironment: viper.GetString("logstash-environment"),
    }

    return InitLoggerFromSettings(settings)
}
```

Note: there is also `SetupLoggingFromParsedLayers(parsedLayers)` that can initialize logging directly from Glazed parsed parameters (no Viper), which we can reuse in the new design.

### 1.4 Existing env parsing middleware (gap vs Viper)

There is an `UpdateFromEnv(prefix)` middleware today, but its env key derivation differs from the Viper semantics and also updates parsed parameters using the env key as the parameter name:

```148:160:glazed/pkg/cmds/middlewares/update.go
name := p.Name
if prefix != "" {
    name = prefix + "_" + name
}
name = strings.ToUpper(name)

if v, ok := os.LookupEnv(name); ok {
    err := parsedLayer.Parameters.UpdateValue(name, p, v, options...)
    if err != nil {
        return err
    }
}
```

Viper’s env semantics are: `ENV_PREFIX + '_' + UPPERCASE(REPLACE_ALL(layerPrefix + paramName, '-', '_'))`, and parsed parameter keys remain the logical parameter names (without env prefix) inside their layer. We will align `UpdateFromEnv` to that.

## 2) Design: Replace Viper with middlewares for config + env parsing

### 2.1 Goals

- Remove Viper from parameter parsing. Use Glazed middlewares only: `LoadParametersFromFile`, `UpdateFromEnv`, `ParseFromCobraCommand`, `SetFromDefaults`.
- Preserve env naming semantics that users rely on today:
  - `ENV_PREFIX + '_' + UPPER(REPLACE_ALL(layerPrefix + paramName, '-', '_'))`
  - Example: app `pinocchio`, layer prefix `openai-`, param `api-key` → `PINOcCHIO_OPENAI_API_KEY` (actually `PINOCCHIO_OPENAI_API_KEY`).
- Replace `--load-parameters-from-file` with automatic config discovery and loading via middleware.
- Keep logging working without Viper by using `SetupLoggingFromParsedLayers`.

### 2.2 New default middleware chain

Use this chain by default for Cobra commands (order listed here is low→high precedence; execution is reverse order):

1) `SetFromDefaults(sources.WithSource("defaults"))`
2) `LoadParametersFromFile(resolvedConfigPath, sources.WithSource("config"))`
3) `UpdateFromEnv(appEnvPrefix, sources.WithSource("env"))`
4) `ParseFromCobraCommand(cmd, sources.WithSource("flags"))`

Because middlewares execute in reverse order, flags override env, env overrides config, and config overrides defaults.

### 2.3 Config discovery (replaces `--load-parameters-from-file` and Viper config search)

Introduce a small helper that mirrors Clay’s `InitViperWithAppName` search logic without Viper:

- `ResolveAppConfigPath(appName string, explicit string) (string, error)`
  - If `explicit != ""`, use it.
  - Else: check, in order: `$XDG_CONFIG_HOME/<appName>/config.yaml`, `$HOME/.<appName>/config.yaml`, `/etc/<appName>/config.yaml`.
  - If not found, return empty string (middleware simply no-ops).

We can place this in `glazed/pkg/cli/appconfig/appconfig.go` or `glazed/pkg/config/resolve.go` and reuse also from the config editing commands.

### 2.4 Align env var mapping semantics in UpdateFromEnv

Update `UpdateFromEnv` to compute env keys exactly like the Viper path used today, while storing parsed values under the parameter’s logical name:

```go
// New behavior (proposed)
// For each layer and ParameterDefinition p:
key := l.GetPrefix() + p.Name                             // e.g. "openai-" + "api-key"
envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_")) // OPENAI_API_KEY
if appEnvPrefix != "" {                                     // e.g. "PINOCCHIO"
    envKey = appEnvPrefix + "_" + envKey                   // PINOCCHIO_OPENAI_API_KEY
}
if v, ok := os.LookupEnv(envKey); ok {
    // IMPORTANT: store under the logical parameter key (p.Name),
    // not the envKey, to remain consistent with how flags/config map.
    err := parsedLayer.Parameters.UpdateValue(p.Name, p, v,
        sources.WithSource("env"),
        sources.WithMetadata(map[string]interface{}{
            "env_key": envKey,
        }),
    )
    if err != nil { return err }
}
```

This fixes two issues in the current middleware: it replaces hyphens by underscores and uses the layer prefix, and it no longer writes values under the env key name.

### 2.5 Cobra integration (no Viper)

Update the Cobra parser hooks to remove Viper calls and inject the chain from §2.2.

- In `glazed/pkg/cli/cobra-parser.go`:
  - `ParseCommandSettingsLayer`: drop `GatherFlagsFromViper(...)`. Optionally add `UpdateFromEnv("GLAZED", ...)` so command-settings can be influenced by env if desired, and then apply defaults. If we keep a minimal `--load-parameters-from-file` for this internal layer during transition, parse it via flags (already handled by `ParseFromCobraCommand`).
  - `CobraCommandDefaultMiddlewares`: extend to include config and env middlewares based on a config resolver and app env prefix provided via `CobraParserConfig`.

Example shape (conceptual):

```go
// New fields (example) in CobraParserConfig
type CobraParserConfig struct {
    MiddlewaresFunc CobraMiddlewaresFunc
    ShortHelpLayers []string
    // New:
    AppName        string // used to derive default config path + env prefix
    ConfigPath     string // explicit path (overrides resolver), may come from a root flag
}

func CobraCommandDefaultMiddlewares(parsed *values.Values, cmd *cobra.Command, args []string) ([]middlewares.Middleware, error) {
    cfg := /* read from parser instance */
    configPath, _ := appconfig.ResolveAppConfigPath(cfg.AppName, cfg.ConfigPath)
    return []middlewares.Middleware{
        sources.FromCobra(cmd, sources.WithSource("flags")),
        sources.FromArgs(args, sources.WithSource("arguments")),
        sources.FromFile(configPath, sources.WithSource("config")),
        sources.FromEnv(strings.ToUpper(cfg.AppName), sources.WithSource("env")),
        sources.FromDefaults(sources.WithSource("defaults")),
    }, nil
}
```

This keeps the precedence identical to today’s documented pattern:

```285:289:glazed/pkg/doc/topics/21-cmds-middlewares.md
combined := middlewares.Chain(
    sources.FromDefaults(),
    sources.FromEnv("APP"),
    middlewares.GatherFlagsFromViper(),
)
```

except we replace `GatherFlagsFromViper()` with `LoadParametersFromFile()` and rely on the same env mapping in `UpdateFromEnv`.

### 2.6 Logging without Viper

- Prefer `logging.SetupLoggingFromParsedLayers(parsedLayers)` during command execution (already available):
  - Apps can call this early in `Run(E)` or in a small bootstrap that parses only the logging layer via the new middleware chain (config + env + defaults) before running subcommands.
- Keep `InitLoggerFromViper` for backward compatibility for one release cycle, but update docs and examples to encourage the parsed-layers path.

### 2.7 Minimal application changes (example: Pinocchio)

Replace:

- `clay.InitViper("pinocchio", rootCmd)` with:
  - `logging.AddLoggingLayerToRootCommand(rootCmd, "pinocchio")`
  - Provide `cli.WithParserConfig(cli.CobraParserConfig{ AppName: "pinocchio", /* optional ConfigPath from a top-level --config flag */ })` to `BuildCobraCommand(…)` calls.
  - In `PersistentPreRunE`, either:
    - keep current Viper call temporarily, or
    - run a tiny parse over only the logging layer using the new middlewares and call `SetupLoggingFromParsedLayers`.

Replace direct `viper.Get…` reads (e.g., repositories list) with either:

- a small `ParameterLayer` dedicated to those settings and reading them from parsed layers, or
- a one-time config YAML read via `LoadParametersFromFile` into a temporary `ParsedLayers` and extracting the value, to avoid reintroducing Viper.

## 3) Concrete change list

### 3.1 Code changes

- glazed
  - `pkg/cmds/middlewares/update.go`
    - Fix env key derivation (prefix + layer prefix + name, hyphen→underscore, uppercase) and store under `p.Name`.
    - Add `env_key` metadata similar to Viper path.
  - `pkg/cli/cobra-parser.go`
    - Remove `GatherFlagsFromViper` from `ParseCommandSettingsLayer`.
    - Extend `CobraCommandDefaultMiddlewares` to include `LoadParametersFromFile` and `UpdateFromEnv` using a new config resolver and app env prefix from `CobraParserConfig`.
    - Add `AppName` and `ConfigPath` to `CobraParserConfig` (or equivalent wiring) to drive the resolver and env prefix.
  - `pkg/config` or `pkg/cli/appconfig`
    - Add `ResolveAppConfigPath(appName, explicit) (string, error)` that mirrors Clay’s search.
  - `pkg/cmds/logging`
    - Keep as-is; update examples to prefer `SetupLoggingFromParsedLayers`.
  - `pkg/cmds/parameters/viper.go` and `pkg/cmds/middlewares/cobra.go`
    - Deprecate usage paths (do not remove immediately). Add docstring deprecation notes and move calls out of defaults.

- clay
  - `pkg/init.go`
    - Introduce `InitGlazed(appName string, rootCmd *cobra.Command)` that:
      - Adds logging layer
      - Optionally exposes a `--config` root flag (string) to flow into `CobraParserConfig.ConfigPath`
      - Does not call Viper; leaves parsing to Glazed middlewares.
    - Keep `InitViper` for backward compatibility but mark as deprecated in docstring.
  - `pkg/sql/cobra.go`
    - Replace `middlewares.GatherFlagsFromViper(...)` with `sources.FromFile(…)` plus `sources.FromEnv(…)` restricted to the whitelisted layers via `WrapWithWhitelistedLayers`.

### 3.2 Documentation updates

- glazed
  - `pkg/doc/topics/21-cmds-middlewares.md`
    - Replace examples that use `GatherFlagsFromViper()` with `LoadParametersFromFile()` + `UpdateFromEnv()`.
    - Document env naming rules and show that layer prefixes participate in env key names.
  - `pkg/doc/tutorials/build-first-command.md`
    - Remove Viper-specific instructions (`viper.BindPFlags`, `InitLoggerFromViper`); recommend `SetupLoggingFromParsedLayers`.
  - `pkg/doc/topics/13-layers-and-parsed-layers.md`
    - Update the “Common Middlewares” and example chains to the new default.
  - `doc/cmd-middlewares-guide.md`
    - Same replacements; add a “migration from Viper” section.

- clay
  - `README.md`
    - Show `InitGlazed(...)` usage and middleware-based config/env instead of `InitViper`.

### 3.3 Backward compatibility and migration

- Keep Viper integration paths (`GatherFlagsFromViper`, `InitViper`, `InitLoggerFromViper`) for one release, but remove them from defaults and clearly mark as deprecated.
- Provide a concise migration guide:
  - Add `AppName`/`ConfigPath` to `CobraParserConfig` when building commands.
  - Remove `InitViper` calls; wire the new `InitGlazed` and rely on the default middleware chain.
  - If reading app settings directly from Viper, move them to a dedicated layer or one-off `LoadParametersFromFile` read.

## 4) Appendix: Key references

- Clay Viper init and env mapping

```14:44:clay/pkg/init.go
// See above: InitViperWithAppName
```

```46:79:clay/pkg/init.go
// See above: InitViper (BindPFlags + InitLoggerFromViper)
```

- Viper-backed parameter ingestion

```95:146:glazed/pkg/cmds/middlewares/cobra.go
// See above: GatherFlagsFromViper middleware
```

```9:29:glazed/pkg/cmds/parameters/viper.go
// See above: GatherFlagsFromViper (per-parameter)
```

- Cobra parser Viper usage

```228:269:glazed/pkg/cli/cobra-parser.go
// See above: ParseCommandSettingsLayer using GatherFlagsFromViper
```

- Logging from Viper vs parsed layers

```125:141:glazed/pkg/cmds/logging/init.go
// See above: InitLoggerFromViper
```

- Current env middleware discrepancy

```148:160:glazed/pkg/cmds/middlewares/update.go
// See above: current UpdateFromEnv behavior (to be aligned)
```


