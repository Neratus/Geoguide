//go:build postgres
// +build postgres

package place_repo

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
	*PostgresPlaceRepo,
	*country_repo.PostgresCountryRepo,
	*city_repo.PostgresCityRepo,
	*user_repo.PostgresUserRepo,
) {
	t.Helper()
	tx, err := testPool.Begin(context.Background())
	require.NoError(t, err)

	testQueries := postgreSQL.New(tx)
	placeRepo := &PostgresPlaceRepo{Pool: testPool, Queries: testQueries}
	countryRepo := &country_repo.PostgresCountryRepo{Pool: testPool, Queries: testQueries}
	cityRepo := &city_repo.PostgresCityRepo{Pool: testPool, Queries: testQueries}
	userRepo := &user_repo.PostgresUserRepo{Pool: testPool, Queries: testQueries}

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	return placeRepo, countryRepo, cityRepo, userRepo
}

func TestPostgresPlaceRepository(t *testing.T) {

	t.Run("Place_Save_Valid", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()

		err := placeRepo.Save(ctx, place)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, place.GetId())
	})

	t.Run("Place_FindByID_Found", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		originalPlace := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, originalPlace))

		found, err := placeRepo.FindByID(ctx, originalPlace.GetId())

		require.NoError(t, err)
		require.Equal(t, originalPlace.GetId(), found.GetId())
		require.Equal(t, originalPlace.GetName(), found.GetName())
	})

	t.Run("Place_FindByID_NotFound", func(t *testing.T) {
		placeRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		nonExistentID := domain.PlaceID(uuid.New())

		found, err := placeRepo.FindByID(ctx, nonExistentID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Place_FindByID_InvalidID", func(t *testing.T) {
		placeRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		invalidID := domain.PlaceID(uuid.Nil)

		found, err := placeRepo.FindByID(ctx, invalidID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Place_FindByCity_Found", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place1 := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		place2 := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place1))
		require.NoError(t, placeRepo.Save(ctx, place2))

		found, err := placeRepo.FindByCity(ctx, city.GetId())

		require.NoError(t, err)
		require.Len(t, found, 2)
	})

	t.Run("Place_FindByCity_Empty", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		found, err := placeRepo.FindByCity(ctx, city.GetId())

		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("Place_FindByCategory_Found", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).WithCategory("Музей").Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		found, err := placeRepo.FindByCategory(ctx, "Музей")

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(found), 1)
		require.Equal(t, "Музей", found[0].GetCategory())
	})

	t.Run("Place_Update_Success", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		place.SetName("Updated Place Name")
		err := placeRepo.Update(ctx, place)

		require.NoError(t, err)
		updated, err := placeRepo.FindByID(ctx, place.GetId())
		require.NoError(t, err)
		require.Equal(t, "Updated Place Name", updated.GetName())
	})

	t.Run("Place_Delete_Success", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		err := placeRepo.Delete(ctx, place.GetId())

		require.NoError(t, err)
		_, err = placeRepo.FindByID(ctx, place.GetId())
		require.Error(t, err)
	})

	t.Run("Place_UpdateRating_Success", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		err := placeRepo.UpdateRating(ctx, place.GetId(), 4.5, 10)

		require.NoError(t, err)

		updated, err := placeRepo.FindByID(ctx, place.GetId())
		require.NoError(t, err)
		require.Equal(t, place.GetId(), updated.GetId())
	})

	t.Run("Review_Save_Valid", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, userRepo := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		review := builders.NewReviewBuilder().
			WithUserID(user.GetId()).
			WithPlaceID(place.GetId()).
			Build()

		err := placeRepo.SaveReview(ctx, review)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, review.GetId())
	})

	t.Run("Review_FindByID_Found", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, userRepo := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		originalReview := builders.NewReviewBuilder().
			WithUserID(user.GetId()).
			WithPlaceID(place.GetId()).
			WithRating(5).
			Build()
		require.NoError(t, placeRepo.SaveReview(ctx, originalReview))

		found, err := placeRepo.FindReviewByID(ctx, originalReview.GetId())

		require.NoError(t, err)
		require.Equal(t, originalReview.GetId(), found.GetId())
		require.Equal(t, 5, found.GetRating())
	})

	t.Run("Review_FindByID_NotFound", func(t *testing.T) {
		placeRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		nonExistentID := domain.ReviewID(uuid.New())

		found, err := placeRepo.FindReviewByID(ctx, nonExistentID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Review_FindReviewsByPlace_Found", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, userRepo := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		review1 := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place.GetId()).Build()
		review2 := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place.GetId()).Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review1))
		require.NoError(t, placeRepo.SaveReview(ctx, review2))

		require.NoError(t, placeRepo.ModerateReview(ctx, review1.GetId(), true, ""))
		require.NoError(t, placeRepo.ModerateReview(ctx, review2.GetId(), true, ""))

		found, err := placeRepo.FindReviewsByPlace(ctx, place.GetId())

		require.NoError(t, err)
		require.Len(t, found, 2)
	})

	t.Run("Review_FindReviewsByPlace_Empty", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, _ := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		found, err := placeRepo.FindReviewsByPlace(ctx, place.GetId())

		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("Review_FindReviewsByUser_Found", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, userRepo := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		place1 := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		place2 := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place1))
		require.NoError(t, placeRepo.Save(ctx, place2))

		review1 := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place1.GetId()).Build()
		review2 := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place2.GetId()).Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review1))
		require.NoError(t, placeRepo.SaveReview(ctx, review2))

		found, err := placeRepo.FindReviewsByUser(ctx, user.GetId())

		require.NoError(t, err)
		require.Len(t, found, 2)
	})

	t.Run("Review_FindReviewsByUserAndPlace_Found", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, userRepo := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		review := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place.GetId()).Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review))

		user2 := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user2))
		review2 := builders.NewReviewBuilder().WithUserID(user2.GetId()).WithPlaceID(place.GetId()).Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review2))

		require.NoError(t, placeRepo.ModerateReview(ctx, review.GetId(), true, ""))
		require.NoError(t, placeRepo.ModerateReview(ctx, review2.GetId(), true, ""))

		found, err := placeRepo.FindReviewsByUserAndPlace(ctx, user.GetId(), place.GetId())

		require.NoError(t, err)
		require.Len(t, found, 1)
		require.Equal(t, user.GetId(), found[0].GetUserID())
	})

	t.Run("Review_Update_Success", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, userRepo := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		review := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place.GetId()).WithRating(3).Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review))

		review.SetRating(5)
		review.SetComment("Updated comment")
		err := placeRepo.UpdateReview(ctx, review)

		require.NoError(t, err)
		updated, err := placeRepo.FindReviewByID(ctx, review.GetId())
		require.NoError(t, err)
		require.Equal(t, 5, updated.GetRating())
		require.Equal(t, "Updated comment", updated.GetComment())
	})

	t.Run("Review_Delete_Success", func(t *testing.T) {
		placeRepo, countryRepo, cityRepo, userRepo := setupTest(t)
		ctx := context.Background()

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		review := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place.GetId()).Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review))

		err := placeRepo.DeleteReview(ctx, review.GetId())

		require.NoError(t, err)
		_, err = placeRepo.FindReviewByID(ctx, review.GetId())
		require.Error(t, err)
	})
}
