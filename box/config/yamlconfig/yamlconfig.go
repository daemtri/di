package yamlconfig

import (
	"context"
	"fmt"
	"os"

	"github.com/daemtri/di/box"
	"github.com/daemtri/di/box/config/jsonconfig"
	"sigs.k8s.io/yaml"
)

type ConfigLoader struct {
	Configfile string `flag:"config" default:"./config.yaml" usage:"配置文件路径"`
}

func (c *ConfigLoader) Build(ctx context.Context) (box.ConfigLoader, error) {
	return c, nil
}

func (c *ConfigLoader) Load(ctx context.Context, setter func([]jsonconfig.ConfigItem)) error {
	items, err := Load(c.Configfile)
	if err != nil {
		return fmt.Errorf("配置文件加载失败: %w", err)
	}
	setter(items)
	return nil
}

func Load(configFile string) ([]jsonconfig.ConfigItem, error) {
	yamlRawConfig, err := os.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("配置文件读取失败: %w", err)
		}
		return nil, err
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
