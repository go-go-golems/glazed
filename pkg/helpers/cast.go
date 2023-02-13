package helpers

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

func CastToInt64List[From Number](list []From) ([]int64, bool) {
	ret := []int64{}

	for _, item := range list {
		ret = append(ret, int64(item))
	}

	return ret, true
}

func CastToNumberList[To Number, From Number](list []From) ([]To, bool) {
	ret := []To{}

	for _, item := range list {
		ret = append(ret, To(item))
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
