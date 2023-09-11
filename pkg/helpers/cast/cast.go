package cast

import (
	"github.com/pkg/errors"
	"reflect"
)

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64
}

type FloatNumber interface {
	float32 | float64
}

type SignedInt interface {
	int | int8 | int16 | int32 | int64
}

type UnsignedInt interface {
	uint | uint8 | uint16 | uint32 | uint64 | uintptr
}

// CastList casts a list of From objects to To, by casting it to an interface{} in between.
func CastList[To any, From any](list []From) ([]To, bool) {
	ret := []To{}

	for _, item := range list {
		var v interface{} = item
		casted, ok := v.(To)
		if !ok {
			return ret, false
		}

		ret = append(ret, casted)
	}

	return ret, true
}

// CastList2 attempts even harder to cast a list of From object to To, by checking if we might
// be dealing with a list masquerading as a interface{}, then a []interface{}, before checking for []To.
func CastList2[To any, From any](list interface{}) ([]To, bool) {
	ret := []To{}

	switch l := list.(type) {
	case []interface{}:
		for _, item := range l {
			casted, ok := item.(To)
			if !ok {
				return ret, false
			}

			ret = append(ret, casted)
		}
	case []From:
		return CastList[To, From](l)
	case []To:
		ret = append(ret, l...)
	default:
		return ret, false
	}

	return ret, true
}

// CastListToInterfaceList attempts to convert the given value to a list of interface{}.
//
// The function checks if the provided value is a slice or an array. If so,
// it returns a slice of interface{}, where each item in the original slice or array
// is converted to its interface{} representation.
func CastListToInterfaceList(value interface{}) ([]interface{}, error) {
	val := reflect.ValueOf(value)

	// Check if the value is a slice or array
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		// Create an empty slice of interface{} with the appropriate length
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = val.Index(i).Interface()
		}
		return result, nil
	default:
		return nil, errors.New("the provided value is not a list")
	}
}

// CastToNumberList casts a list of From objects to To.
// This is useful for transform between different int types, for example.
func CastToNumberList[To Number, From Number](list []From) ([]To, bool) {
	ret := []To{}

	for _, item := range list {
		ret = append(ret, To(item))
	}

	return ret, true
}

func CastListToIntList2[To SignedInt | UnsignedInt](list interface{}) ([]To, bool) {
	switch t := list.(type) {
	case []interface{}:
		return CastInterfaceListToIntList[To](t)
	case []int:
		return CastToNumberList[To, int](t)
	case []int8:
		return CastToNumberList[To, int8](t)
	case []int16:
		return CastToNumberList[To, int16](t)
	case []int32:
		return CastToNumberList[To, int32](t)
	case []int64:
		return CastToNumberList[To, int64](t)
	case []uint:
		return CastToNumberList[To, uint](t)
	case []uint8:
		return CastToNumberList[To, uint8](t)
	case []uint16:
		return CastToNumberList[To, uint16](t)
	case []uint32:
		return CastToNumberList[To, uint32](t)
	case []uint64:
		return CastToNumberList[To, uint64](t)
	case []uintptr:
		return CastToNumberList[To, uintptr](t)
	default:
		return []To{}, false
	}
}

func CastInterfaceListToIntList[To SignedInt | UnsignedInt](list []interface{}) ([]To, bool) {
	ret := []To{}

	for _, item := range list {
		f, ok := CastNumberInterfaceToInt[To](item)
		if !ok {
			return ret, false
		}
		ret = append(ret, f)
	}

	return ret, true
}

func CastListToFloatList2[To FloatNumber](list interface{}) ([]To, bool) {
	switch t := list.(type) {
	case []interface{}:
		return CastInterfaceListToFloatList[To](t)
	case []float32:
		return CastToNumberList[To, float32](t)
	case []float64:
		return CastToNumberList[To, float64](t)
	default:
		return []To{}, false
	}
}

func CastInterfaceListToFloatList[To FloatNumber](list []interface{}) ([]To, bool) {
	ret := []To{}

	for _, item := range list {
		f, ok := CastNumberInterfaceToFloat[To](item)
		if !ok {
			return ret, false
		}
		ret = append(ret, f)
	}

	return ret, true
}
func CastStringMap[To any, From any](m map[string]From) (map[string]To, bool) {
	ret := map[string]To{}

	for k, v := range m {
		var item interface{} = v
		casted, ok := item.(To)
		if !ok {
			return ret, false
		}

		ret[k] = casted
	}

	return ret, true
}

func CastInterfaceToStringMap[To any, From any](m interface{}) (map[string]To, bool) {
	ret := map[string]To{}

	switch m := m.(type) {
	case map[string]To:
		return m, true
	case map[string]interface{}:
		return CastStringMap[To, interface{}](m)
	case map[string]From:
		return CastStringMap[To, From](m)
	default:
		return ret, false
	}
}

func CastMapToInterfaceMap[From any](m map[string]From) map[string]interface{} {
	ret := map[string]interface{}{}

	for k, v := range m {
		ret[k] = v
	}

	return ret
}

func CastNumberInterfaceToInt[To SignedInt | UnsignedInt](i interface{}) (To, bool) {
	switch i := i.(type) {
	case To:
		return i, true
	case int:
		return To(i), true
	case int8:
		return To(i), true
	case int16:
		return To(i), true
	case int32:
		return To(i), true
	case int64:
		return To(i), true
	case uint:
		return To(i), true
	case uint8:
		return To(i), true
	case uint16:
		return To(i), true
	case uint32:
		return To(i), true
	case uint64:
		return To(i), true
	case uintptr:
		return To(i), true
	default:
		return 0, false
	}
}

func CastFloatInterfaceToFloat[To FloatNumber](i interface{}) (To, bool) {
	switch i := i.(type) {
	case To:
		return i, true
	case float32:
		return To(i), true
	case float64:
		return To(i), true
	default:
		return 0, false
	}
}

func CastNumberInterfaceToFloat[To FloatNumber](i interface{}) (To, bool) {
	switch i := i.(type) {
	case To:
		return i, true
	case int:
		return To(i), true
	case int8:
		return To(i), true
	case int16:
		return To(i), true
	case int32:
		return To(i), true
	case int64:
		return To(i), true
	case uint:
		return To(i), true
	case uint8:
		return To(i), true
	case uint16:
		return To(i), true
	case uint32:
		return To(i), true
	case uint64:
		return To(i), true
	case uintptr:
		return To(i), true
	case float32:
		return To(i), true
	case float64:
		return To(i), true
	default:
		return 0, false
	}
}

func CastInterfaceToIntList[To SignedInt | UnsignedInt](i interface{}) ([]To, bool) {
	switch i := i.(type) {
	case []int:
		return CastToNumberList[To, int](i)
	case []int8:
		return CastToNumberList[To, int8](i)
	case []int16:
		return CastToNumberList[To, int16](i)
	case []int32:
		return CastToNumberList[To, int32](i)
	case []int64:
		return CastToNumberList[To, int64](i)
	case []uint:
		return CastToNumberList[To, uint](i)
	case []uint8:
		return CastToNumberList[To, uint8](i)
	case []uint16:
		return CastToNumberList[To, uint16](i)
	case []uint32:
		return CastToNumberList[To, uint32](i)
	case []uint64:
		return CastToNumberList[To, uint64](i)
	case []uintptr:
		return CastToNumberList[To, uintptr](i)
	case []interface{}:
		return CastInterfaceListToIntList[To](i)
	default:
		return nil, false
	}
}

func CastInterfaceToFloatList[To FloatNumber](i interface{}) ([]To, bool) {
	switch i := i.(type) {
	case []float32:
		return CastToNumberList[To, float32](i)
	case []float64:
		return CastToNumberList[To, float64](i)
	case []interface{}:
		return CastInterfaceListToFloatList[To](i)
	default:
		return nil, false
	}
}

func CastStringMap2[To any, From any](m interface{}) (map[string]To, bool) {
	casted, ok := m.(map[string]From)
	if !ok {
		// try to cast to map[string]interface{}
		casted2, ok := m.(map[string]interface{})
		if !ok {
			return map[string]To{}, false
		}
		return CastStringMap[To, interface{}](casted2)
	}

	return CastStringMap[To, From](casted)
}

func CastMapMember[To any](m map[string]interface{}, k string) (*To, bool) {
	v, ok := m[k]
	if !ok {
		return nil, false
	}

	casted, ok := v.(To)
	if !ok {
		return nil, false
	}

	return &casted, true
}
