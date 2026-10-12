package admin

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetPendingReviewsUseCase struct {
	placeRepo interfaces.PlaceRepository
	logger    *slog.Logger
}

func NewGetPendingReviewsUseCase(placeRepo interfaces.PlaceRepository, logger *slog.Logger) *GetPendingReviewsUseCase {
	return &GetPendingReviewsUseCase{placeRepo: placeRepo, logger: logger}
}

func (uc *GetPendingReviewsUseCase) Execute(ctx context.Context, req dto.GetPendingReviewsRequest) (*dto.GetPendingReviewsResponse, error) {
	uc.logger.Info("getting pending reviews", "limit", req.Limit, "offset", req.Offset)

	if !req.IsModerator {
		return nil, usecase_errors.ErrForbidden
	}

	reviews, err := uc.placeRepo.FindPendingReviews(ctx, req.Limit, req.Offset)
	if err != nil {
		uc.logger.Error("failed to get pending reviews", "error", err)
		return nil, err
	}

	return &dto.GetPendingReviewsResponse{Reviews: reviews}, nil
}
