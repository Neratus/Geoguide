package dto

import "github.com/Neratus/geoguide/internal/domain"

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

type GetTransportNodesByCityRequest struct {
	CityID domain.CityID
	Limit  int
	Offset int
}

type TransportNodeResponse struct {
	ID          domain.NodeID
	Name        string
	NodeType    string
	Coordinates domain.Coordinates
	Address     string
	CityID      domain.CityID
	ImageID     domain.ImageID
}
