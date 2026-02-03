package schema

import (
	"fmt"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a field section
func createSection(t *testing.T, slug, name string, paramDefs ...*fields.Definition) Section {
	section, err := NewSection(slug, name, WithFields(paramDefs...))
	require.NoError(t, err)
	require.NotNil(t, section)
	return section
}

func TestNewSchema(t *testing.T) {
	sections := NewSchema()
	assert.NotNil(t, sections)
	assert.Equal(t, 0, sections.Len())
}

func TestSchemaSubset(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	section2 := createSection(t, "section2", "Section 2")
	section3 := createSection(t, "section3", "Section 3")

	sections := NewSchema(WithSections(section1, section2, section3))

	subset := sections.Subset("section1", "section3")

	assert.Equal(t, 2, subset.Len())
	val, present := subset.Get("section1")
	assert.NotNil(t, val)
	assert.True(t, present)
	val, present = subset.Get("section2")
	assert.Nil(t, val)
	assert.False(t, present)
	val, present = subset.Get("section3")
	assert.NotNil(t, val)
	assert.True(t, present)
}

func TestSchemaForEach(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	section2 := createSection(t, "section2", "Section 2")

	sections := NewSchema(WithSections(section1, section2))

	count := 0
	sections.ForEach(func(key string, p Section) {
		count++
		assert.Contains(t, []string{"section1", "section2"}, key)
	})

	assert.Equal(t, 2, count)
}

func TestSchemaForEachE(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	section2 := createSection(t, "section2", "Section 2")

	sections := NewSchema(WithSections(section1, section2))

	count := 0
	err := sections.ForEachE(func(key string, p Section) error {
		count++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSchemaAppendSections(t *testing.T) {
	sections := NewSchema()
	section1 := createSection(t, "section1", "Section 1")
	section2 := createSection(t, "section2", "Section 2")

	sections.AppendSections(section1, section2)

	assert.Equal(t, 2, sections.Len())
	val, present := sections.Get("section1")
	assert.Equal(t, section1, val)
	assert.True(t, present)
	val, present = sections.Get("section2")
	assert.Equal(t, section2, val)
	assert.True(t, present)
}

func TestSchemaPrependSections(t *testing.T) {
	section0 := createSection(t, "section0", "Section 0")

	sections := NewSchema(
		WithSections(section0),
	)
	section1 := createSection(t, "section1", "Section 1")
	section2 := createSection(t, "section2", "Section 2")

	sections.PrependSections(section1, section2)

	assert.Equal(t, 3, sections.Len())
	first := sections.Oldest()
	assert.Equal(t, "section1", first.Key)
	assert.Equal(t, section1, first.Value)
	second := first.Next()
	assert.Equal(t, "section2", second.Key)
	assert.Equal(t, section2, second.Value)
	third := second.Next()
	assert.Equal(t, "section0", third.Key)
	assert.Equal(t, "Section 0", third.Value.GetName())
	assert.Nil(t, third.Next())
}

func TestSchemaMerge(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	section2 := createSection(t, "section2", "Section 2")
	sections1 := NewSchema(WithSections(section1))
	sections2 := NewSchema(WithSections(section2))

	merged := sections1.Merge(sections2)

	assert.Equal(t, 2, merged.Len())
	val, present := merged.Get("section1")
	assert.NotNil(t, val)
	assert.True(t, present)
	val, present = merged.Get("section2")
	assert.True(t, present)
	assert.NotNil(t, val)
}

func TestSchemaAsList(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	section2 := createSection(t, "section2", "Section 2")
	sections := NewSchema(WithSections(section1, section2))

	list := sections.AsList()

	assert.Equal(t, 2, len(list))
	assert.Contains(t, list, section1)
	assert.Contains(t, list, section2)
}

func TestSchemaClone(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	sections := NewSchema(WithSections(section1))

	cloned := sections.Clone()

	assert.Equal(t, sections.Len(), cloned.Len())
	v1, p1 := sections.Get("section1")
	assert.True(t, p1)
	assert.NotNil(t, v1)
	v2, p2 := cloned.Get("section1")
	assert.True(t, p2)
	assert.NotNil(t, v2)
	assert.NotSame(t, v1, v2)
	assert.Equal(t, v1.GetSlug(), v2.GetSlug())
}

func TestSchemaGetAllDefinitions(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1",
		fields.New("param1", fields.TypeString),
	)
	section2 := createSection(t, "section2", "Section 2",
		fields.New("param2", fields.TypeInteger),
	)

	sections := NewSchema(WithSections(section1, section2))

	allDefs := sections.GetAllDefinitions()

	assert.Equal(t, 2, allDefs.Len())
	val, present := allDefs.Get("param1")
	assert.True(t, present)
	assert.NotNil(t, val)
	val, present = allDefs.Get("param2")
	assert.True(t, present)
	assert.NotNil(t, val)
}

func TestSchemaWithSections(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	section2 := createSection(t, "section2", "Section 2")

	sections := NewSchema(WithSections(section1, section2))

	assert.Equal(t, 2, sections.Len())
	val, present := sections.Get("section1")
	assert.True(t, present)
	assert.Equal(t, section1, val)
	val, present = sections.Get("section2")
	assert.True(t, present)
	assert.Equal(t, section2, val)
}

func TestSchemaWithDuplicateSlugs(t *testing.T) {
	section1 := createSection(t, "duplicate", "Section 1")
	section2 := createSection(t, "duplicate", "Section 2")

	sections := NewSchema(WithSections(section1, section2))

	assert.Equal(t, 1, sections.Len())
	val, present := sections.Get("duplicate")
	assert.True(t, present)
	assert.Equal(t, "Section 2", val.GetName())
}

func TestSchemaSubsetWithMissingSections(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	sections := NewSchema(WithSections(section1))

	subset := sections.Subset("section1", "non_existent")

	assert.Equal(t, 1, subset.Len())
	_, present := subset.Get("section1")
	assert.True(t, present)
	_, present = subset.Get("non_existent")
	assert.False(t, present)
}

func TestSchemaMergeWithOverlappingSections(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1 - Original")
	section2 := createSection(t, "section2", "Section 2")
	sections1 := NewSchema(WithSections(section1, section2))

	section1Duplicate := createSection(t, "section1", "Section 1 - Duplicate")
	section3 := createSection(t, "section3", "Section 3")
	sections2 := NewSchema(WithSections(section1Duplicate, section3))

	merged := sections1.Merge(sections2)

	assert.Equal(t, 3, merged.Len())
	val, present := merged.Get("section1")
	assert.True(t, present)
	assert.Equal(t, "Section 1 - Duplicate", val.GetName())
	_, present = merged.Get("section2")
	assert.True(t, present)
	_, present = merged.Get("section3")
	assert.True(t, present)
}

func TestSchemaWithLargeNumberOfSections(t *testing.T) {
	numSections := 1000
	sections := NewSchema()

	for i := 0; i < numSections; i++ {
		section := createSection(t, fmt.Sprintf("section%d", i), fmt.Sprintf("Section %d", i))
		sections.AppendSections(section)
	}

	assert.Equal(t, numSections, sections.Len())
	_, present := sections.Get("section0")
	assert.True(t, present)
	_, present = sections.Get(fmt.Sprintf("section%d", numSections-1))
	assert.True(t, present)
}

func TestSchemaWithUnicodeSectionNames(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1 - 你好")
	section2 := createSection(t, "section2", "Section 2 - こんにちは")

	sections := NewSchema(WithSections(section1, section2))

	assert.Equal(t, 2, sections.Len())
	val, present := sections.Get("section1")
	assert.True(t, present)
	assert.Equal(t, "Section 1 - 你好", val.GetName())
	val, present = sections.Get("section2")
	assert.True(t, present)
	assert.Equal(t, "Section 2 - こんにちは", val.GetName())
}
