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

var _ interfaces.CityRepository = (*MockCityRepository)(nil)

type MockCityRepository struct {
	sync.RWMutex
	cities    map[domain.CityID]*domain.City
	districts map[domain.CityDistrictID]*domain.CityDistrict
	nodes     map[domain.NodeID]*domain.TransportNode

	Calls struct {
		Save, FindByID, FindByCountry, FindByName, SearchCities, Update, Delete                                      int
		SaveDistrict, FindDistrictsByCity, UpdateDistrict, DeleteDistrict                                            int
		SaveTransportNode, FindTransportNodeByID, FindTransportNodesByCity, UpdateTransportNode, DeleteTransportNode int
	}
	ForceError struct {
		Save, FindByID, FindByCountry, FindByName, SearchCities, Update, Delete error
	}
}

func NewMockCityRepository() *MockCityRepository {
	return &MockCityRepository{
		cities:    make(map[domain.CityID]*domain.City),
		districts: make(map[domain.CityDistrictID]*domain.CityDistrict),
		nodes:     make(map[domain.NodeID]*domain.TransportNode),
	}
}

func (m *MockCityRepository) Save(ctx context.Context, city *domain.City) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Save++
	if m.ForceError.Save != nil {
		return m.ForceError.Save
	}
	if city.GetId() == (domain.CityID{}) {
		city.SetId(domain.CityID(uuid.New()))
	}
	cpy := *city
	m.cities[city.GetId()] = &cpy
	return nil
}

func (m *MockCityRepository) FindByID(ctx context.Context, id domain.CityID) (*domain.City, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindByID++
	if m.ForceError.FindByID != nil {
		return nil, m.ForceError.FindByID
	}
	city, ok := m.cities[id]
	if !ok {
		return nil, errors.New("city not found")
	}
	cpy := *city
	return &cpy, nil
}

func (m *MockCityRepository) FindByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.City, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindByCountry++
	if m.ForceError.FindByCountry != nil {
		return nil, m.ForceError.FindByCountry
	}
	var result []*domain.City
	for _, c := range m.cities {
		if c.GetCountryId() == countryID {
			cpy := *c
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (m *MockCityRepository) FindByName(ctx context.Context, name string) (*domain.City, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindByName++
	if m.ForceError.FindByName != nil {
		return nil, m.ForceError.FindByName
	}
	for _, c := range m.cities {
		if strings.EqualFold(c.GetName(), name) {
			cpy := *c
			return &cpy, nil
		}
	}
	return nil, errors.New("city not found")
}

func (m *MockCityRepository) SearchCities(ctx context.Context, query string, limit, offset int) ([]*domain.City, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.SearchCities++
	if m.ForceError.SearchCities != nil {
		return nil, m.ForceError.SearchCities
	}
	q := strings.ToLower(query)
	var matched []*domain.City
	for _, c := range m.cities {
		if strings.Contains(strings.ToLower(c.GetName()), q) {
			cpy := *c
			matched = append(matched, &cpy)
		}
	}
	start := offset
	if start > len(matched) {
		start = len(matched)
	}
	end := start + limit
	if end > len(matched) {
		end = len(matched)
	}
	return matched[start:end], nil
}

func (m *MockCityRepository) Update(ctx context.Context, city *domain.City) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Update++
	if m.ForceError.Update != nil {
		return m.ForceError.Update
	}
	if _, ok := m.cities[city.GetId()]; !ok {
		return errors.New("city not found")
	}
	cpy := *city
	m.cities[city.GetId()] = &cpy
	return nil
}

func (m *MockCityRepository) Delete(ctx context.Context, id domain.CityID) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Delete++
	if m.ForceError.Delete != nil {
		return m.ForceError.Delete
	}
	if _, ok := m.cities[id]; !ok {
		return errors.New("city not found")
	}
	delete(m.cities, id)
	return nil
}

func (m *MockCityRepository) SaveDistrict(ctx context.Context, district *domain.CityDistrict) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.SaveDistrict++
	if district.GetId() == (domain.CityDistrictID{}) {
		district.SetId(domain.CityDistrictID(uuid.New()))
	}
	cpy := *district
	m.districts[district.GetId()] = &cpy
	return nil
}

func (m *MockCityRepository) FindDistrictsByCity(ctx context.Context, cityID domain.CityID) ([]*domain.CityDistrict, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindDistrictsByCity++
	var result []*domain.CityDistrict
	for _, d := range m.districts {
		if d.GetCityId() == cityID {
			cpy := *d
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (m *MockCityRepository) UpdateDistrict(ctx context.Context, district *domain.CityDistrict) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.UpdateDistrict++
	if _, ok := m.districts[district.GetId()]; !ok {
		return errors.New("district not found")
	}
	cpy := *district
	m.districts[district.GetId()] = &cpy
	return nil
}

func (m *MockCityRepository) DeleteDistrict(ctx context.Context, id domain.CityDistrictID) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.DeleteDistrict++
	if _, ok := m.districts[id]; !ok {
		return errors.New("district not found")
	}
	delete(m.districts, id)
	return nil
}

func (m *MockCityRepository) SaveTransportNode(ctx context.Context, node *domain.TransportNode) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.SaveTransportNode++
	if node.GetId() == (domain.NodeID{}) {
		node.SetId(domain.NodeID(uuid.New()))
	}
	cpy := *node
	m.nodes[node.GetId()] = &cpy
	return nil
}

func (m *MockCityRepository) FindTransportNodeByID(ctx context.Context, id domain.NodeID) (*domain.TransportNode, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindTransportNodeByID++
	n, ok := m.nodes[id]
	if !ok {
		return nil, errors.New("transport node not found")
	}
	cpy := *n
	return &cpy, nil
}

func (m *MockCityRepository) FindTransportNodesByCity(ctx context.Context, cityID domain.CityID) ([]*domain.TransportNode, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindTransportNodesByCity++
	var result []*domain.TransportNode
	for _, n := range m.nodes {
		if n.GetCityID() == cityID {
			cpy := *n
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (m *MockCityRepository) UpdateTransportNode(ctx context.Context, node *domain.TransportNode) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.UpdateTransportNode++
	if _, ok := m.nodes[node.GetId()]; !ok {
		return errors.New("transport node not found")
	}
	cpy := *node
	m.nodes[node.GetId()] = &cpy
	return nil
}

func (m *MockCityRepository) DeleteTransportNode(ctx context.Context, id domain.NodeID) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.DeleteTransportNode++
	if _, ok := m.nodes[id]; !ok {
		return errors.New("transport node not found")
	}
	delete(m.nodes, id)
	return nil
}
