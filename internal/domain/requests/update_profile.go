package requests

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type UpdateUserProfileRequest struct {
	Token              string
	Username           *string
	Phone              *string
	BirthDate          *time.Time
	CountryOfResidence *string
	AvatarURL          *string
	FavoriteCategories []string
}

type UpdateUserProfileResponse struct {
	ID                 domain.UserID
	Username           string
	Email              string
	Phone              string
	RegisteredAt       string
	BirthDate          string
	CountryOfResidence string
	AvatarURL          string
	FavoriteCategories []string
	IsBlocked          bool
}
