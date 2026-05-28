package authctx

import (
	"context"
	"time"

	"atom-maintenance/internal/domain"
)

type contextKey struct{}
type Principal struct {
	EmployeeID int64
	Role       domain.EmployeeRole
	JTI        string
	AccessTTL  time.Duration
}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(Principal)
	if !ok || p.EmployeeID <= 0 || p.JTI == "" {
		return Principal{}, false
	}
	return p, true
}
