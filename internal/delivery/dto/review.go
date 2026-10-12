package dto

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type CreateReviewRequest struct {
	UserID    domain.UserID
	PlaceID   domain.PlaceID
	Rating    int
	Comment   string
	VisitDate time.Time
	ImageURL  string
}

type CreateReviewResponse struct {
	ID          domain.ReviewID
	Rating      int
	Comment     string
	VisitDate   string
	CreatedAt   string
	IsModerated bool
	IsApproved  bool
}

type DeleteReviewRequest struct {
	ReviewID    domain.ReviewID
	UserID      domain.UserID
	IsModerator bool
}

type DeleteReviewResponse struct {
	Success bool
}

type GetUserReviewsRequest struct {
	UserID domain.UserID
	Limit  int
	Offset int
}

type UserReviewResponse struct {
	ID          domain.ReviewID
	Rating      int
	Comment     string
	VisitDate   string
	CreatedAt   string
	PlaceID     domain.PlaceID
	PlaceName   string
	IsApproved  bool
	IsModerated bool
}

type ReviewByPlaceResponse struct {
	ID                domain.ReviewID
	Rating            int
	Comment           string
	VisitDate         string
	CreatedAt         string
	Username          string
	UserAvatar        string
	IsApproved        bool
	ModerationComment string
}

type GetReviewsByPlaceRequest struct {
	PlaceID      domain.PlaceID
	OnlyApproved bool
	Limit        int
	Offset       int
}
