package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/pquerna/otp/totp"
)

type EnableTwoFactorUseCase struct {
	userRepo interfaces.UserRepository
	logger   *slog.Logger
}

func NewEnableTwoFactorUseCase(
	userRepo interfaces.UserRepository,
	logger *slog.Logger,
) *EnableTwoFactorUseCase {
	return &EnableTwoFactorUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (uc *EnableTwoFactorUseCase) GenerateSecret(ctx context.Context, userID domain.UserID, email string) (secret string, otpauthURI string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GeoGuide",
		AccountName: email,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	return key.Secret(), key.URL(), nil
}

func (uc *EnableTwoFactorUseCase) VerifyAndEnable(ctx context.Context, req requests.EnableTwoFactorRequest) (backupCodes []string, err error) {
	uc.logger.Info("verifying and enabling two factor authentication", "user_id", req.UserID)
	if !totp.Validate(req.Code, req.Secret) {
		return nil, usecase_errors.ErrInvalidTwoFactorCode
	}

	backupCodes = generateBackupCodes(10)
	hashedBackups := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedBackups[i], _ = domain.HashPassword(code)
	}

	err = uc.userRepo.EnableTwoFactor(ctx, req.UserID, req.Secret, hashedBackups)
	if err != nil {
		return nil, fmt.Errorf("failed to enable two factor: %w", err)
	}

	return backupCodes, nil
}

func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = generateRandomCode(8)
	}
	return codes
}

func generateRandomCode(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}
