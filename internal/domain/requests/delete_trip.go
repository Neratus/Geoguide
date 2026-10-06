package requests

import "github.com/Neratus/geoguide/internal/domain"

type DeleteTripRequest struct {
	TripID domain.TripID
}
