package macaron

import (
	"reflect"
)

// Abort 中止handler链，后续的handler不再执行
const Abort = 1

type ReturnHandler func(*Context, []reflect.Value)
