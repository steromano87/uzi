package variables_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetGetWithSingleLevel(t *testing.T) {
	lkm := variables.NewLinearKeysMap()
	lkm.Set("testKey", "testValue")

	assert.EqualValues(t, "testValue", lkm.Get("testKey"))
}

func TestSetGetWithMultipleLevels(t *testing.T) {
	lkm := variables.NewLinearKeysMap()
	lkm.Set("some.inner.level", "testValue")

	if assert.EqualValues(t, "testValue", lkm.Get("some.inner.level")) {
		assert.IsType(t, make(map[string]any), lkm.Get("some.inner"))
	}
}

func TestMultipleSetWithCommonPath(t *testing.T) {
	lkm := variables.NewLinearKeysMap()
	lkm.Set("some.inner.level", "foo")
	lkm.Set("some.other.level", "bar")

	assert.EqualValues(t, "foo", lkm.Get("some.inner.level"))
	assert.EqualValues(t, "bar", lkm.Get("some.other.level"))
}
