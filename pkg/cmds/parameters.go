package cmds

import (
	"bytes"
	"encoding/json"
	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"github.com/tj/go-naturaldate"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Parameter is a declarative way of describing a command line parameter.
// A Parameter can be either a Flag or an Argument.
// Along with metadata (Name, Help) that is useful for help,
// it also specifies a Type, a Default value and if it is Required.
type Parameter struct {
	Name      string        `yaml:"name"`
	ShortFlag string        `yaml:"shortFlag,omitempty"`
	Type      ParameterType `yaml:"type"`
	Help      string        `yaml:"help,omitempty"`
	Default   interface{}   `yaml:"default,omitempty"`
	Choices   []string      `yaml:"choices,omitempty"`
	Required  bool          `yaml:"required,omitempty"`
}

func (p *Parameter) Copy() *Parameter {
	return &Parameter{
		Name:      p.Name,
		ShortFlag: p.ShortFlag,
		Type:      p.Type,
		Help:      p.Help,
		Default:   p.Default,
		Choices:   p.Choices,
		Required:  p.Required,
	}
}

func (p *Parameter) SetValue(value *reflect.Value) error {
	switch p.Type {
	case ParameterTypeString:
		if p.Default == nil {
			value.SetString("")
		} else {
			value.SetString(p.Default.(string))
		}
	case ParameterTypeBool:
		if p.Default == nil {
			value.SetBool(false)
		} else {
			value.SetBool(p.Default.(bool))
		}
	case ParameterTypeInteger:
		if p.Default == nil {
			value.SetInt(0)
		} else {
			value.SetInt(p.Default.(int64))
		}
	case ParameterTypeFloat:
		if p.Default == nil {
			value.SetFloat(0)
		} else {
			value.SetFloat(p.Default.(float64))
		}
	case ParameterTypeStringList:
		if p.Default == nil {
			value.Set(reflect.ValueOf([]string{}))
		} else {
			value.Set(reflect.ValueOf(p.Default.([]string)))
		}
	case ParameterTypeIntegerList:
		if p.Default == nil {
			value.Set(reflect.ValueOf([]int64{}))
		} else {
			value.Set(reflect.ValueOf(p.Default.([]int64)))
		}
	case ParameterTypeFloatList:
		if p.Default == nil {
			value.Set(reflect.ValueOf([]float64{}))
		} else {
			value.Set(reflect.ValueOf(p.Default.([]float64)))
		}
	case ParameterTypeChoice:
		if p.Default == nil {
			value.SetString("")
		} else {
			value.SetString(p.Default.(string))
		}
	case ParameterTypeStringFromFile:
		if p.Default == nil {
			value.SetString("")
		} else {
			value.SetString(p.Default.(string))
		}
	case ParameterTypeObjectListFromFile:
		if p.Default == nil {
			value.Set(reflect.ValueOf([]map[string]interface{}{}))
		} else {
			value.Set(reflect.ValueOf(p.Default.([]map[string]interface{})))
		}
	case ParameterTypeObjectFromFile:
		if p.Default == nil {
			value.Set(reflect.ValueOf(map[string]interface{}{}))
		} else {
			value.Set(reflect.ValueOf(p.Default.(map[string]interface{})))
		}

	default:
		return errors.Errorf("unknown parameter type %s", p.Type)
	}

	return nil
}

func InitializeStruct(s interface{}, parameterDefinitions map[string]*Parameter) error {
	// check that s is indeed a pointer to a struct
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return errors.Errorf("s is not a pointer")
	}
	if reflect.TypeOf(s).Elem().Kind() != reflect.Struct {
		return errors.Errorf("s is not a pointer to a struct")
	}
	st := reflect.TypeOf(s).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		v, ok := field.Tag.Lookup("glazed.parameter")
		if !ok {
			continue
		}
		parameter, ok := parameterDefinitions[v]
		if !ok {
			return errors.Errorf("unknown parameter %s", v)
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		err := parameter.SetValue(&value)
		if err != nil {
			return errors.Wrapf(err, "failed to set value for %s", v)
		}
	}

	return nil
}

type ParameterType string

const (
	ParameterTypeString         ParameterType = "string"
	ParameterTypeStringFromFile ParameterType = "stringFromFile"

	// TODO (2023-02-07) It would be great to have "list of objects from file" here
	// See https://github.com/go-go-golems/glazed/issues/117

	ParameterTypeObjectListFromFile ParameterType = "objectListFromFile"
	ParameterTypeObjectFromFile     ParameterType = "objectFromFile"
	ParameterTypeInteger            ParameterType = "int"
	ParameterTypeFloat              ParameterType = "float"
	ParameterTypeBool               ParameterType = "bool"
	ParameterTypeDate               ParameterType = "date"
	ParameterTypeStringList         ParameterType = "stringList"
	ParameterTypeIntegerList        ParameterType = "intList"
	ParameterTypeFloatList          ParameterType = "floatList"
	ParameterTypeChoice             ParameterType = "choice"
)

func (p *Parameter) CheckParameterDefaultValueValidity() error {
	// we can have no default
	if p.Default == nil {
		return nil
	}

	switch p.Type {
	case ParameterTypeString:
		_, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

	case ParameterTypeStringFromFile:
		_, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

	case ParameterTypeObjectFromFile:
		_, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

	case ParameterTypeInteger:
		_, ok := p.Default.(int)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer: %v", p.Name, p.Default)
		}

	case ParameterTypeFloat:
		_, ok := p.Default.(int)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer: %v", p.Name, p.Default)
		}

	case ParameterTypeBool:
		_, ok := p.Default.(bool)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a bool: %v", p.Name, p.Default)
		}

	case ParameterTypeDate:
		defaultValue, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

		_, err2 := parseDate(defaultValue)
		if err2 != nil {
			return errors.Wrapf(err2, "Default value for parameter %s is not a valid date: %v", p.Name, p.Default)
		}

	case ParameterTypeStringList:
		_, ok := p.Default.([]string)
		if !ok {
			defaultValue, ok := p.Default.([]interface{})
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a string list: %v", p.Name, p.Default)
			}

			// convert to string list
			fixedDefault, err := convertToStringList(defaultValue)
			if err != nil {
				return errors.Wrapf(err, "Could not convert default value for parameter %s to string list: %v", p.Name, p.Default)
			}
			p.Default = fixedDefault
		}

	case ParameterTypeIntegerList:
		_, ok := p.Default.([]int)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer list: %v", p.Name, p.Default)
		}

	case ParameterTypeFloatList:
		_, ok := p.Default.([]float32)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a float list: %v", p.Name, p.Default)
		}

	case ParameterTypeChoice:
		if len(p.Choices) == 0 {
			return errors.Errorf("Parameter %s is a choice parameter but has no choices", p.Name)
		}

		defaultValue, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

		found := false
		for _, choice := range p.Choices {
			if choice == defaultValue {
				found = true
			}
		}
		if !found {
			return errors.Errorf("Default value for parameter %s is not a valid choice: %v", p.Name, p.Default)
		}
	}

	return nil
}

func (p *Parameter) ParseParameter(v []string) (interface{}, error) {
	if len(v) == 0 {
		if p.Required {
			return nil, errors.Errorf("Argument %s not found", p.Name)
		} else {
			return p.Default, nil
		}
	}

	switch p.Type {
	case ParameterTypeString:
		return v[0], nil
	case ParameterTypeInteger:
		i, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
		}
		return i, nil
	case ParameterTypeFloat:
		f, err := strconv.ParseFloat(v[0], 32)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as float", p.Name)
		}
		return float32(f), nil
	case ParameterTypeStringList:
		return v, nil
	case ParameterTypeIntegerList:
		ints := make([]int, 0)
		for _, arg := range v {
			i, err := strconv.Atoi(arg)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
			}
			ints = append(ints, i)
		}
		return ints, nil

	case ParameterTypeBool:
		b, err := strconv.ParseBool(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as bool", p.Name)
		}
		return b, nil

	case ParameterTypeChoice:
		choice := v[0]
		found := false
		for _, c := range p.Choices {
			if c == choice {
				found = true
			}
		}
		if !found {
			return nil, errors.Errorf("Argument %s has invalid choice %s", p.Name, choice)
		}
		return choice, nil

	case ParameterTypeDate:
		parsedDate, err := parseDate(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as date", p.Name)
		}
		return parsedDate, nil

	case ParameterTypeObjectFromFile:
		fileName := v[0]
		f, err := os.Open(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}

		object := interface{}(nil)
		if strings.HasSuffix(fileName, ".json") {
			err = json.NewDecoder(f).Decode(&object)
		} else if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
			err = yaml.NewDecoder(f).Decode(&object)
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", fileName)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", fileName)
		}

		return object, nil

	case ParameterTypeStringFromFile:
		fileName := v[0]
		if fileName == "-" {
			var b bytes.Buffer
			_, err := io.Copy(&b, os.Stdin)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not read from stdin")
			}
			return b.String(), nil
		}

		bs, err := os.ReadFile(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}
		return string(bs), nil

	case ParameterTypeFloatList:
		floats := make([]float64, 0)
		for _, arg := range v {
			// parse to float
			f, err := strconv.ParseFloat(arg, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
			}
			floats = append(floats, f)
		}
		return floats, nil
	}

	return nil, errors.Errorf("Unknown parameter type %s", p.Type)
}

func convertToStringList(value []interface{}) ([]string, error) {
	stringList := make([]string, len(value))
	for i, v := range value {
		s, ok := v.(string)
		if !ok {
			return nil, errors.Errorf("Not a string: %v", v)
		}
		stringList[i] = s
	}
	return stringList, nil
}

func parseDate(value string) (time.Time, error) {
	parsedDate, err := dateparse.ParseAny(value)
	if err != nil {
		refTime_ := time.Now()
		if refTime != nil {
			refTime_ = *refTime
		}
		parsedDate, err = naturaldate.Parse(value, refTime_)
		if err != nil {
			return time.Time{}, errors.Wrapf(err, "Could not parse date: %s", value)
		}
	}

	return parsedDate, nil
}
