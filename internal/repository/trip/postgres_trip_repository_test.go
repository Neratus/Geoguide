//go:build postgres
// +build postgres

package trip_repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	city_repo "github.com/Neratus/geoguide/internal/repository/city"
	configPkg "github.com/Neratus/geoguide/internal/repository/config"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
	place_repo "github.com/Neratus/geoguide/internal/repository/place"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

var testPool *pgxpool.Pool
var testQueries *postgreSQL.Queries

func TestMain(m *testing.M) {
	cfg, err := configPkg.LoadTest()
	if err != nil {
		log.Fatalf("Failed to load test config: %v", err)
	}

	connStr := cfg.PostgresConnString()
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	if err := goose.Up(sqlDB, "../../../migrations"); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}
	sqlDB.Close()

	testPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	testQueries = postgreSQL.New(testPool)

	var testImageID domain.ImageID
	testSlug := fmt.Sprintf("test-image-slug-%s", uuid.New().String()[:8])

	err = testPool.QueryRow(context.Background(),
		`INSERT INTO "StaticPage" (slug, title) VALUES ($1, 'Test Image') RETURNING id`, testSlug).
		Scan(&testImageID)
	if err != nil {
		log.Fatalf("Failed to create test StaticPage: %v", err)
	}
	builders.DefaultImageID = testImageID

	domain.SetConfig(domain.GetConfig())

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func setupTest(t *testing.T) (
	*PostgresTripRepo,
	*user_repo.PostgresUserRepo,
	*country_repo.PostgresCountryRepo,
	*city_repo.PostgresCityRepo,
	*place_repo.PostgresPlaceRepo,
) {
	t.Helper()
	tx, err := testPool.Begin(context.Background())
	require.NoError(t, err)

	queries := postgreSQL.New(tx)
	tripRepo := &PostgresTripRepo{Pool: testPool, Queries: queries}
	userRepo := &user_repo.PostgresUserRepo{Pool: testPool, Queries: queries}
	countryRepo := &country_repo.PostgresCountryRepo{Pool: testPool, Queries: queries}
	cityRepo := &city_repo.PostgresCityRepo{Pool: testPool, Queries: queries}
	placeRepo := &place_repo.PostgresPlaceRepo{Pool: testPool, Queries: queries}

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	return tripRepo, userRepo, countryRepo, cityRepo, placeRepo
}

func TestPostgresTripRepository(t *testing.T) {

	t.Run("Trip_Save_Valid", func(t *testing.T) {
		tripRepo, userRepo, _, _, _ := setupTest(t)
		ctx := context.Background()

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trip := builders.NewTripBuilder().
			WithUserID(user.GetId()).
			WithTitle("Summer Vacation").
			Build()

		err := tripRepo.Save(ctx, trip)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, trip.GetId())
	})

	t.Run("Trip_FindByID_Found", func(t *testing.T) {
		tripRepo, userRepo, _, _, _ := setupTest(t)
		ctx := context.Background()

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		originalTrip := builders.NewTripBuilder().
			WithUserID(user.GetId()).
			WithTitle("Found Trip").
			Build()
		require.NoError(t, tripRepo.Save(ctx, originalTrip))

		found, err := tripRepo.FindByID(ctx, originalTrip.GetId())

		require.NoError(t, err)
		require.Equal(t, originalTrip.GetId(), found.GetId())
		require.Equal(t, "Found Trip", found.GetTitle())
	})

	t.Run("Trip_FindByID_NotFound", func(t *testing.T) {
		tripRepo, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		nonExistentID := domain.TripID(uuid.New())

		found, err := tripRepo.FindByID(ctx, nonExistentID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Trip_FindByID_InvalidID", func(t *testing.T) {
		tripRepo, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		invalidID := domain.TripID(uuid.Nil)

		found, err := tripRepo.FindByID(ctx, invalidID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Trip_FindByUser_Found", func(t *testing.T) {
		tripRepo, userRepo, _, _, _ := setupTest(t)
		ctx := context.Background()

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trip1 := builders.NewTripBuilder().WithUserID(user.GetId()).WithTitle("Trip 1").Build()
		trip2 := builders.NewTripBuilder().WithUserID(user.GetId()).WithTitle("Trip 2").Build()
		require.NoError(t, tripRepo.Save(ctx, trip1))
		require.NoError(t, tripRepo.Save(ctx, trip2))

		trips, err := tripRepo.FindByUser(ctx, user.GetId())

		require.NoError(t, err)
		require.Len(t, trips, 2)
	})

	t.Run("Trip_FindByUser_Empty", func(t *testing.T) {
		tripRepo, userRepo, _, _, _ := setupTest(t)
		ctx := context.Background()

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trips, err := tripRepo.FindByUser(ctx, user.GetId())

		require.NoError(t, err)
		require.Empty(t, trips)
	})

	t.Run("Trip_Update_Success", func(t *testing.T) {
		tripRepo, userRepo, _, _, _ := setupTest(t)
		ctx := context.Background()

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trip := builders.NewTripBuilder().WithUserID(user.GetId()).WithTitle("Old Title").Build()
		require.NoError(t, tripRepo.Save(ctx, trip))

		trip.SetTitle("New Title")
		err := tripRepo.Update(ctx, trip)

		require.NoError(t, err)
		updated, err := tripRepo.FindByID(ctx, trip.GetId())
		require.NoError(t, err)
		require.Equal(t, "New Title", updated.GetTitle())
	})

	t.Run("Trip_Delete_Success", func(t *testing.T) {
		tripRepo, userRepo, _, _, _ := setupTest(t)
		ctx := context.Background()

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		trip := builders.NewTripBuilder().WithUserID(user.GetId()).Build()
		require.NoError(t, tripRepo.Save(ctx, trip))

		err := tripRepo.Delete(ctx, trip.GetId())

		require.NoError(t, err)
		_, err = tripRepo.FindByID(ctx, trip.GetId())
		require.Error(t, err)
	})

	t.Run("TripPlace_AddPlace_Success", func(t *testing.T) {
		tripRepo, userRepo, countryRepo, cityRepo, placeRepo := setupTest(t)
		ctx := context.Background()

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

		err := tripRepo.AddPlace(ctx, trip.GetId(), place.GetId(), 2, nil, 60, "nice view")

		require.NoError(t, err)
		places, err := tripRepo.GetPlaces(ctx, trip.GetId())
		require.NoError(t, err)
		require.Len(t, places, 1)
		require.Equal(t, place.GetId(), places[0].GetPlaceID())
	})

	t.Run("TripPlace_AddPlace_TripNotFound", func(t *testing.T) {
		tripRepo, _, countryRepo, cityRepo, placeRepo := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		nonExistentTripID := domain.TripID(uuid.New())
		err := tripRepo.AddPlace(ctx, nonExistentTripID, place.GetId(), 1, nil, 60, "")

		require.Error(t, err)
	})

	t.Run("TripPlace_GetPlaces_Success", func(t *testing.T) {
		tripRepo, userRepo, countryRepo, cityRepo, placeRepo := setupTest(t)
		ctx := context.Background()

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		trip := builders.NewTripBuilder().WithUserID(user.GetId()).Build()
		require.NoError(t, tripRepo.Save(ctx, trip))

		place1 := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		place2 := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place1))
		require.NoError(t, placeRepo.Save(ctx, place2))

		require.NoError(t, tripRepo.AddPlace(ctx, trip.GetId(), place1.GetId(), 1, nil, 60, ""))
		require.NoError(t, tripRepo.AddPlace(ctx, trip.GetId(), place2.GetId(), 2, nil, 90, ""))

		places, err := tripRepo.GetPlaces(ctx, trip.GetId())

		require.NoError(t, err)
		require.Len(t, places, 2)
	})

	t.Run("TripPlace_GetPlaces_TripNotFound", func(t *testing.T) {
		tripRepo, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		nonExistentTripID := domain.TripID(uuid.New())

		places, err := tripRepo.GetPlaces(ctx, nonExistentTripID)

		require.NoError(t, err)
		require.Empty(t, places)
	})

	t.Run("TripPlace_RemovePlace_Success", func(t *testing.T) {
		tripRepo, userRepo, countryRepo, cityRepo, placeRepo := setupTest(t)
		ctx := context.Background()

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

		require.NoError(t, tripRepo.AddPlace(ctx, trip.GetId(), place.GetId(), 1, nil, 60, ""))

		err := tripRepo.RemovePlace(ctx, trip.GetId(), place.GetId())

		require.NoError(t, err)
		places, err := tripRepo.GetPlaces(ctx, trip.GetId())
		require.NoError(t, err)
		require.Empty(t, places)
	})

	t.Run("TripPlace_UpdateTripPlace_Success", func(t *testing.T) {
		tripRepo, userRepo, countryRepo, cityRepo, placeRepo := setupTest(t)
		ctx := context.Background()

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

		require.NoError(t, tripRepo.AddPlace(ctx, trip.GetId(), place.GetId(), 1, nil, 60, "original notes"))

		places, err := tripRepo.GetPlaces(ctx, trip.GetId())
		require.NoError(t, err)
		require.Len(t, places, 1)

		tripPlace := places[0]
		tripPlace.SetNotes("updated notes")

		err = tripRepo.UpdateTripPlace(ctx, tripPlace)

		require.NoError(t, err)
		updatedPlaces, err := tripRepo.GetPlaces(ctx, trip.GetId())
		require.NoError(t, err)
		require.Len(t, updatedPlaces, 1)
		require.Equal(t, "updated notes", updatedPlaces[0].GetNotes())
	})
}
