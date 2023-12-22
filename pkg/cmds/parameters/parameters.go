package parameters

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/helpers/maps"
	reflect2 "github.com/go-go-golems/glazed/pkg/helpers/reflect"
	"github.com/pkg/errors"
	"github.com/wk8/go-ordered-map/v2"
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
	Name       string        `yaml:"name"`
	ShortFlag  string        `yaml:"shortFlag,omitempty"`
	Type       ParameterType `yaml:"type"`
	Help       string        `yaml:"help,omitempty"`
	Default    interface{}   `yaml:"default,omitempty"`
	Choices    []string      `yaml:"choices,omitempty"`
	Required   bool          `yaml:"required,omitempty"`
	IsArgument bool          `yaml:"-"`
}

func (p *ParameterDefinition) String() string {
	return fmt.Sprintf("{Parameter: %s - %s}", p.Name, p.Type)
}

func (p *ParameterDefinition) Clone() *ParameterDefinition {
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

func WithIsArgument(isArgument bool) ParameterDefinitionOption {
	return func(p *ParameterDefinition) {
		p.IsArgument = isArgument
	}
}

func (p *ParameterDefinition) IsEqualToDefault(i interface{}) bool {
	return reflect.DeepEqual(p.Default, i)
}

func (p *ParameterDefinition) SetDefaultFromValue(value reflect.Value) error {
	// check if value is pointer, do nothing if nil, otherwise dereference
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	if p.CheckValueValidity(value.Interface()) != nil {
		return errors.Errorf("invalid value for parameter %s: %v", p.Name, value.Interface())
	}

	val := reflect.ValueOf(p).Elem()
	f := val.FieldByName("Default")
	if f.CanSet() {
		f.Set(reflect.ValueOf(value.Interface()))
	}

	return nil
}

// SetValueFromDefault assigns the default value of the ParameterDefinition to the given value.
// If the Default value is nil, the value is set to the zero value of the type.
func (p *ParameterDefinition) SetValueFromDefault(value reflect.Value) error {
	if !value.CanSet() {
		return errors.Errorf("cannot set value of %s", p.Name)
	}

	if p.Default != nil {
		return p.SetValueFromInterface(value, p.Default)
	}
	return p.InitializeValueToEmptyValue(value)
}

// InitializeValueToEmptyValue initializes the given value to the empty value of the type of the parameter.
func (p *ParameterDefinition) InitializeValueToEmptyValue(value reflect.Value) error {
	switch p.Type {
	case ParameterTypeString, ParameterTypeChoice, ParameterTypeStringFromFiles, ParameterTypeStringFromFile:
		value.SetString("")
	case ParameterTypeBool:
		value.SetBool(false)
	case ParameterTypeInteger, ParameterTypeFloat:
		return reflect2.SetReflectValue(value, 0)
	case ParameterTypeStringList, ParameterTypeChoiceList, ParameterTypeStringListFromFiles, ParameterTypeStringListFromFile:
		value.Set(reflect.ValueOf([]string{}))
	case ParameterTypeDate:
		value.Set(reflect.ValueOf(time.Time{}))
	case ParameterTypeIntegerList:
		return reflect2.SetReflectValue(value, []int64{})
	case ParameterTypeFloatList:
		return reflect2.SetReflectValue(value, []float64{})
	case ParameterTypeObjectListFromFiles, ParameterTypeObjectListFromFile:
		value.Set(reflect.ValueOf([]map[string]interface{}{}))
	case ParameterTypeObjectFromFile:
		value.Set(reflect.ValueOf(map[string]interface{}{}))
	case ParameterTypeKeyValue:
		value.Set(reflect.ValueOf(map[string]string{}))
	case ParameterTypeFile:
		value.Set(reflect.ValueOf(nil))
	case ParameterTypeFileList:
		value.Set(reflect.ValueOf([]*FileData{}))
	default:
		return errors.Errorf("unknown parameter type %s", p.Type)
	}
	return nil
}

// SetValueFromInterface assigns the given value to the given reflect.Value.
func (p *ParameterDefinition) SetValueFromInterface(value reflect.Value, v interface{}) error {
	err := p.CheckValueValidity(v)
	if err != nil {
		return err
	}

	switch p.Type {
	case ParameterTypeString, ParameterTypeChoice, ParameterTypeStringFromFiles, ParameterTypeStringFromFile:
		strVal, ok := v.(string)
		if !ok {
			return errors.Errorf("expected string value for parameter %s, got %T", p.Name, v)
		}
		value.SetString(strVal)

	case ParameterTypeBool:
		boolVal, ok := v.(bool)
		if !ok {
			return errors.Errorf("expected bool value for parameter %s, got %T", p.Name, v)
		}
		value.SetBool(boolVal)

	case ParameterTypeInteger, ParameterTypeFloat:
		return reflect2.SetReflectValue(value, v)

	case ParameterTypeStringList, ParameterTypeChoiceList, ParameterTypeStringListFromFiles, ParameterTypeStringListFromFile:
		list, ok := cast.CastList2[string, interface{}](v)
		if !ok {
			return errors.Errorf("expected string list for parameter %s, got %T", p.Name, v)
		}
		value.Set(reflect.ValueOf(list))

	case ParameterTypeDate:
		strVal, ok := v.(string)
		if !ok {
			return errors.Errorf("expected string value for parameter %s, got %T", p.Name, v)
		}
		dateTime, err := ParseDate(strVal)
		if err != nil {
			return errors.Wrapf(err, "error parsing value for parameter %s", p.Name)
		}
		value.Set(reflect.ValueOf(dateTime))

	case ParameterTypeIntegerList, ParameterTypeFloatList:
		return reflect2.SetReflectValue(value, v)

	case ParameterTypeFile:
		return reflect2.SetReflectValue(value, v)

	case ParameterTypeFileList:
		list, ok := cast.CastList2[*FileData, interface{}](v)
		if !ok {
			return errors.Errorf("expected list of files for parameter %s, got %T", p.Name, v)
		}
		value.Set(reflect.ValueOf(list))

	case ParameterTypeObjectListFromFiles, ParameterTypeObjectListFromFile:
		list, ok := cast.CastList2[map[string]interface{}, interface{}](v)
		if !ok {
			return errors.Errorf("expected list of maps for parameter %s, got %T", p.Name, v)
		}
		value.Set(reflect.ValueOf(list))

	case ParameterTypeObjectFromFile:
		mapVal, ok := v.(map[string]interface{})
		if !ok {
			return errors.Errorf("expected map for parameter %s, got %T", p.Name, v)
		}
		value.Set(reflect.ValueOf(mapVal))

	case ParameterTypeKeyValue:
		mapVal, ok := v.(map[string]interface{})
		if !ok {
			return errors.Errorf("expected map for parameter %s, got %T", p.Name, v)
		}
		mapStrVal, ok := cast.CastStringMap[string, interface{}](mapVal)
		if !ok {
			return errors.Errorf("expected map of strings for parameter %s, got %T", p.Name, v)
		}
		value.Set(reflect.ValueOf(mapStrVal))

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
	parameterDefinitions ParameterDefinitions,
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
		parameter, ok := parameterDefinitions.Get(v)
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
	parameterDefinitions ParameterDefinitions,
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
		parameter, ok := parameterDefinitions.Get(v)
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

func GlazedStructToMap(s interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{}

	// check that s is indeed a pointer to a struct
	if !maps.IsStructPointer(s) {
		return nil, errors.Errorf("s is not a pointer to a struct")
	}

	st := reflect.TypeOf(s).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		parameterName, ok := field.Tag.Lookup("glazed.parameter")
		if !ok {
			continue
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		ret[parameterName] = value
	}

	return ret, nil
}

func InitializeParameterDefaultsFromParameters(
	parameterDefinitions ParameterDefinitions,
	ps map[string]interface{},
) error {
	for k, v := range ps {
		parameter, ok := parameterDefinitions.Get(k)
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

func InitializeStructFromParameters(s interface{}, ps *ParsedParameters) error {
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
		v_, ok := ps.Get(v)
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
				err := reflect2.SetReflectValue(value.Elem(), v_.Value)
				if err != nil {
					return errors.Wrapf(err, "failed to set value for %s", v)
				}
			}

		} else {
			err := reflect2.SetReflectValue(value, v_.Value)
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

	// ParameterTypeFile and ParameterTypeFileList are a more elaborate version that loads and parses
	// the file content and returns a list of FileData objects (or a single object in the case
	// of ParameterTypeFile).
	ParameterTypeFile     ParameterType = "file"
	ParameterTypeFileList ParameterType = "fileList"

	// TODO(manuel, 2023-09-19) Add some more types and maybe revisit the entire concept of loading things from files
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
	ParameterTypeChoiceList  ParameterType = "choiceList"
)

// IsFileLoadingParameter returns true if the parameter type is one that loads a file, when provided with the given
// value. This slightly odd API is because some types like ParameterTypeKeyValue can be either a string or a file. A
// beginning character of @ indicates a file.
func IsFileLoadingParameter(p ParameterType, v string) bool {
	//exhaustive:ignore
	switch p {
	case ParameterTypeStringFromFile,
		ParameterTypeObjectListFromFile,
		ParameterTypeObjectFromFile,
		ParameterTypeStringListFromFile,
		ParameterTypeObjectListFromFiles,
		ParameterTypeStringListFromFiles,
		ParameterTypeStringFromFiles,
		ParameterTypeFile,
		ParameterTypeFileList:

		return true
	case ParameterTypeKeyValue:
		return strings.HasPrefix(v, "@")
	default:
		return false
	}
}

// IsListParameter returns if the parameter has to be parsed from a list of strings,
// not if its value is actually a string.
func IsListParameter(p ParameterType) bool {
	//exhaustive:ignore
	switch p {
	case ParameterTypeObjectListFromFiles,
		ParameterTypeStringListFromFiles,
		ParameterTypeStringFromFiles,
		ParameterTypeStringList,
		ParameterTypeIntegerList,
		ParameterTypeFloatList,
		ParameterTypeChoiceList,
		ParameterTypeKeyValue,
		ParameterTypeFileList:
		return true
	default:
		return false
	}
}

// CheckParameterDefaultValueValidity checks if the ParameterDefinition's Default is valid.
// This is used when validating loading from a YAML file or setting up cobra flag definitions.
func (p *ParameterDefinition) CheckParameterDefaultValueValidity() error {
	// we can have no default
	v := p.Default
	return p.CheckValueValidity(v)
}

// CheckValueValidity checks if the given value is valid for the ParameterDefinition.
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
			return errors.Errorf("Value for parameter %s is not a string: %v", p.Name, v)
		}

	case ParameterTypeObjectListFromFile:
		fallthrough
	case ParameterTypeObjectListFromFiles:
		_, ok := v.([]interface{})
		if !ok {
			return errors.Errorf("Value for parameter %s is not a list of objects: %v", p.Name, v)
		}

	case ParameterTypeObjectFromFile:
		_, ok := v.(map[string]interface{})
		if !ok {
			return errors.Errorf("Value for parameter %s is not an object: %v", p.Name, v)
		}

	case ParameterTypeInteger:
		_, ok := cast.CastNumberInterfaceToInt[int64](v)
		if !ok {
			return errors.Errorf("Value for parameter %s is not an integer: %v", p.Name, v)
		}

	case ParameterTypeFloat:
		_, ok := cast.CastNumberInterfaceToFloat[float64](v)
		if !ok {
			return errors.Errorf("Value for parameter %s is not a float: %v", p.Name, v)
		}

	case ParameterTypeBool:
		_, ok := v.(bool)
		if !ok {
			return errors.Errorf("Value for parameter %s is not a bool: %v", p.Name, v)
		}

	case ParameterTypeDate:
		switch v_ := v.(type) {
		case string:
			_, err := ParseDate(v_)
			if err != nil {
				return errors.Wrapf(err, "Value for parameter %s is not a valid date: %v", p.Name, v)
			}
		case time.Time:
			return nil
		default:
			return errors.Errorf("Value for parameter %s is not a valid date: %v", p.Name, v)
		}

	case ParameterTypeFile:
		_, ok := v.(*FileData)
		if !ok {
			return errors.Errorf("Value for parameter %s is not a file: %v", p.Name, v)
		}
		return nil

	case ParameterTypeFileList:
		_, ok := cast.CastList2[*FileData, interface{}](v)
		if !ok {
			return errors.Errorf("Value for parameter %s is not a file list: %v", p.Name, v)
		}
		return nil

	case ParameterTypeStringListFromFile:
		fallthrough
	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringList:
		_, ok := v.([]string)
		if !ok {
			v_, ok := v.([]interface{})
			if !ok {
				return errors.Errorf("Value for parameter %s is not a string list: %v", p.Name, v)
			}

			// convert to string list
			fixedDefault, ok := cast.CastList[string, interface{}](v_)
			if !ok {
				return errors.Errorf("Value for parameter %s is not a string list: %v", p.Name, v)
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
			return errors.Errorf("Value for parameter %s is not a float list: %v", p.Name, v)
		}

	case ParameterTypeChoice:
		if len(p.Choices) == 0 {
			return errors.Errorf("ParameterDefinition %s is a choice parameter but has no choices", p.Name)
		}

		v_, ok := v.(string)
		if !ok {
			return errors.Errorf("Value for parameter %s is not a string: %v", p.Name, v)
		}

		err := p.checkChoiceValidity(v_)
		if err != nil {
			return err
		}

	case ParameterTypeChoiceList:
		if len(p.Choices) == 0 {
			return errors.Errorf("ParameterDefinition %s is a choice parameter but has no choices", p.Name)
		}

		v_, ok := cast.CastList2[string, interface{}](v)
		if !ok {
			return errors.Errorf("Value for parameter %s is not a string list: %v", p.Name, v)
		}

		for _, choice := range v_ {
			err := p.checkChoiceValidity(choice)
			if err != nil {
				return err
			}
		}

	case ParameterTypeKeyValue:
		_, ok := v.(map[string]string)
		if !ok {
			v_, ok := v.(map[string]interface{})
			if !ok {
				return errors.Errorf("Value for parameter %s is not a key value list: %v", p.Name, v)
			}

			_, ok = cast.CastStringMap[string, interface{}](v_)
			if !ok {
				return errors.Errorf("Value for parameter %s is not a key value list: %v", p.Name, v)
			}
		}
	}

	return nil
}

func (p *ParameterDefinition) checkChoiceValidity(choice string) error {
	found := false
	for _, choice2 := range p.Choices {
		if choice == choice2 {
			found = true
		}
	}
	if !found {
		return errors.Errorf("Value for parameter %s is not a valid choice: %v", p.Name, choice)
	}
	return nil
}

// LoadParameterDefinitionsFromYAML loads a map of ParameterDefinitions from a YAML file.
// It checks that default values are valid.
// It returns the ParameterDefinitions as a map indexed by name, and as a list.
func LoadParameterDefinitionsFromYAML(yamlContent []byte) ParameterDefinitions {
	parameters_ := NewParameterDefinitions()

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
		parameters_.Set(p.Name, p)
	}

	return parameters_
}

type ParameterDefinitions struct {
	*orderedmap.OrderedMap[string, *ParameterDefinition]
}

func (p ParameterDefinitions) Merge(m ParameterDefinitions) ParameterDefinitions {
	for v := m.Oldest(); v != nil; v = v.Next() {
		p.Set(v.Key, v.Value)
	}
	return p
}

func (p ParameterDefinitions) GetFlags() ParameterDefinitions {
	ret := NewParameterDefinitions()

	for v := p.Oldest(); v != nil; v = v.Next() {
		if !v.Value.IsArgument {
			ret.Set(v.Key, v.Value)
		}
	}

	return ret
}

func (p ParameterDefinitions) GetArguments() ParameterDefinitions {
	ret := NewParameterDefinitions()

	for v := p.Oldest(); v != nil; v = v.Next() {
		if v.Value.IsArgument {
			ret.Set(v.Key, v.Value)
		}
	}

	return ret
}

func (p ParameterDefinitions) ForEachE(f func(definition *ParameterDefinition) error) error {
	for v := p.Oldest(); v != nil; v = v.Next() {
		err := f(v.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p ParameterDefinitions) ForEach(f func(definition *ParameterDefinition)) {
	for v := p.Oldest(); v != nil; v = v.Next() {
		f(v.Value)
	}
}

type ParameterDefinitionsOption func(*ParameterDefinitions)

func WithParameterDefinitions(parameterDefinitions ParameterDefinitions) ParameterDefinitionsOption {
	return func(p *ParameterDefinitions) {
		p.Merge(parameterDefinitions)
	}
}

func WithParameterDefinitionList(parameterDefinitions []*ParameterDefinition) ParameterDefinitionsOption {
	return func(p *ParameterDefinitions) {
		for _, pd := range parameterDefinitions {
			p.Set(pd.Name, pd)
		}
	}
}

func NewParameterDefinitions(options ...ParameterDefinitionsOption) ParameterDefinitions {
	ret := ParameterDefinitions{
		orderedmap.New[string, *ParameterDefinition](),
	}

	for _, o := range options {
		o(&ret)
	}

	return ret
}

func (p ParameterDefinitions) MarshalYAML() (interface{}, error) {
	ret := []*ParameterDefinition{}
	p.ForEach(func(definition *ParameterDefinition) {
		ret = append(ret, definition)
	})
	return ret, nil
}

func (p ParameterDefinitions) UnmarshalYAML(value *yaml.Node) error {
	var parameterDefinitions []*ParameterDefinition
	err := value.Decode(&parameterDefinitions)
	if err != nil {
		return err
	}

	for _, pd := range parameterDefinitions {
		p.Set(pd.Name, pd)
	}

	return nil
}

func (p ParameterDefinitions) Clone() ParameterDefinitions {
	return NewParameterDefinitions().Merge(p)

}
