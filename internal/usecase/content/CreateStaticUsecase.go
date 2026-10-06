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

type CreateStaticPageUseCase struct {
	staticRepo interfaces.StaticPageRepository
	logger     *slog.Logger
}

func NewCreateStaticPageUseCase(
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *CreateStaticPageUseCase {
	return &CreateStaticPageUseCase{
		staticRepo: staticRepo,
		logger:     logger,
	}
}

type CreateStaticPageRequest struct {
	Slug        string
	Title       string
	Content     string
	MetaDesc    string
	ImageAlt    string
	PublishedAt *time.Time
	ImageFile   io.Reader
}

type CreateStaticPageResponse struct {
	ID domain.StaticPageID
}

func (uc *CreateStaticPageUseCase) Execute(ctx context.Context, req CreateStaticPageRequest) (*CreateStaticPageResponse, error) {
	uc.logger.Info("creating static page", "slug", req.Slug)

	existing, _ := uc.staticRepo.FindBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, usecase_errors.ErrSlugAlreadyExists
	}

	page, err := domain.NewStaticPage(
		domain.ImageID{},
		req.Slug,
		req.Title,
		req.Content,
		req.MetaDesc,
		"",
		req.ImageAlt,
	)
	if err != nil {
		uc.logger.Error("invalid static page data", "error", err)
		return nil, err
	}

	if err := uc.staticRepo.Save(ctx, page, req.ImageFile); err != nil {
		uc.logger.Error("failed to save static page", "error", err)
		return nil, err
	}

	uc.logger.Info("static page created", "id", page.GetId())
	return &CreateStaticPageResponse{ID: page.GetId()}, nil
}
