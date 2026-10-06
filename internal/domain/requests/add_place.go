package requests

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type AddPlaceToTripRequest struct {
	TripID      domain.TripID
	PlaceID     domain.PlaceID
	DayNumber   int
	ArrivalTime *time.Time
	DurationMin int
	Notes       string
	UserID      domain.UserID
}
