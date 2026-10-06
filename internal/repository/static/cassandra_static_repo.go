package static_repo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type CassandraStaticRepo struct {
	session     *gocql.Session
	minioClient *minio.Client
	bucketName  string
}

func NewCassandraStaticRepo(session *gocql.Session, minioClient *minio.Client, bucketName string) *CassandraStaticRepo {
	return &CassandraStaticRepo{
		session:     session,
		minioClient: minioClient,
		bucketName:  bucketName,
	}
}

func (r *CassandraStaticRepo) Close() error {
	return nil
}

func (r *CassandraStaticRepo) Save(ctx context.Context, page *domain.StaticPage, file io.Reader) error {
	id := page.GetId()
	if id == uuid.Nil {
		id = domain.ImageID(uuid.New())
		page.SetId(id)
	}
	cassID := gocql.UUID(id)

	if err := r.session.Query(`
		INSERT INTO static_page_by_id (
			id, slug, title, content, meta_description, image_url, image_alt,
			updated_at, published_at, is_published
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cassID, page.GetSlug(), page.GetTitle(), page.GetContent(),
		page.GetMetaDescription(), page.GetImageURL(), page.GetImageAlt(),
		page.GetUpdatedAt(), page.GetPublishedAt(), page.IsPublished()).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed insert static_page_by_id: %w", err)
	}

	if err := r.session.Query(`
		INSERT INTO static_page_by_slug (slug, id) VALUES (?, ?)
	`, page.GetSlug(), cassID).WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM static_page_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
		return fmt.Errorf("failed insert static_page_by_slug: %w", err)
	}

	if file != nil {
		objectName := id.String()
		contentType := "image/jpeg"
		_, err := r.minioClient.PutObject(ctx, r.bucketName, objectName, file, -1, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			_ = r.session.Query(`DELETE FROM static_page_by_id WHERE id = ?`, cassID).WithContext(ctx).Exec()
			_ = r.session.Query(`DELETE FROM static_page_by_slug WHERE slug = ?`, page.GetSlug()).WithContext(ctx).Exec()
			return fmt.Errorf("failed to upload image: %w", err)
		}
		page.SetImageURL(objectName)
	}

	return nil
}

func (r *CassandraStaticRepo) FindByID(ctx context.Context, id domain.ImageID) (*domain.StaticPage, error) {
	var (
		cassID          gocql.UUID
		slug            string
		title           string
		content         string
		metaDescription string
		imageURL        string
		imageAlt        string
		updatedAt       time.Time
		publishedAt     *time.Time
		isPublished     bool
	)
	query := r.session.Query(`
		SELECT id, slug, title, content, meta_description, image_url, image_alt,
		       updated_at, published_at, is_published
		FROM static_page_by_id WHERE id = ? LIMIT 1
	`, gocql.UUID(id))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(&cassID, &slug, &title, &content, &metaDescription,
		&imageURL, &imageAlt, &updatedAt, &publishedAt, &isPublished) {
		_ = iter.Close()
		return nil, domain.ErrStaticPageNotFound
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	page := domain.NewStaticPageFromDB(
		domain.ImageID(cassID), slug, title, content, metaDescription,
		imageURL, imageAlt, updatedAt, publishedAt, isPublished,
	)
	return page, nil
}

func (r *CassandraStaticRepo) FindBySlug(ctx context.Context, slug string) (*domain.StaticPage, error) {
	var id gocql.UUID
	if err := r.session.Query(`SELECT id FROM static_page_by_slug WHERE slug = ? LIMIT 1`, slug).WithContext(ctx).Scan(&id); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, domain.ErrStaticPageNotFound
		}
		return nil, err
	}
	return r.FindByID(ctx, domain.ImageID(id))
}

func (r *CassandraStaticRepo) Update(ctx context.Context, page *domain.StaticPage, file io.Reader) error {
	id := gocql.UUID(page.GetId())

	var oldSlug string
	if err := r.session.Query(`SELECT slug FROM static_page_by_id WHERE id = ? LIMIT 1`, id).WithContext(ctx).Scan(&oldSlug); err != nil {
		return fmt.Errorf("failed to get old slug: %w", err)
	}

	if err := r.session.Query(`
		UPDATE static_page_by_id SET
			slug = ?, title = ?, content = ?, meta_description = ?,
			image_url = ?, image_alt = ?, updated_at = ?, published_at = ?, is_published = ?
		WHERE id = ?
	`, page.GetSlug(), page.GetTitle(), page.GetContent(),
		page.GetMetaDescription(), page.GetImageURL(), page.GetImageAlt(),
		page.GetUpdatedAt(), page.GetPublishedAt(), page.IsPublished(), id).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed update static_page_by_id: %w", err)
	}

	if oldSlug != page.GetSlug() {
		if err := r.session.Query(`DELETE FROM static_page_by_slug WHERE slug = ?`, oldSlug).WithContext(ctx).Exec(); err != nil {
		}
		if err := r.session.Query(`INSERT INTO static_page_by_slug (slug, id) VALUES (?, ?)`, page.GetSlug(), id).WithContext(ctx).Exec(); err != nil {
			return fmt.Errorf("failed update static_page_by_slug: %w", err)
		}
	}

	if file != nil {
		oldObjectName := page.GetId().String()
		_ = r.minioClient.RemoveObject(ctx, r.bucketName, oldObjectName, minio.RemoveObjectOptions{})
		contentType := "image/jpeg"
		_, err := r.minioClient.PutObject(ctx, r.bucketName, oldObjectName, file, -1, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			return fmt.Errorf("failed to upload new image: %w", err)
		}
		page.SetImageURL(oldObjectName)
	}
	return nil
}

func (r *CassandraStaticRepo) Delete(ctx context.Context, id domain.ImageID) error {
	var slug string
	if err := r.session.Query(`SELECT slug FROM static_page_by_id WHERE id = ? LIMIT 1`, gocql.UUID(id)).WithContext(ctx).Scan(&slug); err != nil {
		return domain.ErrStaticPageNotFound
	}
	if err := r.session.Query(`DELETE FROM static_page_by_id WHERE id = ?`, gocql.UUID(id)).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM static_page_by_slug WHERE slug = ?`, slug).WithContext(ctx).Exec(); err != nil {
	}
	_ = r.minioClient.RemoveObject(ctx, r.bucketName, id.String(), minio.RemoveObjectOptions{})
	return nil
}

func (r *CassandraStaticRepo) GetFile(ctx context.Context, id domain.ImageID) (io.ReadCloser, error) {
	_, err := r.minioClient.StatObject(ctx, r.bucketName, id.String(), minio.StatObjectOptions{})
	if err != nil {
		return nil, domain.ErrFileNotFound
	}
	obj, err := r.minioClient.GetObject(ctx, r.bucketName, id.String(), minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return obj, nil
}

func (r *CassandraStaticRepo) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := r.minioClient.PutObject(ctx, r.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (r *CassandraStaticRepo) GetFileURL(ctx context.Context, objectName string) (string, error) {
	if objectName == "" {
		return "", nil
	}
	reqParams := make(url.Values)
	presignedURL, err := r.minioClient.PresignedGetObject(ctx, r.bucketName, objectName, time.Hour, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func (r *CassandraStaticRepo) DeleteFile(ctx context.Context, objectName string) error {
	return r.minioClient.RemoveObject(ctx, r.bucketName, objectName, minio.RemoveObjectOptions{})
}
