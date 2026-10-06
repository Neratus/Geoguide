package catalog

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/mocks"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func makeCountry(name string) *domain.Country {
	return domain.NewCountryFromDB(
		domain.CountryID(uuid.New()), name,
		551695.0, 2100.0, 146000000,
		"RUB", "Не нужна", "Описание "+name, "Безопасно",
		"Лето", "Русский", "+7", "Православие",
		domain.CityID(uuid.New()), domain.ImageID(uuid.New()),
	)
}

func makeCountryWithID(id domain.CountryID, name string) *domain.Country {
	return domain.NewCountryFromDB(
		id, name,
		551695.0, 2100.0, 146000000,
		"RUB", "Не нужна", "Описание "+name, "Безопасно",
		"Лето", "Русский", "+7", "Православие",
		domain.CityID(uuid.New()), domain.ImageID(uuid.New()),
	)
}

func makeCity(name string, countryID domain.CountryID) *domain.City {
	return domain.NewCityFromDB(
		domain.CityID(uuid.New()), name,
		12000000, true,
		domain.NewCoordinates(55.75, 37.61),
		"Описание "+name, "UTC+3", "Советы",
		countryID, domain.ImageID(uuid.New()),
	)
}

func makeHoliday(name, dateStr string, countryID domain.CountryID) *domain.Holiday {
	date, _ := time.Parse("2006-01-02", dateStr)
	return domain.NewHolidayFromDB(
		domain.HolidayID(uuid.New()), name, date,
		"Описание "+name, "Традиции", "История",
		true, countryID, domain.ImageID(uuid.New()),
	)
}

func makeCityWithID(id domain.CityID, name string, countryID domain.CountryID) *domain.City {
	return domain.NewCityFromDB(
		id, name,
		12000000, true,
		domain.NewCoordinates(55.75, 37.61),
		"Описание "+name, "UTC+3", "Советы",
		countryID, domain.ImageID(uuid.New()),
	)
}

func makeDistrict(name string, cityID domain.CityID) *domain.CityDistrict {
	return domain.NewCityDistrictFromDB(
		domain.CityDistrictID(uuid.New()), name,
		"Описание "+name,
		domain.NewCoordinates(55.75, 37.61),
		cityID, domain.ImageID(uuid.New()),
	)
}

func makeTransportNode(name string, cityID domain.CityID) *domain.TransportNode {
	return domain.NewTransportNodeFromDB(
		domain.NodeID(uuid.New()), name,
		"airport",
		domain.NewCoordinates(55.97, 37.41), "Описание",
		cityID, domain.ImageID(uuid.New()),
	)
}
func TestSearchCountriesUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_FilterByQuery", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		uc := NewSearchCountriesUseCase(repo, logger)

		_ = repo.Save(ctx, makeCountry("France"))
		_ = repo.Save(ctx, makeCountry("Germany"))
		_ = repo.Save(ctx, makeCountry("Finland"))

		req := requests.SearchCountriesRequest{Query: "fr", Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(resp), 1)
		require.Equal(t, "France", resp[0].Name)
	})

	t.Run("Success_EmptyQuery_ReturnsAll", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		uc := NewSearchCountriesUseCase(repo, logger)

		_ = repo.Save(ctx, makeCountry("France"))
		_ = repo.Save(ctx, makeCountry("Germany"))

		req := requests.SearchCountriesRequest{Query: "", Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 2)
	})

	t.Run("Error_RepositoryFailure", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		repo.ForceError.FindAll = errors.New("db down")
		uc := NewSearchCountriesUseCase(repo, logger)

		req := requests.SearchCountriesRequest{Query: "test", Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetCitiesByCountryUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetCitiesWithImageURL", func(t *testing.T) {
		cityRepo := mocks.NewMockCityRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetCitiesByCountryUseCase(cityRepo, staticRepo, logger)

		countryID := domain.CountryID(uuid.New())
		city := makeCity("Paris", countryID)
		_ = cityRepo.Save(ctx, city)

		req := requests.GetCitiesByCountryRequest{CountryID: countryID, Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, "Paris", resp[0].Name)
	})

	t.Run("Error_CitiesNotFound", func(t *testing.T) {
		cityRepo := mocks.NewMockCityRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetCitiesByCountryUseCase(cityRepo, staticRepo, logger)

		req := requests.GetCitiesByCountryRequest{CountryID: domain.CountryID(uuid.New()), Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrCitiesNotFound)
		require.Nil(t, resp)
	})

	t.Run("Error_RepositoryFailure", func(t *testing.T) {
		cityRepo := mocks.NewMockCityRepository()
		cityRepo.ForceError.FindByCountry = errors.New("db error")
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetCitiesByCountryUseCase(cityRepo, staticRepo, logger)

		req := requests.GetCitiesByCountryRequest{CountryID: domain.CountryID(uuid.New()), Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetHolidaysByDateUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetHolidays", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		uc := NewGetHolidaysByDateUseCase(repo, logger)

		countryID := domain.CountryID(uuid.New())
		_ = repo.SaveHoliday(ctx, makeHoliday("New Year", "2026-01-01", countryID))
		_ = repo.SaveHoliday(ctx, makeHoliday("Christmas", "2026-12-25", countryID))

		reqDate, _ := time.Parse("2006-01-02", "2026-01-01")
		req := requests.GetHolidaysByDateRequest{Date: reqDate}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, "New Year", resp[0].Name)
	})

	t.Run("Success_EmptyResult", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		uc := NewGetHolidaysByDateUseCase(repo, logger)

		reqDate, _ := time.Parse("2006-01-02", "2026-06-15")
		req := requests.GetHolidaysByDateRequest{Date: reqDate}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Empty(t, resp)
	})

	t.Run("Error_RepositoryFailure", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		repo.ForceError.FindHolidaysByDate = errors.New("db error")
		uc := NewGetHolidaysByDateUseCase(repo, logger)

		reqDate, _ := time.Parse("2006-01-02", "2026-01-01")
		req := requests.GetHolidaysByDateRequest{Date: reqDate}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetCityUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetCityWithImage", func(t *testing.T) {
		cityRepo := mocks.NewMockCityRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetCityUseCase(cityRepo, staticRepo, logger)

		city := makeCity("Moscow", domain.CountryID(uuid.New()))
		_ = cityRepo.Save(ctx, city)

		req := requests.GetCityRequest{ID: city.GetId()}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "Moscow", resp.Name)
		require.Equal(t, city.GetId(), resp.ID)
	})

	t.Run("Error_CityNotFound", func(t *testing.T) {
		cityRepo := mocks.NewMockCityRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetCityUseCase(cityRepo, staticRepo, logger)

		req := requests.GetCityRequest{ID: domain.CityID(uuid.New())}
		resp, err := uc.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrCityNotFound)
		require.Nil(t, resp)
	})
}

func TestGetPlacesByCategoryUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetPlacesByCategory", func(t *testing.T) {
		placeRepo := mocks.NewMockPlaceRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetPlacesByCategoryUseCase(placeRepo, staticRepo, logger)

		placeBuilder := builders.NewPlaceBuilder()
		place := placeBuilder.Build()
		_ = placeRepo.Save(ctx, place)

		req := requests.GetPlacesByCategoryRequest{Category: place.GetCategory()}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, place.GetName(), resp[0].Name)
		require.Equal(t, place.GetCategory(), resp[0].Category)
	})

	t.Run("Success_EmptyCategory", func(t *testing.T) {
		placeRepo := mocks.NewMockPlaceRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetPlacesByCategoryUseCase(placeRepo, staticRepo, logger)

		req := requests.GetPlacesByCategoryRequest{Category: "NonExistent"}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Empty(t, resp)
	})

}

func TestGetCountriesUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetAllCountries", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()

		cityRepo := mocks.NewMockCityRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetCountriesUseCase(repo, cityRepo, staticRepo, logger)

		_ = repo.Save(ctx, makeCountry("France"))
		_ = repo.Save(ctx, makeCountry("Germany"))

		req := requests.GetCountriesRequest{Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 2)
	})

	t.Run("Error_RepositoryFailure", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		repo.ForceError.FindAll = errors.New("db error")
		cityRepo := mocks.NewMockCityRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		uc := NewGetCountriesUseCase(repo, cityRepo, staticRepo, logger)

		req := requests.GetCountriesRequest{Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetCountryUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetCountryByID", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		cityRepo := mocks.NewMockCityRepository()
		uc := NewGetCountryUseCase(repo, cityRepo, staticRepo, logger)

		country := makeCountry("Japan")
		_ = repo.Save(ctx, country)

		req := requests.GetCountryRequest{ID: country.GetId()}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "Japan", resp.Name)
	})

	t.Run("Error_CountryNotFound", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		cityRepo := mocks.NewMockCityRepository()
		uc := NewGetCountryUseCase(repo, cityRepo, staticRepo, logger)

		req := requests.GetCountryRequest{ID: domain.CountryID(uuid.New())}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetHolidaysByCountryUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetHolidaysByCountry", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		uc := NewGetHolidaysByCountryUseCase(repo, logger)

		countryID := domain.CountryID(uuid.New())
		_ = repo.Save(ctx, makeCountryWithID(countryID, "Россия"))
		_ = repo.SaveHoliday(ctx, makeHoliday("День Победы", "2026-05-09", countryID))

		req := requests.GetHolidaysByCountryRequest{CountryID: countryID}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, "День Победы", resp[0].Name)
	})

	t.Run("Error_RepositoryFailure", func(t *testing.T) {
		repo := mocks.NewMockCountryRepository()
		repo.ForceError.FindHolidaysByCountry = errors.New("db error")
		uc := NewGetHolidaysByCountryUseCase(repo, logger)

		req := requests.GetHolidaysByCountryRequest{CountryID: domain.CountryID(uuid.New())}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetPlaceUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetPlaceByID", func(t *testing.T) {
		placeRepo := mocks.NewMockPlaceRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		user := mocks.NewMockUserRepository()
		uc := NewGetPlaceUseCase(placeRepo, user, staticRepo, logger)

		placeBuilder := builders.NewPlaceBuilder()
		place := placeBuilder.Build()
		_ = placeRepo.Save(ctx, place)

		req := requests.GetPlaceRequest{ID: place.GetId()}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, place.GetName(), resp.Name)
	})

	t.Run("Error_PlaceNotFound", func(t *testing.T) {
		placeRepo := mocks.NewMockPlaceRepository()
		staticRepo := mocks.NewMockStaticPageRepo()
		user := mocks.NewMockUserRepository()
		uc := NewGetPlaceUseCase(placeRepo, user, staticRepo, logger)

		req := requests.GetPlaceRequest{ID: domain.PlaceID(uuid.New())}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestGetPlacesUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_GetPlacesByCity", func(t *testing.T) {
		placeRepo := mocks.NewMockPlaceRepository()
		staticRepo := mocks.NewMockStaticPageRepo()

		placeBuilder := builders.NewPlaceBuilder()
		place := placeBuilder.Build()
		_ = placeRepo.Save(ctx, place)

		uc := NewGetPlacesByCityUseCase(placeRepo, staticRepo, logger)

		req := requests.GetPlacesByCityRequest{CityID: place.GetCityID(), Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(resp), 1)
	})

}

func TestSearchCitiesUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("Success_SearchByName", func(t *testing.T) {
		cityRepo := mocks.NewMockCityRepository()
		uc := NewSearchCitiesUseCase(cityRepo, logger)

		countryID := domain.CountryID(uuid.New())
		_ = cityRepo.Save(ctx, makeCity("Paris", countryID))
		_ = cityRepo.Save(ctx, makeCity("London", countryID))

		req := requests.SearchCitiesRequest{Query: "par", Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(resp), 1)
		require.Equal(t, "Paris", resp[0].Name)
	})

	t.Run("Success_NoResults", func(t *testing.T) {
		cityRepo := mocks.NewMockCityRepository()
		uc := NewSearchCitiesUseCase(cityRepo, logger)

		req := requests.SearchCitiesRequest{Query: "xyz_nonexistent", Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Empty(t, resp)
	})

	t.Run("Error_RepositoryFailure", func(t *testing.T) {
		cityRepo := mocks.NewMockCityRepository()
		cityRepo.ForceError.SearchCities = errors.New("db error")
		uc := NewSearchCitiesUseCase(cityRepo, logger)

		req := requests.SearchCitiesRequest{Query: "test", Limit: 10, Offset: 0}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestMockCityRepository_DistrictsDiagnostic(t *testing.T) {
	repo := mocks.NewMockCityRepository()
	ctx := context.Background()

	cityID := domain.CityID(uuid.New())

	d1 := domain.NewCityDistrictFromDB(
		domain.CityDistrictID(uuid.New()), "Центр", "Описание",
		domain.NewCoordinates(55.75, 37.61),
		cityID, domain.ImageID(uuid.New()),
	)

	err := repo.SaveDistrict(ctx, d1)
	require.NoError(t, err)

	found, err := repo.FindDistrictsByCity(ctx, cityID)
	require.NoError(t, err)
	require.Len(t, found, 1, "Мок должен найти 1 район")

	require.Equal(t, cityID, found[0].GetCityId(), "city_id должен совпадать")
}

func TestMockCityRepository_TransportNodesDiagnostic(t *testing.T) {
	repo := mocks.NewMockCityRepository()
	ctx := context.Background()

	cityID := domain.CityID(uuid.New())

	node := domain.NewTransportNodeFromDB(
		domain.NodeID(uuid.New()), "Шереметьево", "airport",
		domain.NewCoordinates(55.97, 37.41), "Описание",
		cityID, domain.ImageID(uuid.New()),
	)

	err := repo.SaveTransportNode(ctx, node)
	require.NoError(t, err)

	found, err := repo.FindTransportNodesByCity(ctx, cityID)
	require.NoError(t, err)
	require.Len(t, found, 1, "Мок должен найти 1 узел")
}
