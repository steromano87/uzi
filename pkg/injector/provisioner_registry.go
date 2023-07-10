package injector

import "errors"

var provisionerRegistry = map[string]Provisioner{}

func RegisterProvisioner(kind string, provisioner Provisioner) {
	provisionerRegistry[kind] = provisioner
}

func CreateProvisioner(kind string, spec map[string]any) (Provisioner, error) {
	provisioner, ok := provisionerRegistry[kind]
	if !ok {
		return nil, errors.New("cannot find a registered provisioner of type " + kind)
	}

	return provisioner, nil
}
