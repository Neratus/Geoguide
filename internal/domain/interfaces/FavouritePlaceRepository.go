package interfaces

import (
	"context"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type FavouriteRepository interface {
	AddFavourite(ctx context.Context, userID, placeID uuid.UUID) error
	RemoveFavourite(ctx context.Context, userID, placeID uuid.UUID) error
	GetFavouritesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Place, error)
}
