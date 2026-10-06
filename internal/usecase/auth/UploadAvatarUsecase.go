package auth

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
)

type UploadAvatarUseCase struct {
	userRepo interfaces.UserRepository
	fileRepo interfaces.StaticPageRepository
	logger   *slog.Logger
}

func NewUploadAvatarUseCase(userRepo interfaces.UserRepository, fileRepo interfaces.StaticPageRepository, logger *slog.Logger) *UploadAvatarUseCase {
	return &UploadAvatarUseCase{
		userRepo: userRepo,
		fileRepo: fileRepo,
		logger:   logger,
	}
}

func (uc *UploadAvatarUseCase) Execute(ctx context.Context, userID domain.UserID, reader io.Reader, size int64, filename, contentType string) (string, error) {
	uc.logger.Info("uploading avatar", "user_id", userID)

	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return "", usecase_errors.ErrInvalidImageType
	}

	ext := strings.ToLower(filepath.Ext(filename))
	objectName := uuid.UUID(userID).String() + ext

	err := uc.fileRepo.UploadFile(ctx, objectName, reader, size, contentType)
	if err != nil {
		uc.logger.Error("failed to upload avatar", "error", err)
		return "", err
	}

	err = uc.userRepo.UpdateAvatar(ctx, userID, objectName)
	if err != nil {
		_ = uc.fileRepo.DeleteFile(ctx, objectName)
		return "", err
	}

	avatarURL, err := uc.fileRepo.GetFileURL(ctx, objectName)
	if err != nil {
		uc.logger.Warn("failed to generate avatar URL", "error", err)
	}

	return avatarURL, nil
}
