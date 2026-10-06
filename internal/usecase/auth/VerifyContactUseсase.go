package auth

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type VerifyContactUseCase struct {
	userRepo    interfaces.UserRepository
	sessionRepo interfaces.SessionRepository
	notifier    interfaces.Notifier
	taskQueue   interfaces.TaskQueue
	logger      *slog.Logger
}

func NewVerifyContactUseCase(
	userRepo interfaces.UserRepository,
	sessionRepo interfaces.SessionRepository,
	notifier interfaces.Notifier,
	taskQueue interfaces.TaskQueue,
	logger *slog.Logger,
) *VerifyContactUseCase {
	return &VerifyContactUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		notifier:    notifier,
		taskQueue:   taskQueue,
		logger:      logger,
	}
}

func (uc *VerifyContactUseCase) Execute(ctx context.Context, req requests.VerifyContactRequest) error {
	uc.logger.Error("verifying contact", "user_id", req.UserID, "contact", req.Contact)
	var purpose string
	if strings.Contains(req.Contact, "@") {
		purpose = "email_verification"
	} else {
		purpose = "phone_verification"
	}

	ok, err := uc.sessionRepo.Validate_code(ctx, req.UserID, req.Contact, purpose, req.Code)
	if err != nil || !ok {
		return usecase_errors.ErrInvalidVerificationCode
	}

	if purpose == "email_verification" {
		err = uc.userRepo.UpdateEmailVerified(ctx, req.UserID, true)
	} else {
		err = uc.userRepo.UpdatePhoneVerified(ctx, req.UserID, true)
	}
	if err != nil {
		return err
	}
	err = uc.sessionRepo.Delete_code(ctx, req.UserID, req.Contact, purpose)
	if err != nil {
		return err
	}
	return nil
}
