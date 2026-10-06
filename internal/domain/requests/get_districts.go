package requests

import "github.com/Neratus/geoguide/internal/domain"

type GetDistrictsByCityRequest struct {
	CityID domain.CityID
	Limit  int
	Offset int
}

type DistrictResponse struct {
	ID          domain.CityDistrictID
	Name        string
	Description string
	Coordinates domain.Coordinates
	CityID      domain.CityID
	ImageID     domain.ImageID
}
