package parameters

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParameterString(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeString,
		WithDefault("default"),
	)

	i, err := parameter.ParseParameter([]string{"test"})
	require.NoError(t, err)
	assert.Equal(t, "test", i)

	_, err = parameter.ParseParameter([]string{"test", "test2"})
	assert.Error(t, err)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, "default", i)
}

func TestParameterStringList(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeStringList,
		WithDefault([]string{"default"}),
	)

	i, err := parameter.ParseParameter([]string{"test"})
	require.NoError(t, err)
	assert.Equal(t, []string{"test"}, i)

	i, err = parameter.ParseParameter([]string{"test", "test2"})
	require.NoError(t, err)
	assert.Equal(t, []string{"test", "test2"}, i)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, []string{"default"}, i)
}

func TestParameterInt(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeInteger,
		WithDefault(1),
	)

	i, err := parameter.ParseParameter([]string{"1"})
	require.NoError(t, err)
	assert.Equal(t, 1, i)

	_, err = parameter.ParseParameter([]string{"test"})
	assert.Error(t, err)

	_, err = parameter.ParseParameter([]string{"1", "2"})
	assert.Error(t, err)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, 1, i)
}

func TestParameterIntegerList(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeIntegerList,
		WithDefault([]int{1}),
	)

	i, err := parameter.ParseParameter([]string{"1"})
	require.NoError(t, err)
	assert.Equal(t, []int{1}, i)

	i, err = parameter.ParseParameter([]string{"1", "2"})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, i)

	_, err = parameter.ParseParameter([]string{"test"})
	assert.Error(t, err)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, []int{1}, i)
}

func TestParameterBool(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeBool,
		WithDefault(true),
	)

	i, err := parameter.ParseParameter([]string{"true"})
	require.NoError(t, err)
	assert.Equal(t, true, i)

	i, err = parameter.ParseParameter([]string{"false"})
	require.NoError(t, err)
	assert.Equal(t, false, i)

	_, err = parameter.ParseParameter([]string{"test"})
	assert.Error(t, err)

	_, err = parameter.ParseParameter([]string{"true", "false"})
	assert.Error(t, err)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, true, i)
}

func TestParameterFloat(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeFloat,
		WithDefault(1.0),
	)

	i, err := parameter.ParseParameter([]string{"1.0"})
	require.NoError(t, err)
	assert.Equal(t, 1.0, i)

	_, err = parameter.ParseParameter([]string{"test"})
	assert.Error(t, err)

	_, err = parameter.ParseParameter([]string{"1.0", "2.0"})
	assert.Error(t, err)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, 1.0, i)
}

func TestParameterFloatList(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeFloatList,
		WithDefault([]float64{1.0}),
	)

	i, err := parameter.ParseParameter([]string{"1.0"})
	require.NoError(t, err)
	assert.Equal(t, []float64{1.0}, i)

	i, err = parameter.ParseParameter([]string{"1.0", "2.0"})
	require.NoError(t, err)
	assert.Equal(t, []float64{1.0, 2.0}, i)

	_, err = parameter.ParseParameter([]string{"test"})
	assert.Error(t, err)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, []float64{1.0}, i)
}

func TestParameterChoice(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeChoice,
		WithDefault("default"),
		WithChoices([]string{"default", "test"}),
	)

	i, err := parameter.ParseParameter([]string{"test"})
	require.NoError(t, err)
	assert.Equal(t, "test", i)

	_, err = parameter.ParseParameter([]string{"test2"})
	assert.Error(t, err)

	_, err = parameter.ParseParameter([]string{"test", "test2"})
	assert.Error(t, err)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, "default", i)
}

func TestParameterTypeKeyValue(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeKeyValue,
		WithDefault(map[string]interface{}{"default": "default"}),
	)

	i, err := parameter.ParseParameter([]string{"test:test"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, i)

	i, err = parameter.ParseParameter([]string{"test:test", "test2:test2"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test", "test2": "test2"}, i)

	_, err = parameter.ParseParameter([]string{"test"})
	assert.Error(t, err)

	i, err = parameter.ParseParameter([]string{})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"default": "default"}, i)
}

func TestParseStringListFromReader(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeStringListFromFile,
		WithDefault([]string{"default"}),
	)

	reader := strings.NewReader("test\ntest2")
	i, err := parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{"test", "test2"}, i)

	reader = strings.NewReader("test")
	i, err = parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{"test"}, i)

	reader = strings.NewReader("")
	i, err = parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{}, i)
}

func TestParseObjectFromFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeObjectFromFile,
		WithDefault(map[string]interface{}{"default": "default"}),
	)

	reader := strings.NewReader(`{"test":"test"}`)
	i, err := parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, i)

	reader = strings.NewReader(`{"test":"test"`)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	reader = strings.NewReader(`{"test":{"test":"test"}}`)
	i, err = parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": map[string]interface{}{"test": "test"}}, i)

	reader = strings.NewReader(``)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	// toplevel list
	reader = strings.NewReader(`["test"]`)
	v, err := parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"test"}, v)

	// toplevel string
	reader = strings.NewReader(`"test"`)
	v, err = parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, "test", v)

	// toplevel int
	reader = strings.NewReader(`1`)
	v, err = parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, 1.0, v)

	// now yaml
	reader = strings.NewReader(`test: test`)
	i, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, i)

	reader = strings.NewReader(`test: test`)
	i, err = parameter.ParseFromReader(reader, "test.yml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, i)

	// nested object
	reader = strings.NewReader(`test: {test: test}`)
	i, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": map[string]interface{}{"test": "test"}}, i)

	// toplevel list
	reader = strings.NewReader("- test\n- test2")
	v, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"test", "test2"}, v)

	// toplevel string
	reader = strings.NewReader(`test`)
	v, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, "test", v)

	// now, one-line CSV with headers
	reader = strings.NewReader(`test,test2
test,test2`)
	i, err = parameter.ParseFromReader(reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test", "test2": "test2"}, i)

	// fail on 2 line CSV
	reader = strings.NewReader(`test,test2
test,test2
test,test2`)
	_, err = parameter.ParseFromReader(reader, "test.csv")
	assert.Error(t, err)

	// fail on CSV without headers
	reader = strings.NewReader(`test,test2`)
	_, err = parameter.ParseFromReader(reader, "test.csv")
	assert.Error(t, err)

	// fail on empty CSV
	reader = strings.NewReader(``)
	_, err = parameter.ParseFromReader(reader, "test.csv")
	assert.Error(t, err)

	// test TSV
	reader = strings.NewReader(`test	test2
test	test2`)
	i, err = parameter.ParseFromReader(reader, "test.tsv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test", "test2": "test2"}, i)

	// try numbers
	reader = strings.NewReader(`test,test2
1,2`)
	i, err = parameter.ParseFromReader(reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "1", "test2": "2"}, i)

	// try quoted numbers as strings
	reader = strings.NewReader(`test,test2
"1","2"`)
	i, err = parameter.ParseFromReader(reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "1", "test2": "2"}, i)
}

func TestParseObjectListFromFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeObjectListFromFile,
		WithDefault([]map[string]interface{}{{"default": "default"}}),
	)

	v, err := parseObjectListFromReader(parameter, `[{"test":"test"}]`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}}, v)

	// two elements
	v, err = parseObjectListFromReader(parameter, `[{"test":"test"},{"test2":"test2"}]`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}, {"test2": "test2"}}, v)

	_, err = parseObjectListFromReader(parameter, `{"test":"test"`, "test.json")
	assert.Error(t, err)

	v, err = parseObjectListFromReader(parameter, `[{"test":{"test":"test"}}]`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": map[string]interface{}{"test": "test"}}}, v)

	// succeed on empty list
	v, err = parseObjectListFromReader(parameter, `[]`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{}, v)

	// fail on empty file
	_, err = parseObjectListFromReader(parameter, ``, "test.json")
	assert.Error(t, err)

	// fail on toplevel list of string
	_, err = parseObjectListFromReader(parameter, `["test"]`, "test.json")
	assert.Error(t, err)

	// now yaml
	v, err = parseObjectListFromReader(parameter, `- test: test`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}}, v)

	v, err = parseObjectListFromReader(parameter, `- test: test`, "test.yml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}}, v)

	// two elements
	v, err = parseObjectListFromReader(parameter, `- test: test
- test2: test2`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}, {"test2": "test2"}}, v)

	// nested object
	v, err = parseObjectListFromReader(parameter, `- test: {test: test}`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": map[string]interface{}{"test": "test"}}}, v)

	// fail on toplevel list of strings
	_, err = parseObjectListFromReader(parameter, `- test
- test2`, "test.yaml")
	assert.Error(t, err)

	// fail on toplevel object
	_, err = parseObjectListFromReader(parameter, `test: test`, "test.yaml")
	assert.Error(t, err)

	// succeed on empty list
	v, err = parseObjectListFromReader(parameter, `[]`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{}, v)

	// fail on empty file
	_, err = parseObjectListFromReader(parameter, ``, "test.yaml")
	assert.Error(t, err)

	// test csv
	v, err = parseObjectListFromReader(parameter, `test,test2
test,test2`, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test", "test2": "test2"}}, v)

	// test csv with 2 lines
	v, err = parseObjectListFromReader(parameter, `test,test2
test,test2
test,test2`, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test", "test2": "test2"}, {"test": "test", "test2": "test2"}}, v)

	// fail on CSV with no headers
	_, err = parseObjectListFromReader(parameter, `test,test2`, "test.csv")
	assert.Error(t, err)

	// empty list on empty CSV
	v, err = parseObjectListFromReader(parameter, ``, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{}, v)

	// succeed on multiline CSV
	v, err = parseObjectListFromReader(parameter, `test,test2
test,test2
test,test2`, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test", "test2": "test2"}, {"test": "test", "test2": "test2"}}, v)
}

func parseObjectListFromReader(parameter *ParameterDefinition, input string, fileName string) ([]map[string]interface{}, error) {
	reader := strings.NewReader(input)
	i, err := parameter.ParseFromReader(reader, fileName)
	if err != nil {
		return nil, err
	}
	v, ok := cast.CastList[map[string]interface{}, interface{}](i.([]interface{}))
	if !ok {
		return nil, fmt.Errorf("failed to cast")
	}
	return v, nil
}

func parseObjectFromReader(parameter *ParameterDefinition, input string, fileName string) (map[string]interface{}, error) {
	reader := strings.NewReader(input)
	i, err := parameter.ParseFromReader(reader, fileName)
	if err != nil {
		return nil, err
	}
	v, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to cast")
	}
	return v, nil
}

func TestParseStringFromFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeStringFromFile,
		WithDefault("default"),
	)

	reader := strings.NewReader("test")
	i, err := parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "test", i)

	// multiline
	reader = strings.NewReader("test\ntest2")
	i, err = parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "test\ntest2", i)

	reader = strings.NewReader("")
	i, err = parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "", i)
}

func TestParseKeyFromFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeKeyValue,
		WithDefault("default"),
	)

	// from json
	v, err := parseObjectFromReader(parameter, `{"test":"test"}`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, v)

	v, err = parseObjectFromReader(parameter, `{"test":1}`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": 1.0}, v)

	v, err = parseObjectFromReader(parameter, `{"test":["test"]}`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": []interface{}{"test"}}, v)

	// succeed on empty dict
	v, err = parseObjectFromReader(parameter, `{}`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, v)

	// fail on empty file
	_, err = parseObjectFromReader(parameter, ``, "test.json")
	assert.Error(t, err)

	// yaml now
	v, err = parseObjectFromReader(parameter, `test: test`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, v)

	v, err = parseObjectFromReader(parameter, `test: 1`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": 1}, v)

	v, err = parseObjectFromReader(parameter, `test: ["test"]`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": []interface{}{"test"}}, v)

	// succeed on empty dict
	v, err = parseObjectFromReader(parameter, `{}`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, v)

	// fail on empty file
	_, err = parseObjectFromReader(parameter, ``, "test.yaml")
	assert.Error(t, err)

	// try CSV
	v, err = parseObjectFromReader(parameter, `test,test2
test,test2`, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test", "test2": "test2"}, v)
}
