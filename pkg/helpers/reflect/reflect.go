package reflect

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"reflect"
	"strconv"
)

func SetReflectValue(dst reflect.Value, src interface{}) error {
	kind := dst.Kind()
	if !dst.IsValid() {
		return fmt.Errorf("cannot set invalid reflect.Value of type %s", kind)
	}
	if !dst.CanSet() {
		return fmt.Errorf("cannot set reflect.Value of type %s", kind)
	}

	if src == nil {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	//exhaustive:ignore
	switch kind {
	case reflect.String:
		if s, ok := src.(string); ok {
			dst.SetString(s)
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if s, ok := src.(string); ok {
			i, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return err
			}
			dst.SetUint(i)
			return nil
		}
		if i, ok := src.(uint64); ok {
			dst.SetUint(i)
			return nil
		}
		if i, ok := src.(uint); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		if i, ok := src.(uint8); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		if i, ok := src.(uint16); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		if i, ok := src.(uint32); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		if i, ok := src.(int64); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		if i, ok := src.(int); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		if i, ok := src.(int8); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		if i, ok := src.(int16); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		if i, ok := src.(int32); ok {
			dst.SetUint(uint64(i))
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if s, ok := src.(string); ok {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			dst.SetInt(i)
			return nil
		}
		if i, ok := src.(int64); ok {
			dst.SetInt(i)
			return nil
		}
		if i, ok := src.(int); ok {
			dst.SetInt(int64(i))
			return nil
		}
		if i, ok := src.(int8); ok {
			dst.SetInt(int64(i))
			return nil
		}
		if i, ok := src.(int16); ok {
			dst.SetInt(int64(i))
			return nil
		}
		if i, ok := src.(int32); ok {
			dst.SetInt(int64(i))
			return nil
		}

		if i, ok := src.(uint64); ok {
			dst.SetInt(int64(i))
			return nil
		}
		if i, ok := src.(uint); ok {
			dst.SetInt(int64(i))
			return nil
		}
		if i, ok := src.(uint8); ok {
			dst.SetInt(int64(i))
			return nil
		}
		if i, ok := src.(uint16); ok {
			dst.SetInt(int64(i))
			return nil
		}
		if i, ok := src.(uint32); ok {
			dst.SetInt(int64(i))
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

	case reflect.Float32, reflect.Float64:
		if f, ok := src.(float64); ok {
			dst.SetFloat(f)
			return nil
		}
		if f, ok := src.(float32); ok {
			dst.SetFloat(float64(f))
			return nil
		}
		if i, ok := src.(int64); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(int); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(int8); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(int16); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(int32); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(uint64); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(uint); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(uint8); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(uint16); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		if i, ok := src.(uint32); ok {
			dst.SetFloat(float64(i))
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)
	case reflect.Bool:
		if b, ok := src.(bool); ok {
			dst.SetBool(b)
			return nil
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)
	case reflect.Slice:
		//exhaustive:ignore
		sliceKind := dst.Type().Elem().Kind()
		//exhaustive:ignore
		switch sliceKind {
		case reflect.String:
			if s, ok := src.([]string); ok {
				dst.Set(reflect.ValueOf(s))
				return nil
			}
			if s, ok := src.([]interface{}); ok {
				v2_, ok := cast.CastList[string, interface{}](s)
				if !ok {
					return fmt.Errorf("cannot cast %T to []string", src)
				}
				dst.Set(reflect.ValueOf(v2_))
				return nil
			}

			return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

		case reflect.Int:
			return SetIntListReflectValue[int](dst, src)

		case reflect.Int8:
			return SetIntListReflectValue[int8](dst, src)

		case reflect.Int16:
			return SetIntListReflectValue[int16](dst, src)

		case reflect.Int32:
			return SetIntListReflectValue[int32](dst, src)

		case reflect.Int64:
			return SetIntListReflectValue[int64](dst, src)

		case reflect.Uint:
			return SetIntListReflectValue[uint](dst, src)

		case reflect.Uint8:
			return SetIntListReflectValue[uint8](dst, src)

		case reflect.Uint16:
			return SetIntListReflectValue[uint16](dst, src)

		case reflect.Uint32:
			return SetIntListReflectValue[uint32](dst, src)

		case reflect.Uint64:
			return SetIntListReflectValue[uint64](dst, src)

		case reflect.Float32:
			return SetFloatListReflectValue[float32](dst, src)

		case reflect.Float64:
			return SetFloatListReflectValue[float64](dst, src)

		case reflect.Map:
			return SetStringMapListReflectValue[interface{}](dst, src)

		default:
			return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)
		}

	case reflect.Map:
		//exhaustive:ignore
		switch dst.Type().Elem().Kind() {
		case reflect.String:
			if m, ok := src.(map[string]string); ok {
				dst.Set(reflect.ValueOf(m))
				return nil
			}

			if m, ok := src.(map[string]interface{}); ok {
				v2_, ok := cast.CastStringMap[string, interface{}](m)
				if !ok {
					return fmt.Errorf("cannot cast %T to map[string]string", src)
				}
				dst.Set(reflect.ValueOf(v2_))
				return nil
			}

		case reflect.Interface:
			if m, ok := src.(map[string]interface{}); ok {
				dst.Set(reflect.ValueOf(m))
				return nil
			}

			if m, ok := src.(map[string]string); ok {
				v2_, ok := cast.CastStringMap[interface{}, string](m)
				if !ok {
					return fmt.Errorf("cannot cast %T to map[string]interface{}", src)
				}
				dst.Set(reflect.ValueOf(v2_))
				return nil

			}

		default:
			return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)
		}
		return fmt.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

	default:
		return fmt.Errorf("unsupported reflect.Value type %s", kind)
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

func SetStringMapListReflectValue[To interface{}](mapSlice reflect.Value, v interface{}) error {
	keyKind := mapSlice.Type().Elem().Key().Kind()
	switch keyKind {
	case reflect.String:
		r, ok := cast.CastList2[map[string]interface{}, interface{}](v)
		if !ok {
			return fmt.Errorf("cannot cast %T to []map[string]interface{}", v)
		}
		mapSlice.Set(reflect.ValueOf(r))
		return nil
	}
	return fmt.Errorf("cannot set reflect.Value of type %s from %T", keyKind, v)
}
