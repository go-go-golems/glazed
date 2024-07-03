package parameters

import (
	"encoding/json"
	reflect2 "github.com/go-go-golems/glazed/pkg/helpers/reflect"
	"github.com/pkg/errors"
	"path/filepath"
	"reflect"
	"strings"
)

type tagOptions struct {
	Name       string
	FromJson   bool
	IsWildcard bool
}

// parsedTagOptions processes a structure tag string into a struct of type tagOptions.
//
// - Split tag on ,
// - Retrieves tag name
// - Checks if "from_json" is present
// - Returns an error if the tag is empty or nil otherwise.
//
// Returns: parsed *tagOptions
// Error: If the tag is empty.
func parsedTagOptions(tag string) (*tagOptions, error) {
	options := strings.Split(tag, ",")
	if len(options) == 0 {
		return nil, errors.Errorf("invalid empty glazed.parameter tag")
	}
	name := options[0]
	options = options[1:]

	// if the name contains a *, then we are using a glob wildcard
	isWildcard := strings.Contains(name, "*")

	fromJson := false
	for _, o := range options {
		if o == "from_json" {
			fromJson = true
			break
		}
	}

	return &tagOptions{
		Name:       name,
		FromJson:   fromJson,
		IsWildcard: isWildcard,
	}, nil
}

// InitializeStruct initializes a struct from a ParsedParameters map.
//
// It iterates through the struct fields looking for those tagged with
// "glazed.parameter". For each tagged field, it will lookup the corresponding
// parameter value in the ParsedParameters map and set the field's value.
//
// If the tag open `from_json` is appended to `glazed.parameter` and the parameter
// value is a string, bytes, rawMessage or FileData, the value is parsed from json.
//
// If the tag contains a wildcard, the function will match parameter names against the
// wildcard pattern and store the matches in a map in the destination field. The map can
// be of any type, as long as it is a map of strings. The same logic as for normal fields
// will be applied to the map entries.
//
// Struct fields that are pointers to other structs are handled recursively.
//
// Struct fields that are pointers will be dereferenced. If the pointer is nil, a new value
// will be allocated and set.
//
// s should be a pointer to the struct to initialize.
//
// ps is the ParsedParameters map to lookup parameter values from.
//
// Example struct:
//
//	type CreateIndexSettings struct {
//		Index               string               `glazed.parameter:"index"`
//		Settings            *IndexSettings       `glazed.parameter:"settings,from_json"`
//		Mappings            *parameters.FileData `glazed.parameter:"mappings"`
//		Aliases             *map[string]Alias    `glazed.parameter:"aliases,from_json"`
//		WaitForActiveShards string               `glazed.parameter:"wait_for_active_shards"`
//		ApiKeys             map[string]string    `glazed.parameter:"*_api_key"`
//	}
//
// Corresponding ParameterDefinitions:
//
//	parameters.NewParameterDefinition(
//		"index",
//		parameters.ParameterTypeString,
//		parameters.WithHelp("Name of the index to create"),
//		parameters.WithRequired(true),
//	),
//	parameters.NewParameterDefinition(
//		"settings",
//		parameters.ParameterTypeFile,
//		parameters.WithHelp("JSON file containing index settings"),
//	),
//	parameters.NewParameterDefinition(
//		"mappings",
//		parameters.ParameterTypeFile,
//		parameters.WithHelp("JSON file containing index mappings"),
//	),
//	parameters.NewParameterDefinition(
//		"aliases",
//		parameters.ParameterTypeFile,
//		parameters.WithHelp("JSON file containing index aliases"),
//	),
//	parameters.NewParameterDefinition(
//		"wait_for_active_shards",
//		parameters.ParameterTypeString,
//		parameters.WithHelp("Set the number of active shards to wait for before the operation returns."),
//	),
//	parameters.NewParameterDefinition(
//		"openai_api_key",
//		parameters.ParameterTypeString,
//		parameters.WithHelp("OpenAI API key"),
//	),
//	parameters.NewParameterDefinition(
//		"google_api_key",
//		parameters.ParameterTypeString,
//		parameters.WithHelp("Google API key"),
//	),
//
// Returns an error if:
// - s is not a pointer to a struct
// - A tagged field does not have a matching parameter value in ps
// - Failed to set the value of a field
func (p *ParsedParameters) InitializeStruct(s interface{}) error {
	if s == nil {
		return errors.Errorf("Can't initialize nil struct")
	}
	of := reflect.TypeOf(s)
	if of.Kind() != reflect.Ptr {
		return errors.Errorf("s is not a pointer")
	}
	if of.Elem().Kind() != reflect.Struct {
		return errors.Errorf("s is not a pointer to a struct")
	}
	st := of.Elem()
	v := reflect.ValueOf(s).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag, ok := field.Tag.Lookup("glazed.parameter")
		if !ok {
			continue
		}
		options, err := parsedTagOptions(tag)
		if err != nil {
			return errors.Wrapf(err, "failed to parse glazed.parameter tag for field %s", field.Name)
		}

		if options.IsWildcard {
			dst := v.FieldByName(field.Name)
			if dst.Kind() != reflect.Map {
				return errors.Errorf("wildcard parameters require a map field, field %s is not a map", field.Name)
			}
			if err := p.setWildcardValues(dst, options.Name, options.FromJson); err != nil {
				return errors.Wrapf(err, "failed to set wildcard values for %s", options.Name)
			}
		} else {
			parameter, ok := p.Get(options.Name)
			if !ok {
				continue
			}
			dst := v.FieldByName(field.Name)
			if err := p.setTargetValue(dst, parameter.Value, options.FromJson); err != nil {
				return errors.Wrapf(err, "failed to set value for %s", options.Name)
			}
		}
	}
	return nil
}

// setWildcardValues matches parameter names from ParsedParameters against a supplied pattern using the
// filepath.Match method.
//
// The pattern may consist of literal characters,
// character ranges enclosed in square brackets, and the wildcards * and ?.
// The wildcard * matches zero or more characters, and the wildcard ? matches a single character.
//
// The function sets the matches in a map in the destination reflect.Value,
// using the parameter name as the key and the parameter value as the map value.
//
// This way, you can match multiple flags at once and stored them in a map.
//
// Parameters:
//   - dst: A reflect.Value acting as the container for storing the matched keys and their parameter values.
//   - pattern: String pattern used for matching parameter names. It can include literal characters,
//     character ranges enclosed in brackets and wildcards.
func (p *ParsedParameters) setWildcardValues(dst reflect.Value, pattern string, fromJson bool) error {
	// Get the type of elements in dst
	var elemType reflect.Type
	if dst.Kind() != reflect.Map {
		return errors.New("destination is not a map")
	}
	elemType = dst.Type().Elem()
	keyType := dst.Type().Key()

	if dst.IsNil() {
		dst.Set(reflect.MakeMapWithSize(reflect.MapOf(keyType, elemType), 0))
	}

	err := p.ForEachE(func(paramName string, parameter *ParsedParameter) error {
		if matched, _ := filepath.Match(pattern, paramName); matched {
			// Check if the type of parameter.Value is compatible with the map's value type
			if reflect.TypeOf(parameter.Value) != elemType {
				return errors.Errorf("type mismatch: expected type %s, got %s", elemType, reflect.TypeOf(parameter.Value))
			}

			paramValue := reflect.ValueOf(parameter.Value)

			// Check if the type of parameter.Value is directly assignable to the map's value type
			if !paramValue.Type().AssignableTo(elemType) {
				// Check if the type of parameter.Value can be converted to the map's value type
				if paramValue.Type().ConvertibleTo(elemType) {
					paramValue = paramValue.Convert(elemType)
				} else {
					return errors.Errorf("type mismatch: cannot assign type %s to type %s", paramValue.Type(), elemType)
				}
			}

			// Set the value in the map
			keyValue := reflect.ValueOf(paramName)
			if !keyValue.Type().AssignableTo(keyType) {
				if keyValue.Type().ConvertibleTo(keyType) {
					keyValue = keyValue.Convert(keyType)
				} else {
					return errors.Errorf("type mismatch: cannot assign type %s to type %s", keyValue.Type(), keyType)
				}
			}

			dst.SetMapIndex(keyValue, paramValue)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *ParsedParameters) setTargetValue(dst reflect.Value, value interface{}, fromJson bool) error {
	valueType := reflect.TypeOf(value)

	// Handle pointer destination
	wasPointer := false
	if dst.Kind() == reflect.Ptr {
		wasPointer = true
		if dst.IsNil() {
			newValue := reflect.New(dst.Type().Elem())
			dst.Set(newValue)
		}

		// The destination is now the value pointed to by the pointer
		dst = dst.Elem()
	}

	// Handle JSON unmarshalling
	if fromJson {
		if !wasPointer {
			return errors.Errorf("from_json tag can only be used on pointer fields")
		}
		jsonData, ok := value.([]byte)
		if !ok {
			str, ok := value.(string)
			if !ok {
				return errors.Errorf("expected string or []byte for JSON unmarshalling, got %T", value)
			}
			jsonData = []byte(str)
		}
		if err := json.Unmarshal(jsonData, dst.Addr().Interface()); err != nil {
			return errors.Wrapf(err, "failed to unmarshal JSON")
		}
		return nil
	}

	dstTypeS := dst.Type().String()
	valueTypeS := valueType.String()
	_ = dstTypeS
	_ = valueTypeS

	// Direct assignment if types are compatible
	if dst.Type() == valueType {
		dst.Set(reflect.ValueOf(value))
		return nil
	}

	// if valueType is a pointer to a value of dst.Type(), we can assign it directly
	if valueType.Kind() == reflect.Ptr && valueType.Elem() == dst.Type() {
		dst.Set(reflect.ValueOf(value).Elem())
		return nil
	}

	// Recursive struct initialization
	if dst.Kind() == reflect.Struct {
		return p.InitializeStruct(dst.Addr().Interface())
	}

	err := reflect2.SetReflectValue(dst, value)
	if err != nil {
		return errors.Wrapf(err, "failed to set value for field")
	}

	return nil
}

// StructToDataMap transforms a struct into a map[string]interface{} based on the `glazed.parameter` annotations.
//
// If a struct field is annotated with `glazed.parameter:"<pattern>*"` (contains a wildcard `*`), the field
// is expected to be a map. The function will match the map keys against the wildcard pattern and include
// the matching key-value pairs in the resulting data map.
//
// Returns an error if:
// - Parsing the `glazed.parameter` tag fails for any field.
// - A field annotated with a wildcard is not a map.
func StructToDataMap(s interface{}) (map[string]interface{}, error) {
	if s == nil {
		return nil, errors.New("cannot convert nil struct to data map")
	}

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, errors.New("input must be a struct or a pointer to a struct")
	}

	dataMap := make(map[string]interface{})

	structType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := structType.Field(i)
		tag, ok := field.Tag.Lookup("glazed.parameter")
		if !ok {
			continue
		}

		options, err := parsedTagOptions(tag)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse glazed.parameter tag for field %s", field.Name)
		}

		fieldValue := v.Field(i)
		if options.IsWildcard {
			if fieldValue.Kind() != reflect.Map {
				return nil, errors.Errorf("wildcard parameters require a map field, field %s is not a map", field.Name)
			}
			if err := setWildcardDataMapValues(dataMap, fieldValue, options.Name); err != nil {
				return nil, errors.Wrapf(err, "failed to set wildcard values for %s", options.Name)
			}
		} else {
			if err := setDataMapValue(dataMap, options.Name, fieldValue.Interface(), options.FromJson); err != nil {
				return nil, errors.Wrapf(err, "failed to set value for %s", options.Name)
			}
		}
	}

	return dataMap, nil
}

func setWildcardDataMapValues(dataMap map[string]interface{}, fieldValue reflect.Value, pattern string) error {
	iter := fieldValue.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		if matched, _ := filepath.Match(pattern, key); matched {
			value := iter.Value().Interface()
			dataMap[key] = value
		}
	}
	return nil
}

func setDataMapValue(dataMap map[string]interface{}, key string, value interface{}, fromJson bool) error {
	dataMap[key] = value
	return nil
}
