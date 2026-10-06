package trip_repo

import (
	"testing"

	city_repo "github.com/Neratus/geoguide/internal/repository/city"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
	place_repo "github.com/Neratus/geoguide/internal/repository/place"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
)

func TestMockTripRepository(t *testing.T) {
	tripRepo, _ := NewMockTripRepository()
	userRepo, _ := user_repo.NewMockUserRepo()
	placeRepo, _ := place_repo.NewMockPlaceRepository()
	countryRepo, _ := country_repo.NewMockCountryRepository()
	cityRepo, _ := city_repo.NewMockCityRepository()

	t.Run("SaveNewTrip", func(t *testing.T) {
		tripRepository_SaveNewTrip(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("SaveExistingTrip", func(t *testing.T) {
		tripRepository_SaveExistingTrip(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		tripRepository_FindByID_NotFound(t, tripRepo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		tripRepository_FindByID_Found(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("FindByUser", func(t *testing.T) {
		tripRepository_FindByUser(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("UpdateTrip_Success", func(t *testing.T) {
		tripRepository_UpdateTrip_Success(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("DeleteTrip_Success", func(t *testing.T) {
		tripRepository_DeleteTrip_Success(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("AddPlace_Success", func(t *testing.T) {
		tripRepository_AddPlace_Success(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("AddPlace_TripNotFound", func(t *testing.T) {
		tripRepository_AddPlace_TripNotFound(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("RemovePlace_Success", func(t *testing.T) {
		tripRepository_RemovePlace_Success(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("GetPlaces_TripNotFound", func(t *testing.T) {
		tripRepository_GetPlaces_TripNotFound(t, tripRepo)
	})
	t.Run("GetPlaces_Success", func(t *testing.T) {
		tripRepository_GetPlaces_Success(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
	t.Run("UpdateTripPlace_Success", func(t *testing.T) {
		tripRepository_UpdateTripPlace_Success(t, tripRepo, userRepo, countryRepo, cityRepo, placeRepo)
	})
}
