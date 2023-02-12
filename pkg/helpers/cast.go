package helpers

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
