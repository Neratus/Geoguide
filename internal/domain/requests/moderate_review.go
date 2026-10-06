package requests

import "github.com/Neratus/geoguide/internal/domain"

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
