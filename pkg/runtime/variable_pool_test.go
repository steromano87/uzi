package runtime_test

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type VariablePoolTestSuite struct {
	suite.Suite
	logger zerolog.Logger
}

func (suite *VariablePoolTestSuite) SetupTest() {
	suite.logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
}

func (suite *VariablePoolTestSuite) TestSetGet() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", "testValue")
	actualValue, err := vp.Get("test")

	assert.Equal(suite.T(), "testValue", actualValue)
	assert.NoError(suite.T(), err)
}

func (suite *VariablePoolTestSuite) TestGetNonExisting() {
	vp := runtime.NewVariablePool(&suite.logger)
	actualValue, err := vp.Get("nonExisting")

	assert.Nil(suite.T(), actualValue, "Expected nil to be returned")
	assert.IsType(suite.T(), runtime.ErrVariableNotFound{}, err)
}

func (suite *VariablePoolTestSuite) TestGetString() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", "anotherValue")
	actualValue, err := vp.Get("test")

	assert.Equal(suite.T(), "anotherValue", actualValue)
	assert.IsType(suite.T(), "", actualValue, "It is not a string")
	assert.NoError(suite.T(), err)
}

func (suite *VariablePoolTestSuite) TestGetStringNonExisting() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", "anotherValue")
	actualValue, err := vp.GetInt("nonExisting")

	assert.Empty(suite.T(), actualValue, "Expected empty to be returned")
	assert.IsType(suite.T(), runtime.ErrVariableNotFound{}, err)
}

func (suite *VariablePoolTestSuite) TestGetStringBadType() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", 1)
	actualValue, err := vp.GetString("test")
	assert.Equal(suite.T(), "1", actualValue)
	assert.NoError(suite.T(), err)
}

func (suite *VariablePoolTestSuite) TestGetInt() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", 1)
	actualValue, err := vp.GetInt("test")

	assert.Equal(suite.T(), 1, actualValue)
	assert.NoError(suite.T(), err)
}

func (suite *VariablePoolTestSuite) TestGetIntNonExisting() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", 7)
	actualValue, err := vp.GetInt("nonExisting")

	assert.Zero(suite.T(), actualValue, "Expected zero to be returned")
	assert.IsType(suite.T(), runtime.ErrVariableNotFound{}, err)
}

func (suite *VariablePoolTestSuite) TestGetIntBadType() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", "a")
	actualValue, err := vp.GetInt("test")

	assert.Equal(suite.T(), 0, actualValue, "Expected zero-valued int")
	assert.IsType(suite.T(), runtime.ErrBadVariableCast{}, err)
}

func (suite *VariablePoolTestSuite) TestGetBool() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", true)
	actualValue, err := vp.GetBool("test")

	assert.Equal(suite.T(), true, actualValue)
	assert.NoError(suite.T(), err)
}

func (suite *VariablePoolTestSuite) TestGetBoolNonExisting() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", true)
	actualValue, err := vp.GetBool("nonExisting")

	assert.False(suite.T(), actualValue, "Expected false to be returned")
	assert.IsType(suite.T(), runtime.ErrVariableNotFound{}, err)
}

func (suite *VariablePoolTestSuite) TestGetBoolBadType() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", "booleanValue")
	actualValue, err := vp.GetBool("test")

	assert.Equal(suite.T(), false, actualValue, "Expected zero-valued bool (false)")
	assert.IsType(suite.T(), runtime.ErrBadVariableCast{}, err)
}

func (suite *VariablePoolTestSuite) TestDelete() {
	vp := runtime.NewVariablePool(&suite.logger)
	vp.Set("test", "finalValue")
	vp.Delete("test")
	actualResult, err := vp.Get("test")

	assert.Nil(suite.T(), actualResult)
	assert.IsType(suite.T(), runtime.ErrVariableNotFound{}, err)
}

func TestVariablePoolTestSuite(t *testing.T) {
	suite.Run(t, new(VariablePoolTestSuite))
}
