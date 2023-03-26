package wire

import (
	"reflect"
	_ "unsafe"

	"github.com/daemtri/di"
	"github.com/daemtri/di/box"
)

//go:linkname diProvide github.com/daemtri/di.provide
func diProvide(reg di.Registry, typ reflect.Type, flaggerBuilder any, buildFunc func(di.Context) (any, error)) di.Constructor

type ProviderSet struct {
	sets []*injectBuilder
}

func NewSet(providers ...any) ProviderSet {
	ps := ProviderSet{
		sets: make([]*injectBuilder, 0, len(providers)),
	}
	for i := range providers {
		ps.sets = append(ps.sets, Inject(providers[i]))
	}
	return ps
}

func Build(providers ...any) {
	for i := range providers {
		if ps, ok := providers[i].(ProviderSet); ok {
			for j := range ps.sets {
				diProvide(box.Default(), ps.sets[j].pType, ps.sets[j], ps.sets[j].Build)
			}
		} else {
			ib := Inject(providers[i])
			diProvide(box.Default(), ib.pType, ib, ib.Build)
		}
	}
}
