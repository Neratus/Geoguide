package favourites

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
)

type RemoveFavouriteUseCase struct {
	favouriteRepo interfaces.FavouriteRepository
	logger        *slog.Logger
}

func NewRemoveFavouriteUseCase(
	favouriteRepo interfaces.FavouriteRepository,
	logger *slog.Logger,
) *RemoveFavouriteUseCase {
	return &RemoveFavouriteUseCase{
		favouriteRepo: favouriteRepo,
		logger:        logger,
	}
}

type RemoveFavouriteRequest struct {
	UserID  uuid.UUID
	PlaceID uuid.UUID
}

func (uc *RemoveFavouriteUseCase) Execute(ctx context.Context, req RemoveFavouriteRequest) error {
	uc.logger.Info("removing favourite", "user_id", req.UserID, "place_id", req.PlaceID)
	if err := uc.favouriteRepo.RemoveFavourite(ctx, req.UserID, req.PlaceID); err != nil {
		uc.logger.Error("failed to remove favourite", "error", err)
		return err
	}
	return nil
}
