package runtime

import (
	"fmt"
	"github.com/rs/zerolog"
)

type VariablePool struct {
	variables map[string]interface{}
	logger    *zerolog.Logger
}

func NewVariablePool(parentLogger *zerolog.Logger) *VariablePool {
	pool := new(VariablePool)
	pool.variables = make(map[string]interface{})

	poolLogger := parentLogger.With().Str("component", "Variable Pool").Logger()
	pool.logger = &poolLogger

	return pool
}

func (pool *VariablePool) Set(name string, value interface{}) {
	pool.variables[name] = value
	pool.logger.Info().Msgf("Set variable '%s' with value '%s'", name, value)
}

func (pool *VariablePool) GetString(name string) (string, error) {
	value, err := pool.Get(name)
	return fmt.Sprintf("%v", value), err
}

func (pool *VariablePool) GetInt(name string) (int, error) {
	value, err := pool.Get(name)
	if err != nil {
		return 0, err
	}

	convertedValue, isOk := value.(int)
	if !isOk {
		err = ErrBadVariableCast{
			Name:     name,
			CastType: "int",
			RawValue: value,
		}
		pool.logger.Warn().Err(err).Msgf("Caught an error while retrieving variable '%s'", name)
		return 0, err
	}

	return convertedValue, nil
}

func (pool *VariablePool) GetBool(name string) (bool, error) {
	value, err := pool.Get(name)
	if err != nil {
		return false, err
	}

	convertedValue, isOk := value.(bool)
	if !isOk {
		err = ErrBadVariableCast{
			Name:     name,
			CastType: "bool",
			RawValue: value,
		}
		pool.logger.Warn().Err(err).Msgf("Caught an error while retrieving variable '%s'", name)
		return false, err
	}

	return convertedValue, nil
}

func (pool *VariablePool) Get(name string) (interface{}, error) {
	pool.logger.Info().Msgf("Requested variable '%s'", name)
	value, isPresent := pool.variables[name]

	if !isPresent {
		err := ErrVariableNotFound{Name: name}
		pool.logger.Warn().Err(err).Msgf("Caught an error while retrieving variable '%s'", name)
		return nil, err
	}

	return value, nil
}

func (pool *VariablePool) Delete(name string) {
	delete(pool.variables, name)
	pool.logger.Info().Msgf("Deleted variable '%s'", name)
}
