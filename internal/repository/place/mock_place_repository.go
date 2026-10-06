package place_repo

import (
	"context"
	"errors"
	"sync"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type MockPlaceRepo struct {
	sync.RWMutex
	places       map[domain.PlaceID]domain.Place
	reviews      map[domain.ReviewID]domain.Review
	lastPlaceID  uint64
	lastReviewID uint64
}

func NewMockPlaceRepository() (*MockPlaceRepo, error) {
	return &MockPlaceRepo{
		places:  make(map[domain.PlaceID]domain.Place),
		reviews: make(map[domain.ReviewID]domain.Review),
	}, nil
}

func (r *MockPlaceRepo) GetPopularPlacesReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.PlaceReportData, error) {
	return nil, nil
}

func (r *MockPlaceRepo) Save(ctx context.Context, place *domain.Place) error {
	r.Lock()
	defer r.Unlock()
	if place.GetId() == uuid.Nil {
		r.lastPlaceID++
		newID := domain.Uint64ToUUID(r.lastPlaceID)
		place.SetId(newID)
		r.places[newID] = *place
		return nil
	}
	if _, ok := r.places[place.GetId()]; !ok {
		return errors.New("place not found")
	}
	r.places[place.GetId()] = *place
	return nil
}

func (r *MockPlaceRepo) FindByID(ctx context.Context, id domain.PlaceID) (*domain.Place, error) {
	r.RLock()
	defer r.RUnlock()
	place, ok := r.places[id]
	if !ok {
		return nil, errors.New("place not found")
	}
	cpy := place
	return &cpy, nil
}

func (r *MockPlaceRepo) FindByCity(ctx context.Context, cityID domain.CityID) ([]*domain.Place, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Place
	for _, c := range r.places {
		if c.GetCityID() == cityID {
			cpy := c
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockPlaceRepo) FindByCategory(ctx context.Context, category string) ([]*domain.Place, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Place
	for _, c := range r.places {
		if c.GetCategory() == category {
			cpy := c
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockPlaceRepo) Update(ctx context.Context, place *domain.Place) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.places[place.GetId()]; !ok {
		return errors.New("place not found")
	}
	r.places[place.GetId()] = *place
	return nil
}
func (r *MockPlaceRepo) Delete(ctx context.Context, id domain.PlaceID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.places[id]; !ok {
		return errors.New("place not found")
	}
	delete(r.places, id)
	return nil
}
func (r *MockPlaceRepo) UpdateRating(ctx context.Context, id domain.PlaceID, newAvgRating float64, newReviewsCount int) error {
	r.Lock()
	defer r.Unlock()
	place, ok := r.places[id]
	if !ok {
		return errors.New("place not found")
	}
	place.SetAvgRating(newAvgRating)
	place.SetReviewCnt(int32(newReviewsCount))
	r.places[id] = place
	return nil
}

func (r *MockPlaceRepo) SaveReview(ctx context.Context, review *domain.Review) error {
	r.Lock()
	defer r.Unlock()
	if review.GetId() == uuid.Nil {
		r.lastReviewID++
		newID := domain.Uint64ToUUID(r.lastReviewID)
		review.SetId(newID)
		r.reviews[newID] = *review
		return nil
	}
	if _, ok := r.reviews[review.GetId()]; !ok {
		return errors.New("review not found")
	}
	r.reviews[review.GetId()] = *review
	return nil
}
func (r *MockPlaceRepo) FindReviewByID(ctx context.Context, id domain.ReviewID) (*domain.Review, error) {
	r.RLock()
	defer r.RUnlock()
	for _, c := range r.reviews {
		if c.GetId() == id {
			cpy := c
			return &cpy, nil
		}
	}
	return nil, errors.New("review not found")
}
func (r *MockPlaceRepo) FindReviewsByPlace(ctx context.Context, placeID domain.PlaceID) ([]*domain.Review, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Review
	for _, c := range r.reviews {
		if c.GetPlaceID() == placeID {
			cpy := c
			result = append(result, &cpy)
		}
	}
	return result, nil
}
func (r *MockPlaceRepo) FindReviewsByUser(ctx context.Context, userID domain.UserID) ([]*domain.Review, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Review
	for _, c := range r.reviews {
		if c.GetUserID() == userID {
			cpy := c
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockPlaceRepo) FindReviewsByUserAndPlace(ctx context.Context, userID domain.UserID, placeID domain.PlaceID) ([]*domain.Review, error) {
	r.RLock()
	defer r.RUnlock()
	var result []*domain.Review
	for _, c := range r.reviews {
		if c.GetUserID() == userID && c.GetPlaceID() == placeID {
			cpy := c
			result = append(result, &cpy)
		}
	}
	return result, nil
}

func (r *MockPlaceRepo) UpdateReview(ctx context.Context, review *domain.Review) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.reviews[review.GetId()]; !ok {
		return errors.New("review not found")
	}
	r.reviews[review.GetId()] = *review
	return nil
}
func (r *MockPlaceRepo) DeleteReview(ctx context.Context, id domain.ReviewID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.reviews[id]; !ok {
		return errors.New("review not found")
	}
	delete(r.reviews, id)
	return nil
}
func (r *MockPlaceRepo) ModerateReview(ctx context.Context, id domain.ReviewID, approved bool, moderationComment string) error {
	r.Lock()
	defer r.Unlock()
	review, ok := r.reviews[id]
	if !ok {
		return errors.New("review not found")
	}
	review.SetApproved(approved)
	review.SetModerationComment(moderationComment)
	r.reviews[id] = review
	return nil
}

func (r *MockPlaceRepo) FindPendingReviews(ctx context.Context, limit, offset int) ([]*domain.Review, error) {
	return nil, nil
}
