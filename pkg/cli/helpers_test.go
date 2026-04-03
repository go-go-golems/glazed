package cli

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	helpers "github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/stretchr/testify/require"
)

func TestPrintParsedFieldsRedactsSecretValues(t *testing.T) {
	section, err := schema.NewSection(
		schema.DefaultSlug,
		"Test",
		schema.WithFields(
			fields.New("api-key", fields.TypeSecret, fields.WithHelp("Secret API key")),
		),
	)
	require.NoError(t, err)

	parsedValues := values.New()
	sectionValues := parsedValues.GetOrCreate(section)
	definition, ok := section.GetDefinitions().Get("api-key")
	require.True(t, ok)
	err = sectionValues.Fields.UpdateValue(
		"api-key",
		definition,
		"super-secret-token",
		fields.WithSource("config"),
		fields.WithMetadata(map[string]interface{}{
			"map-value": "super-secret-token",
		}),
	)
	require.NoError(t, err)

	output, err := helpers.CaptureOutput(func() error {
		printParsedFields(parsedValues)
		return nil
	})
	require.NoError(t, err)
	require.NotContains(t, output, "super-secret-token")
	require.Contains(t, output, "su***en")
}
