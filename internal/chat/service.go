package chat

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateOrGetConversation(ctx context.Context, listingID string, initiatorID string) (Conversation, error) {
	if !isValidUUID(listingID) {
		return Conversation{}, ErrInvalidListingID
	}

	currentConversation, err := s.repository.CreateOrGetConversation(ctx, listingID, initiatorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Conversation{}, ErrConversationUnavailable
	}

	if errors.Is(err, ErrConversationUnavailable) ||
		errors.Is(err, ErrCannotStartConversationWithSelf) {
		return Conversation{}, err
	}

	if err != nil {
		return Conversation{}, fmt.Errorf(
			"create or get conversation: %w",
			err,
		)
	}

	return currentConversation, nil
}

func (s *Service) GetConversationsByUserID(ctx context.Context, userID string) ([]Conversation, error) {
	conversations, err := s.repository.FindConversationsByUserID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"find conversations by user id: %w",
			err,
		)
	}

	return conversations, nil
}

func (s *Service) GetMessagesByConversationID(ctx context.Context, conversationID string, userID string) ([]Message, error) {
	ok := isValidUUID(conversationID)
	if !ok {
		return []Message{}, ErrInvalidConversationID
	}

	messages, err := s.repository.FindMessagesByConversationID(ctx, conversationID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find conversation messages: %w", err)
	}

	return messages, nil
}

func (s *Service) CreateMessage(ctx context.Context, conversationID string, senderID string, text string) (Message, error) {
	if !isValidUUID(conversationID) {
		return Message{}, ErrInvalidConversationID
	}

	text, err := normalizeMessageText(text)
	if err != nil {
		return Message{}, err
	}

	message, err := s.repository.CreateMessage(ctx, conversationID, senderID, text)
	if errors.Is(err, pgx.ErrNoRows) {
		return Message{}, ErrConversationNotFound
	}
	if err != nil {
		return Message{}, fmt.Errorf("create message: %w", err)
	}

	return message, nil
}
