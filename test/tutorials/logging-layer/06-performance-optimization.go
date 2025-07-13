package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog/log"
)

func calculateComplexDebuggingInfo() map[string]interface{} {
	// Simulate expensive operation
	time.Sleep(10 * time.Millisecond)
	return map[string]interface{}{
		"timestamp":    time.Now(),
		"memory_usage": "150MB",
		"cpu_usage":    "45%",
		"connections":  42,
	}
}

func demonstrateOptimization(userID string, itemCount int, duration time.Duration) {
	// Efficient: Check if debug is enabled before expensive operations
	if log.Debug().Enabled() {
		expensiveData := calculateComplexDebuggingInfo()
		log.Debug().
			Interface("debug_data", expensiveData).
			Msg("Detailed debug information")
	}

	// Always efficient: Simple field logging
	log.Info().
		Str("user", userID).
		Int("count", itemCount).
		Dur("elapsed", duration).
		Msg("Operation completed")
}

func main() {
	fmt.Println("=== Testing performance optimization ===")

	// Test with debug level (expensive operations should run)
	fmt.Println("\n--- With debug level (expensive operations enabled) ---")
	settings := &logging.LoggingSettings{
		LogLevel:  "debug",
		LogFormat: "text",
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	start := time.Now()
	demonstrateOptimization("user123", 100, time.Since(start))
	fmt.Printf("Time with debug enabled: %v\n", time.Since(start))

	// Test with info level (expensive operations should be skipped)
	fmt.Println("\n--- With info level (expensive operations disabled) ---")
	settings = &logging.LoggingSettings{
		LogLevel:  "info",
		LogFormat: "text",
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	start = time.Now()
	demonstrateOptimization("user456", 200, time.Since(start))
	fmt.Printf("Time with debug disabled: %v\n", time.Since(start))

	// Test multiple operations to show performance difference
	fmt.Println("\n--- Performance comparison (10 operations) ---")

	// Debug enabled
	settings = &logging.LoggingSettings{
		LogLevel:  "debug",
		LogFormat: "json",
	}
	logging.InitLoggerFromSettings(settings)

	start = time.Now()
	for i := 0; i < 10; i++ {
		demonstrateOptimization(fmt.Sprintf("user%d", i), i*10, time.Microsecond*time.Duration(i))
	}
	debugTime := time.Since(start)

	// Debug disabled
	settings = &logging.LoggingSettings{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logging.InitLoggerFromSettings(settings)

	start = time.Now()
	for i := 0; i < 10; i++ {
		demonstrateOptimization(fmt.Sprintf("user%d", i), i*10, time.Microsecond*time.Duration(i))
	}
	infoTime := time.Since(start)

	fmt.Printf("\nPerformance results:\n")
	fmt.Printf("Debug enabled:  %v\n", debugTime)
	fmt.Printf("Debug disabled: %v\n", infoTime)
	fmt.Printf("Speedup: %.2fx\n", float64(debugTime)/float64(infoTime))
}
