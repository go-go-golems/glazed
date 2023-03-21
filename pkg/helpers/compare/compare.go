package compare

import (
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
)

func IsLowerThan(a, b interface{}) bool {
	if IsOfNumberType(a) && IsOfNumberType(b) {
		af, ok := cast.CastNumberInterfaceToFloat[float64](a)
		if !ok {
			return false
		}
		bf, ok := cast.CastNumberInterfaceToFloat[float64](b)
		if !ok {
			return false
		}
		return af < bf
	}

	if IsString(a) && IsString(b) {
		return a.(string) < b.(string)
	}

	return false
}

func IsOfNumberType(a interface{}) bool {
	switch a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

func IsString(a interface{}) bool {
	switch a.(type) {
	case string:
		return true
	default:
		return false
	}
}
