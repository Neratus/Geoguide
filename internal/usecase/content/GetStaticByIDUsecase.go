package content

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetStaticPageByIDUseCase struct {
	staticRepo interfaces.StaticPageRepository
	logger     *slog.Logger
}

func NewGetStaticPageByIDUseCase(
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *GetStaticPageByIDUseCase {
	return &GetStaticPageByIDUseCase{
		staticRepo: staticRepo,
		logger:     logger,
	}
}

type GetStaticPageByIDRequest struct {
	ID domain.StaticPageID
}

type StaticPageResponse struct {
	ID          domain.StaticPageID
	Slug        string
	Title       string
	Content     string
	MetaDesc    string
	ImageURL    string
	ImageAlt    string
	IsPublished bool
	PublishedAt string
}

func (uc *GetStaticPageByIDUseCase) Execute(ctx context.Context, req GetStaticPageByIDRequest) (*StaticPageResponse, error) {
	uc.logger.Info("getting static page by ID", "id", req.ID)

	page, err := uc.staticRepo.FindByID(ctx, req.ID)
	if err != nil || page == nil {
		uc.logger.Warn("static page not found", "id", req.ID)
		return nil, usecase_errors.ErrStaticPageNotFound
	}

	publishedAt := ""
	if !page.GetPublishedAt().IsZero() {
		publishedAt = page.GetPublishedAt().Format("2006-01-02 15:04:05")
	}

	return &StaticPageResponse{
		ID:          page.GetId(),
		Slug:        page.GetSlug(),
		Title:       page.GetTitle(),
		Content:     page.GetContent(),
		MetaDesc:    page.GetMetaDescription(),
		ImageURL:    page.GetImageURL(),
		ImageAlt:    page.GetImageAlt(),
		IsPublished: page.IsPublished(),
		PublishedAt: publishedAt,
	}, nil
}
