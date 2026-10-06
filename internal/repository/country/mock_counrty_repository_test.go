package country_repo

import (
	"testing"
)

func TestMockCountryRepository(t *testing.T) {
	t.Run("SaveNewCountry", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_SaveNewCountry(t, repo)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindByID_NotFound(t, repo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindByID_Found(t, repo)
	})
	t.Run("FindAll", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindAll(t, repo)
	})
	t.Run("FindByName_NotFound", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindByName_NotFound(t, repo)
	})
	t.Run("FindByName_Found", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindByName_Found(t, repo)
	})
	t.Run("UpdateCountry_Success", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_UpdateCountry_Success(t, repo)
	})
	t.Run("DeleteCountry_Success", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_DeleteCountry_Success(t, repo)
	})

	t.Run("SaveNewHoliday", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_SaveNewHoliday(t, repo)
	})
	t.Run("FindHolidayByID_NotFound", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindHolidayByID_NotFound(t, repo)
	})
	t.Run("FindHolidayByID_Found", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindHolidayByID_Found(t, repo)
	})
	t.Run("FindHolidaysByCountry", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindHolidaysByCountry(t, repo)
	})
	t.Run("FindHolidaysByDate", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_FindHolidaysByDate(t, repo)
	})
	t.Run("UpdateHoliday_Success", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_UpdateHoliday_Success(t, repo)
	})
	t.Run("DeleteHoliday_Success", func(t *testing.T) {
		repo, _ := NewMockCountryRepository()
		countryRepository_DeleteHoliday_Success(t, repo)
	})
}
