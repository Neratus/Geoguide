//go:build postgres
// +build postgres

package country_repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	configPkg "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
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

func setupTest(t *testing.T) *PostgresCountryRepo {
	t.Helper()
	tx, err := testPool.Begin(context.Background())
	require.NoError(t, err)

	queries := postgreSQL.New(tx)
	repo := &PostgresCountryRepo{
		Pool:    testPool,
		Queries: queries,
	}

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	return repo
}

func TestPostgresCountryRepository(t *testing.T) {

	t.Run("Country_Save_Valid", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().WithName("France").Build()

		err := repo.Save(ctx, country)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, country.GetId())
	})

	t.Run("Country_FindByID_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().WithName("Italy").Build()
		require.NoError(t, repo.Save(ctx, country))

		found, err := repo.FindByID(ctx, country.GetId())

		require.NoError(t, err)
		require.Equal(t, country.GetId(), found.GetId())
		require.Equal(t, "Italy", found.GetName())
	})

	t.Run("Country_FindByID_NotFound", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		nonExistentID := domain.CountryID(uuid.New())

		found, err := repo.FindByID(ctx, nonExistentID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Country_FindByID_InvalidID", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		invalidID := domain.CountryID(uuid.Nil)

		found, err := repo.FindByID(ctx, invalidID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Country_FindByName_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().WithName("Japan").Build()
		require.NoError(t, repo.Save(ctx, country))

		found, err := repo.FindByName(ctx, "Japan")

		require.NoError(t, err)
		require.Equal(t, country.GetId(), found.GetId())
	})

	t.Run("Country_FindByName_NotFound", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()

		found, err := repo.FindByName(ctx, "NonExistentCountry")

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Country_FindAll", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		require.NoError(t, repo.Save(ctx, builders.NewCountryBuilder().WithName("USA").Build()))
		require.NoError(t, repo.Save(ctx, builders.NewCountryBuilder().WithName("Canada").Build()))

		all, err := repo.FindAll(ctx)

		require.NoError(t, err)
		require.Len(t, all, 2)
	})

	t.Run("Country_Update_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().WithName("Brazil").Build()
		require.NoError(t, repo.Save(ctx, country))

		country.SetName("Federative Republic of Brazil")
		err := repo.Update(ctx, country)

		require.NoError(t, err)
		updated, err := repo.FindByID(ctx, country.GetId())
		require.NoError(t, err)
		require.Equal(t, "Federative Republic of Brazil", updated.GetName())
	})

	t.Run("Country_Update_PartialSuccess", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().WithName("Egypt").Build()
		require.NoError(t, repo.Save(ctx, country))

		originalName := country.GetName()

		country.SetDescription("Updated description only")
		err := repo.Update(ctx, country)
		require.NoError(t, err)

		updated, err := repo.FindByID(ctx, country.GetId())
		require.NoError(t, err)

		require.Equal(t, "Updated description only", updated.GetDescription())
		require.Equal(t, originalName, updated.GetName())
	})

	t.Run("Country_Delete_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().WithName("ToDelete").Build()
		require.NoError(t, repo.Save(ctx, country))

		err := repo.Delete(ctx, country.GetId())

		require.NoError(t, err)
		_, err = repo.FindByID(ctx, country.GetId())
		require.Error(t, err)
	})

	// --- Holiday Scenarios ---

	t.Run("Holiday_Save_Valid", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, repo.Save(ctx, country))

		date := time.Date(2024, 7, 14, 0, 0, 0, 0, time.UTC)
		holiday := domain.NewHolidayFromDB(uuid.New(), "Bastille Day", date, "Desc", "Tradition", "History", true, country.GetId(), builders.DefaultImageID)

		err := repo.SaveHoliday(ctx, holiday)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, holiday.GetId())
	})

	t.Run("Holiday_FindByID_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, repo.Save(ctx, country))

		date := time.Date(2024, 7, 14, 0, 0, 0, 0, time.UTC)
		holiday := domain.NewHolidayFromDB(uuid.New(), "Test Holiday", date, "Desc", "Tradition", "History", true, country.GetId(), builders.DefaultImageID)
		require.NoError(t, repo.SaveHoliday(ctx, holiday))

		found, err := repo.FindHolidayByID(ctx, holiday.GetId())

		require.NoError(t, err)
		require.Equal(t, holiday.GetId(), found.GetId())
	})

	t.Run("Holiday_FindByID_NotFound", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		nonExistentID := domain.HolidayID(uuid.New())

		found, err := repo.FindHolidayByID(ctx, nonExistentID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Holiday_FindByID_InvalidID", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		invalidID := domain.HolidayID(uuid.Nil)

		found, err := repo.FindHolidayByID(ctx, invalidID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Holiday_FindHolidaysByCountry_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, repo.Save(ctx, country))

		date := time.Date(2024, 7, 14, 0, 0, 0, 0, time.UTC)
		h1 := domain.NewHolidayFromDB(uuid.New(), "H1", date, "D", "T", "H", true, country.GetId(), builders.DefaultImageID)
		h2 := domain.NewHolidayFromDB(uuid.New(), "H2", date, "D", "T", "H", true, country.GetId(), builders.DefaultImageID)
		require.NoError(t, repo.SaveHoliday(ctx, h1))
		require.NoError(t, repo.SaveHoliday(ctx, h2))

		found, err := repo.FindHolidaysByCountry(ctx, country.GetId())

		require.NoError(t, err)
		require.Len(t, found, 2)
	})

	t.Run("Holiday_FindHolidaysByCountry_Empty", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, repo.Save(ctx, country))

		found, err := repo.FindHolidaysByCountry(ctx, country.GetId())

		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("Holiday_FindHolidaysByDate_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, repo.Save(ctx, country))

		date := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
		h := domain.NewHolidayFromDB(uuid.New(), "Christmas", date, "D", "T", "H", true, country.GetId(), builders.DefaultImageID)
		require.NoError(t, repo.SaveHoliday(ctx, h))

		found, err := repo.FindHolidaysByDate(ctx, "2024-12-25")

		require.NoError(t, err)
		require.Len(t, found, 1)
		require.Equal(t, "Christmas", found[0].GetName())
	})

	t.Run("Holiday_FindHolidaysByDate_NotFound", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()

		found, err := repo.FindHolidaysByDate(ctx, "2099-01-01")

		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("Holiday_Update_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, repo.Save(ctx, country))

		date := time.Date(2024, 7, 14, 0, 0, 0, 0, time.UTC)
		holiday := domain.NewHolidayFromDB(uuid.New(), "Old Name", date, "Desc", "T", "H", true, country.GetId(), builders.DefaultImageID)
		require.NoError(t, repo.SaveHoliday(ctx, holiday))

		holiday.SetName("New Name")
		err := repo.UpdateHoliday(ctx, holiday)

		require.NoError(t, err)
		updated, err := repo.FindHolidayByID(ctx, holiday.GetId())
		require.NoError(t, err)
		require.Equal(t, "New Name", updated.GetName())
	})

	t.Run("Holiday_Delete_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, repo.Save(ctx, country))

		date := time.Date(2024, 7, 14, 0, 0, 0, 0, time.UTC)
		holiday := domain.NewHolidayFromDB(uuid.New(), "ToDelete", date, "Desc", "T", "H", true, country.GetId(), builders.DefaultImageID)
		require.NoError(t, repo.SaveHoliday(ctx, holiday))

		err := repo.DeleteHoliday(ctx, holiday.GetId())

		require.NoError(t, err)
		_, err = repo.FindHolidayByID(ctx, holiday.GetId())
		require.Error(t, err)
	})
}
