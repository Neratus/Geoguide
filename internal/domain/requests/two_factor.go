package requests

import "github.com/Neratus/geoguide/internal/domain"

type EnableTwoFactorRequest struct {
	UserID domain.UserID
	Secret string
	Code   string
}

type VerifyTwoFactorRequest struct {
	ChallengeID string
	Code        string
}
