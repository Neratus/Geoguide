package trip_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type CassandraTripRepo struct {
	session *gocql.Session
}

func NewCassandraTripRepo(session *gocql.Session) *CassandraTripRepo {
	return &CassandraTripRepo{session: session}
}

func (r *CassandraTripRepo) Close() error {
	return nil
}

func (r *CassandraTripRepo) Save(ctx context.Context, trip *domain.Trip) error {
	id := trip.GetId()
	if id == uuid.Nil {
		id = domain.TripID(uuid.New())
		trip.SetId(id)
	}
	cassID := gocql.UUID(id)
	userID := gocql.UUID(trip.GetUserID())
	imageID := gocql.UUID(trip.GetImageID())

	if err := r.session.Query(`
        INSERT INTO trip_by_id (
            id, title, start_date, end_date, budget, status, notes, created_at, user_id, image_id
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, cassID, trip.GetTitle(), trip.GetStartDate(), trip.GetEndDate(),
		trip.GetBudget(), string(trip.GetStatus()), trip.GetNotes(),
		trip.GetCreatedAt(), userID, imageID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert trip_by_id: %w", err)
	}

	if err := r.session.Query(`
        INSERT INTO trip_by_user (
            user_id, start_date, id, title, end_date, budget, status, notes, created_at, image_id
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, userID, trip.GetStartDate(), cassID, trip.GetTitle(), trip.GetEndDate(),
		trip.GetBudget(), string(trip.GetStatus()), trip.GetNotes(),
		trip.GetCreatedAt(), imageID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM trip_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert trip_by_user: %w", err)
	}
	return nil
}

func (r *CassandraTripRepo) FindByID(ctx context.Context, id domain.TripID) (*domain.Trip, error) {
	var (
		cassID    gocql.UUID
		title     string
		startDate time.Time
		endDate   time.Time
		budget    float64
		status    string
		notes     string
		createdAt time.Time
		userID    gocql.UUID
		imageID   gocql.UUID
	)
	query := r.session.Query(`
        SELECT id, title, start_date, end_date, budget, status, notes, created_at, user_id, image_id
        FROM trip_by_id WHERE id = ? LIMIT 1
    `, gocql.UUID(id))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(&cassID, &title, &startDate, &endDate, &budget, &status, &notes, &createdAt, &userID, &imageID) {
		_ = iter.Close()
		return nil, domain.ErrTripNotFound
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return domain.NewTripFromDB(
		domain.TripID(cassID), title, startDate, endDate, createdAt,
		budget, status, notes, domain.UserID(userID), domain.ImageID(imageID),
	), nil
}

func (r *CassandraTripRepo) FindByUser(ctx context.Context, userID domain.UserID) ([]*domain.Trip, error) {
	var trips []*domain.Trip
	iter := r.session.Query(`
        SELECT id, title, start_date, end_date, budget, status, notes, created_at, image_id
        FROM trip_by_user WHERE user_id = ?
    `, gocql.UUID(userID)).WithContext(ctx).Iter()
	var (
		cassID    gocql.UUID
		title     string
		startDate time.Time
		endDate   time.Time
		budget    float64
		status    string
		notes     string
		createdAt time.Time
		imageID   gocql.UUID
	)
	for iter.Scan(&cassID, &title, &startDate, &endDate, &budget, &status, &notes, &createdAt, &imageID) {
		trip := domain.NewTripFromDB(
			domain.TripID(cassID), title, startDate, endDate, createdAt,
			budget, status, notes, userID, domain.ImageID(imageID),
		)
		trips = append(trips, trip)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return trips, nil
}

func (r *CassandraTripRepo) Update(ctx context.Context, trip *domain.Trip) error {
	id := gocql.UUID(trip.GetId())
	userID := gocql.UUID(trip.GetUserID())
	var oldStartDate time.Time
	if err := r.session.Query(`SELECT start_date FROM trip_by_id WHERE id = ?`, id).WithContext(ctx).Scan(&oldStartDate); err != nil {
		return err
	}

	if err := r.session.Query(`
        UPDATE trip_by_id SET
            title = ?, start_date = ?, end_date = ?, budget = ?, status = ?,
            notes = ?, user_id = ?, image_id = ?
        WHERE id = ?
    `, trip.GetTitle(), trip.GetStartDate(), trip.GetEndDate(), trip.GetBudget(),
		string(trip.GetStatus()), trip.GetNotes(), userID, gocql.UUID(trip.GetImageID()),
		id).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update trip_by_id: %w", err)
	}

	if !oldStartDate.Equal(trip.GetStartDate()) {
		_ = r.session.Query(`DELETE FROM trip_by_user WHERE user_id = ? AND start_date = ? AND id = ?`, userID, oldStartDate, id).WithContext(ctx).Exec()
	}
	if err := r.session.Query(`
        INSERT INTO trip_by_user (
            user_id, start_date, id, title, end_date, budget, status, notes, created_at, image_id
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, userID, trip.GetStartDate(), id, trip.GetTitle(), trip.GetEndDate(),
		trip.GetBudget(), string(trip.GetStatus()), trip.GetNotes(),
		trip.GetCreatedAt(), gocql.UUID(trip.GetImageID())).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update trip_by_user: %w", err)
	}
	return nil
}

func (r *CassandraTripRepo) Delete(ctx context.Context, id domain.TripID) error {
	trip, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if trip == nil {
		return nil
	}
	cassID := gocql.UUID(id)
	userID := gocql.UUID(trip.GetUserID())
	startDate := trip.GetStartDate()
	if err := r.session.Query(`DELETE FROM trip_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM trip_by_user WHERE user_id = ? AND start_date = ? AND id = ?`, userID, startDate, cassID).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraTripRepo) AddPlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID, dayNumber int, arrivalTime *time.Time, durationMin int, notes string) error {
	tripCassID := gocql.UUID(tripID)
	placeCassID := gocql.UUID(placeID)
	var count int
	if err := r.session.Query(`SELECT COUNT(*) FROM trip_by_id WHERE id = ?`, tripCassID).WithContext(ctx).Scan(&count); err != nil || count == 0 {
		return domain.ErrTripNotFound
	}
	id := gocql.UUID(uuid.New())
	if err := r.session.Query(`
        INSERT INTO trip_place_by_trip (
            trip_id, day_number, place_id, id, arrival_time, duration_min, notes, visit_status, actual_cost
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, tripCassID, dayNumber, placeCassID, id,
		arrivalTime, durationMin, notes, string(domain.TripPlaceStatusPlanned), 0.0).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed add place: %w", err)
	}
	return nil
}

func (r *CassandraTripRepo) RemovePlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID) error {
	if err := r.session.Query(`DELETE FROM trip_place_by_trip WHERE trip_id = ? AND place_id = ?`,
		gocql.UUID(tripID), gocql.UUID(placeID)).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed remove place: %w", err)
	}
	return nil
}

func (r *CassandraTripRepo) GetPlaces(ctx context.Context, tripID domain.TripID) ([]*domain.TripPlace, error) {
	var places []*domain.TripPlace
	iter := r.session.Query(`
        SELECT id, day_number, arrival_time, duration_min, notes, visit_status, actual_cost, place_id
        FROM trip_place_by_trip WHERE trip_id = ?
    `, gocql.UUID(tripID)).WithContext(ctx).Iter()
	var (
		id          gocql.UUID
		dayNumber   int
		arrivalTime *time.Time
		durationMin int
		notes       string
		visitStatus string
		actualCost  float64
		placeID     gocql.UUID
	)
	for iter.Scan(&id, &dayNumber, &arrivalTime, &durationMin, &notes, &visitStatus, &actualCost, &placeID) {
		tp := domain.NewTripPlaceFromDB(
			domain.TripPlaceID(id), dayNumber, arrivalTime, durationMin,
			notes, visitStatus, actualCost, tripID, domain.PlaceID(placeID),
		)
		places = append(places, tp)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return places, nil
}

func (r *CassandraTripRepo) UpdateTripPlace(ctx context.Context, tripPlace *domain.TripPlace) error {
	if err := r.session.Query(`DELETE FROM trip_place_by_trip WHERE trip_id = ? AND place_id = ?`,
		gocql.UUID(tripPlace.GetTripID()), gocql.UUID(tripPlace.GetPlaceID())).WithContext(ctx).Exec(); err != nil {
		return err
	}
	newID := gocql.UUID(uuid.New())
	if err := r.session.Query(`
        INSERT INTO trip_place_by_trip (
            trip_id, day_number, place_id, id, arrival_time, duration_min, notes, visit_status, actual_cost
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, gocql.UUID(tripPlace.GetTripID()), tripPlace.GetDayNumber(), gocql.UUID(tripPlace.GetPlaceID()),
		newID, tripPlace.GetArrivalTime(), tripPlace.GetDurationMin(), tripPlace.GetNotes(),
		string(tripPlace.GetVisitStatus()), tripPlace.GetActualCost()).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraTripRepo) GetTripStatisticsReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.TripStatData, error) {
	var results []domain.TripStatData
	iter := r.session.Query(`SELECT id, title, user_id, start_date, end_date FROM trip_by_id`).WithContext(ctx).Iter()
	var (
		id        gocql.UUID
		title     string
		userID    gocql.UUID
		startDate time.Time
		endDate   time.Time
	)
	for iter.Scan(&id, &title, &userID, &startDate, &endDate) {
		if dateFrom != "" {
			from, _ := time.Parse("2006-01-02", dateFrom)
			if startDate.Before(from) {
				continue
			}
		}
		if dateTo != "" {
			to, _ := time.Parse("2006-01-02", dateTo)
			if endDate.After(to) {
				continue
			}
		}
		var placesCount int
		var totalCost float64
		iter2 := r.session.Query(`SELECT COUNT(*), SUM(actual_cost) FROM trip_place_by_trip WHERE trip_id = ?`, id).WithContext(ctx).Iter()
		iter2.Scan(&placesCount, &totalCost)
		iter2.Close()
		results = append(results, domain.TripStatData{
			TripID:      domain.TripID(id),
			Title:       title,
			UserID:      domain.UserID(userID),
			StartDate:   startDate,
			EndDate:     endDate,
			PlacesCount: placesCount,
			TotalCost:   totalCost,
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
