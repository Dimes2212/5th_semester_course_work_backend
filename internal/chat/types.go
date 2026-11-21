package chat

import "time"

type Conversation struct {
	ID                  string
	ListingID           string
	ListingTitle        string
	ListingAuthorID     string
	InitiatorID         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastMessage         *Message
	LastMessageText     *string
	CounterpartID       string
	CounterpartFullName string
}

type Message struct {
	ID             string
	ConversationID string
	SenderID       string
	Text           string
	CreatedAt      time.Time
}

type messageResponse struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

type conversationResponse struct {
	ID              string    `json:"id"`
	ListingID       string    `json:"listing_id"`
	ListingTitle    string    `json:"listing_title"`
	InitiatorID     string    `json:"initiator_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LastMessageText *string   `json:"last_message_text"`
}

type createMessageRequest struct {
	Text string `json:"text"`
}

type MessageCreatedEvent struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}
