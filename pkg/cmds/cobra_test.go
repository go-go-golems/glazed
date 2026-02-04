package cmds

import (
	"embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
	"gopkg.in/yaml.v3"
)

func TestAddZeroArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	definitions := fields.NewDefinitions()
	err := definitions.AddFieldsToCobraCommand(cmd, "")
	// assert that err is nil
	require.Nil(t, err)
}

func TestAddSingleRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}

	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:     "foo",
				Required: true,
				Type:     fields.TypeString,
			},
		))
	require.Nil(t, err)
	desc := CommandDescription{
		Schema: schema.NewSchema(schema.WithSections(defaultSection)),
	}
	err = defaultSection.AddSectionToCobraCommand(cmd)

	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo"}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"bar"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1.Value)

	_, err = defaultArguments.GatherArguments([]string{}, false, false)
	assert.Error(t, err)

	_, err = defaultArguments.GatherArguments([]string{"foo", "bla"}, false, false)
	assert.Error(t, err)
}

func TestAddTwoRequiredArguments(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:     "foo",
				Required: true,
				Type:     fields.TypeString,
			},
			&fields.Definition{
				Name:     "bar",
				Required: true,
				Type:     fields.TypeString,
			},
		),
	)
	require.Nil(t, err)
	desc := NewCommandDescription("test", WithSections(defaultSection))
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar"}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"bar", "foo"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 2, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1.Value)
	v2, ok := values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "foo", v2.Value)

	_, err = defaultArguments.GatherArguments([]string{}, false, false)
	assert.Error(t, err)

	_, err = defaultArguments.GatherArguments([]string{"bar"}, false, false)
	assert.Error(t, err)

	_, err = defaultArguments.GatherArguments([]string{"bar", "foo", "baz"}, false, false)
	assert.Error(t, err)
}

func TestOneRequiredOneOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:     "foo",
				Required: true,
				Type:     fields.TypeString,
			},
			&fields.Definition{
				Name:    "bar",
				Type:    fields.TypeString,
				Default: cast.InterfaceAddr("baz"),
			},
		),
	)
	require.Nil(t, err)
	desc := CommandDescription{
		Schema: schema.NewSchema(schema.WithSections(defaultSection)),
	}
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"bar", "foo"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 2, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1.Value)
	v2, ok := values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "foo", v2.Value)

	values, err = defaultArguments.GatherArguments([]string{"foo"}, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "foo", v1.Value)
	v2, ok = values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "baz", v2.Value)

	_, err = defaultArguments.GatherArguments([]string{}, false, false)
	assert.Error(t, err)

	_, err = defaultArguments.GatherArguments([]string{"bar", "foo", "baz"}, false, false)
	assert.Error(t, err)
}

func TestOneOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:    "foo",
				Default: cast.InterfaceAddr("123"),
				Type:    fields.TypeString,
			},
		),
	)
	require.Nil(t, err)
	desc := NewCommandDescription("test", WithSections(defaultSection))
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"foo"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "foo", v1.Value)

	values, err = defaultArguments.GatherArguments([]string{}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "123", v1.Value)
}

func TestDefaultIntValue(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:    "foo",
				Default: cast.InterfaceAddr(123),
				Type:    fields.TypeInteger,
			},
		),
	)
	require.Nil(t, err)
	desc := CommandDescription{
		Schema: schema.NewSchema(schema.WithSections(defaultSection)),
	}
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, 123, v1.Value)

	values, err = defaultArguments.GatherArguments([]string{"234"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, 234, v1.Value)

	_, err = defaultArguments.GatherArguments([]string{"foo"}, false, false)
	assert.Error(t, err)
}

type DefaultTypeTestCase struct {
	Type  fields.Type
	Value interface{}
	Args  []string
}

func TestInvalidDefaultValue(t *testing.T) {
	cmd := &cobra.Command{}
	failingTypes := []DefaultTypeTestCase{
		{Type: fields.TypeString, Value: 123},
		{Type: fields.TypeString, Value: []string{"foo"}},
		{Type: fields.TypeInteger, Value: "foo"},
		{Type: fields.TypeInteger, Value: []int{1}},
		// so oddly enough this is a valid word
		{Type: fields.TypeDate, Value: "22#@!"},
		{Type: fields.TypeStringList, Value: "foo"},
		{Type: fields.TypeIntegerList, Value: "foo"},
		{Type: fields.TypeStringList, Value: []int{1, 2, 3}},
		{Type: fields.TypeStringList, Value: []int{}},
		{Type: fields.TypeIntegerList, Value: []string{"1", "2", "3"}},
		{Type: fields.TypeIntegerList, Value: []string{}},
	}
	for _, failingType := range failingTypes {
		defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
			schema.WithArguments(
				&fields.Definition{
					Name:    "foo",
					Default: cast.InterfaceAddr(failingType.Value),
					Type:    failingType.Type,
				},
			),
		)
		require.Nil(t, err)
		err = defaultSection.AddSectionToCobraCommand(cmd)
		if err == nil {
			t.Errorf("Expected error for type %s and value %v\n", failingType.Type, failingType.Value)
		}
		assert.Error(t, err)
	}
}

func TestTwoOptionalArguments(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name: "foo",
			},
			&fields.Definition{
				Name: "bar",
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "blop"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestFailAddingRequiredAfterOptional(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name: "foo",
			},
			&fields.Definition{
				Name:     "bar",
				Required: true,
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	assert.Error(t, err)
}

func TestAddStringListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:     "foo",
				Required: true,
				Type:     fields.TypeStringList,
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))
}

func TestAddStringListOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:    "foo",
				Type:    fields.TypeStringList,
				Default: cast.InterfaceAddr([]string{"baz"}),
			},
		),
	)
	require.Nil(t, err)
	desc := NewCommandDescription("test", WithSections(defaultSection))
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"bar", "foo"}, false, false)
	require.Nil(t, err)
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"bar", "foo"}, v1.Value)

	values, err = defaultArguments.GatherArguments([]string{"foo"}, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"foo"}, v1.Value)

	values, err = defaultArguments.GatherArguments([]string{}, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"baz"}, v1.Value)
}

func TestFailAddingArgumentAfterStringList(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name: "foo",
				Type: fields.TypeStringList,
			},
			&fields.Definition{
				Name: "bar",
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	assert.Error(t, err)
}

func TestAddIntegerListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:     "foo",
				Required: true,
				Type:     fields.TypeIntegerList,
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"1", "2"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListRequiredAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:     "foo",
				Required: true,
			},
			&fields.Definition{
				Name:     "bar",
				Type:     fields.TypeStringList,
				Required: true,
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListOptionalAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:     "foo",
				Required: true,
			},
			&fields.Definition{
				Name:    "bar",
				Type:    fields.TypeStringList,
				Default: cast.InterfaceAddr([]string{"blop"}),
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListOptionalAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name:    "foo",
				Type:    fields.TypeString,
				Default: cast.InterfaceAddr("blop"),
			},
			&fields.Definition{
				Name:    "bar",
				Type:    fields.TypeStringList,
				Default: cast.InterfaceAddr([]string{"bloppp"}),
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListRequiredAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithArguments(
			&fields.Definition{
				Name: "foo",
			},
			&fields.Definition{
				Name:     "bar",
				Type:     fields.TypeStringList,
				Required: true,
			},
		),
	)
	require.Nil(t, err)
	err = defaultSection.AddSectionToCobraCommand(cmd)
	assert.Error(t, err)
}

// expectedCommandResults is a struct that contains the expected results of a command,
// which is a list of parsed flag fields, a list of parsed argument fields,
// as well as potential errors.
//
// This is used to quickly check the result of passing a set of arguments to a command.
type expectedCommandResults struct {
	Name                   string                 `yaml:"name"`
	ExpectedArgumentFields map[string]interface{} `yaml:"argumentFields"`
	ExpectedFlagFields     map[string]interface{} `yaml:"flagFields"`
	ExpectedFlagError      bool                   `yaml:"flagError"`
	ExpectedArgumentError  bool                   `yaml:"argumentError"`
	Args                   []string               `yaml:"args"`
}

type commandDescription struct {
	CommandDescription
	Flags     []*fields.Definition `yaml:"flags"`
	Arguments []*fields.Definition `yaml:"arguments"`
}

type commandTest struct {
	Description *commandDescription       `yaml:"description"`
	Tests       []*expectedCommandResults `yaml:"tests"`
}

//go:embed "test-data/cobra/*.yaml"
var cobraData embed.FS

func TestCommandArgumentsParsing(t *testing.T) {
	// enumerate all the test files in cobraData
	files, err := cobraData.ReadDir("test-data/cobra")
	require.NoError(t, err)

	for _, file := range files {
		// load yaml from file
		testSuite := &commandTest{}
		fileData, err := cobraData.ReadFile("test-data/cobra/" + file.Name())
		require.NoError(t, err)

		err = yaml.Unmarshal(fileData, testSuite)
		require.NoError(t, err)

		section, err := schema.NewSection(schema.DefaultSlug, "Default",
			schema.WithArguments(testSuite.Description.Arguments...),
			schema.WithFields(testSuite.Description.Flags...),
		)
		require.NoError(t, err)
		testSuite.Description.Schema = schema.NewSchema(schema.WithSections(section))

		if testSuite.Description.Name != "string-from-file" {
			// XXX hack to debug
			continue
		}

		for _, test := range testSuite.Tests {
			test2 := test
			if test2.ExpectedArgumentFields == nil {
				test2.ExpectedArgumentFields = map[string]interface{}{}
			}
			if test2.ExpectedFlagFields == nil {
				test2.ExpectedFlagFields = map[string]interface{}{}
			}

			t.Run(
				fmt.Sprintf("%s/%s", testSuite.Description.Name, test2.Name),
				func(t *testing.T) {
					testCommandParseHelper(t, &testSuite.Description.CommandDescription, test2)
				})
		}
	}
}

func testCommandParseHelper(
	t *testing.T,
	desc *CommandDescription,
	expected *expectedCommandResults,
) {
	var flagsError error
	var argsError error
	var flagFields *fields.FieldValues
	var argumentFields *fields.FieldValues

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			flagFields, flagsError = desc.GetDefaultFlags().GatherFlagsFromCobraCommand(cmd, false, false, "")
			if flagsError != nil {
				return
			}
			argumentFields, argsError = desc.GetDefaultArguments().GatherArguments(args, false, false)
			if argsError != nil {
				return
			}
		},
	}

	defaultSection, ok := desc.GetDefaultSection()
	require.True(t, ok)
	defaultSection_, ok := defaultSection.(schema.CobraSection)
	require.True(t, ok)

	err := defaultSection_.AddSectionToCobraCommand(cmd)
	require.Nil(t, err)

	cmd.SetArgs(expected.Args)

	_ = capturer.CaptureStderr(func() {
		err = cmd.Execute()
	})

	if expected.ExpectedFlagError || expected.ExpectedArgumentError {
		assert.Errorf(t, err, "expected error for %v", expected.Args)
	} else {
		assert.NoErrorf(t, err, "unexpected error for %v", expected.Args)
	}

	if err != nil {
		return
	}

	if expected.ExpectedFlagError {
		assert.Errorf(t, flagsError, "expected flag error for %v", expected.Args)
		return
	} else {
		assert.NoErrorf(t, flagsError, "Unexpected error parsing flags: %v", expected.Args)
	}
	if expected.ExpectedArgumentError {
		assert.Errorf(t, argsError, "expected error for %v", expected.Args)
		return
	} else {
		assert.NoErrorf(t, argsError, "expected no error for %v", expected.Args)
	}

	assertJsonEquivalent(t, expected.ExpectedArgumentFields, argumentFields)
	assertJsonEquivalent(t, expected.ExpectedFlagFields, flagFields)
}

func assertJsonEquivalent(t *testing.T, expected interface{}, actual interface{}) {
	expectedBytes, err := json.MarshalIndent(expected, "", "  ")
	require.NoError(t, err)
	actualBytes, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBytes), string(actualBytes))
}
