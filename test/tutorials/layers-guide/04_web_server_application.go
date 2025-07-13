package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Example 4: Web Server Application
// Complete layer implementation for a web server with database, logging, and server configuration

// Settings structs for type safety
type ServerSettings struct {
	Host         string `glazed.parameter:"host"`
	Port         int    `glazed.parameter:"port"`
	ReadTimeout  string `glazed.parameter:"read-timeout"`
	WriteTimeout string `glazed.parameter:"write-timeout"`
}

type LoggingSettings struct {
	Level  string `glazed.parameter:"log-level"`
	Format string `glazed.parameter:"log-format"`
	File   string `glazed.parameter:"log-file"`
}

type DatabaseSettings struct {
	Host     string `glazed.parameter:"db-host"`
	Port     int    `glazed.parameter:"db-port"`
	Name     string `glazed.parameter:"db-name"`
	Username string `glazed.parameter:"db-username"`
	Password string `glazed.parameter:"db-password"`
}

// Layer creation functions
func NewServerLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"server",
		"Web Server Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"host",
				parameters.ParameterTypeString,
				parameters.WithDefault("localhost"),
				parameters.WithHelp("Server host to bind to"),
				parameters.WithShortFlag("H"),
			),
			parameters.NewParameterDefinition(
				"port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(8080),
				parameters.WithHelp("Server port to listen on"),
				parameters.WithShortFlag("p"),
			),
			parameters.NewParameterDefinition(
				"read-timeout",
				parameters.ParameterTypeString,
				parameters.WithDefault("30s"),
				parameters.WithHelp("HTTP read timeout"),
			),
			parameters.NewParameterDefinition(
				"write-timeout",
				parameters.ParameterTypeString,
				parameters.WithDefault("30s"),
				parameters.WithHelp("HTTP write timeout"),
			),
		),
	)
}

func NewLoggingLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"logging",
		"Logging Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"log-level",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
				parameters.WithDefault("info"),
				parameters.WithHelp("Logging level"),
			),
			parameters.NewParameterDefinition(
				"log-format",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("text", "json"),
				parameters.WithDefault("text"),
				parameters.WithHelp("Log output format"),
			),
			parameters.NewParameterDefinition(
				"log-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Log file path (default: stderr)"),
			),
		),
	)
}

func NewDatabaseLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"database",
		"Database Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"db-host",
				parameters.ParameterTypeString,
				parameters.WithDefault("localhost"),
				parameters.WithHelp("Database host"),
			),
			parameters.NewParameterDefinition(
				"db-port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(5432),
				parameters.WithHelp("Database port"),
			),
			parameters.NewParameterDefinition(
				"db-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Database name"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"db-username",
				parameters.ParameterTypeString,
				parameters.WithHelp("Database username"),
			),
			parameters.NewParameterDefinition(
				"db-password",
				parameters.ParameterTypeSecret,
				parameters.WithHelp("Database password"),
			),
		),
	)
}

// Command creation with layer composition
func NewServerCommand() (*cmds.CommandDescription, error) {
	// Create layers
	serverLayer, err := NewServerLayer()
	if err != nil {
		return nil, err
	}

	loggingLayer, err := NewLoggingLayer()
	if err != nil {
		return nil, err
	}

	databaseLayer, err := NewDatabaseLayer()
	if err != nil {
		return nil, err
	}

	// Compose command with relevant layers
	return cmds.NewCommandDescription(
		"serve",
		cmds.WithShort("Start the web server"),
		cmds.WithLong("Start the web server with the specified configuration"),
		cmds.WithLayersList(serverLayer, databaseLayer, loggingLayer),
	), nil
}

func NewHealthCheckCommand() (*cmds.CommandDescription, error) {
	// Health check only needs server configuration, not database
	serverLayer, err := NewServerLayer()
	if err != nil {
		return nil, err
	}

	loggingLayer, err := NewLoggingLayer()
	if err != nil {
		return nil, err
	}

	return cmds.NewCommandDescription(
		"health",
		cmds.WithShort("Check server health"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"endpoint",
				parameters.ParameterTypeString,
				parameters.WithDefault("/health"),
				parameters.WithHelp("Health check endpoint"),
			),
		),
		cmds.WithLayersList(serverLayer, loggingLayer), // No database layer
	), nil
}

// Settings extraction helpers
func GetServerSettings(parsedLayers *layers.ParsedLayers) (*ServerSettings, error) {
	settings := &ServerSettings{}
	err := parsedLayers.InitializeStruct("server", settings)
	return settings, err
}

func GetLoggingSettings(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error) {
	settings := &LoggingSettings{}
	err := parsedLayers.InitializeStruct("logging", settings)
	return settings, err
}

func GetDatabaseSettings(parsedLayers *layers.ParsedLayers) (*DatabaseSettings, error) {
	settings := &DatabaseSettings{}
	err := parsedLayers.InitializeStruct("database", settings)
	return settings, err
}

func main() {
	fmt.Println("=== Testing Web Server Application ===")

	// Test server command creation
	fmt.Println("\n1. Creating Server Command:")
	serverCmd, err := NewServerCommand()
	if err != nil {
		log.Fatalf("Failed to create server command: %v", err)
	}

	fmt.Printf("Command: %s\n", serverCmd.Name)
	fmt.Printf("Short: %s\n", serverCmd.Short)
	fmt.Printf("Long: %s\n", serverCmd.Long)

	// Check layers
	allLayers := serverCmd.Layers
	fmt.Printf("Layers (%d total):\n", allLayers.Len())
	allLayers.ForEach(func(slug string, layer layers.ParameterLayer) {
		fmt.Printf("  - %s: %s\n", slug, layer.GetName())
	})

	// Test health check command creation
	fmt.Println("\n2. Creating Health Check Command:")
	healthCmd, err := NewHealthCheckCommand()
	if err != nil {
		log.Fatalf("Failed to create health command: %v", err)
	}

	fmt.Printf("Command: %s\n", healthCmd.Name)
	fmt.Printf("Short: %s\n", healthCmd.Short)

	// Check layers
	healthLayers := healthCmd.Layers
	fmt.Printf("Layers (%d total):\n", healthLayers.Len())
	healthLayers.ForEach(func(slug string, layer layers.ParameterLayer) {
		fmt.Printf("  - %s: %s\n", slug, layer.GetName())
	})

	// Test parameter parsing and extraction
	fmt.Println("\n3. Testing Parameter Parsing and Settings Extraction:")

	// Initialize with defaults
	parsedLayers, err := serverCmd.Layers.InitializeFromDefaults()
	if err != nil {
		log.Fatalf("Failed to initialize parsed layers: %v", err)
	}

	// Extract settings
	serverSettings, err := GetServerSettings(parsedLayers)
	if err != nil {
		log.Fatalf("Failed to extract server settings: %v", err)
	}

	loggingSettings, err := GetLoggingSettings(parsedLayers)
	if err != nil {
		log.Fatalf("Failed to extract logging settings: %v", err)
	}

	databaseSettings, err := GetDatabaseSettings(parsedLayers)
	if err != nil {
		log.Fatalf("Failed to extract database settings: %v", err)
	}

	// Display extracted settings
	fmt.Printf("Server Settings:\n")
	fmt.Printf("  Host: %s\n", serverSettings.Host)
	fmt.Printf("  Port: %d\n", serverSettings.Port)
	fmt.Printf("  ReadTimeout: %s\n", serverSettings.ReadTimeout)
	fmt.Printf("  WriteTimeout: %s\n", serverSettings.WriteTimeout)

	fmt.Printf("Logging Settings:\n")
	fmt.Printf("  Level: %s\n", loggingSettings.Level)
	fmt.Printf("  Format: %s\n", loggingSettings.Format)
	fmt.Printf("  File: %s\n", loggingSettings.File)

	fmt.Printf("Database Settings:\n")
	fmt.Printf("  Host: %s\n", databaseSettings.Host)
	fmt.Printf("  Port: %d\n", databaseSettings.Port)
	fmt.Printf("  Name: %s\n", databaseSettings.Name)
	fmt.Printf("  Username: %s\n", databaseSettings.Username)
	fmt.Printf("  Password: %s\n", databaseSettings.Password)
}
