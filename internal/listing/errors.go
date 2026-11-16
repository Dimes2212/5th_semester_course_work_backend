package listing

import "errors"

var (
	ErrCategoryRequired         = errors.New("category is required")
	ErrLocationRequired         = errors.New("location is required")
	ErrInvalidKind              = errors.New("invalid listing kind")
	ErrTitleRequired            = errors.New("title is required")
	ErrTitleTooLong             = errors.New("title is too long")
	ErrDescriptionTooLong       = errors.New("description is too long")
	ErrLocationDetailsTooLong   = errors.New("location details are too long")
	ErrEventAtRequired          = errors.New("event time is required")
	ErrListingNotFound          = errors.New("listing not found")
	ErrInvalidListingID         = errors.New("invalid listing id")
	ErrInvalidCursor            = errors.New("invalid cursor")
	ErrInvalidCategoryID        = errors.New("invalid category id")
	ErrInvalidLocationID        = errors.New("invalid location id")
	ErrSearchQueryTooLong       = errors.New("search query is too long")
	ErrListingPhotoLimitReached = errors.New("listing photo limit reached")
	ErrPhotoTooLarge            = errors.New("photo is too large")
	ErrInvalidPhotoType         = errors.New("invalid photo type")
	ErrEmptyPhoto               = errors.New("photo is empty")
	ErrInvalidPhotoRequest      = errors.New("invalid photo request")
	ErrInvalidListingPhotoID    = errors.New("invalid listing photo id")
	ErrListingPhotoNotFound     = errors.New("listing photo not found")
)
