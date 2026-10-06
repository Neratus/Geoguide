package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
)

type RegisterUserUseCase struct {
	userRepo    interfaces.UserRepository
	sessionRepo interfaces.SessionRepository
	notifier    interfaces.Notifier
	taskQueue   interfaces.TaskQueue
	logger      *slog.Logger
	rateLimiter interfaces.RateLimiter
}

func NewRegisterUserUseCase(
	userRepo interfaces.UserRepository,
	sessionRepo interfaces.SessionRepository,
	notifier interfaces.Notifier,
	taskQueue interfaces.TaskQueue,
	logger *slog.Logger,
	rateLimiter interfaces.RateLimiter,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		notifier:    notifier,
		taskQueue:   taskQueue,
		logger:      logger,
		rateLimiter: rateLimiter,
	}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, req requests.RegisterUserRequest) (*requests.RegisterUserResponse, error) {
	uc.logger.Info("registering user", "email", req.Email)

	key := fmt.Sprintf("rate:register:%s", req.Email)
	allowed, err := uc.rateLimiter.Allow(ctx, key, 3, 15*time.Minute)
	if !allowed {
		return nil, usecase_errors.ErrTooManyAttempts
	}

	existing, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, usecase_errors.ErrUserAlreadyExists
	}

	hashed, err := domain.HashPassword(req.Password)
	if err != nil {
		uc.logger.Error("failed to hash password", "error", err)
		return nil, err
	}

	user, err := domain.NewUser(
		uuid.Nil,
		req.Username,
		req.Email,
		hashed,
		req.Phone,
		req.BirthDate,
		req.CountryOfResidence,
		domain.UserRoleUser,
	)
	if err != nil {
		uc.logger.Error("invalid user data", "error", err)
		return nil, err
	}

	if err := uc.userRepo.Save(ctx, user); err != nil {
		uc.logger.Error("failed to save user", "error", err)
		return nil, err
	}

	emailCode := domain.GenerateSixDigitCode()
	fmt.Println(emailCode)
	if err := uc.sessionRepo.Save_code(ctx, user.GetId(), user.GetEmail(), "email_verification", emailCode, 15*time.Minute); err != nil {
		uc.logger.Error("failed to save email code", "error", err)
	} else {
		if err := uc.taskQueue.Publish(ctx, domain.SendVerificationCodeTask{
			UserID:  user.GetId(),
			Contact: user.GetEmail(),
			Purpose: "email_verification",
			Code:    emailCode,
		}); err != nil {
			uc.logger.Warn("failed to publish email verification task", "error", err)
		}
	}

	return &requests.RegisterUserResponse{
		ID:       user.GetId(),
		Username: user.GetUsername(),
		Email:    user.GetEmail(),
		Message:  "verification codes sent",
	}, nil
}
