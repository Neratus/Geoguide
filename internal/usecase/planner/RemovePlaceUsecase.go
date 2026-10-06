package planner

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
)

type RemovePlaceFromTripUseCase struct {
	tripRepo interfaces.TripRepository
	logger   *slog.Logger
}

func NewRemovePlaceFromTripUseCase(
	tripRepo interfaces.TripRepository,
	logger *slog.Logger,
) *RemovePlaceFromTripUseCase {
	return &RemovePlaceFromTripUseCase{
		tripRepo: tripRepo,
		logger:   logger,
	}
}

func (uc *RemovePlaceFromTripUseCase) Execute(ctx context.Context, req requests.RemovePlaceFromTripRequest) error {
	uc.logger.Info("removing place from trip", "trip_id", req.TripID, "place_id", req.PlaceID)

	if err := uc.tripRepo.RemovePlace(ctx, req.TripID, req.PlaceID); err != nil {
		uc.logger.Error("failed to remove place from trip", "error", err)
		return err
	}

	uc.logger.Info("place removed from trip successfully")
	return nil
}
