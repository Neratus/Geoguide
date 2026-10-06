package favour_repo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type CassandraFavouriteRepository struct {
	session *gocql.Session
}

func NewCassandraFavouriteRepository(session *gocql.Session) *CassandraFavouriteRepository {
	return &CassandraFavouriteRepository{session: session}
}

func (r *CassandraFavouriteRepository) Close() error {
	return nil
}

func (r *CassandraFavouriteRepository) AddFavourite(ctx context.Context, userID, placeID uuid.UUID) error {
	place, err := r.getPlaceByID(ctx, placeID)
	if err != nil {
		return fmt.Errorf("failed to get place: %w", err)
	}
	if place == nil {
		return domain.ErrPlaceNotFound
	}
	now := time.Now()
	if err := r.session.Query(`
		INSERT INTO favourite_by_user (
			user_id, favorited_at, place_id, name, category, description, coordinates,
			address, opening_hours, price_info, avg_visit_duration_min, avg_rating,
			reviews_count, contact_phone, website, city_id, city_name, district_id,
			district_name, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, gocql.UUID(userID), now, gocql.UUID(placeID),
		place.GetName(), place.GetCategory(), place.GetDescription(),
		place.CoordStr(), place.GetAddress(), place.GetOpeningHours(), place.GetPriceInfo(),
		place.GetAvgVisitDurationMin(), place.GetAvgRating(), place.GetReviewCnt(),
		place.GetContactPhone(), place.GetWebsite(), gocql.UUID(place.GetCityID()),
		place.GetCityName(), gocql.UUID(place.GetDistrictID()), place.GetDistrictName(),
		gocql.UUID(place.GetImageID())).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to add favourite: %w", err)
	}
	return nil
}

func (r *CassandraFavouriteRepository) RemoveFavourite(ctx context.Context, userID, placeID uuid.UUID) error {
	var favoritedAt time.Time
	iter := r.session.Query(`SELECT favorited_at FROM favourite_by_user WHERE user_id = ? AND place_id = ?`,
		gocql.UUID(userID), gocql.UUID(placeID)).WithContext(ctx).Iter()
	if !iter.Scan(&favoritedAt) {
		_ = iter.Close()
		return nil
	}
	_ = iter.Close()
	if err := r.session.Query(`DELETE FROM favourite_by_user WHERE user_id = ? AND favorited_at = ? AND place_id = ?`,
		gocql.UUID(userID), favoritedAt, gocql.UUID(placeID)).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to remove favourite: %w", err)
	}
	return nil
}

func (r *CassandraFavouriteRepository) GetFavouritesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Place, error) {
	var places []domain.Place
	iter := r.session.Query(`
		SELECT place_id, name, category, description, coordinates, address, opening_hours,
		       price_info, avg_visit_duration_min, avg_rating, reviews_count, contact_phone,
		       website, city_id, city_name, district_id, district_name, image_id
		FROM favourite_by_user WHERE user_id = ?
	`, gocql.UUID(userID)).WithContext(ctx).Iter()

	var (
		cassID                                                                  gocql.UUID
		name, category, description, coordStr, address, openingHours, priceInfo string
		avgVisitDurationMin                                                     int32
		avgRating                                                               float64
		reviewsCount                                                            int32
		contactPhone, website, cityName, districtName                           string
		cityID, districtID, imageID                                             gocql.UUID
	)

	for iter.Scan(&cassID, &name, &category, &description, &coordStr, &address,
		&openingHours, &priceInfo, &avgVisitDurationMin, &avgRating, &reviewsCount,
		&contactPhone, &website, &cityID, &cityName, &districtID, &districtName, &imageID) {
		coords := parseCoordinates(coordStr)
		place := domain.NewPlaceFromDB(
			domain.PlaceID(cassID), name, category, description, coords,
			address, openingHours, priceInfo, avgVisitDurationMin, avgRating,
			reviewsCount, contactPhone, website, domain.CityID(cityID),
			domain.CityDistrictID(districtID), domain.ImageID(imageID),
		)

		places = append(places, *place)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	start := offset
	if start > len(places) {
		return []domain.Place{}, nil
	}
	end := start + limit
	if end > len(places) {
		end = len(places)
	}
	return places[start:end], nil
}

func (r *CassandraFavouriteRepository) getPlaceByID(ctx context.Context, placeID uuid.UUID) (*domain.Place, error) {
	var (
		cassID                                                                  gocql.UUID
		name, category, description, coordStr, address, openingHours, priceInfo string
		avgVisitDurationMin                                                     int32
		avgRating                                                               float64
		reviewsCount                                                            int32
		contactPhone, website, cityName, districtName                           string
		cityID, districtID, imageID                                             gocql.UUID
	)
	query := r.session.Query(`
		SELECT id, name, category, description, coordinates, address, opening_hours, price_info,
		       avg_visit_duration_min, avg_rating, reviews_count, contact_phone, website,
		       city_id, city_name, district_id, district_name, image_id
		FROM place_by_id WHERE id = ? LIMIT 1
	`, gocql.UUID(placeID))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(&cassID, &name, &category, &description, &coordStr, &address,
		&openingHours, &priceInfo, &avgVisitDurationMin, &avgRating, &reviewsCount,
		&contactPhone, &website, &cityID, &cityName, &districtID, &districtName, &imageID) {
		_ = iter.Close()
		return nil, nil
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	coords := parseCoordinates(coordStr)
	return domain.NewPlaceFromDB(
		domain.PlaceID(cassID), name, category, description, coords,
		address, openingHours, priceInfo, avgVisitDurationMin, avgRating,
		reviewsCount, contactPhone, website, domain.CityID(cityID),
		domain.CityDistrictID(districtID), domain.ImageID(imageID),
	), nil
}

func parseCoordinates(coordStr string) domain.Coordinates {
	parts := strings.Split(coordStr, ",")
	if len(parts) != 2 {
		return domain.Coordinates{}
	}
	lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return domain.NewCoordinates(lat, lng)
}
