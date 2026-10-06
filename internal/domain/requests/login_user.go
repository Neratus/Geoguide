package requests

import (
	"github.com/Neratus/geoguide/internal/domain"
)

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
