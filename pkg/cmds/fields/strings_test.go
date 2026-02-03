package fields

import (
	"reflect"
	"testing"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
)

// TestGatherFlagsFromStringList_ValidArgumentsAndParameters tests the function with valid arguments and fields.
func TestGatherFlagsFromStringList_ValidArgumentsAndParameters(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		params         *Definitions
		want           map[string]interface{}
		wantArgs       []string
		wantErr        bool
		ignoreRequired bool
		onlyProvided   bool
		prefix         string
	}{
		// "--verbose -o file.txt" (bool, string)
		{
			name: "bool, string",
			args: []string{"--verbose", "-o", "file.txt"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "verbose", ShortFlag: "v", Type: TypeBool},
				{Name: "output", ShortFlag: "o", Type: TypeString},
			})),
			want: map[string]interface{}{
				"verbose": true,
				"output":  "file.txt",
			},
			wantErr: false,
		},
		// "--debug --log-level info" (bool, string)
		{
			name: "bool, string",
			args: []string{"--debug", "--log-level", "info"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "debug", ShortFlag: "", Type: TypeBool},
				{Name: "log-level", ShortFlag: "l", Type: TypeString},
			})),
			want: map[string]interface{}{
				"debug":     true,
				"log-level": "info",
			},
			wantErr: false,
		},
		// "-d -l info" (bool, string, using short flags)
		{
			name: "bool, string short flags",
			args: []string{"-d", "-l", "info"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "debug", ShortFlag: "d", Type: TypeBool},
				{Name: "log-level", ShortFlag: "l", Type: TypeString},
			})),
			want: map[string]interface{}{
				"debug":     true,
				"log-level": "info",
			},
			wantErr: false,
		},
		// "--output=file.txt --verbose=true" (string, bool, using =)
		{
			name: "string, bool using =",
			args: []string{"--output=file.txt", "--verbose=true"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "output", ShortFlag: "o", Type: TypeString},
				{Name: "verbose", ShortFlag: "v", Type: TypeBool},
			})),
			want: map[string]interface{}{
				"output":  "file.txt",
				"verbose": true,
			},
			wantErr: false,
		},
		// "--verbose=false" (bool false)
		{
			name: "bool false",
			args: []string{"--verbose=false"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "verbose", ShortFlag: "v", Type: TypeBool},
			})),
			want: map[string]interface{}{
				"verbose": false,
			},
		},
		// "--output file.txt -v" (string, bool, different order)
		{
			name: "string, bool",
			args: []string{"--output", "file.txt", "-v"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "output", ShortFlag: "o", Type: TypeString},
				{Name: "verbose", ShortFlag: "v", Type: TypeBool},
			})),
			want: map[string]interface{}{
				"output":  "file.txt",
				"verbose": true,
			},
			wantErr: false,
		},
		// "--output file.txt --verbose" (string, bool, different order)
		{
			name: "string, bool another order",
			args: []string{"--output", "file.txt", "--verbose"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "output", ShortFlag: "o", Type: TypeString},
				{Name: "verbose", ShortFlag: "v", Type: TypeBool},
			})),
			want: map[string]interface{}{
				"output":  "file.txt",
				"verbose": true,
			},
			wantErr: false,
		},
		// "--size 100 --color red" (integer, string)
		{
			name: "integer, string",
			args: []string{"--size", "100", "--color", "red"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "size", ShortFlag: "s", Type: TypeInteger},
				{Name: "color", ShortFlag: "c", Type: TypeString},
			})),
			want: map[string]interface{}{
				"size":  100,
				"color": "red",
			},
			wantErr: false,
		},
		// "--size 100 -c red" (integer, string, using short flag)
		{
			name: "integer, string short flag",
			args: []string{"--size", "100", "-c", "red"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "size", ShortFlag: "s", Type: TypeInteger},
				{Name: "color", ShortFlag: "c", Type: TypeString},
			})),
			want: map[string]interface{}{
				"size":  100,
				"color": "red",
			},
			wantErr: false,
		},
		// "--enable-feature" (bool, no value)
		{
			name: "bool no value",
			args: []string{"--enable-feature"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "enable-feature", ShortFlag: "e", Type: TypeBool},
			})),
			want: map[string]interface{}{
				"enable-feature": true,
			},
			wantErr: false,
		},
		// "-e" (bool, no value, using short flag)
		{
			name: "bool short flag no value",
			args: []string{"-e"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "enable-feature", ShortFlag: "e", Type: TypeBool},
			})),
			want: map[string]interface{}{
				"enable-feature": true,
			},
			wantErr: false,
		},
		// "--float 3.14" (float)
		{
			name: "float",
			args: []string{"--float", "3.14"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "float", ShortFlag: "f", Type: TypeFloat},
			})),
			want: map[string]interface{}{
				"float": 3.14,
			},
			wantErr: false,
		},
		// "--int-list=1,2 --int-list 3,4 --int-list=5
		{
			name: "int list",
			args: []string{"--int-list=1,2", "--int-list", "3,4", "--int-list=5"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "int-list", ShortFlag: "i", Type: TypeIntegerList},
			})),
			want: map[string]interface{}{
				"int-list": []int{1, 2, 3, 4, 5},
			},
		},
		// "--choice A" (choice)
		{
			name: "choice",
			args: []string{"--choice", "A"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "choice", ShortFlag: "c", Type: TypeChoice, Choices: []string{"A", "B", "C"}},
			})),
			want: map[string]interface{}{
				"choice": "A",
			},
			wantErr: false,
		},
		// "--string-list item1,item2,item3" (string list)
		{
			name: "string list",
			args: []string{"--string-list", "item1,item2,item3"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "string-list", ShortFlag: "s", Type: TypeStringList},
			})),
			want: map[string]interface{}{
				"string-list": []string{"item1", "item2", "item3"},
			},
			wantErr: false,
		},
		// "--string-list item1 --string-list item2 --string-list item3" (string list)
		{
			name: "multiple string list",
			args: []string{"--string-list", "item1", "--string-list", "item2", "--string-list", "item3"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "string-list", ShortFlag: "s", Type: TypeStringList},
			})),
			want: map[string]interface{}{
				"string-list": []string{"item1", "item2", "item3"},
			},
			wantErr: false,
		},
		// "--string-list item1 --integer 1 --string-list item2 --string-list item3" (string list, integer)
		{
			name: "string list with integer",
			args: []string{"--string-list", "item1", "--integer", "1", "--string-list", "item2", "--string-list", "item3"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "string-list", ShortFlag: "s", Type: TypeStringList},
				{Name: "integer", ShortFlag: "i", Type: TypeInteger},
			})),
			want: map[string]interface{}{
				"string-list": []string{"item1", "item2", "item3"},
				"integer":     1,
			},
			wantErr: false,
		},
		// "--integer-list 1,2,3" (integer list)
		{
			name: "integer list",
			args: []string{"--integer-list", "1,2,3"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "integer-list", ShortFlag: "il", Type: TypeIntegerList},
			})),
			want: map[string]interface{}{
				"integer-list": []int{1, 2, 3},
			},
			wantErr: false,
		},
		// "--float-list 1.1,2.2,3.3" (float list)
		{
			name: "float list",
			args: []string{"--float-list", "1.1,2.2,3.3"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "float-list", ShortFlag: "fl", Type: TypeFloatList},
			})),
			want: map[string]interface{}{
				"float-list": []float64{1.1, 2.2, 3.3},
			},
			wantErr: false,
		},
		// "--float-list 1.1 --float-list 2 --float-list 3.3"
		{
			name: "multiple float list",
			args: []string{"--float-list", "1.1", "--float-list", "2", "--float-list", "3.3"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "float-list", ShortFlag: "fl", Type: TypeFloatList},
			})),
			want: map[string]interface{}{
				"float-list": []float64{1.1, 2, 3.3},
			},
			wantErr: false,
		},
		// "--choice-list A,B,C" (choice list)
		{
			name: "choice list",
			args: []string{"--choice-list", "A,B,C"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "choice-list", ShortFlag: "cl", Type: TypeChoiceList, Choices: []string{"A", "B", "C"}},
			})),
			want: map[string]interface{}{
				"choice-list": []string{"A", "B", "C"},
			},
			wantErr: false,
		},
		// "--key-value key1=value1;key2=value2" (key-value)
		{
			name: "key-value",
			args: []string{"--key-value", "key1:value1,key2:value2"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "key-value", ShortFlag: "kv", Type: TypeKeyValue},
			})),
			want: map[string]interface{}{
				"key-value": map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			wantErr: false,
		},
		// "--unknownFlag value" (unknown flag)
		{
			name: "UnknownFlag",
			args: []string{"--unknownFlag", "value"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "knownFlag", ShortFlag: "k", Type: TypeString},
			})),
			wantErr: true,
		},
		// "--flag" (missing value for non-boolean flag)
		{
			name: "MissingValueForNonBooleanFlag",
			args: []string{"--flag"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "flag", ShortFlag: "f", Type: TypeString},
			})),
			wantErr: true,
		},
		// "--size invalidValue" (invalid value for flag)
		{
			name: "InvalidValueForFlag",
			args: []string{"--size", "invalidValue"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "size", ShortFlag: "s", Type: TypeInteger},
			})),
			wantErr: true,
		},
		// "--flag value1 --flag value2" (repeated flags)
		{
			name: "RepeatedFlags",
			args: []string{"--flag", "value1", "--flag", "value2"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "flag", ShortFlag: "f", Type: TypeString},
			})),
			wantErr: true,
		},
		// "--flag value1 -f value2" (repeated flags)
		{
			name: "BooleanFlags",
			args: []string{"--flag", "--anotherFlag"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "flag", ShortFlag: "f", Type: TypeBool},
				{Name: "anotherFlag", ShortFlag: "a", Type: TypeBool},
			})),
			want: map[string]interface{}{
				"flag":        true,
				"anotherFlag": true,
			},
			wantErr: false,
		},
		// "" (empty arguments)
		{
			name: "EmptyArguments",
			args: []string{},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "flag", ShortFlag: "f", Type: TypeString},
			})),
			want:    map[string]interface{}{},
			wantErr: false,
		},
		// "--flag value" (empty parameters)
		{
			name:    "EmptyParameters",
			args:    []string{"--flag", "value"},
			params:  NewDefinitions(),
			wantErr: true,
		},
		// "--flag value" (parameters with empty ShortFlag)
		{
			name: "ParametersWithEmptyShortFlag",
			args: []string{"--flag", "value"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "flag", ShortFlag: "", Type: TypeString},
			})),
			want: map[string]interface{}{
				"flag": "value",
			},
			wantErr: false,
		},
		// "--flag value -f value" (parameters with the same Name and ShortFlag)
		// "--flag1 value1 -f value2" (parameters with different Name but the same ShortFlag)
		{
			name: "ParametersWithDifferentNameSameShortFlag",
			args: []string{"-f", "value1", "-f", "value2"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "flag1", ShortFlag: "f", Type: TypeString},
				{Name: "flag2", ShortFlag: "f", Type: TypeString},
			})),
			wantErr: true,
		},
		// "--flag1 value1 --flag2 value2" (mix of valid and invalid parameters)
		{
			name: "MixOfValidAndInvalidParameters",
			args: []string{"--flag1", "value1", "--flag2", "value2"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "flag1", ShortFlag: "f", Type: TypeString},
				// Assuming invalid parameter doesn't have a type
				{Name: "flag2", ShortFlag: "g"},
			})),
			wantErr: true,
		},
		// "--prefix-integer 1 --prefix-foobar 2" (prefix: integer, integer)
		{
			name: "Prefix: integer, integer",
			args: []string{"--prefix-integer", "1", "--prefix-foobar", "2"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "integer", ShortFlag: "i", Type: TypeInteger},
				{Name: "foobar", ShortFlag: "f", Type: TypeInteger},
			})),
			want: map[string]interface{}{
				"integer": 1,
				"foobar":  2,
			},
			wantErr: false,
			prefix:  "prefix-",
		},
		// "--foobar-bla foo" (param name with _: foobar_bla)
		{
			name: "Param name with _: foobar_bla",
			args: []string{"--foobar-bla", "foo"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "foobar_bla", ShortFlag: "f", Type: TypeString},
			})),
			want: map[string]interface{}{
				"foobar_bla": "foo",
			},
			wantErr: false,
		},
		// "required argument: --required foo" (required argument)
		{
			name: "Required argument",
			args: []string{"--required", "foo"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "required", ShortFlag: "r", Type: TypeString, Required: true},
			})),
			want: map[string]interface{}{
				"required": "foo",
			},
			wantErr: false,
		},
		// error when required argument is missing
		{
			name: "Required argument missing",
			args: []string{},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "required", ShortFlag: "r", Type: TypeString, Required: true},
			})),
			wantErr: true,
		},
		// "foobar: default 2, but missing on command line" (default value)
		{
			name: "Default value",
			args: []string{},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{
					Name:      "foobar",
					ShortFlag: "f",
					Type:      TypeInteger,
					Default:   cast.InterfaceAddr(2),
				},
			})),
			want: map[string]interface{}{
				"foobar": 2,
			},
			wantErr: false,
		},
		// gather arguments: --verbose foobar --output file.txt another argument
		{
			name: "Gather arguments",
			args: []string{"--verbose", "foobar", "--output", "file.txt", "another", "argument"},
			params: NewDefinitions(WithDefinitionList([]*Definition{
				{Name: "verbose", ShortFlag: "v", Type: TypeBool},
				{Name: "output", ShortFlag: "o", Type: TypeString},
			})),
			want: map[string]interface{}{
				"verbose": true,
				"output":  "file.txt",
			},
			wantArgs: []string{"foobar", "another", "argument"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got_, args, err := tt.params.GatherFlagsFromStringList(tt.args, tt.onlyProvided, tt.ignoreRequired, tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("GatherFlagsFromStringList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			got := map[string]interface{}{}
			got_.ForEach(func(key string, p *FieldValue) {
				got[key] = p.Value
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GatherFlagsFromStringList() = %v, want %v", got, tt.want)
			}
			if tt.wantArgs != nil {
				if !reflect.DeepEqual(args, tt.wantArgs) {
					t.Errorf("GatherFlagsFromStringList() = %v, want %v", args, tt.wantArgs)
				}
			}
		})
	}
}
