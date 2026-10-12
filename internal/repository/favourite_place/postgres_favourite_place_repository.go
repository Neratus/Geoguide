package favour_repo

import (
	"context"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresFavouriteRepository struct {
	Pool    *pgxpool.Pool
	Queries *postgreSQL.Queries
}

func NewPostgresFavouriteRepository(cfg *config.Config) (*PostgresFavouriteRepository, error) {
	connStr := cfg.PostgresConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresFavouriteRepository{
		Pool:    pool,
		Queries: postgreSQL.New(pool),
	}, nil
}

func (r *PostgresFavouriteRepository) Close() error {
	if r.Pool != nil {
		r.Pool.Close()
	}
	return nil
}

func (r *PostgresFavouriteRepository) AddFavourite(ctx context.Context, userID, placeID uuid.UUID) error {
	u_id := domain.ToPgUUID(userID)
	p_id := domain.ToPgUUID(placeID)
	return r.Queries.AddFavorite(ctx, postgreSQL.AddFavoriteParams{
		UserID:  u_id,
		PlaceID: p_id,
	})
}

func (r *PostgresFavouriteRepository) RemoveFavourite(ctx context.Context, userID, placeID uuid.UUID) error {
	u_id := domain.ToPgUUID(userID)
	p_id := domain.ToPgUUID(placeID)
	return r.Queries.RemoveFavorite(ctx, postgreSQL.RemoveFavoriteParams{
		UserID:  u_id,
		PlaceID: p_id,
	})
}

func (r *PostgresFavouriteRepository) GetFavouritesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Place, error) {
	id := domain.ToPgUUID(userID)
	rows, err := r.Queries.GetFavoritesByUser(ctx, postgreSQL.GetFavoritesByUserParams{
		UserID: id,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.Place, len(rows))
	for i, row := range rows {
		id := domain.FromPgUUID(row.ID)
		cityID := domain.FromPgUUID(row.ID)
		districtID := domain.FromPgUUID(row.CityID)
		imageID := domain.FromPgUUID(row.ImageID)
		coords := pointToCoordinates(row.Coordinates)
		avgRating := numericToFloat64(row.AvgRating)
		place := domain.NewPlaceFromDB(
			id,
			row.Name,
			row.Category,
			row.Description.String,
			coords,
			row.Address.String,
			row.OpeningHours.String,
			row.PriceInfo.String,
			int32(row.AvgVisitDurationMin.Int32),
			avgRating,
			int32(row.ReviewsCount.Int32),
			row.ContactPhone.String,
			row.Website.String,
			cityID,
			districtID,
			imageID,
		)
		result[i] = *place
	}
	return result, nil
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}

func pointToCoordinates(p pgtype.Point) domain.Coordinates {
	if !p.Valid {
		return domain.NewCoordinates(0, 0)
	}
	return domain.NewCoordinates(p.P.X, p.P.Y)
}
