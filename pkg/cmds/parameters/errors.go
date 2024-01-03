package parameters

import "fmt"

type ErrInvalidValue struct {
	Value interface{}
	Name  string
	Type  ParameterType
}

func (e ErrInvalidValue) Error() string {
	return fmt.Sprintf("invalid value %v for parameter %s of type %s", e.Value, e.Name, e.Type)
}
