package box

import (
	"context"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"github.com/daemtri/di"
	"github.com/daemtri/di/box/flagx"
	"github.com/daemtri/di/container"
	"golang.org/x/exp/slog"
)

type Builder[T any] interface {
	Build(ctx context.Context) (T, error)
}

type buildOptions struct {
	inits         []namedInitFunc
	configLoaders []*configLoaderBuilder
}
type BuildOption interface {
	apply(o *buildOptions)
}

type buildOptionsFunc func(o *buildOptions)

func (of buildOptionsFunc) apply(o *buildOptions) { of(o) }

func UseInit(name string, fn InitFunc) BuildOption {
	return buildOptionsFunc(func(o *buildOptions) {
		if o.inits == nil {
			o.inits = []namedInitFunc{}
		}
		o.inits = append(o.inits, namedInitFunc{
			InitFunc: fn,
			name:     name,
		})
	})
}

// UseConfigLoader register a config loader of the given name
// if name is empty, the loader's type name will be used as the name
// the loader is ordered by the order of the UseConfigLoader call and
// the earlier added loader has a higher priority.
// all loader will be invoked  before all init function and build function
func UseConfigLoader(name string, loader ConfigLoader) BuildOption {
	return buildOptionsFunc(func(o *buildOptions) {
		if o.configLoaders == nil {
			o.configLoaders = make([]*configLoaderBuilder, 0, 1)
		}
		sourceName := name
		if sourceName == "" {
			sourceName = strings.TrimPrefix(reflect.TypeOf(loader).String(), "*")
		}
		o.configLoaders = append(o.configLoaders, &configLoaderBuilder{
			ConfigLoader: loader,
			source:       flagx.NewSource(sourceName),
			name:         name,
		})
	})
}

// Build 递归构建对象以及对象的依赖
// 注意：Build 只能被调用一次，否则会引发重复注册配置文件以及重复解析参数的Panic
func Build[T any](ctx context.Context, opts ...BuildOption) (T, error) {
	defer func() {
		nfsIsParsed = true
	}()
	opt := &buildOptions{}
	for i := range opts {
		opts[i].apply(opt)
	}

	for i := range opt.configLoaders {
		provide[*configLoaderBuilder](opt.configLoaders[i],
			WithFlags(opt.configLoaders[i].name),
			WithName(opt.configLoaders[i].name),
		)
	}

	Provide[*initializer[T]](&initializer[T]{
		beforeFuncs: opt.inits,
	}, WithOptional[*configLoaderBuilder](func(name string, err error) {
		if err != nil {
			slog.Warn("load config failed", "name", name, "error", err)
		}
	}))
	agent, err := di.Build[*initializer[T]](ctx)
	if err != nil {
		return emptyValue[T](), err
	}
	return agent.instance, nil
}

type All[T any] []T

func Invoke[T any](ctx context.Context) T {
	return container.Invoke[T](ctx)
}

// Runable defined a object that can be run
type Runable interface {
	Run(ctx context.Context) error
}

// Bootstrap use to build and run a object
// it will block until the object is stopped
func Bootstrap[T Runable](opts ...BuildOption) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()
	app, err := Build[T](ctx, opts...)
	if err != nil {
		return err
	}
	return app.Run(ctx)
}
