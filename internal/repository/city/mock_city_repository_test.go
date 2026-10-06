package city_repo

import (
	"testing"

	country_repo "github.com/Neratus/geoguide/internal/repository/country"
)

func TestMockCityRepository(t *testing.T) {
	cityRepo, _ := NewMockCityRepository()
	countryRepo, _ := country_repo.NewMockCountryRepository()

	t.Run("CityNotFound", func(t *testing.T) {
		cityRepository_CityNotFound(t, cityRepo)
	})
	t.Run("CityFound", func(t *testing.T) {
		cityRepository_CityFound(t, cityRepo, countryRepo)
	})
	t.Run("CityFindByCountry", func(t *testing.T) {
		cityRepository_CityFindByCountry(t, cityRepo, countryRepo)
	})
	t.Run("CityFindByName", func(t *testing.T) {
		cityRepository_CityFindByName(t, cityRepo, countryRepo)
	})
	t.Run("CityUpdate", func(t *testing.T) {
		cityRepository_CityUpdate(t, cityRepo, countryRepo)
	})
	t.Run("CityDelete", func(t *testing.T) {
		cityRepository_CityDelete(t, cityRepo, countryRepo)
	})

	t.Run("DistrictNotFound", func(t *testing.T) {
		cityRepository_DistrictNotFound(t, cityRepo)
	})
	t.Run("DistrictSaveAndFind", func(t *testing.T) {
		cityRepository_DistrictSaveAndFind(t, cityRepo, countryRepo)
	})
	t.Run("DistrictUpdate", func(t *testing.T) {
		cityRepository_DistrictUpdate(t, cityRepo, countryRepo)
	})
	t.Run("DistrictDelete", func(t *testing.T) {
		cityRepository_DistrictDelete(t, cityRepo, countryRepo)
	})

	t.Run("TransportNodeNotFound", func(t *testing.T) {
		cityRepository_TransportNodeNotFound(t, cityRepo)
	})
	t.Run("TransportNodeSaveAndFind", func(t *testing.T) {
		cityRepository_TransportNodeSaveAndFind(t, cityRepo, countryRepo)
	})
	t.Run("TransportNodeFindByCity", func(t *testing.T) {
		cityRepository_TransportNodeFindByCity(t, cityRepo, countryRepo)
	})
	t.Run("TransportNodeUpdate", func(t *testing.T) {
		cityRepository_TransportNodeUpdate(t, cityRepo, countryRepo)
	})
	t.Run("TransportNodeDelete", func(t *testing.T) {
		cityRepository_TransportNodeDelete(t, cityRepo, countryRepo)
	})
}
