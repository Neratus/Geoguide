package content

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type DeleteStaticPageUseCase struct {
	staticRepo interfaces.StaticPageRepository
	logger     *slog.Logger
}

func NewDeleteStaticPageUseCase(
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *DeleteStaticPageUseCase {
	return &DeleteStaticPageUseCase{
		staticRepo: staticRepo,
		logger:     logger,
	}
}

type DeleteStaticPageRequest struct {
	ID domain.StaticPageID
}

func (uc *DeleteStaticPageUseCase) Execute(ctx context.Context, req DeleteStaticPageRequest) error {
	uc.logger.Info("deleting static page", "page_id", req.ID)

	if err := uc.staticRepo.Delete(ctx, req.ID); err != nil {
		uc.logger.Error("failed to delete static page", "error", err)
		return err
	}

	uc.logger.Info("static page deleted", "page_id", req.ID)
	return nil
}
