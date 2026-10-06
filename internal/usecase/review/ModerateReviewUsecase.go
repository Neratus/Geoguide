package review

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type ModerateReviewUseCase struct {
	placeRepo interfaces.PlaceRepository
	userRepo  interfaces.UserRepository
	notifier  interfaces.Notifier
	taskQueue interfaces.TaskQueue
	logger    *slog.Logger
}

func NewModerateReviewUseCase(
	placeRepo interfaces.PlaceRepository,
	userRepo interfaces.UserRepository,
	notifier interfaces.Notifier,
	taskQueue interfaces.TaskQueue,
	logger *slog.Logger,
) *ModerateReviewUseCase {
	return &ModerateReviewUseCase{
		placeRepo: placeRepo,
		userRepo:  userRepo,
		notifier:  notifier,
		taskQueue: taskQueue,
		logger:    logger,
	}
}

func (uc *ModerateReviewUseCase) Execute(ctx context.Context, req requests.ModerateReviewRequest) (*requests.ModerateReviewResponse, error) {
	uc.logger.Info("moderating review", "review_id", req.ReviewID, "approved", req.Approved)

	review, err := uc.placeRepo.FindReviewByID(ctx, req.ReviewID)
	if err != nil || review == nil {
		return nil, usecase_errors.ErrReviewNotFound
	}
	if review.IsModerated() {
		return nil, usecase_errors.ErrReviewAlreadyModerated
	}

	place, _ := uc.placeRepo.FindByID(ctx, review.GetPlaceID())
	user, _ := uc.userRepo.FindByID(ctx, review.GetUserID())

	if err := uc.placeRepo.ModerateReview(ctx, req.ReviewID, req.Approved, req.Comment); err != nil {
		return nil, err
	}

	if !req.Approved {
		if place != nil && place.GetReviewCnt() > 0 {
			newRating, newCount := domain.RecalculateRatingAfterRemoval(place, review.GetRating())
			if err := uc.placeRepo.UpdateRating(ctx, review.GetPlaceID(), newRating, newCount); err != nil {
				uc.logger.Warn("failed to update place rating after rejection", "error", err)
			}
		}
	}

	userEmail := ""
	username := ""
	if user != nil {
		userEmail = user.GetEmail()
		username = user.GetUsername()
	}
	placeName := ""
	if place != nil {
		placeName = place.GetName()
	}
	_ = uc.taskQueue.Publish(ctx, domain.ReviewModeratedTask{
		ReviewID:  req.ReviewID,
		UserID:    review.GetUserID(),
		UserEmail: userEmail,
		Username:  username,
		PlaceName: placeName,
		Approved:  req.Approved,
		Comment:   req.Comment,
	})

	return &requests.ModerateReviewResponse{
		ID:                req.ReviewID,
		IsApproved:        req.Approved,
		IsModerated:       true,
		ModerationComment: req.Comment,
	}, nil
}
