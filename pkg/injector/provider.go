package injector

import (
	"context"
	"github.com/spf13/viper"
)

type Provider interface {
	Init(ctx context.Context, spec *viper.Viper) (Roster, error)
	TearDown(ctx context.Context) error
}
