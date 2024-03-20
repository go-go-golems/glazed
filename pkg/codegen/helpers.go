package codegen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
)

func StructTypeToJen(typ reflect.Type) (jen.Code, error) {
	isPointer := typ.Kind() == reflect.Ptr
	if isPointer {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.New("input is not a struct")
	}

	var err error
	ret := jen.StructFunc(func(g *jen.Group) {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldType := field.Type

			jenType, err_ := TypeToJen(fieldType)
			if err_ != nil {
				err = err_
				return
			}

			g.Id(field.Name).Add(jenType)
		}
	})

	// found an error while generating the struct
	if err != nil {
		return nil, err
	}

	if isPointer {
		ret = jen.Op("*").Add(ret)
	}

	return ret, nil
}

// TypeToJen converts a reflect.Type to jen.Code
func TypeToJen(t reflect.Type) (jen.Code, error) {
	//nolint: exhaustive
	switch t.Kind() {
	case reflect.String:
		return jen.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return jen.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return jen.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return jen.Float64(), nil
	case reflect.Bool:
		return jen.Bool(), nil
	case reflect.Ptr:
		elemType, err := TypeToJen(t.Elem())
		if err != nil {
			return nil, err
		}
		return jen.Op("*").Add(elemType), nil
	case reflect.Struct:
		c, err := StructTypeToJen(t)
		if err != nil {
			return nil, err
		}
		return c, nil
	case reflect.Array, reflect.Slice:
		elemType, err := TypeToJen(t.Elem())
		if err != nil {
			return nil, err
		}
		return jen.Index().Add(elemType), nil
	case reflect.Map:
		keyType, err := TypeToJen(t.Key())
		if err != nil {
			return nil, err
		}
		elemType, err := TypeToJen(t.Elem())
		if err != nil {
			return nil, err
		}
		return jen.Map(keyType).Add(elemType), nil
	case reflect.Interface:
		return jen.Interface(), nil
	default:
		return nil, errors.Errorf("unsupported type %s", t.Kind())
	}
}

func StructValueToJen(structName string, s interface{}) (jen.Code, error) {
	val := reflect.ValueOf(s)
	typ := val.Type()

	// Check if the input is a struct or a pointer to a struct
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("input is not a struct or a pointer to a struct")
	}

	var err error
	// Start building the struct instantiation
	ret := jen.Var().Id(structName).Op("=").StructFunc(func(g *jen.Group) {
		for i := 0; i < val.NumField(); i++ {
			fieldVal := val.Field(i)
			fieldType := typ.Field(i)

			v, err2 := LiteralToJen(fieldVal)
			if err2 != nil {
				err = err2
				return
			}
			// Add the field and its value to the struct
			g.Id(fieldType.Name).Op(":").Add(v)
		}
	})

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func LiteralToJen(v reflect.Value) (jen.Code, error) {

	//nolint:exhaustive
	switch v.Kind() {
	case reflect.String:
		return jen.Lit(v.String()), nil
	case reflect.Int:
		return jen.Lit(int(v.Int())), nil
	case reflect.Int8:
		return jen.Lit(int8(v.Int())), nil
	case reflect.Int16:
		return jen.Lit(int16(v.Int())), nil
	case reflect.Int32:
		return jen.Lit(int32(v.Int())), nil
	case reflect.Int64:
		return jen.Lit(v.Int()), nil
	case reflect.Uint:
		return jen.Lit(uint(v.Uint())), nil
	case reflect.Uint8:
		return jen.Lit(uint8(v.Uint())), nil
	case reflect.Uint16:
		return jen.Lit(uint16(v.Uint())), nil
	case reflect.Uint32:
		return jen.Lit(uint32(v.Uint())), nil
	case reflect.Uint64:
		return jen.Lit(v.Uint()), nil
	case reflect.Float32:
		return jen.Lit(float32(v.Float())), nil
	case reflect.Float64:
		return jen.Lit(v.Float()), nil
	case reflect.Bool:
		return jen.Lit(v.Bool()), nil
	case reflect.Interface:
		return LiteralToJen(v.Elem())
	case reflect.Slice, reflect.Array:
		var err error
		t, err := TypeToJen(v.Type())
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert slice type to jen")
		}
		ret := jen.Add(t).ValuesFunc(func(g *jen.Group) {
			for i := 0; i < v.Len(); i++ {
				toJen, err_ := LiteralToJen(v.Index(i))
				if err_ != nil {
					err = err_
					return
				}
				g.Add(toJen)
			}
		})

		if err != nil {
			return nil, err
		}

		return ret, nil
	case reflect.Map:
		t, err := TypeToJen(v.Type().Key())
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert map key type to jen")
		}
		v_, err := TypeToJen(v.Type().Elem())
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert map value type to jen")
		}
		ret := jen.Map(t).Add(v_).
			Values(jen.DictFunc(func(d jen.Dict) {
				for _, key := range v.MapKeys() {
					k, err_ := LiteralToJen(key)
					if err_ != nil {
						err = err_
						return
					}
					v, err_ := LiteralToJen(v.MapIndex(key))
					if err_ != nil {
						err = err_
						return
					}
					d[k] = v
				}
			}))
		if err != nil {
			return nil, err
		}
		return ret, nil
	case reflect.Struct:
		var err error
		t := v.Type()
		ret := jen.Id(t.Name())
		if t.Name() == "" {
			c, err := TypeToJen(t)
			if err != nil {
				return nil, err
			}
			ret = jen.Add(c)
		}

		ret = ret.ValuesFunc(func(g *jen.Group) {
			for i := 0; i < v.NumField(); i++ {
				fieldVal := v.Field(i)
				fieldType := v.Type().Field(i)
				toJen, err2 := LiteralToJen(fieldVal)
				if err2 != nil {
					err = err2
					return
				}
				g.Id(fieldType.Name).Op(":").Add(toJen)
			}
		})
		if err != nil {
			return nil, err
		}
		return ret, nil
	// pointer to struct
	case reflect.Ptr:
		// check if pointer to struct
		if v.Elem().Kind() != reflect.Struct {
			return nil, errors.Errorf("unsupported type %s", v.Kind())
		}
		v_, err := LiteralToJen(v.Elem())
		if err != nil {
			return nil, err
		}
		return jen.Op("&").Add(v_), nil
	default:
		// Default case for unsupported types
		return nil, errors.Errorf("unsupported type %s", v.Kind())
	}
}

func ToGoCode(v interface{}) (string, error) {
	return ValueToGoCode(reflect.ValueOf(v))
}

func ValueToGoCode(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return strconv.Quote(v.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Float()), nil
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool()), nil
	case reflect.Slice, reflect.Array:
		return SliceToGoCode(v)
	case reflect.Struct:
		return StructToGoCode(v)
	case reflect.Map:
		return MapToGoCode(v)
	case reflect.Ptr:
		if v.IsNil() {
			return "nil", nil
		}

		if v.Elem().Kind() == reflect.Interface {
			// We have a pointer to an interface, now let's see if it contains a string
			actualValue := v.Elem().Elem() // This will dereference the interface to get the actual value
			if actualValue.Kind() == reflect.String {
				// We found a string inside the interface, return it quoted
				return strconv.Quote(actualValue.String()), nil
			} else {
				// It's not a string, handle other types or return an error
				return ValueToGoCode(actualValue)
			}
		}

		code, err := ValueToGoCode(v.Elem())
		if err != nil {
			return "", err
		}
		// Use a temporary variable for the pointer's value
		s := v.Elem().Type().String()
		if s != "interface {}" {
			s = "*" + s
		}
		return fmt.Sprintf("func() %s { tmp := %s; return &tmp }()", s, code), nil

	case reflect.Interface:
		if v.IsNil() {
			return "nil", nil
		}
		return ValueToGoCode(v.Elem())
	// Add more types as needed
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Sprintf("/* Unsupported type: %s */", v.Type().String()), errors.Errorf("unsupported type %s", v.Kind())
	}

	return "", errors.Errorf("unsupported type %s", v.Kind())
}

func SliceToGoCode(v reflect.Value) (string, error) {
	result := "[]"
	if v.Len() == 0 {
		result += v.Type().Elem().String() + "{}"
		return result, nil
	}

	result += v.Type().Elem().String() + "{"
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			result += ", "
		}
		code, err := ValueToGoCode(v.Index(i))
		if err != nil {
			return "", err
		}
		result += code
	}
	result += "}"
	return result, nil
}

func StructToGoCode(v reflect.Value) (string, error) {
	result := "&" + v.Type().String() + "{"
	for i := 0; i < v.NumField(); i++ {
		if i > 0 {
			result += ", "
		}
		field := v.Type().Field(i)
		code, err := ValueToGoCode(v.Field(i))
		if err != nil {
			return "", err
		}
		result += field.Name + ": " + code
	}
	result += "}"
	return result, nil
}

func MapToGoCode(v reflect.Value) (string, error) {
	result := "map[" + v.Type().Key().String() + "]" + v.Type().Elem().String() + "{"
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		code, err := ValueToGoCode(val)
		if err != nil {
			return "", err
		}
		keyCode, err := ValueToGoCode(key)
		if err != nil {
			return "", err
		}
		result += keyCode + ": " + code + ", "
	}
	result = result[:len(result)-2] + "}" // Remove the last ", "
	return result, nil
}
