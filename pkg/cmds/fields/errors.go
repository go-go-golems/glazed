package fields

import "fmt"

type ErrInvalidValue struct {
	Value interface{}
	Name  string
	Type  Type
}

func (e ErrInvalidValue) Error() string {
	return fmt.Sprintf("invalid value %v for parameter %s of type %s", e.Value, e.Name, e.Type)
}
