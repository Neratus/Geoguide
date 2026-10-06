package content

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type UpdateStaticPageUseCase struct {
	staticRepo interfaces.StaticPageRepository
	logger     *slog.Logger
}

func NewUpdateStaticPageUseCase(
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *UpdateStaticPageUseCase {
	return &UpdateStaticPageUseCase{
		staticRepo: staticRepo,
		logger:     logger,
	}
}

type UpdateStaticPageRequest struct {
	ID          domain.StaticPageID
	Slug        *string
	Title       *string
	Content     *string
	MetaDesc    *string
	ImageAlt    *string
	IsPublished *bool
	PublishedAt *time.Time
	ImageFile   io.Reader
}

func (uc *UpdateStaticPageUseCase) Execute(ctx context.Context, req UpdateStaticPageRequest) error {
	uc.logger.Info("updating static page", "page_id", req.ID)

	page, err := uc.staticRepo.FindByID(ctx, req.ID)
	if err != nil || page == nil {
		uc.logger.Warn("static page not found", "page_id", req.ID)
		return usecase_errors.ErrStaticPageNotFound
	}

	if req.Slug != nil {
		page.SetSlug(*req.Slug)
	}
	if req.Title != nil {
		page.SetTitle(*req.Title)
	}
	if req.Content != nil {
		page.SetContent(*req.Content)
	}
	if req.MetaDesc != nil {
		page.SetMetaDescription(*req.MetaDesc)
	}
	if req.ImageAlt != nil {
		page.SetImageAlt(*req.ImageAlt)
	}
	if req.IsPublished != nil {
		page.SetPublished(*req.IsPublished)
	}
	if req.PublishedAt != nil {
		page.SetPublishedAt(req.PublishedAt)
	} else if req.IsPublished != nil && *req.IsPublished && page.GetPublishedAt().IsZero() {
		time := time.Now()
		page.SetPublishedAt(&time)
	}

	if err := uc.staticRepo.Update(ctx, page, req.ImageFile); err != nil {
		uc.logger.Error("failed to update static page", "error", err)
		return err
	}

	uc.logger.Info("static page updated", "page_id", req.ID)
	return nil
}
