package box

import (
	"flag"
	"github.com/tidwall/sjson"
	"io"
	"reflect"
	"sigs.k8s.io/yaml"
	"strings"
)

type getter interface {
	Get() any
}

// EncodeFlags 保存已经加载的配置到文件中
// format: yaml
func EncodeFlags(w io.Writer) (err error) {
	// nfs 使用的是flag.Commandline
	if !flag.Parsed() {
		flag.Parse()
	}
	jsonValue := `{}`

	nfs.VisitAll(func(p string, f *flag.Flag) {
		if p == "" && (f.Name == "config" || f.Name == "print-config") {
			return
		}
		strValue := f.Value.String()
		if strValue == "" {
			return
		}
		var value any = strValue
		valueGetter, ok := f.Value.(flag.Getter)
		if ok {
			anyValue := valueGetter.Get()
			gt, ok := anyValue.(getter)
			if ok {
				value = gt.Get()
			} else if reflect.TypeOf(f.Value).String() != "*flag.durationValue" {
				value = anyValue
			}
		}
		key := strings.ReplaceAll(f.Name, "-", ".")
		if p != "" {
			key = p + "." + key
		}

		jsonValue, err = sjson.Set(jsonValue, key, value)
		if err != nil {
			return
		}
	})
	yamlValue, err := yaml.JSONToYAML([]byte(jsonValue))
	_, err = w.Write(yamlValue)
	return
}
