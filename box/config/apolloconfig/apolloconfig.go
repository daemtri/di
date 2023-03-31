// Package apolloconfig APOLLO配置加载器
package apolloconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/daemtri/di/box"
	"github.com/shima-park/agollo"
	"golang.org/x/exp/slog"
)

func Init() box.BuildOption {
	return box.UseConfigLoader("apollo", NewConfigLoader())
}

type ConfigLoader struct {
	AppId       string `flag:"appid" default:"" usage:"apollo appid" validate:"required"`
	Cluster     string `flag:"cluster" default:"" usage:"apollo cluster" validate:"required"`
	Addr        string `flag:"addr" default:"" usage:"apollo addr" validate:"required"`
	Namespace   string `flag:"namespace" default:"" usage:"apollo namespace, split by ',', previous namespace will override latest namespace" validate:"required"`
	CachePath   string `flag:"cache" default:"./configs/apollo_system.json" usage:"apollo cache path"`
	SyncTimeout int    `flag:"sync_timeout" default:"5" usage:"apollo sync timeout"`
	Secret      string `flag:"secret" default:"" usage:"apollo secret" validate:"required"`

	// namespaceList store all namespace by split Namespace
	namespaceList []string
	namespacesMap map[string]*namespace
	// configKeyNamespace store namespace by config key
	configKeyNamespace map[string]*namespace
	client             agollo.Agollo
	set                func([]box.ConfigItem)
}

func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		namespacesMap:      make(map[string]*namespace),
		configKeyNamespace: make(map[string]*namespace),
	}
}

func (cl *ConfigLoader) Load(ctx context.Context, set func([]box.ConfigItem)) error {
	if !strings.HasSuffix(cl.CachePath, ".json") {
		cl.CachePath = cl.CachePath + ".json"
	}
	namespaces := strings.Split(cl.Namespace, ",")
	for i := range namespaces {
		cl.namespaceList = append(cl.namespaceList, namespaces[i])
		cl.namespacesMap[namespaces[i]] = &namespace{name: namespaces[i], order: i}
	}
	cl.set = set

	if err := cl.connect(); err != nil {
		return err
	}

	for _, n := range cl.namespaceList {
		if err := cl.loadAndParseByNamespace(n); err != nil {
			return err
		}
	}

	go cl.watch(ctx)
	return nil
}

func (cl *ConfigLoader) setConfigItemsByNamespace(ns string, items []box.ConfigItem) {
	namespace := cl.namespacesMap[ns]
	itemsCanset := make([]box.ConfigItem, 0, len(items))
	for _, item := range items {
		ns, ok := cl.configKeyNamespace[item.Key]
		if ok && ns.order < namespace.order {
			continue
		}
		itemsCanset = append(itemsCanset, item)
		cl.configKeyNamespace[item.Key] = namespace
	}
	cl.set(itemsCanset)
}

func (cl *ConfigLoader) connect() error {
	client, err := agollo.New(
		cl.Addr,
		cl.AppId,
		agollo.Cluster(cl.Cluster),
		agollo.AccessKey(cl.Secret),
		agollo.AutoFetchOnCacheMiss(),
		agollo.PreloadNamespaces(strings.Split(cl.Namespace, ",")...),
		agollo.BackupFile(cl.CachePath),
		agollo.FailTolerantOnBackupExists(),
		agollo.WithLogger(&ApolloLogger{base: slog.Default()}),
	)
	if err != nil {
		return err
	}
	cl.client = client
	return nil
}

func (cl *ConfigLoader) loadAndParseByNamespace(ns string) error {
	cfg := cl.client.GetNameSpace(ns)
	if len(cfg) == 0 {
		return nil
	}
	items, err := cl.parse(cfg)
	if err != nil {
		return err
	}
	cl.setConfigItemsByNamespace(ns, items)
	return nil
}

func (cl *ConfigLoader) parse(cache agollo.Configurations) ([]box.ConfigItem, error) {
	var items = make([]box.ConfigItem, 0, len(cache))
	for key, value := range cache {
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("apollo config value is not string, key: %s", key)
		}
		items = append(items, box.ConfigItem{Key: key, Value: strValue})
	}
	return items, nil
}

func (cl *ConfigLoader) watch(ctx context.Context) {
	// TODO: 这里以后需要注意namespace优先级
	resp := cl.client.Watch()
	for e := range resp {
		slog.Debug("config has changed", "namespace", e.Namespace, "e", e.Changes)
		items := make([]box.ConfigItem, 0, len(e.Changes))
		for _, item := range e.Changes {
			if item.Value == nil {
				continue
			}
			strValue, ok := item.Value.(string)
			if !ok {
				slog.Error("apollo config value is not string", "key", item.Key)
				continue
			}
			items = append(items, box.ConfigItem{Key: item.Key, Value: strValue})
		}
		cl.setConfigItemsByNamespace(e.Namespace, items)
	}
}

type ApolloLogger struct {
	base *slog.Logger
}

func (al *ApolloLogger) Log(kv ...any) {
	al.base.Info("apollo system", kv...)
}
