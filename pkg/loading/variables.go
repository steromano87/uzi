package loading

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"text/template"
)

type Variables struct {
	VariableSet
	globals VariableSet
}

func NewVariables() Variables {
	return Variables{
		VariableSet: VariableSet{},
		globals:     VariableSet{},
	}
}

func (v Variables) Globals() VariableSet {
	return v.globals
}

func (v Variables) Render(t string) (string, error) {
	parsedTemplate, err := template.New("variables").Funcs(sprig.FuncMap()).Parse(t)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	renderingVars := struct {
		Values  map[string]any
		Globals map[string]any
	}{
		Values:  v.VariableSet,
		Globals: v.globals,
	}
	if err = parsedTemplate.Execute(&output, renderingVars); err != nil {
		return "", err
	}

	return output.String(), nil
}

type VariableSet map[string]any

func (vs VariableSet) Set(name string, value any) {
	vs[name] = value
}

func (vs VariableSet) GetString(name string) (string, error) {
	value, err := vs.Get(name)
	return fmt.Sprintf("%v", value), err
}

func (vs VariableSet) GetInt(name string) (int, error) {
	value, err := vs.Get(name)
	if err != nil {
		return 0, err
	}

	convertedValue, isOk := value.(int)
	if !isOk {
		err = BadVariableCastError{
			Name:     name,
			CastType: "int",
			RawValue: value,
		}
		return 0, err
	}

	return convertedValue, nil
}

func (vs VariableSet) GetBool(name string) (bool, error) {
	value, err := vs.Get(name)
	if err != nil {
		return false, err
	}

	convertedValue, isOk := value.(bool)
	if !isOk {
		err = BadVariableCastError{
			Name:     name,
			CastType: "bool",
			RawValue: value,
		}
		return false, err
	}

	return convertedValue, nil
}

func (vs VariableSet) Get(name string) (any, error) {
	value, isPresent := vs[name]

	if !isPresent {
		err := VariableNotFoundError{Name: name}
		return nil, err
	}

	return value, nil
}

func (vs VariableSet) Delete(name string) {
	delete(vs, name)
}
