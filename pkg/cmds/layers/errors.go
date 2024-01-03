package layers

import "fmt"

type ErrInvalidParameterLayer struct {
	Name     string
	Expected string
}

func (e ErrInvalidParameterLayer) Error() string {
	if e.Expected == "" {
		return fmt.Sprintf("invalid parameter layer: %s", e.Name)
	}
	return fmt.Sprintf("invalid parameter layer: %s (expected %s)", e.Name, e.Expected)
}
