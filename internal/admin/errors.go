package admin

import "errors"

var (
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidBlockedValue = errors.New("invalid blocked value")
	ErrInvalidListingID    = errors.New("invalid listing id")
	ErrListingNotFound     = errors.New("listing not found")
)
