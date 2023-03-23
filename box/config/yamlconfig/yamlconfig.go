package yamlconfig

import (
	"flag"
	"os"

	"golang.org/x/exp/slog"

	"github.com/daemtri/di/box/config/jsonconfig"
	"sigs.k8s.io/yaml"
)

func Load(configFile string, fs *flag.FlagSet) error {
	settedFlags := map[string]struct{}{}
	fs.Visit(func(f *flag.Flag) {
		settedFlags[f.Name] = struct{}{}
	})

	yamlRawConfig, err := os.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("配置文件读取失败", "file", configFile)
			return nil
		}
		return err
	}
	jsonRawConfig, err := yaml.YAMLToJSON(yamlRawConfig)
	if err != nil {
		return err
	}
	items, err := jsonconfig.ParseJSONToKeyValue(string(jsonRawConfig))
	if err != nil {
		return err
	}
	for i := range items {
		if _, ok := settedFlags[items[i].Key]; ok {
			continue
		}
		f := fs.Lookup(items[i].Key)
		if f != nil {
			// 需要通过fs.Set方式设置值，以使fs的actual生效
			if err := fs.Set(items[i].Key, items[i].Value); err != nil {
				return err
			}
		}
	}
	return nil
}
