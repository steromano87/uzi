package injector

import "context"

type Provisioner interface {
	Setup(ctx context.Context) (map[string]InjectorClient, error)
	TearDown(ctx context.Context) error
}
