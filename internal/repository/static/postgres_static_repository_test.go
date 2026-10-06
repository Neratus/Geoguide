//go:build postgres
// +build postgres

package static_repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pressly/goose/v3"
)

var testCfg *config.Config

func TestMain(m *testing.M) {
	var err error
	testCfg, err = config.LoadTest()
	if err != nil {
		log.Fatalf("Failed to load test config: %v", err)
	}
	if err := runMigrations(testCfg); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}
	domain.SetConfig(domain.GetConfig())
	code := m.Run()
	os.Exit(code)
}

func runMigrations(cfg *config.Config) error {
	connStr := cfg.PostgresConnString()
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return goose.Up(sqlDB, "../../../migrations")
}

func cleanTablesStatic(ctx context.Context, pool *pgxpool.Pool) {
	tables := []string{
		"TripPlace", "Review", "Trip", "Place", "CityDistrict",
		"TransportNode", "City", "Holiday", "Country", "User", "StaticPage",
	}
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE \"%s\" RESTART IDENTITY CASCADE;", table))
		if err != nil {
			log.Printf("Warning: truncate table %s failed: %v", table, err)
		}
	}
}

func NewTestPostgresStaticRepo(cfg *config.Config) (*PostgresStaticRepo, error) {
	connStr := cfg.PostgresConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	queries := postgreSQL.New(pool)
	cleanTablesStatic(context.Background(), pool)

	minioCfg := cfg.Database.Minio
	minioClient, err := minio.New(cfg.Database.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Database.Minio.AccessKey, cfg.Database.Minio.SecretKey, ""),
		Secure: cfg.Database.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	bucketName := minioCfg.Bucket
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.Database.Minio.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.Database.Minio.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &PostgresStaticRepo{
		pool:        pool,
		queries:     queries,
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}
func TestPostgresStaticRepository(t *testing.T) {
	repo, err := NewTestPostgresStaticRepo(testCfg)
	if err != nil {
		t.Fatalf("Failed to create static repo: %v", err)
	}

	t.Run("SaveNew", func(t *testing.T) {
		cleanTablesStatic(context.Background(), repo.pool)
		staticRepository_SaveNew(t, repo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		cleanTablesStatic(context.Background(), repo.pool)
		staticRepository_FindByID_NotFound(t, repo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		cleanTablesStatic(context.Background(), repo.pool)
		staticRepository_FindByID_Found(t, repo)
	})
	t.Run("Update_Success", func(t *testing.T) {
		cleanTablesStatic(context.Background(), repo.pool)
		staticRepository_Update_Success(t, repo)
	})
	t.Run("Delete_Success", func(t *testing.T) {
		cleanTablesStatic(context.Background(), repo.pool)
		staticRepository_Delete_Success(t, repo)
	})
	t.Run("GetFile_Success", func(t *testing.T) {
		cleanTablesStatic(context.Background(), repo.pool)
		staticRepository_GetFile_Success(t, repo)
	})
}
