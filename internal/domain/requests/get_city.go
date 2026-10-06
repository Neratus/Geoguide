package requests

import "github.com/Neratus/geoguide/internal/domain"

type GetCityRequest struct {
	ID domain.CityID
}

type GetCityResponse struct {
	ID          domain.CityID
	Name        string
	Population  int64
	IsCapital   bool
	Coordinates domain.Coordinates
	Description string
	Timezone    string
	TravelTips  string
	CountryID   domain.CountryID
	ImageID     domain.ImageID
	ImageURL    string
}
