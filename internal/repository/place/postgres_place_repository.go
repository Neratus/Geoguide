package place_repo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPlaceRepo struct {
	Pool    *pgxpool.Pool
	Queries *postgreSQL.Queries
}

func NewPostgresPlaceRepo(cfg *config.Config) (*PostgresPlaceRepo, error) {
	connStr := cfg.PostgresConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresPlaceRepo{
		Pool:    pool,
		Queries: postgreSQL.New(pool),
	}, nil
}

func (r *PostgresPlaceRepo) Close() error {
	if r.Pool != nil {
		r.Pool.Close()
	}
	return nil
}

func (r *PostgresPlaceRepo) Save(ctx context.Context, place *domain.Place) error {
	params := toSavePlaceParams(place)
	id, err := r.Queries.SavePlace(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to save place: %w", err)
	}
	place.SetId(domain.FromPgUUID(id))
	return nil
}

func (r *PostgresPlaceRepo) FindByID(ctx context.Context, id domain.PlaceID) (*domain.Place, error) {
	dbPlace, err := r.Queries.FindPlaceByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find place by ID: %w", err)
	}
	return toDomainPlaceFromRow(dbPlace)
}

func (r *PostgresPlaceRepo) FindByCity(ctx context.Context, cityID domain.CityID) ([]*domain.Place, error) {
	dbPlaces, err := r.Queries.FindPlacesByCity(ctx, domain.ToPgUUID(cityID))
	if err != nil {
		return nil, fmt.Errorf("failed to find places by city: %w", err)
	}
	var result []*domain.Place
	for _, p := range dbPlaces {
		place, err := toDomainPlaceFromCityRow(p)
		if err != nil {
			return nil, err
		}
		result = append(result, place)
	}
	return result, nil
}

func (r *PostgresPlaceRepo) FindReviewsByUserAndPlace(ctx context.Context, userID domain.UserID, placeID domain.PlaceID) ([]*domain.Review, error) {
	rows, err := r.Queries.FindReviewsByUserAndPlace(ctx, postgreSQL.FindReviewsByUserAndPlaceParams{
		UserID:  domain.ToPgUUID(userID),
		PlaceID: domain.ToPgUUID(placeID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by user and place: %w", err)
	}
	reviews := make([]*domain.Review, 0, len(rows))
	for _, row := range rows {
		review, err := toDomainReviewFromReview(row)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (r *PostgresPlaceRepo) FindByCategory(ctx context.Context, category string) ([]*domain.Place, error) {
	dbPlaces, err := r.Queries.FindPlacesByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to find places by category: %w", err)
	}
	var result []*domain.Place
	for _, p := range dbPlaces {
		place, err := toDomainPlaceFromCategoryRow(p)
		if err != nil {
			return nil, err
		}
		result = append(result, place)
	}
	return result, nil
}

func (r *PostgresPlaceRepo) Update(ctx context.Context, place *domain.Place) error {
	params := toUpdatePlaceParams(place)
	err := r.Queries.UpdatePlace(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update place: %w", err)
	}
	return nil
}

func (r *PostgresPlaceRepo) Delete(ctx context.Context, id domain.PlaceID) error {
	err := r.Queries.DeletePlace(ctx, domain.ToPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete place: %w", err)
	}
	return nil
}

func (r *PostgresPlaceRepo) UpdateRating(ctx context.Context, id domain.PlaceID, newAvgRating float64, newReviewsCount int) error {
	ratingStr := strconv.FormatFloat(newAvgRating, 'f', -1, 64)
	var avgRating pgtype.Numeric
	if err := avgRating.Scan(ratingStr); err != nil {
		return fmt.Errorf("failed to convert rating: %w", err)
	}
	params := postgreSQL.UpdateRatingParams{
		ID:           domain.ToPgUUID(id),
		AvgRating:    avgRating,
		ReviewsCount: pgtype.Int4{Int32: int32(newReviewsCount), Valid: true},
	}
	err := r.Queries.UpdateRating(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}
	return nil
}

func (r *PostgresPlaceRepo) SaveReview(ctx context.Context, review *domain.Review) error {
	params := toSaveReviewParams(review)
	id, err := r.Queries.SaveReview(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to save review: %w", err)
	}
	review.SetId(domain.FromPgUUID(id))
	return nil
}

func (r *PostgresPlaceRepo) FindReviewByID(ctx context.Context, id domain.ReviewID) (*domain.Review, error) {
	dbReview, err := r.Queries.FindReviewByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find review by ID: %w", err)
	}
	return toDomainReviewFromReview(dbReview)
}

func (r *PostgresPlaceRepo) FindReviewsByPlace(ctx context.Context, placeID domain.PlaceID) ([]*domain.Review, error) {
	dbReviews, err := r.Queries.FindReviewsByPlace(ctx, domain.ToPgUUID(placeID))
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by place: %w", err)
	}
	result := make([]*domain.Review, 0, len(dbReviews))
	for _, rw := range dbReviews {
		review, err := toDomainReviewFromReview(rw)
		if err != nil {
			return nil, err
		}
		result = append(result, review)
	}
	return result, nil
}

func (r *PostgresPlaceRepo) GetPopularPlacesReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.PlaceReportData, error) {
	rows, err := r.Queries.GetPopularPlacesReport(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get popular places report: %w", err)
	}

	result := make([]domain.PlaceReportData, 0, len(rows))
	for _, row := range rows {
		avgRating := 0.0
		if row.AvgRating.Valid {
			f, _ := row.AvgRating.Float64Value()
			avgRating = f.Float64
		}
		result = append(result, domain.PlaceReportData{
			PlaceID:      domain.FromPgUUID(row.ID),
			Name:         row.Name,
			Category:     row.Category,
			TripsCount:   int(row.TripsCount),
			ReviewsCount: int(row.ReviewsCount),
			AvgRating:    avgRating,
		})
	}
	return result, nil
}

func (r *PostgresPlaceRepo) FindReviewsByUser(ctx context.Context, userID domain.UserID) ([]*domain.Review, error) {
	dbReviews, err := r.Queries.FindReviewsByUser(ctx, domain.ToPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by user: %w", err)
	}
	result := make([]*domain.Review, 0, len(dbReviews))
	for _, rw := range dbReviews {
		review, err := toDomainReviewFromReview(rw)
		if err != nil {
			return nil, err
		}
		result = append(result, review)
	}
	return result, nil
}

func (r *PostgresPlaceRepo) UpdateReview(ctx context.Context, review *domain.Review) error {
	params := toUpdateReviewParams(review)
	err := r.Queries.UpdateReview(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}
	return nil
}

func (r *PostgresPlaceRepo) DeleteReview(ctx context.Context, id domain.ReviewID) error {
	err := r.Queries.DeleteReview(ctx, domain.ToPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	return nil
}

func (r *PostgresPlaceRepo) ModerateReview(ctx context.Context, id domain.ReviewID, approved bool, moderationComment string) error {
	params := postgreSQL.ModerateReviewParams{
		ID:                domain.ToPgUUID(id),
		IsApproved:        pgtype.Bool{Bool: approved, Valid: true},
		ModerationComment: pgtype.Text{String: moderationComment, Valid: moderationComment != ""},
	}
	err := r.Queries.ModerateReview(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to moderate review: %w", err)
	}
	return nil
}

func (r *PostgresPlaceRepo) FindPendingReviews(ctx context.Context, limit, offset int) ([]*domain.Review, error) {
	params := postgreSQL.FindPendingReviewsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
	rows, err := r.Queries.FindPendingReviews(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reviews: %w", err)
	}
	reviews := make([]*domain.Review, 0, len(rows))
	for _, row := range rows {
		review, err := toDomainReviewFromPendingRow(row)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func toDomainReviewFromPendingRow(row postgreSQL.FindPendingReviewsRow) (*domain.Review, error) {
	id := domain.FromPgUUID(row.ID)
	if id == uuid.Nil {
		return nil, errors.New("review id is null")
	}
	userID := domain.FromPgUUID(row.UserID)
	placeID := domain.FromPgUUID(row.PlaceID)
	imageID := domain.FromPgUUID(row.ImageID)

	rating := int(row.Rating)
	comment := ""
	if row.Comment.Valid {
		comment = row.Comment.String
	}
	visitDate := time.Time{}
	if row.VisitDate.Valid {
		visitDate = row.VisitDate.Time
	}
	createdAt := time.Now()
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	isModerated := false
	if row.IsModerated.Valid {
		isModerated = row.IsModerated.Bool
	}
	isApproved := false
	if row.IsApproved.Valid {
		isApproved = row.IsApproved.Bool
	}
	moderationComment := ""
	if row.ModerationComment.Valid {
		moderationComment = row.ModerationComment.String
	}
	username := row.Username

	return domain.NewReviewFromDB(
		id,
		rating,
		comment,
		visitDate,
		createdAt,
		isModerated,
		isApproved,
		moderationComment,
		userID,
		placeID,
		imageID,
		username,
	), nil
}

func toSavePlaceParams(place *domain.Place) postgreSQL.SavePlaceParams {
	var avg pgtype.Numeric
	avg.Scan(place.GetAvgRating())

	var districtIDParam pgtype.UUID
	if place.GetDistrictID() == uuid.Nil {
		districtIDParam = pgtype.UUID{Valid: false}
	} else {
		districtIDParam = pgtype.UUID{Bytes: place.GetDistrictID(), Valid: true}
	}

	var cityIDParam pgtype.UUID
	if place.GetCityID() == uuid.Nil {
		cityIDParam = pgtype.UUID{Valid: false}
	} else {
		cityIDParam = pgtype.UUID{Bytes: place.GetCityID(), Valid: true}
	}

	coords := place.GetCoordinates()
	return postgreSQL.SavePlaceParams{
		Name:                place.GetName(),
		Category:            place.GetCategory(),
		Description:         pgtype.Text{String: place.GetDescription(), Valid: place.GetDescription() != ""},
		Coordinates:         pgtype.Point{P: pgtype.Vec2{X: coords.Lng(), Y: coords.Lat()}, Valid: true},
		Address:             pgtype.Text{String: place.GetAddress(), Valid: place.GetAddress() != ""},
		OpeningHours:        pgtype.Text{String: place.GetOpeningHours(), Valid: place.GetOpeningHours() != ""},
		PriceInfo:           pgtype.Text{String: place.GetPriceInfo(), Valid: place.GetPriceInfo() != ""},
		AvgVisitDurationMin: pgtype.Int4{Int32: int32(place.GetAvgVisitDurationMin()), Valid: true},
		AvgRating:           avg,
		ReviewsCount:        pgtype.Int4{Int32: int32(place.GetReviewCnt()), Valid: true},
		ContactPhone:        pgtype.Text{String: place.GetContactPhone(), Valid: place.GetContactPhone() != ""},
		Website:             pgtype.Text{String: place.GetWebsite(), Valid: place.GetWebsite() != ""},
		CityID:              cityIDParam,
		DistrictID:          districtIDParam,
		ImageID:             domain.ToPgUUID(place.GetImageID()),
	}
}

func toUpdatePlaceParams(place *domain.Place) postgreSQL.UpdatePlaceParams {
	var districtIDParam pgtype.UUID
	if place.GetDistrictID() == uuid.Nil {
		districtIDParam = pgtype.UUID{Valid: false}
	} else {
		districtIDParam = pgtype.UUID{Bytes: place.GetDistrictID(), Valid: true}
	}

	var cityIDParam pgtype.UUID
	if place.GetCityID() == uuid.Nil {
		cityIDParam = pgtype.UUID{Valid: false}
	} else {
		cityIDParam = pgtype.UUID{Bytes: place.GetCityID(), Valid: true}
	}

	var avg pgtype.Numeric
	avg.Scan(place.GetAvgRating())
	coords := place.GetCoordinates()
	return postgreSQL.UpdatePlaceParams{
		ID:                  domain.ToPgUUID(place.GetId()),
		Name:                place.GetName(),
		Category:            place.GetCategory(),
		Description:         pgtype.Text{String: place.GetDescription(), Valid: place.GetDescription() != ""},
		Coordinates:         pgtype.Point{P: pgtype.Vec2{X: coords.Lng(), Y: coords.Lat()}, Valid: true},
		Address:             pgtype.Text{String: place.GetAddress(), Valid: place.GetAddress() != ""},
		OpeningHours:        pgtype.Text{String: place.GetOpeningHours(), Valid: place.GetOpeningHours() != ""},
		PriceInfo:           pgtype.Text{String: place.GetPriceInfo(), Valid: place.GetPriceInfo() != ""},
		AvgVisitDurationMin: pgtype.Int4{Int32: int32(place.GetAvgVisitDurationMin()), Valid: true},
		AvgRating:           avg,
		ReviewsCount:        pgtype.Int4{Int32: int32(place.GetReviewCnt()), Valid: true},
		ContactPhone:        pgtype.Text{String: place.GetContactPhone(), Valid: place.GetContactPhone() != ""},
		Website:             pgtype.Text{String: place.GetWebsite(), Valid: place.GetWebsite() != ""},
		CityID:              cityIDParam,
		DistrictID:          districtIDParam,
		ImageID:             domain.ToPgUUID(place.GetImageID()),
	}
}

func toDomainPlace(dbPlace postgreSQL.Place) (*domain.Place, error) {
	id := domain.FromPgUUID(dbPlace.ID)
	if id == uuid.Nil {
		return nil, errors.New("place id is null")
	}
	cityID := domain.FromPgUUID(dbPlace.CityID)
	districtID := domain.FromPgUUID(dbPlace.DistrictID)
	imageID := domain.FromPgUUID(dbPlace.ImageID)

	coords := domain.NewCoordinates(dbPlace.Coordinates.P.Y, dbPlace.Coordinates.P.X)

	avgRating := 0.0
	if dbPlace.AvgRating.Valid {
		f, _ := dbPlace.AvgRating.Float64Value()
		avgRating = f.Float64
	}
	reviewsCount := int32(0)
	if dbPlace.ReviewsCount.Valid {
		reviewsCount = dbPlace.ReviewsCount.Int32
	}
	avgVisitDuration := int32(0)
	if dbPlace.AvgVisitDurationMin.Valid {
		avgVisitDuration = dbPlace.AvgVisitDurationMin.Int32
	}

	return domain.NewPlaceFromDB(
		id,
		dbPlace.Name,
		dbPlace.Category,
		dbPlace.Description.String,
		coords,
		dbPlace.Address.String,
		dbPlace.OpeningHours.String,
		dbPlace.PriceInfo.String,
		avgVisitDuration,
		avgRating,
		reviewsCount,
		dbPlace.ContactPhone.String,
		dbPlace.Website.String,
		cityID,
		districtID,
		imageID,
	), nil
}

func toSaveReviewParams(review *domain.Review) postgreSQL.SaveReviewParams {
	return postgreSQL.SaveReviewParams{
		Rating:    int32(review.GetRating()),
		Comment:   pgtype.Text{String: review.GetComment(), Valid: review.GetComment() != ""},
		VisitDate: pgtype.Date{Time: review.GetVisitDate(), Valid: true},
		UserID:    domain.ToPgUUID(review.GetUserID()),
		PlaceID:   domain.ToPgUUID(review.GetPlaceID()),
		ImageID:   domain.ToPgUUID(review.GetImageID()),
	}
}

func toUpdateReviewParams(review *domain.Review) postgreSQL.UpdateReviewParams {
	return postgreSQL.UpdateReviewParams{
		ID:        domain.ToPgUUID(review.GetId()),
		Rating:    int32(review.GetRating()),
		Comment:   pgtype.Text{String: review.GetComment(), Valid: review.GetComment() != ""},
		VisitDate: pgtype.Date{Time: review.GetVisitDate(), Valid: true},
		ImageID:   domain.ToPgUUID(review.GetImageID()),
	}
}
func toDomainReviewFromReview(dbReview postgreSQL.Review) (*domain.Review, error) {
	id := domain.FromPgUUID(dbReview.ID)
	if id == uuid.Nil {
		return nil, errors.New("review id is null")
	}
	userID := domain.FromPgUUID(dbReview.UserID)
	placeID := domain.FromPgUUID(dbReview.PlaceID)
	imageID := domain.FromPgUUID(dbReview.ImageID)

	rating := int(dbReview.Rating)
	comment := ""
	if dbReview.Comment.Valid {
		comment = dbReview.Comment.String
	}
	visitDate := time.Time{}
	if dbReview.VisitDate.Valid {
		visitDate = dbReview.VisitDate.Time
	}
	createdAt := time.Now()
	if dbReview.CreatedAt.Valid {
		createdAt = dbReview.CreatedAt.Time
	}
	isModerated := false
	if dbReview.IsModerated.Valid {
		isModerated = dbReview.IsModerated.Bool
	}
	isApproved := false
	if dbReview.IsApproved.Valid {
		isApproved = dbReview.IsApproved.Bool
	}
	moderationComment := ""
	if dbReview.ModerationComment.Valid {
		moderationComment = dbReview.ModerationComment.String
	}

	return domain.NewReviewFromDB(
		id,
		rating,
		comment,
		visitDate,
		createdAt,
		isModerated,
		isApproved,
		moderationComment,
		userID,
		placeID,
		imageID,
		"",
	), nil
}
func toDomainPlaceFromCityRow(row postgreSQL.FindPlacesByCityRow) (*domain.Place, error) {
	id := domain.FromPgUUID(row.ID)
	if id == uuid.Nil {
		return nil, errors.New("place id is null")
	}
	cityID := domain.FromPgUUID(row.CityID)
	districtID := domain.FromPgUUID(row.DistrictID)
	imageID := domain.FromPgUUID(row.ImageID)

	coords := domain.NewCoordinates(row.Coordinates.P.Y, row.Coordinates.P.X)

	avgRating := 0.0
	if row.AvgRating.Valid {
		f, _ := row.AvgRating.Float64Value()
		avgRating = f.Float64
	}
	reviewsCount := int32(0)
	if row.ReviewsCount.Valid {
		reviewsCount = row.ReviewsCount.Int32
	}
	avgVisitDuration := int32(0)
	if row.AvgVisitDurationMin.Valid {
		avgVisitDuration = row.AvgVisitDurationMin.Int32
	}

	return domain.NewPlaceFromDB(
		id,
		row.Name,
		row.Category,
		row.Description.String,
		coords,
		row.Address.String,
		row.OpeningHours.String,
		row.PriceInfo.String,
		avgVisitDuration,
		avgRating,
		reviewsCount,
		row.ContactPhone.String,
		row.Website.String,
		cityID,
		districtID,
		imageID,
	), nil
}

func toDomainPlaceFromCategoryRow(row postgreSQL.FindPlacesByCategoryRow) (*domain.Place, error) {
	id := domain.FromPgUUID(row.ID)
	if id == uuid.Nil {
		return nil, errors.New("place id is null")
	}
	cityID := domain.FromPgUUID(row.CityID)
	districtID := domain.FromPgUUID(row.DistrictID)
	imageID := domain.FromPgUUID(row.ImageID)

	coords := domain.NewCoordinates(row.Coordinates.P.Y, row.Coordinates.P.X)

	avgRating := 0.0
	if row.AvgRating.Valid {
		f, _ := row.AvgRating.Float64Value()
		avgRating = f.Float64
	}
	reviewsCount := int32(0)
	if row.ReviewsCount.Valid {
		reviewsCount = row.ReviewsCount.Int32
	}
	avgVisitDuration := int32(0)
	if row.AvgVisitDurationMin.Valid {
		avgVisitDuration = row.AvgVisitDurationMin.Int32
	}

	return domain.NewPlaceFromDB(
		id,
		row.Name,
		row.Category,
		row.Description.String,
		coords,
		row.Address.String,
		row.OpeningHours.String,
		row.PriceInfo.String,
		avgVisitDuration,
		avgRating,
		reviewsCount,
		row.ContactPhone.String,
		row.Website.String,
		cityID,
		districtID,
		imageID,
	), nil
}

func toDomainPlaceFromRow(row postgreSQL.FindPlaceByIDRow) (*domain.Place, error) {
	id := domain.FromPgUUID(row.ID)
	if id == uuid.Nil {
		return nil, errors.New("place id is null")
	}
	cityID := domain.FromPgUUID(row.CityID)
	districtID := domain.FromPgUUID(row.DistrictID)
	imageID := domain.FromPgUUID(row.ImageID)
	coords := domain.NewCoordinates(row.Coordinates.P.Y, row.Coordinates.P.X)
	avgRating := 0.0
	if row.AvgRating.Valid {
		f, _ := row.AvgRating.Float64Value()
		avgRating = f.Float64
	}
	reviewsCount := int32(0)
	if row.ReviewsCount.Valid {
		reviewsCount = row.ReviewsCount.Int32
	}
	avgVisitDuration := int32(0)
	if row.AvgVisitDurationMin.Valid {
		avgVisitDuration = row.AvgVisitDurationMin.Int32
	}
	return domain.NewPlaceFromDB(
		id,
		row.Name,
		row.Category,
		row.Description.String,
		coords,
		row.Address.String,
		row.OpeningHours.String,
		row.PriceInfo.String,
		avgVisitDuration,
		avgRating,
		reviewsCount,
		row.ContactPhone.String,
		row.Website.String,
		cityID,
		districtID,
		imageID,
	), nil
}
