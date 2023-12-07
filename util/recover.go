package util

import (
	"fmt"
	"runtime"

	"github.com/hudangwei/common/logger"
)

var PanicHandler func(interface{})

func LogPanic(r interface{}) {
	callers := ""
	for i := 0; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	logger.Error(fmt.Sprintf("Recovered from panic: %#v (%v)\n%v\n", r, r, callers))
}

func init() {
	PanicHandler = LogPanic
}

func WithRecover(fn func()) {
	defer func() {
		handler := PanicHandler
		if handler != nil {
			if err := recover(); err != nil {
				handler(err)
			}
		}
	}()

	fn()
}

func CaptureException() {
	handler := PanicHandler
	if handler != nil {
		if err := recover(); err != nil {
			handler(err)
		}
	}
}
