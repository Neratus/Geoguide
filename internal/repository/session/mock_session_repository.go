package session_repo

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type MockSessionRepo struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewMockSessionRepo() (*MockSessionRepo, error) {
	return &MockSessionRepo{
		data: make(map[string]interface{}),
	}, nil
}

func (r *MockSessionRepo) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
	return nil
}

func (r *MockSessionRepo) Get(ctx context.Context, key string, dest interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, ok := r.data[key]
	if !ok {
		return errors.New("key not found")
	}

	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr {
		return errors.New("dest must be a pointer")
	}
	elem := destVal.Elem()
	valVal := reflect.ValueOf(val)
	if !elem.Type().AssignableTo(valVal.Type()) {
		return errors.New("type mismatch")
	}
	elem.Set(valVal)
	return nil
}

func (r *MockSessionRepo) Delete(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}

func (r *MockSessionRepo) Exists(ctx context.Context, key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.data[key]
	return ok
}

func (r *MockSessionRepo) Save_code(ctx context.Context, userID domain.UserID, contact, purpose, code string, ttl time.Duration) error {
	return nil
}

func (r *MockSessionRepo) Validate_code(ctx context.Context, userID domain.UserID, contact, purpose, code string) (bool, error) {
	return false, nil
}

func (r *MockSessionRepo) Delete_code(ctx context.Context, userID domain.UserID, contact, purpose string) error {
	return nil
}

func (r *MockSessionRepo) CreateTwoFactorChallenge(ctx context.Context, userID domain.UserID) (challengeID string, err error) {
	return "", nil
}

func (r *MockSessionRepo) GetUserIDByChallenge(ctx context.Context, challengeID string) (domain.UserID, error) {
	return domain.Uint64ToUUID(0), nil
}

func (r *MockSessionRepo) DeleteChallenge(ctx context.Context, challengeID string) error {
	return nil
}
