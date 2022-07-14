package loading_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type VariablesTestSuite struct {
	suite.Suite
	variables loading.Variables
}

func (suite *VariablesTestSuite) SetupTest() {
	suite.variables = loading.NewVariables()
}

func (suite *VariablesTestSuite) TestSetGet() {
	suite.variables.Set("test", "testValue")
	actualValue, err := suite.variables.Get("test")

	assert.Equal(suite.T(), "testValue", actualValue)
	assert.NoError(suite.T(), err)
}

func (suite *VariablesTestSuite) TestGetNonExisting() {
	actualValue, err := suite.variables.Get("nonExisting")

	assert.Nil(suite.T(), actualValue, "Expected nil to be returned")
	assert.IsType(suite.T(), loading.VariableNotFoundError{}, err)
}

func (suite *VariablesTestSuite) TestGetString() {
	suite.variables.Set("test", "anotherValue")
	actualValue, err := suite.variables.Get("test")

	assert.Equal(suite.T(), "anotherValue", actualValue)
	assert.IsType(suite.T(), "", actualValue, "It is not a string")
	assert.NoError(suite.T(), err)
}

func (suite *VariablesTestSuite) TestGetStringNonExisting() {
	suite.variables.Set("test", "anotherValue")
	actualValue, err := suite.variables.GetInt("nonExisting")

	assert.Empty(suite.T(), actualValue, "Expected empty to be returned")
	assert.IsType(suite.T(), loading.VariableNotFoundError{}, err)
}

func (suite *VariablesTestSuite) TestGetStringBadType() {
	suite.variables.Set("test", 1)
	actualValue, err := suite.variables.GetString("test")
	assert.Equal(suite.T(), "1", actualValue)
	assert.NoError(suite.T(), err)
}

func (suite *VariablesTestSuite) TestGetInt() {
	suite.variables.Set("test", 1)
	actualValue, err := suite.variables.GetInt("test")

	assert.Equal(suite.T(), 1, actualValue)
	assert.NoError(suite.T(), err)
}

func (suite *VariablesTestSuite) TestGetIntNonExisting() {
	suite.variables.Set("test", 7)
	actualValue, err := suite.variables.GetInt("nonExisting")

	assert.Zero(suite.T(), actualValue, "Expected zero to be returned")
	assert.IsType(suite.T(), loading.VariableNotFoundError{}, err)
}

func (suite *VariablesTestSuite) TestGetIntBadType() {
	suite.variables.Set("test", "a")
	actualValue, err := suite.variables.GetInt("test")

	assert.Equal(suite.T(), 0, actualValue, "Expected zero-valued int")
	assert.IsType(suite.T(), loading.BadVariableCastError{}, err)
}

func (suite *VariablesTestSuite) TestGetBool() {
	suite.variables.Set("test", true)
	actualValue, err := suite.variables.GetBool("test")

	assert.Equal(suite.T(), true, actualValue)
	assert.NoError(suite.T(), err)
}

func (suite *VariablesTestSuite) TestGetBoolNonExisting() {
	suite.variables.Set("test", true)
	actualValue, err := suite.variables.GetBool("nonExisting")

	assert.False(suite.T(), actualValue, "Expected false to be returned")
	assert.IsType(suite.T(), loading.VariableNotFoundError{}, err)
}

func (suite *VariablesTestSuite) TestGetBoolBadType() {
	suite.variables.Set("test", "booleanValue")
	actualValue, err := suite.variables.GetBool("test")

	assert.Equal(suite.T(), false, actualValue, "Expected zero-valued bool (false)")
	assert.IsType(suite.T(), loading.BadVariableCastError{}, err)
}

func (suite *VariablesTestSuite) TestDelete() {
	suite.variables.Set("test", "finalValue")
	suite.variables.Delete("test")
	actualResult, err := suite.variables.Get("test")

	assert.Nil(suite.T(), actualResult)
	assert.IsType(suite.T(), loading.VariableNotFoundError{}, err)
}

func (suite *VariablesTestSuite) TestRenderValidTemplate() {
	suite.variables.Set("test", "My Value")
	myTemplate := "The value of 'test' is '{{ .Values.test }}'"
	rendered, err := suite.variables.Render(myTemplate)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), "The value of 'test' is 'My Value'", rendered)
	}
}

func (suite *VariablesTestSuite) TestRenderInvalidTemplate() {
	myTemplate := "Invalid template: {{"
	_, err := suite.variables.Render(myTemplate)

	assert.Error(suite.T(), err)
}

func TestVariablePoolTestSuite(t *testing.T) {
	suite.Run(t, new(VariablesTestSuite))
}
