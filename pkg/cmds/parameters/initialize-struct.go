package parameters

import (
	"encoding/json"
	reflect2 "github.com/go-go-golems/glazed/pkg/helpers/reflect"
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
// Struct fields that are pointers to other structs are handled recursively.
//
// s should be a pointer to the struct to initialize.
//
// ps is the ParsedParameters map to lookup parameter values from.
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

		if dst.Type() == reflect.TypeOf(parameter.Value) {
			dst.Set(reflect.ValueOf(parameter.Value))
			continue
		}

		wasPointer := false
		if kind == reflect.Ptr {
			elem := field.Type.Elem()
			if dst.IsNil() {
				value := reflect.New(elem)
				dst.Set(value)
				dst = value
				kind = elem.Kind()
			}
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
			case []byte:
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

		err = reflect2.SetReflectValue(dst, parameter.Value)
		if err != nil {
			return errors.Wrapf(err, "failed to set value for %s", options.Name)
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
