package parameters

import (
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	reflect2 "github.com/go-go-golems/glazed/pkg/helpers/reflect"
	"github.com/pkg/errors"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
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
	Default    *interface{}  `yaml:"default,omitempty"`
	Choices    []string      `yaml:"choices,omitempty"`
	Required   bool          `yaml:"required,omitempty"`
	IsArgument bool          `yaml:"-"`
}

type ParameterDefinitionOption func(*ParameterDefinition)

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
		p.Default = &defaultValue
	}
}

func WithChoices(choices ...string) ParameterDefinitionOption {
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

func NewParameterDefinition(
	name string,
	parameterType ParameterType,
	options ...ParameterDefinitionOption,
) *ParameterDefinition {
	ret := &ParameterDefinition{
		Name: name,
		Type: parameterType,
	}

	for _, o := range options {
		o(ret)
	}

	return ret
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

func (p *ParameterDefinition) IsEqualToDefault(i interface{}) bool {
	if p.Default == nil {
		return false
	}
	return reflect.DeepEqual(*p.Default, i)
}

// SetDefaultFromValue sets the Default field of the ParameterDefinition
// to the provided value. It handles nil values and dereferencing pointers.
// The value is type checked before being set.
func (p *ParameterDefinition) SetDefaultFromValue(value reflect.Value) error {
	// check if value is pointer, do nothing if nil, otherwise dereference
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	_, err := p.CheckValueValidity(value.Interface())
	if err != nil {
		return errors.Errorf("invalid value for parameter %s: %v", p.Name, value.Interface())
	}

	v_ := value.Interface()

	val := reflect.ValueOf(p).Elem()
	f := val.FieldByName("Default")
	if f.CanSet() {
		f.Set(reflect.ValueOf(&v_))
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
		return p.SetValueFromInterface(value, *p.Default)
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
	case ParameterTypeStringList,
		ParameterTypeChoiceList,
		ParameterTypeStringListFromFiles,
		ParameterTypeStringListFromFile:
		return reflect2.SetReflectValue(value, []string{})
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

// SetValueFromInterface assigns the value v to the given reflect.Value based on the
// ParameterDefinition's type. It handles type checking and conversion for the
// various supported parameter types.
func (p *ParameterDefinition) SetValueFromInterface(value reflect.Value, v interface{}) error {
	_, err := p.CheckValueValidity(v)
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

	case ParameterTypeStringList,
		ParameterTypeChoiceList,
		ParameterTypeStringListFromFiles,
		ParameterTypeStringListFromFile:
		list, err := cast.CastListToStringList(v)
		if err != nil {
			return errors.Errorf("expected string list for parameter %s, got %T", p.Name, v)
		}
		return reflect2.SetReflectValue(value, list)

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

// ParsedParametersFromDefaults uses the parameter definitions default values to create a ParsedParameters
// object.
func (pds *ParameterDefinitions) ParsedParametersFromDefaults() (*ParsedParameters, error) {
	ret := NewParsedParameters()
	err := pds.ForEachE(func(definition *ParameterDefinition) error {
		err := ret.UpdateValue(definition.Name, definition, definition.Default,
			WithParseStepSource("defaults"),
			WithParseStepValue(definition.Default),
		)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// InitializeStructFromDefaults initializes a struct from a map of parameter definitions.
//
// Each field in the struct annotated with tag `glazed.parameter` will be set to the default value of
// the corresponding `ParameterDefinition`. If no `ParameterDefinition` is found for a field, an error
// is returned.
func (pds *ParameterDefinitions) InitializeStructFromDefaults(s interface{}) error {
	parsedParameters, err := pds.ParsedParametersFromDefaults()
	if err != nil {
		return err
	}
	return parsedParameters.InitializeStruct(s)
}

// InitializeDefaultsFromStruct initializes the parameters definitions from a struct.
// Each field in the struct annotated with tag `glazed.parameter` will be used to set
// the default value of the corresponding definition in `parameterDefinitions`.
// If no `ParameterDefinition` is found for a field, an error is returned.
// This is the inverse of InitializeStructFromDefaults.
func (pds *ParameterDefinitions) InitializeDefaultsFromStruct(
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
		tag, ok := field.Tag.Lookup("glazed.parameter")
		if !ok {
			continue
		}

		tagOptions, err := parsedTagOptions(tag)
		if err != nil {
			return err
		}

		if tagOptions.IsWildcard {
			err = pds.ForEachE(func(pd *ParameterDefinition) error {
				if matched, _ := filepath.Match(tagOptions.Name, pd.Name); matched {
					mapValue := reflect.ValueOf(s).Elem().FieldByName(field.Name)

					// check that mapValue is a map of strings
					if mapValue.Kind() != reflect.Map {
						return errors.Errorf("wildcard parameters require a map field, field %s is not a map", field.Name)
					}
					if mapValue.Type().Key().Kind() != reflect.String {
						return errors.Errorf("wildcard parameters require a map of strings, field %s is not a map of strings", field.Name)
					}

					// look up pd.Name in the map and if present, set the default
					value := mapValue.MapIndex(reflect.ValueOf(pd.Name))
					if value.IsValid() {
						err = pd.SetDefaultFromValue(value)
						if err != nil {
							return err
						}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
			continue
		}

		parameter, ok := pds.Get(tagOptions.Name)
		if !ok {
			return errors.Errorf("unknown parameter %s", tag)
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		err = parameter.SetDefaultFromValue(value)
		if err != nil {
			return errors.Wrapf(err, "failed to set default value for %s", tag)
		}
	}

	return nil
}

func (pds *ParameterDefinitions) InitializeDefaultsFromMap(
	ps map[string]interface{},
) error {
	for k, v := range ps {
		parameter, ok := pds.Get(k)
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

// CheckParameterDefaultValueValidity checks if the ParameterDefinition's Default is valid.
// This is used when validating loading from a YAML file or setting up cobra flag definitions.
func (p *ParameterDefinition) CheckParameterDefaultValueValidity() (interface{}, error) {
	// no default at all is valid
	if p.Default == nil {
		return nil, nil
	}
	// we can have no default
	v := *p.Default
	v, err := p.CheckValueValidity(v)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid default value for parameter %s", p.Name)
	}
	return v, nil
}

// CheckValueValidity checks if the given value is valid for the ParameterDefinition, and returns the value cast to the correct type.
func (p *ParameterDefinition) CheckValueValidity(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch p.Type {
	case ParameterTypeStringFromFile:
		fallthrough
	case ParameterTypeStringFromFiles:
		fallthrough
	case ParameterTypeString:
		s, err := cast.ToString(v)
		if err != nil {
			return nil, errors.Errorf("Value for parameter %s is not a string: %v", p.Name, v)
		}
		return s, nil

	case ParameterTypeObjectListFromFile:
		fallthrough
	case ParameterTypeObjectListFromFiles:
		l, ok := v.([]interface{})
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a list of objects: %v", p.Name, v)
		}
		return l, nil

	case ParameterTypeObjectFromFile:
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not an object: %v", p.Name, v)
		}
		return m, nil

	case ParameterTypeInteger:
		i, ok := cast.CastNumberInterfaceToInt[int](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not an integer: %v", p.Name, v)
		}
		return i, nil

	case ParameterTypeFloat:
		f, ok := cast.CastNumberInterfaceToFloat[float64](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a float: %v", p.Name, v)
		}
		return f, nil

	case ParameterTypeBool:
		b, ok := v.(bool)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a bool: %v", p.Name, v)
		}
		return b, nil

	case ParameterTypeDate:
		switch v_ := v.(type) {
		case string:
			d, err := ParseDate(v_)
			if err != nil {
				return nil, errors.Wrapf(err, "Value for parameter %s is not a valid date: %v", p.Name, v)
			}
			return d, nil
		case time.Time:
			return v_, nil
		default:
			return nil, errors.Errorf("Value for parameter %s is not a valid date: %v", p.Name, v)
		}

	case ParameterTypeFile:
		f, ok := v.(*FileData)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a file: %v", p.Name, v)
		}
		return f, nil

	case ParameterTypeFileList:
		l, ok := cast.CastList2[*FileData, interface{}](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a file list: %v", p.Name, v)
		}
		return l, nil

	case ParameterTypeStringListFromFile:
		fallthrough
	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringList:
		l, err := cast.CastListToStringList(v)
		if err != nil {
			v_, ok := v.([]interface{})
			if !ok {
				return nil, errors.Errorf("Value for parameter %s is not a string list: %v", p.Name, v)
			}

			// convert to string list
			fixedDefault, err := cast.CastListToStringList(v_)
			if err != nil {
				return nil, errors.Errorf("Value for parameter %s is not a string list: %v", p.Name, v)
			}
			return fixedDefault, nil
		}
		return l, nil

	case ParameterTypeIntegerList:
		l, ok := cast.CastInterfaceToIntList[int](v)
		if !ok {
			return nil, errors.Errorf("Default value for parameter %s is not an integer list: %v", p.Name, v)
		}
		return l, nil

	case ParameterTypeFloatList:
		l, ok := cast.CastInterfaceToFloatList[float64](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a float list: %v", p.Name, v)
		}
		return l, nil

	case ParameterTypeChoice:
		if len(p.Choices) == 0 {
			return nil, errors.Errorf("ParameterDefinition %s is a choice parameter but has no choices", p.Name)
		}

		s, err := cast.ToString(v)
		if err != nil {
			return nil, errors.Errorf("Value for parameter %s is not a string: %v", p.Name, v)
		}

		err = p.checkChoiceValidity(s)
		if err != nil {
			return nil, err
		}
		return s, nil

	case ParameterTypeChoiceList:
		if len(p.Choices) == 0 {
			return nil, errors.Errorf("ParameterDefinition %s is a choice parameter but has no choices", p.Name)
		}

		l, err := cast.CastListToStringList(v)
		if err != nil {
			return nil, errors.Errorf("Value for parameter %s is not a string list: %v", p.Name, v)
		}

		for _, choice := range l {
			err := p.checkChoiceValidity(choice)
			if err != nil {
				return nil, err
			}
		}
		return l, nil

	case ParameterTypeKeyValue:
		// TypeKeyValue is map[string]string
		m, ok := v.(map[string]string)
		if ok {
			return m, nil
		}
		m_, ok := v.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a key value map: %v", p.Name, v)
		}

		m__, ok := cast.CastStringMap[string, interface{}](m_)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a key value map: %v", p.Name, v)
		}
		return m__, nil

	default:
		return nil, errors.Errorf("unknown parameter type %s", p.Type)
	}
}

func (p *ParameterDefinition) checkChoiceValidity(choice string) error {
	for _, validChoice := range p.Choices {
		if choice == validChoice {
			return nil
		}
	}
	return errors.Errorf("Value %s is not a valid choice for parameter %s. Valid choices are: %v", choice, p.Name, p.Choices)
}

// LoadParameterDefinitionsFromYAML loads a map of ParameterDefinitions from a YAML file.
// It checks that default values are valid.
// It returns the ParameterDefinitions as a map indexed by name, and as a list.
func LoadParameterDefinitionsFromYAML(yamlContent []byte) *ParameterDefinitions {
	parameters_ := NewParameterDefinitions()

	var err error
	var parameters []*ParameterDefinition

	err = yaml.Unmarshal(yamlContent, &parameters)
	if err != nil {
		panic(errors.Wrap(err, "Failed to unmarshal output flags yaml"))
	}

	for _, p := range parameters {
		_, err := p.CheckParameterDefaultValueValidity()
		if err != nil {
			panic(errors.Wrap(err, "Failed to check parameter default value validity"))
		}
		parameters_.Set(p.Name, p)
	}

	return parameters_
}

// ParameterDefinitions is an ordered map of ParameterDefinition.
type ParameterDefinitions struct {
	*orderedmap.OrderedMap[string, *ParameterDefinition]
}

type ParameterDefinitionsOption func(*ParameterDefinitions)

func WithParameterDefinitions(parameterDefinitions *ParameterDefinitions) ParameterDefinitionsOption {
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

func NewParameterDefinitions(options ...ParameterDefinitionsOption) *ParameterDefinitions {
	ret := &ParameterDefinitions{
		orderedmap.New[string, *ParameterDefinition](),
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

// Clone returns a cloned copy of the ParameterDefinitions.
// The parameter definitions are cloned as well.
func (pds *ParameterDefinitions) Clone() *ParameterDefinitions {
	return NewParameterDefinitions().Merge(pds)
}

// Merge merges the parameter definitions from m into p.
// It clones each parameter definition before adding it to p
// so that updates to p do not affect m.
func (pds *ParameterDefinitions) Merge(m *ParameterDefinitions) *ParameterDefinitions {
	for v := m.Oldest(); v != nil; v = v.Next() {
		pds.Set(v.Key, v.Value.Clone())
	}
	return pds
}

// GetFlags returns a new ParameterDefinitions containing only the flag
// parameters. The parameter definitions are not cloned.
func (pds *ParameterDefinitions) GetFlags() *ParameterDefinitions {
	ret := NewParameterDefinitions()

	for v := pds.Oldest(); v != nil; v = v.Next() {
		if !v.Value.IsArgument {
			ret.Set(v.Key, v.Value)
		}
	}

	return ret
}

func (pds *ParameterDefinitions) ToList() []*ParameterDefinition {
	ret := []*ParameterDefinition{}
	for v := pds.Oldest(); v != nil; v = v.Next() {
		ret = append(ret, v.Value)
	}
	return ret
}

func (pds *ParameterDefinitions) GetDefaultValue(key string, defaultValue interface{}) interface{} {
	v, ok := pds.Get(key)
	if !ok || v.Default == nil {
		return defaultValue
	}
	return *v.Default
}

// GetArguments returns a new ParameterDefinitions containing only the argument
// parameters. The parameter definitions are not cloned.
func (pds *ParameterDefinitions) GetArguments() *ParameterDefinitions {
	ret := NewParameterDefinitions()

	for v := pds.Oldest(); v != nil; v = v.Next() {
		if v.Value.IsArgument {
			ret.Set(v.Key, v.Value)
		}
	}

	return ret
}

// ForEachE calls the given function f on each parameter definition in p.
// If f returns an error, ForEachE stops iterating and returns the error immediately.
func (pds *ParameterDefinitions) ForEachE(f func(definition *ParameterDefinition) error) error {
	for v := pds.Oldest(); v != nil; v = v.Next() {
		err := f(v.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

// ForEach calls the given function f on each parameter definition in p.
func (pds *ParameterDefinitions) ForEach(f func(definition *ParameterDefinition)) {
	for v := pds.Oldest(); v != nil; v = v.Next() {
		f(v.Value)
	}
}

func (pds *ParameterDefinitions) MarshalYAML() (interface{}, error) {
	ret := []*ParameterDefinition{}
	pds.ForEach(func(definition *ParameterDefinition) {
		ret = append(ret, definition)
	})
	return ret, nil
}

func (pds *ParameterDefinitions) UnmarshalYAML(value *yaml.Node) error {
	var parameterDefinitions []*ParameterDefinition
	err := value.Decode(&parameterDefinitions)
	if err != nil {
		return err
	}

	for _, pd_ := range parameterDefinitions {
		pds.Set(pd_.Name, pd_)
	}

	return nil
}

func (p *ParameterDefinition) setReflectValue(v reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	switch p.Type {
	case ParameterTypeString, ParameterTypeChoice, ParameterTypeStringFromFiles, ParameterTypeStringFromFile:
		strVal, ok := value.(string)
		if !ok {
			return errors.Errorf("expected string value for parameter %s, got %T", p.Name, value)
		}
		v.SetString(strVal)

	case ParameterTypeBool:
		boolVal, ok := value.(bool)
		if !ok {
			return errors.Errorf("expected bool value for parameter %s, got %T", p.Name, value)
		}
		v.SetBool(boolVal)

	case ParameterTypeInteger, ParameterTypeFloat:
		return reflect2.SetReflectValue(v, value)

	case ParameterTypeStringList,
		ParameterTypeChoiceList,
		ParameterTypeStringListFromFiles,
		ParameterTypeStringListFromFile:
		list, ok := value.([]string)
		if !ok {
			return errors.Errorf("expected string list for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(list))

	case ParameterTypeDate:
		dateTime, ok := value.(time.Time)
		if !ok {
			return errors.Errorf("expected time.Time value for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(dateTime))

	case ParameterTypeIntegerList, ParameterTypeFloatList:
		return reflect2.SetReflectValue(v, value)

	case ParameterTypeFile:
		return reflect2.SetReflectValue(v, value)

	case ParameterTypeFileList:
		list, ok := value.([]*FileData)
		if !ok {
			return errors.Errorf("expected list of files for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(list))

	case ParameterTypeObjectListFromFiles, ParameterTypeObjectListFromFile:
		list, ok := value.([]interface{})
		if !ok {
			return errors.Errorf("expected list of maps for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(list))

	case ParameterTypeObjectFromFile:
		mapVal, ok := value.(map[string]interface{})
		if !ok {
			return errors.Errorf("expected map for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(mapVal))

	case ParameterTypeKeyValue:
		mapVal, ok := value.(map[string]string)
		if !ok {
			return errors.Errorf("expected map of strings for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(mapVal))

	default:
		return errors.Errorf("unknown parameter type %s", p.Type)
	}
	return nil
}
