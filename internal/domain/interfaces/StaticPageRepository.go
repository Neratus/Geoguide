package interfaces

import (
	"context"
	"io"

	"github.com/Neratus/geoguide/internal/domain"
)

type StaticPageRepository interface {
	Save(ctx context.Context, page *domain.StaticPage, file io.Reader) error
	FindByID(ctx context.Context, id domain.ImageID) (*domain.StaticPage, error)
	Update(ctx context.Context, page *domain.StaticPage, file io.Reader) error
	Delete(ctx context.Context, id domain.ImageID) error
	GetFile(ctx context.Context, id domain.ImageID) (io.ReadCloser, error)
	FindBySlug(ctx context.Context, slug string) (*domain.StaticPage, error)
	UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	GetFileURL(ctx context.Context, objectName string) (string, error)
	DeleteFile(ctx context.Context, objectName string) error
}
