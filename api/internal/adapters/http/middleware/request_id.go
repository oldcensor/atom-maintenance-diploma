package middleware

import (
	"net/http"

	"atom-maintenance/pkg"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(pkg.RequestIDHeader)
		if rid == "" {
			rid = pkg.NewRequestID()
		}
		ctx := pkg.WithRequestID(r.Context(), rid)
		w.Header().Set(pkg.RequestIDHeader, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
