package schema

import "fmt"

type ErrInvalidSection struct {
	Name     string
	Expected string
}

func (e ErrInvalidSection) Error() string {
	if e.Expected == "" {
		return fmt.Sprintf("invalid section: %s", e.Name)
	}
	return fmt.Sprintf("invalid section: %s (expected %s)", e.Name, e.Expected)
}
