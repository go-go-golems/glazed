package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/go-go-golems/glazed/pkg/helpers"
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

// ParameterDefinition is a declarative way of describing a command line parameter.
// A ParameterDefinition can be either a Flag or an Argument.
// Along with metadata (Name, Help) that is useful for help,
// it also specifies a Type, a Default value and if it is Required.
type ParameterDefinition struct {
	Name      string        `yaml:"name"`
	ShortFlag string        `yaml:"shortFlag,omitempty"`
	Type      ParameterType `yaml:"type"`
	Help      string        `yaml:"help,omitempty"`
	Default   interface{}   `yaml:"default,omitempty"`
	Choices   []string      `yaml:"choices,omitempty"`
	Required  bool          `yaml:"required,omitempty"`
}

func (p *ParameterDefinition) String() string {
	return fmt.Sprintf("{Parameter: %s - %s}", p.Name, p.Type)
}

func (p *ParameterDefinition) Copy() *ParameterDefinition {
	return &ParameterDefinition{
		Name:      p.Name,
		ShortFlag: p.ShortFlag,
		Type:      p.Type,
		Help:      p.Help,
		Default:   p.Default,
		Choices:   p.Choices,
		Required:  p.Required,
	}
}

// SetValueFromDefault assigns the default value of the ParameterDefinition to the given value.
// If the Default value is nil, the value is set to the zero value of the type.
//
// TODO(manuel, 2023-02-12) Not sure if the setting to the zero value of the type is the best idea, really.
func (p *ParameterDefinition) SetValueFromDefault(value reflect.Value) error {
	if !value.CanSet() {
		return errors.Errorf("cannot set value of %s", p.Name)
	}

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
			return helpers.SetReflectValue(value, 0)
		} else {
			return helpers.SetReflectValue(value, p.Default)
		}
	case ParameterTypeFloat:
		if p.Default == nil {
			return helpers.SetReflectValue(value, 0.0)
		} else {
			return helpers.SetReflectValue(value, p.Default)
		}
	case ParameterTypeStringList:
		if p.Default == nil {
			value.Set(reflect.ValueOf([]string{}))
		} else {
			value.Set(reflect.ValueOf(p.Default.([]string)))
		}
	case ParameterTypeDate:
		// TODO(manuel, 2023-02-12) Not sure exactly if this should be fully parsed at this point, or left up to the flag
		if p.Default == nil {
			// maybe this should be nil too (?)
			value.Set(reflect.ValueOf(time.Time{}))
		} else {
			s := p.Default.(string)
			dateTime, err := parseDate(s)
			if err != nil {
				return errors.Wrapf(err, "error parsing default value for parameter %s", p.Name)
			}
			value.Set(reflect.ValueOf(dateTime))
		}
	case ParameterTypeIntegerList:
		if p.Default == nil {
			return helpers.SetReflectValue(value, []int64{})
		} else {
			return helpers.SetReflectValue(value, p.Default)
		}
	case ParameterTypeFloatList:
		if p.Default == nil {
			return helpers.SetReflectValue(value, []float64{})
		} else {
			return helpers.SetReflectValue(value, p.Default)
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
			list2, b := helpers.CastList2[map[string]interface{}, interface{}](p.Default)
			if !b {
				return errors.Errorf("default value for parameter %s is not a list of maps", p.Name)
			}
			value.Set(reflect.ValueOf(list2))
		}
	case ParameterTypeObjectFromFile:
		if p.Default == nil {
			value.Set(reflect.ValueOf(map[string]interface{}{}))
		} else {
			value.Set(reflect.ValueOf(p.Default.(map[string]interface{})))
		}
	case ParameterTypeKeyValue:
		if p.Default == nil {
			value.Set(reflect.ValueOf(map[string]string{}))
		} else {
			v, ok := p.Default.(map[string]interface{})
			if !ok {
				return errors.Errorf("default value for parameter %s is not a map[string]interface{}", p.Name)
			}
			v2, ok := helpers.CastStringMap[string, interface{}](v)
			if !ok {
				return errors.Errorf("default value for parameter %s is not a map[string]interface{}", p.Name)
			}
			value.Set(reflect.ValueOf(v2))
		}

	default:
		return errors.Errorf("unknown parameter type %s", p.Type)
	}

	return nil
}

func InitializeStructFromParameterDefinitions(s interface{}, parameterDefinitions map[string]*ParameterDefinition) error {
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

		if field.Type.Kind() == reflect.Ptr {
			if value.IsNil() {
				value.Set(reflect.New(field.Type.Elem()))
			}
			if field.Type.Elem().Kind() == reflect.Struct {
				err := InitializeStructFromParameterDefinitions(value.Interface(), parameterDefinitions)
				if err != nil {
					return errors.Wrapf(err, "failed to initialize struct for %s", v)
				}
			} else {
				err := parameter.SetValueFromDefault(value.Elem())
				if err != nil {
					return errors.Wrapf(err, "failed to set value for %s", v)
				}
			}

		}

		err := parameter.SetValueFromDefault(value)
		if err != nil {
			return errors.Wrapf(err, "failed to set value for %s", v)
		}
	}

	return nil
}

func InitializeStructFromParameters(s interface{}, parameters map[string]interface{}) error {
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
		v_, ok := parameters[v]
		if !ok {
			continue
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		if field.Type.Kind() == reflect.Ptr {
			elem := field.Type.Elem()
			if value.IsNil() {
				value.Set(reflect.New(elem))
			} else {
				//exhaustive:ignore
				switch elem.Kind() {
				case reflect.Struct:
					err := InitializeStructFromParameters(value.Interface(), parameters)
					if err != nil {
						return errors.Wrapf(err, "failed to initialize struct for %s", v)
					}
				default:
					err := helpers.SetReflectValue(value.Elem(), v_)
					if err != nil {
						return errors.Wrapf(err, "failed to set value for %s", v)
					}
				}
			}

		} else {
			err := helpers.SetReflectValue(value, v_)
			if err != nil {
				return errors.Wrapf(err, "failed to set value for %s", v)
			}
		}
	}

	return nil
}

// CloneParameterDefinitionsWithDefaultsStruct clones the parameter definitions
// and sets the default values from the struct's tag `glazed.parameter`.
//
// TODO(manuel, 2023-02-12): This function is not necessary if we have a better way of initializing defaults
//
// This is more of a placeholder while we are refactoring things for
// https://github.com/go-go-golems/glazed/issues/132
func CloneParameterDefinitionsWithDefaultsStruct(
	parameterList []*ParameterDefinition,
	s interface{},
) ([]*ParameterDefinition, error) {
	ret := make([]*ParameterDefinition, len(parameterList))

	// gather default by parameter name from s
	defaults := map[string]reflect.Value{}
	// check that s is indeed a pointer to a struct
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return nil, errors.Errorf("s is not a pointer")
	}
	if reflect.TypeOf(s).Elem().Kind() != reflect.Struct {
		return nil, errors.Errorf("s is not a pointer to a struct")
	}
	st := reflect.TypeOf(s).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		v, ok := field.Tag.Lookup("glazed.parameter")
		if !ok {
			continue
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)
		defaults[v] = value
	}

	for i, p := range parameterList {
		ret[i] = p.Copy()
		if v, ok := defaults[p.Name]; ok {
			ret[i].Default = v.Interface()
		}
	}

	return ret, nil
}

type ParameterType string

const (
	ParameterTypeString         ParameterType = "string"
	ParameterTypeStringFromFile ParameterType = "stringFromFile"

	// TODO (2023-02-07) It would be great to have "list of objects from file" here
	// See https://github.com/go-go-golems/glazed/issues/117
	//
	// - string (potentially from file if starting with @)
	// - string/int/float list from file is another useful type

	ParameterTypeObjectListFromFile ParameterType = "objectListFromFile"
	ParameterTypeObjectFromFile     ParameterType = "objectFromFile"

	// ParameterTypeKeyValue signals either a string with comma separate key-value options, or when beginning with @, a file with key-value options
	ParameterTypeKeyValue ParameterType = "keyValue"

	ParameterTypeInteger     ParameterType = "int"
	ParameterTypeFloat       ParameterType = "float"
	ParameterTypeBool        ParameterType = "bool"
	ParameterTypeDate        ParameterType = "date"
	ParameterTypeStringList  ParameterType = "stringList"
	ParameterTypeIntegerList ParameterType = "intList"
	ParameterTypeFloatList   ParameterType = "floatList"
	ParameterTypeChoice      ParameterType = "choice"
)

// IsFileLoadingParameter returns true if the parameter type is one that loads a file, when provided with the given
// value. This slightly odd API is because some types like ParameterTypeKeyValue can be either a string or a file. A
// beginning character of @ indicates a file.
func IsFileLoadingParameter(p ParameterType, v string) bool {
	//exhaustive:ignore
	switch p {
	case ParameterTypeStringFromFile:
		return true
	case ParameterTypeObjectListFromFile:
		return true
	case ParameterTypeObjectFromFile:
		return true
	case ParameterTypeKeyValue:
		return strings.HasPrefix(v, "@")
	default:
		return false
	}
}

func IsListParameter(p ParameterType) bool {
	//exhaustive:ignore
	switch p {
	case ParameterTypeStringList:
		return true
	case ParameterTypeIntegerList:
		return true
	case ParameterTypeFloatList:
		return true
	case ParameterTypeKeyValue:
		return true
	default:
		return false
	}
}

func (p *ParameterDefinition) CheckParameterDefaultValueValidity() error {
	// we can have no default
	if p.Default == nil {
		return nil
	}

	switch p.Type {
	case ParameterTypeStringFromFile:
		fallthrough
	case ParameterTypeString:
		_, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

	case ParameterTypeObjectListFromFile:
		_, ok := p.Default.([]interface{})
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a list of objects: %v", p.Name, p.Default)
		}

	case ParameterTypeObjectFromFile:
		_, ok := p.Default.(map[string]interface{})
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an object: %v", p.Name, p.Default)
		}

	case ParameterTypeInteger:
		_, ok := p.Default.(int)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer: %v", p.Name, p.Default)
		}

	case ParameterTypeFloat:
		_, ok := p.Default.(float64)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a float: %v", p.Name, p.Default)
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
			defaultValue, ok := p.Default.([]interface{})
			if !ok {
				return errors.Errorf("Default value for parameter %s is not an integer list: %v", p.Name, p.Default)
			}
			_, ok = helpers.CastList[int, interface{}](defaultValue)
			if !ok {
				return errors.Errorf("Default value for parameter %s is not an integer list: %v", p.Name, p.Default)
			}
		}

	case ParameterTypeFloatList:
		_, ok := p.Default.([]float64)
		if !ok {
			defaultValue, ok := p.Default.([]interface{})
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a float list: %v", p.Name, p.Default)
			}
			_, ok = helpers.CastList[float64, interface{}](defaultValue)
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a float list: %v", p.Name, p.Default)
			}
		}

	case ParameterTypeChoice:
		if len(p.Choices) == 0 {
			return errors.Errorf("ParameterDefinition %s is a choice parameter but has no choices", p.Name)
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

	case ParameterTypeKeyValue:
		_, ok := p.Default.(map[string]string)
		if !ok {
			defaultValue, ok := p.Default.(map[string]interface{})
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a key value list: %v", p.Name, p.Default)
			}

			_, ok = helpers.CastStringMap[string, interface{}](defaultValue)
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a key value list: %v", p.Name, p.Default)
			}
		}
	}

	return nil
}

func (p *ParameterDefinition) ParseParameter(v []string) (interface{}, error) {
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

	case ParameterTypeObjectListFromFile:
		fileName := v[0]
		f, err := os.Open(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}

		objectList := []interface{}{}
		if strings.HasSuffix(fileName, ".json") {
			err = json.NewDecoder(f).Decode(&objectList)
		} else if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
			err = yaml.NewDecoder(f).Decode(&objectList)
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", fileName)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", fileName)
		}

		return objectList, nil

	case ParameterTypeKeyValue:
		ret := map[string]interface{}{}
		if len(v) == 1 && strings.HasPrefix(v[0], "@") {
			// load from file
			templateDataFile := v[0][1:]

			if strings.HasSuffix(templateDataFile, ".json") {
				err := helpers.LoadJSONFile(templateDataFile, &ret)
				if err != nil {
					return nil, errors.Wrapf(err, "Error loading template data from file %s", templateDataFile)
				}
			} else if strings.HasSuffix(templateDataFile, ".yaml") || strings.HasSuffix(templateDataFile, ".yml") {
				err := helpers.LoadYAMLFile(templateDataFile, &ret)
				if err != nil {
					return nil, errors.Wrapf(err, "Error loading template data from file %s", templateDataFile)
				}
			} else {
				return nil, errors.Errorf("Unknown template data file format for file %s", templateDataFile)
			}
		} else {
			for _, arg := range v {
				// TODO(2023-02-11): The separator could be stored in the parameter itself?
				// It was configurable before.
				//
				// See https://github.com/go-go-golems/glazed/issues/129
				parts := strings.Split(arg, ":")
				if len(parts) != 2 {
					return nil, errors.Errorf("Could not parse argument %s as key=value pair", arg)
				}
				ret[parts[0]] = parts[1]
			}
		}

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

func InitFlagsFromYaml(yamlContent []byte) (map[string]*ParameterDefinition, []*ParameterDefinition) {
	flags := make(map[string]*ParameterDefinition)
	flagList := make([]*ParameterDefinition, 0)

	var err error
	var parameters []*ParameterDefinition

	err = yaml.Unmarshal(yamlContent, &parameters)
	if err != nil {
		panic(errors.Wrap(err, "Failed to unmarshal output flags yaml"))
	}

	for _, p := range parameters {
		err := p.CheckParameterDefaultValueValidity()
		if err != nil {
			panic(errors.Wrap(err, "Failed to check parameter default value validity"))
		}
		flags[p.Name] = p
		flagList = append(flagList, p)
	}

	return flags, flagList
}
