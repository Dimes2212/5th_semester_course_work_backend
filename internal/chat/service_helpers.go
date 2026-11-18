package chat

import (
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
)

func isValidUUID(value string) bool {
	var id pgtype.UUID

	if err := id.Scan(value); err != nil {
		return false
	}

	return id.Valid
}

func normalizeMessageText(text string) (string, error) {
	text = strings.TrimSpace(text)

	if text == "" || utf8.RuneCountInString(text) > 2000 {
		return "", ErrInvalidMessageText
	}

	return text, nil
}
