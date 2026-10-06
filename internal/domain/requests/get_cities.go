package requests

import (
	"github.com/Neratus/geoguide/internal/domain"
)

type GetCitiesByCountryRequest struct {
	CountryID domain.CountryID
	Limit     int
	Offset    int
}

type GetAllCitiesRequest struct {
	Limit  int
	Offset int
}

type SearchCitiesRequest struct {
	Query  string
	Limit  int
	Offset int
}

type CityResponse struct {
	ID          domain.CityID
	Name        string
	Population  int64
	IsCapital   bool
	Coordinates string
	Description string
	Timezone    string
	ImageURL    string
}
