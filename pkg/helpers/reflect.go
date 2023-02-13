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
		return fmt.Errorf("cannot set reflect.Value")
	}

	//exhaustive:ignore
	switch value.Kind() {
	case reflect.String:
		if s, ok := v.(string); ok {
			value.SetString(s)
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if i, ok := v.(uint64); ok {
			value.SetUint(i)
			return nil
		}
		if i, ok := v.(uint); ok {
			value.SetUint(uint64(i))
			return nil
		}
		if i, ok := v.(uint8); ok {
			value.SetUint(uint64(i))
			return nil
		}
		if i, ok := v.(uint16); ok {
			value.SetUint(uint64(i))
			return nil
		}
		if i, ok := v.(uint32); ok {
			value.SetUint(uint64(i))
			return nil
		}
		if i, ok := v.(int64); ok {
			value.SetUint(uint64(i))
			return nil
		}
		if i, ok := v.(int); ok {
			value.SetUint(uint64(i))
			return nil
		}
		if i, ok := v.(int8); ok {
			value.SetUint(uint64(i))
			return nil
		}
		if i, ok := v.(int16); ok {
			value.SetUint(uint64(i))
			return nil
		}
		if i, ok := v.(int32); ok {
			value.SetUint(uint64(i))
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, ok := v.(int64); ok {
			value.SetInt(i)
			return nil
		}
		if i, ok := v.(int); ok {
			value.SetInt(int64(i))
			return nil
		}
		if i, ok := v.(int8); ok {
			value.SetInt(int64(i))
			return nil
		}
		if i, ok := v.(int16); ok {
			value.SetInt(int64(i))
			return nil
		}
		if i, ok := v.(int32); ok {
			value.SetInt(int64(i))
			return nil
		}

		if i, ok := v.(uint64); ok {
			value.SetInt(int64(i))
			return nil
		}
		if i, ok := v.(uint); ok {
			value.SetInt(int64(i))
			return nil
		}
		if i, ok := v.(uint8); ok {
			value.SetInt(int64(i))
			return nil
		}
		if i, ok := v.(uint16); ok {
			value.SetInt(int64(i))
			return nil
		}
		if i, ok := v.(uint32); ok {
			value.SetInt(int64(i))
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
		//exhaustive:ignore
		switch value.Type().Elem().Kind() {
		case reflect.String:
			if s, ok := v.([]string); ok {
				value.Set(reflect.ValueOf(s))
				return nil
			}
			if s, ok := v.([]interface{}); ok {
				v2_, ok := CastList[string, interface{}](s)
				if !ok {
					return fmt.Errorf("cannot cast %T to []string", v)
				}
				value.Set(reflect.ValueOf(v2_))
				return nil
			}

			return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if s, ok := v.([]int64); ok {
				value.Set(reflect.ValueOf(s))
				return nil
			}

			if s, ok := v.([]interface{}); ok {
				v2_, ok := CastList[int64, interface{}](s)
				if !ok {
					return fmt.Errorf("cannot cast %T to []int64", v)
				}
				value.Set(reflect.ValueOf(v2_))
				return nil
			}

			return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)

		case reflect.Float32, reflect.Float64:
			if s, ok := v.([]float64); ok {
				value.Set(reflect.ValueOf(s))
				return nil
			}

			if s, ok := v.([]interface{}); ok {
				v2_, ok := CastList[float64, interface{}](s)
				if !ok {
					return fmt.Errorf("cannot cast %T to []float64", v)
				}
				value.Set(reflect.ValueOf(v2_))
				return nil
			}

			return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)

		default:
			return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
		}

	case reflect.Map:
		//exhaustive:ignore
		switch value.Type().Elem().Kind() {
		case reflect.String:
			if m, ok := v.(map[string]string); ok {
				value.Set(reflect.ValueOf(m))
				return nil
			}

			if m, ok := v.(map[string]interface{}); ok {
				v2_, ok := CastStringMap[string, interface{}](m)
				if !ok {
					return fmt.Errorf("cannot cast %T to map[string]string", v)
				}
				value.Set(reflect.ValueOf(v2_))
				return nil
			}

		case reflect.Interface:
			if m, ok := v.(map[string]interface{}); ok {
				value.Set(reflect.ValueOf(m))
				return nil
			}

		default:
			return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)

	default:
		return fmt.Errorf("unsupported reflect.Value type %s", value.Kind())
	}
}
