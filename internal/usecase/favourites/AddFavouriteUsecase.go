package favourites

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
)

type AddFavouriteUseCase struct {
	favouriteRepo interfaces.FavouriteRepository
	placeRepo     interfaces.PlaceRepository
	logger        *slog.Logger
}

func NewAddFavouriteUseCase(
	favouriteRepo interfaces.FavouriteRepository,
	placeRepo interfaces.PlaceRepository,
	logger *slog.Logger,
) *AddFavouriteUseCase {
	return &AddFavouriteUseCase{
		favouriteRepo: favouriteRepo,
		placeRepo:     placeRepo,
		logger:        logger,
	}
}

type AddFavouriteRequest struct {
	UserID  uuid.UUID
	PlaceID uuid.UUID
}

func (uc *AddFavouriteUseCase) Execute(ctx context.Context, req AddFavouriteRequest) error {
	uc.logger.Info("adding favourite", "user_id", req.UserID, "place_id", req.PlaceID)

	_, err := uc.placeRepo.FindByID(ctx, req.PlaceID)
	if err != nil {
		uc.logger.Warn("place not found", "place_id", req.PlaceID)
		return usecase_errors.ErrPlaceNotFound
	}

	if err := uc.favouriteRepo.AddFavourite(ctx, req.UserID, req.PlaceID); err != nil {
		uc.logger.Error("failed to add favourite", "error", err)
		return err
	}

	uc.logger.Info("favourite added", "user_id", req.UserID, "place_id", req.PlaceID)
	return nil
}
