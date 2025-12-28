---
Title: Implementing Profiles in a Glazed (Cobra) Command
Slug: implementing-profile-middleware
Short: |
  How to wire profiles.yaml into a Cobra CLI using Glazed middlewares (including proper profile-selection resolution).
Topics:
- middleware
- profiles
- configuration
Commands:
Flags:
- profile
- profile-file
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Implementing Profile Middleware in a Glazed (Cobra) command

For the conceptual overview (file format, precedence, config keys), see:

- `topics/15-profiles.md` (“Profiles (profiles.yaml)”)

This document focuses on the implementation pattern: where to insert the middleware and how to avoid the common
“profile selection circularity” when profile selection is allowed to come from env/config.

## Profile Middleware Overview

Profile middleware in Pinocchio is responsible for loading and applying configuration parameters from a specified
profile. This middleware allows developers to define different configurations for various environments or use cases,
which can be dynamically selected at runtime.

## Middleware instantiation (and why the naive approach breaks)

The profile middleware takes **constructor arguments** (`defaultProfileFile`, `profileFile`, `profileName`), so if you
compute those values before env/config are applied, you will capture defaults and load the wrong profile.

The correct approach is usually a **bootstrap parse** of `profile-settings` (and often `command-settings`) first, then
instantiate the profile middleware with the resolved values.

### Skeleton pattern (bootstrap parse + main chain)

```go
func GetCommandMiddlewares(
	parsedCommandLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {
	// 1) Bootstrap parse profile-settings (+ command-settings if config selection depends on it)
	//    using defaults + config + env + cobra, then read the resolved selection.
	//
	// 2) Build main chain:
	//    defaults -> profiles -> config -> env -> args -> flags
	//
	// (See `topics/15-profiles.md` for details.)
	return []middlewares.Middleware{/* ... */}, nil
}
```

## Loading profile configuration (recommended precedence)

The recommended precedence is:

**flags > env > config > profiles > defaults**

Operationally, this usually means inserting `GatherFlagsFromProfiles(...)` *above* defaults but *below* config/env/flags.

## Profile-Specific Overrides

For advanced use cases, combine profile middleware with additional config files using `LoadParametersFromFile` or `LoadParametersFromFiles`:

```go
func GetCobraCommandGeppettoMiddlewares(
    commandSettings *cli.GlazedCommandSettings,
    cmd *cobra.Command,
    args []string,
) ([]middlewares.Middleware, error) {
    // ... existing profile setup ...

    middlewares_ := []middlewares.Middleware{
        // Command line arguments (highest priority)
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.GatherArguments(args),
        
        // Profile-specific override files
        middlewares.LoadParametersFromFile(
            fmt.Sprintf("/etc/%s/overrides.yaml", commandSettings.Profile),
            parameters.WithParseStepSource("profile-overrides"),
        ),
        
        // Profile configuration
        profileMiddleware,
        
        // ... rest of middleware chain ...
    }

    return middlewares_, nil
}
```

This pattern allows for:
- **Profile-specific overrides**: Load additional config files based on the active profile
- **Environment-specific settings**: Different config sources for development, staging, production
- **Dynamic configuration**: Runtime determination of config sources based on profiles

## Example Usage

When running the `YOUR_PROGRAM` command, a developer can specify a profile as follows:

```bash
YOUR_PROGRAM --profile development [command]
```

Or by setting an environment variable:

```bash
export YOUR_PROGRAM_PROFILE=development
YOUR_PROGRAM [command]
```

The middleware will then load the configuration parameters from the `development` profile and apply them to the command.

## Advanced Profile Scenarios

### Custom Profile Sources

You can use the custom profile middleware to load profiles from different sources:

```go
// Load from a specific profile file
middlewares.GatherFlagsFromCustomProfiles(
    "production",
    middlewares.WithProfileFile("/etc/myapp/custom-profiles.yaml"),
    middlewares.WithProfileParseOptions(parameters.WithParseStepSource("custom-profiles")),
)

// Load from another app's profiles
middlewares.GatherFlagsFromCustomProfiles(
    "shared-config",
    middlewares.WithProfileAppName("central-config"),
    middlewares.WithProfileParseOptions(parameters.WithParseStepSource("shared-profiles")),
)
```

### Profile-Specific Configuration Files

Load different configuration files based on the active profile:

```go
// Load profile-specific config file
middlewares.LoadParametersFromFile(
    fmt.Sprintf("/etc/myapp/%s.yaml", commandSettings.Profile),
    parameters.WithParseStepSource("profile-config"),
)
```

With profile files like:
- `/etc/myapp/development.yaml`
- `/etc/myapp/staging.yaml`
- `/etc/myapp/production.yaml`

<!-- Removed: Cross-application loading via CustomViper. Prefer explicit file lists or a ConfigFiles resolver in CobraParserConfig. -->

### Environment Variable Integration

Combine environment variables with custom config loading:

```bash
# Set profile and custom config path via environment
export YOUR_PROGRAM_PROFILE=production
export YOUR_PROGRAM_CUSTOM_CONFIG=/opt/configs/production-override.yaml

# The middleware can then use these environment variables
YOUR_PROGRAM [command]
```

```go
customConfigPath := os.Getenv("YOUR_PROGRAM_CUSTOM_CONFIG")
if customConfigPath != "" {
    middlewares_ = append(middlewares_,
        middlewares.LoadParametersFromFile(
            customConfigPath,
            parameters.WithParseStepSource("env-custom"),
        ),
    )
}
```

### Configuration Hierarchy Example

A complete example showing how profiles and custom config work together:

```go
func GetAdvancedMiddlewares(commandSettings *cli.GlazedCommandSettings) []middlewares.Middleware {
    return []middlewares.Middleware{
        // 1. Command line (highest priority)
        middlewares.ParseFromCobraCommand(cmd),
        
        // 2. Environment-specific overrides
        middlewares.LoadParametersFromFile(
            "/etc/myapp/local-overrides.yaml",
            parameters.WithParseStepSource("config"),
        ),
        
        // 3. Profile-specific configuration
        middlewares.LoadParametersFromFile(
            fmt.Sprintf("/etc/myapp/profiles/%s.yaml", commandSettings.Profile),
            parameters.WithParseStepSource("config"),
        ),
        
        // 4. Custom profile sources
        middlewares.GatherFlagsFromCustomProfiles(
            commandSettings.Profile,
            middlewares.WithProfileFile("/etc/shared/organization-profiles.yaml"),
            middlewares.WithProfileParseOptions(parameters.WithParseStepSource("org-profiles")),
        ),
        
        // 5. Profile middleware (from profiles.yaml)
        middlewares.GatherFlagsFromProfiles(
            defaultProfileFile,
            commandSettings.ProfileFile,
            commandSettings.Profile,
            commandSettings.DefaultProfileName,
        ),
        
        // 6. Shared organization config
        // Provide explicit file paths or use a ConfigFiles resolver
        
        // 7. Application defaults
        middlewares.SetFromDefaults(),
    }
}
```

This creates a configuration hierarchy where:
1. Command line flags override everything
2. Local environment overrides take precedence
3. Profile-specific configs override general profile settings
4. Custom organizational profiles provide company-wide profile settings
5. Standard profile settings override shared config
6. Shared organizational config provides common defaults
7. Application defaults provide the base configuration

### Real-World Profile Scenarios

#### Multi-Environment Organization Setup

```go
// Organization-wide profiles for all applications
middlewares.GatherFlagsFromCustomProfiles(
    commandSettings.Profile,
    middlewares.WithProfileAppName("org-config"),
    middlewares.WithProfileParseOptions(parameters.WithParseStepSource("org-profiles")),
)

// Team-specific profile overrides
middlewares.GatherFlagsFromCustomProfiles(
    fmt.Sprintf("%s-%s", commandSettings.Team, commandSettings.Profile),
    middlewares.WithProfileFile("/etc/team-configs/profiles.yaml"),
    middlewares.WithProfileParseOptions(parameters.WithParseStepSource("team-profiles")),
)
```

#### Dynamic Profile Loading

```go
// Load different profile files based on deployment region
profileFile := fmt.Sprintf("/etc/regional-configs/%s/profiles.yaml", commandSettings.Region)
middlewares.GatherFlagsFromCustomProfiles(
    commandSettings.Profile,
    middlewares.WithProfileFile(profileFile),
    middlewares.WithProfileRequired(true),  // Must exist for regional deployments
)
```

#### Profile Inheritance Chain

```go
// Base organizational profiles
middlewares.GatherFlagsFromCustomProfiles(
    "base",
    middlewares.WithProfileAppName("org-base-config"),
)

// Environment-specific profiles (inherits from base)
middlewares.GatherFlagsFromCustomProfiles(
    commandSettings.Profile,
    middlewares.WithProfileAppName("org-env-config"),
)

// Application-specific profiles (highest precedence)
middlewares.GatherFlagsFromProfiles(
    defaultProfileFile,
    commandSettings.ProfileFile,
    commandSettings.Profile,
    commandSettings.DefaultProfileName,
)
```
