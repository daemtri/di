package box

import (
	"context"

	"github.com/daemtri/di/box/config/jsonconfig"
	"github.com/daemtri/di/box/flagx"
	"github.com/daemtri/di/box/validate"
)

// ConfigLoader defined a interface to load config
type ConfigLoader interface {
	// Load load config from config file or other source,
	// setter is a callback function to set config, it is goroutine safe,
	// that means you can call setter immediately,
	// then start a goroutine to watch config change until context done,
	// and call setter again when config changed
	Load(ctx context.Context, setter func([]jsonconfig.ConfigItem)) error
}

type ConfigLoaderBuilder struct {
	OriginBuilder Builder[ConfigLoader] `flag:""`
	source        flagx.Source
}

func (cb *ConfigLoaderBuilder) ValidateFlags() error {
	return validate.Struct(cb.OriginBuilder)
}

func (cb *ConfigLoaderBuilder) Build(ctx context.Context) (ConfigLoader, error) {
	return cb.OriginBuilder.Build(ctx)
}
