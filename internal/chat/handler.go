package chat

import (
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/requestctx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateOrGetConversation(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	listingID := chi.URLParam(r, "listingID")

	conversation, err := h.service.CreateOrGetConversation(r.Context(), listingID, currentUser.UserID)
	if err != nil {
		writeChatError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeConversationResponse(conversation),
	)
}

func (h *Handler) GetMyConversations(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	conversations, err := h.service.GetConversationsByUserID(r.Context(), currentUser.UserID)
	if err != nil {
		writeChatError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeConversationResponses(conversations),
	)

}

func (h *Handler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversationID := chi.URLParam(r, "conversationID")

	messages, err := h.service.GetMessagesByConversationID(
		r.Context(),
		conversationID,
		currentUser.UserID,
	)
	if err != nil {
		writeChatError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusOK, makeMessageResponses(messages))
}

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var request createMessageRequest

	if err := apihelper.DecodeJSON(w, r, &request); err != nil {
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	conversationID := chi.URLParam(r, "conversationID")

	message, err := h.service.CreateMessage(r.Context(), conversationID, currentUser.UserID, request.Text)
	if err != nil {
		writeChatError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusCreated,
		makeMessageResponse(message),
	)
}
