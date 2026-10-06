package requests

import "github.com/Neratus/geoguide/internal/domain"

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
