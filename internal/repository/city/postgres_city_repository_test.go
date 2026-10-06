//go:build postgres
// +build postgres

package city_repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	configPkg "github.com/Neratus/geoguide/internal/repository/config"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
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

func setupTest(t *testing.T) (*PostgresCityRepo, *country_repo.PostgresCountryRepo) {
	t.Helper()
	tx, err := testPool.Begin(context.Background())
	require.NoError(t, err)

	testQueries := postgreSQL.New(tx)
	cityRepo := &PostgresCityRepo{Pool: testPool, Queries: testQueries}
	countryRepo := &country_repo.PostgresCountryRepo{Pool: testPool, Queries: testQueries}

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	return cityRepo, countryRepo
}

func TestPostgresCityRepository(t *testing.T) {

	t.Run("City_Save_Valid", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		city := builders.NewCityBuilder().WithName("Paris").WithCountryID(country.GetId()).Build()

		err := cityRepo.Save(ctx, city)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, city.GetId())
	})

	t.Run("City_FindByID_Found", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithName("Rome").WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		found, err := cityRepo.FindByID(ctx, city.GetId())

		require.NoError(t, err)
		require.Equal(t, city.GetId(), found.GetId())
		require.Equal(t, "Rome", found.GetName())
	})

	t.Run("City_FindByID_NotFound", func(t *testing.T) {
		cityRepo, _ := setupTest(t)
		ctx := context.Background()
		nonExistentID := domain.CityID(uuid.New())

		found, err := cityRepo.FindByID(ctx, nonExistentID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("City_FindByID_InvalidID", func(t *testing.T) {
		cityRepo, _ := setupTest(t)
		ctx := context.Background()
		invalidID := domain.CityID(uuid.Nil)

		found, err := cityRepo.FindByID(ctx, invalidID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("City_FindByName_Found", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithName("Tokyo").WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		found, err := cityRepo.FindByName(ctx, "Tokyo")

		require.NoError(t, err)
		require.Equal(t, city.GetId(), found.GetId())
	})

	t.Run("City_FindByName_NotFound", func(t *testing.T) {
		cityRepo, _ := setupTest(t)
		ctx := context.Background()

		found, err := cityRepo.FindByName(ctx, "NonExistentCity")

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("City_FindByCountry_Found", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		c1 := builders.NewCityBuilder().WithName("City1").WithCountryID(country.GetId()).Build()
		c2 := builders.NewCityBuilder().WithName("City2").WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, c1))
		require.NoError(t, cityRepo.Save(ctx, c2))

		found, err := cityRepo.FindByCountry(ctx, country.GetId())

		require.NoError(t, err)
		require.Len(t, found, 2)
	})

	t.Run("City_FindByCountry_Empty", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		found, err := cityRepo.FindByCountry(ctx, country.GetId())

		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("City_SearchCities_Found", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		require.NoError(t, cityRepo.Save(ctx, builders.NewCityBuilder().WithName("Berlin").WithCountryID(country.GetId()).Build()))

		found, err := cityRepo.SearchCities(ctx, "Ber", 10, 0)

		require.NoError(t, err)
		require.Len(t, found, 1)
		require.Equal(t, "Berlin", found[0].GetName())
	})

	t.Run("City_SearchCities_Empty", func(t *testing.T) {
		cityRepo, _ := setupTest(t)
		ctx := context.Background()

		found, err := cityRepo.SearchCities(ctx, "XYZ", 10, 0)

		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("City_Update_Success", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithName("Madrid").WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		city.SetName("Madrid Updated")
		err := cityRepo.Update(ctx, city)

		require.NoError(t, err)
		updated, err := cityRepo.FindByID(ctx, city.GetId())
		require.NoError(t, err)
		require.Equal(t, "Madrid Updated", updated.GetName())
	})

	t.Run("City_Update_PartialSuccess", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithName("Lisbon").WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		originalPop := city.GetPopulation()

		city.SetDescription("New description only")
		err := cityRepo.Update(ctx, city)

		require.NoError(t, err)
		updated, err := cityRepo.FindByID(ctx, city.GetId())
		require.NoError(t, err)
		require.Equal(t, "New description only", updated.GetDescription())
		require.Equal(t, originalPop, updated.GetPopulation())
	})

	t.Run("City_Delete_Success", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithName("ToDelete").WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		err := cityRepo.Delete(ctx, city.GetId())

		require.NoError(t, err)
		_, err = cityRepo.FindByID(ctx, city.GetId())
		require.Error(t, err)
	})

	t.Run("District_Save_Valid", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		district := builders.NewCityDistrictBuilder().WithCityID(city.GetId()).Build()

		err := cityRepo.SaveDistrict(ctx, district)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, district.GetId())
	})

	t.Run("District_FindByCity_Found", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		d1 := builders.NewCityDistrictBuilder().WithCityID(city.GetId()).Build()
		d2 := builders.NewCityDistrictBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, cityRepo.SaveDistrict(ctx, d1))
		require.NoError(t, cityRepo.SaveDistrict(ctx, d2))

		found, err := cityRepo.FindDistrictsByCity(ctx, city.GetId())

		require.NoError(t, err)
		require.Len(t, found, 2)
	})

	t.Run("District_FindByCity_Empty", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		found, err := cityRepo.FindDistrictsByCity(ctx, city.GetId())

		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("District_Update_Success", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		district := builders.NewCityDistrictBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, cityRepo.SaveDistrict(ctx, district))

		district.SetName("Updated District")
		err := cityRepo.UpdateDistrict(ctx, district)

		require.NoError(t, err)
		found, err := cityRepo.FindDistrictsByCity(ctx, city.GetId())
		require.NoError(t, err)
		require.Equal(t, "Updated District", found[0].GetName())
	})

	t.Run("District_Delete_Success", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		district := builders.NewCityDistrictBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, cityRepo.SaveDistrict(ctx, district))

		err := cityRepo.DeleteDistrict(ctx, district.GetId())

		require.NoError(t, err)
		found, err := cityRepo.FindDistrictsByCity(ctx, city.GetId())
		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("TransportNode_Save_Valid", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		node := builders.NewTransportNodeBuilder().WithCityID(city.GetId()).Build()

		err := cityRepo.SaveTransportNode(ctx, node)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, node.GetId())
	})

	t.Run("TransportNode_FindByID_Found", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		node := domain.NewTransportNodeFromDB(uuid.New(), "Test Airport", "airport", domain.NewCoordinates(0, 0), "Location", city.GetId(), builders.DefaultImageID)
		require.NoError(t, cityRepo.SaveTransportNode(ctx, node))

		found, err := cityRepo.FindTransportNodeByID(ctx, node.GetId())

		require.NoError(t, err)
		require.Equal(t, node.GetId(), found.GetId())
	})

	t.Run("TransportNode_FindByID_NotFound", func(t *testing.T) {
		cityRepo, _ := setupTest(t)
		ctx := context.Background()
		nonExistentID := domain.NodeID(uuid.New())

		found, err := cityRepo.FindTransportNodeByID(ctx, nonExistentID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("TransportNode_FindByID_InvalidID", func(t *testing.T) {
		cityRepo, _ := setupTest(t)
		ctx := context.Background()
		invalidID := domain.NodeID(uuid.Nil)

		found, err := cityRepo.FindTransportNodeByID(ctx, invalidID)

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("TransportNode_FindByCity_Found", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		n1 := domain.NewTransportNodeFromDB(uuid.New(), "Node1", "type", domain.NewCoordinates(0, 0), "Loc", city.GetId(), builders.DefaultImageID)
		n2 := domain.NewTransportNodeFromDB(uuid.New(), "Node2", "type", domain.NewCoordinates(0, 0), "Loc", city.GetId(), builders.DefaultImageID)
		require.NoError(t, cityRepo.SaveTransportNode(ctx, n1))
		require.NoError(t, cityRepo.SaveTransportNode(ctx, n2))

		found, err := cityRepo.FindTransportNodesByCity(ctx, city.GetId())

		require.NoError(t, err)
		require.Len(t, found, 2)
	})

	t.Run("TransportNode_FindByCity_Empty", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		found, err := cityRepo.FindTransportNodesByCity(ctx, city.GetId())

		require.NoError(t, err)
		require.Empty(t, found)
	})

	t.Run("TransportNode_Update_Success", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		node := domain.NewTransportNodeFromDB(uuid.New(), "Old Name", "type", domain.NewCoordinates(0, 0), "Loc", city.GetId(), builders.DefaultImageID)
		require.NoError(t, cityRepo.SaveTransportNode(ctx, node))

		node.SetName("New Name")
		err := cityRepo.UpdateTransportNode(ctx, node)

		require.NoError(t, err)
		updated, err := cityRepo.FindTransportNodeByID(ctx, node.GetId())
		require.NoError(t, err)
		require.Equal(t, "New Name", updated.GetName())
	})

	t.Run("TransportNode_Delete_Success", func(t *testing.T) {
		cityRepo, countryRepo := setupTest(t)
		ctx := context.Background()
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		node := domain.NewTransportNodeFromDB(uuid.New(), "ToDelete", "type", domain.NewCoordinates(0, 0), "Loc", city.GetId(), builders.DefaultImageID)
		require.NoError(t, cityRepo.SaveTransportNode(ctx, node))

		err := cityRepo.DeleteTransportNode(ctx, node.GetId())

		require.NoError(t, err)
		_, err = cityRepo.FindTransportNodeByID(ctx, node.GetId())
		require.Error(t, err)
	})
}
