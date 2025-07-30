# Analysis and Refactoring Design for Cobra Command Creation API

This document outlines an analysis of the `glazed/pkg/cli/cobra.go` file and proposes a refactoring to simplify and unify the API for creating `cobra` commands.

## 1. Introduction

The primary goal of this refactoring is to consolidate the command creation logic within the `glazed` framework. Specifically, the aim is to eliminate the need for the separate `BuildCobraCommandDualMode` function by integrating its functionality into the main command builder, controlled by a unified set of options. This will create a more cohesive and user-friendly API for developers using the framework.

## 2. Current API Analysis

The current implementation for creating `cobra` commands is powerful but has several areas that can be improved for clarity and consistency.

### Core Components

-   **`cmds.Command` Interfaces**: The framework defines several command interfaces (`BareCommand`, `WriterCommand`, `GlazeCommand`). A single command struct can implement multiple interfaces, allowing it to be executed in different ways.
-   **`BuildCobraCommandFromCommand`**: This is the primary function for creating a `cobra.Command` from a `cmds.Command`. It currently uses a `switch` statement on the command's type, which is restrictive because it only recognizes the first interface it matches, ignoring others the command might implement.
-   **`BuildCobraCommandDualMode`**: A specialized function designed to create commands that can operate in two modes: a "classic" mode (either `WriterCommand` or `BareCommand`) and a "glaze" mode (`GlazeCommand`). The mode is typically selected at runtime via a command-line flag.
-   **`BuildCobraCommandFromGlazeCommand`, `...FromWriterCommand`, `...FromBareCommand`**: These are lower-level builders for specific command types.

### Configuration Mechanisms

There are two distinct and inconsistent mechanisms for configuring command creation:

1.  **`CobraParserOption`**: A function type (`func(*CobraParser)`) used to configure the `CobraParser`, which handles flag and argument parsing. These options are passed to most builder functions.
2.  **`DualModeOption`**: An interface type used exclusively to configure the behavior of `BuildCobraCommandDualMode`.

### Problems with the Current Design

-   **API Fragmentation**: The existence of two top-level builder functions (`BuildCobraCommandFromCommand` and `BuildCobraCommandDualMode`) forces developers to decide upfront which one to use, complicating the API.
-   **Inconsistent Configuration**: The use of two different option patterns (`CobraParserOption` and `DualModeOption`) makes the API less intuitive and harder to learn.
-   **Rigid Type Switching**: The `switch` statement in `BuildCobraCommandFromCommand` is a significant limitation. It prevents a command that implements multiple interfaces (e.g., both `GlazeCommand` and `WriterCommand`) from being used to its full potential, as only the first matched interface is considered.
-   **Isolated Logic**: The logic within `BuildCobraCommandDualMode` for handling dual-mode commands is robust (it checks interfaces at runtime inside the generated `Run` function), but it is not shared or accessible by the main builder, leading to code duplication and divergence.

## 3. Proposed Refactoring Design

To address these issues, I propose a refactoring that unifies the option system and consolidates the builder logic into a single, more flexible function.

### Unify Options

The first step is to replace the two separate option systems with a single, unified one.

-   **Introduce `CobraOption`**: A new, unified option type will be created:
    ```go
    type CobraOption func(*cobraBuilderConfig)
    ```
-   **Create `cobraBuilderConfig`**: This struct will hold all possible configurations for building a command, including parser settings and dual-mode behavior.
    ```go
    type cobraBuilderConfig struct {
        // Parser-related settings can be migrated here
        // For example, by passing this config to the parser builder

        // Builder-related settings for dual mode
        enableDualMode   bool
        glazeToggleFlag  string
        defaultToGlaze   bool
        hiddenGlazeFlags []string
    }
    ```
-   **Refactor Existing Options**: All existing `CobraParserOption` and `DualModeOption` factories will be refactored to produce the new `CobraOption` type. This provides a single, consistent way to configure any aspect of command creation. For example:
    ```go
    func WithGlazeToggleFlag(name string) CobraOption {
        return func(cfg *cobraBuilderConfig) {
            cfg.glazeToggleFlag = name
        }
    }

    func WithDualMode(enabled bool) CobraOption {
        return func(cfg *cobraBuilderConfig) {
            cfg.enableDualMode = enabled
        }
    }
    ```

### Consolidate Builder Logic

With a unified option system, we can consolidate the builder logic.

-   **Deprecate and Remove `BuildCobraCommandDualMode`**: This function will no longer be needed, as its functionality will be absorbed into the main builder.
-   **Enhance `BuildCobraCommandFromCommand`**: This function will be overhauled to become the single entry point for creating all `cobra` commands.
    -   Its signature will change to: `func BuildCobraCommandFromCommand(c cmds.Command, options ...CobraOption) (*cobra.Command, error)`.
    -   It will instantiate a `cobraBuilderConfig`, apply all the provided `CobraOption`s, and then construct the `cobra.Command`.
    -   The decision of which `Run` logic to generate will be based on the configuration.

### New `Run` Function Logic

The core of the refactoring lies in how the `Run` function for the `cobra.Command` is constructed.

-   **If Dual Mode is Enabled** (`cfg.enableDualMode == true`):
    -   The logic will be ported directly from the current `BuildCobraCommandDualMode`.
    -   A toggle flag will be added to the command.
    -   Inside the `Run` function, this flag will be checked to determine whether to execute the `GlazeCommand` interface or the `WriterCommand`/`BareCommand` interface.
-   **If Dual Mode is Disabled** (default):
    -   The restrictive `switch` statement will be replaced with a prioritized interface check.
    -   The logic will check for interfaces in a specific order (e.g., `GlazeCommand`, then `WriterCommand`, then `BareCommand`) and use the `Run` implementation from the first one that matches. This ensures that a multi-interface command is utilized correctly based on a sensible default behavior.

### Pseudocode for the New Builder

```go
func BuildCobraCommandFromCommand(c cmds.Command, options ...CobraOption) (*cobra.Command, error) {
    // 1. Create default config
    cfg := &cobraBuilderConfig{
        glazeToggleFlag: "with-glaze-output",
    }

    // 2. Apply all options
    for _, opt := range options {
        opt(cfg)
    }

    // 3. Create cobra.Command and CobraParser
    // ...

    // 4. Construct the Run function based on config
    cmd.Run = func(cmd *cobra.Command, args []string) {
        // ... (handle parsing, debug flags, etc.)

        useGlazeMode := false
        if cfg.enableDualMode {
            // Determine mode from toggle flag
            if cfg.defaultToGlaze {
                noGlaze, _ := cmd.Flags().GetBool("no-glaze-output")
                useGlazeMode = !noGlaze
            } else {
                useGlazeMode, _ = cmd.Flags().GetBool(cfg.glazeToggleFlag)
            }
        } else {
            // Default behavior: if it's a GlazeCommand, use it.
            if _, ok := c.(cmds.GlazeCommand); ok {
                useGlazeMode = true
            }
        }

        if useGlazeMode {
            // Run as GlazeCommand
        } else {
            // Run as WriterCommand or BareCommand
        }
    }

    return cmd, nil
}
```

This refactoring will result in a cleaner, more maintainable, and more intuitive API for creating `cobra` commands within the `glazed` ecosystem. 