package injector

import "context"

type Provisioner interface {
	Setup(ctx context.Context, spec map[string]any) (map[string]InjectorClient, error)
	TearDown(ctx context.Context) error
}
