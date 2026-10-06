package requests

import "github.com/Neratus/geoguide/internal/domain"

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
