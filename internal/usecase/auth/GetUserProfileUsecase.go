package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetUserProfileUseCase struct {
	userRepo    interfaces.UserRepository
	sessionRepo interfaces.SessionRepository
	staticRepo  interfaces.StaticPageRepository
	logger      *slog.Logger
}

func NewGetUserProfileUseCase(
	userRepo interfaces.UserRepository,
	sessionRepo interfaces.SessionRepository,
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		staticRepo:  staticRepo,
		logger:      logger,
	}
}

func (uc *GetUserProfileUseCase) Execute(ctx context.Context, req requests.GetUserProfileRequest) (*requests.GetUserProfileResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		uc.logger.Warn("user not found", "user_id", req.UserID)
		return nil, usecase_errors.ErrUserNotFound
	}
	uc.logger.Info("user found", "user_id", user.GetId())
	avatarURL := ""
	if user.GetAvatarUrl() != "" {
		url, err := uc.staticRepo.GetFileURL(ctx, user.GetAvatarUrl())
		if err == nil {
			avatarURL = url
		}
	}
	var birthDate string
	if user.GetBirthDate() != nil {
		birthDate = user.GetBirthDate().Format("2006-01-02")
	}
	return &requests.GetUserProfileResponse{
		ID:                 user.GetId(),
		Username:           user.GetUsername(),
		Email:              user.GetEmail(),
		Phone:              user.GetPhone(),
		RegisteredAt:       user.GetRegisteredAt().Format(time.RFC3339),
		BirthDate:          birthDate,
		CountryOfResidence: user.GetCountryOfResidence(),
		AvatarURL:          avatarURL,
		FavoriteCategories: user.GetFavoriteCategories(),
		IsBlocked:          user.IsBlocked(),
		Role:               user.GetRole(),
	}, nil
}
