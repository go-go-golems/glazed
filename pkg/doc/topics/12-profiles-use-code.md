---
Title: Implementing Profiles in a Glazed Command
Slug: implementing-profile-middleware
Short: |
  Guide for developers on how to leverage the profile middleware in Glazed commands.
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

# Implementing Profile Middleware in Pinocchio

This document is intended for developers who want to understand how Pinocchio handles profile-based configuration using
middleware. It explains how the profile middleware is instantiated and how the `--profile` and `--profile-file` flags (
along with their corresponding environment variables) are processed.

## Profile Middleware Overview

Profile middleware in Pinocchio is responsible for loading and applying configuration parameters from a specified
profile. This middleware allows developers to define different configurations for various environments or use cases,
which can be dynamically selected at runtime.

## Middleware Instantiation

The profile middleware is instantiated as part of the `GetCobraCommandGeppettoMiddlewares` function. This function is
responsible for setting up the middleware chain that will be applied to a Cobra command.

Here's an example of how the profile middleware is instantiated:

```go
func GetCobraCommandGeppettoMiddlewares(
	commandSettings *cli.GlazedCommandSettings,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {
	// ... other middleware ...

	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	defaultProfileFile := fmt.Sprintf("%s/YOUR_PROGRAM/profiles.yaml", xdgConfigPath)
	if commandSettings.ProfileFile == "" {
		commandSettings.ProfileFile = defaultProfileFile
	}
	if commandSettings.Profile == "" {
		commandSettings.Profile = "default"
	}

	profileMiddleware := middlewares.GatherFlagsFromProfiles(
		defaultProfileFile,
		commandSettings.ProfileFile,
		commandSettings.Profile,
		parameters.WithParseStepSource("profiles"),
		parameters.WithParseStepMetadata(map[string]interface{}{
			"profileFile": commandSettings.ProfileFile,
			"profile":     commandSettings.Profile,
		}),
	)

	// ... appending profileMiddleware to the middleware chain ...

	return middlewares_, nil
}
```

## Loading Profile Configuration

The `GatherFlagsFromProfiles` middleware loads the profile specified by the `--profile` flag or the `YOUR_PROGRAM_PROFILE`
environment variable. These are loaded from the profile file specified by the `--profile-file` flag or the default profile
path if not specified.

The middleware would usually be inserted in front of the "viper" and "defaults" middlewares, so that the profile and profile-file 
flags can be loaded from the config file or the environment.

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
)
```

