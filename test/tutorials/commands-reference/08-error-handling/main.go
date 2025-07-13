package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// ErrorHandlingCommand demonstrates error handling patterns from the documentation
type ErrorHandlingCommand struct {
	*cmds.CommandDescription
}

// ErrorHandlingSettings mirrors the command parameters
type ErrorHandlingSettings struct {
	Count         int    `glazed.parameter:"count"`
	ErrorType     string `glazed.parameter:"error-type"`
	SimulateError bool   `glazed.parameter:"simulate-error"`
}

func (c *ErrorHandlingCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ErrorHandlingSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}

	// Validate settings following the documentation pattern
	if err := c.validateSettings(s); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	fmt.Printf("Processing %d items with error handling...\n", s.Count)

	// Process with context cancellation support (from documentation)
	for i := 0; i < s.Count; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Simulate different error scenarios
			if s.SimulateError && c.shouldSimulateError(i, s.ErrorType) {
				switch s.ErrorType {
				case "processing":
					return fmt.Errorf("failed to process item %d: simulated processing error", i)
				case "network":
					return fmt.Errorf("failed to add row %d: network timeout", i)
				case "validation":
					return fmt.Errorf("validation failed for item %d: invalid data format", i)
				}
			}

			row := types.NewRow(
				types.MRP("id", i+1),
				types.MRP("name", fmt.Sprintf("item-%d", i+1)),
				types.MRP("status", "processed"),
				types.MRP("error_type_tested", s.ErrorType),
			)

			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add row %d: %w", i, err)
			}

			// Simulate some processing time
			time.Sleep(time.Millisecond * 10)
		}
	}

	return nil
}

// validateSettings implements the validation pattern from the documentation
func (c *ErrorHandlingCommand) validateSettings(s *ErrorHandlingSettings) error {
	if s.Count < 0 {
		return errors.New("count must be non-negative")
	}
	if s.Count > 1000 {
		return errors.New("count cannot exceed 1000")
	}

	validErrorTypes := []string{"none", "processing", "network", "validation"}
	isValid := false
	for _, validType := range validErrorTypes {
		if s.ErrorType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid error type '%s', must be one of: %v", s.ErrorType, validErrorTypes)
	}

	return nil
}

// shouldSimulateError determines when to simulate errors for demonstration
func (c *ErrorHandlingCommand) shouldSimulateError(index int, errorType string) bool {
	if errorType == "none" {
		return false
	}
	// Simulate error on the 3rd item to demonstrate error handling
	return index == 2
}

// BareCommand implementation to demonstrate exit control
func (c *ErrorHandlingCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &ErrorHandlingSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}

	// Demonstrate early exit without error (from documentation)
	if s.Count == 0 {
		fmt.Println("No items to process - exiting early")
		return &cmds.ExitWithoutGlazeError{}
	}

	fmt.Printf("Processing %d items in bare mode...\n", s.Count)
	for i := 0; i < s.Count; i++ {
		if s.SimulateError && c.shouldSimulateError(i, s.ErrorType) {
			return fmt.Errorf("simulated %s error on item %d", s.ErrorType, i+1)
		}
		fmt.Printf("Processed item %d\n", i+1)
	}

	return nil
}

// NewErrorHandlingCommand creates a new error handling demo command
func NewErrorHandlingCommand() (*ErrorHandlingCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"error-handling",
		cmds.WithShort("Demonstrate error handling patterns"),
		cmds.WithLong("Demonstrate graceful error handling, validation, and context cancellation patterns from the documentation"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"count",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(5),
				parameters.WithHelp("Number of items to process"),
			),
			parameters.NewParameterDefinition(
				"error-type",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("none", "processing", "network", "validation"),
				parameters.WithDefault("none"),
				parameters.WithHelp("Type of error to simulate"),
			),
			parameters.NewParameterDefinition(
				"simulate-error",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Whether to simulate errors"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &ErrorHandlingCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance for dual command
var _ cmds.GlazeCommand = &ErrorHandlingCommand{}
var _ cmds.BareCommand = &ErrorHandlingCommand{}

func main() {
	cmd, err := NewErrorHandlingCommand()
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	// Use dual command builder to demonstrate both interfaces
	cobraCmd, err := cli.BuildCobraCommandDualMode(
		cmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatalf("Error building Cobra command: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "error-handling-demo",
		Short: "Demonstration of error handling patterns from Glazed documentation",
		Long: `This demonstrates error handling patterns from the documentation:

1. Settings validation with clear error messages
2. Graceful error handling with context wrapping
3. Context cancellation support
4. Exit control with ExitWithoutGlazeError
5. Different error scenarios for testing

The command implements both BareCommand and GlazeCommand to show
error handling in both modes.`,
	}
	rootCmd.AddCommand(cobraCmd)

	// Add comprehensive examples
	cobraCmd.Example = `  # Normal execution
  error-handling-demo error-handling --count 3

  # Test validation errors
  error-handling-demo error-handling --count -1  # Should fail
  error-handling-demo error-handling --count 1001  # Should fail

  # Test early exit
  error-handling-demo error-handling --count 0

  # Simulate different error types
  error-handling-demo error-handling --simulate-error --error-type processing
  error-handling-demo error-handling --simulate-error --error-type network
  error-handling-demo error-handling --simulate-error --error-type validation

  # Test in both output modes
  error-handling-demo error-handling --count 5  # Bare mode
  error-handling-demo error-handling --structured-output --count 5  # Glaze mode

  # Context cancellation (try Ctrl+C during execution)
  error-handling-demo error-handling --count 100`

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
