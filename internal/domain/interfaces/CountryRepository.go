package interfaces

import (
	"context"

	"github.com/Neratus/geoguide/internal/domain"
)

type CountryRepository interface {
	Save(ctx context.Context, country *domain.Country) error
	FindByID(ctx context.Context, id domain.CountryID) (*domain.Country, error)
	FindHolidayByID(ctx context.Context, id domain.HolidayID) (*domain.Holiday, error)
	FindAll(ctx context.Context) ([]*domain.Country, error)
	FindByName(ctx context.Context, name string) (*domain.Country, error)
	Update(ctx context.Context, country *domain.Country) error
	Delete(ctx context.Context, id domain.CountryID) error

	SaveHoliday(ctx context.Context, holiday *domain.Holiday) error
	FindHolidaysByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.Holiday, error)
	FindHolidaysByDate(ctx context.Context, date string) ([]*domain.Holiday, error)
	UpdateHoliday(ctx context.Context, holiday *domain.Holiday) error
	DeleteHoliday(ctx context.Context, id domain.HolidayID) error
}
