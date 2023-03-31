package jsonconfig

import (
	"encoding/csv"
	"flag"
	"fmt"
	"strings"

	"github.com/daemtri/di/box"
	"github.com/tidwall/gjson"
)

func convertToStringSlice(arr []gjson.Result) []string {
	ret := make([]string, 0, len(arr))
	for i := range arr {
		switch arr[i].Type {
		case gjson.String:
			ret = append(ret, arr[i].Str)
		case gjson.Number:
			ret = append(ret, arr[i].Raw)
		case gjson.False, gjson.True:
			ret = append(ret, arr[i].Raw)
		default:
			panic(fmt.Errorf("数组类型Item只支持bool,number,string,错误的值: %s", arr[i]))
		}
	}
	return ret
}

func parseJSON(fs *flag.FlagSet, prefix string, result *gjson.Result) {
	result.ForEach(func(key, value gjson.Result) bool {
		flagKey := prefix + key.Str
		switch value.Type {
		case gjson.JSON:
			if value.IsArray() {
				var sb strings.Builder
				wt := csv.NewWriter(&sb)
				wt.Write(convertToStringSlice(value.Array()))
				fs.Set(flagKey, sb.String())
			} else if value.IsObject() {
				parseJSON(fs, flagKey+"-", &value)
			}
		default:
			fs.Set(flagKey, value.String())
		}
		return true
	})
}

func ParseJSON(fs *flag.FlagSet, json string) error {
	result := gjson.Parse(json)
	if !result.IsObject() {
		return fmt.Errorf("参数解析失败")
	}
	parseJSON(fs, "", &result)
	return nil
}

func parseJSONToKeyValue(prefix string, result *gjson.Result) []box.ConfigItem {
	kv := make([]box.ConfigItem, 0, 10)
	result.ForEach(func(key, value gjson.Result) bool {
		flagKey := prefix + key.Str
		switch value.Type {
		case gjson.JSON:
			if value.IsArray() {
				var sb strings.Builder
				wt := csv.NewWriter(&sb)
				if err := wt.Write(convertToStringSlice(value.Array())); err != nil {
					panic(err)
				}
				wt.Flush()
				kv = append(kv, box.ConfigItem{
					Key:   flagKey,
					Value: sb.String(),
				})
			} else if value.IsObject() {
				kv = append(kv, parseJSONToKeyValue(flagKey+"-", &value)...)
			}
		default:
			kv = append(kv, box.ConfigItem{
				Key:   flagKey,
				Value: value.String(),
			})
		}
		return true
	})
	return kv
}

func ParseJSONToKeyValue(json string) ([]box.ConfigItem, error) {
	result := gjson.Parse(json)
	if !result.IsObject() {
		return nil, fmt.Errorf("参数解析失败")
	}

	return parseJSONToKeyValue("", &result), nil
}
