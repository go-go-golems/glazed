package main

import (
	"fmt"
	"os"

	"glazed-logging-layer/logging"
)

func main() {
	fmt.Println("Testing LoggingSettings validation...")

	// Test valid settings
	settings := &logging.LoggingSettings{
		Level:      "info",
		Format:     "text",
		File:       "",
		WithCaller: false,
		Verbose:    false,
	}

	if err := settings.Validate(); err != nil {
		fmt.Printf("ERROR: Valid settings failed validation: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Valid settings passed validation")

	// Test invalid log level
	settings.Level = "invalid"
	if err := settings.Validate(); err == nil {
		fmt.Println("ERROR: Invalid log level should have failed validation")
		os.Exit(1)
	}
	fmt.Println("✓ Invalid log level correctly rejected")

	// Test invalid log format
	settings.Level = "info"
	settings.Format = "xml"
	if err := settings.Validate(); err == nil {
		fmt.Println("ERROR: Invalid log format should have failed validation")
		os.Exit(1)
	}
	fmt.Println("✓ Invalid log format correctly rejected")

	// Test invalid logstash port
	settings.Format = "json"
	settings.LogstashHost = "localhost"
	settings.LogstashPort = 99999
	if err := settings.Validate(); err == nil {
		fmt.Println("ERROR: Invalid logstash port should have failed validation")
		os.Exit(1)
	}
	fmt.Println("✓ Invalid logstash port correctly rejected")

	// Test GetLogLevel
	settings.LogstashHost = ""
	settings.Level = "debug"
	if settings.GetLogLevel().String() != "debug" {
		fmt.Printf("ERROR: Expected debug level, got %s\n", settings.GetLogLevel().String())
		os.Exit(1)
	}
	fmt.Println("✓ GetLogLevel works correctly")

	// Test verbose override
	settings.Level = "error"
	settings.Verbose = true
	if settings.GetLogLevel().String() != "debug" {
		fmt.Printf("ERROR: Expected debug level with verbose=true, got %s\n", settings.GetLogLevel().String())
		os.Exit(1)
	}
	fmt.Println("✓ Verbose override works correctly")

	fmt.Println("\nAll settings tests passed! ✅")
}
