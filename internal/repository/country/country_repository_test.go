package country_repo

import (
	"context"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func generateTestCountry() *domain.Country {

	country, err := domain.NewCountry(
		uuid.Nil,
		"France",
		551695,
		2.778e12,
		67390000,
		"EUR",
		"Schengen visa",
		"Country description",
		"Safe to travel",
		"Spring",
		"French",
		"+33",
		"Christianity",
		uuid.Nil,
		uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
	return country
}

func generateTestHoliday(countryID domain.CountryID) *domain.Holiday {
	date, _ := time.Parse("2006-01-02", "2024-07-14")
	holiday, err := domain.NewHoliday(
		uuid.Nil,
		"Bastille Day",
		date,
		"National holiday",
		"Fireworks",
		"Storming of the Bastille",
		true,
		countryID,
		uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
	return holiday
}

func countryRepository_SaveNewCountry(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	err := repo.Save(context.Background(), country)
	require.NoError(t, err)
	require.NotZero(t, country.GetId())
}

func countryRepository_FindByID_NotFound(t *testing.T, repo interfaces.CountryRepository) {
	_, err := repo.FindByID(context.Background(), domain.Int64ToUUID(9999))
	require.Error(t, err)
}

func countryRepository_FindByID_Found(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	err := repo.Save(context.Background(), country)
	require.NoError(t, err)
	id := country.GetId()

	found, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, country.GetId(), found.GetId())
	require.Equal(t, country.GetName(), found.GetName())
}

func countryRepository_FindAll(t *testing.T, repo interfaces.CountryRepository) {
	c1 := generateTestCountry()
	c1.SetName("France")
	_ = repo.Save(context.Background(), c1)

	c2 := generateTestCountry()
	c2.SetName("Italy")
	_ = repo.Save(context.Background(), c2)

	all, err := repo.FindAll(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 2)
}

func countryRepository_FindByName_NotFound(t *testing.T, repo interfaces.CountryRepository) {
	_, err := repo.FindByName(context.Background(), "Nonexistent")
	require.Error(t, err)
}

func countryRepository_FindByName_Found(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	country.SetName("Spain")
	err := repo.Save(context.Background(), country)
	require.NoError(t, err)

	found, err := repo.FindByName(context.Background(), "Spain")
	require.NoError(t, err)
	require.Equal(t, country.GetId(), found.GetId())
}

func countryRepository_UpdateCountry_Success(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	err := repo.Save(context.Background(), country)
	require.NoError(t, err)
	id := country.GetId()

	country.SetName("Updated")
	err = repo.Update(context.Background(), country)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "Updated", found.GetName())
}

func countryRepository_DeleteCountry_Success(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	err := repo.Save(context.Background(), country)
	require.NoError(t, err)
	id := country.GetId()

	err = repo.Delete(context.Background(), id)
	require.NoError(t, err)

	_, err = repo.FindByID(context.Background(), id)
	require.Error(t, err)
}

func countryRepository_SaveNewHoliday(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	err := repo.Save(context.Background(), country)
	require.NoError(t, err)

	holiday := generateTestHoliday(country.GetId())
	err = repo.SaveHoliday(context.Background(), holiday)
	require.NoError(t, err)
	require.NotZero(t, holiday.GetId())
}

func countryRepository_FindHolidayByID_NotFound(t *testing.T, repo interfaces.CountryRepository) {
	_, err := repo.FindHolidayByID(context.Background(), domain.Int64ToUUID(9999))
	require.Error(t, err)
}

func countryRepository_FindHolidayByID_Found(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	_ = repo.Save(context.Background(), country)
	holiday := generateTestHoliday(country.GetId())
	err := repo.SaveHoliday(context.Background(), holiday)
	require.NoError(t, err)
	id := holiday.GetId()

	found, err := repo.FindHolidayByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, holiday.GetId(), found.GetId())
}

func countryRepository_FindHolidaysByCountry(t *testing.T, repo interfaces.CountryRepository) {
	country1 := generateTestCountry()
	country1.SetName("France")
	_ = repo.Save(context.Background(), country1)
	country2 := generateTestCountry()
	country2.SetName("Italy")
	_ = repo.Save(context.Background(), country2)

	h1 := generateTestHoliday(country1.GetId())
	h2 := generateTestHoliday(country1.GetId())
	h3 := generateTestHoliday(country2.GetId())

	_ = repo.SaveHoliday(context.Background(), h1)
	_ = repo.SaveHoliday(context.Background(), h2)
	_ = repo.SaveHoliday(context.Background(), h3)

	holidaysFR, err := repo.FindHolidaysByCountry(context.Background(), country1.GetId())
	require.NoError(t, err)
	require.Len(t, holidaysFR, 2)

	holidaysIT, err := repo.FindHolidaysByCountry(context.Background(), country2.GetId())
	require.NoError(t, err)
	require.Len(t, holidaysIT, 1)
}

func countryRepository_FindHolidaysByDate(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	_ = repo.Save(context.Background(), country)

	date1, _ := time.Parse("2006-01-02", "2024-07-14")
	date2, _ := time.Parse("2006-01-02", "2024-07-14")
	date3, _ := time.Parse("2006-01-02", "2025-01-01")

	h1 := generateTestHoliday(country.GetId())
	h1.SetDate(date1)
	_ = repo.SaveHoliday(context.Background(), h1)

	h2 := generateTestHoliday(country.GetId())
	h2.SetDate(date2)
	_ = repo.SaveHoliday(context.Background(), h2)

	h3 := generateTestHoliday(country.GetId())
	h3.SetDate(date3)
	_ = repo.SaveHoliday(context.Background(), h3)

	holidays, err := repo.FindHolidaysByDate(context.Background(), "2024-07-14")
	require.NoError(t, err)
	require.Len(t, holidays, 2)
}

func countryRepository_UpdateHoliday_Success(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	_ = repo.Save(context.Background(), country)
	holiday := generateTestHoliday(country.GetId())
	err := repo.SaveHoliday(context.Background(), holiday)
	require.NoError(t, err)
	id := holiday.GetId()

	holiday.SetDescription("Updated description")
	err = repo.UpdateHoliday(context.Background(), holiday)
	require.NoError(t, err)

	found, err := repo.FindHolidayByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "Updated description", found.GetDescription())
}

func countryRepository_DeleteHoliday_Success(t *testing.T, repo interfaces.CountryRepository) {
	country := generateTestCountry()
	_ = repo.Save(context.Background(), country)
	holiday := generateTestHoliday(country.GetId())
	err := repo.SaveHoliday(context.Background(), holiday)
	require.NoError(t, err)
	id := holiday.GetId()

	err = repo.DeleteHoliday(context.Background(), id)
	require.NoError(t, err)

	_, err = repo.FindHolidayByID(context.Background(), id)
	require.Error(t, err)
}
