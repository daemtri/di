package box

import (
	"errors"
	"flag"
	"fmt"

	"github.com/daemtri/di"
	"github.com/daemtri/di/box/flagx"
)

var (
	defaultRegistrar = di.GetRegistry()
	nfs              = flagx.NewNamedFlagSets()
	nfsIsParsed      bool
	envPrefix        = "GF"
)

// Default 返回默认di.Registry
func Default() di.Registry {
	return defaultRegistrar
}

func SetEnvPrefix(prefix string) {
	envPrefix = prefix
}

func FlagSet(name ...string) *flag.FlagSet {
	return nfs.FlagSet(name...)
}

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
func SetConfig(items []ConfigItem, source flagx.Source) error {
	var errs error
	for _, item := range items {
		if err := nfs.Set(item.Key, item.Value, source); err != nil {
			errs = errors.Join(errs, fmt.Errorf("配置变更失败: key=%s,value=%s,error=%s", item.Key, item.Value, err))
		}
	}
	err2 := defaultRegistrar.ValidateFlags()
	if err2 != nil {
		errs = errors.Join(errs, err2)
	}
	err3 := retrofit()
	if err3 != nil {
		errs = errors.Join(errs, err3)
	}
	return errs
}
