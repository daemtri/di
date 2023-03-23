package flagx

import (
	"flag"
)

// IsSet 判断flag.FlagSet中某个key是否已经设置过了
func IsSet(fs *flag.FlagSet, name string) bool {
	isSet := false
	fs.Visit(func(of *flag.Flag) {
		if name == of.Name {
			isSet = true
		}
	})
	return isSet
}
