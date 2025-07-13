package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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

// MonitorServersCommand demonstrates GlazeCommand interface from the documentation
type MonitorServersCommand struct {
	*cmds.CommandDescription
}

// MonitorSettings mirrors the command parameters
type MonitorSettings struct {
	Environment string `glazed.parameter:"environment"`
	Count       int    `glazed.parameter:"count"`
}

// Server represents a server in our inventory
type Server struct {
	Hostname      string
	Environment   string
	OSVersion     string
	KernelVersion string
}

// ServerHealth represents the health status of a server
type ServerHealth struct {
	CPUPercent      float64
	MemoryUsedGB    float64
	MemoryTotalGB   float64
	DiskUsedPercent float64
	Status          string
	LastSeen        time.Time
	ActiveAlerts    []string
	UptimeDays      int
}

func (c *MonitorServersCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &MonitorSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Get server data from various sources
	servers := getServersFromInventory(s.Environment, s.Count)

	for _, server := range servers {
		// Check server health
		health := checkServerHealth(server.Hostname)

		// Produce a rich data row with nested information
		row := types.NewRow(
			types.MRP("hostname", server.Hostname),
			types.MRP("environment", server.Environment),
			types.MRP("cpu_percent", health.CPUPercent),
			types.MRP("memory_used_gb", health.MemoryUsedGB),
			types.MRP("memory_total_gb", health.MemoryTotalGB),
			types.MRP("disk_used_percent", health.DiskUsedPercent),
			types.MRP("status", health.Status),
			types.MRP("last_seen", health.LastSeen),
			types.MRP("alerts", health.ActiveAlerts), // Can be an array
			types.MRP("metadata", map[string]interface{}{ // Nested objects work too
				"os_version":  server.OSVersion,
				"kernel":      server.KernelVersion,
				"uptime_days": health.UptimeDays,
			}),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// getServersFromInventory simulates fetching servers from an inventory system
func getServersFromInventory(environment string, count int) []Server {
	servers := make([]Server, count)
	for i := 0; i < count; i++ {
		servers[i] = Server{
			Hostname:      fmt.Sprintf("%s-server-%02d", environment, i+1),
			Environment:   environment,
			OSVersion:     "Ubuntu 22.04.3 LTS",
			KernelVersion: "5.15.0-91-generic",
		}
	}
	return servers
}

// checkServerHealth simulates checking server health
func checkServerHealth(hostname string) ServerHealth {
	rand.Seed(int64(len(hostname))) // Deterministic for demo

	cpuPercent := rand.Float64() * 100
	memoryUsed := 4 + rand.Float64()*12 // 4-16 GB used
	memoryTotal := 16.0
	diskPercent := 20 + rand.Float64()*60 // 20-80% used

	status := "healthy"
	var alerts []string

	if cpuPercent > 80 {
		status = "warning"
		alerts = append(alerts, "High CPU usage")
	}
	if diskPercent > 85 {
		status = "critical"
		alerts = append(alerts, "Disk space critical")
	}
	if memoryUsed/memoryTotal > 0.9 {
		status = "warning"
		alerts = append(alerts, "High memory usage")
	}

	return ServerHealth{
		CPUPercent:      cpuPercent,
		MemoryUsedGB:    memoryUsed,
		MemoryTotalGB:   memoryTotal,
		DiskUsedPercent: diskPercent,
		Status:          status,
		LastSeen:        time.Now().Add(-time.Duration(rand.Intn(60)) * time.Minute),
		ActiveAlerts:    alerts,
		UptimeDays:      rand.Intn(365),
	}
}

// NewMonitorServersCommand creates a new monitor servers command
func NewMonitorServersCommand() (*MonitorServersCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"monitor",
		cmds.WithShort("Monitor server health across environments"),
		cmds.WithLong("Monitor server health and generate structured output in multiple formats"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"environment",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("production", "staging", "development"),
				parameters.WithDefault("production"),
				parameters.WithHelp("Environment to monitor"),
			),
			parameters.NewParameterDefinition(
				"count",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(5),
				parameters.WithHelp("Number of servers to monitor"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &MonitorServersCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &MonitorServersCommand{}

func main() {
	cmd, err := NewMonitorServersCommand()
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		log.Fatalf("Error building Cobra command: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "monitor-demo",
		Short: "Demonstration of GlazeCommand from Glazed documentation",
		Long: `This demonstrates how GlazeCommand generates structured data that can be:
- Formatted as tables, JSON, YAML, CSV automatically
- Filtered and sorted without changing command code
- Piped to other tools for processing`,
	}
	rootCmd.AddCommand(cobraCmd)

	// Add comprehensive examples
	cobraCmd.Example = `  # Human-readable table
  monitor-demo monitor --output table

  # JSON for scripting
  monitor-demo monitor --output json

  # Find problem servers with jq
  monitor-demo monitor --output json | jq '.[] | select(.status != "healthy")'

  # CSV for spreadsheets
  monitor-demo monitor --output csv > servers.csv

  # Filter high CPU usage and sort
  monitor-demo monitor --filter 'cpu_percent > 80' --sort cpu_percent

  # Monitor staging environment
  monitor-demo monitor --environment staging

  # Get more servers
  monitor-demo monitor --count 10`

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
