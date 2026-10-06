package review

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type CreateReviewUseCase struct {
	placeRepo interfaces.PlaceRepository
	notifier  interfaces.Notifier
	taskQueue interfaces.TaskQueue
	logger    *slog.Logger
}

func NewCreateReviewUseCase(
	placeRepo interfaces.PlaceRepository,
	notifier interfaces.Notifier,
	taskQueue interfaces.TaskQueue,
	logger *slog.Logger,
) *CreateReviewUseCase {
	return &CreateReviewUseCase{
		placeRepo: placeRepo,
		notifier:  notifier,
		taskQueue: taskQueue,
		logger:    logger,
	}
}

func (uc *CreateReviewUseCase) Execute(ctx context.Context, req requests.CreateReviewRequest) (*requests.CreateReviewResponse, error) {
	uc.logger.Info("creating review", "user_id", req.UserID, "place_id", req.PlaceID, "rating", req.Rating)

	place, err := uc.placeRepo.FindByID(ctx, domain.PlaceID(req.PlaceID))
	if err != nil || place == nil {
		uc.logger.Warn("place not found", "place_id", req.PlaceID)
		return nil, usecase_errors.ErrPlaceNotFound
	}

	existingReviews, err := uc.placeRepo.FindReviewsByUserAndPlace(ctx, req.UserID, req.PlaceID)
	if err != nil {
		uc.logger.Error("failed to check existing reviews", "error", err)
		return nil, err
	}
	if len(existingReviews) > 0 {
		return nil, usecase_errors.ErrReviewAlreadyExists
	}

	review, err := domain.NewReview(
		domain.ReviewID{},
		req.Rating,
		req.Comment,
		req.VisitDate,
		req.UserID,
		req.PlaceID,
		domain.ImageID{},
	)
	if err != nil {
		uc.logger.Error("invalid review data", "error", err)
		return nil, err
	}

	if err := uc.placeRepo.SaveReview(ctx, review); err != nil {
		uc.logger.Error("failed to save review", "error", err)
		return nil, err
	}

	newRating, newCount := domain.CalculateNewAverageRating(place, req.Rating)
	if err := uc.placeRepo.UpdateRating(ctx, req.PlaceID, newRating, newCount); err != nil {
		uc.logger.Warn("failed to update place rating", "error", err)
	}

	_ = uc.taskQueue.Publish(ctx, domain.NewReviewTask{
		ReviewID:  review.GetId(),
		PlaceID:   req.PlaceID,
		PlaceName: place.GetName(),
		UserID:    req.UserID,
		Username:  "",
		Rating:    req.Rating,
		Comment:   req.Comment,
	})

	uc.logger.Info("review created successfully", "review_id", review.GetId())
	return &requests.CreateReviewResponse{
		ID:          review.GetId(),
		Rating:      review.GetRating(),
		Comment:     review.GetComment(),
		VisitDate:   review.GetVisitDate().Format("2006-01-02"),
		CreatedAt:   review.GetCreatedAt().Format("2006-01-02 15:04:05"),
		IsModerated: review.IsModerated(),
		IsApproved:  review.IsApproved(),
	}, nil
}
