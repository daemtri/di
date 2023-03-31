package yamlconfig

import (
	"context"
	"fmt"
	"os"

	"github.com/daemtri/di/box"
	"github.com/daemtri/di/box/config/jsonconfig"
	"sigs.k8s.io/yaml"
)

func Init() box.BuildOption {
	return box.UseConfigLoader("", &ConfigLoader{})
}

type ConfigLoader struct {
	Configfile string `flag:"config" default:"./config.yaml" usage:"配置文件路径"`
}

func (c *ConfigLoader) Load(ctx context.Context, setter func([]box.ConfigItem)) error {
	items, err := Load(c.Configfile)
	if err != nil {
		return err
	}
	setter(items)
	return nil
}

func Load(configFile string) ([]box.ConfigItem, error) {
	yamlRawConfig, err := os.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("配置文件读取失败: %w", err)
		}
		return nil, err
	}
	if len(yamlRawConfig) == 0 {
		return nil, nil
	}
	jsonRawConfig, err := yaml.YAMLToJSON(yamlRawConfig)
	if err != nil {
		return nil, fmt.Errorf("配置文件解析失败: %w", err)
	}
	items, err := jsonconfig.ParseJSONToKeyValue(string(jsonRawConfig))
	if err != nil {
		return nil, fmt.Errorf("配置文件解析失败: %w", err)
	}
	return items, nil
}
