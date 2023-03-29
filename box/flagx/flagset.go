package flagx

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Source int

func (s Source) String() string {
	switch s {
	case SourceNone:
		return "none"
	case SourceRemote:
		return "remote"
	case SourceFile:
		return "file"
	case SourceEnvrioment:
		return "env"
	case SourceArgs:
		return "args"
	}
	return fmt.Sprintf("unknown source: %d", s)
}

const (
	SourceNone Source = iota
	SourceRemote
	SourceFile
	SourceEnvrioment
	SourceArgs
)

// NamedFlagSets 存储了命名参数集合
type NamedFlagSets struct {
	// order 为命名参数的名称排序
	order []string
	// flagSets 存储所有所有参数集合，key为名称
	flagSets map[string]*flag.FlagSet

	// keySource 存储所有参数的来源
	keySource map[string]Source

	fs *flag.FlagSet
}

// FlagSet 返回一个以name为名称的flagSet
// 如果不存在，则新建一个，并保存到FlagSets映射，添加排序
func (nfs *NamedFlagSets) FlagSet(name ...string) *flag.FlagSet {
	prefix := ""
	if len(name) > 0 {
		prefix = strings.TrimPrefix(strings.Join(name, "-"), "-")
	}
	if nfs.flagSets == nil {
		nfs.flagSets = map[string]*flag.FlagSet{}
	}
	if _, ok := nfs.flagSets[prefix]; !ok {
		nfs.flagSets[prefix] = flag.NewFlagSet(prefix, flag.ExitOnError)
		nfs.order = append(nfs.order, prefix)
	}
	return nfs.flagSets[prefix]
}

func (nfs *NamedFlagSets) VisitAll(fn func(p string, f *flag.Flag)) {
	for i := range nfs.order {
		prefix := nfs.order[i]
		fs := nfs.flagSets[prefix]
		fs.VisitAll(func(f *flag.Flag) {
			fn(prefix, f)
		})
	}
}

func envKey(prefix string, name string) (key string) {
	if prefix == "" {
		return strings.ReplaceAll(strings.ToUpper(name), "-", "_")
	}
	return strings.ReplaceAll(strings.ToUpper(prefix+"_"+name), "-", "_")

}

// CanSet 判断key是否可以被source设置,如果已经被更高优先级的source设置，则返回false
func (nfs *NamedFlagSets) CanSet(key string, source Source) bool {
	return nfs.keySource[key] <= source
}

type envFlag struct {
	envKey  string
	flagKey string
}

// BindFlagSet 将所有的flag绑定到fs中，并从环境变量中读取
func (nfs *NamedFlagSets) BindFlagSet(fs *flag.FlagSet, envPrefix string) {
	nfs.fs = fs

	envFlags := make([]envFlag, 0, fs.NFlag())
	nfs.VisitAll(func(p string, f *flag.Flag) {
		name := f.Name
		if p != "" {
			name = p + "-" + f.Name
		}
		key := envKey(envPrefix, name)
		envFlags = append(envFlags, envFlag{
			envKey:  key,
			flagKey: name,
		})
		fs.Var(f.Value, name, fmt.Sprintf("%s (env %s)", f.Usage, key))
		fs.Lookup(name).DefValue = f.DefValue
	})
	// parse flags from os.Args
	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
	nfs.keySource = make(map[string]Source, fs.NFlag())
	// record os.Args flags
	fs.Visit(func(f *flag.Flag) {
		nfs.keySource[f.Name] = SourceArgs
	})
	// parse flags from env
	for i := range envFlags {
		if !nfs.CanSet(envFlags[i].flagKey, SourceEnvrioment) {
			continue
		}
		envValue, ok := os.LookupEnv(envFlags[i].envKey)
		if !ok {
			continue
		}
		if err := fs.Set(envFlags[i].flagKey, envValue); err != nil {
			panic(err)
		}
		nfs.keySource[envFlags[i].flagKey] = SourceEnvrioment
	}
}

func (nfs *NamedFlagSets) Set(key string, value string, source Source) error {
	if !nfs.CanSet(key, source) {
		return fmt.Errorf("can not set %s from %s, already set from %s", key, source, nfs.keySource[key])
	}
	if err := nfs.fs.Set(key, value); err != nil {
		return err
	}
	nfs.keySource[key] = source
	return nil
}
