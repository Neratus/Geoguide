package admin

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetUsersUseCase struct {
	userRepo interfaces.UserRepository
	logger   *slog.Logger
}

func NewGetUsersUseCase(userRepo interfaces.UserRepository, logger *slog.Logger) *GetUsersUseCase {
	return &GetUsersUseCase{userRepo: userRepo, logger: logger}
}

func (uc *GetUsersUseCase) Execute(ctx context.Context, req dto.GetUsersRequest) (*dto.GetUsersResponse, error) {
	uc.logger.Info("getting users", "limit", req.Limit, "offset", req.Offset, "search", req.Search)

	if !req.IsModerator {
		return nil, usecase_errors.ErrForbidden
	}

	users, err := uc.userRepo.FindAll(ctx, req.Limit, req.Offset, req.Search)
	if err != nil {
		uc.logger.Error("failed to get users", "error", err)
		return nil, err
	}

	return &dto.GetUsersResponse{Users: users}, nil
}
