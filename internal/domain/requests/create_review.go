package requests

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
