package admin

import (
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/listing"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
)

type UserDetails struct {
	ID        string
	FullName  string
	Email     string
	Role      user.Role
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	BlockedAt *time.Time
}

type userResponse struct {
	ID        string     `json:"id"`
	FullName  string     `json:"full_name"`
	Email     string     `json:"email"`
	Role      user.Role  `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	BlockedAt *time.Time `json:"blocked_at"`
}

type listingResponse struct {
	ID                 string                 `json:"id"`
	AuthorID           string                 `json:"author_id"`
	CategoryID         int16                  `json:"category_id"`
	LocationID         int16                  `json:"location_id"`
	LocationDetails    *string                `json:"location_details"`
	Kind               listing.ListingKind    `json:"kind"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	EventAt            time.Time              `json:"event_at"`
	Status             listing.ListingStatus  `json:"status"`
	ResolvedWithUserID *string                `json:"resolved_with_user_id"`
	ResolvedAt         *time.Time             `json:"resolved_at"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	DeletedAt          *time.Time             `json:"deleted_at"`
	Photos             []listingPhotoResponse `json:"photos"`
}

type listingPhotoResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Position int16  `json:"position"`
}
