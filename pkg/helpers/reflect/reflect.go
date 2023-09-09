package reflect

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"reflect"
	"strconv"
)

func SetReflectValue(value reflect.Value, v interface{}) error {
	if !value.IsValid() {
		return fmt.Errorf("cannot set invalid reflect.Value")
	}
	if !value.CanSet() {
		return fmt.Errorf("cannot set reflect.Value")
	}

	if v == nil {
		value.Set(reflect.Zero(value.Type()))
		return nil
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
		if s, ok := v.(string); ok {
			i, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return err
			}
			value.SetUint(i)
			return nil
		}
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
		if s, ok := v.(string); ok {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			value.SetInt(i)
			return nil
		}
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
		if f, ok := v.(float32); ok {
			value.SetFloat(float64(f))
			return nil
		}
		if i, ok := v.(int64); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(int); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(int8); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(int16); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(int32); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(uint64); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(uint); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(uint8); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(uint16); ok {
			value.SetFloat(float64(i))
			return nil
		}
		if i, ok := v.(uint32); ok {
			value.SetFloat(float64(i))
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
		sliceKind := value.Type().Elem().Kind()
		//exhaustive:ignore
		switch sliceKind {
		case reflect.String:
			if s, ok := v.([]string); ok {
				value.Set(reflect.ValueOf(s))
				return nil
			}
			if s, ok := v.([]interface{}); ok {
				v2_, ok := cast.CastList[string, interface{}](s)
				if !ok {
					return fmt.Errorf("cannot cast %T to []string", v)
				}
				value.Set(reflect.ValueOf(v2_))
				return nil
			}

			return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)

		case reflect.Int:
			return SetIntListReflectValue[int](value, v)

		case reflect.Int8:
			return SetIntListReflectValue[int8](value, v)

		case reflect.Int16:
			return SetIntListReflectValue[int16](value, v)

		case reflect.Int32:
			return SetIntListReflectValue[int32](value, v)

		case reflect.Int64:
			return SetIntListReflectValue[int64](value, v)

		case reflect.Uint:
			return SetIntListReflectValue[uint](value, v)

		case reflect.Uint8:
			return SetIntListReflectValue[uint8](value, v)

		case reflect.Uint16:
			return SetIntListReflectValue[uint16](value, v)

		case reflect.Uint32:
			return SetIntListReflectValue[uint32](value, v)

		case reflect.Uint64:
			return SetIntListReflectValue[uint64](value, v)

		case reflect.Float32:
			return SetFloatListReflectValue[float32](value, v)

		case reflect.Float64:
			return SetFloatListReflectValue[float64](value, v)

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
				v2_, ok := cast.CastStringMap[string, interface{}](m)
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

			if m, ok := v.(map[string]string); ok {
				v2_, ok := cast.CastStringMap[interface{}, string](m)
				if !ok {
					return fmt.Errorf("cannot cast %T to map[string]interface{}", v)
				}
				value.Set(reflect.ValueOf(v2_))
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

func SetIntListReflectValue[To cast.Number](value reflect.Value, v interface{}) error {
	if s, ok := v.([]int64); ok {
		s2, ok := cast.CastToNumberList[To, int64](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]int); ok {
		s2, ok := cast.CastToNumberList[To, int](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]int8); ok {
		s2, ok := cast.CastToNumberList[To, int8](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]int16); ok {
		s2, ok := cast.CastToNumberList[To, int16](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]int32); ok {
		s2, ok := cast.CastToNumberList[To, int32](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint64); ok {
		s2, ok := cast.CastToNumberList[To, uint64](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint); ok {
		s2, ok := cast.CastToNumberList[To, uint](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint8); ok {
		s2, ok := cast.CastToNumberList[To, uint8](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint16); ok {
		s2, ok := cast.CastToNumberList[To, uint16](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint32); ok {
		s2, ok := cast.CastToNumberList[To, uint32](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]interface{}); ok {
		v2_, ok := cast.CastList[int, interface{}](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		v3_, ok := cast.CastToNumberList[To, int](v2_)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(v3_))
		return nil
	}

	return fmt.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
}

func SetFloatListReflectValue[To cast.FloatNumber](value reflect.Value, v interface{}) error {
	if s, ok := v.([]float64); ok {
		s2, ok := cast.CastToNumberList[To, float64](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]float32); ok {
		s2, ok := cast.CastToNumberList[To, float32](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]interface{}); ok {
		v2_, ok := cast.CastInterfaceListToFloatList[To](s)
		if !ok {
			return fmt.Errorf("cannot cast %T to []%T", v, To(0))
		}

		value.Set(reflect.ValueOf(v2_))
		return nil
	}

	return SetIntListReflectValue[To](value, v)
}
