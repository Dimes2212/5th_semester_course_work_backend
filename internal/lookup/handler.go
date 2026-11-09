package lookup

import (
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetCategories(
	w http.ResponseWriter,
	r *http.Request,
) {
	categories, err := h.service.GetCategories(r.Context())
	if err != nil {
		apihelper.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
		)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		categories,
	)
}

func (h *Handler) GetLocations(
	w http.ResponseWriter,
	r *http.Request,
) {
	locations, err := h.service.GetLocations(r.Context())
	if err != nil {
		apihelper.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
		)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		locations,
	)
}
