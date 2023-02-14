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

// CastToNumberList casts a list of From objects to To.
// This is useful for transform between different int types, for example.
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
