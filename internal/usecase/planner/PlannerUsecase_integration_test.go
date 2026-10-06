//go:build postgres
// +build postgres

package planner

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	city_repo "github.com/Neratus/geoguide/internal/repository/city"
	configPkg "github.com/Neratus/geoguide/internal/repository/config"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
	place_repo "github.com/Neratus/geoguide/internal/repository/place"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	trip_repo "github.com/Neratus/geoguide/internal/repository/trip"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

var testPool *pgxpool.Pool
var testQueries *postgreSQL.Queries

func setupPostgresForMain(baseConnStr, migrationsPath string) (*pgxpool.Pool, func(), error) {
	ctx := context.Background()
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	admin, err := pgx.Connect(ctx, baseConnStr)
	if err != nil {
		return nil, nil, fmt.Errorf("admin connect: %w", err)
	}
	if _, err := admin.Exec(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		_ = admin.Close(ctx)
		return nil, nil, fmt.Errorf("create schema: %w", err)
	}
	_ = admin.Close(ctx)

	poolCfg, err := pgxpool.ParseConfig(baseConnStr)
	if err != nil {
		return nil, nil, err
	}
	if poolCfg.ConnConfig.RuntimeParams == nil {
		poolCfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolCfg.ConnConfig.RuntimeParams["search_path"] = schema

	sqlDB := stdlib.OpenDB(*poolCfg.ConnConfig)
	if err := goose.SetDialect("postgres"); err != nil {
		sqlDB.Close()
		return nil, nil, err
	}
	if err := goose.Up(sqlDB, migrationsPath); err != nil {
		sqlDB.Close()
		return nil, nil, err
	}
	sqlDB.Close()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, nil, err
	}

	teardown := func() {
		pool.Close()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		conn, err := pgx.Connect(c, baseConnStr)
		if err == nil {
			_, _ = conn.Exec(c, `DROP SCHEMA "`+schema+`" CASCADE`)
			_ = conn.Close(c)
		}
	}
	return pool, teardown, nil
}

func TestMain(m *testing.M) {
	cfg, err := configPkg.LoadTest()
	if err != nil {
		log.Fatalf("Failed to load test config: %v", err)
	}

	pool, teardown, err := setupPostgresForMain(cfg.PostgresConnString(), "../../../migrations")
	if err != nil {
		log.Fatalf("setup postgres: %v", err)
	}

	testPool = pool
	testQueries = postgreSQL.New(pool)

	var testImageID domain.ImageID
	testSlug := "test-planner-image-" + uuid.NewString()[:8]
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO "StaticPage" (slug, title) VALUES ($1, 'Test') RETURNING id`, testSlug).
		Scan(&testImageID); err != nil {
		log.Fatalf("Failed to create test StaticPage: %v", err)
	}
	builders.DefaultImageID = testImageID

	domain.SetConfig(domain.GetConfig())

	code := m.Run()
	teardown()
	os.Exit(code)
}

func setupTest(t *testing.T) (
	*trip_repo.PostgresTripRepo,
	*user_repo.PostgresUserRepo,
	*place_repo.PostgresPlaceRepo,
	*city_repo.PostgresCityRepo,
	*country_repo.PostgresCountryRepo,
) {
	t.Helper()

	tripRepo := &trip_repo.PostgresTripRepo{Pool: testPool, Queries: testQueries}
	userRepo := &user_repo.PostgresUserRepo{Pool: testPool, Queries: testQueries}
	placeRepo := &place_repo.PostgresPlaceRepo{Pool: testPool, Queries: testQueries}
	cityRepo := &city_repo.PostgresCityRepo{Pool: testPool, Queries: testQueries}
	countryRepo := &country_repo.PostgresCountryRepo{Pool: testPool, Queries: testQueries}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tables := []string{"TripPlace", "Trip", "Place", "City", "Country", "User"}
		for _, table := range tables {
			_, _ = testPool.Exec(ctx, "TRUNCATE TABLE \""+table+"\" RESTART IDENTITY CASCADE;")
		}
	})

	return tripRepo, userRepo, placeRepo, cityRepo, countryRepo
}

func TestCreateTripUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	tripRepo, userRepo, _, _, _ := setupTest(t)

	uc := NewCreateTripUseCase(tripRepo, logger)

	t.Run("Success_CreateTrip", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		req := requests.CreateTripRequest{
			UserID:    user.GetId(),
			Title:     "Поездка в Париж",
			StartDate: time.Now().Add(24 * time.Hour),
			EndDate:   time.Now().Add(48 * time.Hour),
			Budget:    5000.0,
			Notes:     "Отдых",
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "Поездка в Париж", resp.Title)
		require.Equal(t, "DRAFT", resp.Status)

		savedTrip, err := tripRepo.FindByID(ctx, resp.ID)
		require.NoError(t, err)
		require.Equal(t, user.GetId(), savedTrip.GetUserID())
		require.Equal(t, "Поездка в Париж", savedTrip.GetTitle())
	})
}

func TestGetTripUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	tripRepo, userRepo, _, _, _ := setupTest(t)

	uc := NewGetTripUseCase(tripRepo, logger)

	t.Run("Success_GetTrip", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trip := builders.NewTripBuilder().
			WithUserID(user.GetId()).
			WithTitle("Тестовая поездка").
			Build()
		require.NoError(t, tripRepo.Save(ctx, trip))

		req := requests.GetTripRequest{
			TripID: trip.GetId(),
			UserID: user.GetId(),
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "Тестовая поездка", resp.Title)
	})

	t.Run("Error_TripNotFound", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		req := requests.GetTripRequest{
			TripID: domain.TripID(uuid.New()),
			UserID: user.GetId(),
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrTripNotFound)
		require.Nil(t, resp)
	})

	t.Run("Error_Unauthorized", func(t *testing.T) {
		owner := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, owner))

		requester := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, requester))

		trip := builders.NewTripBuilder().
			WithUserID(owner.GetId()).
			WithTitle("Чужая поездка").
			Build()
		require.NoError(t, tripRepo.Save(ctx, trip))

		req := requests.GetTripRequest{
			TripID: trip.GetId(),
			UserID: requester.GetId(),
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrTripNotFound)
		require.Nil(t, resp)
	})
}

func TestGetUserTripsUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	tripRepo, userRepo, _, _, _ := setupTest(t)

	uc := NewGetUserTripsUseCase(tripRepo, logger)

	t.Run("Success_GetUserTrips", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trip1 := builders.NewTripBuilder().WithUserID(user.GetId()).WithTitle("Trip 1").Build()
		trip2 := builders.NewTripBuilder().WithUserID(user.GetId()).WithTitle("Trip 2").Build()
		require.NoError(t, tripRepo.Save(ctx, trip1))
		require.NoError(t, tripRepo.Save(ctx, trip2))

		req := requests.GetUserTripsRequest{UserID: user.GetId()}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 2)
	})
}

func TestAddPlaceToTripUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	tripRepo, userRepo, placeRepo, cityRepo, countryRepo := setupTest(t)

	uc := NewAddPlaceToTripUseCase(tripRepo, placeRepo, logger)

	t.Run("Success_AddPlace", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		trip := builders.NewTripBuilder().WithUserID(user.GetId()).Build()
		require.NoError(t, tripRepo.Save(ctx, trip))

		req := requests.AddPlaceToTripRequest{
			TripID:      trip.GetId(),
			PlaceID:     place.GetId(),
			UserID:      user.GetId(),
			DayNumber:   1,
			DurationMin: 120,
		}

		err := uc.Execute(ctx, req)

		require.NoError(t, err)

		places, err := tripRepo.GetPlaces(ctx, trip.GetId())
		require.NoError(t, err)
		require.Len(t, places, 1)
		require.Equal(t, place.GetId(), places[0].GetPlaceID())
	})

	t.Run("Error_PlaceNotFound", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trip := builders.NewTripBuilder().WithUserID(user.GetId()).Build()
		require.NoError(t, tripRepo.Save(ctx, trip))

		req := requests.AddPlaceToTripRequest{
			TripID:  trip.GetId(),
			PlaceID: domain.PlaceID(uuid.New()),
			UserID:  user.GetId(),
		}

		err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrPlaceNotFound)
	})

	t.Run("Error_TripNotFound", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		req := requests.AddPlaceToTripRequest{
			TripID:  domain.TripID(uuid.New()),
			PlaceID: place.GetId(),
			UserID:  user.GetId(),
		}

		err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrTripNotFound)
	})
}

func TestDeleteTripUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	tripRepo, userRepo, _, _, _ := setupTest(t)

	uc := NewDeleteTripUseCase(tripRepo, logger)

	t.Run("Success_DeleteTrip", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trip := builders.NewTripBuilder().WithUserID(user.GetId()).Build()
		require.NoError(t, tripRepo.Save(ctx, trip))

		req := requests.DeleteTripRequest{TripID: trip.GetId()}

		err := uc.Execute(ctx, req)

		require.NoError(t, err)

		_, err = tripRepo.FindByID(ctx, trip.GetId())
		require.Error(t, err)
	})

}
