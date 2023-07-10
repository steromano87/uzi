package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc"
	"reflect"
)

const (
	providedProvisionerKind = "provided"
	injectorsSpecsKeys      = "injectors"
)

type ProvidedProvisioner struct {
	injectors []ProvidedInjectorSpec
}

type ProvidedInjectorSpec struct {
	Name       string
	Address    string
	Port       uint16
	Connection *grpc.ClientConn
}

func init() {
	RegisterProvisioner(providedProvisionerKind, &ProvidedProvisioner{})
}

func (p ProvidedProvisioner) Setup(_ context.Context, spec map[string]any) (map[string]InjectorClient, error) {
	clients := map[string]InjectorClient{}

	// Check if configuration key exists
	injectorSpecs, ok := spec[injectorsSpecsKeys]
	if !ok {
		return nil, errors.New(fmt.Sprintf("invalid spec for 'provided' provisioner, key %s not found", injectorsSpecsKeys))
	}

	// Check if key content is a slice
	specKind := reflect.TypeOf(injectorSpecs).Kind()
	if specKind != reflect.Slice {
		return nil, errors.New(fmt.Sprintf("invalid type for provisioner specs: expected slice, found %s", specKind.String()))
	}

	specAsSlice := injectorSpecs.([]map[string]any)
	for _, rawInjectorSpec := range specAsSlice {
		injectorSpec := ProvidedInjectorSpec{}
		err := mapstructure.Decode(rawInjectorSpec, &injectorSpec)
		if err != nil {
			return nil, err
		}

		// Connect to provided injector
		connection, err := grpc.Dial(injectorSpec.Address)
		if err != nil {
			return nil, err
		}
		injectorSpec.Connection = connection
		client := NewInjectorClient(connection)

		clients[injectorSpec.Name] = client
		p.injectors = append(p.injectors, injectorSpec)
	}

	return clients, nil
}

func (p ProvidedProvisioner) TearDown(_ context.Context) error {
	for _, injector := range p.injectors {
		err := injector.Connection.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
