package requests

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type ModerateUserRequest struct {
	UserID      domain.UserID
	Block       bool
	Reason      string
	ModeratorID domain.UserID
}

type ModerateUserResponse struct {
	UserID      domain.UserID
	IsBlocked   bool
	BlockedAt   *time.Time
	BlockReason string
}
