package mocks

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
)

var _ interfaces.TripRepository = (*MockTripRepository)(nil)

type MockTripRepository struct {
	sync.RWMutex
	Trips      map[domain.TripID]*domain.Trip
	Places     map[domain.TripID][]*domain.TripPlace
	lastTripID uint64
}

func NewMockTripRepository() *MockTripRepository {
	return &MockTripRepository{
		Trips:  make(map[domain.TripID]*domain.Trip),
		Places: make(map[domain.TripID][]*domain.TripPlace),
	}
}

func (m *MockTripRepository) Save(ctx context.Context, trip *domain.Trip) error {
	m.Lock()
	defer m.Unlock()

	id := trip.GetId()
	if id == uuid.Nil {
		m.lastTripID++
		newID := domain.Uint64ToUUID(m.lastTripID)
		trip.SetId(newID)
		id = newID
	}

	cpy := *trip
	m.Trips[id] = &cpy
	return nil
}

func (m *MockTripRepository) FindByID(ctx context.Context, id domain.TripID) (*domain.Trip, error) {
	m.RLock()
	defer m.RUnlock()

	trip, ok := m.Trips[id]
	if !ok {
		return nil, errors.New("trip not found")
	}
	cpy := *trip
	return &cpy, nil
}

func (m *MockTripRepository) FindByUser(ctx context.Context, userID domain.UserID) ([]*domain.Trip, error) {
	m.RLock()
	defer m.RUnlock()

	var result []*domain.Trip
	for _, t := range m.Trips {
		if t.GetUserID() == userID {
			cpy := *t
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (m *MockTripRepository) Update(ctx context.Context, trip *domain.Trip) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Trips[trip.GetId()]; !ok {
		return errors.New("trip not found")
	}
	cpy := *trip
	m.Trips[trip.GetId()] = &cpy
	return nil
}

func (m *MockTripRepository) Delete(ctx context.Context, id domain.TripID) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Trips[id]; !ok {
		return errors.New("trip not found")
	}
	delete(m.Trips, id)
	delete(m.Places, id)
	return nil
}

func (m *MockTripRepository) AddPlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID, dayNumber int, arrivalTime *time.Time, durationMin int, notes string) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Trips[tripID]; !ok {
		return errors.New("trip not found")
	}

	tp, _ := domain.NewTripPlace(
		domain.TripPlaceID(uuid.New()),
		dayNumber,
		arrivalTime,
		durationMin,
		notes,
		"PLANNED",
		0.0,
		tripID,
		placeID,
	)

	m.Places[tripID] = append(m.Places[tripID], tp)
	return nil
}

func (m *MockTripRepository) RemovePlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Trips[tripID]; !ok {
		return errors.New("trip not found")
	}

	Places := m.Places[tripID]
	for i, p := range Places {
		if p.GetPlaceID() == placeID {
			m.Places[tripID] = append(Places[:i], Places[i+1:]...)
			return nil
		}
	}
	return errors.New("place not found in trip")
}

func (m *MockTripRepository) GetPlaces(ctx context.Context, tripID domain.TripID) ([]*domain.TripPlace, error) {
	m.RLock()
	defer m.RUnlock()

	if _, ok := m.Trips[tripID]; !ok {
		return nil, errors.New("trip not found")
	}

	Places := m.Places[tripID]
	var result []*domain.TripPlace
	for _, p := range Places {
		cpy := *p
		result = append(result, &cpy)
	}
	return result, nil
}

func (m *MockTripRepository) UpdateTripPlace(ctx context.Context, tripPlace *domain.TripPlace) error {
	m.Lock()
	defer m.Unlock()

	tripID := tripPlace.GetTripID()
	placeID := tripPlace.GetPlaceID()

	Places, ok := m.Places[tripID]
	if !ok {
		return errors.New("trip not found")
	}

	for i, p := range Places {
		if p.GetPlaceID() == placeID {
			cpy := *tripPlace
			m.Places[tripID][i] = &cpy
			return nil
		}
	}
	return errors.New("trip place not found")
}

func (m *MockTripRepository) GetTripStatisticsReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.TripStatData, error) {
	return []domain.TripStatData{}, nil
}
