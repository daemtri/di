package box

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

func init() {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("godotenv.Load failed", "err", err)
		}
	}
}

// InitFunc 初始化函数
type InitFunc func(context.Context) error

type namedInitFunc struct {
	InitFunc
	name string
}

type initializer[T any] struct {
	beforeFuncs []namedInitFunc
	instance    T
}

func (it *initializer[T]) Build(ctx context.Context) (*initializer[T], error) {
	// register config loader
	configLoaders := Invoke[All[*configLoaderBuilder]](ctx)

	// parser args and envronment
	printConfig := nfs.FlagSet().Bool("print-config", false, "print configuration information")
	nfs.BindFlagSet(flag.CommandLine, envPrefix)

	// load config from config file or other source
	for i := range configLoaders {
		if err := configLoaders[i].Load(ctx, func(items []ConfigItem) {
			SetConfig(items, configLoaders[i].source)
		}); err != nil {
			return nil, fmt.Errorf("load configuration %s failed: %w", configLoaders[i].source, err)
		}
	}

	// print config
	if *printConfig {
		err := EncodeFlags(os.Stdout)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stdout, "EncodeFlags error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	for i := range it.beforeFuncs {
		if err := it.beforeFuncs[i].InitFunc(ctx); err != nil {
			return nil, fmt.Errorf("execution initFunc %s returned error: %w", it.beforeFuncs[i].name, err)
		}
	}

	it.instance = Invoke[T](ctx)
	return it, nil
}
