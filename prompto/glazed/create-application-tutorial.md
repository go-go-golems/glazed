# Building a Full Glazed Application

This tutorial builds upon the [Creating a New Command](create-command-tutorial.md) guide and shows how to create a complete application with multiple commands, custom layers, and middleware. We'll use a Slack CLI application as an example.

## 1. Project Structure

A typical glazed application has the following structure:

```
cmd/
  myapp/
    main.go                 # Main application entry point
    cmds/
      root.go              # Root command definition
      users/               # User-related commands
        root.go            # Users root command
        list.go            # List users command
        get.go             # Get user command
      channels/            # Channel-related commands
        root.go            # Channels root command
        list.go           # List channels command
        get.go            # Get channel command
    pkg/
      layers/             # Custom parameter layers
        myapp.go          # Application-specific layer
      middlewares/        # Custom middleware functions
        middlewares.go    # Shared middleware functions
      client/            # Application-specific client code
        client.go        # Client implementation
      utils/            # Shared utilities
        utils.go        # Utility functions
```

## 2. Creating Custom Parameter Layers

Custom parameter layers allow you to define reusable sets of parameters that can be shared across commands.

```go
// pkg/layers/myapp.go
package layers

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

const MyAppSlug = "myapp"

type MyAppSettings struct {
    Token string `glazed.parameter:"token"`
}

func NewMyAppParameterLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        MyAppSlug,
        "MyApp authentication settings",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "token",
                parameters.ParameterTypeString,
                parameters.WithHelp("API Token"),
                parameters.WithRequired(false),
            ),
        ),
    )
}
```

## 3. Creating Custom Middleware

Middleware functions allow you to customize how parameters are parsed and processed:

```go
// pkg/middlewares/middlewares.go
package middlewares

import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/spf13/cobra"
    "myapp/pkg/layers"
)

func GetCobraCommandMiddlewares(
    commandSettings *cli.GlazedCommandSettings,
    cmd *cobra.Command,
    args []string,
) ([]middlewares.Middleware, error) {
    middlewares_ := []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd,
            parameters.WithParseStepSource("cobra"),
        ),
        middlewares.GatherArguments(args,
            parameters.WithParseStepSource("arguments"),
        ),
    }

    if commandSettings.LoadParametersFromFile != "" {
        middlewares_ = append(middlewares_,
            middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
    }

    // Allow settings to be set from a config file or environment variables
    appMiddleware := middlewares.WrapWithWhitelistedLayers(
        []string{layers.MyAppSlug},
        middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("viper")),
    )
    middlewares_ = append(middlewares_, appMiddleware,
        middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
    )

    return middlewares_, nil
}
```

## 4. Creating Commands

### 4.1 Root Command

The root command sets up the application and initializes shared components:

```go
// cmds/root.go
package cmds

import (
    "github.com/spf13/cobra"
    "myapp/cmds/users"
    "myapp/cmds/channels"
)

func NewRootCmd() (*cobra.Command, error) {
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "MyApp CLI tool",
        Long:  "A CLI tool for interacting with MyApp",
    }

    // Add subcommands
    usersCmd, err := users.NewUsersCmd()
    if err != nil {
        return nil, fmt.Errorf("could not create users command: %w", err)
    }
    rootCmd.AddCommand(usersCmd)

    channelsCmd, err := channels.NewChannelsCmd()
    if err != nil {
        return nil, fmt.Errorf("could not create channels command: %w", err)
    }
    rootCmd.AddCommand(channelsCmd)

    return rootCmd, nil
}
```

### 4.2 Subcommand Groups

Create root commands for each group of related commands:

```go
// cmds/users/root.go
package users

import (
    "github.com/spf13/cobra"
)

func NewUsersCmd() (*cobra.Command, error) {
    cmd := &cobra.Command{
        Use:   "users",
        Short: "User management commands",
        Long:  "Commands for managing users",
    }

    listCmd, err := NewListCmd()
    if err != nil {
        return nil, err
    }
    cmd.AddCommand(listCmd)

    getCmd, err := NewGetCmd()
    if err != nil {
        return nil, err
    }
    cmd.AddCommand(getCmd)

    return cmd, nil
}
```

### 4.3 Individual Commands

Create individual commands that use the shared layers and middleware:

```go
// cmds/users/list.go
package users

import (
    "context"
    "fmt"

    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/spf13/cobra"
    "myapp/pkg/client"
    appLayers "myapp/pkg/layers"
    appMiddlewares "myapp/pkg/middlewares"
)

type ListCommand struct {
    *cmds.CommandDescription
}

type ListSettings struct {
    appLayers.MyAppSettings
}

func NewListCmd() (*cobra.Command, error) {
    glazedCmd, err := NewListGlazedCmd()
    if err != nil {
        return nil, fmt.Errorf("could not create list users glazed command: %w", err)
    }

    cmd, err := cli.BuildCobraCommandFromCommand(glazedCmd,
        cli.WithCobraShortHelpLayers(appLayers.MyAppSlug),
        cli.WithCobraMiddlewaresFunc(appMiddlewares.GetCobraCommandMiddlewares),
    )
    if err != nil {
        return nil, fmt.Errorf("could not create list users cobra command: %w", err)
    }

    return cmd, nil
}

func NewListGlazedCmd() (*ListCommand, error) {
    glazedParameterLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
    }

    appLayer, err := appLayers.NewMyAppParameterLayer()
    if err != nil {
        return nil, fmt.Errorf("could not create MyApp parameter layer: %w", err)
    }

    layers_ := layers.NewParameterLayers(layers.WithLayers(
        glazedParameterLayer,
        appLayer,
    ))

    return &ListCommand{
        CommandDescription: cmds.NewCommandDescription(
            "list",
            cmds.WithShort("List users"),
            cmds.WithLong("List all users in the system"),
            cmds.WithLayers(layers_),
        ),
    }, nil
}

func (c *ListCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &ListSettings{}
    if err := parsedLayers.InitializeStruct(appLayers.MyAppSlug, &s.MyAppSettings); err != nil {
        return err
    }

    if s.Token == "" {
        return fmt.Errorf("token is required")
    }

    cl := client.New(s.Token)
    users, err := cl.ListUsers(ctx)
    if err != nil {
        return fmt.Errorf("failed to list users: %w", err)
    }

    for _, user := range users {
        row := types.NewRow(
            types.MRP("id", user.ID),
            types.MRP("name", user.Name),
            types.MRP("email", user.Email),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return fmt.Errorf("failed to add row: %w", err)
        }
    }

    return nil
}
```

## 5. Main Application

The main application ties everything together:

```go
// main.go
package main

import (
    "os"

    clay "github.com/go-go-golems/clay/pkg"
    "github.com/go-go-golems/glazed/pkg/help"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
    "myapp/cmds"
)

func main() {
    // Initialize logging
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

    // Create root command
    rootCmd, err := cmds.NewRootCmd()
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to create root command")
    }

    // Setup logging for all commands
    rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
        err := clay.InitLogger()
        cobra.CheckErr(err)
    }

    // Initialize viper for configuration
    err = clay.InitViper("myapp", rootCmd)
    cobra.CheckErr(err)
    err = clay.InitLogger()
    cobra.CheckErr(err)

    // Setup help system
    helpSystem := help.NewHelpSystem()
    helpSystem.SetupCobraRootCommand(rootCmd)

    // Execute
    if err := rootCmd.Execute(); err != nil {
        log.Fatal().Err(err).Msg("Failed to execute root command")
    }
}
```

## 6. Configuration

The application can be configured through:
- Command line flags
- Environment variables (prefixed with MYAPP_)
- Configuration file (myapp.yaml)

Example configuration file:
```yaml
token: "your-api-token"
```

Environment variables:
```bash
export MYAPP_TOKEN="your-api-token"
```

## 7. Running the Application

The application can be run with various commands:

```bash
# List users
myapp users list

# Get specific user
myapp users get --id 123

# List channels
myapp channels list

# Get specific channel
myapp channels get --id general

# Use configuration file
myapp users list --config myapp.yaml

# Override token
myapp users list --token "different-token"
```

## 8. Best Practices

1. **Layer Organization**
   - Keep related parameters in their own layer
   - Use meaningful slugs for layer identification
   - Document layer parameters thoroughly

2. **Middleware**
   - Keep middleware functions focused and composable
   - Use whitelisting to control which layers can be set from different sources
   - Handle errors gracefully

3. **Command Structure**
   - Group related commands together
   - Use consistent naming conventions
   - Provide helpful short and long descriptions

4. **Error Handling**
   - Use descriptive error messages
   - Wrap errors with context
   - Log errors appropriately

5. **Configuration**
   - Use viper for flexible configuration
   - Support multiple configuration methods
   - Document configuration options

6. **Testing**
   - Test each command independently
   - Mock external dependencies
   - Test error conditions

This tutorial demonstrates how to build a complete glazed application with multiple commands, custom layers, and middleware. The example shows how to organize code, handle configuration, and follow best practices for building maintainable CLI applications. 