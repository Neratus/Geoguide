package planner

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type DeleteTripUseCase struct {
	tripRepo interfaces.TripRepository
	logger   *slog.Logger
}

func NewDeleteTripUseCase(
	tripRepo interfaces.TripRepository,
	logger *slog.Logger,
) *DeleteTripUseCase {
	return &DeleteTripUseCase{
		tripRepo: tripRepo,
		logger:   logger,
	}
}

func (uc *DeleteTripUseCase) Execute(ctx context.Context, req dto.DeleteTripRequest) error {
	uc.logger.Info("deleting trip", "trip_id", req.TripID)

	if err := uc.tripRepo.Delete(ctx, req.TripID); err != nil {
		uc.logger.Error("failed to delete trip", "error", err)
		return err
	}

	uc.logger.Info("trip deleted successfully")
	return nil
}
