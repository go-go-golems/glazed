package reflect

import (
	"math"
	"reflect"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
)

// NOTE(manuel, 2024-07-03) This is quite a mess of a function and I'm not even entirely sure
// what it's purpose is, and what the src can be (it tries to handle *interface{} but doesn't seem to handle
// other point values that easily?

func SetReflectValue(dst reflect.Value, src interface{}) error {
	kind := dst.Kind()
	if !dst.IsValid() {
		return errors.Errorf("cannot set invalid reflect.Value of type %s", kind)
	}
	if !dst.CanSet() {
		return errors.Errorf("cannot set reflect.Value of type %s", kind)
	}

	if src == nil {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	setValue := func(src interface{}) error {
		//exhaustive:ignore
		switch kind {
		case reflect.String:
			srcValue := reflect.ValueOf(src)
			srcKind := srcValue.Kind()

			assignableToString := srcValue.Type().AssignableTo(reflect.TypeOf(""))

			if srcKind == reflect.String || (srcKind == reflect.Ptr && srcValue.Elem().Kind() == reflect.String) {
				dst.SetString(srcValue.String())
				return nil
			}

			if assignableToString {
				dst.Set(srcValue)
				return nil
			}

			return errors.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

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
				if i < 0 {
					return errors.Errorf("cannot convert negative int64 %d to uint64", i)
				}
				dst.SetUint(uint64(i))
				return nil
			}
			if i, ok := src.(int); ok {
				if i < 0 {
					return errors.Errorf("cannot convert negative int %d to uint64", i)
				}
				dst.SetUint(uint64(i))
				return nil
			}
			if i, ok := src.(int8); ok {
				if i < 0 {
					return errors.Errorf("cannot convert negative int8 %d to uint64", i)
				}
				dst.SetUint(uint64(i))
				return nil
			}
			if i, ok := src.(int16); ok {
				if i < 0 {
					return errors.Errorf("cannot convert negative int16 %d to uint64", i)
				}
				dst.SetUint(uint64(i))
				return nil
			}
			if i, ok := src.(int32); ok {
				if i < 0 {
					return errors.Errorf("cannot convert negative int32 %d to uint64", i)
				}
				dst.SetUint(uint64(i))
				return nil
			}
			return errors.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

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
				if i > uint64(math.MaxInt64) { // MaxInt64
					return errors.Errorf("uint64 value %d overflows int64", i)
				}
				dst.SetInt(int64(i))
				return nil
			}
			if i, ok := src.(uint); ok {
				if i > uint(math.MaxInt64) { // MaxInt64
					return errors.Errorf("uint value %d overflows int64", i)
				}
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
			return errors.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

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
			return errors.Errorf("cannot set reflect.Value of type %s from %T", kind, src)
		case reflect.Bool:
			if b, ok := src.(bool); ok {
				dst.SetBool(b)
				return nil
			}
			return errors.Errorf("cannot set reflect.Value of type %s from %T", kind, src)
		case reflect.Slice:
			//exhaustive:ignore
			dstElementKind := dst.Type().Elem().Kind()
			//exhaustive:ignore
			switch dstElementKind {
			case reflect.String:
				s, err := cast.CastListToStringList(src)
				if err != nil {
					return errors.Errorf("cannot cast %T to []string", src)
				}

				// Get the type of the target slice elements
				targetElemType := dst.Type().Elem()

				// Get the type of the source slice elements
				sourceElemType := reflect.TypeOf(src).Elem()

				// Check if the source slice is assignable to the target slice
				if reflect.TypeOf(s).AssignableTo(dst.Type()) {
					// Direct assignment is possible
					dst.Set(reflect.ValueOf(s))
					return nil
				}

				if sourceElemType.Kind() == dstElementKind {
					// Direct assignment is not possible, perform type conversion
					newSlice := reflect.MakeSlice(dst.Type(), len(s), len(s))
					for i, s := range s {
						newSlice.Index(i).Set(reflect.ValueOf(s).Convert(targetElemType))
					}
					dst.Set(newSlice)
					return nil
				}

				return errors.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

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
				// try to cast each element of src into type of dst
				dstElemType := dst.Type().Elem()
				srcVal := reflect.ValueOf(src)
				newSlice := reflect.MakeSlice(reflect.SliceOf(dstElemType), srcVal.Len(), srcVal.Cap())

				for i := 0; i < srcVal.Len(); i++ {
					srcElem := srcVal.Index(i)
					if srcElem.Kind() == reflect.Interface && !srcElem.IsNil() {
						srcElem = srcElem.Elem()
					}
					dstElem := reflect.New(dstElemType).Elem()

					if srcElem.Type().AssignableTo(dstElemType) {
						dstElem.Set(srcElem)
					} else if srcElem.Type().ConvertibleTo(dstElemType) {
						dstElem.Set(srcElem.Convert(dstElemType))
					} else {
						return errors.Errorf("cannot convert element %d of type %s to %s", i, srcElem.Type(), dstElemType)
					}

					newSlice.Index(i).Set(dstElem)
				}

				dst.Set(newSlice)
				return nil
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
						return errors.Errorf("cannot cast %T to map[string]string", src)
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
						return errors.Errorf("cannot cast %T to map[string]interface{}", src)
					}
					dst.Set(reflect.ValueOf(v2_))
					return nil

				}

			default:
				return errors.Errorf("cannot set reflect.Value of type %s from %T", kind, src)
			}
			return errors.Errorf("cannot set reflect.Value of type %s from %T", kind, src)

		default:
			return errors.Errorf("unsupported reflect.Value type %s", kind)
		}

	}

	err := setValue(src)
	if err != nil {
		// try again, checking if src is an *interface{} and unwrapping that
		if i, ok := src.(*interface{}); ok {
			// if i == nil, we can't do anything
			if i == nil {
				return nil
			}

			src = *i
			err = setValue(src)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func SetIntListReflectValue[To cast.Number](value reflect.Value, v interface{}) error {
	if s, ok := v.([]int64); ok {
		s2, ok := cast.CastToNumberList[To, int64](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]int); ok {
		s2, ok := cast.CastToNumberList[To, int](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]int8); ok {
		s2, ok := cast.CastToNumberList[To, int8](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]int16); ok {
		s2, ok := cast.CastToNumberList[To, int16](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]int32); ok {
		s2, ok := cast.CastToNumberList[To, int32](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint64); ok {
		s2, ok := cast.CastToNumberList[To, uint64](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint); ok {
		s2, ok := cast.CastToNumberList[To, uint](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint8); ok {
		s2, ok := cast.CastToNumberList[To, uint8](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint16); ok {
		s2, ok := cast.CastToNumberList[To, uint16](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]uint32); ok {
		s2, ok := cast.CastToNumberList[To, uint32](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]interface{}); ok {
		v2_, ok := cast.CastList[int, interface{}](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		v3_, ok := cast.CastToNumberList[To, int](v2_)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(v3_))
		return nil
	}

	return errors.Errorf("cannot set reflect.Value of type %s from %T", value.Kind(), v)
}

func SetFloatListReflectValue[To cast.FloatNumber](value reflect.Value, v interface{}) error {
	if s, ok := v.([]float64); ok {
		s2, ok := cast.CastToNumberList[To, float64](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]float32); ok {
		s2, ok := cast.CastToNumberList[To, float32](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}
		value.Set(reflect.ValueOf(s2))
		return nil
	}

	if s, ok := v.([]interface{}); ok {
		v2_, ok := cast.CastInterfaceListToFloatList[To](s)
		if !ok {
			return errors.Errorf("cannot cast %T to []%T", v, To(0))
		}

		value.Set(reflect.ValueOf(v2_))
		return nil
	}

	return SetIntListReflectValue[To](value, v)
}

func SetStringMapListReflectValue[To interface{}](mapSlice reflect.Value, v interface{}) error {
	keyKind := mapSlice.Type().Elem().Key().Kind()
	if keyKind == reflect.String {
		r, ok := cast.CastList2[map[string]interface{}, interface{}](v)
		if !ok {
			return errors.Errorf("cannot cast %T to []map[string]interface{}", v)
		}
		mapSlice.Set(reflect.ValueOf(r))
		return nil
	}
	return errors.Errorf("cannot set reflect.Value of type %s from %T", keyKind, v)
}

// StripInterface takes a reflect.Value and returns the underlying type by recursively
// stripping off interface{} types while preserving pointer types.
//
// For example:
// - string -> string
// - interface{}(string) -> string
// - *string -> *string
// - interface{}(*string) -> *string
// - *interface{}(string) -> *string
// - interface{}(*interface{}(string)) -> *string
//
// If the value is invalid or nil, returns nil.
func StripInterface(v reflect.Value) reflect.Type {
	if !v.IsValid() {
		return nil
	}

	v = StripInterfaceFromValue(v)
	if !v.IsValid() {
		return nil
	}

	return v.Type()
}

// StripInterfaceFromValue takes a reflect.Value and returns the underlying value by recursively
// stripping off interface{} types while preserving pointer types.
//
// For example:
// - string -> string value
// - interface{}(string) -> string value
// - *string -> *string value
// - interface{}(*string) -> *string value
// - *interface{}(string) -> *string value
// - interface{}(*interface{}(string)) -> *string value
//
// If the value is invalid or nil, returns an invalid reflect.Value.
func StripInterfaceFromValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}

	// If it's a pointer type, we want to keep the pointer but strip interfaces from the element type
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		// Get the element value after stripping interfaces
		elemValue := StripInterfaceFromValue(v.Elem())
		if !elemValue.IsValid() {
			return reflect.Value{}
		}
		// Create a new pointer to the stripped element value
		ptr := reflect.New(elemValue.Type())
		ptr.Elem().Set(elemValue)
		return ptr
	}

	// Handle interface stripping
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = StripInterfaceFromValue(v.Elem())
	}

	return v
}

// StripInterfaceValue takes an interface{} and returns the underlying value by recursively
// stripping off interface{} and pointer types.
//
// For example:
// - string -> string value
// - interface{}(string) -> string value
// - *string -> string value
// - interface{}(*string) -> string value
// - *interface{}(string) -> string value
// - interface{}(*interface{}(string)) -> string value
//
// If the value is nil, returns nil.
func StripInterfaceValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	value := StripInterfaceFromValue(reflect.ValueOf(v))
	if !value.IsValid() {
		return nil
	}
	return value.Interface()
}
