package place_repo

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

type CassandraPlaceRepo struct {
	session *gocql.Session
}

func NewCassandraPlaceRepo(session *gocql.Session) *CassandraPlaceRepo {
	return &CassandraPlaceRepo{session: session}
}

func (r *CassandraPlaceRepo) Close() error {
	return nil
}

func (r *CassandraPlaceRepo) Save(ctx context.Context, place *domain.Place) error {
	id := place.GetId()
	if id == uuid.Nil {
		id = domain.PlaceID(uuid.New())
		place.SetId(id)
	}
	cassID := gocql.UUID(id)

	coords := place.GetCoordinates()
	coordStr := fmt.Sprintf("%f,%f", coords.Lat(), coords.Lng())
	cityID := gocql.UUID(place.GetCityID())
	districtID := gocql.UUID(place.GetDistrictID())
	imageID := gocql.UUID(place.GetImageID())

	if err := r.session.Query(`
		INSERT INTO place_by_id (
			id, name, category, description, coordinates, address, opening_hours, price_info,
			avg_visit_duration_min, avg_rating, reviews_count, contact_phone, website,
			city_id, district_id, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cassID, place.GetName(), place.GetCategory(), place.GetDescription(),
		coordStr, place.GetAddress(), place.GetOpeningHours(), place.GetPriceInfo(),
		place.GetAvgVisitDurationMin(), place.GetAvgRating(), place.GetReviewCnt(),
		place.GetContactPhone(), place.GetWebsite(), cityID, districtID, imageID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert place_by_id: %w", err)
	}

	if err := r.session.Query(`
		INSERT INTO place_by_city (
			city_id, avg_rating, id, name, category, description, coordinates, address,
			opening_hours, price_info, avg_visit_duration_min, reviews_count, contact_phone,
			website, district_id, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cityID, place.GetAvgRating(), cassID, place.GetName(), place.GetCategory(),
		place.GetDescription(), coordStr, place.GetAddress(), place.GetOpeningHours(),
		place.GetPriceInfo(), place.GetAvgVisitDurationMin(), place.GetReviewCnt(),
		place.GetContactPhone(), place.GetWebsite(), districtID, imageID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM place_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert place_by_city: %w", err)
	}

	if err := r.session.Query(`
		INSERT INTO place_by_category (
			category, avg_rating, id, name, description, city_id, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, place.GetCategory(), place.GetAvgRating(), cassID, place.GetName(),
		place.GetDescription(), cityID, imageID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM place_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		_ = r.session.Query(`DELETE FROM place_by_city WHERE city_id = ? AND avg_rating = ? AND id = ?`, cityID, place.GetAvgRating(), cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert place_by_category: %w", err)
	}
	return nil
}

func (r *CassandraPlaceRepo) FindByID(ctx context.Context, id domain.PlaceID) (*domain.Place, error) {
	var (
		cassID                                                                  gocql.UUID
		name, category, description, coordStr, address, openingHours, priceInfo string
		avgVisitDurationMin                                                     int32
		avgRating                                                               float64
		reviewsCount                                                            int32
		contactPhone, website                                                   string
		cityID, districtID, imageID                                             gocql.UUID
	)
	query := r.session.Query(`
		SELECT id, name, category, description, coordinates, address, opening_hours, price_info,
		       avg_visit_duration_min, avg_rating, reviews_count, contact_phone, website,
		       city_id, district_id, image_id
		FROM place_by_id WHERE id = ? LIMIT 1
	`, gocql.UUID(id))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(&cassID, &name, &category, &description, &coordStr, &address,
		&openingHours, &priceInfo, &avgVisitDurationMin, &avgRating, &reviewsCount,
		&contactPhone, &website, &cityID, &districtID, &imageID) {
		_ = iter.Close()
		return nil, domain.ErrPlaceNotFound
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

func (r *CassandraPlaceRepo) FindByCity(ctx context.Context, cityID domain.CityID) ([]*domain.Place, error) {
	var places []*domain.Place
	iter := r.session.Query(`
		SELECT id, name, category, description, coordinates, address, opening_hours, price_info,
		       avg_visit_duration_min, avg_rating, reviews_count, contact_phone, website,
		       district_id, image_id
		FROM place_by_city WHERE city_id = ?
	`, gocql.UUID(cityID)).WithContext(ctx).Iter()
	var (
		cassID                                                                  gocql.UUID
		name, category, description, coordStr, address, openingHours, priceInfo string
		avgVisitDurationMin                                                     int32
		avgRating                                                               float64
		reviewsCount                                                            int32
		contactPhone, website                                                   string
		districtID, imageID                                                     gocql.UUID
	)
	for iter.Scan(&cassID, &name, &category, &description, &coordStr, &address,
		&openingHours, &priceInfo, &avgVisitDurationMin, &avgRating, &reviewsCount,
		&contactPhone, &website, &districtID, &imageID) {
		coords := parseCoordinates(coordStr)
		place := domain.NewPlaceFromDB(
			domain.PlaceID(cassID), name, category, description, coords,
			address, openingHours, priceInfo, avgVisitDurationMin, avgRating,
			reviewsCount, contactPhone, website, cityID,
			domain.CityDistrictID(districtID), domain.ImageID(imageID),
		)
		places = append(places, place)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return places, nil
}

func (r *CassandraPlaceRepo) FindByCategory(ctx context.Context, category string) ([]*domain.Place, error) {
	var places []*domain.Place
	iter := r.session.Query(`
		SELECT id, name, description, city_id, image_id, avg_rating
		FROM place_by_category WHERE category = ?
	`, category).WithContext(ctx).Iter()
	var (
		cassID            gocql.UUID
		name, description string
		cityID, imageID   gocql.UUID
		avgRating         float64
	)
	for iter.Scan(&cassID, &name, &description, &cityID, &imageID, &avgRating) {
		place, err := r.FindByID(ctx, domain.PlaceID(cassID))
		if err != nil {
			continue
		}
		places = append(places, place)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return places, nil
}

func (r *CassandraPlaceRepo) Update(ctx context.Context, place *domain.Place) error {
	id := gocql.UUID(place.GetId())
	cityID := gocql.UUID(place.GetCityID())
	coordStr := fmt.Sprintf("%f,%f", place.GetCoordinates().Lat(), place.GetCoordinates().Lng())
	imageID := gocql.UUID(place.GetImageID())
	districtID := gocql.UUID(place.GetDistrictID())
	old, err := r.FindByID(ctx, place.GetId())
	if err != nil {
		return err
	}
	if old == nil {
		return domain.ErrPlaceNotFound
	}
	if err := r.session.Query(`
		UPDATE place_by_id SET
			name = ?, category = ?, description = ?, coordinates = ?, address = ?,
			opening_hours = ?, price_info = ?, avg_visit_duration_min = ?,
			avg_rating = ?, reviews_count = ?, contact_phone = ?, website = ?,
			city_id = ?, district_id = ?, image_id = ?
		WHERE id = ?
	`, place.GetName(), place.GetCategory(), place.GetDescription(), coordStr,
		place.GetAddress(), place.GetOpeningHours(), place.GetPriceInfo(),
		place.GetAvgVisitDurationMin(), place.GetAvgRating(), place.GetReviewCnt(),
		place.GetContactPhone(), place.GetWebsite(), cityID, districtID, imageID, id).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update place_by_id: %w", err)
	}
	_ = r.session.Query(`DELETE FROM place_by_city WHERE city_id = ? AND avg_rating = ? AND id = ?`, gocql.UUID(old.GetCityID()), old.GetAvgRating(), id).WithContext(ctx).Exec()
	if err := r.session.Query(`
		INSERT INTO place_by_city (
			city_id, avg_rating, id, name, category, description, coordinates, address,
			opening_hours, price_info, avg_visit_duration_min, reviews_count, contact_phone,
			website, district_id, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cityID, place.GetAvgRating(), id, place.GetName(), place.GetCategory(),
		place.GetDescription(), coordStr, place.GetAddress(), place.GetOpeningHours(),
		place.GetPriceInfo(), place.GetAvgVisitDurationMin(), place.GetReviewCnt(),
		place.GetContactPhone(), place.GetWebsite(), districtID, imageID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update place_by_city: %w", err)
	}
	return nil
}

func (r *CassandraPlaceRepo) Delete(ctx context.Context, id domain.PlaceID) error {
	place, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if place == nil {
		return nil
	}
	cassID := gocql.UUID(id)
	cityID := gocql.UUID(place.GetCityID())
	if err := r.session.Query(`DELETE FROM place_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM place_by_city WHERE city_id = ? AND avg_rating = ? AND id = ?`, cityID, place.GetAvgRating(), cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM place_by_category WHERE category = ? AND avg_rating = ? AND id = ?`, place.GetCategory(), place.GetAvgRating(), cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraPlaceRepo) UpdateRating(ctx context.Context, placeID domain.PlaceID, avgRating float64, reviewsCount int) error {
	if err := r.session.Query(`UPDATE place_by_id SET avg_rating = ?, reviews_count = ? WHERE id = ?`, avgRating, reviewsCount, gocql.UUID(placeID)).WithContext(ctx).Exec(); err != nil {
		return err
	}
	place, err := r.FindByID(ctx, placeID)
	if err != nil {
		return err
	}
	if place == nil {
		return nil
	}
	oldRating := place.GetAvgRating()
	cityID := gocql.UUID(place.GetCityID())
	cassID := gocql.UUID(placeID)
	_ = r.session.Query(`DELETE FROM place_by_city WHERE city_id = ? AND avg_rating = ? AND id = ?`, cityID, oldRating, cassID).WithContext(ctx).Exec()
	coordStr := fmt.Sprintf("%f,%f", place.GetCoordinates().Lat(), place.GetCoordinates().Lng())
	if err := r.session.Query(`
		INSERT INTO place_by_city (
			city_id, avg_rating, id, name, category, description, coordinates, address,
			opening_hours, price_info, avg_visit_duration_min, reviews_count, contact_phone,
			website, district_id, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cityID, avgRating, cassID, place.GetName(), place.GetCategory(),
		place.GetDescription(), coordStr, place.GetAddress(), place.GetOpeningHours(),
		place.GetPriceInfo(), place.GetAvgVisitDurationMin(), reviewsCount,
		place.GetContactPhone(), place.GetWebsite(), gocql.UUID(place.GetDistrictID()),
		gocql.UUID(place.GetImageID())).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraPlaceRepo) SaveReview(ctx context.Context, review *domain.Review) error {
	id := review.GetId()
	if id == uuid.Nil {
		id = domain.ReviewID(uuid.New())
		review.SetId(id)
	}
	cassID := gocql.UUID(id)
	placeID := gocql.UUID(review.GetPlaceID())
	userID := gocql.UUID(review.GetUserID())
	imageID := gocql.UUID(review.GetImageID())

	if err := r.session.Query(`
		INSERT INTO review_by_id (
			id, place_id, user_id, rating, comment, visit_date, created_at,
			is_moderated, is_approved, moderation_comment, image_id, username
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cassID, placeID, userID, review.GetRating(), review.GetComment(),
		review.GetVisitDate(), review.GetCreatedAt(), review.IsModerated(),
		review.IsApproved(), review.GetModerationComment(), imageID,
		review.GetUsername()).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert review_by_id: %w", err)
	}
	if err := r.session.Query(`
		INSERT INTO review_by_place (
			place_id, created_at, id, rating, comment, visit_date, username,
			is_approved, is_moderated, moderation_comment, image_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, placeID, review.GetCreatedAt(), cassID, review.GetRating(),
		review.GetComment(), review.GetVisitDate(), review.GetUsername(),
		review.IsApproved(), review.IsModerated(), review.GetModerationComment(),
		imageID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM review_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert review_by_place: %w", err)
	}
	if err := r.session.Query(`
		INSERT INTO review_by_user (
			user_id, created_at, id, place_id, rating, comment, visit_date,
			is_approved, is_moderated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, review.GetCreatedAt(), cassID, placeID, review.GetRating(),
		review.GetComment(), review.GetVisitDate(), review.IsApproved(),
		review.IsModerated()).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM review_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		_ = r.session.Query(`DELETE FROM review_by_place WHERE place_id = ? AND created_at = ? AND id = ?`, placeID, review.GetCreatedAt(), cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert review_by_user: %w", err)
	}
	return nil
}

func (r *CassandraPlaceRepo) FindReviewByID(ctx context.Context, id domain.ReviewID) (*domain.Review, error) {
	var (
		cassID      gocql.UUID
		placeID     gocql.UUID
		userID      gocql.UUID
		rating      int
		comment     string
		visitDate   time.Time
		createdAt   time.Time
		isModerated bool
		isApproved  bool
		modComment  string
		imageID     gocql.UUID
		username    string
	)
	query := r.session.Query(`
		SELECT id, place_id, user_id, rating, comment, visit_date, created_at,
		       is_moderated, is_approved, moderation_comment, image_id, username
		FROM review_by_id WHERE id = ? LIMIT 1
	`, gocql.UUID(id))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(&cassID, &placeID, &userID, &rating, &comment, &visitDate, &createdAt,
		&isModerated, &isApproved, &modComment, &imageID, &username) {
		_ = iter.Close()
		return nil, domain.ErrReviewNotFound
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return domain.NewReviewFromDB(
		domain.ReviewID(cassID), rating, comment, visitDate, createdAt,
		isModerated, isApproved, modComment, domain.UserID(userID),
		domain.PlaceID(placeID), domain.ImageID(imageID), username,
	), nil
}

func (r *CassandraPlaceRepo) FindReviewsByPlace(ctx context.Context, placeID domain.PlaceID) ([]*domain.Review, error) {
	var reviews []*domain.Review
	iter := r.session.Query(`
        SELECT id, created_at, rating, comment, visit_date, username,
               is_approved, is_moderated, moderation_comment, image_id
        FROM review_by_place WHERE place_id = ?
    `, gocql.UUID(placeID)).WithContext(ctx).Iter()
	var (
		id                                   gocql.UUID
		createdAt                            time.Time
		rating                               int
		comment, username, moderationComment string
		visitDate                            time.Time
		isApproved, isModerated              bool
		imageID                              gocql.UUID
	)
	for iter.Scan(&id, &createdAt, &rating, &comment, &visitDate, &username,
		&isApproved, &isModerated, &moderationComment, &imageID) {
		review := domain.NewReviewFromDB(
			domain.ReviewID(id), rating, comment, visitDate, createdAt,
			isModerated, isApproved, moderationComment,
			domain.UserID(uuid.Nil), placeID,
			domain.ImageID(imageID), username,
		)
		reviews = append(reviews, review)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *CassandraPlaceRepo) FindReviewsByUser(ctx context.Context, userID domain.UserID) ([]*domain.Review, error) {
	var reviews []*domain.Review
	iter := r.session.Query(`
		SELECT id, created_at, place_id, rating, comment, visit_date,
		       is_approved, is_moderated
		FROM review_by_user WHERE user_id = ?
	`, gocql.UUID(userID)).WithContext(ctx).Iter()
	var (
		id                      gocql.UUID
		createdAt               time.Time
		placeID                 gocql.UUID
		rating                  int
		comment                 string
		visitDate               time.Time
		isApproved, isModerated bool
	)
	for iter.Scan(&id, &createdAt, &placeID, &rating, &comment, &visitDate,
		&isApproved, &isModerated) {
		review := domain.NewReviewFromDB(
			domain.ReviewID(id), rating, comment, visitDate, createdAt,
			isModerated, isApproved, "", userID, domain.PlaceID(placeID),
			domain.ImageID(uuid.Nil), "",
		)
		reviews = append(reviews, review)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *CassandraPlaceRepo) FindReviewsByUserAndPlace(ctx context.Context, userID domain.UserID, placeID domain.PlaceID) ([]*domain.Review, error) {
	userReviews, err := r.FindReviewsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Review, 0)
	for _, rev := range userReviews {
		if rev.GetPlaceID() == placeID {
			result = append(result, rev)
		}
	}
	return result, nil
}

func (r *CassandraPlaceRepo) UpdateReview(ctx context.Context, review *domain.Review) error {
	id := gocql.UUID(review.GetId())
	placeID := gocql.UUID(review.GetPlaceID())
	userID := gocql.UUID(review.GetUserID())
	if err := r.session.Query(`
		UPDATE review_by_id SET
			rating = ?, comment = ?, visit_date = ?, image_id = ?
		WHERE id = ?
	`, review.GetRating(), review.GetComment(), review.GetVisitDate(),
		gocql.UUID(review.GetImageID()), id).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`
		UPDATE review_by_place SET
			rating = ?, comment = ?, visit_date = ?, image_id = ?
		WHERE place_id = ? AND created_at = ? AND id = ?
	`, review.GetRating(), review.GetComment(), review.GetVisitDate(),
		gocql.UUID(review.GetImageID()), placeID, review.GetCreatedAt(), id).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`
		UPDATE review_by_user SET
			rating = ?, comment = ?, visit_date = ?
		WHERE user_id = ? AND created_at = ? AND id = ?
	`, review.GetRating(), review.GetComment(), review.GetVisitDate(),
		userID, review.GetCreatedAt(), id).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraPlaceRepo) DeleteReview(ctx context.Context, id domain.ReviewID) error {
	review, err := r.FindReviewByID(ctx, id)
	if err != nil {
		return err
	}
	if review == nil {
		return nil
	}
	cassID := gocql.UUID(id)
	placeID := gocql.UUID(review.GetPlaceID())
	userID := gocql.UUID(review.GetUserID())
	createdAt := review.GetCreatedAt()
	if err := r.session.Query(`DELETE FROM review_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM review_by_place WHERE place_id = ? AND created_at = ? AND id = ?`, placeID, createdAt, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM review_by_user WHERE user_id = ? AND created_at = ? AND id = ?`, userID, createdAt, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraPlaceRepo) ModerateReview(ctx context.Context, reviewID domain.ReviewID, approved bool, comment string) error {
	review, err := r.FindReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return domain.ErrReviewNotFound
	}
	cassID := gocql.UUID(reviewID)
	placeID := gocql.UUID(review.GetPlaceID())
	userID := gocql.UUID(review.GetUserID())
	createdAt := review.GetCreatedAt()
	if err := r.session.Query(`
		UPDATE review_by_id SET is_moderated = true, is_approved = ?, moderation_comment = ?
		WHERE id = ?
	`, approved, comment, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`
		UPDATE review_by_place SET is_moderated = true, is_approved = ?, moderation_comment = ?
		WHERE place_id = ? AND created_at = ? AND id = ?
	`, approved, comment, placeID, createdAt, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`
		UPDATE review_by_user SET is_moderated = true, is_approved = ?
		WHERE user_id = ? AND created_at = ? AND id = ?
	`, approved, userID, createdAt, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraPlaceRepo) GetPopularPlacesReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.PlaceReportData, error) {
	var results []domain.PlaceReportData
	iter := r.session.Query(`SELECT id, name, category, reviews_count, avg_rating FROM place_by_id`).WithContext(ctx).Iter()
	var (
		id             gocql.UUID
		name, category string
		reviewsCount   int32
		avgRating      float64
	)
	for iter.Scan(&id, &name, &category, &reviewsCount, &avgRating) {
		results = append(results, domain.PlaceReportData{
			PlaceID:      domain.PlaceID(id),
			Name:         name,
			Category:     category,
			TripsCount:   0,
			ReviewsCount: int(reviewsCount),
			AvgRating:    avgRating,
		})
		if len(results) >= limit && limit > 0 {
			break
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *CassandraPlaceRepo) FindPendingReviews(ctx context.Context, limit, offset int) ([]*domain.Review, error) {
	var reviews []*domain.Review
	iter := r.session.Query(`
		SELECT place_id, created_at, id, rating, comment, visit_date, username, image_id
		FROM review_by_place WHERE is_moderated = false ALLOW FILTERING
	`).WithContext(ctx).Iter()
	var (
		placeID           gocql.UUID
		createdAt         time.Time
		id                gocql.UUID
		rating            int
		comment, username string
		visitDate         time.Time
		imageID           gocql.UUID
	)
	for iter.Scan(&placeID, &createdAt, &id, &rating, &comment, &visitDate, &username, &imageID) {
		review := domain.NewReviewFromDB(
			domain.ReviewID(id), rating, comment, visitDate, createdAt,
			false, false, "",
			domain.UserID(uuid.Nil),
			domain.PlaceID(placeID),
			domain.ImageID(imageID),
			username,
		)
		reviews = append(reviews, review)
		if len(reviews) >= limit+offset && limit > 0 {
			break
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	if offset > 0 && len(reviews) > offset {
		reviews = reviews[offset:]
	}
	if limit > 0 && len(reviews) > limit {
		reviews = reviews[:limit]
	}
	return reviews, nil
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
