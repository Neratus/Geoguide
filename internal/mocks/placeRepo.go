package mocks

import (
	"context"
	"errors"
	"sync"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type MockPlaceRepository struct {
	sync.RWMutex
	Places  map[domain.PlaceID]*domain.Place
	Reviews map[domain.ReviewID]*domain.Review

	ForceError struct {
		FindByCategory error
		FindByCity     error
		FindByID       error
		Save           error
	}
}

func NewMockPlaceRepository() *MockPlaceRepository {
	return &MockPlaceRepository{
		Places:  make(map[domain.PlaceID]*domain.Place),
		Reviews: make(map[domain.ReviewID]*domain.Review),
	}
}

func (m *MockPlaceRepository) Save(ctx context.Context, place *domain.Place) error {
	if m.ForceError.Save != nil {
		return m.ForceError.Save
	}
	m.Lock()
	defer m.Unlock()
	if place.GetId() == (domain.PlaceID{}) {
		place.SetId(domain.PlaceID(uuid.New()))
	}
	cpy := *place
	m.Places[place.GetId()] = &cpy
	return nil
}

func (m *MockPlaceRepository) FindByID(ctx context.Context, id domain.PlaceID) (*domain.Place, error) {
	if m.ForceError.FindByID != nil {
		return nil, m.ForceError.FindByID
	}
	m.RLock()
	defer m.RUnlock()
	place, ok := m.Places[id]
	if !ok {
		return nil, errors.New("place not found")
	}
	cpy := *place
	return &cpy, nil
}

func (m *MockPlaceRepository) FindByCity(ctx context.Context, cityID domain.CityID) ([]*domain.Place, error) {
	if m.ForceError.FindByCity != nil {
		return nil, m.ForceError.FindByCity
	}
	m.RLock()
	defer m.RUnlock()
	var result []*domain.Place
	for _, p := range m.Places {
		if p.GetCityID() == cityID {
			cpy := *p
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (m *MockPlaceRepository) FindByCategory(ctx context.Context, category string) ([]*domain.Place, error) {
	if m.ForceError.FindByCategory != nil {
		return nil, m.ForceError.FindByCategory
	}
	m.RLock()
	defer m.RUnlock()
	var result []*domain.Place
	for _, p := range m.Places {
		if p.GetCategory() == category {
			cpy := *p
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (m *MockPlaceRepository) Update(ctx context.Context, place *domain.Place) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.Places[place.GetId()]; !ok {
		return errors.New("place not found")
	}
	m.Places[place.GetId()] = place
	return nil
}

func (m *MockPlaceRepository) Delete(ctx context.Context, id domain.PlaceID) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.Places[id]; !ok {
		return errors.New("place not found")
	}
	delete(m.Places, id)
	return nil
}

func (m *MockPlaceRepository) UpdateRating(ctx context.Context, id domain.PlaceID, newAvgRating float64, newReviewsCount int) error {
	m.Lock()
	defer m.Unlock()

	return nil
}

func (m *MockPlaceRepository) SaveReview(ctx context.Context, review *domain.Review) error {
	m.Lock()
	defer m.Unlock()
	if review.GetId() == uuid.Nil {
		review.SetId(domain.ReviewID(uuid.New()))
	}
	m.Reviews[review.GetId()] = review
	return nil
}

func (m *MockPlaceRepository) FindReviewByID(ctx context.Context, id domain.ReviewID) (*domain.Review, error) {
	m.RLock()
	defer m.RUnlock()
	review, ok := m.Reviews[id]
	if !ok {
		return nil, errors.New("review not found")
	}
	return review, nil
}

func (m *MockPlaceRepository) FindReviewsByPlace(ctx context.Context, placeID domain.PlaceID) ([]*domain.Review, error) {
	m.RLock()
	defer m.RUnlock()
	var result []*domain.Review
	for _, r := range m.Reviews {
		if r.GetPlaceID() == placeID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockPlaceRepository) FindReviewsByUserAndPlace(ctx context.Context, userID domain.UserID, placeID domain.PlaceID) ([]*domain.Review, error) {
	m.RLock()
	defer m.RUnlock()
	var result []*domain.Review
	for _, r := range m.Reviews {
		if r.GetUserID() == userID && r.GetPlaceID() == placeID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockPlaceRepository) FindReviewsByUser(ctx context.Context, userID domain.UserID) ([]*domain.Review, error) {
	m.RLock()
	defer m.RUnlock()
	var result []*domain.Review
	for _, r := range m.Reviews {
		if r.GetUserID() == userID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockPlaceRepository) UpdateReview(ctx context.Context, review *domain.Review) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.Reviews[review.GetId()]; !ok {
		return errors.New("review not found")
	}
	m.Reviews[review.GetId()] = review
	return nil
}

func (m *MockPlaceRepository) DeleteReview(ctx context.Context, id domain.ReviewID) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.Reviews[id]; !ok {
		return errors.New("review not found")
	}
	delete(m.Reviews, id)
	return nil
}

func (m *MockPlaceRepository) ModerateReview(ctx context.Context, id domain.ReviewID, approved bool, comment string) error {
	m.Lock()
	defer m.Unlock()
	review, ok := m.Reviews[id]
	if !ok {
		return errors.New("review not found")
	}
	review.SetModerated(true)
	review.SetApproved(approved)
	review.SetModerationComment(comment)
	return nil
}

func (m *MockPlaceRepository) GetPopularPlacesReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.PlaceReportData, error) {
	return []domain.PlaceReportData{}, nil
}

func (m *MockPlaceRepository) FindPendingReviews(ctx context.Context, limit, offset int) ([]*domain.Review, error) {
	return []*domain.Review{}, nil
}
