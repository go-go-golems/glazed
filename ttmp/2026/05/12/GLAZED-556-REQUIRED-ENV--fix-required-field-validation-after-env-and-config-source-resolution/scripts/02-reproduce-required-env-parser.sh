#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"
TEST_FILE="pkg/cli/required_env_repro_test.go"
cleanup() { rm -f "$TEST_FILE"; }
trap cleanup EXIT
cat > "$TEST_FILE" <<'GOEOF'
package cli

import (
	"os"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestReproIssue556RequiredEnvBackedField(t *testing.T) {
	section, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithFields(fields.New(
			"required-name",
			fields.TypeString,
			fields.WithRequired(true),
		)),
	)
	require.NoError(t, err)

	parser, err := NewCobraParserFromSections(schema.NewSchema(schema.WithSections(section)), &CobraParserConfig{
		ShortHelpSections:          []string{schema.DefaultSlug},
		SkipCommandSettingsSection: true,
		AppName:                    "REQ_ENV_TEST",
	})
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "probe"}
	require.NoError(t, parser.AddToCobraCommand(cmd))
	t.Setenv("REQ_ENV_TEST_REQUIRED_NAME", "from-env")
	_, err = parser.Parse(cmd, nil)
	if err != nil {
		t.Fatalf("BUG REPRODUCED: required env-backed field failed before env could satisfy it: %v", err)
	}
}

func TestReproIssue556OptionalEnvBackedField(t *testing.T) {
	section, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithFields(fields.New("optional-name", fields.TypeString)),
	)
	require.NoError(t, err)

	parser, err := NewCobraParserFromSections(schema.NewSchema(schema.WithSections(section)), &CobraParserConfig{
		ShortHelpSections:          []string{schema.DefaultSlug},
		SkipCommandSettingsSection: true,
		AppName:                    "REQ_ENV_TEST",
	})
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "probe"}
	require.NoError(t, parser.AddToCobraCommand(cmd))
	require.NoError(t, os.Setenv("REQ_ENV_TEST_OPTIONAL_NAME", "from-env"))
	t.Cleanup(func() { _ = os.Unsetenv("REQ_ENV_TEST_OPTIONAL_NAME") })
	parsed, err := parser.Parse(cmd, nil)
	require.NoError(t, err)
	fv, ok := parsed.GetField(schema.DefaultSlug, "optional-name")
	require.True(t, ok)
	require.Equal(t, "from-env", fv.Value)
}
GOEOF

go test ./pkg/cli -run 'TestReproIssue556' -count=1 -v
