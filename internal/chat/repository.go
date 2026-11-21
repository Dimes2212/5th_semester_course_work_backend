package chat

import (
	"context"
	"errors"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/listing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateOrGetConversation(
	ctx context.Context,
	listingID string,
	initiatorID string,
) (Conversation, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Conversation{}, err
	}
	defer tx.Rollback(ctx)

	var currentListing listing.Listing

	err = tx.QueryRow(ctx, `
		SELECT
			id,
			author_id,
			category_id,
			location_id,
			location_details,
			kind,
			title,
			description,
			event_at,
			status,
			resolved_with_user_id,
			resolved_at,
			created_at,
			updated_at,
			deleted_at
		FROM listings
		WHERE id = $1
		  AND deleted_at IS NULL
		FOR SHARE
	`, listingID).Scan(
		&currentListing.ID,
		&currentListing.AuthorID,
		&currentListing.CategoryID,
		&currentListing.LocationID,
		&currentListing.LocationDetails,
		&currentListing.Kind,
		&currentListing.Title,
		&currentListing.Description,
		&currentListing.EventAt,
		&currentListing.Status,
		&currentListing.ResolvedWithUserID,
		&currentListing.ResolvedAt,
		&currentListing.CreatedAt,
		&currentListing.UpdatedAt,
		&currentListing.DeletedAt,
	)
	if err != nil {
		return Conversation{}, err
	}

	if currentListing.Status != listing.ListingStatusActive {
		return Conversation{}, ErrConversationUnavailable
	}

	if initiatorID == currentListing.AuthorID {
		return Conversation{}, ErrCannotStartConversationWithSelf
	}

	var existingConversation Conversation

	err = tx.QueryRow(ctx, `
		SELECT
			id,
			listing_id,
			initiator_id,
			created_at,
			updated_at
		FROM conversations
		WHERE initiator_id = $1
		  AND listing_id = $2
	`, initiatorID, listingID).Scan(
		&existingConversation.ID,
		&existingConversation.ListingID,
		&existingConversation.InitiatorID,
		&existingConversation.CreatedAt,
		&existingConversation.UpdatedAt,
	)
	if err == nil {
		existingConversation.ListingTitle = currentListing.Title
		existingConversation.ListingAuthorID = currentListing.AuthorID

		if err := tx.Commit(ctx); err != nil {
			return Conversation{}, err
		}

		return existingConversation, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return Conversation{}, err
	}

	var createdConversation Conversation

	err = tx.QueryRow(ctx, `
		INSERT INTO conversations (
			listing_id,
			initiator_id
		)
		VALUES ($1, $2)
		ON CONFLICT (listing_id, initiator_id)
		DO UPDATE SET
			updated_at = conversations.updated_at
		RETURNING
			id,
			listing_id,
			initiator_id,
			created_at,
			updated_at
	`, listingID, initiatorID).Scan(
		&createdConversation.ID,
		&createdConversation.ListingID,
		&createdConversation.InitiatorID,
		&createdConversation.CreatedAt,
		&createdConversation.UpdatedAt,
	)
	if err != nil {
		return Conversation{}, err
	}

	createdConversation.ListingTitle = currentListing.Title
	createdConversation.ListingAuthorID = currentListing.AuthorID

	if err := tx.Commit(ctx); err != nil {
		return Conversation{}, err
	}

	return createdConversation, nil
}

func (r *Repository) FindConversationsByUserID(ctx context.Context, userID string) ([]Conversation, error) {
	rows, err := r.db.Query(ctx, `
		WITH conversations_started_by_user AS (
			SELECT
				id,
				listing_id,
				initiator_id,
				created_at,
				updated_at
			FROM conversations
			WHERE initiator_id = $1
		), listings_created_by_user AS (
			SELECT id
			FROM listings
			WHERE author_id = $1
		), conversations_received_by_user AS (
			SELECT
				id,
				listing_id,
				initiator_id,
				created_at,
				updated_at
			FROM conversations
			WHERE listing_id IN (
				SELECT id
				FROM listings_created_by_user
			)
		), all_user_conversations AS (
			SELECT *
			FROM conversations_started_by_user
			UNION
			SELECT *
			FROM conversations_received_by_user
		), messages_ranked_by_newest AS (
			SELECT
				message.conversation_id,
				message.text,
				ROW_NUMBER() OVER (
					PARTITION BY message.conversation_id
					ORDER BY message.created_at DESC, message.id DESC
				) AS row_number
			FROM messages AS message
			WHERE message.conversation_id IN (
				SELECT id
				FROM all_user_conversations
			)
		)
		SELECT
			conversation.id,
			conversation.listing_id,
			listing.title,
			conversation.initiator_id,
			conversation.created_at,
			conversation.updated_at,
			last_message.text AS last_message_text
		FROM all_user_conversations AS conversation
		JOIN listings AS listing
		  ON listing.id = conversation.listing_id
		LEFT JOIN messages_ranked_by_newest AS last_message
		  ON last_message.conversation_id = conversation.id
		 AND last_message.row_number = 1
		ORDER BY conversation.updated_at DESC, conversation.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]Conversation, 0)

	for rows.Next() {
		var conversation Conversation

		if err := rows.Scan(
			&conversation.ID,
			&conversation.ListingID,
			&conversation.ListingTitle,
			&conversation.InitiatorID,
			&conversation.CreatedAt,
			&conversation.UpdatedAt,
			&conversation.LastMessageText,
		); err != nil {
			return nil, err
		}

		conversations = append(conversations, conversation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *Repository) FindMessagesByConversationID(ctx context.Context, conversationID string, userID string) ([]Message, error) {
	var hasAccess bool
	messages := make([]Message, 0)

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM conversations AS conversation
			JOIN listings AS listing
			ON listing.id = conversation.listing_id
			WHERE conversation.id = $1
			AND (
				conversation.initiator_id = $2
				OR listing.author_id = $2
			)
		);
	`, conversationID, userID).Scan(&hasAccess)

	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pgx.ErrNoRows
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, conversation_id, sender_id, text, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC, id ASC
	`, conversationID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var message Message

		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.SenderID,
			&message.Text,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil

}

func (r *Repository) CreateMessage(ctx context.Context, conversationID string, senderID string, text string) (Message, string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Message{}, "", err
	}
	defer tx.Rollback(ctx)

	var initiatorID string
	var listingAuthorID string

	err = tx.QueryRow(ctx, `
		SELECT
			conversation.initiator_id,
			listing.author_id
		FROM conversations AS conversation
		JOIN listings AS listing
		  ON listing.id = conversation.listing_id
		WHERE conversation.id = $1
		  AND (
			  conversation.initiator_id = $2
			  OR listing.author_id = $2
		  )
	`, conversationID, senderID).Scan(&initiatorID, &listingAuthorID)

	if err != nil {
		return Message{}, "", err
	}

	recipientID := listingAuthorID
	if senderID == listingAuthorID {
		recipientID = initiatorID
	}

	var newMessage Message

	err = tx.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, sender_id, text, created_at
	`, conversationID, senderID, text).Scan(
		&newMessage.ID,
		&newMessage.ConversationID,
		&newMessage.SenderID,
		&newMessage.Text,
		&newMessage.CreatedAt,
	)

	if err != nil {
		return Message{}, "", err
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE conversations
		SET updated_at = now()
		WHERE id = $1 
	`, conversationID)

	if err != nil {
		return Message{}, "", err
	}

	if commandTag.RowsAffected() == 0 {
		return Message{}, "", pgx.ErrNoRows
	}

	if err := tx.Commit(ctx); err != nil {
		return Message{}, "", err
	}

	return newMessage, recipientID, nil
}
