package middleware

import (
	"net/http"

	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/authctx"
	"atom-maintenance/pkg/respond"
)

func RequireRole(min domain.EmployeeRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := authctx.PrincipalFrom(r.Context())
			if !ok {
				respond.Error(w, domain.ErrUnauthorized)
				return
			}
			if !p.Role.AtLeast(min) {
				respond.Error(w, domain.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
