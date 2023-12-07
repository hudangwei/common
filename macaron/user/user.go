package user

import (
	"context"
)

type User struct {
	Uid       int64  `json:"uid"`
	UserId    string `json:"userid"`
	UserName  string `json:"username"`
	Avatar    string `json:"avatar"`
	SessionId string `json:"sessionid"`
}

var ctxUserKey = struct{}{}

func NewContextWithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, ctxUserKey, u)
}

func UserFromContext(ctx context.Context) *User {
	v := ctx.Value(ctxUserKey)
	if u, ok := v.(*User); ok {
		return u
	}
	return nil
}
