// ### context工具
package ctxutil

import (
	"context"
	"time"
)

// 封装超时context。
// 如果timeout为0，则返回原始context。
func Timeout(ctx context.Context, timeout int) (context.Context, context.CancelFunc) {
	ctx = Ensure(ctx)

	if timeout <= 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
}

// 确保ctx不为nil。
func Ensure(ctx context.Context) context.Context {
	if ctx == nil {
		return context.TODO()
	}

	return ctx
}
