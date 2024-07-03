This guide explains how to set up a new program with a main.go structure exposing glazed commands.

## 1. Main Package Setup

In your main.go file, set up the following structure:

```go
package main

import (
    "embed"
    clay "github.com/go-go-golems/clay/pkg"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/help"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "os"
)

//go:embed docs/*
var docsFS embed.FS

var rootCmd = &cobra.Command{
    Use:   "your-app",
    Short: "Your application description",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        err := clay.InitLogger()
        cobra.CheckErr(err)
    },
}

func main() {
    // Implementation will be added later
}
```

This setup imports necessary packages and embeds documentation files. The `rootCmd` is defined with a `PersistentPreRun` function to initialize logging for all commands.

## 2. Root Command Initialization

Create an `initRootCmd` function to set up the root command:

```go
func initRootCmd() (*help.HelpSystem, error) {
    helpSystem := help.NewHelpSystem()
    
    // Load embedded documentation
    err := helpSystem.LoadSectionsFromFS(docsFS, ".")
    if err != nil {
        return nil, err
    }

    helpSystem.SetupCobraRootCommand(rootCmd)

    err = clay.InitViper("your-app", rootCmd)
    if err != nil {
        return nil, err
    }

    err = clay.InitLogger()
    if err != nil {
        return nil, err
    }

    return helpSystem, nil
}
```

This function initializes several crucial components:
- Help System: Loads documentation and sets up the help command.
- Viper: Initializes configuration management.
- Logger: Sets up application logging.

These components are essential for a full-featured Glazed application.

## 3. Command Registration

Create a function to register your Glazed commands:

```go
func registerCommands(helpSystem *help.HelpSystem) error {
    // Example of registering a custom Glazed command
    customCommand, err := cmds.NewCustomCommand()
    if err != nil {
        return err
    }
    cobraCustomCommand, err := glazed.BuildCobraCommand(customCommand)
    if err != nil {
        return err
    }
    rootCmd.AddCommand(cobraCustomCommand)

    // Register other commands here

    return nil
}
```

This function maps your Glazed commands to Cobra commands and adds them to the root command. This is where you'll register all your application's commands.

## 4. Main Function Structure

Implement the `main` function:

```go
func main() {
    helpSystem, err := initRootCmd()
    cobra.CheckErr(err)

    err = registerCommands(helpSystem)
    cobra.CheckErr(err)

    err = rootCmd.Execute()
    cobra.CheckErr(err)
}
```

This function initializes the root command, registers all commands, and executes the root command.

## 5. Optional Components

### Custom Command Loader (for YAML-defined commands)

If you're using YAML to define commands, you can create a custom command loader:

```go
type YourCustomCommandLoader struct{}

func (l *YourCustomCommandLoader) LoadCommands(
    fs_ fs.FS,
    filePath string,
    options []cmds.CommandDescriptionOption,
    aliasOptions []alias.Option,
) ([]cmds.Command, error) {
    // Implement your custom command loading logic here
}
```

### Run Command Functionality

To allow executing commands from files:

```go
var runCommandCmd = &cobra.Command{
    Use:   "run-command",
    Short: "Run a command from a file",
    Args:  cobra.ExactArgs(1),
    Run:   runCommandFunc,
}

func runCommandFunc(cmd *cobra.Command, args []string) {
    // Implement logic to load and execute command from file
}
```

Add this to your `initRootCmd` function:

```go
rootCmd.AddCommand(runCommandCmd)
```

### Custom Middleware

You can define custom middleware to modify command behavior:

```go
func YourMiddlewareFunc(cmd *cobra.Command, glazedCommand cmds.GlazeCommand) error {
    // Implement your middleware logic here
}
```

### Custom Help Layers

To provide specialized help information:

```go
const YourCustomHelpSlug = "your-custom-help"

// Implement your custom help layer
```


