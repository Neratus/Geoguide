package auth

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type DisableTwoFactorUseCase struct {
	userRepo interfaces.UserRepository
	logger   *slog.Logger
}

func NewDisableTwoFactorUseCase(
	userRepo interfaces.UserRepository,
	logger *slog.Logger,
) *DisableTwoFactorUseCase {
	return &DisableTwoFactorUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (uc *DisableTwoFactorUseCase) Execute(ctx context.Context, userID domain.UserID) error {
	uc.logger.Info("disabling two factor authentication", "user_id", userID)
	err := uc.userRepo.DisableTwoFactor(ctx, userID)
	if err != nil {
		uc.logger.Error("failed to disable two factor authentication", "user_id", userID, "error", err)
		return err
	}
	return nil
}
