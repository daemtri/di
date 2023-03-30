package yamlconfig

import (
	"fmt"
	"os"

	"github.com/daemtri/di/box/config/jsonconfig"
	"sigs.k8s.io/yaml"
)

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
