package macaron

import (
	"net/http"
	"reflect"

	"github.com/hudangwei/common/macaron/user"
)

type Context struct {
	Injector
	handlers []Handler
	action   Handler
	index    int
	IsAbort  bool //中止handler链，后续的handler不再执行

	//自定义数据
	InputType      reflect.Type
	Req            *http.Request
	RespWriter     http.ResponseWriter
	PathParamsFunc func(key string) string
	User           user.User
}

func (ctx *Context) handler() Handler {
	if ctx.index < len(ctx.handlers) {
		return ctx.handlers[ctx.index]
	}
	if ctx.index == len(ctx.handlers) {
		return ctx.action
	}
	panic("invalid index for context handler")
}

func (ctx *Context) Next() {
	ctx.index++
	ctx.run()
}

func (ctx *Context) run() {
	for ctx.index <= len(ctx.handlers) {
		handler := ctx.handler()
		// pretty.Println(ctx.Req.URL.Path, ctx.index, handler)
		vals, err := ctx.Invoke(handler)
		if err != nil {
			panic(err)
		}
		ctx.index++

		if len(vals) > 0 {
			if vals[0].Kind() == reflect.Int {
				if vals[0].Int() == 1 {
					ctx.IsAbort = true
					return
				}
			} else {
				ev := ctx.GetVal(reflect.TypeOf(ReturnHandler(nil)))
				if !ev.IsValid() {
					return
				}
				handleReturn := ev.Interface().(ReturnHandler)
				handleReturn(ctx, vals)
			}
		}
		if ctx.IsAbort {
			return
		}
	}
}
