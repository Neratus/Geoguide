package planner

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
)

type CreateTripUseCase struct {
	tripRepo interfaces.TripRepository
	logger   *slog.Logger
}

func NewCreateTripUseCase(
	tripRepo interfaces.TripRepository,
	logger *slog.Logger,
) *CreateTripUseCase {
	return &CreateTripUseCase{
		tripRepo: tripRepo,
		logger:   logger,
	}
}

func (uc *CreateTripUseCase) Execute(ctx context.Context, req requests.CreateTripRequest) (*requests.CreateTripResponse, error) {
	uc.logger.Info("creating trip", "user_id", req.UserID, "title", req.Title)

	trip, err := domain.NewTrip(
		domain.TripID{},
		req.Title,
		req.StartDate,
		req.EndDate,
		req.Budget,
		"DRAFT",
		req.Notes,
		req.UserID,
		domain.ImageID{},
	)
	if err != nil {
		uc.logger.Error("invalid trip data", "error", err)
		return nil, err
	}

	if err := uc.tripRepo.Save(ctx, trip); err != nil {
		uc.logger.Error("failed to save trip", "error", err)
		return nil, err
	}

	uc.logger.Info("trip created successfully", "trip_id", trip.GetId())
	return &requests.CreateTripResponse{
		ID:        trip.GetId(),
		Title:     trip.GetTitle(),
		StartDate: trip.GetStartDate().Format("2006-01-02"),
		EndDate:   trip.GetEndDate().Format("2006-01-02"),
		Budget:    trip.GetBudget(),
		Status:    trip.GetStatus(),
	}, nil
}
