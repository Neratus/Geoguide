package requests

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type CreateTripRequest struct {
	UserID    domain.UserID
	Title     string
	StartDate time.Time
	EndDate   time.Time
	Budget    float64
	Notes     string
}

type CreateTripResponse struct {
	ID        domain.TripID
	Title     string
	StartDate string
	EndDate   string
	Budget    float64
	Status    string
}
