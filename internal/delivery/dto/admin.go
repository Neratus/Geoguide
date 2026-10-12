package dto

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type GetUsersRequest struct {
	Limit       int
	Offset      int
	Search      string
	IsModerator bool
}

type GetUsersResponse struct {
	Users []*domain.User
}

type GetPendingReviewsRequest struct {
	Limit       int
	Offset      int
	IsModerator bool
	IsApproved  bool
}

type GetPendingReviewsResponse struct {
	Reviews []*domain.Review
}

type ModerateReviewRequest struct {
	ReviewID    domain.ReviewID
	Approved    bool
	Comment     string
	ModeratorID domain.UserID
}

type ModerateReviewResponse struct {
	ID                domain.ReviewID
	IsApproved        bool
	IsModerated       bool
	ModerationComment string
}

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
