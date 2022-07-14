package loading

import "fmt"

type VariableNotFoundError struct {
	Name string
}

func (vnf VariableNotFoundError) Error() string {
	return fmt.Sprintf("variable '%s' not found", vnf.Name)
}
