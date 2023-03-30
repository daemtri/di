package box

import (
	"flag"
	"fmt"
	"reflect"
	"strings"

	"github.com/daemtri/di"
	"github.com/daemtri/di/box/flagvar"
)

func init() {

	// 补全基础类型
	di.RegisterFlagBinder(reflect.TypeOf(uint8(1)), bindBaseFlag[uint8])
	di.RegisterFlagBinder(reflect.TypeOf(uint16(1)), bindBaseFlag[uint16])
	di.RegisterFlagBinder(reflect.TypeOf(uint32(1)), bindBaseFlag[uint32])
	di.RegisterFlagBinder(reflect.TypeOf(int8(1)), bindBaseFlag[int8])
	di.RegisterFlagBinder(reflect.TypeOf(int16(1)), bindBaseFlag[int16])
	di.RegisterFlagBinder(reflect.TypeOf(int32(1)), bindBaseFlag[int32])
	di.RegisterFlagBinder(reflect.TypeOf(float32(1)), bindBaseFlag[float32])

	// 支持切片类型
	di.RegisterFlagBinder(reflect.TypeOf([]uint{}), bindSliceFlag[uint])
	di.RegisterFlagBinder(reflect.TypeOf([]uint8{}), bindSliceFlag[uint8])
	di.RegisterFlagBinder(reflect.TypeOf([]uint16{}), bindSliceFlag[uint16])
	di.RegisterFlagBinder(reflect.TypeOf([]uint32{}), bindSliceFlag[uint32])
	di.RegisterFlagBinder(reflect.TypeOf([]uint64{}), bindSliceFlag[uint64])
	di.RegisterFlagBinder(reflect.TypeOf([]int{}), bindSliceFlag[int])
	di.RegisterFlagBinder(reflect.TypeOf([]int8{}), bindSliceFlag[int8])
	di.RegisterFlagBinder(reflect.TypeOf([]int16{}), bindSliceFlag[int16])
	di.RegisterFlagBinder(reflect.TypeOf([]int32{}), bindSliceFlag[int32])
	di.RegisterFlagBinder(reflect.TypeOf([]int64{}), bindSliceFlag[int64])
	di.RegisterFlagBinder(reflect.TypeOf([]float32{}), bindSliceFlag[float32])
	di.RegisterFlagBinder(reflect.TypeOf([]float64{}), bindSliceFlag[float64])
	di.RegisterFlagBinder(reflect.TypeOf([]bool{}), bindSliceFlag[bool])
	di.RegisterFlagBinder(reflect.TypeOf([]string{}), bindSliceFlag[string])

	// 支持map类型
	di.RegisterFlagBinder(reflect.TypeOf(map[string]uint{}), bindMapFlag[uint])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]uint8{}), bindMapFlag[uint8])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]uint16{}), bindMapFlag[uint16])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]uint32{}), bindMapFlag[uint32])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]uint64{}), bindMapFlag[uint64])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]int{}), bindMapFlag[int])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]int8{}), bindMapFlag[int8])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]int16{}), bindMapFlag[int16])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]int32{}), bindMapFlag[int32])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]int64{}), bindMapFlag[int64])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]float32{}), bindMapFlag[float32])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]float64{}), bindMapFlag[float64])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]bool{}), bindMapFlag[bool])
	di.RegisterFlagBinder(reflect.TypeOf(map[string]string{}), bindMapFlag[string])
}

func bindBaseFlag[T flagvar.BaseType](fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
	ptr := (*T)(value.Addr().UnsafePointer())
	defValue := *ptr
	if value.IsZero() && def != "" {
		val, err := flagvar.StringToKind[value.Kind()](def)
		if err != nil {
			return fmt.Errorf("参数%s默认值解析错误: %w", name, err)
		}
		defValue = val.(T)
		*ptr = defValue
	}
	fs.Var(flagvar.Base(ptr, defValue), name, usage)
	return nil
}

func bindSliceFlag[T flagvar.BaseType](fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
	ptr := (*[]T)(value.Addr().UnsafePointer())
	defValue := *ptr
	if value.IsZero() && def != "" {
		var err error
		defValue, err = flagvar.StringSliceTo[T](strings.Split(def, ","))
		if err != nil {
			return fmt.Errorf("参数%s默认值解析错误: %w", name, err)
		}
		*ptr = defValue
	}
	fs.Var(flagvar.Slice(ptr, defValue...), name, usage)
	return nil
}

func bindMapFlag[T flagvar.BaseType](fs *flag.FlagSet, value reflect.Value, name, def, usage string) error {
	ptr := (*map[string]T)(value.Addr().UnsafePointer())
	if value.IsZero() && def != "" {
		ssm := flagvar.StringSliceToStringStringMap(strings.Split(def, ","))
		defValue, err := flagvar.StringStringMapTo[T](ssm)
		if err != nil {
			return fmt.Errorf("参数%s默认值解析错误: %w", name, err)
		}
		*ptr = defValue
	}
	fs.Var(flagvar.StringMap(ptr), name, usage)
	return nil
}
