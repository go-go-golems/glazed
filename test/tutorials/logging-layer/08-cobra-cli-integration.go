package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "test-app",
		Short: "Test application with logging",
		Run: func(cmd *cobra.Command, args []string) {
			// Get values from flags
			logLevel, _ := cmd.Flags().GetString("log-level")
			logFormat, _ := cmd.Flags().GetString("log-format")
			logFile, _ := cmd.Flags().GetString("log-file")
			withCaller, _ := cmd.Flags().GetBool("with-caller")
			logToStdout, _ := cmd.Flags().GetBool("log-to-stdout")

			// Setup logging manually (since the documented function doesn't exist)
			settings := &logging.LoggingSettings{
				LogLevel:    logLevel,
				LogFormat:   logFormat,
				LogFile:     logFile,
				WithCaller:  withCaller,
				LogToStdout: logToStdout,
			}

			if err := logging.InitLoggerFromSettings(settings); err != nil {
				fmt.Printf("Failed to setup logging: %v\n", err)
				os.Exit(1)
			}

			log.Info().Msg("Application started")
			log.Debug().Str("app", "test-app").Msg("Debug information")
			log.Info().
				Str("format", logFormat).
				Str("level", logLevel).
				Msg("Logging configuration applied")
			log.Info().Msg("Application completed")
		},
	}

	// Add logging flags using the documented approach
	err := logging.AddLoggingLayerToRootCommand(rootCmd, "test-app")
	if err != nil {
		fmt.Printf("Failed to add logging layer: %v\n", err)
		os.Exit(1)
	}

	// Execute with different flag combinations
	fmt.Println("=== Testing Cobra CLI integration ===")

	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "Default settings",
			args: []string{},
		},
		{
			name: "Debug level with text format",
			args: []string{"--log-level", "debug", "--log-format", "text"},
		},
		{
			name: "JSON format with caller info",
			args: []string{"--log-format", "json", "--with-caller"},
		},
		{
			name: "File logging",
			args: []string{"--log-file", "/tmp/test-cli.log", "--log-format", "json"},
		},
		{
			name: "Dual output (file + stdout)",
			args: []string{"--log-file", "/tmp/test-cli.log", "--log-to-stdout", "--log-format", "text"},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.name)

		// Reset command args
		rootCmd.SetArgs(tc.args)

		// Execute command
		if err := rootCmd.Execute(); err != nil {
			fmt.Printf("Command failed: %v\n", err)
		}
	}

	// Clean up
	os.Remove("/tmp/test-cli.log")
}
