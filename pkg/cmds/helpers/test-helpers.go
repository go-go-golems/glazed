package helpers

import (
	"fmt"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/errgo.v2/fmt/errors"
)

// Package helpers provides structures and helper functions required for
// conducting table-driven tests in Golang using external YAML files.
//
// The YAML file typically includes test cases with input, expected output,
// and a descriptive name.
// All these will be loaded into structs that use the Test* structs defined here.
//
// This approach makes your tests more clear by extracting verbose the definitions and values.
//
// Here are the steps to use this package:
//
// 1. Define your test cases inside a YAML file.
//    The YAML can use the Test* structs defined here for specific fields.
// 2. Load the YAML file using the LoadTestFromYAML function into your table-driven test struct.
//
// 3. Use the New* functions to convert the data loaded from YAML into appropriate form for the glazed library (typically Values and schema sections)
//
// 4. Write a standard table-driven test loop.
//
// ---
//
// test_cases.yaml
// ===============
// // fill with your test cases

// - name: "Empty sections and values"
//  sections: []
//  values: []
//  expectedSections: []
//  expectedError: false
//
//- name: "Single section with default"
//  sections:
//    - name: "section1"
//      definitions:
//        - name: "param1"
//          type: "string"
//          default: "defaultVal"
//  values:
//    - name: "section1"
//  expectedSections:
//    - name: "section1"
//      values:
//        param1: "defaultVal"
//  ... # potentially additional fields
//
//- name: "Multiple sections with defaults"
//  sections:
//    - name: "section1"
//      definitions:
//        - name: "param1"
//          type: "string"
//          default: "default1"
//        - name: "param2"
//          type: "integer"
//          default: 42
//    - name: "section2"
//      definitions:
//        - name: "param3"
//          type: "bool"
//          default: true
//  values:
//    - name: "section1"
//		fields:
//		  - name: "param1"
//          value: "existingValue"
//    - name: "section2"
//  expectedSections:
//    - name: "section1"
//      values:
//        param1: "default1"
//        param2: 42
//    - name: "section2"
//      values:
//        param3: true
//  expectedError: false
//  ... # potentially additional fields
//
// ---
//
// type test struct {
// 	name            string
// 	sections        []helpers.TestSection
// 	values          []helpers.TestSectionValues
// 	expectedSections []helpers.TestExpectedSection
// 	expectedError   bool
// }
//
// func TestYourFunc(t *testing.T) {
//  tests := LoadTestFromYAML(test_cases.yaml)
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			schema_ := helpers.NewTestSchema(tt.sections)
// 			parsedValues := helpers.NewTestValues(schema_, tt.values)
//
//          YourTest()
//
// 			if err != nil {
// 				t.Errorf("YourTest() error = %v", err)
// 				return
// 			}
//
// 			for _, l_ := range tt.expectedSections {
// 				l, ok := parsedValues.Get(l_.Name)
// 				require.True(t, ok)
//
// 				actual := l.Fields.ToMap()
// 				assert.Equal(t, l_.Values, actual)
// 			}
// 		})
// 	}
// }

type TestSection struct {
	Name        string               `yaml:"name"`
	Definitions []*fields.Definition `yaml:"definitions,omitempty"`
	Prefix      string               `yaml:"prefix"`
}

type TestParsedField struct {
	Name  string      `yaml:"name"`
	Value interface{} `yaml:"value"`
}

type TestSectionValues struct {
	Name   string            `yaml:"name"`
	Fields []TestParsedField `yaml:"fields"`
}

type TestExpectedSection struct {
	Name   string                        `yaml:"name"`
	Values map[string]interface{}        `yaml:"values"`
	Logs   map[string][]fields.ParseStep `yaml:"logs"`
}

type TestMiddlewareName string

const TestMiddlewareSetFromDefaults = "setFromDefaults"
const TestMiddlewareUpdateFromMap = "updateFromMap"
const TestMiddlewareUpdateFromMapAsDefault = "updateFromMapAsDefault"
const TestMiddlewareUpdateFromEnv = "updateFromEnv"
const TestWhitelistSections = "whitelistSections"
const TestWhitelistSectionsFirst = "whitelistSectionsFirst"
const TestWhitelistSectionFields = "whitelistSectionFields"
const TestWhitelistSectionFieldsFirst = "whitelistSectionFieldsFirst"
const TestBlacklistSections = "blacklistSections"
const TestBlacklistSectionsFirst = "blacklistSectionsFirst"
const TestBlacklistSectionFields = "blacklistSectionFields"
const TestBlacklistSectionFieldsFirst = "blacklistSectionFieldsFirst"

type TestParseOptionName string

const TestParseOptionSource = "source"
const TestParseOptionValue = "value"
const TestParseOptionMetadata = "metadata"

type TestParseOption struct {
	Name  TestParseOptionName `yaml:"name"`
	Value interface{}         `yaml:"value"`
}

type TestMiddleware struct {
	Name     TestMiddlewareName                 `yaml:"name"`
	Options  []TestParseOption                  `yaml:"options"`
	Map      *map[string]map[string]interface{} `yaml:"map"`
	Prefix   *string                            `yaml:"prefix"`
	Sections *[]string                          `yaml:"sections"`
	Fields   *map[string][]string               `yaml:"fields"`
}

type TestMiddlewares []TestMiddleware

func (t TestMiddlewares) ToMiddlewares() ([]sources.Middleware, error) {
	ret := []sources.Middleware{}
	for _, m := range t {
		options := []fields.ParseOption{}
		for _, o := range m.Options {
			switch o.Name {
			case TestParseOptionSource:
				options = append(options, fields.WithSource(o.Value.(string)))
			case TestParseOptionValue:
				options = append(options, fields.WithParseStepValue(o.Value))
			case TestParseOptionMetadata:
				options = append(options, fields.WithMetadata(o.Value.(map[string]interface{})))
			default:
				return nil, errors.Newf("unknown option name %s", o.Name)
			}
		}

		switch m.Name {
		case TestMiddlewareSetFromDefaults:
			ret = append(ret, sources.FromDefaults(options...))
		case TestMiddlewareUpdateFromMap:
			ret = append(ret, sources.FromMap(*m.Map, options...))
		case TestMiddlewareUpdateFromMapAsDefault:
			ret = append(ret, sources.FromMapAsDefault(*m.Map, options...))
		case TestMiddlewareUpdateFromEnv:
			ret = append(ret, sources.FromEnv(*m.Prefix, options...))
		case TestWhitelistSections:
			ret = append(ret, sources.WhitelistSections(*m.Sections))
		case TestWhitelistSectionsFirst:
			ret = append(ret, sources.WhitelistSectionsFirst(*m.Sections))
		case TestWhitelistSectionFields:
			ret = append(ret, sources.WhitelistSectionFields(*m.Fields))
		case TestWhitelistSectionFieldsFirst:
			ret = append(ret, sources.WhitelistSectionFieldsFirst(*m.Fields))
		case TestBlacklistSections:
			ret = append(ret, sources.BlacklistSections(*m.Sections))
		case TestBlacklistSectionsFirst:
			ret = append(ret, sources.BlacklistSectionsFirst(*m.Sections))
		case TestBlacklistSectionFields:
			ret = append(ret, sources.BlacklistSectionFields(*m.Fields))
		case TestBlacklistSectionFieldsFirst:
			ret = append(ret, sources.BlacklistSectionFieldsFirst(*m.Fields))
		default:
			return nil, errors.Newf("unknown middleware name %s", m.Name)
		}
	}

	return ret, nil
}

// NewTestSection is a helper function to create a Section from a definition bundle.
func NewTestSection(l TestSection) schema.Section {
	definitions_ := []*fields.Definition{}
	definitions_ = append(definitions_, l.Definitions...)
	ret, err := schema.NewSection(l.Name, l.Name,
		schema.WithFields(definitions_...))
	if err != nil {
		panic(err)
	}

	return ret
}

func NewTestSchema(ls []TestSection) *schema.Schema {
	ret := schema.NewSchema()
	for _, l := range ls {
		ret.Set(l.Name, NewTestSection(l))
	}
	return ret
}

// NewTestSectionValues helper function to create Values from TestParsedField.
func NewTestSectionValues(pl schema.Section, l TestSectionValues) *values.SectionValues {
	params_ := fields.NewFieldValues()
	pds := pl.GetDefinitions()
	for _, p := range l.Fields {
		pd, ok := pds.Get(p.Name)
		if !ok {
			panic("field definition not found")
		}
		err := params_.UpdateValue(p.Name, pd, p.Value)
		if err != nil {
			panic(err)
		}
	}

	ret, err := values.NewSectionValues(pl, values.WithFields(params_))
	if err != nil {
		panic(err)
	}

	return ret
}

func NewTestValues(pls *schema.Schema, ls ...TestSectionValues) *values.Values {
	ret := values.New()
	for _, l := range ls {
		pl, ok := pls.Get(l.Name)
		if !ok {
			panic("section not found")
		}
		ret.Set(l.Name, NewTestSectionValues(pl, l))
	}
	return ret
}

// compareValues handles comparison of values with special cases for slice types
func compareValues(t *testing.T, expected, actual interface{}, key string) {
	normalizedExpected, err := cast.NormalizeValue(expected)
	require.NoError(t, err, "failed to normalize expected value for key %s", key)

	normalizedActual, err := cast.NormalizeValue(actual)
	require.NoError(t, err, "failed to normalize actual value for key %s", key)

	assert.Equal(t, normalizedExpected, normalizedActual, "mismatch for key %s", key)
}

func TestExpectedOutputs(t *testing.T, expectedSections []TestExpectedSection, parsedValues *values.Values) {
	expectedSections_ := map[string]TestExpectedSection{}
	for _, section := range expectedSections {
		expectedSections_[section.Name] = section
		l, ok := parsedValues.Get(section.Name)
		require.True(t, ok)

		actual, err := l.Fields.ToInterfaceMap()
		require.NoError(t, err)

		// Compare each value using the helper
		for k, expectedValue := range section.Values {
			compareValues(t, expectedValue, actual[k], k)
		}

		for k, expectedLog := range section.Logs {
			actual, ok := l.Fields.Get(k)
			require.True(t, ok)

			// Compare each log entry using the helper
			for i, expectedEntry := range expectedLog {
				if i < len(actual.Log) {
					compareValues(t, expectedEntry.Value, actual.Log[i].Value, fmt.Sprintf("%s.Log[%d].Value", k, i))
					assert.Equal(t, expectedEntry.Source, actual.Log[i].Source, fmt.Sprintf("%s.Log[%d].Source", k, i))
				}
			}
		}
	}

	parsedValues.ForEach(func(key string, l *values.SectionValues) {
		if _, ok := expectedSections_[key]; !ok {
			t.Errorf("did not expect section %s to be present", key)
		}
	})
}
