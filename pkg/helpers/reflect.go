package helpers

import (
	"fmt"
	"reflect"
)

func SetReflectValue(value reflect.Value, v interface{}) error {
	if !value.IsValid() {
		return fmt.Errorf("cannot set invalid reflect.Value")
	}
	if !value.CanSet() {
		return fmt.Errorf("cannot set unexported reflect.Value")
	}

	switch value.Kind() {
	case reflect.String:
		if s, ok := v.(string); ok {
			value.SetString(s)
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, ok := v.(int64); ok {
			value.SetInt(i)
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
	case reflect.Float32, reflect.Float64:
		if f, ok := v.(float64); ok {
			value.SetFloat(f)
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
	case reflect.Bool:
		if b, ok := v.(bool); ok {
			value.SetBool(b)
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
	case reflect.Slice:
		if s, ok := v.([]interface{}); ok {
			value.Set(reflect.ValueOf(s))
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
	case reflect.Map:
		if m, ok := v.(map[string]interface{}); ok {
			value.Set(reflect.ValueOf(m))
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
	default:
		return fmt.Errorf("unsupported reflect.Value type %s", value.Kind())
	}
}
