# Building a Full Glazed Application

This tutorial explains how to build a complete application using the glazed framework. It focuses on best practices and architectural patterns, building upon these core concepts:
- [Creating Individual Commands](create-command-tutorial.md) - Understanding basic command creation
- [Command Description Structure](command-description.md) - Deep dive into command metadata and configuration
- [Working with Rows and Middlewares](TUTORIAL.md) - Understanding data processing in glazed

## 1. Best Practices

### 1.1 Layer Design

1. **Single Responsibility**
   - Each layer should handle one aspect of configuration (e.g., authentication, database, output formatting)
   - Keep parameter definitions focused and cohesive
   - Avoid mixing unrelated parameters in the same layer

2. **Clear Naming**
   - Use descriptive slugs that indicate the layer's purpose (e.g., "auth", "db", "output")
   - Follow consistent naming patterns across layers
   - Document the purpose of each layer

3. **Parameter Organization**
   - Group related parameters together
   - Use clear, descriptive parameter names
   - Provide helpful descriptions and defaults
   - Consider validation requirements

4. **Reusability**
   - Design layers to be reusable across commands
   - Extract common parameters into shared layers
   - Use composition to combine layers

### 1.2 Middleware Organization

1. **Composability**
   - Keep middleware functions small and focused
   - Chain middleware in a logical order
   - Make dependencies explicit
   - Allow for easy insertion/removal of middleware

2. **Error Handling**
   - Provide clear error messages
   - Handle edge cases gracefully
   - Propagate errors with context
   - Log appropriate debugging information

3. **Performance**
   - Consider the order of middleware execution
   - Avoid unnecessary processing
   - Cache results when appropriate
   - Profile and optimize hot paths

4. **Configuration**
   - Use whitelisting to control layer access
   - Support multiple configuration sources
   - Handle defaults appropriately
   - Validate configuration early

### 1.3 Command Structure

1. **Organization**
   - Group related commands together
   - Use a consistent hierarchy
   - Keep command groups focused
   - Consider command dependencies

2. **Documentation**
   - Provide clear short and long descriptions
   - Document all flags and arguments
   - Include examples in help text
   - Keep help messages concise but informative

3. **Error Handling**
   - Use descriptive error messages
   - Handle user input errors gracefully
   - Provide suggestions for common mistakes
   - Log errors with appropriate context

4. **Testing**
   - Test each command independently
   - Mock external dependencies
   - Test error conditions
   - Verify help text and documentation

### 1.4 Data Processing

1. **Row Handling**
   - Use appropriate row types for your data
   - Handle missing or invalid data gracefully
   - Consider memory usage with large datasets
   - Implement proper cleanup

2. **Output Formatting**
   - Support multiple output formats
   - Handle special characters properly
   - Consider terminal capabilities
   - Format output for readability

3. **Performance**
   - Process data in chunks when possible
   - Use appropriate data structures
   - Consider memory vs. CPU tradeoffs
   - Profile and optimize as needed

## 2. Project Organization

A typical project structure following these best practices:

```
cmd/
  myapp/
    main.go                 # Application entry point
    cmds/                  # Command implementations
      root.go             # Root command setup
      command_group/      # Related commands
        root.go          # Group command setup
        command1.go      # Individual command
        command2.go      # Individual command
    pkg/
      layers/            # Custom parameter layers
        auth.go         # Authentication layer
        db.go          # Database layer
        output.go      # Output formatting layer
      middlewares/       # Custom middleware
        parse.go       # Parameter parsing
        validate.go    # Validation logic
        transform.go   # Data transformation
      domain/           # Application-specific code
      utils/           # Shared utilities
```

## 3. Common Patterns

1. **Layer Definition**
   ```go
   // Define focused, single-purpose layers
   type AuthSettings struct {
       Token string `glazed.parameter:"token"`
   }

   func NewAuthLayer() (layers.ParameterLayer, error) {
       return layers.NewParameterLayer(
           "auth",
           "Authentication settings",
           // ... focused parameter definitions
       )
   }
   ```

2. **Middleware Chain**
   ```go
   // Compose middleware in a logical order
   middlewares := []middlewares.Middleware{
       // 1. Parse input
       middlewares.ParseFromCobraCommand(cmd),
       // 2. Load configuration
       middlewares.GatherFlagsFromViper(),
       // 3. Validate
       customMiddlewares.ValidateSettings(),
       // 4. Transform
       customMiddlewares.TransformData(),
   }
   ```

3. **Error Handling**
   ```go
   // Provide helpful error messages
   if err := validateInput(input); err != nil {
       return fmt.Errorf("invalid input %q: %w. Expected format: %s",
           input, err, expectedFormat)
   }
   ```

## 4. Component Sketches

### 4.1 Main Application

```go
// main.go
package main

import (
    "os"

    clay "github.com/go-go-golems/clay/pkg"
    "github.com/go-go-golems/glazed/pkg/help"
    help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
    "myapp/cmds"
)

func main() {
    // 2. Create root command
    rootCmd, err := cmds.NewRootCmd()
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to create root command")
    }

    // 4. Initialize configuration and logger
    err = clay.InitViper("myapp", rootCmd)
    cobra.CheckErr(err)

    // 5. Setup help system
    helpSystem := help.NewHelpSystem()
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

    // 6. Execute
    if err := rootCmd.Execute(); err != nil {
        log.Fatal().Err(err).Msg("Failed to execute root command")
    }
}
```

### 4.2 Root Command

```go
// cmds/root.go
package cmds

import (
    "github.com/spf13/cobra"
    "myapp/cmds/group1"
    "myapp/cmds/group2"
)

func NewRootCmd() (*cobra.Command, error) {
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "MyApp CLI tool",
        Long:  "A CLI tool for interacting with MyApp",
    }

    // Add command groups
    group1Cmd, err := group1.NewGroup1Cmd()
    if err != nil {
        return nil, fmt.Errorf("could not create group1 command: %w", err)
    }
    rootCmd.AddCommand(group1Cmd)

    group2Cmd, err := group2.NewGroup2Cmd()
    if err != nil {
        return nil, fmt.Errorf("could not create group2 command: %w", err)
    }
    rootCmd.AddCommand(group2Cmd)

    return rootCmd, nil
}
```

### 4.3 Command Group

```go
// cmds/group1/root.go
package group1

import (
    "github.com/spf13/cobra"
)

func NewGroup1Cmd() (*cobra.Command, error) {
    cmd := &cobra.Command{
        Use:   "group1",
        Short: "Group1 commands",
        Long:  "Commands for managing group1 resources",
    }

    // Add subcommands
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

### 4.4 Individual Command

```go
// cmds/group1/list.go
package group1

import (
    "context"
    "fmt"

    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/spf13/cobra"
    "myapp/pkg/layers"
)

type ListCommand struct {
    *cmds.CommandDescription
}

type ListSettings struct {
    layers.AppSettings
}

func NewListCmd() (*cobra.Command, error) {
    glazedCmd, err := NewListGlazedCmd()
    if err != nil {
        return nil, fmt.Errorf("could not create list glazed command: %w", err)
    }

    cmd, err := cli.BuildCobraCommandFromCommand(glazedCmd,
        cli.WithCobraShortHelpLayers(layers.AppSlug),
        cli.WithCobraMiddlewaresFunc(middlewares.GetCobraCommandMiddlewares),
    )
    if err != nil {
        return nil, fmt.Errorf("could not create list cobra command: %w", err)
    }

    return cmd, nil
}

func NewListGlazedCmd() (*ListCommand, error) {
    // 1. Create glazed parameter layer
    glazedParameterLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
    }

    // 2. Create app layer
    appLayer, err := layers.NewAppParameterLayer()
    if err != nil {
        return nil, fmt.Errorf("could not create App parameter layer: %w", err)
    }

    // 3. Combine layers
    layers_ := layers.NewParameterLayers(layers.WithLayers(
        glazedParameterLayer,
        appLayer,
    ))

    // 4. Create command
    return &ListCommand{
        CommandDescription: cmds.NewCommandDescription(
            "list",
            cmds.WithShort("List resources"),
            cmds.WithLong("List all resources in the system"),
            cmds.WithLayers(layers_),
        ),
    }, nil
}

func (c *ListCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // 1. Parse settings
    s := &ListSettings{}
    if err := parsedLayers.InitializeStruct(layers.AppSlug, &s.AppSettings); err != nil {
        return err
    }

    // 2. Validate
    if err := s.Validate(); err != nil {
        return err
    }

    // 3. Process data
    items, err := c.fetchItems(ctx, s)
    if err != nil {
        return fmt.Errorf("failed to fetch items: %w", err)
    }

    // 4. Output rows
    for _, item := range items {
        row := types.NewRow(
            types.MRP("id", item.ID),
            types.MRP("name", item.Name),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return fmt.Errorf("failed to add row: %w", err)
        }
    }

    return nil
}
```

## 5. Further Reading

- [Command Description Documentation](command-description.md) - Detailed command configuration
- [Middlewares Tutorial](TUTORIAL.md) - Data processing and transformation
- [Creating Commands Tutorial](create-command-tutorial.md) - Basic command creation

This guide focuses on best practices for building glazed applications. For specific implementation details, refer to the linked documentation. 