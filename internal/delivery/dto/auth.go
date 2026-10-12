package dto

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type EnableTwoFactorRequest struct {
	UserID domain.UserID
	Secret string
	Code   string
}

type VerifyTwoFactorRequest struct {
	ChallengeID string
	Code        string
}

type LoginUserRequest struct {
	Username string
	Password string
}

type LoginUserResponse struct {
	ID                domain.UserID `json:"id"`
	Username          string        `json:"username"`
	Email             string        `json:"email"`
	Token             string        `json:"token,omitempty"`
	RequiresTwoFactor bool          `json:"requires_two_factor,omitempty"`
	ChallengeID       string        `json:"challenge_id,omitempty"`
}

type LogoutRequest struct {
	Token string
}

type LogoutResponse struct {
	Success bool
}

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
