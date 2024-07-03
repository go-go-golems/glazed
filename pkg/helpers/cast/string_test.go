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
