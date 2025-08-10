package userid

import (
	"context"
)

type userIdKey struct{}

// MustFromCtx extracts the player id from the context.
func MustFromCtx(ctx context.Context) string {
	pId, ok := ctx.Value(userIdKey{}).(string)
	if !ok {
		panic("player id not found in context")
	}

	return pId
}

// ToCtx adds the player id to the context.
func ToCtx(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, userIdKey{}, userId)
}
