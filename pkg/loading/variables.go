package loading

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"text/template"
)

type Variables struct {
	Values  map[string]any
	Globals map[string]any
}

func NewVariables() Variables {
	return Variables{
		Values:  map[string]any{},
		Globals: map[string]any{},
	}
}

func (v Variables) Set(name string, value any) {
	v.Values[name] = value
}

func (v Variables) GetString(name string) (string, error) {
	value, err := v.Get(name)
	return fmt.Sprintf("%v", value), err
}

func (v Variables) GetInt(name string) (int, error) {
	value, err := v.Get(name)
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

func (v Variables) GetBool(name string) (bool, error) {
	value, err := v.Get(name)
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

func (v Variables) Get(name string) (any, error) {
	value, isPresent := v.Values[name]

	if !isPresent {
		err := VariableNotFoundError{Name: name}
		return nil, err
	}

	return value, nil
}

func (v Variables) Delete(name string) {
	delete(v.Values, name)
}

func (v Variables) Render(t string) (string, error) {
	parsedTemplate, err := template.New("variables").Funcs(template.FuncMap(sprig.FuncMap())).Parse(t)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	if err = parsedTemplate.Execute(&output, v); err != nil {
		return "", err
	}

	return output.String(), nil
}
