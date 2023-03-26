package box

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/daemtri/di"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

type BBuilder[T any] interface {
	Build(ctx context.Context) (T, error)
}

type bBuilderFunc[T any] struct {
	BBuilder[T]
}

func (b *bBuilderFunc[T]) Build(ctx Context) (T, error) {
	return b.BBuilder.Build(ctx.Unwrap())
}

type Builder[T any] interface {
	Build(ctx Context) (T, error)
}

type buildOptions struct {
	name string
	init func(ctx Context) error
}
type BuildOption interface {
	apply(o *buildOptions)
}

type buildOptionsFunc func(o *buildOptions)

func (of buildOptionsFunc) apply(o *buildOptions) { of(o) }

func Select(name string) BuildOption {
	return buildOptionsFunc(func(o *buildOptions) {
		o.name = name
	})
}

type multiInit []func(Context) error

func (m multiInit) init(ctx Context) error {
	for i := range m {
		if err := m[i](ctx); err != nil {
			return fmt.Errorf("运行(%T)返回错误: %w", m[i], err)
		}
	}
	return nil
}

func UseInit(fn ...func(Context) error) BuildOption {
	var initFunc func(Context) error
	if len(fn) == 1 {
		initFunc = fn[0]
	} else {
		initFunc = multiInit(fn).init
	}
	return buildOptionsFunc(func(o *buildOptions) {
		o.init = initFunc
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
			slog.Warn("godotenv.Load失败", "err", err)
		}
	}
	nfs.FlagSet().StringVar(&configFile, "config", configFile, "配置文件地址")
	printConfig := nfs.FlagSet().Bool("print-config", false, "打印配置信息")
	nfs.BindEnvAndFlags(envPrefix, flag.CommandLine)
	if err := configLoadFunc(configFile, flag.CommandLine); err != nil {
		slog.Warn("配置文件加载错误", "error", err.Error())
	}
	if *printConfig {
		err := EncodeFlags(os.Stdout)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stdout, "EncodeFlags error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	reg := defaultRegistrar

	if opt.name != "" {
		reg = defaultRegistrar.Named(opt.name)
	}
	if opt.init != nil {
		Provide[*initializer[T]](&initializer[T]{
			beforeFunc: opt.init,
		})
		nfsIsParsed = true
		agent, err := di.Build[*initializer[T]](reg, ctx)
		if err != nil {
			return emptyValue[T](), err
		}
		return agent.instance, nil
	}
	nfsIsParsed = true

	return di.Build[T](reg, ctx)
}

func Must[T any](ctx Context) T {
	return di.Must[T](ctx)
}

func MustAll[T any](ctx Context) map[string]T {
	return di.MustAll[T](ctx)
}

// Exists 判断类型T是否已经在容器内提供了
func Exists[T any](ctx Context) bool {
	return di.Exists[T](ctx)
}
