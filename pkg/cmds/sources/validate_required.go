package sources

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

// ValidateRequired returns a middleware that validates required fields after all
// earlier source middlewares have merged their values into parsedValues.
func ValidateRequired() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			if err := next(schema_, parsedValues); err != nil {
				return err
			}
			return ValidateRequiredValues(schema_, parsedValues)
		}
	}
}

// ValidateRequiredValues checks required field definitions against final merged
// values. Source middlewares should collect values without enforcing requiredness;
// this helper is the final gate for normal command execution.
func ValidateRequiredValues(schema_ *schema.Schema, parsedValues *values.Values) error {
	missing := []string{}

	err := schema_.ForEachE(func(_ string, section schema.Section) error {
		sectionValues, sectionOK := parsedValues.Get(section.GetSlug())
		return section.GetDefinitions().ForEachE(func(def *fields.Definition) error {
			if !def.Required {
				return nil
			}

			if !sectionOK {
				missing = append(missing, requiredFieldName(section, def))
				return nil
			}

			fieldValue, ok := sectionValues.Fields.Get(def.Name)
			if !ok || isRequiredValueEmpty(def, fieldValue.Value) {
				missing = append(missing, requiredFieldName(section, def))
			}
			return nil
		})
	})
	if err != nil {
		return err
	}

	if len(missing) > 0 {
		return fmt.Errorf("required field(s) missing: %s", strings.Join(missing, ", "))
	}

	return nil
}

func requiredFieldName(section schema.Section, def *fields.Definition) string {
	if section.GetSlug() == schema.DefaultSlug {
		return def.Name
	}
	return section.GetSlug() + "." + def.Name
}

func isRequiredValueEmpty(def *fields.Definition, value interface{}) bool {
	if value == nil {
		return true
	}

	// For human-entered scalar text values, an empty or whitespace-only string is
	// not a meaningful required value. Numeric zero and boolean false are valid
	// provided values and are handled by the generic fallback below.
	switch def.Type {
	case fields.TypeString,
		fields.TypeSecret,
		fields.TypeFile,
		fields.TypeStringFromFile,
		fields.TypeStringFromFiles,
		fields.TypeChoice:
		if s, ok := value.(string); ok {
			return strings.TrimSpace(s) == ""
		}
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}
