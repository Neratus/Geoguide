package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/Neratus/geoguide/internal/domain"
)

type AddPlaceToTripRequest struct {
	TripID      domain.TripID
	PlaceID     domain.PlaceID
	DayNumber   int
	ArrivalTime *time.Time
	DurationMin int
	Notes       string
	UserID      domain.UserID
}

type GetPlaceRequest struct {
	ID uuid.UUID
}

type ReviewResponse struct {
	ID         uuid.UUID
	Rating     int
	Comment    string
	VisitDate  string
	CreatedAt  string
	Username   string
	IsApproved bool
}

type GetPlaceResponse struct {
	ID                  uuid.UUID
	Name                string
	Category            string
	Description         string
	Coordinates         string
	Address             string
	OpeningHours        string
	PriceInfo           string
	AvgVisitDurationMin int
	AvgRating           float64
	ReviewsCount        int
	ContactPhone        string
	Website             string
	CityID              uuid.UUID
	CityName            string
	DistrictID          uuid.UUID
	DistrictName        string
	ImageURL            string
	Reviews             []ReviewResponse
}

type GetPlacesByCityRequest struct {
	CityID   uuid.UUID
	Category string
	Limit    int
	Offset   int
	SortBy   string
}

type GetPlacesByCategoryRequest struct {
	Category string
	Limit    int
	Offset   int
}

type PlaceResponse struct {
	ID                  uuid.UUID
	Name                string
	Category            string
	Description         string
	Coordinates         string
	Address             string
	OpeningHours        string
	PriceInfo           string
	AvgVisitDurationMin int
	AvgRating           float64
	ReviewsCount        int
	ContactPhone        string
	Website             string
	CityID              uuid.UUID
	CityName            string
	DistrictID          uuid.UUID
	DistrictName        string
	ImageURL            string
	Reviews             []ReviewResponse
}

type RemovePlaceFromTripRequest struct {
	TripID  domain.TripID
	PlaceID domain.PlaceID
	UserID  domain.UserID
}
