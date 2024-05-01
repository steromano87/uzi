package schedule

import (
	"errors"
	"github.com/spf13/viper"
)

type ProfileParseFunc func(rawConfig *viper.Viper) (Profile, error)

var profilesRegistry = make(map[string]ProfileParseFunc)

func RegisterParseFunc(kind string, parseFunc ProfileParseFunc) {
	profilesRegistry[kind] = parseFunc
}

func Parse(kind string, rawConfig *viper.Viper) (Profile, error) {
	parserFunc, ok := profilesRegistry[kind]
	if !ok {
		return nil, errors.New("unknown profile kind: " + kind)
	}

	profile, err := parserFunc(rawConfig)
	if err != nil {
		return nil, err
	}

	return profile, nil
}
