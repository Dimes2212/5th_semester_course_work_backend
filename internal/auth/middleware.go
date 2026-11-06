package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
)

type Middleware struct {
	service *Service
}
type contextKey uint8

const identityContextKey contextKey = iota

func NewMiddleware(service *Service) *Middleware {
	return &Middleware{
		service: service,
	}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			writeError(w, http.StatusUnauthorized, "access_token_required")
			return
		}

		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "invalid_access_token")
			return
		}

		identity, err := m.service.ParseAccessToken(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_access_token")
			return
		}

		ctx := context.WithValue(r.Context(), identityContextKey, identity)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := IdentityFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if identity.Role != user.RoleAdmin {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityContextKey).(Identity)
	return identity, ok
}
