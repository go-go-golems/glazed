package settings

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseStructuredOutputSettings(t *testing.T, args ...string) *values.SectionValues {
	t.Helper()
	section, err := NewStructuredOutputSection()
	require.NoError(t, err)

	schema_ := schema.NewSchema(schema.WithSections(section))
	parsedValues := values.New()
	err = sources.Execute(
		schema_,
		parsedValues,
		sources.UpdateFromStringList("", args, fields.WithSource("test")),
		sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
	)
	require.NoError(t, err)

	sectionValues, ok := parsedValues.Get(StructuredOutputSlug)
	require.True(t, ok)
	return sectionValues
}

func TestStructuredOutputSettingsDefaults(t *testing.T) {
	settings, err := DecodeStructuredOutputSettings(parseStructuredOutputSettings(t))
	require.NoError(t, err)
	assert.Equal(t, OutputTable, settings.Format)
	assert.Empty(t, settings.OutputFields)
	assert.Zero(t, settings.MaxOutputRows)
}

func TestStructuredOutputSettingsDecodeAndNormalize(t *testing.T) {
	settings, err := DecodeStructuredOutputSettings(parseStructuredOutputSettings(
		t,
		"--format", "jsonl",
		"--output-fields", " name,id,name, ",
		"--max-output-rows", "2",
	))
	require.NoError(t, err)
	assert.Equal(t, OutputJSONL, settings.Format)
	assert.Equal(t, []string{"name", "id"}, settings.OutputFields)
	assert.Equal(t, 2, settings.MaxOutputRows)
}

func TestStructuredOutputRejectsNegativeMaxRows(t *testing.T) {
	_, err := DecodeStructuredOutputSettings(parseStructuredOutputSettings(
		t,
		"--max-output-rows", "-1",
	))
	require.EqualError(t, err, "max-output-rows must be greater than or equal to zero")
}

func TestStructuredOutputProjectsAndCapsJSONLines(t *testing.T) {
	sectionValues := parseStructuredOutputSettings(
		t,
		"--format", "jsonl",
		"--output-fields", "name,id",
		"--max-output-rows", "2",
	)
	buf := &bytes.Buffer{}
	processor, _, err := SetupStructuredOutput(sectionValues, buf)
	require.NoError(t, err)

	ctx := context.Background()
	for i, name := range []string{"Ada", "Grace", "Katherine"} {
		require.NoError(t, processor.AddRow(ctx, types.NewRow(
			types.MRP("id", i+1),
			types.MRP("name", name),
			types.MRP("internal", true),
		)))
	}
	require.NoError(t, processor.Close(ctx))
	assert.Equal(t, "{\"id\":1,\"name\":\"Ada\"}\n{\"id\":2,\"name\":\"Grace\"}\n", buf.String())
}

func TestStructuredOutputCSVPreservesProjectionOrderAndCap(t *testing.T) {
	sectionValues := parseStructuredOutputSettings(
		t,
		"--format", "csv",
		"--output-fields", "name,id",
		"--max-output-rows", "1",
	)
	buf := &bytes.Buffer{}
	processor, _, err := SetupStructuredOutput(sectionValues, buf)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, processor.AddRow(ctx, types.NewRow(
		types.MRP("id", 1),
		types.MRP("name", "Ada"),
	)))
	require.NoError(t, processor.AddRow(ctx, types.NewRow(
		types.MRP("id", 2),
		types.MRP("name", "Grace"),
	)))
	require.NoError(t, processor.Close(ctx))
	assert.Equal(t, "name,id\nAda,1\n", buf.String())
}

func TestEveryStructuredOutputFormatProducesOutput(t *testing.T) {
	for _, format := range structuredOutputFormats {
		t.Run(format, func(t *testing.T) {
			sectionValues := parseStructuredOutputSettings(t, "--format", format)
			buf := &bytes.Buffer{}
			processor, outputFormatter, err := SetupStructuredOutput(sectionValues, buf)
			require.NoError(t, err)
			require.NotNil(t, outputFormatter)

			ctx := context.Background()
			require.NoError(t, processor.AddRow(ctx, types.NewRow(
				types.MRP("id", 1),
				types.MRP("name", "Ada"),
			)))
			require.NoError(t, processor.Close(ctx))
			assert.NotEmpty(t, buf.String())
		})
	}
}
