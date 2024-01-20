package settings

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func makeAndParse(t *testing.T, defaults *TemplateFlagsDefaults, args ...string) *TemplateSettings {
	layer, err := NewTemplateParameterLayer()
	require.NoError(t, err)
	err = layer.InitializeParameterDefaultsFromStruct(defaults)
	require.NoError(t, err)

	layers_ := layers.NewParameterLayers(layers.WithLayers(layer))
	parsedLayers := layers.NewParsedLayers()
	err = middlewares.ExecuteMiddlewares(layers_, parsedLayers,
		middlewares.UpdateFromStringList("", args, parameters.WithParseStepSource("string-list")),
		middlewares.SetFromDefaults(parameters.WithParseStepSource(parameters.SourceDefaults)),
	)
	require.NoError(t, err)

	ps, ok := parsedLayers.Get(GlazedTemplateLayerSlug)
	require.True(t, ok)

	settings, err := NewTemplateSettings(ps)
	require.NoError(t, err)

	return settings
}

func TestTemplateFlags(t *testing.T) {
	settings := makeAndParse(t, NewTemplateFlagsDefaults(), "--template", "test")
	assert.Equal(t, "test", settings.Templates["_0"])

	settings = makeAndParse(t, NewTemplateFlagsDefaults())
	assert.Equal(t, 0, len(settings.Templates))

	defaults := NewTemplateFlagsDefaults()
	defaults.Template = "test2"
	settings = makeAndParse(t, defaults)
	assert.Equal(t, "test2", settings.Templates["_0"])
}

func TestTemplateFieldFlag(t *testing.T) {
	settings := makeAndParse(t,
		NewTemplateFlagsDefaults(),
		"--template-field", "test1:test1",
		"--template-field", "test2:test2")
	assert.Equal(t, "test1", settings.Templates["test1"])
	assert.Equal(t, "test2", settings.Templates["test2"])

	settings = makeAndParse(t, NewTemplateFlagsDefaults())
	assert.Equal(t, 0, len(settings.Templates))

	defaults := NewTemplateFlagsDefaults()
	defaults.TemplateField = map[string]string{
		"test3": "test3",
		"test4": "test4",
	}
	settings = makeAndParse(t, defaults)
	assert.Equal(t, "test3", settings.Templates["test3"])
	assert.Equal(t, "test4", settings.Templates["test4"])

	settings = makeAndParse(t, defaults, "--template-field", "test5:test5,test6:test6")
	assert.Equal(t, "test5", settings.Templates["test5"])
	assert.Equal(t, "test6", settings.Templates["test6"])
}
