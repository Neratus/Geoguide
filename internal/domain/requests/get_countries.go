package requests

import (
	"github.com/Neratus/geoguide/internal/domain"
)

type GetCountriesRequest struct {
	Limit  int
	Offset int
	Search string
}

type CountryResponse struct {
	ID         domain.CountryID
	Name       string
	Capital    string
	Area       float64
	Population int64
	Currency   string
	Language   string
	PhoneCode  string
	ImageURL   string
}

type SearchCountriesRequest struct {
	Query  string
	Limit  int
	Offset int
}

type CountrySearchResponse struct {
	ID       domain.CountryID
	Name     string
	Capital  string
	Currency string
}
