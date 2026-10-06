package auth

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type ModerateUserUsecase struct {
	userRepo  interfaces.UserRepository
	notifier  interfaces.Notifier
	taskQueue interfaces.TaskQueue
	logger    *slog.Logger
}

func NewModerateUserUsecase(
	userRepo interfaces.UserRepository,
	notifier interfaces.Notifier,
	taskQueue interfaces.TaskQueue,
	logger *slog.Logger,
) *ModerateUserUsecase {
	return &ModerateUserUsecase{
		userRepo:  userRepo,
		notifier:  notifier,
		taskQueue: taskQueue,
		logger:    logger,
	}
}

func (uc *ModerateUserUsecase) Execute(ctx context.Context, req requests.ModerateUserRequest) (*requests.ModerateUserResponse, error) {
	uc.logger.Info("moderating user", "user_id", req.UserID, "block", req.Block, "moderator_id", req.ModeratorID)

	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil || user == nil {
		uc.logger.Warn("user not found", "user_id", req.UserID)
		return nil, usecase_errors.ErrUserNotFound
	}

	if req.ModeratorID == req.UserID {
		uc.logger.Warn("moderator tried to block themselves", "user_id", req.UserID)
		return nil, usecase_errors.ErrCannotModerateSelf
	}

	if req.Block {
		if req.Reason == "" {
			return nil, usecase_errors.ErrBlockReasonRequired
		}

		if user.IsBlocked() {
			uc.logger.Warn("user already blocked", "user_id", req.UserID)
			return nil, usecase_errors.ErrUserAlreadyBlocked
		}

		if err := uc.userRepo.BlockUser(ctx, req.UserID, req.Reason); err != nil {
			uc.logger.Error("failed to block user", "error", err, "user_id", req.UserID)
			return nil, err
		}

		if err := uc.taskQueue.Publish(ctx, domain.UserBlockedTask{
			UserID:      user.GetId(),
			Email:       user.GetEmail(),
			Username:    user.GetUsername(),
			Reason:      req.Reason,
			ModeratorID: req.ModeratorID,
		}); err != nil {
			uc.logger.Warn("failed to publish user blocked task", "error", err)
		}

		uc.logger.Info("user blocked successfully", "user_id", req.UserID, "reason", req.Reason)

		user, _ = uc.userRepo.FindByID(ctx, req.UserID)

		return &requests.ModerateUserResponse{
			UserID:      user.GetId(),
			IsBlocked:   true,
			BlockedAt:   user.GetBlockedAt(),
			BlockReason: user.GetBlockReason(),
		}, nil
	}

	if !user.IsBlocked() {
		uc.logger.Warn("user not blocked", "user_id", req.UserID)
		return nil, usecase_errors.ErrUserNotBlocked
	}

	if err := uc.userRepo.UnblockUser(ctx, req.UserID); err != nil {
		uc.logger.Error("failed to unblock user", "error", err, "user_id", req.UserID)
		return nil, err
	}

	if err := uc.taskQueue.Publish(ctx, domain.UserUnblockedTask{
		UserID:      user.GetId(),
		Email:       user.GetEmail(),
		Username:    user.GetUsername(),
		ModeratorID: req.ModeratorID,
	}); err != nil {
		uc.logger.Warn("failed to publish user unblocked task", "error", err)
	}

	uc.logger.Info("user unblocked successfully", "user_id", req.UserID)

	user, _ = uc.userRepo.FindByID(ctx, req.UserID)

	return &requests.ModerateUserResponse{
		UserID:      user.GetId(),
		IsBlocked:   false,
		BlockedAt:   nil,
		BlockReason: "",
	}, nil
}
