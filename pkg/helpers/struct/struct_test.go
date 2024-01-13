package _struct

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// Define a simple struct for testing
type SimpleStruct struct {
	IntField    int
	StringField string
}

// Define a nested struct for testing
type NestedStruct struct {
	Simple  SimpleStruct
	Pointer *SimpleStruct
}

func TestCloneStructWithSimpleStruct(t *testing.T) {
	original := SimpleStruct{IntField: 42, StringField: "Hello, world!"}
	cloned := CloneStruct(original).(SimpleStruct)

	assert.Equal(t, original, cloned, "Cloned struct should match the original")
}

func TestCloneStructWithNestedStruct(t *testing.T) {
	original := NestedStruct{
		Simple:  SimpleStruct{IntField: 42, StringField: "Hello, world!"},
		Pointer: &SimpleStruct{IntField: 100, StringField: "Nested struct"},
	}
	cloned := CloneStruct(original).(NestedStruct)

	assert.Equal(t, original, cloned, "Cloned struct should match the original")
	assert.False(t, &original.Pointer == &cloned.Pointer, "Cloned struct's pointer field should not point to the same address as the original")
}

func TestCloneStructWithPointerFields(t *testing.T) {
	original := &SimpleStruct{IntField: 42, StringField: "Hello, world!"}
	cloned := CloneStruct(original).(SimpleStruct)

	assert.Equal(t, *original, cloned, "Cloned struct should match the original")
}

func TestCloneStructWithUnexportedFields(t *testing.T) {
	original := StructWithUnexportedFields{ExportedField: "Visible", unexportedField: 42}
	cloned := CloneStruct(original).(StructWithUnexportedFields)

	assert.Equal(t, original.ExportedField, cloned.ExportedField, "Cloned struct's exported field should match the original")

	originalValue := reflect.ValueOf(original).FieldByName("unexportedField").Int()
	clonedValue := reflect.ValueOf(cloned).FieldByName("unexportedField").Int()

	assert.NotEqual(t, originalValue, clonedValue, "Cloned struct's unexported field should not be cloned")
}

func TestCloneStructWithSlicesAndMaps(t *testing.T) {
	original := StructWithSlicesAndMaps{
		SliceField: []int{1, 2, 3},
		MapField:   map[string]int{"one": 1, "two": 2},
	}
	cloned := CloneStruct(original).(StructWithSlicesAndMaps)

	assert.Equal(t, original.SliceField, cloned.SliceField, "Cloned struct's slice field should match the original")
	assert.Equal(t, original.MapField, cloned.MapField, "Cloned struct's map field should match the original")

	assert.True(t, &original.SliceField[0] == &cloned.SliceField[0], "Cloned struct's slice field should point to the same underlying data as the original")
	assert.True(t, areMapsSame(original.MapField, cloned.MapField), "Cloned struct's map field should point to the same underlying data as the original")
}

func TestCloneStructWithEmptyStruct(t *testing.T) {
	original := EmptyStruct{}
	cloned := CloneStruct(original).(EmptyStruct)

	assert.Equal(t, original, cloned, "Cloned empty struct should match the original")
}

func TestCloneStructWithNonStructInput(t *testing.T) {
	original := "I am not a struct"
	cloned := CloneStruct(original)

	assert.Nil(t, cloned, "Cloned value should be nil when non-struct input is provided")
}

func TestCloneStructWithPrimitiveTypes(t *testing.T) {
	originalInt := 42
	clonedInt := CloneStruct(originalInt)

	assert.Nil(t, clonedInt, "Cloned value should be nil when primitive type input is provided")

	originalString := "Hello, world!"
	clonedString := CloneStruct(originalString)

	assert.Nil(t, clonedString, "Cloned value should be nil when primitive type input is provided")
}

func TestCloneStructWithZeroValues(t *testing.T) {
	original := StructWithVariousTypes{}
	cloned := CloneStruct(original).(StructWithVariousTypes)

	assert.Equal(t, original, cloned, "Cloned struct with zero values should match the original")
}

func TestCloneStructWithInterfaceFields(t *testing.T) {
	original := StructWithInterfaceField{Field: "I am an interface field"}
	cloned := CloneStruct(original).(StructWithInterfaceField)

	assert.Equal(t, original, cloned, "Cloned struct with interface field should match the original")
	assert.IsType(t, original.Field, cloned.Field, "Cloned struct's interface field should be of the expected type")
}

type StructWithUnexportedFields struct {
	ExportedField   string
	unexportedField int
}

// Define a struct with slices and maps for testing
type StructWithSlicesAndMaps struct {
	SliceField []int
	MapField   map[string]int
}

// Define an empty struct for testing
type EmptyStruct struct{}

func areMapsSame(a, b map[string]int) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

// Define a struct with different types of fields for testing zero values
type StructWithVariousTypes struct {
	IntField     int
	StringField  string
	SliceField   []int
	PointerField *int
}

// Define a struct with an interface field for testing
type StructWithInterfaceField struct {
	Field interface{}
}
