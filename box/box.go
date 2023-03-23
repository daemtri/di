package box

import (
	"errors"
	"flag"
	"fmt"

	"github.com/daemtri/di"
	config "github.com/daemtri/di/box/config/jsonconfig"
	"github.com/daemtri/di/box/config/yamlconfig"
	"github.com/daemtri/di/box/flagx"
)

var (
	defaultRegistrar = di.NewRegistry()

	nfs         = flagx.NamedFlagSets{}
	nfsIsParsed bool

	envPrefix = "GF"

	configFile     = "./config.yaml"
	configLoadFunc = yamlconfig.Load
)

// Default 返回默认di.Registry
func Default() di.Registry {
	return defaultRegistrar
}

func SetEnvPrefix(prefix string) {
	envPrefix = prefix
}

func SetConfigLoader(defaultFile string, fn func(configFile string, fs *flag.FlagSet) error) {
	configFile = defaultFile
	if fn != nil {
		configLoadFunc = fn
	}
}

func FlagSet(name ...string) *flag.FlagSet {
	return nfs.FlagSet(name...)
}

type Context = di.Context

// Retrofiter 定义了一个可以重新构建对象的接口
type Retrofiter interface {
	Retrofit() error
}

// Retrofit 遍历reg种所有已经构建完成的对象
// 如果builder实现了Retrofit，则触发一次Retrofit
func retrofit() error {
	var err error
	defaultRegistrar.Visit(func(v di.Value) {
		if v.Instance() != nil {
			if r, ok := v.Builder().(Retrofiter); ok {
				err2 := r.Retrofit()
				if err2 != nil {
					err = errors.Join(err, err2)
				}
			}
		}
	})
	return err
}

// SetConfig 设置配置
func SetConfig(items []config.ConfigItem) error {
	var err error
	for _, item := range items {
		if err1 := flag.CommandLine.Set(item.Key, item.Value); err != nil {
			err = errors.Join(err, fmt.Errorf("配置变更失败: key=%s,value=%s,error=%s", item.Key, item.Value, err1))
		}
	}
	err2 := defaultRegistrar.ValidateFlags()
	if err2 != nil {
		err = errors.Join(err, err2)
	}
	err3 := retrofit()
	if err3 != nil {
		err = errors.Join(err, err3)
	}
	return err
}
