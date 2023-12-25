package parameters

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

type ExpectError string

const ErrorExpected ExpectError = "ErrorExpected"
const ErrorNotExpected ExpectError = "ErrorNotExpected"

type ParameterTestCase struct {
	Name     string
	Input    []string
	Expected interface{}
	WantErr  ExpectError
}

func TestMain(m *testing.M) {
	// Set default time for unit tests
	refTime_ := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local)
	refTime = &refTime_

	m.Run()
}

func getLastYear() time.Time {
	now := *refTime
	lastYear := now.AddDate(-1, 0, 0) // Subtract one year from the current date
	return lastYear
}

func getNextMondayAt3PM() time.Time {
	now := *refTime
	// Calculate the number of days to add to get to the next Monday
	daysUntilNextMonday := (8 - int(now.Weekday())) % 7
	if daysUntilNextMonday == 0 {
		// If today is Monday and it's past 3 PM, set to next Monday
		daysUntilNextMonday = 7
	}

	nextMonday := now.AddDate(0, 0, daysUntilNextMonday)
	// Set the time to 15:00:00
	return time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 15, 0, 0, 0, nextMonday.Location())
}

func getTomorrow() time.Time {
	now := *refTime
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

// todayWithTime creates a new time.Time object with today's date but with the specified hour, minute, second, and nanosecond values.
func todayWithTime(hour, minute, second, nanosecond int) time.Time {
	now := *refTime
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, nanosecond, now.Location())
}

type ParameterTest struct {
	Name          string
	ParameterType ParameterType
	DefaultValue  interface{}
	Choices       []string
	Cases         []ParameterTestCase
}

func TestParameterDate(t *testing.T) {
	cases := []ParameterTestCase{
		{
			Name:     "Natural language input for past date, no error expected",
			Input:    []string{"last year"},
			Expected: getLastYear(),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Natural language date and time input, no error expected",
			Input:    []string{"next Monday at 3 PM"},
			Expected: getNextMondayAt3PM(),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Valid full date and time input, no error expected",
			Input:    []string{"2023-12-01 15:00"},
			Expected: time.Date(2023, time.December, 01, 15, 00, 0, 0, time.Local), // Adjust time zone as required
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Valid date only input, no error expected",
			Input:    []string{"2023-12-01"},
			Expected: time.Date(2023, time.December, 01, 0, 0, 0, 0, time.Local), // Time defaults to 00:00
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Natural language date input, no error expected",
			Input:    []string{"tomorrow"},
			Expected: getTomorrow(),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Invalid date format, error expected",
			Input:    []string{"2023/12/01"},
			Expected: time.Date(2023, time.December, 01, 0, 0, 0, 0, time.Local), // Time defaults to 00:00
		},
		{
			Name:    "Invalid non-date string, error expected",
			Input:   []string{"oisdjfoisudof,,inot a date"},
			WantErr: ErrorExpected,
		},
		{
			Name:     "No input uses default date, no error expected",
			Input:    []string{},
			Expected: *refTime,
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Valid date and time with seconds, no error expected",
			Input:    []string{"2023-12-01 15:00:30"},
			Expected: time.Date(2023, time.December, 01, 15, 00, 30, 0, time.Local), // Adjust time zone as required
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Natural language date input with alternative format, no error expected",
			Input:    []string{"December 1st, 2023"},
			Expected: time.Date(2023, time.December, 01, 0, 0, 0, 0, time.Local),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Valid time with hour and minute, no error expected",
			Input:    []string{"15:30"},
			Expected: todayWithTime(15, 30, 0, 0), // Function to set today's date with specific time
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Valid time with hour, minute and seconds, no error expected",
			Input:    []string{"23:59:59"},
			Expected: todayWithTime(23, 59, 59, 0),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Valid time with hour only, no error expected",
			Input:    []string{"8 AM"},
			Expected: todayWithTime(8, 0, 0, 0),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "12-hour format with AM/PM, no error expected",
			Input:    []string{"3 PM"},
			Expected: todayWithTime(15, 0, 0, 0),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:    "Invalid time format (negative minute), error expected",
			Input:   []string{"10:-30"},
			WantErr: ErrorExpected,
		},
		{
			Name:     "Valid ISO date and time, no error expected",
			Input:    []string{"2023-12-01T15:00:00Z"},
			Expected: time.Date(2023, time.December, 01, 15, 00, 00, 0, time.UTC),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Valid date and time with space separator, no error expected",
			Input:    []string{"2023-12-01 15:00:00"},
			Expected: time.Date(2023, time.December, 01, 15, 00, 00, 0, time.Local),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:     "Date only, no error expected",
			Input:    []string{"2023-12-01"},
			Expected: time.Date(2023, time.December, 01, 0, 0, 0, 0, time.Local),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:    "Invalid time format in date-time combination, error expected",
			Input:   []string{"2023-12-01 25:00:00"},
			WantErr: ErrorExpected,
		},
		{
			Name:     "Valid 12-hour format date and time, no error expected",
			Input:    []string{"December 1, 2023, 3:00 PM"},
			Expected: time.Date(2023, time.December, 01, 15, 00, 00, 0, time.Local),
			WantErr:  ErrorNotExpected,
		},
		{
			Name:  "Date and time with timezone, no error expected",
			Input: []string{"2023-12-01T15:00:00-05:00"}, // Eastern Time
			Expected: func() time.Time {
				t, _ := time.Parse(time.RFC3339, "2023-12-01T15:00:00-05:00")
				return t
			}(),
			WantErr: ErrorNotExpected,
		},
	}

	parameter := NewParameterDefinition(
		"test",
		ParameterTypeDate,
		WithDefault(*refTime),
	)

	for _, tc := range cases {
		t.Run(fmt.Sprintf("Date: %s", tc.Name), func(t *testing.T) {
			got, err := parameter.ParseParameter(tc.Input)
			if tc.WantErr == ErrorExpected {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// clean out milliseconds
				expected := tc.Expected.(time.Time).Truncate(time.Second)
				got_ := got.Value.(time.Time).Truncate(time.Second)
				assert.Equal(t, expected, got_)
			}
		})
	}
}

func TestParameters(t *testing.T) {
	tests := []ParameterTest{
		{
			Name:          "ParameterString",
			ParameterType: ParameterTypeString,
			DefaultValue:  "default",
			Cases: []ParameterTestCase{
				{Name: "Valid single string, no error expected", Input: []string{"test"}, Expected: "test", WantErr: ErrorNotExpected},
				{Name: "Multiple strings for non-list, error expected", Input: []string{"test", "test2"}, WantErr: ErrorExpected},
				{Name: "No input uses default, no error expected", Input: []string{}, Expected: "default", WantErr: ErrorNotExpected},
			},
		},
		{
			Name:          "ParameterStringList",
			ParameterType: ParameterTypeStringList,
			DefaultValue:  []string{"default"},
			Cases: []ParameterTestCase{
				{Name: "Valid single string in list, no error expected", Input: []string{"test"}, Expected: []string{"test"}, WantErr: ErrorNotExpected},
				{Name: "Valid multiple strings in list, no error expected", Input: []string{"test", "test2"}, Expected: []string{"test", "test2"}, WantErr: ErrorNotExpected},
				{Name: "No input uses default list, no error expected", Input: []string{}, Expected: []string{"default"}, WantErr: ErrorNotExpected},
			},
		},
		{
			Name:          "ParameterInt",
			ParameterType: ParameterTypeInteger,
			DefaultValue:  1,
			Cases: []ParameterTestCase{
				{Name: "Valid integer string, no error expected", Input: []string{"1"}, Expected: 1, WantErr: ErrorNotExpected},
				{Name: "Invalid non-integer string, error expected", Input: []string{"test"}, WantErr: ErrorExpected},
				{Name: "Multiple integers for non-list, error expected", Input: []string{"1", "2"}, WantErr: ErrorExpected},
				{Name: "No input uses default integer, no error expected", Input: []string{}, Expected: 1, WantErr: ErrorNotExpected},
			},
		},
		{
			Name:          "ParameterIntegerList",
			ParameterType: ParameterTypeIntegerList,
			DefaultValue:  []int{1},
			Cases: []ParameterTestCase{
				{Name: "Valid single integer in list, no error expected", Input: []string{"1"}, Expected: []int{1}, WantErr: ErrorNotExpected},
				{Name: "Valid multiple integers in list, no error expected", Input: []string{"1", "2"}, Expected: []int{1, 2}, WantErr: ErrorNotExpected},
				{Name: "Invalid non-integer string in list, error expected", Input: []string{"test"}, WantErr: ErrorExpected},
				{Name: "No input uses default integer list, no error expected", Input: []string{}, Expected: []int{1}, WantErr: ErrorNotExpected},
			},
		},

		{
			Name:          "ParameterBool",
			ParameterType: ParameterTypeBool,
			DefaultValue:  true,
			Cases: []ParameterTestCase{
				{Name: "Valid 'true' input, no error expected", Input: []string{"true"}, Expected: true, WantErr: ErrorNotExpected},
				{Name: "Valid 'false' input, no error expected", Input: []string{"false"}, Expected: false, WantErr: ErrorNotExpected},
				{Name: "Invalid non-boolean string, error expected", Input: []string{"test"}, WantErr: ErrorExpected},
				{Name: "Multiple boolean values, error expected", Input: []string{"true", "false"}, WantErr: ErrorExpected},
				{Name: "No input uses default boolean, no error expected", Input: []string{}, Expected: true, WantErr: ErrorNotExpected},
			},
		},
		{
			Name:          "ParameterFloat",
			ParameterType: ParameterTypeFloat,
			DefaultValue:  1.0,
			Cases: []ParameterTestCase{
				{Name: "Valid float input, no error expected", Input: []string{"1.0"}, Expected: 1.0, WantErr: ErrorNotExpected},
				{Name: "Invalid non-float string, error expected", Input: []string{"test"}, WantErr: ErrorExpected},
				{Name: "Multiple floats for non-list, error expected", Input: []string{"1.0", "2.0"}, WantErr: ErrorExpected},
				{Name: "No input uses default float, no error expected", Input: []string{}, Expected: 1.0, WantErr: ErrorNotExpected},
			},
		},
		{
			Name:          "ParameterFloatList",
			ParameterType: ParameterTypeFloatList,
			DefaultValue:  []float64{1.0},
			Cases: []ParameterTestCase{
				{Name: "Valid single float in list, no error expected", Input: []string{"1.0"}, Expected: []float64{1.0}, WantErr: ErrorNotExpected},
				{Name: "Valid multiple floats in list, no error expected", Input: []string{"1.0", "2.0"}, Expected: []float64{1.0, 2.0}, WantErr: ErrorNotExpected},
				{Name: "Invalid non-float string in list, error expected", Input: []string{"test"}, WantErr: ErrorExpected},
				{Name: "No input uses default float list, no error expected", Input: []string{}, Expected: []float64{1.0}, WantErr: ErrorNotExpected},
			},
		},
		{
			Name:          "ParameterChoice",
			ParameterType: ParameterTypeChoice,
			DefaultValue:  "default",
			Choices:       []string{"default", "test"},
			Cases: []ParameterTestCase{
				{Name: "Valid choice from list, no error expected", Input: []string{"test"}, Expected: "test", WantErr: ErrorNotExpected},
				{Name: "Invalid choice not in list, error expected", Input: []string{"test2"}, WantErr: ErrorExpected},
				{Name: "Multiple choices for non-list, error expected", Input: []string{"test", "test2"}, WantErr: ErrorExpected},
				{Name: "No input uses default choice, no error expected", Input: []string{}, Expected: "default", WantErr: ErrorNotExpected},
			},
		},
		{
			Name:          "ParameterChoiceList",
			ParameterType: ParameterTypeChoiceList,
			DefaultValue:  []string{"default"},
			Choices:       []string{"default", "test", "option1", "option2"},
			Cases: []ParameterTestCase{
				{Name: "Valid single choice in list, no error expected", Input: []string{"test"}, Expected: []string{"test"}, WantErr: ErrorNotExpected},
				{Name: "Valid multiple choices in list, no error expected", Input: []string{"test", "option1"}, Expected: []string{"test", "option1"}, WantErr: ErrorNotExpected},
				{Name: "Invalid single choice not in list, error expected", Input: []string{"test2"}, WantErr: ErrorExpected},
				{Name: "Mixed valid and invalid choices, error expected", Input: []string{"test", "test2"}, WantErr: ErrorExpected},
				{Name: "No input uses default choice list, no error expected", Input: []string{}, Expected: []string{"default"}, WantErr: ErrorNotExpected},
			},
		},
		{
			Name:          "ParameterTypeKeyValue",
			ParameterType: ParameterTypeKeyValue,
			DefaultValue:  map[string]interface{}{"default": "default"},
			Cases: []ParameterTestCase{
				{Name: "Valid single key-value pair, no error expected", Input: []string{"test:test"}, Expected: map[string]interface{}{"test": "test"}, WantErr: ErrorNotExpected},
				{Name: "Valid multiple key-value pairs, no error expected", Input: []string{"test:test", "test2:test2"}, Expected: map[string]interface{}{"test": "test", "test2": "test2"}, WantErr: ErrorNotExpected},
				{Name: "Invalid input without colon separator, error expected", Input: []string{"test"}, WantErr: ErrorExpected},
				{Name: "No input uses default key-value map, no error expected", Input: []string{}, Expected: map[string]interface{}{"default": "default"}, WantErr: ErrorNotExpected},
			},
		},
	}

	for _, tt := range tests {
		parameter := NewParameterDefinition(
			"test",
			tt.ParameterType,
			WithChoices(tt.Choices),
			WithDefault(tt.DefaultValue))

		for _, tc := range tt.Cases {
			t.Run(fmt.Sprintf("%s: %s", tt.Name, tc.Name), func(t *testing.T) {
				got, err := parameter.ParseParameter(tc.Input)
				if tc.WantErr == ErrorExpected {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tc.Expected, got.Value)
				}
			})
		}
	}
}

func TestParseStringListFromReader(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeStringListFromFile,
		WithDefault([]string{"default"}),
	)

	reader := strings.NewReader("test\ntest2")
	i, err := parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{"test", "test2"}, i.Value)

	reader = strings.NewReader("test")
	i, err = parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{"test"}, i.Value)

	reader = strings.NewReader("")
	i, err = parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{}, i.Value)

	// try single column CSV with header
	reader = strings.NewReader("test\ntest2\ntest3\ntest4")
	i, err = parameter.ParseFromReader(reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []string{"test2", "test3", "test4"}, i.Value)

	// test single string list json
	reader = strings.NewReader(`["test","test2"]`)
	i, err = parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []string{"test", "test2"}, i.Value)

	// fail single string
	reader = strings.NewReader(`"test"`)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	// test fail int
	reader = strings.NewReader(`1`)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	// test fail int list
	reader = strings.NewReader(`[1,2]`)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	// test fail mixed list
	reader = strings.NewReader(`["test",1]`)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	// test fail empty json
	reader = strings.NewReader(`{}`)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	// test succeed empty list
	reader = strings.NewReader(`[]`)
	i, err = parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []string{}, i.Value)

	// test yaml
	reader = strings.NewReader(`- test
- test2`)
	i, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []string{"test", "test2"}, i.Value)

	// test empty csv (just headers)
	reader = strings.NewReader(`test`)
	i, err = parameter.ParseFromReader(reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []string{}, i.Value)

}

func TestParseObjectFromFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeObjectFromFile,
		WithDefault(map[string]interface{}{"default": "default"}),
	)

	reader := strings.NewReader(`{"test":"test"}`)
	i, err := parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, i.Value)

	reader = strings.NewReader(`{"test":"test"`)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	reader = strings.NewReader(`{"test":{"test":"test"}}`)
	i, err = parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": map[string]interface{}{"test": "test"}}, i.Value)

	reader = strings.NewReader(``)
	_, err = parameter.ParseFromReader(reader, "test.json")
	assert.Error(t, err)

	// toplevel list
	reader = strings.NewReader(`["test"]`)
	v, err := parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"test"}, v.Value)

	// toplevel string
	reader = strings.NewReader(`"test"`)
	v, err = parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, "test", v.Value)

	// toplevel int
	reader = strings.NewReader(`1`)
	v, err = parameter.ParseFromReader(reader, "test.json")
	require.NoError(t, err)
	assert.Equal(t, 1.0, v.Value)

	// now yaml
	reader = strings.NewReader(`test: test`)
	i, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, i.Value)

	reader = strings.NewReader(`test: test`)
	i, err = parameter.ParseFromReader(reader, "test.yml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, i.Value)

	// nested object
	reader = strings.NewReader(`test: {test: test}`)
	i, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": map[string]interface{}{"test": "test"}}, i.Value)

	// toplevel list
	reader = strings.NewReader("- test\n- test2")
	v, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"test", "test2"}, v.Value)

	// toplevel string
	reader = strings.NewReader(`test`)
	v, err = parameter.ParseFromReader(reader, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, "test", v.Value)

	// now, one-line CSV with headers
	reader = strings.NewReader(`test,test2
test,test2`)
	i, err = parameter.ParseFromReader(reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test", "test2": "test2"}, i.Value)

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
	assert.Equal(t, map[string]interface{}{"test": "test", "test2": "test2"}, i.Value)

	// try numbers
	reader = strings.NewReader(`test,test2
1,2`)
	i, err = parameter.ParseFromReader(reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "1", "test2": "2"}, i.Value)

	// try quoted numbers as strings
	reader = strings.NewReader(`test,test2
"1","2"`)
	i, err = parameter.ParseFromReader(reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "1", "test2": "2"}, i.Value)
}

func TestParseObjectListFromFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeObjectListFromFile,
		WithDefault([]map[string]interface{}{{"default": "default"}}),
	)

	v, err := parseObjectListFromString(parameter, `[{"test":"test"}]`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}}, v)

	// two elements
	v, err = parseObjectListFromString(parameter, `[{"test":"test"},{"test2":"test2"}]`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}, {"test2": "test2"}}, v)

	_, err = parseObjectListFromString(parameter, `{"test":"test"`, "test.json")
	assert.Error(t, err)

	v, err = parseObjectListFromString(parameter, `[{"test":{"test":"test"}}]`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": map[string]interface{}{"test": "test"}}}, v)

	// succeed on empty list
	v, err = parseObjectListFromString(parameter, `[]`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{}, v)

	// fail on empty file
	_, err = parseObjectListFromString(parameter, ``, "test.json")
	assert.Error(t, err)

	// fail on toplevel list of string
	_, err = parseObjectListFromString(parameter, `["test"]`, "test.json")
	assert.Error(t, err)

	// now yaml
	v, err = parseObjectListFromString(parameter, `- test: test`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}}, v)

	v, err = parseObjectListFromString(parameter, `- test: test`, "test.yml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}}, v)

	// two elements
	v, err = parseObjectListFromString(parameter, `- test: test
- test2: test2`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}, {"test2": "test2"}}, v)

	// nested object
	v, err = parseObjectListFromString(parameter, `- test: {test: test}`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": map[string]interface{}{"test": "test"}}}, v)

	// fail on toplevel list of strings
	_, err = parseObjectListFromString(parameter, `- test
- test2`, "test.yaml")
	assert.Error(t, err)

	// fail on toplevel object
	v, err = parseObjectListFromString(parameter, `test: test`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test"}}, v)

	// succeed on empty list
	v, err = parseObjectListFromString(parameter, `[]`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{}, v)

	// fail on empty file
	_, err = parseObjectListFromString(parameter, ``, "test.yaml")
	assert.Error(t, err)

	// test csv
	v, err = parseObjectListFromString(parameter, `test,test2
test,test2`, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test", "test2": "test2"}}, v)

	// test csv with 2 lines
	v, err = parseObjectListFromString(parameter, `test,test2
test,test2
test,test2`, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test", "test2": "test2"}, {"test": "test", "test2": "test2"}}, v)

	// fail on CSV with no headers
	_, err = parseObjectListFromString(parameter, `test,test2`, "test.csv")
	assert.Error(t, err)

	// empty list on empty CSV
	v, err = parseObjectListFromString(parameter, ``, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{}, v)

	// succeed on multiline CSV
	v, err = parseObjectListFromString(parameter, `test,test2
test,test2
test,test2`, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"test": "test", "test2": "test2"}, {"test": "test", "test2": "test2"}}, v)
}

func parseObjectListFromString(parameter *ParameterDefinition, input string, fileName string) ([]map[string]interface{}, error) {
	reader := strings.NewReader(input)
	i, err := parameter.ParseFromReader(reader, fileName)
	if err != nil {
		return nil, err
	}
	v, ok := cast.CastList[map[string]interface{}, interface{}](i.Value.([]interface{}))
	if !ok {
		return nil, fmt.Errorf("failed to cast")
	}
	return v, nil
}

func parseObjectFromString(parameter *ParameterDefinition, input string, fileName string) (map[string]interface{}, error) {
	reader := strings.NewReader(input)
	i, err := parameter.ParseFromReader(reader, fileName)
	if err != nil {
		return nil, err
	}
	v, ok := i.Value.(map[string]interface{})
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
	assert.Equal(t, "test", i.Value)

	// multiline
	reader = strings.NewReader("test\ntest2")
	i, err = parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "test\ntest2", i.Value)

	reader = strings.NewReader("")
	i, err = parameter.ParseFromReader(reader, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "", i.Value)
}

func TestParseKeyFromFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeKeyValue,
		WithDefault("default"),
	)

	// from json
	v, err := parseObjectFromString(parameter, `{"test":"test"}`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, v)

	v, err = parseObjectFromString(parameter, `{"test":1}`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": 1.0}, v)

	v, err = parseObjectFromString(parameter, `{"test":["test"]}`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": []interface{}{"test"}}, v)

	// succeed on empty dict
	v, err = parseObjectFromString(parameter, `{}`, "test.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, v)

	// fail on empty file
	_, err = parseObjectFromString(parameter, ``, "test.json")
	assert.Error(t, err)

	// yaml now
	v, err = parseObjectFromString(parameter, `test: test`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test"}, v)

	v, err = parseObjectFromString(parameter, `test: 1`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": 1}, v)

	v, err = parseObjectFromString(parameter, `test: ["test"]`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": []interface{}{"test"}}, v)

	// succeed on empty dict
	v, err = parseObjectFromString(parameter, `{}`, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, v)

	// fail on empty file
	_, err = parseObjectFromString(parameter, ``, "test.yaml")
	assert.Error(t, err)

	// try CSV
	v, err = parseObjectFromString(parameter, `test,test2
test,test2`, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "test", "test2": "test2"}, v)
}

func TestParseStringFromFileRealFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeStringFromFile,
		WithDefault("default"),
	)

	v, err := parameter.ParseParameter([]string{"test-data/string.txt"})
	require.NoError(t, err)
	assert.Equal(t, "string1\n", v.Value)

	parameter = NewParameterDefinition("test", ParameterTypeStringFromFiles,
		WithDefault("default"),
	)
	v, err = parameter.ParseParameter([]string{"test-data/string.txt"})
	require.NoError(t, err)
	assert.Equal(t, "string1\n", v.Value)

	v, err = parameter.ParseParameter([]string{"test-data/string.txt", "test-data/string2.txt"})
	require.NoError(t, err)
	assert.Equal(t, "string1\nstring2\n", v.Value)
}

func TestParseStringListFromFileRealFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeStringListFromFile,
		WithDefault([]string{"default"}),
	)

	v, err := parameter.ParseParameter([]string{"test-data/string.txt"})
	require.NoError(t, err)
	assert.Equal(t, []string{"string1"}, v.Value)

	v, err = parameter.ParseParameter([]string{"test-data/stringList.csv"})
	require.NoError(t, err)
	assert.Equal(t, []string{"stringList1", "stringList2"}, v.Value)

	v, err = parameter.ParseParameter([]string{"test-data/stringList.csv", "test-data/stringList2.csv"})
	require.NoError(t, err)
	assert.Equal(t, []string{"stringList1", "stringList2", "stringList3", "stringList4"}, v.Value)

	parameter = NewParameterDefinition("test", ParameterTypeStringListFromFiles,
		WithDefault("default"),
	)
	v, err = parameter.ParseParameter([]string{"test-data/string.txt"})
	require.NoError(t, err)
	assert.Equal(t, []string{"string1"}, v.Value)

	v, err = parameter.ParseParameter([]string{"test-data/string.txt", "test-data/string2.txt"})
	require.NoError(t, err)
	assert.Equal(t, []string{"string1", "string2"}, v.Value)
}

func TestParseObjectListFromFileRealFile(t *testing.T) {
	parameter := NewParameterDefinition("test", ParameterTypeObjectListFromFile,
		WithDefault([]interface{}{}),
	)

	v, err := parameter.ParseParameter([]string{"test-data/object.json"})
	require.NoError(t, err)
	assert.Equal(t, []interface{}{map[string]interface{}{"name": "object1", "type": "object"}}, v.Value)

	v, err = parameter.ParseParameter([]string{"test-data/objectList.json"})
	require.NoError(t, err)
	assert.Equal(t,
		[]interface{}{
			map[string]interface{}{"name": "objectList1", "type": "object"},
			map[string]interface{}{"name": "objectList2", "type": "object"},
		}, v.Value)

	v, err = parameter.ParseParameter([]string{"test-data/objectList3.csv"})
	require.NoError(t, err)
	assert.Equal(t,
		[]interface{}{
			map[string]interface{}{"name": "objectList5", "type": "object"},
			map[string]interface{}{"name": "objectList6", "type": "object"},
		}, v.Value)

	parameter = NewParameterDefinition("test", ParameterTypeObjectListFromFiles,
		WithDefault([]interface{}{}),
	)

	v, err = parameter.ParseParameter([]string{"test-data/object.json"})
	require.NoError(t, err)
	assert.Equal(t, []interface{}{map[string]interface{}{"name": "object1", "type": "object"}}, v.Value)

	v, err = parameter.ParseParameter([]string{"test-data/object.json", "test-data/object2.json"})
	require.NoError(t, err)
	assert.Equal(t,
		[]interface{}{
			map[string]interface{}{"name": "object1", "type": "object"},
			map[string]interface{}{"name": "object2", "type": "object"},
		},
		v.Value)

	v, err = parameter.ParseParameter([]string{
		"test-data/objectList.json",
		"test-data/objectList2.yaml",
		"test-data/object.json",
		"test-data/object2.json",
		"test-data/objectList3.csv"})
	require.NoError(t, err)
	assert.Equal(t,
		[]interface{}{
			map[string]interface{}{"name": "objectList1", "type": "object"},
			map[string]interface{}{"name": "objectList2", "type": "object"},
			map[string]interface{}{"name": "objectList3", "type": "object"},
			map[string]interface{}{"name": "objectList4", "type": "object"},
			map[string]interface{}{"name": "object1", "type": "object"},
			map[string]interface{}{"name": "object2", "type": "object"},
			map[string]interface{}{"name": "objectList5", "type": "object"},
			map[string]interface{}{"name": "objectList6", "type": "object"},
		},
		v.Value)
}

func TestParseDate(t *testing.T) {
	// set default time for unit tests
	refTime_ := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	refTime = &refTime_

	testCases := []struct {
		Value  string
		Result time.Time
	}{
		{Value: "2018-01-01", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local)},
		{Value: "2018/01/01", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local)},
		//{Value: "January First 2018", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "January 1st 2018", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local)},
		{Value: "2018-01-01T00:00:00+00:00", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "2018-01-01T00:00:00+01:00", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.FixedZone("", 3600))},
		{Value: "2018-01-01T00:00:00-01:00", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.FixedZone("", -3600))},
		{Value: "2018", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local)},
		{Value: "2018-01", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local)},
		{Value: "last year", Result: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "last hour", Result: time.Date(2017, 12, 31, 23, 0, 0, 0, time.UTC)},
		{Value: "last month", Result: time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "last week", Result: time.Date(2017, 12, 25, 0, 0, 0, 0, time.UTC)},
		{Value: "last monday", Result: time.Date(2017, 12, 25, 0, 0, 0, 0, time.UTC)},
		{Value: "10 days ago", Result: time.Date(2017, 12, 22, 0, 0, 0, 0, time.UTC)},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("ParseDate - %s", testCase.Value), func(t *testing.T) {

			result, err := ParseDate(testCase.Value)
			require.Nil(t, err)
			if !result.Equal(testCase.Result) {
				t.Errorf("Expected %s to parse to %s, got %s", testCase.Value, testCase.Result, result)
			}
		})
	}
}

type DefaultTypeTestCase struct {
	Type    ParameterType
	Value   interface{}
	Choices []string
	Args    []string
}

func TestValidDefaultValue(t *testing.T) {
	testCases := []DefaultTypeTestCase{
		{Type: ParameterTypeString, Value: "foo"},
		{Type: ParameterTypeInteger, Value: 123},
		{Type: ParameterTypeBool, Value: false},
		{Type: ParameterTypeDate, Value: "2018-01-01"},
		{Type: ParameterTypeStringList, Value: []string{"foo", "bar"}},
		{Type: ParameterTypeIntegerList, Value: []int{1, 2, 3}},
		{Type: ParameterTypeStringList, Value: []string{}},
		{Type: ParameterTypeIntegerList, Value: []int{}},
		{Type: ParameterTypeChoice, Value: "foo", Choices: []string{"foo", "bar"}},
		{Type: ParameterTypeChoiceList, Value: []string{"foo", "bar"}, Choices: []string{"foo", "bar"}},
	}
	for _, testCase := range testCases {
		t.Run(string(testCase.Type), func(t *testing.T) {
			param := &ParameterDefinition{
				Name:    "foo",
				Default: testCase.Value,
				Type:    testCase.Type,
				Choices: testCase.Choices,
			}
			err := param.CheckParameterDefaultValueValidity()
			assert.Nil(t, err)
		})
	}
}

func TestValidChoiceDefaultValue(t *testing.T) {
	param := &ParameterDefinition{
		Name:    "foo",
		Default: "bar",
		Type:    ParameterTypeChoice,
		Choices: []string{"foo", "bar"},
	}
	err := param.CheckParameterDefaultValueValidity()
	assert.Nil(t, err)
}

func TestInvalidChoiceDefaultValue(t *testing.T) {
	testCases := []interface{}{
		"baz",
		123,
		"flop",
	}
	for _, testCase := range testCases {
		param := &ParameterDefinition{
			Name:    "foo",
			Default: testCase,
			Type:    ParameterTypeChoice,
			Choices: []string{"foo", "bar"},
		}
		err := param.CheckParameterDefaultValueValidity()
		assert.Error(t, err)
	}
}
