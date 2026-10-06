package requests

import "github.com/google/uuid"

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
