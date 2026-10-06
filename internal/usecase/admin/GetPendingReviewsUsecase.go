package admin

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetPendingReviewsUseCase struct {
	placeRepo interfaces.PlaceRepository
	logger    *slog.Logger
}

func NewGetPendingReviewsUseCase(placeRepo interfaces.PlaceRepository, logger *slog.Logger) *GetPendingReviewsUseCase {
	return &GetPendingReviewsUseCase{placeRepo: placeRepo, logger: logger}
}

func (uc *GetPendingReviewsUseCase) Execute(ctx context.Context, req requests.GetPendingReviewsRequest) (*requests.GetPendingReviewsResponse, error) {
	uc.logger.Info("getting pending reviews", "limit", req.Limit, "offset", req.Offset)

	if !req.IsModerator {
		return nil, usecase_errors.ErrForbidden
	}

	reviews, err := uc.placeRepo.FindPendingReviews(ctx, req.Limit, req.Offset)
	if err != nil {
		uc.logger.Error("failed to get pending reviews", "error", err)
		return nil, err
	}

	return &requests.GetPendingReviewsResponse{Reviews: reviews}, nil
}
