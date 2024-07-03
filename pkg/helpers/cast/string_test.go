package cast_test

import (
	. "github.com/go-go-golems/glazed/pkg/helpers/cast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MyStringAlias string
type MyString string

func TestToString_RegularString(t *testing.T) {
	input := "hello"
	expected := "hello"
	result, err := ToString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestToString_StringTypeAlias(t *testing.T) {
	input := MyStringAlias("hello")
	expected := "hello"
	result, err := ToString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestToString_StringTypeDeclaration(t *testing.T) {
	input := MyString("hello")
	expected := "hello"
	result, err := ToString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestToString_PointerToString(t *testing.T) {
	str := "hello"
	input := &str
	_, err := ToString(input)
	require.Error(t, err)
}

func TestToString_PointerToStringTypeAlias(t *testing.T) {
	str := MyStringAlias("hello")
	input := &str
	_, err := ToString(input)
	require.Error(t, err)
}

func TestToString_PointerToStringTypeDeclaration(t *testing.T) {
	str := MyString("hello")
	input := &str
	_, err := ToString(input)
	require.Error(t, err)
}

func TestToString_InterfaceHoldingInt(t *testing.T) {
	var input interface{} = 123
	_, err := ToString(input)
	require.Error(t, err)
}

func TestToString_InterfaceHoldingFloat(t *testing.T) {
	var input interface{} = 123.45
	_, err := ToString(input)
	require.Error(t, err)
}

func TestToString_InterfaceHoldingBool(t *testing.T) {
	var input interface{} = true
	_, err := ToString(input)
	require.Error(t, err)
}

func TestToString_InterfaceHoldingStruct(t *testing.T) {
	type MyStruct struct {
		Field string
	}
	var input interface{} = MyStruct{Field: "value"}
	_, err := ToString(input)
	require.Error(t, err)
}

func TestToString_InterfaceHoldingString(t *testing.T) {
	var input interface{} = "hello"
	expected := "hello"
	result, err := ToString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestToString_InterfaceHoldingStringTypeAlias(t *testing.T) {
	var input interface{} = MyStringAlias("hello")
	expected := "hello"
	result, err := ToString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestToString_InterfaceHoldingStringTypeDeclaration(t *testing.T) {
	var input interface{} = MyString("hello")
	expected := "hello"
	result, err := ToString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_RegularStringSlice(t *testing.T) {
	input := []string{"hello", "world"}
	expected := []string{"hello", "world"}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_StringTypeAliasSlice(t *testing.T) {
	input := []MyStringAlias{"hello", "world"}
	expected := []string{"hello", "world"}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_StringTypeDeclarationSlice(t *testing.T) {
	input := []MyString{"hello", "world"}
	expected := []string{"hello", "world"}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_MixedTypeSlice(t *testing.T) {
	input := []interface{}{"hello", MyStringAlias("world"), MyString("foo")}
	expected := []string{"hello", "world", "foo"}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_InterfaceHoldingString(t *testing.T) {
	input := []interface{}{"hello", "world"}
	expected := []string{"hello", "world"}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_InterfaceHoldingStringTypeAlias(t *testing.T) {
	input := []interface{}{MyStringAlias("hello"), MyStringAlias("world")}
	expected := []string{"hello", "world"}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_InterfaceHoldingStringTypeDeclaration(t *testing.T) {
	input := []interface{}{MyString("hello"), MyString("world")}
	expected := []string{"hello", "world"}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_InvalidTypeInSlice(t *testing.T) {
	input := []interface{}{"hello", 123, "world"}
	_, err := CastListToStringList(input)
	require.Error(t, err)
}

func TestCastListToStringList_NonSliceInput(t *testing.T) {
	input := "not a slice"
	_, err := CastListToStringList(input)
	require.Error(t, err)
}

func TestCastListToStringList_EmptySlice(t *testing.T) {
	input := []string{}
	expected := []string{}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastListToStringList_NilSlice(t *testing.T) {
	var input []string
	expected := []string{}
	result, err := CastListToStringList(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastStringListToList_StringTypeAlias(t *testing.T) {
	input := []string{"hello", "world"}
	expected := []MyStringAlias{"hello", "world"}
	result, err := CastStringListToList[MyStringAlias](input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastStringListToList_StringTypeDeclaration(t *testing.T) {
	input := []string{"hello", "world"}
	expected := []MyString{"hello", "world"}
	result, err := CastStringListToList[MyString](input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCastStringListToList_EmptyList(t *testing.T) {
	input := []string{}
	expected := []MyStringAlias{}
	result, err := CastStringListToList[MyStringAlias](input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
