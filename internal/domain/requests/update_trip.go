package requests

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type UpdateTripRequest struct {
	TripID    domain.TripID
	Title     *string
	StartDate *time.Time
	EndDate   *time.Time
	Budget    *float64
	Status    *string
	Notes     *string
	UserID    domain.UserID
}
