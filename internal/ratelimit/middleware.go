package ratelimit

import (
	"net/http"
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func ByClientIP(requestLimit int, windowLength time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(
		requestLimit,
		windowLength,
		func(r *http.Request) (string, error) {
			return httprate.CanonicalizeIP(
				middleware.GetClientIP(r.Context()),
			), nil
		},
		httprate.WithLimitHandler(func(w http.ResponseWriter, _ *http.Request) {
			apihelper.WriteError(
				w,
				http.StatusTooManyRequests,
				"rate_limit_exceeded",
			)
		}),
	)
}
