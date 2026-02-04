package values

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewValues(t *testing.T) {
	parsedValues := New()
	assert.NotNil(t, parsedValues)
	assert.Equal(t, 0, parsedValues.Len())
}

func TestValuesWithSectionValues(t *testing.T) {
	section := createSection(t, "test", "Test Section")
	sectionValues := createSectionValues(t, section, nil)

	parsedValues := New(WithSectionValues("test", sectionValues))

	assert.Equal(t, 1, parsedValues.Len())
	val, present := parsedValues.Get("test")
	assert.True(t, present)
	assert.Equal(t, sectionValues, val)
}

func TestValuesClone(t *testing.T) {
	section := createSection(t, "test", "Test Section")
	sectionValues := createSectionValues(t, section, nil)
	parsedValues := New(WithSectionValues("test", sectionValues))

	cloned := parsedValues.Clone()

	assert.Equal(t, parsedValues.Len(), cloned.Len())
	originalVal, _ := parsedValues.Get("test")
	clonedVal, present := cloned.Get("test")
	assert.True(t, present)
	assert.NotSame(t, originalVal, clonedVal)
	assert.Equal(t, originalVal.Section, clonedVal.Section)
}

func TestValuesGetOrCreate(t *testing.T) {
	parsedValues := New()
	section := createSection(t, "test", "Test Section")

	// Get non-existent section (should create)
	sectionValues := parsedValues.GetOrCreate(section)
	assert.NotNil(t, sectionValues)
	assert.Equal(t, section, sectionValues.Section)

	// Get existing section
	sameSection := parsedValues.GetOrCreate(section)
	assert.Equal(t, sectionValues, sameSection)
}

func TestValuesGetDataMap(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1",
		fields.New("param1", fields.TypeString),
	)
	sectionValues1 := createSectionValues(t, section1, map[string]interface{}{"param1": "value1"})

	section2 := createSection(t, "section2", "Section 2",
		fields.New("param2", fields.TypeInteger),
	)
	sectionValues2 := createSectionValues(t, section2, map[string]interface{}{"param2": 42})

	parsedValues := New(
		WithSectionValues("section1", sectionValues1),
		WithSectionValues("section2", sectionValues2),
	)

	dataMap := parsedValues.GetDataMap()
	assert.Equal(t, 2, len(dataMap))
	assert.Equal(t, "value1", dataMap["param1"])
	assert.Equal(t, 42, dataMap["param2"])
}

func TestValuesInitializeStruct(t *testing.T) {
	type TestStruct struct {
		Param1 string `glazed:"param1"`
		Param2 int    `glazed:"param2"`
	}

	section := createSection(t, "test", "Test Section",
		fields.New("param1", fields.TypeString),
		fields.New("param2", fields.TypeInteger),
	)
	sectionValues := createSectionValues(t, section, map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	})
	parsedValues := New(WithSectionValues("test", sectionValues))

	var result TestStruct
	err := parsedValues.DecodeSectionInto("test", &result)
	assert.NoError(t, err)
	assert.Equal(t, "value1", result.Param1)
	assert.Equal(t, 42, result.Param2)
}

func TestValuesAllFieldValues(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1",
		fields.New("param1", fields.TypeString),
	)
	sectionValues1 := createSectionValues(t, section1, map[string]interface{}{"param1": "value1"})

	section2 := createSection(t, "section2", "Section 2",
		fields.New("param2", fields.TypeInteger),
	)
	sectionValues2 := createSectionValues(t, section2, map[string]interface{}{"param2": 42})

	parsedValues := New(
		WithSectionValues("section1", sectionValues1),
		WithSectionValues("section2", sectionValues2),
	)

	allParams := parsedValues.AllFieldValues()
	assert.Equal(t, 2, allParams.Len())
	param1, present := allParams.Get("param1")
	assert.True(t, present)
	assert.Equal(t, "value1", param1.Value)
	param2, present := allParams.Get("param2")
	assert.True(t, present)
	assert.Equal(t, 42, param2.Value)
}

func TestValuesGetField(t *testing.T) {
	section := createSection(t, "test", "Test Section",
		fields.New("param", fields.TypeString),
	)
	sectionValues := createSectionValues(t, section, map[string]interface{}{"param": "value"})
	parsedValues := New(WithSectionValues("test", sectionValues))

	param, present := parsedValues.GetField("test", "param")
	assert.True(t, present)
	assert.Equal(t, "value", param.Value)

	_, present = parsedValues.GetField("non_existent", "param")
	assert.False(t, present)

	_, present = parsedValues.GetField("test", "non_existent")
	assert.False(t, present)
}

func TestValuesDefaultSectionValues(t *testing.T) {
	parsedValues := New()

	defaultSection := parsedValues.DefaultSectionValues()
	assert.NotNil(t, defaultSection)
	assert.Equal(t, DefaultSlug, defaultSection.Section.GetSlug())

	// Calling it again should return the same section
	sameDefaultSection := parsedValues.DefaultSectionValues()
	assert.Equal(t, defaultSection, sameDefaultSection)
}

func TestValuesForEach(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	sectionValues1 := createSectionValues(t, section1, nil)

	section2 := createSection(t, "section2", "Section 2")
	sectionValues2 := createSectionValues(t, section2, nil)

	parsedValues := New(
		WithSectionValues("section1", sectionValues1),
		WithSectionValues("section2", sectionValues2),
	)

	count := 0
	parsedValues.ForEach(func(k string, v *SectionValues) {
		count++
		assert.Contains(t, []string{"section1", "section2"}, k)
	})
	assert.Equal(t, 2, count)
}

func TestValuesForEachE(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	sectionValues1 := createSectionValues(t, section1, nil)

	section2 := createSection(t, "section2", "Section 2")
	sectionValues2 := createSectionValues(t, section2, nil)

	parsedValues := New(
		WithSectionValues("section1", sectionValues1),
		WithSectionValues("section2", sectionValues2),
	)

	count := 0
	err := parsedValues.ForEachE(func(k string, v *SectionValues) error {
		count++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// Test with error
	errorOnSecond := parsedValues.ForEachE(func(k string, v *SectionValues) error {
		if k == "section2" {
			return assert.AnError
		}
		return nil
	})
	assert.Error(t, errorOnSecond)
}

func TestSectionValuesInitializeStructWithUnexportedFields(t *testing.T) {
	section := createSection(t, "test", "Test Section",
		fields.New("exported", fields.TypeString),
	)
	sectionValues := createSectionValues(t, section, map[string]interface{}{"exported": "value"})

	type TestStruct struct {
		Exported   string `glazed:"exported"`
		unexported string
	}

	result := TestStruct{
		unexported: "unexported",
	}
	err := sectionValues.DecodeInto(&result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result.Exported)
	assert.Equal(t, "unexported", result.unexported)

}

func TestValuesInitializeStructWithNonPointer(t *testing.T) {
	parsedValues := New()
	section := createSection(t, "test", "Test Section",
		fields.New("param", fields.TypeString),
	)
	sectionValues := createSectionValues(t, section, map[string]interface{}{"param": "value"})
	parsedValues.Set("test", sectionValues)

	type TestStruct struct {
		Param string `glazed:"param"`
	}

	var result TestStruct
	err := parsedValues.DecodeSectionInto("test", result) // Note: passing result, not &result
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pointer")
}

func TestValuesGetFieldNonExistentSection(t *testing.T) {
	parsedValues := New()

	_, present := parsedValues.GetField("non_existent", "param")
	assert.False(t, present)
}

func TestValuesGetOrCreateNilSection(t *testing.T) {
	parsedValues := New()

	// Depending on how you want to handle this case, you might expect an error or a new empty section
	// Alternatively, if you want to panic:
	assert.Panics(t, func() { parsedValues.GetOrCreate(nil) })
}

func TestValuesInitializeStructUnsupportedTypes(t *testing.T) {
	section := createSection(t, "test", "Test Section",
		fields.New("supported", fields.TypeString),
		fields.New("unsupported", fields.TypeString),
	)
	parsedValues := map[string]interface{}{
		"supported":   "value",
		"unsupported": make(chan int), // channels are not supported
	}
	options := make([]SectionValuesOption, 0, len(parsedValues))
	for key, value := range parsedValues {
		options = append(options, WithFieldValue(key, value))
	}
	_, err := NewSectionValues(section, options...)
	assert.Error(t, err)
}

func TestValuesForEachEWithError(t *testing.T) {
	section1 := createSection(t, "section1", "Section 1")
	sectionValues1 := createSectionValues(t, section1, nil)
	section2 := createSection(t, "section2", "Section 2")
	sectionValues2 := createSectionValues(t, section2, nil)

	parsedValues := New(
		WithSectionValues("section1", sectionValues1),
		WithSectionValues("section2", sectionValues2),
	)

	count := 0
	err := parsedValues.ForEachE(func(k string, v *SectionValues) error {
		count++
		if k == "section2" {
			return errors.New("intentional error")
		}
		return nil
	})

	assert.Error(t, err)
	assert.Equal(t, "intentional error", err.Error())
	assert.Equal(t, 2, count) // The loop should have proceeded to the second section before stopping
}

func TestValuesInitializeStructStringTypes(t *testing.T) {
	// Define custom types
	type StringAlias string
	type StringDeclaration = string

	// Define the test struct
	type TestStruct struct {
		StringField            string              `glazed:"string_field"`
		StringAliasField       StringAlias         `glazed:"string_alias_field"`
		StringDeclarationField StringDeclaration   `glazed:"string_declaration_field"`
		StringListField        []string            `glazed:"string_list_field"`
		StringAliasListField   []StringAlias       `glazed:"string_alias_list_field"`
		StringDeclListField    []StringDeclaration `glazed:"string_decl_list_field"`
	}

	// Create a field section with all the necessary definitions
	section := createSection(t, "test", "Test Section",
		fields.New("string_field", fields.TypeString),
		fields.New("string_alias_field", fields.TypeString),
		fields.New("string_declaration_field", fields.TypeString),
		fields.New("string_list_field", fields.TypeStringList),
		fields.New("string_alias_list_field", fields.TypeStringList),
		fields.New("string_decl_list_field", fields.TypeStringList),
	)

	// Create a parsed section with test values
	sectionValues := createSectionValues(t, section, map[string]interface{}{
		"string_field":             "regular string",
		"string_alias_field":       "aliased string",
		"string_declaration_field": "declared string",
		"string_list_field":        []string{"a", "b", "c"},
		"string_alias_list_field":  []string{"x", "y", "z"},
		"string_decl_list_field":   []string{"1", "2", "3"},
	})

	// Create Values and add the parsed section
	parsedValues := New(WithSectionValues("test", sectionValues))

	// Initialize the struct
	var result TestStruct
	err := parsedValues.DecodeSectionInto("test", &result)

	// Assert no error occurred
	assert.NoError(t, err)

	// Verify each field was correctly initialized
	assert.Equal(t, "regular string", result.StringField)
	assert.Equal(t, StringAlias("aliased string"), result.StringAliasField)
	assert.Equal(t, StringDeclaration("declared string"), result.StringDeclarationField)
	assert.Equal(t, []string{"a", "b", "c"}, result.StringListField)
	assert.Equal(t, []StringAlias{"x", "y", "z"}, result.StringAliasListField)
	assert.Equal(t, []StringDeclaration{"1", "2", "3"}, result.StringDeclListField)

	// Additional type checks
	assert.IsType(t, "", result.StringField)
	assert.IsType(t, StringAlias(""), result.StringAliasField)
	assert.IsType(t, StringDeclaration(""), result.StringDeclarationField)
	assert.IsType(t, []string{}, result.StringListField)
	assert.IsType(t, []StringAlias{}, result.StringAliasListField)
	assert.IsType(t, []StringDeclaration{}, result.StringDeclListField)
}

func TestValuesInitializeStructStringPointerTypes(t *testing.T) {
	// Define custom types
	type StringAlias string
	type StringDeclaration = string

	// Define the test struct with pointer fields
	type TestStruct struct {
		StringPtrField            *string            `glazed:"string_ptr_field"`
		StringAliasPtrField       *StringAlias       `glazed:"string_alias_ptr_field"`
		StringDeclarationPtrField *StringDeclaration `glazed:"string_declaration_ptr_field"`
	}

	// Create a field section with all the necessary definitions
	section := createSection(t, "test", "Test Section",
		fields.New("string_ptr_field", fields.TypeString),
		fields.New("string_alias_ptr_field", fields.TypeString),
		fields.New("string_declaration_ptr_field", fields.TypeString),
	)

	// Create a parsed section with test values
	sectionValues := createSectionValues(t, section, map[string]interface{}{
		"string_ptr_field":             "regular string",
		"string_alias_ptr_field":       "aliased string",
		"string_declaration_ptr_field": "declared string",
	})

	// Create Values and add the parsed section
	parsedValues := New(WithSectionValues("test", sectionValues))

	// Initialize the struct
	var result TestStruct
	err := parsedValues.DecodeSectionInto("test", &result)

	// Assert no error occurred
	assert.NoError(t, err)

	// Verify each field was correctly initialized
	assert.NotNil(t, result.StringPtrField)
	assert.NotNil(t, result.StringAliasPtrField)
	assert.NotNil(t, result.StringDeclarationPtrField)

	assert.Equal(t, "regular string", *result.StringPtrField)
	assert.Equal(t, StringAlias("aliased string"), *result.StringAliasPtrField)
	assert.Equal(t, StringDeclaration("declared string"), *result.StringDeclarationPtrField)

	// Additional type checks
	assert.IsType(t, (*string)(nil), result.StringPtrField)
	assert.IsType(t, (*StringAlias)(nil), result.StringAliasPtrField)
	assert.IsType(t, (*StringDeclaration)(nil), result.StringDeclarationPtrField)
}
