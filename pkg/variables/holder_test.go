package variables_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleStringTemplating(t *testing.T) {
	templateString := "My name is {{ .Locals.name }}"
	holder := variables.NewHolder()
	holder.SetLocal("name", "Carl")

	renderedString, err := holder.Render(templateString)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "My name is Carl", renderedString)
	}
}

func TestMultipleLevelRendering(t *testing.T) {
	templateString := "My name is {{ .Locals.user.name }}"
	holder := variables.NewHolder()
	holder.SetLocal("user.name", "Carl")

	renderedString, err := holder.Render(templateString)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "My name is Carl", renderedString)
	}
}

func TestDifferentRenderingWithVariableChange(t *testing.T) {
	templateString := "My name is {{ .Locals.name }}"
	holder := variables.NewHolder()
	holder.SetLocal("name", "Carl")

	renderedString, err := holder.Render(templateString)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "My name is Carl", renderedString)
	}

	holder.SetLocal("name", "John")
	newRenderedString, err := holder.Render(templateString)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "My name is John", newRenderedString)
	}
}

func TestStringTemplatingWithFunctions(t *testing.T) {
	templateString := "My name is {{ .Locals.name | upper }}"
	holder := variables.NewHolder()
	holder.SetLocal("name", "Carl")

	renderedString, err := holder.Render(templateString)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "My name is CARL", renderedString)
	}
}
