package auth

import (
	"net/http"
	"strings"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/requestctx"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
)

type Middleware struct {
	service *Service
}

func NewMiddleware(service *Service) *Middleware {
	return &Middleware{
		service: service,
	}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			apihelper.WriteError(w, http.StatusUnauthorized, "access_token_required")
			return
		}

		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			apihelper.WriteError(w, http.StatusUnauthorized, "invalid_access_token")
			return
		}

		identity, err := m.service.ParseAccessToken(parts[1])
		if err != nil {
			apihelper.WriteError(w, http.StatusUnauthorized, "invalid_access_token")
			return
		}

		ctx := requestctx.WithCurrentUser(
			r.Context(),
			requestctx.CurrentUser{
				UserID: identity.UserID,
				Role:   string(identity.Role),
			},
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
		if !ok {
			apihelper.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if currentUser.Role != string(user.RoleAdmin) {
			apihelper.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}
