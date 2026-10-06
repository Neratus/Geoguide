package requests

import "github.com/google/uuid"

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
