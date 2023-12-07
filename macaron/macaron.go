package macaron

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/hudangwei/common/macaron/user"
)

type Handler interface{}

type Macaron struct {
	Injector
	handlers []Handler
	action   Handler
}

func New() *Macaron {
	m := &Macaron{
		Injector: NewInjector(),
		action:   func() {},
	}
	m.Use(Logger())
	m.Use(Recovery())
	return m
}

func (m *Macaron) Use(handler Handler) {
	handler = validateAndWrapHandler(handler)
	m.handlers = append(m.handlers, handler)
}

func (m *Macaron) Action(handler Handler) {
	handler = validateAndWrapHandler(handler)
	m.action = handler
}

func (m *Macaron) newGinRoute(ctx *gin.Context, handler Handler, inputType reflect.Type) {
	var handlers []Handler
	handlers = append(handlers, m.handlers...)
	handlers = append(handlers, handler)
	c := new(Context)
	c.Injector = NewInjector()
	c.handlers = handlers
	c.action = m.action
	c.index = 0
	c.InputType = inputType
	c.Req = ctx.Request
	c.RespWriter = ctx.Writer
	c.PathParamsFunc = ctx.Param

	if u := user.UserFromContext(ctx.Request.Context()); u != nil {
		c.User.SessionId = u.SessionId
		c.User.Uid = u.Uid
		c.User.UserId = u.UserId
		c.User.UserName = u.UserName
		c.User.Avatar = u.Avatar
	}
	c.SetParent(m)
	c.Map(c)
	c.run()
}

func (m *Macaron) Wraps(handler Handler) func(*gin.Context) {
	inputType := getHandlerInput(handler)
	handler = validateAndWrapHandler(handler)
	return func(ctx *gin.Context) {
		m.newGinRoute(ctx, handler, inputType)
	}
}

func getHandlerInput(handler Handler) reflect.Type {
	t := reflect.TypeOf(handler)
	if t.Kind() != reflect.Func {
		panic("Macaron handler must be a callable function")
	}
	if t.NumIn() != 1 && t.NumIn() != 2 {
		panic("invalid num of args")
	}
	ctxArg := t.In(0)
	ctxType := reflect.TypeOf(new(Context))
	if ctxArg != ctxType {
		panic("The first input arg must be *Context")
	}

	if t.NumIn() == 2 {
		input := t.In(1)
		if input.Kind() == reflect.Ptr {
			input = input.Elem()
		}
		if input.Kind() != reflect.Struct {
			panic("input type must be struct or struct ptr")
		}
		return input
	}

	return nil
}

type ContextInvoker func(ctx *Context)

func (invoke ContextInvoker) Invoke(params []interface{}) ([]reflect.Value, error) {
	invoke(params[0].(*Context))
	return nil, nil
}

func validateAndWrapHandler(h Handler) Handler {
	if reflect.TypeOf(h).Kind() != reflect.Func {
		panic("Macaron handler must be a callable function")
	}

	if !IsFastInvoker(h) {
		switch v := h.(type) {
		case func(*Context):
			return ContextInvoker(v)
		}
	}
	return h
}
