package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

// HealthReportCommand demonstrates WriterCommand interface from the documentation
type HealthReportCommand struct {
	*cmds.CommandDescription
}

// HealthReportSettings mirrors the command parameters
type HealthReportSettings struct {
	Hostname string `glazed.parameter:"hostname"`
	Verbose  bool   `glazed.parameter:"verbose"`
}

func (c *HealthReportCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &HealthReportSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Generate a comprehensive system health report
	fmt.Fprintf(w, "System Health Report\n")
	fmt.Fprintf(w, "Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(w, "Host: %s\n\n", s.Hostname)

	// Check various system components
	components := []string{"CPU", "Memory", "Disk", "Network"}
	for _, component := range components {
		status, details := checkComponentHealth(component)
		fmt.Fprintf(w, "%s Status: %s\n", component, status)
		if s.Verbose {
			fmt.Fprintf(w, "  Details: %s\n", details)
		}
	}

	// Add recommendations if any issues found
	if recommendations := generateRecommendations(); len(recommendations) > 0 {
		fmt.Fprintf(w, "\nRecommendations:\n")
		for i, rec := range recommendations {
			fmt.Fprintf(w, "%d. %s\n", i+1, rec)
		}
	}

	return nil
}

// checkComponentHealth simulates checking system component health
func checkComponentHealth(component string) (status string, details string) {
	switch component {
	case "CPU":
		return "OK", fmt.Sprintf("Usage: %.1f%%, Cores: %d", 45.2, runtime.NumCPU())
	case "Memory":
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return "OK", fmt.Sprintf("Allocated: %.2f MB, System: %.2f MB",
			float64(m.Alloc)/1024/1024, float64(m.Sys)/1024/1024)
	case "Disk":
		return "WARNING", "Root partition 85% full"
	case "Network":
		return "OK", "All interfaces up, latency < 10ms"
	default:
		return "UNKNOWN", "Component not recognized"
	}
}

// generateRecommendations generates health recommendations
func generateRecommendations() []string {
	return []string{
		"Consider disk cleanup - root partition is getting full",
		"Monitor CPU usage during peak hours",
		"Update system packages - 15 updates available",
	}
}

// NewHealthReportCommand creates a new health report command
func NewHealthReportCommand() (*HealthReportCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"health-report",
		cmds.WithShort("Generate a system health report"),
		cmds.WithLong("Generate a comprehensive system health report that can be output to files or stdout"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"hostname",
				parameters.ParameterTypeString,
				parameters.WithDefault("localhost"),
				parameters.WithHelp("Hostname to include in the report"),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Include detailed component information"),
			),
		),
	)

	return &HealthReportCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.WriterCommand = &HealthReportCommand{}

func main() {
	cmd, err := NewHealthReportCommand()
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		log.Fatalf("Error building Cobra command: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "health-demo",
		Short: "Demonstration of WriterCommand from Glazed documentation",
		Long: `This demonstrates how WriterCommand allows the same command to output to:
- stdout for immediate viewing
- files for archival
- network connections for monitoring systems`,
	}
	rootCmd.AddCommand(cobraCmd)

	// Add examples in help
	cobraCmd.Example = `  # Output to stdout
  health-demo health-report

  # Output to file  
  health-demo health-report > health-report.txt

  # Verbose output to file
  health-demo health-report --verbose > detailed-health.txt

  # With custom hostname
  health-demo health-report --hostname production-server-01`

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
