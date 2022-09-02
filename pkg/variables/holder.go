package variables

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	"text/template"
)

type Holder struct {
	locals   LinearKeysMap
	globals  LinearKeysMap
	iterVars LinearKeysMap
}

func NewHolder() *Holder {
	return &Holder{
		locals:   NewLinearKeysMap(),
		globals:  NewLinearKeysMap(),
		iterVars: NewLinearKeysMap(),
	}
}

func (h *Holder) SetGlobals(input map[string]any) {
	newGlobals := NewLinearKeysMap()
	newGlobals.FromMap(input)
	h.globals = newGlobals
}

func (h *Holder) UpdateIterVars(input map[string]any) {
	newIterVars := NewLinearKeysMap()
	newIterVars.FromMap(input)
	h.iterVars = newIterVars
}

func (h *Holder) SetLocal(key string, value any) {
	h.locals.Set(key, value)
}

func (h *Holder) Globals() map[string]any {
	return h.globals.AsMap()
}

func (h *Holder) IterVars() map[string]any {
	return h.iterVars.AsMap()
}

func (h *Holder) Render(t string) (string, error) {
	parsedTemplate, err := template.New("variables").Funcs(sprig.FuncMap()).Parse(t)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	renderingVars := struct {
		Locals   map[string]any
		Globals  map[string]any
		IterVars map[string]any
	}{
		Locals:   h.locals.AsMap(),
		Globals:  h.globals.AsMap(),
		IterVars: h.iterVars.AsMap(),
	}
	if err = parsedTemplate.Execute(&output, renderingVars); err != nil {
		return "", err
	}

	return output.String(), nil
}
