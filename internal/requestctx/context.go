package requestctx

import "context"

type CurrentUser struct {
	UserID string
	Role   string
}

type contextKey uint8

const currentUserKey contextKey = iota

func WithCurrentUser(ctx context.Context, currentUser CurrentUser) context.Context {
	return context.WithValue(
		ctx, currentUserKey, currentUser,
	)
}

func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	currentUser, ok := ctx.Value(currentUserKey).(CurrentUser)

	return currentUser, ok
}
