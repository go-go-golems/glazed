package cmds

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

//go:embed "data/parameters_test.yaml"
var testFlagsYaml []byte

var testParameterDefinitions map[string]*ParameterDefinition
var testParameterDefinitionsList []*ParameterDefinition

func init() {
	testParameterDefinitions, testParameterDefinitionsList = InitFlagsFromYaml(testFlagsYaml)
}

func TestSetValueFromDefault(t *testing.T) {
	intFlag := testParameterDefinitions["int-flag"]

	testStruct := struct {
		Int int `glazed.parameter:"int-flag"`
	}{
		Int: 234,
	}
	_ = testStruct

	i := 234

	// get values of testStruct.Int
	iValue := reflect.ValueOf(&i).Elem()

	err := intFlag.SetValueFromDefault(iValue)
	require.NoError(t, err)
	assert.Equal(t, 42, i)

	intFlag = testParameterDefinitions["int-flag-without-default"]
	err = intFlag.SetValueFromDefault(iValue)
	require.NoError(t, err)
	assert.Equal(t, 0, i)

	i = 42

	intFlag = testParameterDefinitions["int-flag-with-empty-default"]
	err = intFlag.SetValueFromDefault(iValue)
	require.NoError(t, err)
	assert.Equal(t, i, 0)
}
