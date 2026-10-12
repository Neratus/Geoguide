package dto

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type CreateTripRequest struct {
	UserID    domain.UserID
	Title     string
	StartDate time.Time
	EndDate   time.Time
	Budget    float64
	Notes     string
}

type CreateTripResponse struct {
	ID        domain.TripID
	Title     string
	StartDate string
	EndDate   string
	Budget    float64
	Status    string
}

type GetTripRequest struct {
	TripID domain.TripID
	UserID domain.UserID
}

type TripPlaceResponse struct {
	PlaceID     domain.PlaceID
	DayNumber   int
	ArrivalTime string
	DurationMin int
	Notes       string
	VisitStatus string
	ActualCost  float64
}

type GetTripResponse struct {
	ID        domain.TripID
	Title     string
	StartDate string
	EndDate   string
	Budget    float64
	Status    string
	Notes     string
	Places    []TripPlaceResponse
}

type GetUserTripsRequest struct {
	UserID domain.UserID
}

type UserTripResponse struct {
	ID        domain.TripID
	Title     string
	StartDate string
	EndDate   string
	Status    string
}

type DeleteTripRequest struct {
	TripID domain.TripID
}

type UpdateTripRequest struct {
	TripID    domain.TripID
	Title     *string
	StartDate *time.Time
	EndDate   *time.Time
	Budget    *float64
	Status    *string
	Notes     *string
	UserID    domain.UserID
}
