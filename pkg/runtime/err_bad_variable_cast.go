package runtime

import (
	"fmt"
)

type ErrBadVariableCast struct {
	Name     string
	CastType string
	RawValue interface{}
}

func (vc ErrBadVariableCast) Error() string {
	return fmt.Sprintf("error when casting variable '%s' as %s, raw value is '%v'", vc.Name, vc.CastType, vc.RawValue)
}
