package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runTest(testFile string) (bool, string) {
	fmt.Printf("Running %s...\n", testFile)

	cmd := exec.Command("go", "run", testFile)
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
	}

	return true, string(output)
}

func main() {
	fmt.Println("=== Logging Layer Documentation Test Suite ===")
	fmt.Println("Running all test programs to validate documentation examples...")

	tests := []struct {
		file        string
		description string
	}{
		{"01-basic-integration.go", "Basic integration pattern"},
		{"02-structured-logging.go", "Structured logging with fields"},
		{"03-contextual-loggers.go", "Contextual logger creation"},
		{"04-formats-and-levels.go", "Different formats and log levels"},
		{"05-file-logging.go", "File output and rotation"},
		{"06-performance-optimization.go", "Performance optimization patterns"},
		{"07-missing-functions-test.go", "Documented but missing functions"},
		{"08-cobra-cli-integration.go", "Cobra CLI flag integration"},
		{"missing-functions.go", "Implementation of missing functions"},
	}

	results := make(map[string]bool)
	var failed []string

	for _, test := range tests {
		fmt.Printf("\n--- %s ---\n", test.description)

		success, output := runTest(test.file)
		results[test.file] = success

		if success {
			fmt.Printf("✅ PASSED\n")
			// Show first few lines of output
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) > 3 {
				for i := 0; i < 3; i++ {
					fmt.Printf("   %s\n", lines[i])
				}
				fmt.Printf("   ... (%d more lines)\n", len(lines)-3)
			} else {
				for _, line := range lines {
					fmt.Printf("   %s\n", line)
				}
			}
		} else {
			fmt.Printf("❌ FAILED\n")
			failed = append(failed, test.file)
			fmt.Printf("   %s\n", output)
		}
	}

	// Summary
	fmt.Printf("\n=== TEST SUMMARY ===\n")
	passed := 0
	for _, success := range results {
		if success {
			passed++
		}
	}

	fmt.Printf("Tests passed: %d/%d\n", passed, len(tests))

	if len(failed) > 0 {
		fmt.Printf("Failed tests:\n")
		for _, test := range failed {
			fmt.Printf("  - %s\n", test)
		}
	} else {
		fmt.Printf("All tests passed! ✅\n")
	}

	// Check if report exists
	reportPath := "report.md"
	if _, err := os.Stat(reportPath); err == nil {
		fmt.Printf("\nDetailed analysis available in: %s\n", reportPath)
	}

	// List all created files
	fmt.Printf("\nTest files created:\n")
	files, err := filepath.Glob("*.go")
	if err == nil {
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
		}
	}

	if _, err := os.Stat("report.md"); err == nil {
		fmt.Printf("  - report.md\n")
	}

	fmt.Printf("\nAll test programs demonstrate the concepts from the logging layer documentation.\n")
	fmt.Printf("See report.md for detailed findings and recommendations.\n")
}
