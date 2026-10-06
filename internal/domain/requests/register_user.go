package requests

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type RegisterUserRequest struct {
	Username           string
	Email              string
	Password           string
	Phone              string
	BirthDate          *time.Time
	CountryOfResidence string
}

type RegisterUserResponse struct {
	ID       domain.UserID
	Username string
	Email    string
	Message  string
}

type VerifyContactRequest struct {
	UserID  domain.UserID
	Contact string
	Code    string
}
