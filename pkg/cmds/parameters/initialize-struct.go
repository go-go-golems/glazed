package parameters

import (
	"encoding/json"
	reflect_helpers "github.com/go-go-golems/glazed/pkg/helpers/reflect"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

// InitializeStruct initializes a struct from a ParsedParameters map.
//
// It iterates through the struct fields looking for those tagged with
// "glazed.parameter". For each tagged field, it will lookup the corresponding
// parameter value in the ParsedParameters map and set the field's value.
//
// If the tag open `from_json` is appended to `glazed.parameter` and the parameter
// value is a string, bytes, rawMessage or FileData, the value is parsed from json.
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
//
// Returns an error if:
// - s is not a pointer to a struct
// - A tagged field does not have a matching parameter value in ps
// - Failed to set the value of a field
func (p *ParsedParameters) InitializeStruct(s interface{}) error {
	if s == nil {
		return errors.Errorf("Can't initialize nil struct")
	}

	// check that s is indeed a pointer to a struct
	of := reflect.TypeOf(s)
	if of.Kind() != reflect.Ptr {
		return errors.Errorf("s is not a pointer")
	}
	if of.Elem().Kind() != reflect.Struct {
		return errors.Errorf("s is not a pointer to a struct")
	}
	st := of.Elem()

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

		parameter, ok := p.Get(options.Name)
		if !ok {
			continue
		}
		dst := reflect.ValueOf(s).Elem().FieldByName(field.Name)
		kind := field.Type.Kind()

		if dst.Type() == reflect.TypeOf(parameter.Value) && !options.FromJson {
			dst.Set(reflect.ValueOf(parameter.Value))
			continue
		}

		wasPointer := false
		if kind == reflect.Ptr {
			elem := field.Type.Elem()
			kind = elem.Kind()
			if dst.IsNil() {
				kind = elem.Kind()
				value := reflect.New(elem)
				dst.Set(value)
			}
			dst = dst.Elem()
			wasPointer = true
		}

		if options.FromJson {
			if !wasPointer {
				return errors.Errorf("from_json tag can only be used on pointer fields")
			}

			switch parameter.Value.(type) {
			case *FileData:
				err := json.Unmarshal([]byte(parameter.Value.(*FileData).Content), dst.Interface())
				if err != nil {
					return errors.Wrapf(err, "failed to unmarshal json for %s", options.Name)
				}
			case string:
				err := json.Unmarshal([]byte(parameter.Value.(string)), dst.Interface())
				if err != nil {
					return errors.Wrapf(err, "failed to unmarshal json for %s", options.Name)
				}
			case []byte, json.RawMessage:
				err := json.Unmarshal(parameter.Value.([]byte), dst.Interface())
				if err != nil {
					return errors.Wrapf(err, "failed to unmarshal json for %s", options.Name)
				}
			default:
				return errors.Errorf("unsupported type for %s", options.Name)
			}

			continue
		}

		if kind == reflect.Struct {
			err := p.InitializeStruct(dst.Interface())
			if err != nil {
				return errors.Wrapf(err, "failed to initialize struct for %s", options.Name)
			}
			continue
		}

		kind2 := dst.Kind()
		_ = kind2
		err = reflect_helpers.SetReflectValue(dst, parameter.Value)
		if err != nil {
			return errors.Wrapf(err, "failed to set value %v for %s from value %v", options.Name, dst, parameter.Value)
		}
	}

	return nil
}

type tagOptions struct {
	Name     string
	FromJson bool
}

func parsedTagOptions(tag string) (*tagOptions, error) {
	options := strings.Split(tag, ",")
	if len(options) == 0 {
		return nil, errors.Errorf("invalid empty glazed.parameter tag")
	}
	name := options[0]
	options = options[1:]

	fromJson := false
	for _, o := range options {
		if o == "from_json" {
			fromJson = true
			break
		}
	}

	return &tagOptions{
		Name:     name,
		FromJson: fromJson,
	}, nil
}
