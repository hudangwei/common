package macaron

import (
	"log"
	"time"
)

func Logger() Handler {
	return func(ctx *Context) {
		start := time.Now()
		ctx.Next()
		log.Printf("route path:%s cost %v\n", ctx.Req.URL.Path, time.Since(start))
	}
}
