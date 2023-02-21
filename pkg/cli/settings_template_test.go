package cli

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func makeCommand(t *testing.T, defaults *TemplateFlagsDefaults) *cobra.Command {
	cmd := &cobra.Command{}
	tpl, err := NewTemplateParameterLayer()
	require.NoError(t, err)
	tpl.Defaults = defaults

	err = tpl.AddFlags(cmd)
	require.NoError(t, err)

	return cmd
}

func makeAndParse(t *testing.T, defaults *TemplateFlagsDefaults, args ...string) *TemplateSettings {
	cmd := makeCommand(t, defaults)
	err := cmd.ParseFlags(args)
	require.NoError(t, err)

	tpl, err := NewTemplateParameterLayer()
	require.NoError(t, err)

	err = tpl.ParseFlags(cmd)
	require.NoError(t, err)

	return tpl.Settings
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
