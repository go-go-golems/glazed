package fields

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

// Definition is a declarative way of describing a command line parameter.
// A Definition can be either a Flag or an Argument.
// Along with metadata (Name, Help) that is useful for help,
// it also specifies a Type, a Default value and if it is Required.
type Definition struct {
	Name       string       `yaml:"name"`
	ShortFlag  string       `yaml:"shortFlag,omitempty"`
	Type       Type         `yaml:"type"`
	Help       string       `yaml:"help,omitempty"`
	Default    *interface{} `yaml:"default,omitempty"`
	Choices    []string     `yaml:"choices,omitempty"`
	Required   bool         `yaml:"required,omitempty"`
	IsArgument bool         `yaml:"-"`
}

type Option func(*Definition)

func WithHelp(help string) Option {
	return func(p *Definition) {
		p.Help = help
	}
}

func WithShortFlag(shortFlag string) Option {
	return func(p *Definition) {
		p.ShortFlag = shortFlag
	}
}

func WithDefault(defaultValue interface{}) Option {
	return func(p *Definition) {
		p.Default = &defaultValue
	}
}

func WithChoices(choices ...string) Option {
	return func(p *Definition) {
		p.Choices = choices
	}
}

func WithRequired(required bool) Option {
	return func(p *Definition) {
		p.Required = required
	}
}

func WithIsArgument(isArgument bool) Option {
	return func(p *Definition) {
		p.IsArgument = isArgument
	}
}

func New(
	name string,
	parameterType Type,
	options ...Option,
) *Definition {
	ret := &Definition{
		Name: name,
		Type: parameterType,
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (p *Definition) String() string {
	return fmt.Sprintf("{Parameter: %s - %s}", p.Name, p.Type)
}

func (p *Definition) Clone() *Definition {
	return &Definition{
		Name:       p.Name,
		ShortFlag:  p.ShortFlag,
		Type:       p.Type,
		Help:       p.Help,
		Default:    p.Default,
		Choices:    p.Choices,
		Required:   p.Required,
		IsArgument: p.IsArgument,
	}
}

func (p *Definition) IsEqualToDefault(i interface{}) bool {
	if p.Default == nil {
		return false
	}
	return reflect.DeepEqual(*p.Default, i)
}

// SetDefaultFromValue sets the Default field of the Definition
// to the provided value. It handles nil values and dereferencing pointers.
// The value is type checked before being set.
func (p *Definition) SetDefaultFromValue(value reflect.Value) error {
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

// SetValueFromDefault assigns the default value of the Definition to the given value.
// If the Default value is nil, the value is set to the zero value of the type.
func (p *Definition) SetValueFromDefault(value reflect.Value) error {
	if !value.CanSet() {
		return errors.Errorf("cannot set value of %s", p.Name)
	}

	if p.Default != nil {
		return p.SetValueFromInterface(value, *p.Default)
	}
	return p.InitializeValueToEmptyValue(value)
}

// InitializeValueToEmptyValue initializes the given value to the empty value of the type of the parameter.
func (p *Definition) InitializeValueToEmptyValue(value reflect.Value) error {
	switch p.Type {
	case TypeString, TypeSecret, TypeChoice, TypeStringFromFiles, TypeStringFromFile:
		value.SetString("")
	case TypeBool:
		value.SetBool(false)
	case TypeInteger, TypeFloat:
		return reflect2.SetReflectValue(value, 0)
	case TypeStringList,
		TypeChoiceList,
		TypeStringListFromFiles,
		TypeStringListFromFile:
		return reflect2.SetReflectValue(value, []string{})
	case TypeDate:
		value.Set(reflect.ValueOf(time.Time{}))
	case TypeIntegerList:
		return reflect2.SetReflectValue(value, []int64{})
	case TypeFloatList:
		return reflect2.SetReflectValue(value, []float64{})
	case TypeObjectListFromFiles, TypeObjectListFromFile:
		value.Set(reflect.ValueOf([]map[string]interface{}{}))
	case TypeObjectFromFile:
		value.Set(reflect.ValueOf(map[string]interface{}{}))
	case TypeKeyValue:
		value.Set(reflect.ValueOf(map[string]string{}))
	case TypeFile:
		value.Set(reflect.ValueOf(nil))
	case TypeFileList:
		value.Set(reflect.ValueOf([]*FileData{}))
	default:
		return errors.Errorf("unknown parameter type %s", p.Type)
	}
	return nil
}

// SetValueFromInterface assigns the value v to the given reflect.Value based on the
// Definition's type. It handles type checking and conversion for the
// various supported parameter types.
func (p *Definition) SetValueFromInterface(value reflect.Value, v interface{}) error {
	_, err := p.CheckValueValidity(v)
	if err != nil {
		return err
	}

	switch p.Type {
	case TypeString, TypeSecret, TypeChoice, TypeStringFromFiles, TypeStringFromFile:
		strVal, ok := v.(string)
		if !ok {
			return errors.Errorf("expected string value for parameter %s, got %T", p.Name, v)
		}
		value.SetString(strVal)

	case TypeBool:
		boolVal, ok := v.(bool)
		if !ok {
			return errors.Errorf("expected bool value for parameter %s, got %T", p.Name, v)
		}
		value.SetBool(boolVal)

	case TypeInteger, TypeFloat:
		return reflect2.SetReflectValue(value, v)

	case TypeStringList,
		TypeChoiceList,
		TypeStringListFromFiles,
		TypeStringListFromFile:
		list, err := cast.CastListToStringList(v)
		if err != nil {
			return errors.Errorf("expected string list for parameter %s, got %T", p.Name, v)
		}
		return reflect2.SetReflectValue(value, list)

	case TypeDate:
		strVal, ok := v.(string)
		if !ok {
			return errors.Errorf("expected string value for parameter %s, got %T", p.Name, v)
		}
		dateTime, err := ParseDate(strVal)
		if err != nil {
			return errors.Wrapf(err, "error parsing value for parameter %s", p.Name)
		}
		value.Set(reflect.ValueOf(dateTime))

	case TypeIntegerList, TypeFloatList:
		return reflect2.SetReflectValue(value, v)

	case TypeFile:
		return reflect2.SetReflectValue(value, v)

	case TypeFileList:
		list, ok := cast.CastList2[*FileData, interface{}](v)
		if !ok {
			return errors.Errorf("expected list of files for parameter %s, got %T", p.Name, v)
		}
		value.Set(reflect.ValueOf(list))

	case TypeObjectListFromFiles, TypeObjectListFromFile:
		list, ok := cast.CastList2[map[string]interface{}, interface{}](v)
		if !ok {
			return errors.Errorf("expected list of maps for parameter %s, got %T", p.Name, v)
		}
		value.Set(reflect.ValueOf(list))

	case TypeObjectFromFile:
		mapVal, ok := v.(map[string]interface{})
		if !ok {
			return errors.Errorf("expected map for parameter %s, got %T", p.Name, v)
		}
		value.Set(reflect.ValueOf(mapVal))

	case TypeKeyValue:
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
func (pds *Definitions) ParsedParametersFromDefaults() (*ParsedParameters, error) {
	ret := NewParsedParameters()
	err := pds.ForEachE(func(definition *Definition) error {
		if definition.Default == nil {
			return nil
		}
		err := ret.UpdateValue(definition.Name, definition, *definition.Default,
			WithSource("defaults"),
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
// Each field in the struct annotated with tag `glazed` will be set to the default value of
// the corresponding `Definition`. If no `Definition` is found for a field, an error
// is returned.
func (pds *Definitions) InitializeStructFromDefaults(s interface{}) error {
	parsedParameters, err := pds.ParsedParametersFromDefaults()
	if err != nil {
		return err
	}
	return parsedParameters.InitializeStruct(s)
}

// InitializeDefaultsFromStruct initializes the parameters definitions from a struct.
// Each field in the struct annotated with tag `glazed` will be used to set
// the default value of the corresponding definition in `parameterDefinitions`.
// If no `Definition` is found for a field, an error is returned.
// This is the inverse of InitializeStructFromDefaults.
func (pds *Definitions) InitializeDefaultsFromStruct(
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
		tag, ok := field.Tag.Lookup("glazed")
		if !ok {
			continue
		}

		tagOptions, err := parsedTagOptions(tag)
		if err != nil {
			return err
		}

		if tagOptions.IsWildcard {
			err = pds.ForEachE(func(pd *Definition) error {
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
			return errors.Errorf("unknown parameter %s when initializing defaults from struct", tag)
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		err = parameter.SetDefaultFromValue(value)
		if err != nil {
			return errors.Wrapf(err, "failed to set default value for %s when initializing defaults from struct", tag)
		}
	}

	return nil
}

func (pds *Definitions) InitializeDefaultsFromMap(
	ps map[string]interface{},
) error {
	for k, v := range ps {
		parameter, ok := pds.Get(k)
		if !ok {
			return errors.Errorf("unknown parameter when initializing defaults from map: %s", k)
		}
		err := parameter.SetDefaultFromValue(reflect.ValueOf(v))
		if err != nil {
			return errors.Wrapf(err, "failed to set default value for %s from map", k)
		}
	}

	return nil
}

// CheckParameterDefaultValueValidity checks if the Definition's Default is valid.
// This is used when validating loading from a YAML file or setting up cobra flag definitions.
func (p *Definition) CheckParameterDefaultValueValidity() (interface{}, error) {
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

// CheckValueValidity checks if the given value is valid for the Definition, and returns the value cast to the correct type.
func (p *Definition) CheckValueValidity(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	v = reflect2.StripInterfaceValue(v)

	switch p.Type {
	case TypeStringFromFile:
		fallthrough
	case TypeStringFromFiles:
		fallthrough
	case TypeString:
		fallthrough
	case TypeSecret:
		s, err := cast.ToString(v)
		if err != nil {
			return nil, errors.Errorf("Value for parameter %s is not a string: %v", p.Name, v)
		}
		return s, nil

	case TypeObjectListFromFile:
		fallthrough
	case TypeObjectListFromFiles:
		l, ok := cast.CastList2[map[string]interface{}, interface{}](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s (type %T) is not a list of objects: %v", p.Name, v, v)
		}
		return l, nil

	case TypeObjectFromFile:
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not an object: %v", p.Name, v)
		}
		return m, nil

	case TypeInteger:
		i, ok := cast.CastNumberInterfaceToInt[int](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not an integer: %v", p.Name, v)
		}
		return i, nil

	case TypeFloat:
		f, ok := cast.CastNumberInterfaceToFloat[float64](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a float: %v", p.Name, v)
		}
		return f, nil

	case TypeBool:
		b, ok := v.(bool)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a bool: %v", p.Name, v)
		}
		return b, nil

	case TypeDate:
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

	case TypeFile:
		f, ok := v.(*FileData)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a file (got type %T): %v", p.Name, v, v)
		}
		return f, nil

	case TypeFileList:
		l, ok := cast.CastList2[*FileData, interface{}](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a file list: %v", p.Name, v)
		}
		return l, nil

	case TypeStringListFromFile:
		fallthrough
	case TypeStringListFromFiles:
		fallthrough
	case TypeStringList:
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

	case TypeIntegerList:
		l, ok := cast.CastInterfaceToIntList[int](v)
		if !ok {
			return nil, errors.Errorf("Default value for parameter %s is not an integer list: %v", p.Name, v)
		}
		return l, nil

	case TypeFloatList:
		l, ok := cast.CastInterfaceToFloatList[float64](v)
		if !ok {
			return nil, errors.Errorf("Value for parameter %s is not a float list: %v", p.Name, v)
		}
		return l, nil

	case TypeChoice:
		if len(p.Choices) == 0 {
			return nil, errors.Errorf("Definition %s is a choice parameter but has no choices", p.Name)
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

	case TypeChoiceList:
		if len(p.Choices) == 0 {
			return nil, errors.Errorf("Definition %s is a choice parameter but has no choices", p.Name)
		}

		l, err := cast.CastListToStringList(v)
		if err != nil {
			return nil, errors.Errorf("Value for parameter %s is not a choice list: %v", p.Name, v)
		}

		for _, choice := range l {
			err := p.checkChoiceValidity(choice)
			if err != nil {
				return nil, err
			}
		}
		return l, nil

	case TypeKeyValue:
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

func (p *Definition) checkChoiceValidity(choice string) error {
	for _, validChoice := range p.Choices {
		if choice == validChoice {
			return nil
		}
	}
	return errors.Errorf("Value %s is not a valid choice for parameter %s. Valid choices are: %v", choice, p.Name, p.Choices)
}

// LoadDefinitionsFromYAML loads a map of Definitions from a YAML file.
// It checks that default values are valid.
// It returns the Definitions as a map indexed by name, and as a list.
func LoadDefinitionsFromYAML(yamlContent []byte) *Definitions {
	parameters_ := NewDefinitions()

	var err error
	var parameters []*Definition

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

// Definitions is an ordered map of Definition.
type Definitions struct {
	*orderedmap.OrderedMap[string, *Definition]
}

type DefinitionsOption func(*Definitions)

func WithDefinitions(parameterDefinitions *Definitions) DefinitionsOption {
	return func(p *Definitions) {
		p.Merge(parameterDefinitions)
	}
}

func WithDefinitionList(parameterDefinitions []*Definition) DefinitionsOption {
	return func(p *Definitions) {
		for _, pd := range parameterDefinitions {
			p.Set(pd.Name, pd)
		}
	}
}

func NewDefinitions(options ...DefinitionsOption) *Definitions {
	ret := &Definitions{
		orderedmap.New[string, *Definition](),
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

// Clone returns a cloned copy of the Definitions.
// The parameter definitions are cloned as well.
func (pds *Definitions) Clone() *Definitions {
	return NewDefinitions().Merge(pds)
}

// Merge merges the parameter definitions from m into p.
// It clones each parameter definition before adding it to p
// so that updates to p do not affect m.
func (pds *Definitions) Merge(m *Definitions) *Definitions {
	for v := m.Oldest(); v != nil; v = v.Next() {
		pds.Set(v.Key, v.Value.Clone())
	}
	return pds
}

// GetFlags returns a new Definitions containing only the flag
// fields. The parameter definitions are not cloned.
func (pds *Definitions) GetFlags() *Definitions {
	ret := NewDefinitions()

	for v := pds.Oldest(); v != nil; v = v.Next() {
		if !v.Value.IsArgument {
			ret.Set(v.Key, v.Value)
		}
	}

	return ret
}

func (pds *Definitions) ToList() []*Definition {
	ret := []*Definition{}
	for v := pds.Oldest(); v != nil; v = v.Next() {
		ret = append(ret, v.Value)
	}
	return ret
}

func (pds *Definitions) GetDefaultValue(key string, defaultValue interface{}) interface{} {
	v, ok := pds.Get(key)
	if !ok || v.Default == nil {
		return defaultValue
	}
	return *v.Default
}

// GetArguments returns a new Definitions containing only the argument
// fields. The parameter definitions are not cloned.
func (pds *Definitions) GetArguments() *Definitions {
	ret := NewDefinitions()

	for v := pds.Oldest(); v != nil; v = v.Next() {
		if v.Value.IsArgument {
			ret.Set(v.Key, v.Value)
		}
	}

	return ret
}

// ForEachE calls the given function f on each parameter definition in p.
// If f returns an error, ForEachE stops iterating and returns the error immediately.
func (pds *Definitions) ForEachE(f func(definition *Definition) error) error {
	for v := pds.Oldest(); v != nil; v = v.Next() {
		err := f(v.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

// ForEach calls the given function f on each parameter definition in p.
func (pds *Definitions) ForEach(f func(definition *Definition)) {
	for v := pds.Oldest(); v != nil; v = v.Next() {
		f(v.Value)
	}
}

func (pds *Definitions) MarshalYAML() (interface{}, error) {
	ret := []*Definition{}
	pds.ForEach(func(definition *Definition) {
		ret = append(ret, definition)
	})
	return ret, nil
}

func (pds *Definitions) UnmarshalYAML(value *yaml.Node) error {
	var parameterDefinitions []*Definition
	err := value.Decode(&parameterDefinitions)
	if err != nil {
		return err
	}

	for _, pd_ := range parameterDefinitions {
		pds.Set(pd_.Name, pd_)
	}

	return nil
}

func (p *Definition) setReflectValue(v reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	switch p.Type {
	case TypeString, TypeSecret, TypeChoice, TypeStringFromFiles, TypeStringFromFile:
		strVal, ok := value.(string)
		if !ok {
			return errors.Errorf("expected string value for parameter %s, got %T", p.Name, value)
		}
		v.SetString(strVal)

	case TypeBool:
		boolVal, ok := value.(bool)
		if !ok {
			return errors.Errorf("expected bool value for parameter %s, got %T", p.Name, value)
		}
		v.SetBool(boolVal)

	case TypeInteger, TypeFloat:
		return reflect2.SetReflectValue(v, value)

	case TypeStringList,
		TypeChoiceList,
		TypeStringListFromFiles,
		TypeStringListFromFile:
		list, ok := value.([]string)
		if !ok {
			return errors.Errorf("expected string list for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(list))

	case TypeDate:
		dateTime, ok := value.(time.Time)
		if !ok {
			return errors.Errorf("expected time.Time value for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(dateTime))

	case TypeIntegerList, TypeFloatList:
		return reflect2.SetReflectValue(v, value)

	case TypeFile:
		return reflect2.SetReflectValue(v, value)

	case TypeFileList:
		list, ok := value.([]*FileData)
		if !ok {
			return errors.Errorf("expected list of files for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(list))

	case TypeObjectListFromFiles, TypeObjectListFromFile:
		list, ok := value.([]interface{})
		if !ok {
			return errors.Errorf("expected list of maps for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(list))

	case TypeObjectFromFile:
		mapVal, ok := value.(map[string]interface{})
		if !ok {
			return errors.Errorf("expected map for parameter %s, got %T", p.Name, value)
		}
		v.Set(reflect.ValueOf(mapVal))

	case TypeKeyValue:
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
