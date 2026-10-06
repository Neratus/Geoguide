package interfaces

import (
	"context"

	"github.com/Neratus/geoguide/internal/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id domain.UserID) error
	BlockUser(ctx context.Context, id domain.UserID, reason string) error
	UnblockUser(ctx context.Context, id domain.UserID) error
	FindAll(ctx context.Context, limit, offset int, search string) ([]*domain.User, error)

	UpdateEmailVerified(ctx context.Context, userID domain.UserID, verified bool) error
	UpdatePhoneVerified(ctx context.Context, userID domain.UserID, verified bool) error
	EnableTwoFactor(ctx context.Context, userID domain.UserID, secret string, backupCodes []string) error
	DisableTwoFactor(ctx context.Context, userID domain.UserID) error

	GetUserActivityReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.UserActivityData, error)
	UpdateAvatar(ctx context.Context, userID domain.UserID, avatarURL string) error
}
