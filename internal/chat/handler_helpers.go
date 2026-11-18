package chat

import (
	"errors"
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
)

func makeConversationResponse(
	conversation Conversation,
) conversationResponse {
	return conversationResponse{
		ID:              conversation.ID,
		ListingID:       conversation.ListingID,
		ListingTitle:    conversation.ListingTitle,
		InitiatorID:     conversation.InitiatorID,
		CreatedAt:       conversation.CreatedAt,
		UpdatedAt:       conversation.UpdatedAt,
		LastMessageText: conversation.LastMessageText,
	}
}

func writeChatError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidListingID):
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_listing_id",
		)

	case errors.Is(err, ErrConversationUnavailable):
		apihelper.WriteError(
			w,
			http.StatusNotFound,
			"conversation_unavailable",
		)

	case errors.Is(err, ErrCannotStartConversationWithSelf):
		apihelper.WriteError(
			w,
			http.StatusConflict,
			"cannot_start_conversation_with_self",
		)
	case errors.Is(err, ErrInvalidConversationID):
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_conversation_id")

	case errors.Is(err, ErrConversationNotFound):
		apihelper.WriteError(w, http.StatusNotFound, "conversation_not_found")

	case errors.Is(err, ErrInvalidMessageText):
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_message_text")

	default:
		apihelper.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
		)
	}
}

func makeConversationResponses(conversations []Conversation) []conversationResponse {
	responses := make([]conversationResponse, 0, len(conversations))

	for _, conversation := range conversations {
		responses = append(responses, makeConversationResponse(conversation))
	}

	return responses
}

func makeMessageResponse(message Message) messageResponse {
	return messageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Text:           message.Text,
		CreatedAt:      message.CreatedAt,
	}
}

func makeMessageResponses(messages []Message) []messageResponse {
	responses := make([]messageResponse, 0, len(messages))

	for _, message := range messages {
		responses = append(responses, makeMessageResponse(message))
	}

	return responses
}
