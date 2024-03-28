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

