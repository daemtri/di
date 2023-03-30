package box

import (
	"context"

	"github.com/daemtri/di/box/config/jsonconfig"
)

type ConfigLoader interface {
	Load(ctx context.Context, setter func([]jsonconfig.ConfigItem)) error
}
