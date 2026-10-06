package mocks

import (
	"context"
	"errors"
	"sync"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type MockUserRepository struct {
	sync.RWMutex
	Users map[domain.UserID]*domain.User

	Calls struct {
		Save                  int
		FindByID              int
		FindByUsername        int
		FindByEmail           int
		Update                int
		Delete                int
		BlockUser             int
		UnblockUser           int
		FindAll               int
		UpdateEmailVerified   int
		UpdatePhoneVerified   int
		EnableTwoFactor       int
		DisableTwoFactor      int
		GetUserActivityReport int
		UpdateAvatar          int
	}
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[domain.UserID]*domain.User),
	}
}

func (m *MockUserRepository) Save(ctx context.Context, user *domain.User) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Save++

	if user.GetId() == uuid.Nil {
		user.SetId(domain.UserID(uuid.New()))
	}
	cpy := *user
	m.Users[user.GetId()] = &cpy
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindByID++

	user, ok := m.Users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	cpy := *user
	return &cpy, nil
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindByUsername++

	for _, user := range m.Users {
		if user.GetUsername() == username {
			cpy := *user
			return &cpy, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindByEmail++

	for _, user := range m.Users {
		if user.GetEmail() == email {
			cpy := *user
			return &cpy, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Update++

	if _, ok := m.Users[user.GetId()]; !ok {
		return errors.New("user not found")
	}
	cpy := *user
	m.Users[user.GetId()] = &cpy
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id domain.UserID) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.Delete++

	if _, ok := m.Users[id]; !ok {
		return errors.New("user not found")
	}
	delete(m.Users, id)
	return nil
}

func (m *MockUserRepository) BlockUser(ctx context.Context, id domain.UserID, reason string) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.BlockUser++

	user, ok := m.Users[id]
	if !ok {
		return errors.New("user not found")
	}
	user.Block(reason)
	return nil
}

func (m *MockUserRepository) UnblockUser(ctx context.Context, id domain.UserID) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.UnblockUser++

	user, ok := m.Users[id]
	if !ok {
		return errors.New("user not found")
	}
	user.Unblock()
	return nil
}

func (m *MockUserRepository) FindAll(ctx context.Context, limit, offset int, search string) ([]*domain.User, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.FindAll++

	var result []*domain.User
	total := 0

	for _, user := range m.Users {
		if search == "" || containsIgnoreCase(user.GetUsername(), search) || containsIgnoreCase(user.GetEmail(), search) {
			total++
			if len(result) < limit {
				cpy := *user
				result = append(result, &cpy)
			}
		}
	}
	return result, nil
}

func (m *MockUserRepository) UpdateEmailVerified(ctx context.Context, userID domain.UserID, verified bool) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.UpdateEmailVerified++

	user, ok := m.Users[userID]
	if !ok {
		return errors.New("user not found")
	}
	user.SetEmailVerified(verified)
	return nil
}

func (m *MockUserRepository) UpdatePhoneVerified(ctx context.Context, userID domain.UserID, verified bool) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.UpdatePhoneVerified++

	user, ok := m.Users[userID]
	if !ok {
		return errors.New("user not found")
	}
	user.SetPhoneVerified(verified)
	return nil
}

func (m *MockUserRepository) EnableTwoFactor(ctx context.Context, userID domain.UserID, secret string, backupCodes []string) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.EnableTwoFactor++

	_, ok := m.Users[userID]
	if !ok {
		return errors.New("user not found")
	}

	return nil
}

func (m *MockUserRepository) DisableTwoFactor(ctx context.Context, userID domain.UserID) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.DisableTwoFactor++

	_, ok := m.Users[userID]
	if !ok {
		return errors.New("user not found")
	}

	return nil
}

func (m *MockUserRepository) GetUserActivityReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.UserActivityData, error) {
	m.RLock()
	defer m.RUnlock()
	m.Calls.GetUserActivityReport++

	return []domain.UserActivityData{}, nil
}

func (m *MockUserRepository) UpdateAvatar(ctx context.Context, userID domain.UserID, avatarURL string) error {
	m.Lock()
	defer m.Unlock()
	m.Calls.UpdateAvatar++

	user, ok := m.Users[userID]
	if !ok {
		return errors.New("user not found")
	}
	user.SetAvatarUrl(avatarURL)
	return nil
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}
