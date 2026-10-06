package favourites

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
)

type GetFavouritesUseCase struct {
	favouriteRepo interfaces.FavouriteRepository
	logger        *slog.Logger
}

func NewGetFavouritesUseCase(
	favouriteRepo interfaces.FavouriteRepository,
	logger *slog.Logger,
) *GetFavouritesUseCase {
	return &GetFavouritesUseCase{
		favouriteRepo: favouriteRepo,
		logger:        logger,
	}
}

type GetFavouritesRequest struct {
	UserID uuid.UUID
	Limit  int
	Offset int
}

type GetFavouritesResponse struct {
	Places []domain.Place
	Total  int
}

func (uc *GetFavouritesUseCase) Execute(ctx context.Context, req GetFavouritesRequest) (*GetFavouritesResponse, error) {
	uc.logger.Info("getting favourites", "user_id", req.UserID, "limit", req.Limit, "offset", req.Offset)

	places, err := uc.favouriteRepo.GetFavouritesByUser(ctx, req.UserID, req.Limit, req.Offset)
	if err != nil {
		uc.logger.Error("failed to get favourites", "error", err)
		return nil, err
	}

	return &GetFavouritesResponse{
		Places: places,
	}, nil
}
