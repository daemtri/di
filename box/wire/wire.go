package wire

import (
	"context"
	"reflect"
	_ "unsafe"

	"github.com/daemtri/di"
	"github.com/daemtri/di/box"
)

//go:linkname diProvide github.com/daemtri/di.provide
func diProvide(reg di.Registry, typ reflect.Type, flaggerBuilder any, buildFunc func(context.Context) (any, error))

type ProviderSet struct {
	sets []*anyFunctionBuilder
}

func NewSet(providers ...any) ProviderSet {
	ps := ProviderSet{
		sets: make([]*anyFunctionBuilder, 0, len(providers)),
	}
	for i := range providers {
		ps.sets = append(ps.sets, newAnyFunctionBuilder(providers[i]))
	}
	return ps
}

func Build(providers ...any) {
	for i := range providers {
		if ps, ok := providers[i].(ProviderSet); ok {
			for j := range ps.sets {
				diProvide(box.Default(), ps.sets[j].targetType, ps.sets[j], ps.sets[j].Build)
			}
		} else {
			ib := newAnyFunctionBuilder(providers[i])
			diProvide(box.Default(), ib.targetType, ib, ib.Build)
		}
	}
}
