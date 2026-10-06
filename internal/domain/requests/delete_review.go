package requests

import "github.com/Neratus/geoguide/internal/domain"

type DeleteReviewRequest struct {
	ReviewID    domain.ReviewID
	UserID      domain.UserID
	IsModerator bool
}

type DeleteReviewResponse struct {
	Success bool
}
