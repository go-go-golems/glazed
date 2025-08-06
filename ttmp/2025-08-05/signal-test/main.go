package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"

	"github.com/spf13/cobra"
)

// Test command that isolates context cancellation behavior
type SignalTestCommand struct {
	*cmds.CommandDescription
}

type SignalTestSettings struct {
	TestType        string `glazed.parameter:"test-type"`
	Duration        int    `glazed.parameter:"duration"`
	CreateNotifyCtx bool   `glazed.parameter:"create-notify-context"`
	Host            string `glazed.parameter:"host"`
	Port            int    `glazed.parameter:"port"`
}

func (c *SignalTestCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &SignalTestSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	fmt.Printf("[TEST] Starting signal test: %s\n", settings.TestType)
	fmt.Printf("[TEST] Duration: %d seconds\n", settings.Duration)
	fmt.Printf("[TEST] Create NotifyContext: %v\n", settings.CreateNotifyCtx)

	// Monitor original context cancellation
	go func() {
		<-ctx.Done()
		fmt.Printf("[TEST] Original context cancelled: %v\n", ctx.Err())
	}()

	testCtx := ctx

	// Conditionally create signal.NotifyContext to test interference
	if settings.CreateNotifyCtx {
		fmt.Printf("[TEST] Creating signal.NotifyContext...\n")
		var cancel context.CancelFunc
		var stop context.CancelFunc
		testCtx, cancel = context.WithCancel(ctx)
		defer cancel()
		testCtx, stop = signal.NotifyContext(testCtx, os.Interrupt, syscall.SIGTERM)
		defer stop()

		// Monitor the NotifyContext
		go func() {
			<-testCtx.Done()
			fmt.Printf("[TEST] NotifyContext cancelled: %v\n", testCtx.Err())
		}()
	}

	// Run the specified test
	switch settings.TestType {
	case "sleep":
		return c.testSleep(testCtx, settings)
	case "tcp-connect":
		return c.testTCPConnect(testCtx, settings)
	case "tcp-dial-context":
		return c.testTCPDialContext(testCtx, settings)
	case "raw-socket":
		return c.testRawSocket(testCtx, settings)
	default:
		return fmt.Errorf("unknown test type: %s", settings.TestType)
	}
}

func (c *SignalTestCommand) testSleep(ctx context.Context, settings *SignalTestSettings) error {
	fmt.Printf("[TEST] Starting sleep test for %d seconds...\n", settings.Duration)

	select {
	case <-time.After(time.Duration(settings.Duration) * time.Second):
		fmt.Printf("[TEST] Sleep completed normally\n")
		return nil
	case <-ctx.Done():
		fmt.Printf("[TEST] Sleep cancelled by context: %v\n", ctx.Err())
		return ctx.Err()
	}
}

func (c *SignalTestCommand) testTCPConnect(ctx context.Context, settings *SignalTestSettings) error {
	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	fmt.Printf("[TEST] Starting TCP connect test to %s...\n", addr)

	// Use dialer.DialContext to test context cancellation at network level
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		fmt.Printf("[TEST] TCP connect failed: %v\n", err)
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("[TEST] Error closing connection: %v\n", err)
		}
	}()

	fmt.Printf("[TEST] TCP connect succeeded to %s\n", addr)
	return nil
}

func (c *SignalTestCommand) testTCPDialContext(ctx context.Context, settings *SignalTestSettings) error {
	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	fmt.Printf("[TEST] Starting TCP DialContext test to %s...\n", addr)

	// Create a custom dialer to match what lib/pq uses
	dialer := &net.Dialer{
		Timeout: time.Duration(settings.Duration) * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		fmt.Printf("[TEST] TCP DialContext failed: %v\n", err)
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("[TEST] Error closing connection: %v\n", err)
		}
	}()

	fmt.Printf("[TEST] TCP DialContext succeeded to %s\n", addr)
	return nil
}

func (c *SignalTestCommand) testRawSocket(ctx context.Context, settings *SignalTestSettings) error {
	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	fmt.Printf("[TEST] Starting raw socket test to %s...\n", addr)

	// Resolve address first
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	// Create raw socket connection (similar to what happens in net stack)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("[TEST] Raw socket failed: %v\n", err)
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("[TEST] Error closing connection: %v\n", err)
		}
	}()

	fmt.Printf("[TEST] Raw socket succeeded to %s\n", addr)
	return nil
}

func NewSignalTestCommand() (*SignalTestCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	commandSettingsLayer, err := cli.NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"signal-test",
		cmds.WithShort("Test context cancellation behavior with different scenarios"),
		cmds.WithLong(`
Test different context cancellation scenarios to isolate the signal handling bug.

Test Types:
  sleep           - Simple context cancellation with time.Sleep
  tcp-connect     - Network connection using net.DialContext
  tcp-dial-context - Custom dialer context cancellation (like lib/pq)
  raw-socket      - Raw TCP socket connection

Examples:
  signal-test --test-type sleep --duration 10
  signal-test --test-type tcp-connect --host 127.0.0.1 --port 5432
  signal-test --test-type tcp-connect --host 127.0.0.1 --port 5432 --create-notify-context
		`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"test-type",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("sleep", "tcp-connect", "tcp-dial-context", "raw-socket"),
				parameters.WithDefault("sleep"),
				parameters.WithHelp("Type of cancellation test to run"),
				parameters.WithShortFlag("t"),
			),
			parameters.NewParameterDefinition(
				"duration",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(10),
				parameters.WithHelp("Test duration in seconds"),
				parameters.WithShortFlag("d"),
			),
			parameters.NewParameterDefinition(
				"create-notify-context",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Create signal.NotifyContext to test interference"),
				parameters.WithShortFlag("n"),
			),
			parameters.NewParameterDefinition(
				"host",
				parameters.ParameterTypeString,
				parameters.WithDefault("127.0.0.1"),
				parameters.WithHelp("Host for network tests"),
			),
			parameters.NewParameterDefinition(
				"port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(5432),
				parameters.WithHelp("Port for network tests"),
				parameters.WithShortFlag("p"),
			),
		),
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)

	return &SignalTestCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &SignalTestCommand{}

func main() {
	rootCmd := &cobra.Command{
		Use:   "signal-test",
		Short: "Test signal handling and context cancellation behavior",
		Long:  "Isolate and test different context cancellation scenarios",
	}

	// Create the test command
	signalTestCmd, err := NewSignalTestCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
		os.Exit(1)
	}

	// Convert to Cobra command
	cobraSignalTestCmd, err := cli.BuildCobraCommand(signalTestCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraSignalTestCmd)

	// Setup help system
	helpSystem := help.NewHelpSystem()
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
