package chat

import "errors"

var (
	ErrConversationUnavailable         = errors.New("conversation unavailable")
	ErrCannotStartConversationWithSelf = errors.New("cannot start conversation with self")
	ErrInvalidListingID                = errors.New("invalid listing id")
	ErrInvalidConversationID           = errors.New("invalid conversation id")
	ErrConversationNotFound            = errors.New("conversation not found")
	ErrInvalidMessageText              = errors.New("invalid message text")
)
