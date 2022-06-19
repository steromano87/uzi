package runtime

import "fmt"

type ErrVariableNotFound struct {
	Name string
}

func (vnf ErrVariableNotFound) Error() string {
	return fmt.Sprintf("variable '%s' not found", vnf.Name)
}
