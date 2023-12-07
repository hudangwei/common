package macaron

import (
	"fmt"
	"log"
	"runtime"
)

func Recovery() Handler {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				logPanic(err)
			}
		}()
		ctx.Next()
	}
}

func logPanic(r interface{}) {
	callers := ""
	for i := 0; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	log.Printf("Recovered from panic: %#v (%v)\n%v\n", r, r, callers)
}
