package variables

import "strings"

type LinearKeysMap struct {
	values map[string]any
}

func NewLinearKeysMap() LinearKeysMap {
	return LinearKeysMap{
		values: make(map[string]any),
	}
}

func (m *LinearKeysMap) Get(key string) any {
	keyTokens := strings.Split(key, ".")

	var source any
	source = m.values

	for index, token := range keyTokens {
		if index < len(keyTokens)-1 {
			intermediateValue := m.getInner(source, token)
			if intermediateValue == nil {
				return nil
			}
			source = intermediateValue
		} else {
			return m.getInner(source, token)
		}
	}

	return nil
}

func (m *LinearKeysMap) Set(key string, value any) {
	keyTokens := strings.Split(key, ".")

	intermediateMap := m.values

	for index, token := range keyTokens {
		if index < len(keyTokens)-1 {
			_, ok := intermediateMap[token]
			if !ok {
				intermediateMap[token] = make(map[string]any)
			}

			intermediateValueAsMap, ok := intermediateMap[token].(map[string]any)
			if ok {
				intermediateMap = intermediateValueAsMap
			} else {
				intermediateMap[token] = make(map[string]any)
				intermediateMap = make(map[string]any)
			}
		} else {
			intermediateMap[token] = value
		}
	}
}

func (m *LinearKeysMap) AsMap() map[string]any {
	return m.values
}

func (m *LinearKeysMap) FromMap(input map[string]any) {
	m.values = input
}

func (m *LinearKeysMap) getInner(source any, key string) any {
	sourceAsMap, ok := source.(map[string]any)
	if !ok {
		return nil
	}

	value, ok := sourceAsMap[key]
	if !ok {
		return nil
	}

	return value
}
