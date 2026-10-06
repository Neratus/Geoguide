package place_repo

import (
	"testing"

	city_repo "github.com/Neratus/geoguide/internal/repository/city"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
)

func TestMockPlaceRepository(t *testing.T) {
	t.Run("SaveNewPlace", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_SaveNewPlace(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("SaveExistingPlace", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_SaveExistingPlace(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		placeRepository_FindByID_NotFound(t, placeRepo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_FindByID_Found(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindByCity", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_FindByCity(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindByCategory", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_FindByCategory(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("UpdatePlace_Success", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_UpdatePlace_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("DeletePlace_Success", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_DeletePlace_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("UpdateRating_Success", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_UpdateRating_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("SaveNewReview", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_SaveNewReview(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("SaveExistingReview", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_SaveExistingReview(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindReviewByID_NotFound", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		placeRepository_FindReviewByID_NotFound(t, placeRepo)
	})
	t.Run("FindReviewByID_Found", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_FindReviewByID_Found(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindReviewsByPlace", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_FindReviewsByPlace(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindReviewsByUser", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_FindReviewsByUser(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("UpdateReview_Success", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_UpdateReview_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("DeleteReview_Success", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_DeleteReview_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("ModerateReview_Success", func(t *testing.T) {
		placeRepo, _ := NewMockPlaceRepository()
		countryRepo, _ := country_repo.NewMockCountryRepository()
		cityRepo, _ := city_repo.NewMockCityRepository()
		userRepo, _ := user_repo.NewMockUserRepo()
		placeRepository_ModerateReview_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
}
