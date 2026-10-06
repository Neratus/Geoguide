package trip_repo

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

type PostgresTripRepo struct {
	Pool    *pgxpool.Pool
	Queries *postgreSQL.Queries
}

func NewPostgresTripRepo(cfg *config.Config) (*PostgresTripRepo, error) {
	connStr := cfg.PostgresConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresTripRepo{
		Pool:    pool,
		Queries: postgreSQL.New(pool),
	}, nil
}

func (r *PostgresTripRepo) Close() error {
	if r.Pool != nil {
		r.Pool.Close()
	}
	return nil
}

func (r *PostgresTripRepo) Save(ctx context.Context, trip *domain.Trip) error {
	tripParams := toSaveTripParams(trip)
	id, err := r.Queries.SaveTrip(ctx, tripParams)
	if err != nil {
		return fmt.Errorf("failed to save trip: %w", err)
	}
	trip.SetId(domain.FromPgUUID(id))
	return err
}

func (r *PostgresTripRepo) FindByID(ctx context.Context, id domain.TripID) (*domain.Trip, error) {
	dbtrip, err := r.Queries.FindTripByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find trip by ID: %w", err)
	}
	return toDomainTrip(dbtrip)
}

func (r *PostgresTripRepo) FindByUser(ctx context.Context, userID domain.UserID) ([]*domain.Trip, error) {
	dbtrip, err := r.Queries.FindTripsByUser(ctx, domain.ToPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to find trip by user: %w", err)
	}
	var result []*domain.Trip
	for _, c := range dbtrip {
		city, err := toDomainTrip(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert trip: %w", err)
		}
		result = append(result, city)
	}
	return result, nil
}

func (r *PostgresTripRepo) Update(ctx context.Context, trip *domain.Trip) error {
	params := toUpdateTripParams(trip)
	err := r.Queries.UpdateTrip(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update trip: %w", err)
	}
	return nil
}

func (r *PostgresTripRepo) GetTripStatisticsReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.TripStatData, error) {
	params := postgreSQL.GetTripStatisticsReportParams{
		DateFrom: pgtype.Text{String: dateFrom, Valid: true},
		DateTo:   pgtype.Text{String: dateTo, Valid: true},
		Limit:    int32(limit),
	}
	rows, err := r.Queries.GetTripStatisticsReport(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip statistics report: %w", err)
	}

	result := make([]domain.TripStatData, 0, len(rows))
	for _, row := range rows {
		totalCost := 0.0
		if row.TotalCost.Valid {
			f, _ := row.TotalCost.Float64Value()
			totalCost = f.Float64
		}
		result = append(result, domain.TripStatData{
			TripID:      domain.FromPgUUID(row.ID),
			Title:       row.Title,
			UserID:      domain.FromPgUUID(row.UserID),
			StartDate:   row.StartDate.Time,
			EndDate:     row.EndDate.Time,
			PlacesCount: int(row.PlacesCount),
			TotalCost:   totalCost,
		})
	}
	return result, nil
}

func (r *PostgresTripRepo) Delete(ctx context.Context, id domain.TripID) error {
	err := r.Queries.DeleteTrip(ctx, domain.ToPgUUID(id))
	return err
}

func (r *PostgresTripRepo) AddPlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID, dayNumber int, arrivalTime *time.Time, durationMin int, notes string) error {
	params := postgreSQL.AddTripPlaceParams{
		TripID:      domain.ToPgUUID(tripID),
		PlaceID:     domain.ToPgUUID(placeID),
		DayNumber:   int32(dayNumber),
		DurationMin: pgtype.Int4{Int32: int32(durationMin), Valid: true},
		Notes:       pgtype.Text{String: notes, Valid: notes != ""},
		VisitStatus: pgtype.Text{String: string(domain.TripPlaceStatusPlanned), Valid: true},
		ActualCost:  pgtype.Numeric{Valid: false},
	}
	if arrivalTime != nil {
		params.ArrivalTime = pgtype.Time{
			Microseconds: int64(arrivalTime.Hour())*1_000_000 + int64(arrivalTime.Minute())*10_000 + int64(arrivalTime.Second())*100,
			Valid:        true,
		}
	}
	_, err := r.Queries.AddTripPlace(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to add place to trip: %w", err)
	}
	return nil
}

func (r *PostgresTripRepo) RemovePlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID) error {
	params := postgreSQL.RemoveTripPlaceParams{
		TripID:  domain.ToPgUUID(tripID),
		PlaceID: domain.ToPgUUID(placeID),
	}
	err := r.Queries.RemoveTripPlace(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to remove place from trip: %w", err)
	}
	return nil
}

func (r *PostgresTripRepo) GetPlaces(ctx context.Context, tripID domain.TripID) ([]*domain.TripPlace, error) {
	dbPlaces, err := r.Queries.GetTripPlaces(ctx, domain.ToPgUUID(tripID))
	if err != nil {
		return nil, fmt.Errorf("failed to get trip places: %w", err)
	}
	result := make([]*domain.TripPlace, 0, len(dbPlaces))
	for _, dbPlace := range dbPlaces {
		place, err := toDomainTripPlace(dbPlace)
		if err != nil {
			return nil, fmt.Errorf("failed to convert trip place: %w", err)
		}
		result = append(result, place)
	}
	return result, nil
}

func (r *PostgresTripRepo) UpdateTripPlace(ctx context.Context, tripPlace *domain.TripPlace) error {
	var arrivalTime pgtype.Time
	if t := tripPlace.GetArrivalTime(); t != nil {
		arrivalTime = pgtype.Time{
			Microseconds: int64(t.Hour())*1_000_000 + int64(t.Minute())*10_000 + int64(t.Second())*100,
			Valid:        true,
		}
	}
	var actualCost pgtype.Numeric
	actualCost.Scan(tripPlace.GetActualCost())

	params := postgreSQL.UpdateTripPlaceParams{
		ID:          domain.ToPgUUID(tripPlace.GetId()),
		DayNumber:   int32(tripPlace.GetDayNumber()),
		ArrivalTime: arrivalTime,
		DurationMin: pgtype.Int4{Int32: int32(tripPlace.GetDurationMin()), Valid: true},
		Notes:       pgtype.Text{String: tripPlace.GetNotes(), Valid: true},
		VisitStatus: pgtype.Text{String: string(tripPlace.GetVisitStatus()), Valid: true},
		ActualCost:  actualCost,
	}
	err := r.Queries.UpdateTripPlace(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update trip place: %w", err)
	}
	return nil
}

func toSaveTripParams(trip *domain.Trip) postgreSQL.SaveTripParams {
	var budget pgtype.Numeric
	budgetStr := strconv.FormatFloat(trip.GetBudget(), 'f', -1, 64)
	if err := budget.Scan(budgetStr); err != nil {
		budget = pgtype.Numeric{Valid: false}
	}
	return postgreSQL.SaveTripParams{
		Title:     trip.GetTitle(),
		StartDate: pgtype.Date{Time: trip.GetStartDate(), Valid: true},
		EndDate:   pgtype.Date{Time: trip.GetEndDate(), Valid: true},
		Budget:    budget,
		Status:    pgtype.Text{String: string(trip.GetStatus()), Valid: true},
		Notes:     pgtype.Text{String: trip.GetNotes(), Valid: trip.GetNotes() != ""},
		UserID:    domain.ToPgUUID(trip.GetUserID()),
		ImageID:   domain.ToPgUUID(trip.GetImageID()),
	}
}

func toUpdateTripParams(trip *domain.Trip) postgreSQL.UpdateTripParams {
	var budget pgtype.Numeric
	budgetStr := strconv.FormatFloat(trip.GetBudget(), 'f', -1, 64)
	_ = budget.Scan(budgetStr)
	return postgreSQL.UpdateTripParams{
		ID:        domain.ToPgUUID(trip.GetId()),
		Title:     trip.GetTitle(),
		StartDate: pgtype.Date{Time: trip.GetStartDate(), Valid: true},
		EndDate:   pgtype.Date{Time: trip.GetEndDate(), Valid: true},
		Budget:    budget,
		Status:    pgtype.Text{String: string(trip.GetStatus()), Valid: true},
		Notes:     pgtype.Text{String: trip.GetNotes(), Valid: trip.GetNotes() != ""},
		ImageID:   domain.ToPgUUID(trip.GetImageID()),
	}
}
func toDomainTrip(dbTrip postgreSQL.Trip) (*domain.Trip, error) {
	id := domain.FromPgUUID(dbTrip.ID)
	if id == uuid.Nil {
		return nil, errors.New("trip id is null")
	}
	userID := domain.FromPgUUID(dbTrip.UserID)
	imageID := domain.FromPgUUID(dbTrip.ImageID)

	if !dbTrip.StartDate.Valid {
		return nil, errors.New("start_date is null")
	}
	if !dbTrip.EndDate.Valid {
		return nil, errors.New("end_date is null")
	}
	if !dbTrip.CreatedAt.Valid {
		return nil, errors.New("created_at is null")
	}
	if !dbTrip.Budget.Valid {
		return nil, errors.New("budget is null")
	}

	budgetFloat, err := dbTrip.Budget.Float64Value()
	if err != nil {
		return nil, fmt.Errorf("failed to parse budget: %w", err)
	}
	budget := budgetFloat.Float64

	status := ""
	if dbTrip.Status.Valid {
		status = dbTrip.Status.String
	}

	notes := ""
	if dbTrip.Notes.Valid {
		notes = dbTrip.Notes.String
	}

	return domain.NewTripFromDB(
		id,
		dbTrip.Title,
		dbTrip.StartDate.Time,
		dbTrip.EndDate.Time,
		dbTrip.CreatedAt.Time,
		budget,
		status,
		notes,
		userID,
		imageID,
	), nil
}

func toDomainTripPlace(dbTP postgreSQL.TripPlace) (*domain.TripPlace, error) {
	id := domain.FromPgUUID(dbTP.ID)
	if id == uuid.Nil {
		return nil, errors.New("trip_place id is null")
	}
	tripID := domain.FromPgUUID(dbTP.TripID)
	placeID := domain.FromPgUUID(dbTP.PlaceID)

	var arrivalTime *time.Time
	if dbTP.ArrivalTime.Valid {
		microseconds := dbTP.ArrivalTime.Microseconds
		hour := microseconds / 1_000_000
		minute := (microseconds % 1_000_000) / 10_000
		second := (microseconds % 10_000) / 100
		t := time.Date(0, 1, 1, int(hour), int(minute), int(second), 0, time.UTC)
		arrivalTime = &t
	}

	durationMin := 0
	if dbTP.DurationMin.Valid {
		durationMin = int(dbTP.DurationMin.Int32)
	}

	actualCost := 0.0
	if dbTP.ActualCost.Valid {
		f, err := dbTP.ActualCost.Float64Value()
		if err == nil {
			actualCost = f.Float64
		}
	}

	notes := ""
	if dbTP.Notes.Valid {
		notes = dbTP.Notes.String
	}

	status := string(domain.TripPlaceStatusPlanned)
	if dbTP.VisitStatus.Valid {
		status = dbTP.VisitStatus.String
	}

	return domain.NewTripPlaceFromDB(
		id,
		int(dbTP.DayNumber),
		arrivalTime,
		durationMin,
		notes,
		status,
		actualCost,
		tripID,
		placeID,
	), nil
}
