//go:build go1.19

package di

import (
	"encoding"
	"flag"
)

func textVar(fs *flag.FlagSet, p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) {
	fs.TextVar(p, name, value, usage)
}
