package country_repo

import (
	"context"
	"errors"
	"sync"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type MockCountryRepository struct {
	sync.RWMutex
	counties      map[domain.CountryID]domain.Country
	holidays      map[domain.HolidayID]domain.Holiday
	lastCountryID uint64
	lastHolidayID uint64
}

func NewMockCountryRepository() (*MockCountryRepository, error) {
	return &MockCountryRepository{
		counties: make(map[domain.CountryID]domain.Country),
		holidays: make(map[domain.HolidayID]domain.Holiday),
	}, nil
}

func (r *MockCountryRepository) Save(ctx context.Context, country *domain.Country) error {
	r.Lock()
	defer r.Unlock()
	if country.GetId() == uuid.Nil {
		r.lastCountryID++
		newID := domain.Uint64ToUUID(r.lastCountryID)
		country.SetId(newID)
		r.counties[newID] = *country
		return nil
	}
	if _, ok := r.counties[country.GetId()]; !ok {
		return errors.New("country not found")
	}
	r.counties[country.GetId()] = *country
	return nil

}

func (r *MockCountryRepository) FindByID(ctx context.Context, id domain.CountryID) (*domain.Country, error) {
	r.RLock()
	defer r.RUnlock()
	country, ok := r.counties[id]
	if !ok {
		return nil, errors.New("country not found")
	}
	cpy := country
	return &cpy, nil
}

func (r *MockCountryRepository) FindHolidayByID(ctx context.Context, id domain.HolidayID) (*domain.Holiday, error) {
	r.RLock()
	defer r.RUnlock()
	holiday, ok := r.holidays[id]
	if !ok {
		return nil, errors.New("holiday not found")
	}
	cpy := holiday
	return &cpy, nil
}

func (r *MockCountryRepository) FindAll(ctx context.Context) ([]*domain.Country, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Country
	for _, c := range r.counties {
		cpy := c
		result = append(result, &cpy)
	}
	return result, nil
}

func (r *MockCountryRepository) FindByName(ctx context.Context, name string) (*domain.Country, error) {
	r.RLock()
	defer r.RUnlock()
	for _, c := range r.counties {
		if c.GetName() == name {
			cpy := c
			return &cpy, nil
		}
	}
	return nil, errors.New("Country not found")
}

func (r *MockCountryRepository) Update(ctx context.Context, country *domain.Country) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.counties[country.GetId()]; !ok {
		return errors.New("Country not found")
	}
	r.counties[country.GetId()] = *country
	return nil
}

func (r *MockCountryRepository) Delete(ctx context.Context, id domain.CountryID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.counties[id]; !ok {
		return errors.New("country not found")
	}
	delete(r.counties, id)
	return nil
}

func (r *MockCountryRepository) SaveHoliday(ctx context.Context, holiday *domain.Holiday) error {
	r.Lock()
	defer r.Unlock()
	if holiday.GetId() == uuid.Nil {
		r.lastHolidayID++
		newID := domain.Uint64ToUUID(r.lastHolidayID)
		holiday.SetId(newID)
		r.holidays[newID] = *holiday
		return nil
	}
	if _, ok := r.holidays[holiday.GetId()]; !ok {
		return errors.New("holiday not found")
	}
	r.holidays[holiday.GetId()] = *holiday
	return nil
}

func (r *MockCountryRepository) FindHolidaysByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.Holiday, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Holiday
	for _, c := range r.holidays {
		if c.GetCountryID() == countryID {
			cpy := c
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockCountryRepository) FindHolidaysByDate(ctx context.Context, date string) ([]*domain.Holiday, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Holiday
	for _, c := range r.holidays {
		if c.GetDate().Format("2006-01-02") == date {
			cpy := c
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockCountryRepository) UpdateHoliday(ctx context.Context, holiday *domain.Holiday) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.holidays[holiday.GetId()]; !ok {
		return errors.New("holiday not found")
	}
	r.holidays[holiday.GetId()] = *holiday
	return nil
}

func (r *MockCountryRepository) DeleteHoliday(ctx context.Context, id domain.HolidayID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.holidays[id]; !ok {
		return errors.New("holiday not found")
	}
	delete(r.holidays, id)
	return nil
}
