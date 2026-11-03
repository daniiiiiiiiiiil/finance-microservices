package ports

import "context"

type UserContext struct {
	UserID  int
	Email   string
	IsAdmin bool
}

type contextKey string

const userContextKey contextKey = "user_context"

func WithUser(ctx context.Context, u UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func UserFromContext(ctx context.Context) (UserContext, bool) {
	u, ok := ctx.Value(userContextKey).(UserContext)
	return u, ok
}
