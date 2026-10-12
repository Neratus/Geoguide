package review

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type DeleteReviewUseCase struct {
	placeRepo interfaces.PlaceRepository
	logger    *slog.Logger
}

func NewDeleteReviewUseCase(
	placeRepo interfaces.PlaceRepository,
	logger *slog.Logger,
) *DeleteReviewUseCase {
	return &DeleteReviewUseCase{
		placeRepo: placeRepo,
		logger:    logger,
	}
}

func (uc *DeleteReviewUseCase) Execute(ctx context.Context, req dto.DeleteReviewRequest) (*dto.DeleteReviewResponse, error) {
	uc.logger.Info("deleting review", "review_id", req.ReviewID)

	review, err := uc.placeRepo.FindReviewByID(ctx, req.ReviewID)
	if err != nil || review == nil {
		return nil, usecase_errors.ErrReviewNotFound
	}
	if !req.IsModerator && review.GetUserID() != req.UserID {
		return nil, usecase_errors.ErrNotAuthorized
	}
	if err := uc.placeRepo.DeleteReview(ctx, req.ReviewID); err != nil {
		return nil, err
	}
	place, _ := uc.placeRepo.FindByID(ctx, review.GetPlaceID())
	if place != nil && place.GetReviewCnt() > 0 {
		newRating, newCount := domain.RecalculateRatingAfterRemoval(place, review.GetRating())
		if err := uc.placeRepo.UpdateRating(ctx, review.GetPlaceID(), newRating, newCount); err != nil {
			uc.logger.Warn("failed to update place rating after deletion", "error", err)
		}
	}
	return &dto.DeleteReviewResponse{Success: true}, nil
}
