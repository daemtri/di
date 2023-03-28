package di

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var (
	flagKindBinds = map[reflect.Kind]FlagBind{
		reflect.String: func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
			ptr := (*string)(value.Addr().UnsafePointer())
			if value.IsZero() && def != "" {
				*ptr = def
			}
			fs.StringVar(ptr, name, *ptr, usage)
			return nil
		},
		reflect.Int: func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
			ptr := (*int)(value.Addr().UnsafePointer())
			if value.IsZero() && def != "" {
				if defValue, err := strconv.ParseInt(def, 10, strconv.IntSize); err != nil {
					return err
				} else {
					*ptr = int(defValue)
				}
			}
			fs.IntVar(ptr, name, *ptr, usage)
			return nil
		},
		reflect.Uint: func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
			ptr := (*uint)(value.Addr().UnsafePointer())
			if value.IsZero() && def != "" {
				if defValue, err := strconv.ParseUint(def, 10, strconv.IntSize); err != nil {
					return err
				} else {
					*ptr = uint(defValue)
				}
			}
			fs.UintVar(ptr, name, *ptr, usage)
			return nil
		},
		reflect.Int64: func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
			ptr := (*int64)(value.Addr().UnsafePointer())
			if value.IsZero() && def != "" {
				if defValue, err := strconv.ParseInt(def, 10, 64); err != nil {
					return err
				} else {
					*ptr = defValue
				}
			}
			fs.Int64Var(ptr, name, *ptr, usage)
			return nil
		},
		reflect.Uint64: func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
			ptr := (*uint64)(value.Addr().UnsafePointer())
			if value.IsZero() && def != "" {
				if defValue, err := strconv.ParseUint(def, 10, 64); err != nil {
					return err
				} else {
					*ptr = defValue
				}
			}
			fs.Uint64Var(ptr, name, *ptr, usage)
			return nil
		},
		reflect.Bool: func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
			ptr := (*bool)(value.Addr().UnsafePointer())
			if value.IsZero() && def != "" {
				if defValue, err := strconv.ParseBool(def); err != nil {
					return err
				} else {
					*ptr = defValue
				}
			}
			fs.BoolVar(ptr, name, *ptr, usage)
			return nil
		},
		reflect.Float64: func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
			ptr := (*float64)(value.Addr().UnsafePointer())
			if value.IsZero() && def != "" {
				if defValue, err := strconv.ParseFloat(def, 64); err != nil {
					return err
				} else {
					*ptr = defValue
				}
			}
			fs.Float64Var(ptr, name, *ptr, usage)
			return nil
		},
	}

	// flagTypeBinds 注册类型对应的绑定
	flagTypeBinds = map[reflect.Type]FlagBind{
		reflect.TypeOf(time.Duration(0)): func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
			ptr := (*time.Duration)(value.Addr().UnsafePointer())
			if value.IsZero() && def != "" {
				if defValue, err := time.ParseDuration(def); err != nil {
					return err
				} else {
					*ptr = defValue
				}
			}
			fs.DurationVar(ptr, name, *ptr, usage)
			return nil
		},
	}
)

type FlagBind func(fs *flag.FlagSet, value reflect.Value, name, def, usage string) error

func RegisterFlagBinder(typ reflect.Type, bind FlagBind) {
	if _, ok := flagTypeBinds[typ]; ok {
		panic(fmt.Errorf("flag type :%s has been bound", typ))
	}
	flagTypeBinds[typ] = bind
}
