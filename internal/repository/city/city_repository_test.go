package city_repo

import (
	"context"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func generateTestCityWithCountry(countryID domain.CountryID) *domain.City {
	coords := domain.NewCoordinates(48.8566, 2.3522)
	city, err := domain.NewCity(
		uuid.Nil, "Paris", 2148000, true, coords,
		"Capital of France", "Europe/Paris", "Great city",
		countryID, uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
	return city
}

func generateTestDistrictWithCity(cityID domain.CityID) *domain.CityDistrict {
	coords := domain.NewCoordinates(48.8566, 2.3522)
	district, err := domain.NewCityDistrict(
		"Louvre District", "Central area", coords, cityID, uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
	return district
}

func generateTestTransportNodeWithCity(cityID domain.CityID) *domain.TransportNode {
	coords := domain.NewCoordinates(48.8566, 2.3522)
	node, err := domain.NewTransportNode(
		uuid.Nil, "Paris-Charles de Gaulle", "airport", coords,
		"Roissy-en-France", cityID, uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
	return node
}

func cityRepository_CityNotFound(t *testing.T, repo interfaces.CityRepository) {
	_, err := repo.FindByID(context.Background(), uuid.Nil)
	require.Error(t, err)
}

func cityRepository_CityFound(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Beautiful country", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, city.GetId())

	found, err := cityRepo.FindByID(ctx, city.GetId())
	require.NoError(t, err)
	require.Equal(t, city.GetId(), found.GetId())
	require.Equal(t, city.GetName(), found.GetName())
	require.Equal(t, city.GetPopulation(), found.GetPopulation())
	require.Equal(t, city.IsCapital(), found.IsCapital())
	require.Equal(t, city.GetCoordinates(), found.GetCoordinates())
	require.Equal(t, city.GetDescription(), found.GetDescription())
	require.Equal(t, city.GetTimezone(), found.GetTimezone())
	require.Equal(t, city.GetTravelTips(), found.GetTravelTips())
	require.Equal(t, city.GetCountryId(), found.GetCountryId())
}

func cityRepository_CityFindByCountry(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city1 := generateTestCityWithCountry(country.GetId())
	city1.SetName("Paris")
	err = cityRepo.Save(ctx, city1)
	require.NoError(t, err)

	city2 := generateTestCityWithCountry(country.GetId())
	city2.SetName("Lyon")
	err = cityRepo.Save(ctx, city2)
	require.NoError(t, err)

	otherCountry, err := domain.NewCountry(
		uuid.Nil, "Germany", 357582.0, 3860000.0, 83200000,
		"EUR", "Schengen", "Description", "Safe", "Summer",
		"German", "+49", "Christian", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, otherCountry)
	require.NoError(t, err)

	city3 := generateTestCityWithCountry(otherCountry.GetId())
	city3.SetName("Berlin")
	err = cityRepo.Save(ctx, city3)
	require.NoError(t, err)

	found, err := cityRepo.FindByCountry(ctx, country.GetId())
	require.NoError(t, err)
	require.Len(t, found, 2)

	names := make(map[string]bool)
	for _, c := range found {
		names[c.GetName()] = true
	}
	require.True(t, names["Paris"])
	require.True(t, names["Lyon"])
	require.False(t, names["Berlin"])
}

func cityRepository_CityFindByName(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "TestLand", 1000.0, 1000.0, 100000,
		"USD", "None", "", "", "", "English", "+1", "",
		uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	city.SetName("UniqueCity")
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	found, err := cityRepo.FindByName(ctx, "UniqueCity")
	require.NoError(t, err)
	require.Equal(t, city.GetId(), found.GetId())

	_, err = cityRepo.FindByName(ctx, "NonExisting")
	require.Error(t, err)
}

func cityRepository_CityUpdate(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	city.SetName("NewName")
	city.SetPopulation(3000000)
	err = cityRepo.Update(ctx, city)
	require.NoError(t, err)

	updated, err := cityRepo.FindByID(ctx, city.GetId())
	require.NoError(t, err)
	require.Equal(t, "NewName", updated.GetName())
	require.Equal(t, int64(3000000), updated.GetPopulation())
}

func cityRepository_CityDelete(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)
	id := city.GetId()

	err = cityRepo.Delete(ctx, id)
	require.NoError(t, err)

	_, err = cityRepo.FindByID(ctx, id)
	require.Error(t, err)
}

func cityRepository_DistrictNotFound(t *testing.T, repo interfaces.CityRepository) {
	found, err := repo.FindDistrictsByCity(context.Background(), uuid.Nil)
	require.NoError(t, err)
	require.Empty(t, found)
}

func cityRepository_DistrictSaveAndFind(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	district := generateTestDistrictWithCity(city.GetId())
	err = cityRepo.SaveDistrict(ctx, district)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, district.GetId())

	found, err := cityRepo.FindDistrictsByCity(ctx, city.GetId())
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, district.GetId(), found[0].GetId())
}

func cityRepository_DistrictUpdate(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	district := generateTestDistrictWithCity(city.GetId())
	err = cityRepo.SaveDistrict(ctx, district)
	require.NoError(t, err)

	district.SetName("NewDistrictName")
	err = cityRepo.UpdateDistrict(ctx, district)
	require.NoError(t, err)

	found, err := cityRepo.FindDistrictsByCity(ctx, city.GetId())
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, "NewDistrictName", found[0].GetName())
}

func cityRepository_DistrictDelete(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	district := generateTestDistrictWithCity(city.GetId())
	err = cityRepo.SaveDistrict(ctx, district)
	require.NoError(t, err)
	id := district.GetId()

	err = cityRepo.DeleteDistrict(ctx, id)
	require.NoError(t, err)

	found, err := cityRepo.FindDistrictsByCity(ctx, city.GetId())
	require.NoError(t, err)
	require.Empty(t, found)
}

func cityRepository_TransportNodeNotFound(t *testing.T, repo interfaces.CityRepository) {
	_, err := repo.FindTransportNodeByID(context.Background(), uuid.Nil)
	require.Error(t, err)
}

func cityRepository_TransportNodeSaveAndFind(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	node := generateTestTransportNodeWithCity(city.GetId())
	err = cityRepo.SaveTransportNode(ctx, node)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, node.GetId())

	found, err := cityRepo.FindTransportNodeByID(ctx, node.GetId())
	require.NoError(t, err)
	require.Equal(t, node.GetName(), found.GetName())
}

func cityRepository_TransportNodeFindByCity(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	node1 := generateTestTransportNodeWithCity(city.GetId())
	node1.SetName("CDG")
	err = cityRepo.SaveTransportNode(ctx, node1)
	require.NoError(t, err)

	node2 := generateTestTransportNodeWithCity(city.GetId())
	node2.SetName("Orly")
	err = cityRepo.SaveTransportNode(ctx, node2)
	require.NoError(t, err)

	otherCity := generateTestCityWithCountry(country.GetId())
	otherCity.SetName("Lyon")
	err = cityRepo.Save(ctx, otherCity)
	require.NoError(t, err)

	node3 := generateTestTransportNodeWithCity(otherCity.GetId())
	node3.SetName("LYS")
	err = cityRepo.SaveTransportNode(ctx, node3)
	require.NoError(t, err)

	found, err := cityRepo.FindTransportNodesByCity(ctx, city.GetId())
	require.NoError(t, err)
	require.Len(t, found, 2)

	names := make(map[string]bool)
	for _, n := range found {
		names[n.GetName()] = true
	}
	require.True(t, names["CDG"])
	require.True(t, names["Orly"])
	require.False(t, names["LYS"])
}

func cityRepository_TransportNodeUpdate(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	node := generateTestTransportNodeWithCity(city.GetId())
	err = cityRepo.SaveTransportNode(ctx, node)
	require.NoError(t, err)
	id := node.GetId()

	node.SetName("Updated Name")
	err = cityRepo.UpdateTransportNode(ctx, node)
	require.NoError(t, err)

	updated, err := cityRepo.FindTransportNodeByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Updated Name", updated.GetName())
}

func cityRepository_TransportNodeDelete(t *testing.T, cityRepo interfaces.CityRepository, countryRepo interfaces.CountryRepository) {
	ctx := context.Background()

	country, err := domain.NewCountry(
		uuid.Nil, "France", 551695.0, 2716000.0, 67390000,
		"EUR", "Schengen", "Description", "Safe", "Spring",
		"French", "+33", "Secular", uuid.Nil, uuid.Nil,
	)
	require.NoError(t, err)
	err = countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCityWithCountry(country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	node := generateTestTransportNodeWithCity(city.GetId())
	err = cityRepo.SaveTransportNode(ctx, node)
	require.NoError(t, err)
	id := node.GetId()

	err = cityRepo.DeleteTransportNode(ctx, id)
	require.NoError(t, err)

	_, err = cityRepo.FindTransportNodeByID(ctx, id)
	require.Error(t, err)
}
