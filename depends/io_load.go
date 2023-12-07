package depends

import (
	"fmt"
	"sync"
)

var dependIO sync.Map

type Configger interface {
	LoadConfig(interface{}, string) (interface{}, error)
}

type IOModule interface {
	Open(Configger, string) error
}

func RegisterIO(depend IOModule, name string) {
	_, ok := dependIO.Load(name)
	if ok {
		panic(fmt.Sprintf("io module[%s] already regist", name))
	}
	dependIO.Store(name, depend)
}

func RangeIOModule(fn func(string, IOModule) bool) {
	dependIO.Range(func(k, v interface{}) bool {
		name := k.(string)
		module := v.(IOModule)
		return fn(name, module)
	})
}

func ExistIOModule(name string) bool {
	_, ok := dependIO.Load(name)
	return ok
}
