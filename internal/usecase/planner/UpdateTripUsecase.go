package planner

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type UpdateTripUseCase struct {
	tripRepo interfaces.TripRepository
	logger   *slog.Logger
}

func NewUpdateTripUseCase(
	tripRepo interfaces.TripRepository,
	logger *slog.Logger,
) *UpdateTripUseCase {
	return &UpdateTripUseCase{
		tripRepo: tripRepo,
		logger:   logger,
	}
}

func (uc *UpdateTripUseCase) Execute(ctx context.Context, req requests.UpdateTripRequest) error {
	uc.logger.Info("updating trip", "trip_id", req.TripID)

	trip, err := uc.tripRepo.FindByID(ctx, req.TripID)
	if err != nil || trip == nil || trip.GetUserID() != req.UserID {
		uc.logger.Warn("trip not found", "trip_id", req.TripID)
		return usecase_errors.ErrTripNotFound
	}

	if req.Title != nil {
		trip.SetTitle(*req.Title)
	}
	if req.StartDate != nil {
		trip.SetStartDate(*req.StartDate)
	}
	if req.EndDate != nil {
		trip.SetEndDate(*req.EndDate)
	}
	if req.Budget != nil {
		trip.SetBudget(*req.Budget)
	}
	if req.Status != nil {
		trip.SetStatus(*req.Status)
	}
	if req.Notes != nil {
		trip.SetNotes(*req.Notes)
	}

	if err := uc.tripRepo.Update(ctx, trip); err != nil {
		uc.logger.Error("failed to update trip", "error", err)
		return err
	}

	uc.logger.Info("trip updated successfully")
	return nil
}
