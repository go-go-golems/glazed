package parameters

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"strings"
)

func RenderValue(type_ ParameterType, value interface{}) (string, error) {
	switch type_ {
	case ParameterTypeString:
		fallthrough
	case ParameterTypeStringFromFile:
		fallthrough
	case ParameterTypeObjectListFromFile:
		fallthrough
	case ParameterTypeObjectFromFile:
		fallthrough
	case ParameterTypeDate:
		fallthrough
	case ParameterTypeChoice:
		s, ok := value.(string)
		if !ok {
			return "", errors.Errorf("expected string, got %T", value)
		}
		return s, nil

	case ParameterTypeKeyValue:
		m, ok := cast.CastInterfaceToStringMap[string, interface{}](value)
		if !ok {
			return "", errors.Errorf("expected map[string]string, got %T", value)
		}
		s := []string{}
		for k, v := range m {
			s = append(s, k+":"+v)
		}
		return strings.Join(s, ","), nil

	case ParameterTypeInteger:
		v, ok := cast.CastNumberInterfaceToInt[int64](value)
		if !ok {
			return "", errors.Errorf("expected int, got %T", value)
		}
		return fmt.Sprintf("%d", v), nil

	case ParameterTypeFloat:
		v, ok := cast.CastNumberInterfaceToFloat[float64](value)
		if !ok {
			return "", errors.Errorf("expected float, got %T", v)
		}
		return fmt.Sprintf("%f", v), nil

	case ParameterTypeBool:
		v, ok := value.(bool)
		if !ok {
			return "", errors.Errorf("expected bool, got %T", value)
		}
		if v {
			return "true", nil
		}
		return "false", nil

	case ParameterTypeStringListFromFile:
		fallthrough
	case ParameterTypeStringList:
		l, ok := cast.CastList2[string, interface{}](value)
		if !ok {
			return "", errors.Errorf("expected []string, got %T", value)
		}
		return strings.Join(l, ","), nil

	case ParameterTypeIntegerList:
		l, ok := cast.CastListToIntList2[int64](value)
		if !ok {
			return "", errors.Errorf("expected []interface{}, got %T", value)
		}
		s := []string{}
		for _, i := range l {
			s = append(s, fmt.Sprintf("%d", i))
		}
		return strings.Join(s, ","), nil

	case ParameterTypeFloatList:
		l, ok := cast.CastListToFloatList2[float64](value)
		if !ok {
			return "", errors.Errorf("expected []interface{}, got %T", value)
		}
		s := []string{}
		for _, i := range l {
			s = append(s, fmt.Sprintf("%f", i))
		}
		return strings.Join(s, ","), nil
	}

	return "", errors.Errorf("unknown type %s", type_)
}
