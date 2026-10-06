package static_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

type PostgresStaticRepo struct {
	pool        *pgxpool.Pool
	queries     *postgreSQL.Queries
	minioClient *minio.Client
	bucketName  string
}

func NewPostgresStaticRepo(cfg *config.Config, minioClient *minio.Client) (*PostgresStaticRepo, error) {
	connStr := cfg.PostgresConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	queries := postgreSQL.New(pool)

	return &PostgresStaticRepo{
		pool:        pool,
		queries:     queries,
		minioClient: minioClient,
		bucketName:  cfg.Database.Minio.Bucket,
	}, nil
}

func (r *PostgresStaticRepo) Close() error {
	if r.pool != nil {
		r.pool.Close()
	}
	return nil
}

func (r *PostgresStaticRepo) Save(ctx context.Context, page *domain.StaticPage, file io.Reader) error {
	params := toSaveStaticParams(page)
	id, err := r.queries.SaveStatic(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to save static page: %w", err)
	}
	page.SetId(domain.FromPgUUID(id))

	if file != nil {
		objectName := page.GetId().String()
		contentType := "image/jpeg"
		_, err = r.minioClient.PutObject(ctx, r.bucketName, objectName, file, -1, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			_ = r.queries.DeleteStatic(ctx, domain.ToPgUUID(page.GetId()))
			return fmt.Errorf("failed to upload image to minio: %w", err)
		}
		page.SetImageURL(objectName)
	}
	return nil
}

func (r *PostgresStaticRepo) FindByID(ctx context.Context, id domain.ImageID) (*domain.StaticPage, error) {
	dbPage, err := r.queries.FindStaticByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find static page by ID: %w", err)
	}
	return toDomainStatic(dbPage)
}

func (r *PostgresStaticRepo) Update(ctx context.Context, page *domain.StaticPage, file io.Reader) error {
	if file != nil {
		old, err := r.FindByID(ctx, page.GetId())
		if err != nil {
			return fmt.Errorf("failed to find existing page: %w", err)
		}
		if old.GetId() != uuid.Nil {
			_ = r.minioClient.RemoveObject(ctx, r.bucketName, old.GetId().String(), minio.RemoveObjectOptions{})
		}
		objectName := page.GetId().String()
		contentType := "image/jpeg"
		_, err = r.minioClient.PutObject(ctx, r.bucketName, objectName, file, -1, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			return fmt.Errorf("failed to upload new image: %w", err)
		}
		page.SetImageURL(objectName)
	}

	params := toUpdateStaticParams(page)
	err := r.queries.UpdateStatic(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update static page: %w", err)
	}
	return nil
}

func (r *PostgresStaticRepo) FindBySlug(ctx context.Context, slug string) (*domain.StaticPage, error) {
	row, err := r.queries.FindStaticPageBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find static page by slug: %w", err)
	}
	return toDomainStatic(row)
}

func (r *PostgresStaticRepo) Delete(ctx context.Context, id domain.ImageID) error {
	page, err := r.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find static page: %w", err)
	}
	if page.GetId() != uuid.Nil {
		_, errStat := r.minioClient.StatObject(ctx, r.bucketName, page.GetId().String(), minio.StatObjectOptions{})
		if errStat == nil {
			err = r.minioClient.RemoveObject(ctx, r.bucketName, page.GetId().String(), minio.RemoveObjectOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete image from minio: %w", err)
			}
		}
	}
	err = r.queries.DeleteStatic(ctx, domain.ToPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete static page: %w", err)
	}
	return nil
}

func (r *PostgresStaticRepo) GetFile(ctx context.Context, id domain.ImageID) (io.ReadCloser, error) {
	_, err := r.minioClient.StatObject(ctx, r.bucketName, id.String(), minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	obj, err := r.minioClient.GetObject(ctx, r.bucketName, id.String(), minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return obj, nil
}

func (r *PostgresStaticRepo) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := r.minioClient.PutObject(ctx, r.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (r *PostgresStaticRepo) GetFileURL(ctx context.Context, objectName string) (string, error) {
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

func (r *PostgresStaticRepo) DeleteFile(ctx context.Context, objectName string) error {
	return r.minioClient.RemoveObject(ctx, r.bucketName, objectName, minio.RemoveObjectOptions{})
}

func toSaveStaticParams(page *domain.StaticPage) postgreSQL.SaveStaticParams {
	id := page.GetId()
	if uuid.UUID(id) == uuid.Nil {
		id = uuid.New()
	}
	return postgreSQL.SaveStaticParams{
		ID:              uuidToPgUUID(id),
		Slug:            page.GetSlug(),
		Title:           page.GetTitle(),
		Content:         pgtype.Text{String: page.GetContent(), Valid: page.GetContent() != ""},
		MetaDescription: pgtype.Text{String: page.GetMetaDescription(), Valid: page.GetMetaDescription() != ""},
		ImageUrl:        pgtype.Text{String: page.GetImageURL(), Valid: page.GetImageURL() != ""},
		ImageAlt:        pgtype.Text{String: page.GetImageAlt(), Valid: page.GetImageAlt() != ""},
		IsPublished:     pgtype.Bool{Bool: page.IsPublished(), Valid: true},
	}
}

func toUpdateStaticParams(page *domain.StaticPage) postgreSQL.UpdateStaticParams {
	return postgreSQL.UpdateStaticParams{
		ID:              domain.ToPgUUID(page.GetId()),
		Slug:            page.GetSlug(),
		Title:           page.GetTitle(),
		Content:         pgtype.Text{String: page.GetContent(), Valid: page.GetContent() != ""},
		MetaDescription: pgtype.Text{String: page.GetMetaDescription(), Valid: page.GetMetaDescription() != ""},
		ImageUrl:        pgtype.Text{String: page.GetImageURL(), Valid: page.GetImageURL() != ""},
		ImageAlt:        pgtype.Text{String: page.GetImageAlt(), Valid: page.GetImageAlt() != ""},
		IsPublished:     pgtype.Bool{Bool: page.IsPublished(), Valid: true},
	}
}
func toDomainStatic(dbRow postgreSQL.StaticPage) (*domain.StaticPage, error) {
	id := domain.FromPgUUID(dbRow.ID)
	if id == uuid.Nil {
		return nil, errors.New("static page id is null")
	}
	var publishedAt *time.Time
	if dbRow.PublishedAt.Valid {
		publishedAt = &dbRow.PublishedAt.Time
	}
	return domain.NewStaticPageFromDB(
		id,
		dbRow.Slug,
		dbRow.Title,
		dbRow.Content.String,
		dbRow.MetaDescription.String,
		dbRow.ImageUrl.String,
		dbRow.ImageAlt.String,
		dbRow.UpdatedAt.Time,
		publishedAt,
		dbRow.IsPublished.Bool,
	), nil
}

func uuidToPgUUID(u uuid.UUID) pgtype.UUID {
	if u == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}
