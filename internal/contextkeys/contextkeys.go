package contextkeys

import (
	"context"
)

type ctxKey string

const (
	usernameKey ctxKey = "username"
)

func SetUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

func UsernameFromContext(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey).(string)
	return username
}
