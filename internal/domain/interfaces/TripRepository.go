package interfaces

import (
	"context"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type TripRepository interface {
	Save(ctx context.Context, trip *domain.Trip) error
	FindByID(ctx context.Context, id domain.TripID) (*domain.Trip, error)
	FindByUser(ctx context.Context, userID domain.UserID) ([]*domain.Trip, error)
	Update(ctx context.Context, trip *domain.Trip) error
	Delete(ctx context.Context, id domain.TripID) error

	AddPlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID, dayNumber int, arrivalTime *time.Time, durationMin int, notes string) error
	RemovePlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID) error
	GetPlaces(ctx context.Context, tripID domain.TripID) ([]*domain.TripPlace, error)
	UpdateTripPlace(ctx context.Context, tripPlace *domain.TripPlace) error

	GetTripStatisticsReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.TripStatData, error)
}
