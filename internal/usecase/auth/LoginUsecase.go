package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/services"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type LoginUseCase struct {
	userRepo    interfaces.UserRepository
	sessionRepo interfaces.SessionRepository
	notifier    interfaces.Notifier
	taskQueue   interfaces.TaskQueue
	logger      *slog.Logger
	rateLimiter interfaces.RateLimiter
}

func NewLoginUseCase(
	userRepo interfaces.UserRepository,
	sessionRepo interfaces.SessionRepository,
	notifier interfaces.Notifier,
	taskQueue interfaces.TaskQueue,
	logger *slog.Logger,
	rateLimiter interfaces.RateLimiter,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		notifier:    notifier,
		taskQueue:   taskQueue,
		logger:      logger,
		rateLimiter: rateLimiter,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req requests.LoginUserRequest, ip, userAgent string) (*requests.LoginUserResponse, error) {
	uc.logger.Info("login user", "username", req.Username, "ip", ip)

	key := fmt.Sprintf("rate:register:%s", req.Username)
	allowed, err := uc.rateLimiter.Allow(ctx, key, 3, 15*time.Minute)
	if !allowed {
		return nil, usecase_errors.ErrTooManyAttempts
	}

	user, err := uc.userRepo.FindByUsername(ctx, req.Username)
	if err != nil || user == nil {
		uc.logger.Warn("user not found", "username", req.Username)
		return nil, usecase_errors.ErrInvalidCredentials
	}

	if user.IsBlocked() {
		uc.logger.Warn("user is blocked", "username", req.Username)
		return nil, usecase_errors.ErrUserBlocked
	}

	if err := domain.ComparePassword(user.GetPasswordHash(), req.Password); err != nil {
		uc.logger.Warn("invalid password", "username", req.Username)
		return nil, usecase_errors.ErrInvalidCredentials
	}
	if !user.IsEmailVerified() {
		return nil, usecase_errors.ErrEmailNotVerified
	}

	if user.IsTwoFactorEnabled() {
		challengeID, err := uc.sessionRepo.CreateTwoFactorChallenge(ctx, user.GetId())
		if err != nil {
			uc.logger.Error("failed to create 2fa challenge", "error", err)
			return nil, err
		}
		return &requests.LoginUserResponse{
			RequiresTwoFactor: true,
			ChallengeID:       challengeID,
		}, nil
	}

	token, err := services.GenerateJWT(user.GetId(), user.GetRole())
	if err != nil {
		return nil, err
	}
	if err := uc.sessionRepo.Set(ctx, token, user.GetId(), 24*time.Hour); err != nil {
		uc.logger.Error("failed to create session", "error", err)
		return nil, err
	}

	if err := uc.taskQueue.Publish(ctx, domain.LoginNotificationTask{
		UserID:    user.GetId(),
		Email:     user.GetEmail(),
		Username:  user.GetUsername(),
		IP:        ip,
		UserAgent: userAgent,
	}); err != nil {
		uc.logger.Warn("failed to publish login notification task", "error", err)
	}

	uc.logger.Info("user logged in successfully", "user_id", user.GetId())

	return &requests.LoginUserResponse{
		ID:       user.GetId(),
		Username: user.GetUsername(),
		Email:    user.GetEmail(),
		Token:    token,
	}, nil
}
