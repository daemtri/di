package flagx

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// NamedFlagSets 存储了命名参数集合
type NamedFlagSets struct {
	// order 为命名参数的名称排序
	order []string
	// flagSets 存储所有所有参数集合，key为名称
	flagSets map[string]*flag.FlagSet
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

func (nfs *NamedFlagSets) Parse() {
	nfs.VisitAll(func(p string, f *flag.Flag) {
		name := p + "-" + f.Name
		flag.Var(f.Value, name, f.Usage)
		flag.Lookup(name).DefValue = f.DefValue
	})
	flag.Parse()
}

func envKey(prefix string, name string) (key string) {
	if prefix == "" {
		return strings.ReplaceAll(strings.ToUpper(name), "-", "_")
	}
	return strings.ReplaceAll(strings.ToUpper(prefix+"_"+name), "-", "_")

}

type envFlag struct {
	envKey  string
	flagKey string
}

func (nfs *NamedFlagSets) BindEnvAndFlags(envPrefix string, fs *flag.FlagSet) {
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
	for i := range envFlags {
		envValue, ok := os.LookupEnv(envFlags[i].envKey)
		if !ok {
			continue
		}
		if err := fs.Set(envFlags[i].flagKey, envValue); err != nil {
			panic(err)
		}
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}
