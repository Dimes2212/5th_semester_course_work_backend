package listing

import (
	"io"
	"time"
)

type ListingKind string

const (
	ListingKindLost  ListingKind = "lost"
	ListingKindFound ListingKind = "found"
)

type ListingStatus string

const (
	ListingStatusActive   ListingStatus = "active"
	ListingStatusResolved ListingStatus = "resolved"
	ListingStatusArchived ListingStatus = "archived"
)

type Listing struct {
	ID                 string
	AuthorID           string
	CategoryID         int16
	LocationID         int16
	LocationDetails    *string
	Kind               ListingKind
	Title              string
	Description        string
	EventAt            time.Time
	Status             ListingStatus
	ResolvedWithUserID *string
	ResolvedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
	Photos             []ListingPhoto
}

type ListingPhoto struct {
	ID         string
	ListingID  string
	StorageKey string
	Position   int16
	URL        string
}

type CreateListingInput struct {
	CategoryID      int16
	LocationID      int16
	LocationDetails *string
	Kind            ListingKind
	Title           string
	Description     string
	EventAt         time.Time
}

type createListingRequest struct {
	CategoryID      int16       `json:"category_id"`
	LocationID      int16       `json:"location_id"`
	LocationDetails *string     `json:"location_details"`
	Kind            ListingKind `json:"kind"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	EventAt         time.Time   `json:"event_at"`
}

type ListingFeedCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

type GetListingsFeedInput struct {
	Cursor     string
	Query      string
	Kind       *ListingKind
	CategoryID *int16
	LocationID *int16
}

type ListingFeedFilter struct {
	Query      *string
	Kind       *ListingKind
	CategoryID *int16
	LocationID *int16
}

type GetListingsFeedResult struct {
	Listings   []Listing
	NextCursor *string
	HasMore    bool
}

type listingPhotoResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Position int16  `json:"position"`
}

type listingResponse struct {
	ID                 string                 `json:"id"`
	AuthorID           string                 `json:"author_id"`
	CategoryID         int16                  `json:"category_id"`
	LocationID         int16                  `json:"location_id"`
	LocationDetails    *string                `json:"location_details"`
	Kind               ListingKind            `json:"kind"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	EventAt            time.Time              `json:"event_at"`
	Status             ListingStatus          `json:"status"`
	ResolvedWithUserID *string                `json:"resolved_with_user_id"`
	ResolvedAt         *time.Time             `json:"resolved_at"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	Photos             []listingPhotoResponse `json:"photos"`
}

type listingsFeedResponse struct {
	Items      []listingResponse `json:"items"`
	NextCursor *string           `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
}

type UpdateListingInput struct {
	CategoryID      int16
	LocationID      int16
	LocationDetails *string
	Kind            ListingKind
	Title           string
	Description     string
	EventAt         time.Time
}

type updateListingRequest struct {
	CategoryID      int16       `json:"category_id"`
	LocationID      int16       `json:"location_id"`
	LocationDetails *string     `json:"location_details"`
	Kind            ListingKind `json:"kind"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	EventAt         time.Time   `json:"event_at"`
}

type deleteListingResponse struct {
	ID        string     `json:"id"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type AddListingPhotoInput struct {
	File        io.Reader
	Size        int64
	ContentType string
}

type AddListingPhotoResult struct {
	Photo ListingPhoto
	URL   string
}

type addListingPhotoResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Position int16  `json:"position"`
}
