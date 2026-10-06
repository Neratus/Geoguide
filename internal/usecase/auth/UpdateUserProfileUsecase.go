package auth

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
)

type UpdateUserProfileUseCase struct {
	userRepo    interfaces.UserRepository
	sessionRepo interfaces.SessionRepository
	logger      *slog.Logger
}

func NewUpdateUserProfileUseCase(
	userRepo interfaces.UserRepository,
	sessionRepo interfaces.SessionRepository,
	logger *slog.Logger,
) *UpdateUserProfileUseCase {
	return &UpdateUserProfileUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (uc *UpdateUserProfileUseCase) Execute(ctx context.Context, req requests.UpdateUserProfileRequest) (*requests.UpdateUserProfileResponse, error) {
	uc.logger.Info("updating user profile", "token", req.Token)

	var userID uuid.UUID
	if err := uc.sessionRepo.Get(ctx, req.Token, &userID); err != nil {
		uc.logger.Warn("invalid or expired session", "error", err)
		return nil, usecase_errors.ErrInvalidSession
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		uc.logger.Warn("user not found", "user_id", userID)
		return nil, usecase_errors.ErrUserNotFound
	}

	if user.IsBlocked() {
		uc.logger.Warn("user is blocked", "user_id", userID)
		return nil, usecase_errors.ErrUserBlocked
	}

	if req.Username != nil && *req.Username != "" {
		if *req.Username != user.GetUsername() {
			existingUser, _ := uc.userRepo.FindByUsername(ctx, *req.Username)
			if existingUser != nil && existingUser.GetId() != userID {
				return nil, usecase_errors.ErrUsernameAlreadyExists
			}
		}
		user.SetUsername(*req.Username)
	}

	if req.Phone != nil {
		user.SetPhone(*req.Phone)
	}

	if req.BirthDate != nil {
		user.SetBirthDate(req.BirthDate)
	}

	if req.CountryOfResidence != nil {
		user.SetCountryOfResidence(*req.CountryOfResidence)
	}

	if req.AvatarURL != nil {
		user.SetAvatarUrl(*req.AvatarURL)
	}

	if req.FavoriteCategories != nil {
		user.SetFavoriteCategories(req.FavoriteCategories)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.Error("failed to update user", "error", err, "user_id", userID)
		return nil, err
	}

	uc.logger.Info("user profile updated successfully", "user_id", userID)

	return &requests.UpdateUserProfileResponse{
		ID:                 user.GetId(),
		Username:           user.GetUsername(),
		Email:              user.GetEmail(),
		Phone:              user.GetPhone(),
		RegisteredAt:       user.GetRegisteredAt().Format("2006-01-02 15:04:05"),
		BirthDate:          user.GetBirthDate().Format("2006-01-02"),
		CountryOfResidence: user.GetCountryOfResidence(),
		AvatarURL:          user.GetAvatarUrl(),
		FavoriteCategories: user.GetFavoriteCategories(),
		IsBlocked:          user.IsBlocked(),
	}, nil
}
