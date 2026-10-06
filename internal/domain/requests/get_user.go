package requests

import "github.com/Neratus/geoguide/internal/domain"

type GetUserProfileRequest struct {
	UserID domain.UserID
}

type GetUserProfileResponse struct {
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
	Role               string
}
