package flagx

import (
	"fmt"
	"sync"
)

var (
	mux              sync.Mutex
	sourceNames      []string
	sourceArgs       = NewSource("args")
	sourceEnvrioment = NewSource("envrioment")
)

type Source interface {
	order() int
	String() string
}

type source struct {
	index int
	name  string
}

func (s source) order() int {
	return s.index
}

func (s source) String() string {
	return fmt.Sprintf("[%d]%s", s.index, s.name)
}

func NewSource(name string) Source {
	mux.Lock()
	defer mux.Unlock()
	for i := range sourceNames {
		if sourceNames[i] == name {
			panic(fmt.Errorf("dumplicate source name: %s", name))
		}
	}
	sourceNames = append(sourceNames, name)
	return source{index: len(sourceNames) - 1, name: name}
}
