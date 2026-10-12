package dto

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
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

type GetCountryRequest struct {
	ID uuid.UUID
}

type HolidayResponse struct {
	ID          uuid.UUID
	Name        string
	Date        time.Time
	Description string
	IsNational  bool
}

type GetCountryResponse struct {
	ID               uuid.UUID
	Name             string
	Area             float64
	Population       int64
	GDP              float64
	Currency         string
	VisaRequirements string
	Description      string
	SafetyTips       string
	BestSeason       string
	Language         string
	PhoneCode        string
	Religion         string
	CapitalID        uuid.UUID
	CapitalName      string
	ImageURL         string
	Holidays         []HolidayResponse
}

type GetHolidaysByCountryRequest struct {
	CountryID uuid.UUID
}

type HolidaySearchResponse struct {
	ID          uuid.UUID
	Name        string
	Date        time.Time
	Description string
	Traditions  string
	History     string
	IsNational  bool
}

type GetHolidaysByDateRequest struct {
	Date time.Time
}

type HolidaySearchDateResponse struct {
	ID          domain.HolidayID
	Name        string
	Date        time.Time
	Description string
	Traditions  string
	History     string
	IsNational  bool
	CountryID   domain.CountryID
	ImageID     domain.ImageID
}
