package parameters

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"reflect"
	"testing"
	"time"
)

//go:embed "test-data/parameters_test.yaml"
var testFlagsYaml []byte

var testParameterDefinitions map[string]*ParameterDefinition
var testParameterDefinitionsList []*ParameterDefinition

type ValidityTest struct {
	Name                string        `yaml:"name"`
	Valid               bool          `yaml:"valid"`
	Type                ParameterType `yaml:"type"`
	Value               interface{}   `yaml:"value"`
	Choices             []string      `yaml:"choices,omitempty"`
	parameterDefinition *ParameterDefinition
}

//go:embed "test-data/parameters_validity_test.yaml"
var validityTestYaml []byte

var testParameterValidList []*ValidityTest

func loadValidityTestDataFromYAML(s []byte) ([]*ValidityTest, error) {
	var tests []*ValidityTest
	err := yaml.Unmarshal(s, &tests)
	if err != nil {
		return nil, err
	}

	for _, test := range tests {
		test.parameterDefinition = &ParameterDefinition{
			Name:     test.Name,
			Type:     test.Type,
			Default:  nil,
			Choices:  test.Choices,
			Required: true,
		}
	}

	return tests, nil
}

func init() {
	testParameterDefinitions, testParameterDefinitionsList = LoadParameterDefinitionsFromYAML(testFlagsYaml)
	var err error
	testParameterValidList, err = loadValidityTestDataFromYAML(validityTestYaml)
	if err != nil {
		panic(err)
	}
}

func TestParameterValidity(t *testing.T) {
	for _, validityTest := range testParameterValidList {
		t.Run(validityTest.Name, func(t *testing.T) {
			err := validityTest.parameterDefinition.CheckValueValidity(validityTest.Value)
			if validityTest.Valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
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

	floatFlag = testParameterDefinitions["float-flag-with-int-default"]
	err = floatFlag.SetValueFromDefault(fValue)
	require.NoError(t, err)
	assert.Equal(t, 42.0, f)
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

func TestSetValueFromDefaultChoice(t *testing.T) {
	choiceFlag := testParameterDefinitions["choice-flag"]

	c := "foo"

	// get values of testStruct.Choice
	cValue := reflect.ValueOf(&c).Elem()

	err := choiceFlag.SetValueFromDefault(cValue)
	require.NoError(t, err)
	assert.Equal(t, "default", c)

	choiceFlag = testParameterDefinitions["choice-flag-without-default"]
	err = choiceFlag.SetValueFromDefault(cValue)
	require.NoError(t, err)
	assert.Equal(t, "", c)

	choiceFlag = &ParameterDefinition{
		Name:    "choice-flag-with-invalid-default",
		Type:    ParameterTypeChoice,
		Default: "invalid",
		Choices: []string{"foo", "bar"},
	}
	err = choiceFlag.SetValueFromDefault(cValue)
	require.Error(t, err)
}

func TestSetValueFromDefaultChoiceList(t *testing.T) {
	choiceListFlag := testParameterDefinitions["choice-list-flag"]

	cl := []string{"foo", "bar"}

	// get values of testStruct.ChoiceList
	clValue := reflect.ValueOf(&cl).Elem()

	err := choiceListFlag.SetValueFromDefault(clValue)
	require.NoError(t, err)
	assert.Equal(t, []string{"default", "choice1", "choice2"}, cl)

	choiceListFlag = testParameterDefinitions["choice-list-flag-without-default"]
	err = choiceListFlag.SetValueFromDefault(clValue)
	require.NoError(t, err)
	assert.Equal(t, []string{}, cl)

	choiceListFlag = &ParameterDefinition{
		Name:    "choice-list-flag-with-invalid-default",
		Type:    ParameterTypeChoiceList,
		Default: []string{"invalid"},
		Choices: []string{"foo", "bar"},
	}
	err = choiceListFlag.SetValueFromDefault(clValue)
	require.Error(t, err)
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

func TestSetValueFromDefaultInt32List(t *testing.T) {
	intListFlag := testParameterDefinitions["int-list-flag"]
	require.NotNil(t, intListFlag)

	il := []int32{4, 5, 6}

	// get values of testStruct.IntList
	ilValue := reflect.ValueOf(&il).Elem()

	err := intListFlag.SetValueFromDefault(ilValue)
	require.NoError(t, err)
	assert.Equal(t, []int32{1, 2, 3}, il)

	intListFlag = testParameterDefinitions["int-list-flag-without-default"]
	require.NotNil(t, intListFlag)
	err = intListFlag.SetValueFromDefault(ilValue)
	require.NoError(t, err)
	assert.Equal(t, []int32{}, il)

	il = []int32{4, 5, 6}

	intListFlag = testParameterDefinitions["int-list-flag-with-empty-default"]
	require.NotNil(t, intListFlag)
	err = intListFlag.SetValueFromDefault(ilValue)
	require.NoError(t, err)
	assert.Equal(t, []int32{}, il)
}

func TestSetValueFromDefaultFloatList(t *testing.T) {
	floatListFlag := testParameterDefinitions["float-list-flag"]

	fl := []float64{4.0, 5.0, 6.0}

	// get values of testStruct.FloatList
	flValue := reflect.ValueOf(&fl).Elem()

	err := floatListFlag.SetValueFromDefault(flValue)
	require.NoError(t, err)
	assert.Equal(t, []float64{1.1, 2.2, 3.3, 4.0, 5.0}, fl)

	floatListFlag = testParameterDefinitions["float-list-flag-without-default"]
	err = floatListFlag.SetValueFromDefault(flValue)
	require.NoError(t, err)
	assert.Equal(t, []float64{}, fl)

	fl = []float64{4.0, 5.0, 6.0}

	floatListFlag = testParameterDefinitions["float-list-flag-with-empty-default"]
	err = floatListFlag.SetValueFromDefault(flValue)
	require.NoError(t, err)
	assert.Equal(t, []float64{}, fl)
}

func TestSetValueFromDefaultFloat32List(t *testing.T) {
	floatListFlag := testParameterDefinitions["float-list-flag"]
	require.NotNil(t, floatListFlag)

	fl := []float32{4.0, 5.0, 6.0}

	// get values of testStruct.FloatList
	flValue := reflect.ValueOf(&fl).Elem()

	err := floatListFlag.SetValueFromDefault(flValue)
	require.NoError(t, err)
	assert.Equal(t, []float32{1.1, 2.2, 3.3, 4, 5}, fl)

	floatListFlag = testParameterDefinitions["float-list-flag-without-default"]
	require.NotNil(t, floatListFlag)
	err = floatListFlag.SetValueFromDefault(flValue)
	require.NoError(t, err)
	assert.Equal(t, []float32{}, fl)

	fl = []float32{4.0, 5.0, 6.0}

	floatListFlag = testParameterDefinitions["float-list-flag-with-empty-default"]
	require.NotNil(t, floatListFlag)
	err = floatListFlag.SetValueFromDefault(flValue)
	require.NoError(t, err)
	assert.Equal(t, []float32{}, fl)
}

func TestSetValueFromDefaultObjectFromFile(t *testing.T) {
	objectFromFileFlag := testParameterDefinitions["object-from-file-flag"]

	fl := map[string]interface{}{"foo": "bar"}
	oValue := reflect.ValueOf(&fl).Elem()

	err := objectFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"name": "default", "value": 42}, fl)

	objectFromFileFlag = testParameterDefinitions["object-from-file-flag-without-default"]
	err = objectFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, fl)

	fl = map[string]interface{}{"foo": "bar"}

	objectFromFileFlag = testParameterDefinitions["object-from-file-flag-with-empty-default"]
	err = objectFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, fl)
}

func TestSetValueFromDefaultObjectListFromFile(t *testing.T) {
	objectListFromFileFlag := testParameterDefinitions["object-list-from-file-flag"]

	fl := []map[string]interface{}{{"foo": "bar"}}
	oValue := reflect.ValueOf(&fl).Elem()

	err := objectListFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{
		{"name": "default1", "value": 42},
		{"name": "default2", "value": 43},
	}, fl)

	objectListFromFileFlag = testParameterDefinitions["object-list-from-file-flag-without-default"]
	err = objectListFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{}, fl)

	fl = []map[string]interface{}{{"foo": "bar"}}

	objectListFromFileFlag = testParameterDefinitions["object-list-from-file-flag-with-empty-default"]
	err = objectListFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{}, fl)
}

func TestSetValueFromDefaultStringFromFile(t *testing.T) {
	stringFromFileFlag := testParameterDefinitions["string-from-file-flag"]

	fl := "foo"
	oValue := reflect.ValueOf(&fl).Elem()

	err := stringFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, "default", fl)

	stringFromFileFlag = testParameterDefinitions["string-from-file-flag-without-default"]
	err = stringFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, "", fl)

	fl = "foo"

	stringFromFileFlag = testParameterDefinitions["string-from-file-flag-with-empty-default"]
	err = stringFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, "", fl)
}

func TestSetValueFromDefaultStringListFromFile(t *testing.T) {
	stringListFromFileFlag := testParameterDefinitions["string-list-from-file-flag"]

	fl := []string{"foo"}
	oValue := reflect.ValueOf(&fl).Elem()

	err := stringListFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, []string{"default1", "default2"}, fl)

	stringListFromFileFlag = testParameterDefinitions["string-list-from-file-flag-without-default"]
	err = stringListFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, []string{}, fl)

	fl = []string{"foo"}

	stringListFromFileFlag = testParameterDefinitions["string-list-from-file-flag-with-empty-default"]
	err = stringListFromFileFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, []string{}, fl)
}

func TestSetValueFromDefaultKeyValue(t *testing.T) {
	keyValueFlag := testParameterDefinitions["key-value-flag"]

	fl := map[string]string{"foo": "bar"}
	oValue := reflect.ValueOf(&fl).Elem()

	err := keyValueFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, fl)

	keyValueFlag = testParameterDefinitions["key-value-flag-without-default"]
	err = keyValueFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{}, fl)

	fl = map[string]string{"foo": "bar"}

	keyValueFlag = testParameterDefinitions["key-value-flag-with-empty-default"]
	err = keyValueFlag.SetValueFromDefault(oValue)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{}, fl)
}
