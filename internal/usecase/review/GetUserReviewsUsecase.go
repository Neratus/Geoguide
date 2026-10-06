package review

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetUserReviewsUseCase struct {
	placeRepo interfaces.PlaceRepository
	logger    *slog.Logger
}

func NewGetUserReviewsUseCase(
	placeRepo interfaces.PlaceRepository,
	logger *slog.Logger,
) *GetUserReviewsUseCase {
	return &GetUserReviewsUseCase{
		placeRepo: placeRepo,
		logger:    logger,
	}
}

func (uc *GetUserReviewsUseCase) Execute(ctx context.Context, req requests.GetUserReviewsRequest) ([]requests.UserReviewResponse, error) {
	uc.logger.Info("getting user reviews", "user_id", req.UserID)

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	reviews, err := uc.placeRepo.FindReviewsByUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if len(reviews) == 0 {
		return nil, usecase_errors.ErrReviewsNotFound
	}

	resp := make([]requests.UserReviewResponse, 0, len(reviews))
	for _, r := range reviews {
		resp = append(resp, requests.UserReviewResponse{
			ID:          r.GetId(),
			Rating:      r.GetRating(),
			Comment:     r.GetComment(),
			VisitDate:   r.GetVisitDate().Format("2006-01-02"),
			CreatedAt:   r.GetCreatedAt().Format("2006-01-02 15:04:05"),
			PlaceID:     r.GetPlaceID(),
			PlaceName:   r.GetPlaceName(),
			IsApproved:  r.IsApproved(),
			IsModerated: r.IsModerated(),
		})
	}
	return resp, nil
}
