//go:build postgres
// +build postgres

package catalog

import (
	"context"
	"fmt"
	"io"
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
		return nil, nil, fmt.Errorf("parse config: %w", err)
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
	testSlug := "test-catalog-image-" + uuid.NewString()[:8]
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
	*country_repo.PostgresCountryRepo,
	*city_repo.PostgresCityRepo,
	*place_repo.PostgresPlaceRepo,
	*user_repo.PostgresUserRepo,
) {
	t.Helper()
	countryRepo := &country_repo.PostgresCountryRepo{Pool: testPool, Queries: testQueries}
	cityRepo := &city_repo.PostgresCityRepo{Pool: testPool, Queries: testQueries}
	placeRepo := &place_repo.PostgresPlaceRepo{Pool: testPool, Queries: testQueries}
	userRepo := &user_repo.PostgresUserRepo{Pool: testPool, Queries: testQueries}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tables := []string{"Review", "Place", "CityDistrict", "TransportNode", "City", "Holiday", "Country", "User"}
		for _, table := range tables {
			_, _ = testPool.Exec(ctx, "TRUNCATE TABLE \""+table+"\" RESTART IDENTITY CASCADE;")
		}
	})

	return countryRepo, cityRepo, placeRepo, userRepo
}

type stubStaticRepo struct{}

func (s *stubStaticRepo) Save(ctx context.Context, page *domain.StaticPage, reader io.Reader) error {
	return nil
}
func (s *stubStaticRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.StaticPage, error) {
	return nil, nil
}
func (s *stubStaticRepo) Update(ctx context.Context, page *domain.StaticPage, reader io.Reader) error {
	return nil
}
func (s *stubStaticRepo) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	return nil
}
func (s *stubStaticRepo) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (s *stubStaticRepo) DeleteFile(ctx context.Context, fileID string) error { return nil }
func (s *stubStaticRepo) FindBySlug(ctx context.Context, slug string) (*domain.StaticPage, error) {
	return nil, nil
}
func (s *stubStaticRepo) GetFile(ctx context.Context, fileID domain.ImageID) (io.ReadCloser, error) {
	return nil, nil
}
func (s *stubStaticRepo) GetFileURL(ctx context.Context, fileID string) (string, error) {
	return "https://cdn.example.com/" + fileID, nil
}

func TestSearchCountriesUseCase_Integration(t *testing.T) {
	t.Run("Success_FilterByQuery", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, _, _, _ := setupTest(t)
		uc := NewSearchCountriesUseCase(countryRepo, newTestLogger())

		require.NoError(t, countryRepo.Save(ctx, builders.NewCountryBuilder().WithName("France").Build()))
		require.NoError(t, countryRepo.Save(ctx, builders.NewCountryBuilder().WithName("Germany").Build()))
		require.NoError(t, countryRepo.Save(ctx, builders.NewCountryBuilder().WithName("Finland").Build()))

		req := requests.SearchCountriesRequest{Query: "fr", Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(resp), 1)
		require.Equal(t, "France", resp[0].Name)
	})

	t.Run("Success_EmptyQuery_ReturnsAll", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, _, _, _ := setupTest(t)
		uc := NewSearchCountriesUseCase(countryRepo, newTestLogger())

		require.NoError(t, countryRepo.Save(ctx, builders.NewCountryBuilder().WithName("Italy").Build()))
		require.NoError(t, countryRepo.Save(ctx, builders.NewCountryBuilder().WithName("Spain").Build()))

		req := requests.SearchCountriesRequest{Query: "", Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 2)
	})
}

func TestGetCitiesByCountryUseCase_Integration(t *testing.T) {
	t.Run("Success_GetCitiesWithImageURL", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, _, _ := setupTest(t)
		uc := NewGetCitiesByCountryUseCase(cityRepo, &stubStaticRepo{}, newTestLogger())

		country := builders.NewCountryBuilder().WithName("Japan").Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		city := builders.NewCityBuilder().WithName("Tokyo").WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		req := requests.GetCitiesByCountryRequest{CountryID: country.GetId(), Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, "Tokyo", resp[0].Name)
	})

	t.Run("Error_CitiesNotFound", func(t *testing.T) {
		ctx := context.Background()
		_, cityRepo, _, _ := setupTest(t)
		uc := NewGetCitiesByCountryUseCase(cityRepo, &stubStaticRepo{}, newTestLogger())

		req := requests.GetCitiesByCountryRequest{CountryID: domain.CountryID(uuid.New()), Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrCitiesNotFound)
		require.Nil(t, resp)
	})
}

func TestGetHolidaysByDateUseCase_Integration(t *testing.T) {
	t.Run("Success_GetHolidays", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, _, _, _ := setupTest(t)
		uc := NewGetHolidaysByDateUseCase(countryRepo, newTestLogger())

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		date1, _ := time.Parse("2006-01-02", "2026-01-01")
		date2, _ := time.Parse("2006-01-02", "2026-12-25")
		h1 := domain.NewHolidayFromDB(uuid.New(), "New Year", date1, "Desc", "Trad", "Hist", true, country.GetId(), builders.DefaultImageID)
		h2 := domain.NewHolidayFromDB(uuid.New(), "Christmas", date2, "Desc", "Trad", "Hist", true, country.GetId(), builders.DefaultImageID)
		require.NoError(t, countryRepo.SaveHoliday(ctx, h1))
		require.NoError(t, countryRepo.SaveHoliday(ctx, h2))

		reqDate, _ := time.Parse("2006-01-02", "2026-01-01")
		resp, err := uc.Execute(ctx, requests.GetHolidaysByDateRequest{Date: reqDate})

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, "New Year", resp[0].Name)
	})

	t.Run("Success_EmptyResult", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, _, _, _ := setupTest(t)
		uc := NewGetHolidaysByDateUseCase(countryRepo, newTestLogger())

		reqDate, _ := time.Parse("2006-01-02", "2099-06-15")
		resp, err := uc.Execute(ctx, requests.GetHolidaysByDateRequest{Date: reqDate})

		require.NoError(t, err)
		require.Empty(t, resp)
	})
}

func TestGetCityUseCase_Integration(t *testing.T) {
	t.Run("Success_GetCityWithImage", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, _, _ := setupTest(t)
		uc := NewGetCityUseCase(cityRepo, &stubStaticRepo{}, newTestLogger())

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithName("Moscow").WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		resp, err := uc.Execute(ctx, requests.GetCityRequest{ID: city.GetId()})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "Moscow", resp.Name)
		require.Equal(t, city.GetId(), resp.ID)
	})

	t.Run("Error_CityNotFound", func(t *testing.T) {
		ctx := context.Background()
		_, cityRepo, _, _ := setupTest(t)
		uc := NewGetCityUseCase(cityRepo, &stubStaticRepo{}, newTestLogger())

		resp, err := uc.Execute(ctx, requests.GetCityRequest{ID: domain.CityID(uuid.New())})
		require.ErrorIs(t, err, usecase_errors.ErrCityNotFound)
		require.Nil(t, resp)
	})
}

func TestGetPlacesByCategoryUseCase_Integration(t *testing.T) {
	t.Run("Success_GetPlacesByCategory", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, placeRepo, _ := setupTest(t)
		uc := NewGetPlacesByCategoryUseCase(placeRepo, &stubStaticRepo{}, newTestLogger())

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).WithCategory("Музей").Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		resp, err := uc.Execute(ctx, requests.GetPlacesByCategoryRequest{Category: "Музей"})

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, "Музей", resp[0].Category)
	})

	t.Run("Success_EmptyCategory", func(t *testing.T) {
		ctx := context.Background()
		_, _, placeRepo, _ := setupTest(t)
		uc := NewGetPlacesByCategoryUseCase(placeRepo, &stubStaticRepo{}, newTestLogger())

		resp, err := uc.Execute(ctx, requests.GetPlacesByCategoryRequest{Category: "NonExistent"})
		require.NoError(t, err)
		require.Empty(t, resp)
	})
}

func TestGetCountriesUseCase_Integration(t *testing.T) {
	t.Run("Success_GetAllCountries", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, _, _ := setupTest(t)
		uc := NewGetCountriesUseCase(countryRepo, cityRepo, &stubStaticRepo{}, newTestLogger())

		require.NoError(t, countryRepo.Save(ctx, builders.NewCountryBuilder().WithName("France").Build()))
		require.NoError(t, countryRepo.Save(ctx, builders.NewCountryBuilder().WithName("Germany").Build()))

		resp, err := uc.Execute(ctx, requests.GetCountriesRequest{Limit: 10, Offset: 0})

		require.NoError(t, err)
		require.Len(t, resp, 2)
	})
}

func TestGetCountryUseCase_Integration(t *testing.T) {
	t.Run("Success_GetCountryByID", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, _, _ := setupTest(t)
		uc := NewGetCountryUseCase(countryRepo, cityRepo, &stubStaticRepo{}, newTestLogger())

		country := builders.NewCountryBuilder().WithName("Japan").Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		resp, err := uc.Execute(ctx, requests.GetCountryRequest{ID: country.GetId()})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "Japan", resp.Name)
	})

	t.Run("Error_CountryNotFound", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, _, _ := setupTest(t)
		uc := NewGetCountryUseCase(countryRepo, cityRepo, &stubStaticRepo{}, newTestLogger())

		resp, err := uc.Execute(ctx, requests.GetCountryRequest{ID: domain.CountryID(uuid.New())})
		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetHolidaysByCountryUseCase_Integration(t *testing.T) {
	t.Run("Success_GetHolidaysByCountry", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, _, _, _ := setupTest(t)
		uc := NewGetHolidaysByCountryUseCase(countryRepo, newTestLogger())

		country := builders.NewCountryBuilder().WithName("Россия").Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		date, _ := time.Parse("2006-01-02", "2026-05-09")
		h := domain.NewHolidayFromDB(uuid.New(), "День Победы", date, "Desc", "Trad", "Hist", true, country.GetId(), builders.DefaultImageID)
		require.NoError(t, countryRepo.SaveHoliday(ctx, h))

		resp, err := uc.Execute(ctx, requests.GetHolidaysByCountryRequest{CountryID: country.GetId()})

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, "День Победы", resp[0].Name)
	})
}

func TestGetPlaceUseCase_Integration(t *testing.T) {
	t.Run("Success_GetPlaceByID", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, placeRepo, userRepo := setupTest(t)
		uc := NewGetPlaceUseCase(placeRepo, userRepo, &stubStaticRepo{}, newTestLogger())

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		resp, err := uc.Execute(ctx, requests.GetPlaceRequest{ID: place.GetId()})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, place.GetName(), resp.Name)
	})

	t.Run("Error_PlaceNotFound", func(t *testing.T) {
		ctx := context.Background()
		_, _, placeRepo, userRepo := setupTest(t)
		uc := NewGetPlaceUseCase(placeRepo, userRepo, &stubStaticRepo{}, newTestLogger())

		resp, err := uc.Execute(ctx, requests.GetPlaceRequest{ID: domain.PlaceID(uuid.New())})
		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetPlacesByCityUseCase_Integration(t *testing.T) {
	t.Run("Success_GetPlacesByCity", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, placeRepo, _ := setupTest(t)
		uc := NewGetPlacesByCityUseCase(placeRepo, &stubStaticRepo{}, newTestLogger())

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		resp, err := uc.Execute(ctx, requests.GetPlacesByCityRequest{CityID: city.GetId(), Limit: 10, Offset: 0})

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(resp), 1)
	})
}

func TestSearchCitiesUseCase_Integration(t *testing.T) {
	t.Run("Success_SearchByName", func(t *testing.T) {
		ctx := context.Background()
		countryRepo, cityRepo, _, _ := setupTest(t)
		uc := NewSearchCitiesUseCase(cityRepo, newTestLogger())

		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		require.NoError(t, cityRepo.Save(ctx, builders.NewCityBuilder().WithName("Paris").WithCountryID(country.GetId()).Build()))
		require.NoError(t, cityRepo.Save(ctx, builders.NewCityBuilder().WithName("London").WithCountryID(country.GetId()).Build()))

		resp, err := uc.Execute(ctx, requests.SearchCitiesRequest{Query: "par", Limit: 10, Offset: 0})

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(resp), 1)
		require.Equal(t, "Paris", resp[0].Name)
	})

	t.Run("Success_NoResults", func(t *testing.T) {
		ctx := context.Background()
		_, cityRepo, _, _ := setupTest(t)
		uc := NewSearchCitiesUseCase(cityRepo, newTestLogger())

		resp, err := uc.Execute(ctx, requests.SearchCitiesRequest{Query: "xyz_nonexistent", Limit: 10, Offset: 0})
		require.NoError(t, err)
		require.Empty(t, resp)
	})
}
