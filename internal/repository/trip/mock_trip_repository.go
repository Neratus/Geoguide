package trip_repo

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type MockTripRepo struct {
	sync.RWMutex
	trips      map[domain.TripID]domain.Trip
	tripPlaces map[domain.TripID][]*domain.TripPlace
	lastTripID uint64
}

func NewMockTripRepository() (*MockTripRepo, error) {
	return &MockTripRepo{
		trips:      make(map[domain.TripID]domain.Trip),
		tripPlaces: make(map[domain.TripID][]*domain.TripPlace),
	}, nil
}

func (r *MockTripRepo) Save(ctx context.Context, trip *domain.Trip) error {
	r.Lock()
	defer r.Unlock()
	if trip.GetId() == uuid.Nil {
		r.lastTripID++
		newID := domain.Uint64ToUUID(r.lastTripID)
		trip.SetId(newID)
		r.trips[newID] = *trip
		if _, ok := r.tripPlaces[newID]; !ok {
			r.tripPlaces[newID] = []*domain.TripPlace{}
		}
		return nil
	}
	if _, ok := r.trips[trip.GetId()]; !ok {
		return errors.New("trip not found")
	}
	r.trips[trip.GetId()] = *trip
	return nil
}

func (r *MockTripRepo) FindByID(ctx context.Context, id domain.TripID) (*domain.Trip, error) {
	r.RLock()
	defer r.RUnlock()
	trip, ok := r.trips[id]
	if !ok {
		return nil, errors.New("trip not found")
	}
	cpy := trip
	return &cpy, nil
}

func (r *MockTripRepo) FindByUser(ctx context.Context, userID domain.UserID) ([]*domain.Trip, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Trip
	for _, t := range r.trips {
		if t.GetUserID() == userID {
			cpy := t
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockTripRepo) Update(ctx context.Context, trip *domain.Trip) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.trips[trip.GetId()]; !ok {
		return errors.New("trip not found")
	}
	r.trips[trip.GetId()] = *trip
	return nil
}

func (r *MockTripRepo) Delete(ctx context.Context, id domain.TripID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.trips[id]; !ok {
		return errors.New("trip not found")
	}
	delete(r.trips, id)
	delete(r.tripPlaces, id)
	return nil
}

func (r *MockTripRepo) AddPlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID, dayNumber int, arrivalTime *time.Time, durationMin int, notes string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.trips[tripID]; !ok {
		return errors.New("trip not found")
	}
	tp := domain.NewTripPlaceFromDB(
		uuid.Nil,
		dayNumber,
		arrivalTime,
		durationMin,
		notes,
		domain.TripPlaceStatusPlanned,
		0.0,
		tripID,
		placeID,
	)
	r.tripPlaces[tripID] = append(r.tripPlaces[tripID], tp)
	return nil
}

func (r *MockTripRepo) RemovePlace(ctx context.Context, tripID domain.TripID, placeID domain.PlaceID) error {
	r.Lock()
	defer r.Unlock()
	places, ok := r.tripPlaces[tripID]
	if !ok {
		return errors.New("trip not found")
	}
	for i, tp := range places {
		if tp.GetPlaceID() == placeID {
			r.tripPlaces[tripID] = append(places[:i], places[i+1:]...)
			return nil
		}
	}
	return errors.New("place not found in trip")
}

func (r *MockTripRepo) GetPlaces(ctx context.Context, tripID domain.TripID) ([]*domain.TripPlace, error) {
	r.RLock()
	defer r.RUnlock()
	places, ok := r.tripPlaces[tripID]
	if !ok {
		return []*domain.TripPlace{}, nil
	}
	result := make([]*domain.TripPlace, len(places))
	for i, tp := range places {
		cpy := *tp
		result[i] = &cpy
	}
	return result, nil
}

func (r *MockTripRepo) UpdateTripPlace(ctx context.Context, tripPlace *domain.TripPlace) error {
	r.Lock()
	defer r.Unlock()
	tripID := tripPlace.GetTripID()
	placeID := tripPlace.GetPlaceID()
	places, ok := r.tripPlaces[tripID]
	if !ok {
		return errors.New("trip not found")
	}
	for i, tp := range places {
		if tp.GetPlaceID() == placeID {
			r.tripPlaces[tripID][i] = tripPlace
			return nil
		}
	}
	return errors.New("place not found in trip")
}

func (r *MockTripRepo) GetTripStatisticsReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.TripStatData, error) {
	return nil, nil
}
