package requests

import "github.com/Neratus/geoguide/internal/domain"

type RemovePlaceFromTripRequest struct {
	TripID  domain.TripID
	PlaceID domain.PlaceID
	UserID  domain.UserID
}
