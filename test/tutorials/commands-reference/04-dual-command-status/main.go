package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
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

// StatusCommand demonstrates Dual Command from the documentation
// It implements both BareCommand and GlazeCommand interfaces
type StatusCommand struct {
	*cmds.CommandDescription
}

// StatusSettings mirrors the command parameters
type StatusSettings struct {
	ShowDetails bool `glazed.parameter:"show-details"`
}

// Mock system data
type SystemStatus struct {
	CPUUsage    float64
	MemoryUsage string
	DiskUsage   float64
	Uptime      time.Duration
	Processes   int
}

// Implement BareCommand for classic mode (human-readable output)
func (c *StatusCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &StatusSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	status := getSystemStatus()

	// Human-readable output
	fmt.Printf("System Status:\n")
	fmt.Printf("  CPU: %.1f%%\n", status.CPUUsage)
	fmt.Printf("  Memory: %s\n", status.MemoryUsage)
	fmt.Printf("  Disk: %.1f%%\n", status.DiskUsage)
	fmt.Printf("  Uptime: %v\n", status.Uptime.Truncate(time.Minute))

	if s.ShowDetails {
		fmt.Printf("\nDetailed Information:\n")
		fmt.Printf("  Processes: %d\n", status.Processes)
		fmt.Printf("  CPU Cores: %d\n", runtime.NumCPU())
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	}

	return nil
}

// Implement GlazeCommand for structured output mode
func (c *StatusCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &StatusSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	status := getSystemStatus()

	// Structured data output
	row := types.NewRow(
		types.MRP("cpu_usage", status.CPUUsage),
		types.MRP("memory_usage", status.MemoryUsage),
		types.MRP("disk_usage", status.DiskUsage),
		types.MRP("uptime_seconds", int(status.Uptime.Seconds())),
		types.MRP("timestamp", time.Now()),
	)

	if s.ShowDetails {
		// Add detailed information when requested
		details := map[string]interface{}{
			"processes":  status.Processes,
			"cpu_cores":  runtime.NumCPU(),
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		}
		row.Set("details", details)
	}

	return gp.AddRow(ctx, row)
}

// getSystemStatus simulates getting system status
func getSystemStatus() SystemStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemStatus{
		CPUUsage:    15.5 + rand.Float64()*30, // 15-45%
		MemoryUsage: fmt.Sprintf("%.1f GB / 16.0 GB", float64(m.Alloc)/1024/1024/1024),
		DiskUsage:   45.0 + rand.Float64()*20,                      // 45-65%
		Uptime:      time.Hour*24*7 + time.Hour*3 + time.Minute*22, // 7 days, 3 hours, 22 minutes
		Processes:   150 + rand.Intn(50),                           // 150-200 processes
	}
}

// NewStatusCommand creates a new status command
func NewStatusCommand() (*StatusCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"status",
		cmds.WithShort("Show system status"),
		cmds.WithLong("Show system status with both human-readable and structured output modes"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"show-details",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Show detailed system information"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &StatusCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &StatusCommand{}
var _ cmds.GlazeCommand = &StatusCommand{}

func main() {
	cmd, err := NewStatusCommand()
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	// Use dual command builder for commands that implement multiple interfaces
	cobraCmd, err := cli.BuildCobraCommandDualMode(
		cmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatalf("Error building Cobra command: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "status-demo",
		Short: "Demonstration of Dual Command from Glazed documentation",
		Long: `This demonstrates how Dual Commands provide both:
- Human-readable text output for interactive use
- Structured data output for scripting and automation
All from the same command implementation.`,
	}
	rootCmd.AddCommand(cobraCmd)

	// Add comprehensive examples showing both modes
	cobraCmd.Example = `  # Human-readable output (default)
  status-demo status

  # Human-readable with details
  status-demo status --show-details

  # Structured JSON output
  status-demo status --structured-output --output json

  # Structured table output with details
  status-demo status --structured-output --show-details --output table

  # CSV for monitoring systems
  status-demo status --structured-output --output csv

  # Filter and format for specific use cases
  status-demo status --structured-output --output json | jq '.cpu_usage'`

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
