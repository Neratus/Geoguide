package auth

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type LogoutUseCase struct {
	sessionRepo interfaces.SessionRepository
	logger      *slog.Logger
}

func NewLogoutUseCase(
	sessionRepo interfaces.SessionRepository,
	logger *slog.Logger,
) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, req requests.LogoutRequest) (*requests.LogoutResponse, error) {
	uc.logger.Info("logging out user", "token", req.Token)

	if !uc.sessionRepo.Exists(ctx, req.Token) {
		uc.logger.Warn("session not found", "token", req.Token)
		return nil, usecase_errors.ErrInvalidSession
	}

	if err := uc.sessionRepo.Delete(ctx, req.Token); err != nil {
		uc.logger.Error("failed to delete session", "error", err)
		return nil, err
	}

	uc.logger.Info("user logged out successfully", "token", req.Token)

	return &requests.LogoutResponse{
		Success: true,
	}, nil
}
