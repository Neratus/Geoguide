package review

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetReviewsByPlaceUseCase struct {
	placeRepo interfaces.PlaceRepository
	logger    *slog.Logger
}

func NewGetReviewsByPlaceUseCase(
	placeRepo interfaces.PlaceRepository,
	logger *slog.Logger,
) *GetReviewsByPlaceUseCase {
	return &GetReviewsByPlaceUseCase{
		placeRepo: placeRepo,
		logger:    logger,
	}
}

func (uc *GetReviewsByPlaceUseCase) Execute(ctx context.Context, req requests.GetReviewsByPlaceRequest) ([]requests.ReviewByPlaceResponse, error) {
	uc.logger.Info("getting reviews by place", "place_id", req.PlaceID)

	place, err := uc.placeRepo.FindByID(ctx, req.PlaceID)
	if err != nil || place == nil {
		return nil, usecase_errors.ErrPlaceNotFound
	}

	reviews, err := uc.placeRepo.FindReviewsByPlace(ctx, req.PlaceID)
	if err != nil {
		uc.logger.Error("failed to get reviews", "error", err)
		return nil, err
	}

	filtered := reviews
	if req.OnlyApproved {
		filtered = make([]*domain.Review, 0, len(reviews))
		for _, r := range reviews {
			if r.IsApproved() {
				filtered = append(filtered, r)
			}
		}
	}

	resp := make([]requests.ReviewByPlaceResponse, 0, len(filtered))
	for _, r := range filtered {
		resp = append(resp, requests.ReviewByPlaceResponse{
			ID:                r.GetId(),
			Rating:            r.GetRating(),
			Comment:           r.GetComment(),
			VisitDate:         r.GetVisitDate().Format("2006-01-02"),
			CreatedAt:         r.GetCreatedAt().Format("2006-01-02 15:04:05"),
			Username:          r.GetUsername(),
			UserAvatar:        r.GetUserAvatar(),
			IsApproved:        r.IsApproved(),
			ModerationComment: r.GetModerationComment(),
		})
	}
	return resp, nil
}
