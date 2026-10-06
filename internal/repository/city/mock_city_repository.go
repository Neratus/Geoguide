package city_repo

import (
	"context"
	"errors"
	"sync"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type MockCityRepo struct {
	sync.RWMutex
	cities         map[domain.CityID]domain.City
	districts      map[domain.CityDistrictID]domain.CityDistrict
	nodes          map[domain.NodeID]domain.TransportNode
	lastCityID     uint64
	lastDistrictID uint64
	lastNodeID     uint64
}

func NewMockCityRepository() (*MockCityRepo, error) {
	return &MockCityRepo{
		cities:    make(map[domain.CityID]domain.City),
		districts: make(map[domain.CityDistrictID]domain.CityDistrict),
		nodes:     make(map[domain.NodeID]domain.TransportNode),
	}, nil
}

func (r *MockCityRepo) Save(ctx context.Context, city *domain.City) error {
	r.Lock()
	defer r.Unlock()
	id := city.GetId()
	if domain.UUIDToInt64(id) == 0 {
		r.lastCityID++
		newID := domain.Uint64ToUUID(r.lastCityID)
		city.SetId(newID)
		r.cities[newID] = *city
		return nil
	}
	if _, ok := r.cities[city.GetId()]; !ok {
		return errors.New("city not found")
	}
	r.cities[city.GetId()] = *city
	return nil
}

func (r *MockCityRepo) FindByID(ctx context.Context, id domain.CityID) (*domain.City, error) {
	r.RLock()
	defer r.RUnlock()
	city, ok := r.cities[id]
	if !ok {
		return nil, errors.New("city not found")
	}
	cpy := city
	return &cpy, nil
}

func (r *MockCityRepo) FindByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.City, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.City
	for _, c := range r.cities {
		if c.GetCountryId() == countryID {
			cpy := c
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockCityRepo) FindByName(ctx context.Context, name string) (*domain.City, error) {
	r.RLock()
	defer r.RUnlock()
	for _, c := range r.cities {
		if c.GetName() == name {
			cpy := c
			return &cpy, nil
		}
	}
	return nil, errors.New("city not found")
}

func (r *MockCityRepo) SearchCities(ctx context.Context, query string, limit, offset int) ([]*domain.City, error) {
	r.RLock()
	defer r.RUnlock()
	for _, c := range r.cities {
		if c.GetName() == query {
			return nil, nil
		}
	}
	return nil, errors.New("city not found")
}

func (r *MockCityRepo) Update(ctx context.Context, city *domain.City) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.cities[city.GetId()]; !ok {
		return errors.New("city not found")
	}
	r.cities[city.GetId()] = *city
	return nil
}

func (r *MockCityRepo) Delete(ctx context.Context, id domain.CityID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.cities[id]; !ok {
		return errors.New("city not found")
	}
	delete(r.cities, id)
	return nil
}

func (r *MockCityRepo) SaveDistrict(ctx context.Context, district *domain.CityDistrict) error {
	r.Lock()
	defer r.Unlock()
	if district.GetId() == uuid.Nil {
		r.lastDistrictID++
		newID := domain.Uint64ToUUID(r.lastDistrictID)
		district.SetId(newID)
		r.districts[newID] = *district
		return nil
	}
	if _, ok := r.districts[district.GetId()]; !ok {
		return errors.New("district not found")
	}
	r.districts[district.GetId()] = *district
	return nil
}

func (r *MockCityRepo) FindDistrictsByCity(ctx context.Context, cityID domain.CityID) ([]*domain.CityDistrict, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.CityDistrict
	for _, d := range r.districts {
		if d.GetCityId() == cityID {
			cpy := d
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockCityRepo) UpdateDistrict(ctx context.Context, district *domain.CityDistrict) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.districts[district.GetId()]; !ok {
		return errors.New("district not found")
	}
	r.districts[district.GetId()] = *district
	return nil
}

func (r *MockCityRepo) DeleteDistrict(ctx context.Context, id domain.CityDistrictID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.districts[id]; !ok {
		return errors.New("district not found")
	}
	delete(r.districts, id)
	return nil
}

func (r *MockCityRepo) SaveTransportNode(ctx context.Context, node *domain.TransportNode) error {
	r.Lock()
	defer r.Unlock()
	if node.GetId() == uuid.Nil {
		r.lastNodeID++
		newID := domain.Uint64ToUUID(r.lastNodeID)
		node.SetId(newID)
		r.nodes[newID] = *node
		return nil
	}
	if _, ok := r.nodes[node.GetId()]; !ok {
		return errors.New("transport node not found")
	}
	r.nodes[node.GetId()] = *node
	return nil
}

func (r *MockCityRepo) FindTransportNodeByID(ctx context.Context, id domain.NodeID) (*domain.TransportNode, error) {
	r.RLock()
	defer r.RUnlock()
	node, ok := r.nodes[id]
	if !ok {
		return nil, errors.New("transport node not found")
	}
	cpy := node
	return &cpy, nil
}

func (r *MockCityRepo) FindTransportNodesByCity(ctx context.Context, cityID domain.CityID) ([]*domain.TransportNode, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.TransportNode
	for _, n := range r.nodes {
		if n.GetCityID() == cityID {
			cpy := n
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockCityRepo) UpdateTransportNode(ctx context.Context, node *domain.TransportNode) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.nodes[node.GetId()]; !ok {
		return errors.New("transport node not found")
	}
	r.nodes[node.GetId()] = *node
	return nil
}

func (r *MockCityRepo) DeleteTransportNode(ctx context.Context, id domain.NodeID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.nodes[id]; !ok {
		return errors.New("transport node not found")
	}
	delete(r.nodes, id)
	return nil
}
