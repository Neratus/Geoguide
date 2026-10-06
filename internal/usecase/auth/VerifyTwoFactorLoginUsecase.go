package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/services"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type VerifyTwoFactorLoginUseCase struct {
	userRepo    interfaces.UserRepository
	sessionRepo interfaces.SessionRepository
	logger      *slog.Logger
}

func NewVerifyTwoFactorLoginUseCase(
	userRepo interfaces.UserRepository,
	sessionRepo interfaces.SessionRepository,
	logger *slog.Logger,
) *VerifyTwoFactorLoginUseCase {
	return &VerifyTwoFactorLoginUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (uc *VerifyTwoFactorLoginUseCase) Execute(ctx context.Context, req requests.VerifyTwoFactorRequest) (string, error) {
	uc.logger.Info("verifying two factor login", "challenge_id", req.ChallengeID)
	userID, err := uc.sessionRepo.GetUserIDByChallenge(ctx, req.ChallengeID)
	if err != nil {
		return "", usecase_errors.ErrInvalidTwoFactorCode
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}

	valid := false
	if totp.Validate(req.Code, user.GetTwoFactorSecret()) {
		valid = true
	} else if isBackupCodeValid(user.GetBackupCodes(), req.Code) {
		valid = true
		newBackups := removeUsedBackupCode(user.GetBackupCodes(), req.Code)
		uc.userRepo.EnableTwoFactor(ctx, userID, user.GetTwoFactorSecret(), newBackups)
	}

	if !valid {
		return "", usecase_errors.ErrInvalidTwoFactorCode
	}

	token, err := services.GenerateJWT(userID, user.GetRole())
	if err != nil {
		uc.logger.Error("failed to generate JWT", "error", err)
		return "", err
	}
	if err := uc.sessionRepo.Set(ctx, token, userID, 24*time.Hour); err != nil {
		return "", err
	}

	uc.sessionRepo.DeleteChallenge(ctx, req.ChallengeID)

	return token, nil
}

func isBackupCodeValid(hashedBackups []string, code string) bool {
	for _, hc := range hashedBackups {
		if err := bcrypt.CompareHashAndPassword([]byte(hc), []byte(code)); err == nil {
			return true
		}
	}
	return false
}

func removeUsedBackupCode(backups []string, usedCode string) []string {
	newList := []string{}
	for _, bc := range backups {
		if bc != usedCode {
			newList = append(newList, bc)
		}
	}
	return newList
}
