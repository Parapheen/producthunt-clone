package user

import "context"

type ContextKey string

const ContextKeyUser ContextKey = "user"

func GetUserFromContext(ctx context.Context) *User {
	if u, ok := ctx.Value(ContextKeyUser).(*User); ok {
		return u
	}
	return nil
}

func SetUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, ContextKeyUser, user)
}
