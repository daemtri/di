package di

import (
	"flag"
)

type FlagSetWatcher struct {
	oldFlagsKeyValues map[string]string
	fs                *flag.FlagSet
}

func (fsw *FlagSetWatcher) WatchFlags(fs *flag.FlagSet) {
	if fsw.oldFlagsKeyValues == nil {
		fsw.oldFlagsKeyValues = make(map[string]string)
	}
	fsw.fs = fs
}

func (fsw *FlagSetWatcher) FlagsChanged() bool {
	changed := false
	fsw.fs.VisitAll(func(f *flag.Flag) {
		if len(fsw.oldFlagsKeyValues) > 0 && fsw.oldFlagsKeyValues[f.Name] != f.Value.String() {
			changed = true
		}
		fsw.oldFlagsKeyValues[f.Name] = f.Value.String()
	})
	return changed
}
