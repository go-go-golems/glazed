package values

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/stretchr/testify/require"
)

type testSection struct {
	name        string
	slug        string
	description string
	prefix      string
	definitions *fields.Definitions
}

func (t *testSection) GetDefinitions() *fields.Definitions {
	return t.definitions
}

func (t *testSection) GetName() string {
	return t.name
}

func (t *testSection) GetDescription() string {
	return t.description
}

func (t *testSection) GetPrefix() string {
	return t.prefix
}

func (t *testSection) GetSlug() string {
	return t.slug
}

func createSection(t *testing.T, slug, name string, paramDefs ...*fields.Definition) Section {
	definitions := fields.NewDefinitions()
	for _, def := range paramDefs {
		definitions.Set(def.Name, def)
	}
	return &testSection{
		name:        name,
		slug:        slug,
		description: "",
		prefix:      "",
		definitions: definitions,
	}
}

func createSectionValues(t *testing.T, section Section, parsedValues map[string]interface{}) *SectionValues {
	sectionValues, err := NewSectionValues(section)
	require.NoError(t, err)
	if len(parsedValues) == 0 {
		return sectionValues
	}

	for key, value := range parsedValues {
		definition, ok := section.GetDefinitions().Get(key)
		require.True(t, ok, "definition %s missing", key)
		parsed := &fields.FieldValue{Definition: definition}
		err := parsed.Update(value)
		require.NoError(t, err)
		sectionValues.Fields.Set(key, parsed)
	}

	return sectionValues
}
