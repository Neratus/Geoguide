package user_repo

import (
	"context"
	"errors"
	"sync"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type MockUserRepo struct {
	sync.RWMutex
	users      map[domain.UserID]domain.User
	lastUserID uint64
}

func NewMockUserRepo() (*MockUserRepo, error) {
	return &MockUserRepo{
		users: make(map[domain.UserID]domain.User),
	}, nil
}

func (r *MockUserRepo) Save(ctx context.Context, user *domain.User) error {
	r.Lock()
	defer r.Unlock()
	if user.GetId() == uuid.Nil {
		r.lastUserID++
		newID := domain.Uint64ToUUID(uint64(r.lastUserID))
		user.SetId(newID)
		r.users[newID] = *user
		return nil
	}
	if _, ok := r.users[user.GetId()]; !ok {
		return errors.New("city not found")
	}
	r.users[user.GetId()] = *user
	return nil
}
func (r *MockUserRepo) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	r.RLock()
	defer r.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	cpy := user
	return &cpy, nil
}

func (r *MockUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.RLock()
	defer r.RUnlock()
	for _, u := range r.users {
		if u.GetUsername() == username {
			cpy := u
			return &cpy, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *MockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.RLock()
	defer r.RUnlock()
	for _, u := range r.users {
		if u.GetEmail() == email {
			cpy := u
			return &cpy, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *MockUserRepo) Update(ctx context.Context, user *domain.User) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.users[user.GetId()]; !ok {
		return errors.New("district not found")
	}
	r.users[user.GetId()] = *user
	return nil
}

func (r *MockUserRepo) Delete(ctx context.Context, id domain.UserID) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.users[id]; !ok {
		return errors.New("district not found")
	}
	delete(r.users, id)
	return nil
}

func (r *MockUserRepo) BlockUser(ctx context.Context, id domain.UserID, reason string) error {
	r.Lock()
	defer r.Unlock()
	user, ok := r.users[id]
	if !ok {
		return errors.New("district not found")
	}
	user.Block(reason)
	r.users[id] = user
	return nil
}

func (r *MockUserRepo) UnblockUser(ctx context.Context, id domain.UserID) error {
	r.Lock()
	defer r.Unlock()
	user, ok := r.users[id]
	if !ok {
		return errors.New("district not found")
	}
	user.Unblock()
	r.users[id] = user
	return nil
}

func (r *MockUserRepo) GetUserActivityReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.UserActivityData, error) {
	return nil, nil
}

func (r *MockUserRepo) UpdateEmailVerified(ctx context.Context, userID domain.UserID, verified bool) error {
	return nil
}

func (r *MockUserRepo) UpdatePhoneVerified(ctx context.Context, userID domain.UserID, verified bool) error {
	return nil
}

func (r *MockUserRepo) EnableTwoFactor(ctx context.Context, userID domain.UserID, secret string, backupCodes []string) error {
	return nil
}

func (r *MockUserRepo) DisableTwoFactor(ctx context.Context, userID domain.UserID) error {
	return nil
}

func (r *MockUserRepo) FindAll(ctx context.Context, limit, offset int, search string) ([]*domain.User, error) {
	return nil, nil
}

func (r *MockUserRepo) UpdateAvatar(ctx context.Context, userID domain.UserID, avatarURL string) error {
	return nil
}
