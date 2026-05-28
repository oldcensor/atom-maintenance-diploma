package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"atom-maintenance/internal/domain"
	"atom-maintenance/internal/ports"
	"atom-maintenance/pkg/authctx"
	jwtpkg "atom-maintenance/pkg/jwt"
	"atom-maintenance/pkg/respond"
)

const (
	blacklistPrefix  = "blacklist:"
	revokedEmpPrefix = "revoked:emp:"
)

func Auth(jwt *jwtpkg.Provider, cache ports.Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				respond.Error(w, domain.ErrUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			employeeID, role, jti, ttl, err := jwt.VerifyAccess(tokenStr)
			if err != nil {
				respond.Error(w, domain.ErrUnauthorized)
				return
			}

			val, err := cache.Get(r.Context(), blacklistPrefix+jti)
			if err == nil && val == "revoked" {
				respond.Error(w, domain.ErrUnauthorized)
				return
			}

			val, err = cache.Get(r.Context(), fmt.Sprintf("%s%d", revokedEmpPrefix, employeeID))
			if err == nil && val == "revoked" {
				respond.Error(w, domain.ErrUnauthorized)
				return
			}

			ctx := authctx.WithPrincipal(r.Context(), authctx.Principal{
				EmployeeID: employeeID,
				Role:       role,
				JTI:        jti,
				AccessTTL:  ttl,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
