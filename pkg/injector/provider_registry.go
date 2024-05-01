package injector

import "errors"

var providerRegistry = make(map[string]Provider)

func RegisterProvider(kind string, provider Provider) {
	providerRegistry[kind] = provider
}

func GetProvider(kind string) (Provider, error) {
	provider, ok := providerRegistry[kind]
	if !ok {
		return nil, errors.New("unknown provider: " + kind)
	}

	return provider, nil
}
