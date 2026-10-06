package requests

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

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
