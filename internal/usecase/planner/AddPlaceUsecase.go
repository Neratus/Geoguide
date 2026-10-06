package planner

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type AddPlaceToTripUseCase struct {
	tripRepo  interfaces.TripRepository
	placeRepo interfaces.PlaceRepository
	logger    *slog.Logger
}

func NewAddPlaceToTripUseCase(
	tripRepo interfaces.TripRepository,
	placeRepo interfaces.PlaceRepository,
	logger *slog.Logger,
) *AddPlaceToTripUseCase {
	return &AddPlaceToTripUseCase{
		tripRepo:  tripRepo,
		placeRepo: placeRepo,
		logger:    logger,
	}
}

func (uc *AddPlaceToTripUseCase) Execute(ctx context.Context, req requests.AddPlaceToTripRequest) error {
	uc.logger.Info("adding place to trip", "trip_id", req.TripID, "place_id", req.PlaceID)

	place, err := uc.placeRepo.FindByID(ctx, req.PlaceID)
	if err != nil || place == nil {
		uc.logger.Warn("place not found", "place_id", req.PlaceID)
		return usecase_errors.ErrPlaceNotFound
	}

	trip, err := uc.tripRepo.FindByID(ctx, req.TripID)
	if err != nil || trip == nil {
		uc.logger.Warn("trip not found", "trip_id", req.TripID)
		return usecase_errors.ErrTripNotFound
	}
	if trip.GetUserID() != req.UserID {
		uc.logger.Warn("trip not found", "trip_id", req.TripID)
		return usecase_errors.ErrTripNotFound
	}

	if err := uc.tripRepo.AddPlace(ctx, req.TripID, req.PlaceID, req.DayNumber, req.ArrivalTime, req.DurationMin, req.Notes); err != nil {
		uc.logger.Error("failed to add place to trip", "error", err)
		return err
	}

	uc.logger.Info("place added to trip successfully")
	return nil
}
