package interfaces

import (
	"context"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
)

type SessionRepository interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool

	Save_code(ctx context.Context, userID domain.UserID, contact, purpose, code string, ttl time.Duration) error
	Validate_code(ctx context.Context, userID domain.UserID, contact, purpose, code string) (bool, error)
	Delete_code(ctx context.Context, userID domain.UserID, contact, purpose string) error

	CreateTwoFactorChallenge(ctx context.Context, userID domain.UserID) (challengeID string, err error)
	GetUserIDByChallenge(ctx context.Context, challengeID string) (domain.UserID, error)
	DeleteChallenge(ctx context.Context, challengeID string) error
}
