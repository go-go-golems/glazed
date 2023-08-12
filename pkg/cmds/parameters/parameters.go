package parameters

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	reflect2 "github.com/go-go-golems/glazed/pkg/helpers/reflect"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"reflect"
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

type ParameterDefinitionOption func(*ParameterDefinition)

func NewParameterDefinition(name string, parameterType ParameterType, options ...ParameterDefinitionOption) *ParameterDefinition {
	ret := &ParameterDefinition{
		Name: name,
		Type: parameterType,
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func WithHelp(help string) ParameterDefinitionOption {
	return func(p *ParameterDefinition) {
		p.Help = help
	}
}

func WithShortFlag(shortFlag string) ParameterDefinitionOption {
	return func(p *ParameterDefinition) {
		p.ShortFlag = shortFlag
	}
}

func WithDefault(defaultValue interface{}) ParameterDefinitionOption {
	return func(p *ParameterDefinition) {
		p.Default = defaultValue
	}
}

func WithChoices(choices []string) ParameterDefinitionOption {
	return func(p *ParameterDefinition) {
		p.Choices = choices
	}
}

func WithRequired(required bool) ParameterDefinitionOption {
	return func(p *ParameterDefinition) {
		p.Required = required
	}
}

func (p *ParameterDefinition) IsEqualToDefault(i interface{}) bool {
	return reflect.DeepEqual(p.Default, i)
}

func (p *ParameterDefinition) SetDefaultFromInterface(i interface{}) error {
	switch p.Type {
	case ParameterTypeString:
		v, ok := i.(string)
		if !ok {
			return errors.Errorf("expected string for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeBool:
		v, ok := i.(bool)
		if !ok {
			return errors.Errorf("expected bool for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeInteger:
		v, ok := cast.CastNumberInterfaceToInt[int64](i)
		if !ok {
			return errors.Errorf("expected int64 for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeFloat:

		v, ok := cast.CastFloatInterfaceToFloat[float64](i)
		if !ok {
			return errors.Errorf("expected float64 for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeStringList:
		v, ok := cast.CastList2[string, interface{}](i)
		if !ok {
			return errors.Errorf("expected string list for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeDate:
		p.Default = i.(time.Time).Format(time.RFC3339)
	case ParameterTypeIntegerList:
		v, ok := cast.CastInterfaceToIntList[int64](i)
		if !ok {
			return errors.Errorf("expected integer list for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeFloatList:
		v, ok := cast.CastInterfaceToFloatList[float64](i)
		if !ok {
			return errors.Errorf("expected float list for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeChoice:
		p.Default = i.(string)
	case ParameterTypeStringFromFiles:
		fallthrough
	case ParameterTypeStringFromFile:
		p.Default = i.(string)
	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringListFromFile:
		v, ok := cast.CastList2[string, interface{}](i)
		if !ok {
			return errors.Errorf("expected string list for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeKeyValue:
		v, ok := cast.CastInterfaceToStringMap[string, interface{}](i)
		if !ok {
			return errors.Errorf("expected string map for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeObjectFromFile:
		v, ok := cast.CastInterfaceToStringMap[interface{}, interface{}](i)
		if !ok {
			return errors.Errorf("expected object for parameter %s, got %T", p.Name, i)
		}
		p.Default = v
	case ParameterTypeObjectListFromFiles:
		fallthrough
	case ParameterTypeObjectListFromFile:
		v, ok := cast.CastList2[map[string]interface{}, interface{}](i)
		if !ok {
			return errors.Errorf("expected object list for parameter %s, got %T", p.Name, i)
		}
		p.Default = v

	}

	return nil
}

func (p *ParameterDefinition) SetDefaultFromValue(value reflect.Value) error {
	// check if value is pointer, do nothing if nil, otherwise dereference
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	switch p.Type {
	case ParameterTypeString:
		p.Default = value.String()
	case ParameterTypeBool:
		p.Default = value.Bool()
	case ParameterTypeInteger:
		p.Default = value.Int()
	case ParameterTypeFloat:
		p.Default = value.Float()

	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringListFromFile:
		fallthrough
	case ParameterTypeStringList:
		v, ok := cast.CastList2[string, interface{}](value.Interface())
		if !ok {
			return errors.Errorf("expected string list for parameter %s, got %T", p.Name, value.Interface())
		}
		p.Default = v

	case ParameterTypeDate:
		p.Default = value.Interface().(time.Time).Format(time.RFC3339)
	case ParameterTypeIntegerList:
		v, ok := cast.CastList2[int64, interface{}](value.Interface())
		if !ok {
			return errors.Errorf("expected integer list for parameter %s, got %T", p.Name, value.Interface())
		}
		p.Default = v
	case ParameterTypeFloatList:
		v, ok := cast.CastList2[float64, interface{}](value.Interface())
		if !ok {
			return errors.Errorf("expected float list for parameter %s, got %T", p.Name, value.Interface())
		}
		p.Default = v

	case ParameterTypeChoice:
		p.Default = value.String()
	case ParameterTypeStringFromFiles:
		fallthrough
	case ParameterTypeStringFromFile:
		p.Default = value.String()
	case ParameterTypeKeyValue:
		v, ok := cast.CastInterfaceToStringMap[string, interface{}](value.Interface())
		if !ok {
			return errors.Errorf("expected string map for parameter %s, got %T", p.Name, value.Interface())
		}
		p.Default = v
	case ParameterTypeObjectFromFile:
		v, ok := cast.CastInterfaceToStringMap[interface{}, interface{}](value.Interface())
		if !ok {
			return errors.Errorf("expected object for parameter %s, got %T", p.Name, value.Interface())
		}
		p.Default = v

	case ParameterTypeObjectListFromFiles:
		fallthrough
	case ParameterTypeObjectListFromFile:
		v, ok := cast.CastList2[map[string]interface{}, interface{}](value.Interface())
		if !ok {
			return errors.Errorf("expected object list for parameter %s, got %T", p.Name, value.Interface())
		}
		p.Default = v
	}
	return nil
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
			return reflect2.SetReflectValue(value, 0)
		} else {
			return reflect2.SetReflectValue(value, p.Default)
		}
	case ParameterTypeFloat:
		if p.Default == nil {
			return reflect2.SetReflectValue(value, 0.0)
		} else {
			return reflect2.SetReflectValue(value, p.Default)
		}
	case ParameterTypeStringList:
		if p.Default == nil {
			value.Set(reflect.ValueOf([]string{}))
		} else {
			v_, ok := cast.CastList2[string, interface{}](p.Default)
			if !ok {
				return errors.Errorf("expected string list for parameter %s, got %T", p.Name, p.Default)
			}
			value.Set(reflect.ValueOf(v_))
		}
	case ParameterTypeDate:
		// TODO(manuel, 2023-02-12) Not sure exactly if this should be fully parsed at this point, or left up to the flag
		if p.Default == nil {
			// maybe this should be nil too (?)
			value.Set(reflect.ValueOf(time.Time{}))
		} else {
			s := p.Default.(string)
			dateTime, err := ParseDate(s)
			if err != nil {
				return errors.Wrapf(err, "error parsing default value for parameter %s", p.Name)
			}
			value.Set(reflect.ValueOf(dateTime))
		}
	case ParameterTypeIntegerList:
		if p.Default == nil {
			return reflect2.SetReflectValue(value, []int64{})
		} else {
			return reflect2.SetReflectValue(value, p.Default)
		}
	case ParameterTypeFloatList:
		if p.Default == nil {
			return reflect2.SetReflectValue(value, []float64{})
		} else {
			return reflect2.SetReflectValue(value, p.Default)
		}
	case ParameterTypeChoice:
		if p.Default == nil {
			value.SetString("")
		} else {
			value.SetString(p.Default.(string))
		}
	case ParameterTypeStringFromFiles:
		fallthrough
	case ParameterTypeStringFromFile:
		if p.Default == nil {
			value.SetString("")
		} else {
			value.SetString(p.Default.(string))
		}
	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringListFromFile:
		if p.Default == nil {
			value.Set(reflect.ValueOf([]string{}))
		} else {
			list, b := cast.CastList2[string, interface{}](p.Default)
			if !b {
				return errors.Errorf("default value for parameter %s is not a list of strings", p.Name)
			}
			value.Set(reflect.ValueOf(list))
		}
	case ParameterTypeObjectListFromFiles:
		fallthrough
	case ParameterTypeObjectListFromFile:
		if p.Default == nil {
			value.Set(reflect.ValueOf([]map[string]interface{}{}))
		} else {
			list2, b := cast.CastList2[map[string]interface{}, interface{}](p.Default)
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
			v2, ok := cast.CastStringMap[string, interface{}](v)
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

// InitializeStructFromParameterDefinitions initializes a struct from a map of parameter definitions.
//
// Each field in the struct annotated with tag `glazed.parameter` will be set to the default value of
// the corresponding `ParameterDefinition`. If no `ParameterDefinition` is found for a field, an error
// is returned.
func InitializeStructFromParameterDefinitions(
	s interface{},
	parameterDefinitions map[string]*ParameterDefinition,
) error {
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

// InitializeParameterDefinitionsFromStruct initializes a map of parameter definitions from a struct.
// Each field in the struct annotated with tag `glazed.parameter` will be used to set
// the default value of the corresponding definition in `parameterDefinitions`.
// If no `ParameterDefinition` is found for a field, an error is returned.
// This is the inverse of InitializeStructFromParameterDefinitions.
func InitializeParameterDefinitionsFromStruct(
	parameterDefinitions map[string]*ParameterDefinition,
	s interface{},
) error {
	// check that s is indeed a pointer to a struct
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return errors.Errorf("s is not a pointer")
	}
	// check if nil
	if reflect.ValueOf(s).IsNil() {
		return nil
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

		err := parameter.SetDefaultFromValue(value)
		if err != nil {
			return errors.Wrapf(err, "failed to set default value for %s", v)
		}
	}

	return nil
}

func InitializeParameterDefaultsFromParameters(
	parameterDefinitions map[string]*ParameterDefinition,
	ps map[string]interface{},
) error {
	for k, v := range ps {
		parameter, ok := parameterDefinitions[k]
		if !ok {
			return errors.Errorf("unknown parameter %s", k)
		}
		err := parameter.SetDefaultFromValue(reflect.ValueOf(v))
		if err != nil {
			return errors.Wrapf(err, "failed to set default value for %s", k)
		}
	}

	return nil
}

func InitializeStructFromParameters(s interface{}, ps map[string]interface{}) error {
	if s == nil {
		return errors.Errorf("Can't initialize nil struct")
	}
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
		v_, ok := ps[v]
		if !ok {
			continue
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		if field.Type.Kind() == reflect.Ptr {
			elem := field.Type.Elem()
			if value.IsNil() {
				value.Set(reflect.New(elem))
			}
			//exhaustive:ignore
			switch elem.Kind() {
			case reflect.Struct:
				err := InitializeStructFromParameters(value.Interface(), ps)
				if err != nil {
					return errors.Wrapf(err, "failed to initialize struct for %s", v)
				}
			default:
				err := reflect2.SetReflectValue(value.Elem(), v_)
				if err != nil {
					return errors.Wrapf(err, "failed to set value for %s", v)
				}
			}

		} else {
			err := reflect2.SetReflectValue(value, v_)
			if err != nil {
				return errors.Wrapf(err, "failed to set value for %s", v)
			}
		}
	}

	return nil
}

type ParameterType string

const (
	ParameterTypeString ParameterType = "string"

	// TODO(2023-02-13, manuel) Should the "default" of a stringFromFile be the filename, or the string?
	//
	// See https://github.com/go-go-golems/glazed/issues/137

	ParameterTypeStringFromFile  ParameterType = "stringFromFile"
	ParameterTypeStringFromFiles ParameterType = "stringFromFiles"

	// TODO (2023-02-07) It would be great to have "list of objects from file" here
	// See https://github.com/go-go-golems/glazed/issues/117
	//
	// - string (potentially from file if starting with @)
	// - string/int/float list from file is another useful type

	ParameterTypeObjectListFromFile  ParameterType = "objectListFromFile"
	ParameterTypeObjectListFromFiles ParameterType = "objectListFromFiles"
	ParameterTypeObjectFromFile      ParameterType = "objectFromFile"
	ParameterTypeStringListFromFile  ParameterType = "stringListFromFile"
	ParameterTypeStringListFromFiles ParameterType = "stringListFromFiles"

	// ParameterTypeKeyValue signals either a string with comma separate key-value options,
	// or when beginning with @, a file with key-value options
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
	case ParameterTypeStringListFromFile:
		return true
	case ParameterTypeObjectListFromFiles:
		return true
	case ParameterTypeStringListFromFiles:
		return true
	case ParameterTypeStringFromFiles:
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
	case ParameterTypeObjectListFromFiles:
		return true
	case ParameterTypeStringListFromFiles:
		return true
	case ParameterTypeStringFromFiles:
		return true
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
	v := p.Default
	return p.CheckValueValidity(v)
}

func (p *ParameterDefinition) CheckValueValidity(v interface{}) error {
	if v == nil {
		return nil
	}

	switch p.Type {
	case ParameterTypeStringFromFile:
		fallthrough
	case ParameterTypeStringFromFiles:
		fallthrough
	case ParameterTypeString:
		_, ok := v.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, v)
		}

	case ParameterTypeObjectListFromFile:
		fallthrough
	case ParameterTypeObjectListFromFiles:
		_, ok := v.([]interface{})
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a list of objects: %v", p.Name, v)
		}

	case ParameterTypeObjectFromFile:
		_, ok := v.(map[string]interface{})
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an object: %v", p.Name, v)
		}

	case ParameterTypeInteger:
		_, ok := cast.CastNumberInterfaceToInt[int64](v)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer: %v", p.Name, v)
		}

	case ParameterTypeFloat:
		_, ok := cast.CastFloatInterfaceToFloat[float64](v)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a float: %v", p.Name, v)
		}

	case ParameterTypeBool:
		_, ok := v.(bool)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a bool: %v", p.Name, v)
		}

	case ParameterTypeDate:
		switch v_ := v.(type) {
		case string:
			_, err := ParseDate(v_)
			if err != nil {
				return errors.Wrapf(err, "Default value for parameter %s is not a valid date: %v", p.Name, v)
			}
		case time.Time:
			return nil
		default:
			return errors.Errorf("Default value for parameter %s is not a valid date: %v", p.Name, v)
		}

	case ParameterTypeStringListFromFile:
		fallthrough
	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringList:
		_, ok := v.([]string)
		if !ok {
			defaultValue, ok := v.([]interface{})
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a string list: %v", p.Name, v)
			}

			// convert to string list
			fixedDefault, ok := cast.CastList[string, interface{}](defaultValue)
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a string list: %v", p.Name, v)
			}
			_ = fixedDefault
		}

	case ParameterTypeIntegerList:
		_, ok := cast.CastInterfaceToIntList[int64](v)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer list: %v", p.Name, v)
		}

	case ParameterTypeFloatList:
		_, ok := cast.CastInterfaceToFloatList[float64](v)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a float list: %v", p.Name, v)
		}

	case ParameterTypeChoice:
		if len(p.Choices) == 0 {
			return errors.Errorf("ParameterDefinition %s is a choice parameter but has no choices", p.Name)
		}

		defaultValue, ok := v.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, v)
		}

		found := false
		for _, choice := range p.Choices {
			if choice == defaultValue {
				found = true
			}
		}
		if !found {
			return errors.Errorf("Default value for parameter %s is not a valid choice: %v", p.Name, v)
		}

	case ParameterTypeKeyValue:
		_, ok := v.(map[string]string)
		if !ok {
			defaultValue, ok := v.(map[string]interface{})
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a key value list: %v", p.Name, v)
			}

			_, ok = cast.CastStringMap[string, interface{}](defaultValue)
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a key value list: %v", p.Name, v)
			}
		}
	}

	return nil
}

func LoadParameterDefinitionsFromYAML(yamlContent []byte) (map[string]*ParameterDefinition, []*ParameterDefinition) {
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
