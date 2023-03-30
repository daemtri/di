package box

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/daemtri/di"
	"github.com/daemtri/di/box/flagx"
	"github.com/daemtri/di/container"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

type Builder[T any] interface {
	Build(ctx context.Context) (T, error)
}

type buildOptions struct {
	init          func(ctx context.Context) error
	configLoaders []Builder[ConfigLoader]
}
type BuildOption interface {
	apply(o *buildOptions)
}

type buildOptionsFunc func(o *buildOptions)

func (of buildOptionsFunc) apply(o *buildOptions) { of(o) }

type multiInit []InitFunc

func (m multiInit) init(ctx context.Context) error {
	for i := range m {
		if err := m[i](ctx); err != nil {
			return fmt.Errorf("execution of (%T) returned error: %w", m[i], err)
		}
	}
	return nil
}

// InitFunc 初始化函数
type InitFunc func(context.Context) error

func UseInit(fn ...InitFunc) BuildOption {
	var initFunc InitFunc
	if len(fn) == 1 {
		initFunc = fn[0]
	} else {
		initFunc = multiInit(fn).init
	}
	return buildOptionsFunc(func(o *buildOptions) {
		o.init = initFunc
	})
}

func UseConfigLoader(cl ...Builder[ConfigLoader]) BuildOption {
	return buildOptionsFunc(func(o *buildOptions) {
		if o.configLoaders == nil {
			o.configLoaders = make([]Builder[ConfigLoader], 0, len(cl))
		}
		o.configLoaders = append(o.configLoaders, cl...)
	})
}

// Build 递归构建对象以及对象的依赖
// 注意：Build 只能被调用一次，否则会引发重复注册配置文件以及重复解析参数的Panic
func Build[T any](ctx context.Context, opts ...BuildOption) (T, error) {
	opt := &buildOptions{}
	for i := range opts {
		opts[i].apply(opt)
	}

	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("godotenv.Load failed", "err", err)
		}
	}

	for i := range opt.configLoaders {
		provide[ConfigLoader](&ConfigLoaderBuilder{OriginBuilder: opt.configLoaders[i], source: flagx.NewSource("x")})
	}

	printConfig := nfs.FlagSet().Bool("print-config", false, "print configuration information")
	nfs.BindFlagSet(flag.CommandLine, envPrefix)
	nfs.FlagSet().StringVar(&configFile, "config", configFile, "configuration file path")
	if items, err := configLoadFunc(configFile); err != nil {
		slog.Warn("local configuration file not found", "error", err.Error())
	} else if err := SetConfig(items, flagx.SourceLocal); err != nil {
		slog.Warn("set local configuration failed", "error", err.Error())
	}

	if *printConfig {
		err := EncodeFlags(os.Stdout)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stdout, "EncodeFlags error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if opt.init != nil {
		Provide[*initializer[T]](&initializer[T]{
			beforeFunc: opt.init,
		})
		nfsIsParsed = true
		agent, err := di.Build[*initializer[T]](ctx)
		if err != nil {
			return emptyValue[T](), err
		}
		return agent.instance, nil
	}
	nfsIsParsed = true

	return di.Build[T](ctx)
}

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
