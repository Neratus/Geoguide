package interfaces

import "github.com/Neratus/geoguide/internal/domain"

type MapClient interface {
	Geocode(address string) (lat, lng float64, err error)
	GetDistance(p1, p2 domain.Coordinates) (meters int, durationSec int, err error)
	OptimizeOrder(points []domain.Coordinates) ([]int, error)
}
