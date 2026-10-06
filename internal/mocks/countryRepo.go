package mocks

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
)

var _ interfaces.CountryRepository = (*MockCountryRepository)(nil)

type MockCountryRepository struct {
	sync.RWMutex
	countries map[domain.CountryID]*domain.Country
	holidays  map[domain.HolidayID]*domain.Holiday

	Calls struct {
		Save, FindByID, FindAll, FindByName, Update, Delete                                                   int
		SaveHoliday, FindHolidayByID, FindHolidaysByCountry, FindHolidaysByDate, UpdateHoliday, DeleteHoliday int
	}
	ForceError struct {
		Save, FindByID, FindAll, FindByName, Update, Delete    error
		SaveHoliday, FindHolidaysByDate, FindHolidaysByCountry error
	}
}

func NewMockCountryRepository() *MockCountryRepository {
	return &MockCountryRepository{
		countries: make(map[domain.CountryID]*domain.Country),
		holidays:  make(map[domain.HolidayID]*domain.Holiday),
	}
}

func (m *MockCountryRepository) Save(ctx context.Context, country *domain.Country) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Save++
	if m.ForceError.Save != nil {
		return m.ForceError.Save
	}
	if country.GetId() == (domain.CountryID{}) {
		country.SetId(domain.CountryID(uuid.New()))
	}
	cpy := *country
	m.countries[country.GetId()] = &cpy
	return nil
}

func (m *MockCountryRepository) FindByID(ctx context.Context, id domain.CountryID) (*domain.Country, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindByID++
	if m.ForceError.FindByID != nil {
		return nil, m.ForceError.FindByID
	}
	c, ok := m.countries[id]
	if !ok {
		return nil, errors.New("country not found")
	}
	cpy := *c
	return &cpy, nil
}

func (m *MockCountryRepository) FindAll(ctx context.Context) ([]*domain.Country, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindAll++
	if m.ForceError.FindAll != nil {
		return nil, m.ForceError.FindAll
	}
	var result []*domain.Country
	for _, c := range m.countries {
		cpy := *c
		result = append(result, &cpy)
	}
	return result, nil
}

func (m *MockCountryRepository) FindByName(ctx context.Context, name string) (*domain.Country, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindByName++
	if m.ForceError.FindByName != nil {
		return nil, m.ForceError.FindByName
	}
	for _, c := range m.countries {
		if strings.EqualFold(c.GetName(), name) {
			cpy := *c
			return &cpy, nil
		}
	}
	return nil, errors.New("country not found")
}

func (m *MockCountryRepository) Update(ctx context.Context, country *domain.Country) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Update++
	if m.ForceError.Update != nil {
		return m.ForceError.Update
	}
	if _, ok := m.countries[country.GetId()]; !ok {
		return errors.New("country not found")
	}
	cpy := *country
	m.countries[country.GetId()] = &cpy
	return nil
}

func (m *MockCountryRepository) Delete(ctx context.Context, id domain.CountryID) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Delete++
	if m.ForceError.Delete != nil {
		return m.ForceError.Delete
	}
	if _, ok := m.countries[id]; !ok {
		return errors.New("country not found")
	}
	delete(m.countries, id)
	return nil
}

func (m *MockCountryRepository) FindHolidayByID(ctx context.Context, id domain.HolidayID) (*domain.Holiday, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindHolidayByID++
	h, ok := m.holidays[id]
	if !ok {
		return nil, errors.New("holiday not found")
	}
	cpy := *h
	return &cpy, nil
}

func (m *MockCountryRepository) SaveHoliday(ctx context.Context, holiday *domain.Holiday) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.SaveHoliday++
	if m.ForceError.SaveHoliday != nil {
		return m.ForceError.SaveHoliday
	}
	if holiday.GetId() == (domain.HolidayID{}) {
		holiday.SetId(domain.HolidayID(uuid.New()))
	}
	cpy := *holiday
	m.holidays[holiday.GetId()] = &cpy
	return nil
}

func (m *MockCountryRepository) FindHolidaysByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.Holiday, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindHolidaysByCountry++
	var result []*domain.Holiday
	for _, h := range m.holidays {
		if h.GetCountryID() == countryID {
			cpy := *h
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (m *MockCountryRepository) FindHolidaysByDate(ctx context.Context, date string) ([]*domain.Holiday, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindHolidaysByDate++
	if m.ForceError.FindHolidaysByDate != nil {
		return nil, m.ForceError.FindHolidaysByDate
	}
	var result []*domain.Holiday
	for _, h := range m.holidays {
		if h.GetDate().Format("2006-01-02") == date {
			cpy := *h
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (m *MockCountryRepository) UpdateHoliday(ctx context.Context, holiday *domain.Holiday) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.UpdateHoliday++
	if _, ok := m.holidays[holiday.GetId()]; !ok {
		return errors.New("holiday not found")
	}
	cpy := *holiday
	m.holidays[holiday.GetId()] = &cpy
	return nil
}

func (m *MockCountryRepository) DeleteHoliday(ctx context.Context, id domain.HolidayID) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.DeleteHoliday++
	if _, ok := m.holidays[id]; !ok {
		return errors.New("holiday not found")
	}
	delete(m.holidays, id)
	return nil
}
