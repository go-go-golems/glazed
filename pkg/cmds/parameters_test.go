package cmds

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

//go:embed "data/parameters_test.yaml"
var testFlagsYaml []byte

var testParameterDefinitions map[string]*ParameterDefinition
var testParameterDefinitionsList []*ParameterDefinition

func init() {
	testParameterDefinitions, testParameterDefinitionsList = InitFlagsFromYaml(testFlagsYaml)
}

func TestSetValueFromDefaultInt(t *testing.T) {
	intFlag := testParameterDefinitions["int-flag"]

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

func TestSetValueFromDefaultInt32(t *testing.T) {
	intFlag := testParameterDefinitions["int-flag"]

	var i int32 = 234

	// get values of testStruct.Int
	iValue := reflect.ValueOf(&i).Elem()

	err := intFlag.SetValueFromDefault(iValue)
	require.NoError(t, err)
	assert.Equal(t, int32(42), i)

	intFlag = testParameterDefinitions["int-flag-without-default"]
	err = intFlag.SetValueFromDefault(iValue)
	require.NoError(t, err)
	assert.Equal(t, int32(0), i)

	i = 42

	intFlag = testParameterDefinitions["int-flag-with-empty-default"]
	err = intFlag.SetValueFromDefault(iValue)
	require.NoError(t, err)
	assert.Equal(t, int32(0), i)

}

func TestSetValueFromDefaultFloat(t *testing.T) {
	floatFlag := testParameterDefinitions["float-flag"]

	f := 234.0

	// get values of testStruct.Float
	fValue := reflect.ValueOf(&f).Elem()

	err := floatFlag.SetValueFromDefault(fValue)
	require.NoError(t, err)
	assert.Equal(t, 42.42, f)

	floatFlag = testParameterDefinitions["float-flag-without-default"]
	err = floatFlag.SetValueFromDefault(fValue)
	require.NoError(t, err)
	assert.Equal(t, 0.0, f)

	f = 42.0

	floatFlag = testParameterDefinitions["float-flag-with-empty-default"]
	err = floatFlag.SetValueFromDefault(fValue)
	require.NoError(t, err)
	assert.Equal(t, 0.0, f)
}

func TestSetValueFromDefaultFloat32(t *testing.T) {
	floatFlag := testParameterDefinitions["float-flag"]

	var f float32 = 234.0

	// get values of testStruct.Float
	fValue := reflect.ValueOf(&f).Elem()

	err := floatFlag.SetValueFromDefault(fValue)
	require.NoError(t, err)
	assert.Equal(t, float32(42.42), f)

	floatFlag = testParameterDefinitions["float-flag-without-default"]
	err = floatFlag.SetValueFromDefault(fValue)
	require.NoError(t, err)
	assert.Equal(t, float32(0.0), f)

	f = 42.0

	floatFlag = testParameterDefinitions["float-flag-with-empty-default"]
	err = floatFlag.SetValueFromDefault(fValue)
	require.NoError(t, err)
	assert.Equal(t, float32(0.0), f)
}

func TestSetValueFromDefaultDate(t *testing.T) {
	dateFlag := testParameterDefinitions["date-flag"]

	d := time.Now()

	// get values of testStruct.Date
	dValue := reflect.ValueOf(&d).Elem()

	parsedTime, err := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	require.NoError(t, err)

	err = dateFlag.SetValueFromDefault(dValue)
	require.NoError(t, err)
	assert.Equal(t, parsedTime, d)

	dateFlag = testParameterDefinitions["date-flag-without-default"]
	err = dateFlag.SetValueFromDefault(dValue)
	require.NoError(t, err)
	assert.Equal(t, time.Time{}, d)
}

func TestSetValueFromDefaultString(t *testing.T) {
	stringFlag := testParameterDefinitions["string-flag"]

	s := "test"

	// get values of testStruct.String
	sValue := reflect.ValueOf(&s).Elem()

	err := stringFlag.SetValueFromDefault(sValue)
	require.NoError(t, err)
	assert.Equal(t, "default", s)

	stringFlag = testParameterDefinitions["string-flag-without-default"]
	err = stringFlag.SetValueFromDefault(sValue)
	require.NoError(t, err)
	assert.Equal(t, "", s)

	s = "foo"

	stringFlag = testParameterDefinitions["string-flag-with-empty-default"]
	err = stringFlag.SetValueFromDefault(sValue)
	require.NoError(t, err)
	assert.Equal(t, "", s)
}

func TestSetValueFromDefaultBool(t *testing.T) {
	boolFlag := testParameterDefinitions["bool-flag"]

	b := false

	// get values of testStruct.Bool
	bValue := reflect.ValueOf(&b).Elem()

	err := boolFlag.SetValueFromDefault(bValue)
	require.NoError(t, err)
	assert.Equal(t, true, b)

	boolFlag = testParameterDefinitions["bool-flag-without-default"]
	err = boolFlag.SetValueFromDefault(bValue)
	require.NoError(t, err)
	assert.Equal(t, false, b)

	b = true

	boolFlag = testParameterDefinitions["bool-flag-with-empty-default"]
	err = boolFlag.SetValueFromDefault(bValue)
	require.NoError(t, err)
	assert.Equal(t, false, b)
}

func TestSetValueFromDefaultIntList(t *testing.T) {
	intListFlag := testParameterDefinitions["int-list-flag"]

	il := []int{4, 5, 6}

	// get values of testStruct.IntList
	ilValue := reflect.ValueOf(&il).Elem()

	err := intListFlag.SetValueFromDefault(ilValue)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, il)

	intListFlag = testParameterDefinitions["int-list-flag-without-default"]
	err = intListFlag.SetValueFromDefault(ilValue)
	require.NoError(t, err)
	assert.Equal(t, []int{}, il)

	il = []int{4, 5, 6}

	intListFlag = testParameterDefinitions["int-list-flag-with-empty-default"]
	err = intListFlag.SetValueFromDefault(ilValue)
	require.NoError(t, err)
	assert.Equal(t, []int{}, il)
}
