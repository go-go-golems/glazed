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
	case ParameterTypeStringListFromFile:
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
		m, ok := value.(map[string]string)
		if !ok {
			return "", errors.Errorf("expected map[string]string, got %T", value)
		}
		s := []string{}
		for k, v := range m {
			s = append(s, k+":"+v)
		}
		return strings.Join(s, ","), nil

	case ParameterTypeInteger:
		return fmt.Sprintf("%d", value), nil

	case ParameterTypeFloat:
		return fmt.Sprintf("%f", value), nil

	case ParameterTypeBool:
		v, ok := value.(bool)
		if !ok {
			return "", errors.Errorf("expected bool, got %T", value)
		}
		if v {
			return "true", nil
		}
		return "false", nil

	case ParameterTypeStringList:
		l, ok := value.([]string)
		if !ok {
			return "", errors.Errorf("expected []string, got %T", value)
		}
		return strings.Join(l, ","), nil

	case ParameterTypeIntegerList:
		v, ok := value.([]interface{})
		if !ok {
			return "", errors.Errorf("expected []interface{}, got %T", value)
		}
		l, ok := cast.CastInterfaceListToIntList[int64](v)
		if !ok {
			return "", errors.Errorf("expected []int64, got %T", value)
		}
		s := []string{}
		for _, i := range l {
			s = append(s, fmt.Sprintf("%d", i))
		}
		return strings.Join(s, ","), nil

	case ParameterTypeFloatList:
		v, ok := value.([]interface{})
		if !ok {
			return "", errors.Errorf("expected []interface{}, got %T", value)
		}
		l, ok := cast.CastInterfaceListToFloatList[float64](v)
		if !ok {
			return "", errors.Errorf("expected []float64, got %T", value)
		}
		s := []string{}
		for _, i := range l {
			s = append(s, fmt.Sprintf("%f", i))
		}
		return strings.Join(s, ","), nil
	}

	return "", errors.Errorf("unknown type %s", type_)
}
