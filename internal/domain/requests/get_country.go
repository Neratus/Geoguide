package requests

import (
	"time"

	"github.com/google/uuid"
)

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
