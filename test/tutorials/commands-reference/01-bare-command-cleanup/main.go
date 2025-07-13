package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

// CleanupCommand demonstrates BareCommand interface from the documentation
type CleanupCommand struct {
	*cmds.CommandDescription
}

// CleanupSettings mirrors the command parameters
type CleanupSettings struct {
	Directory string `glazed.parameter:"directory"`
	OlderThan string `glazed.parameter:"older-than"`
	DryRun    bool   `glazed.parameter:"dry-run"`
}

func (c *CleanupCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &CleanupSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	fmt.Printf("Starting cleanup in %s...\n", s.Directory)

	olderThan, err := parseDuration(s.OlderThan)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	files, err := findOldFiles(s.Directory, olderThan)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("Directory is clean - no files older than %s found.\n", s.OlderThan)
		return nil
	}

	fmt.Printf("Found %d files to clean up:\n", len(files))
	for i, file := range files {
		fmt.Printf("  %d. %s\n", i+1, file)
		if !s.DryRun {
			if err := os.Remove(file); err != nil {
				fmt.Printf("     Failed to remove: %s\n", err)
			} else {
				fmt.Printf("     Removed\n")
			}
		}
	}

	if s.DryRun {
		fmt.Printf("Dry run completed. Use --no-dry-run to actually remove files.\n")
	} else {
		fmt.Printf("Cleanup completed successfully.\n")
	}

	return nil
}

// findOldFiles scans directory for files older than the specified duration
func findOldFiles(dir string, olderThan time.Duration) ([]string, error) {
	var files []string
	cutoff := time.Now().Add(-olderThan)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.ModTime().Before(cutoff) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// parseDuration parses duration strings like "7d", "24h", "30m"
func parseDuration(s string) (time.Duration, error) {
	// Simple parser for demo purposes
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}
	
	unit := s[len(s)-1:]
	valueStr := s[:len(s)-1]
	
	// Try standard time.ParseDuration first
	value, err := time.ParseDuration(s)
	if err == nil {
		return value, nil
	}
	
	// Handle day units that time.ParseDuration doesn't support
	if unit == "d" {
		days, err := strconv.Atoi(valueStr)
		if err != nil {
			return 0, fmt.Errorf("invalid number of days: %s", valueStr)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	
	return 0, fmt.Errorf("unsupported duration format: %s", s)
}

// NewCleanupCommand creates a new cleanup command
func NewCleanupCommand() (*CleanupCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"cleanup",
		cmds.WithShort("Clean up old files in a directory"),
		cmds.WithLong("Remove files older than a specified duration from a directory"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"directory",
				parameters.ParameterTypeString,
				parameters.WithDefault("./test-data"),
				parameters.WithHelp("Directory to clean up"),
			),
			parameters.NewParameterDefinition(
				"older-than",
				parameters.ParameterTypeString,
				parameters.WithDefault("7d"), // 1 week
				parameters.WithHelp("Remove files older than this duration (e.g., 7d, 24h, 30m)"),
			),
			parameters.NewParameterDefinition(
				"dry-run",
				parameters.ParameterTypeBool,
				parameters.WithDefault(true),
				parameters.WithHelp("Show what would be deleted without actually deleting"),
			),
		),
	)

	return &CleanupCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &CleanupCommand{}

func main() {
	// Create some test files first
	if err := setupTestData(); err != nil {
		log.Fatalf("Failed to setup test data: %v", err)
	}

	cmd, err := NewCleanupCommand()
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		log.Fatalf("Error building Cobra command: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "cleanup-demo",
		Short: "Demonstration of BareCommand from Glazed documentation",
	}
	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// setupTestData creates some test files for demonstration
func setupTestData() error {
	testDir := "./test-data"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return err
	}

	// Create some old files (simulate old files by setting mod time)
	oldFiles := []string{"old1.txt", "old2.log", "ancient.tmp"}
	for _, filename := range oldFiles {
		path := filepath.Join(testDir, filename)
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		file.WriteString("This is an old file")
		file.Close()

		// Make it appear old
		oldTime := time.Now().Add(-time.Hour * 24 * 8) // 8 days old
		if err := os.Chtimes(path, oldTime, oldTime); err != nil {
			return err
		}
	}

	// Create some new files
	newFiles := []string{"new1.txt", "recent.log"}
	for _, filename := range newFiles {
		path := filepath.Join(testDir, filename)
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		file.WriteString("This is a new file")
		file.Close()
	}

	return nil
}
