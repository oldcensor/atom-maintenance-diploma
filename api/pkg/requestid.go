package pkg

import (
	"context"

	"github.com/google/uuid"
)

type reqIDKey struct{}

const RequestIDHeader = "X-Request-ID"

func NewRequestID() string { return uuid.NewString() }

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, reqIDKey{}, id)
}

func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(reqIDKey{}).(string); ok {
		return v
	}
	return ""
}
