package a

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
)

func sameLineEnvSuppressed() string {
	return os.Getenv("SAME_LINE") //glazedclilint:ignore legacy bootstrap env read in test fixture
}

func previousLineEnvSuppressed() string {
	//glazedclilint:ignore legacy bootstrap env read in test fixture
	return os.Getenv("PREVIOUS_LINE")
}

func trailingSuppressionDoesNotSuppressNextStatement() string {
	first := os.Getenv("TRAILING")   //glazedclilint:ignore legacy bootstrap env read in test fixture
	second := os.Getenv("NEXT_LINE") // want `use Glazed config/env middleware`
	return first + second
}

func sameLineRawFlagSuppressed() {
	_ = flag.String("legacy", "", "legacy") //glazedclilint:ignore legacy standard flag adapter in test fixture
}

func previousLineCobraFlagSuppressed() *cobra.Command {
	var address string
	cmd := &cobra.Command{}
	//glazedclilint:ignore legacy Cobra bridge in test fixture
	cmd.Flags().StringVar(&address, "address", ":8080", "listen address")
	return cmd
}

func previousLineMultiLineCobraFlagSuppressed() *cobra.Command {
	var address string
	cmd := &cobra.Command{}
	//glazedclilint:ignore legacy multi-line Cobra bridge in test fixture
	cmd.Flags().StringVar(
		&address,
		"address",
		":8080",
		"listen address",
	)
	return cmd
}

func invalidSuppressionStillReports() string {
	//glazedclilint:ignore // want `glazedclilint suppression requires a reason`
	return os.Getenv("INVALID") // want `use Glazed config/env middleware`
}
