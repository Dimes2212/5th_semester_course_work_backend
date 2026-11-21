package health

import (
	"context"
	"net/http"
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/jackc/pgx/v5/pgxpool"
)

const databasePingTimeout = 2 * time.Second

type Handler struct {
	db *pgxpool.Pool
}

type response struct {
	Status string `json:"status"`
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), databasePingTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		apihelper.WriteJSON(
			w,
			http.StatusServiceUnavailable,
			response{Status: "unavailable"},
		)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		response{Status: "ok"},
	)
}
