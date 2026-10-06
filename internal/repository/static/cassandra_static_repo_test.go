//go:build cassandra
// +build cassandra

package static_repo

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	"github.com/gocql/gocql"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

var cassandraSession *gocql.Session
var minioClient *minio.Client
var bucketName string

func loadCassandraConfig() (*config.Config, error) {
	v := viper.New()
	v.SetConfigFile("../../../config/config.cassandra.yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func TestMain(m *testing.M) {
	cfg, err := loadCassandraConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if cfg.Database.PrimaryType != "cassandra" {
		log.Println("Skipping Cassandra tests: primary_type != cassandra")
		os.Exit(0)
	}
	cassandraSession, err = createCassandraSessionWithKeyspace(cfg)
	if err != nil {
		log.Fatalf("Failed to create Cassandra session: %v", err)
	}
	defer cassandraSession.Close()

	minioCfg := cfg.Database.Minio
	minioClient, err = minio.New(minioCfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioCfg.AccessKey, minioCfg.SecretKey, ""),
		Secure: minioCfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create minio client: %v", err)
	}
	bucketName = minioCfg.Bucket
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	}

	if err := ensureTablesStatic(cassandraSession); err != nil {
		log.Fatalf("Failed to ensure tables: %v", err)
	}
	cleanCassandraTables(cassandraSession)
	domain.SetConfig(domain.GetConfig())

	code := m.Run()
	os.Exit(code)
}

func createCassandraSessionWithKeyspace(cfg *config.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Database.Cassandra.Hosts...)
	cluster.Timeout = 10 * time.Second
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	keyspace := cfg.Database.Cassandra.Keyspace
	var count int
	iter := session.Query(`SELECT COUNT(*) FROM system_schema.keyspaces WHERE keyspace_name = ?`, keyspace).Iter()
	iter.Scan(&count)
	iter.Close()
	if count == 0 {
		createKeyspace := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
			WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`, keyspace)
		if err := session.Query(createKeyspace).Exec(); err != nil {
			return nil, err
		}
		log.Printf("Keyspace %s created", keyspace)
	}
	clusterWithKeyspace := gocql.NewCluster(cfg.Database.Cassandra.Hosts...)
	clusterWithKeyspace.Keyspace = keyspace
	clusterWithKeyspace.Consistency = gocql.ParseConsistency(cfg.Database.Cassandra.Consistency)
	clusterWithKeyspace.Timeout = 10 * time.Second
	return clusterWithKeyspace.CreateSession()
}

func ensureTablesStatic(session *gocql.Session) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS static_page_by_id (
			id UUID PRIMARY KEY,
			slug TEXT,
			title TEXT,
			content TEXT,
			meta_description TEXT,
			image_url TEXT,
			image_alt TEXT,
			updated_at TIMESTAMP,
			published_at TIMESTAMP,
			is_published BOOLEAN
		)`,
		`CREATE TABLE IF NOT EXISTS static_page_by_slug (
			slug TEXT PRIMARY KEY,
			id UUID
		)`,
	}
	for _, q := range queries {
		if err := session.Query(q).Exec(); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	log.Println("Cassandra tables for static page ready")
	return nil
}

func cleanCassandraTables(session *gocql.Session) {
	tables := []string{"static_page_by_id", "static_page_by_slug"}
	for _, t := range tables {
		if err := session.Query(fmt.Sprintf("TRUNCATE TABLE %s", t)).Exec(); err != nil {
			log.Printf("Warning truncate %s: %v", t, err)
		}
	}
}

func TestCassandraStaticRepository(t *testing.T) {
	repo := NewCassandraStaticRepo(cassandraSession, minioClient, bucketName)

	t.Run("SaveNew", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		staticRepository_SaveNew(t, repo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		staticRepository_FindByID_NotFound(t, repo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		staticRepository_FindByID_Found(t, repo)
	})
	t.Run("Update_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		staticRepository_Update_Success(t, repo)
	})
	t.Run("Delete_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		staticRepository_Delete_Success(t, repo)
	})
	t.Run("GetFile_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		staticRepository_GetFile_Success(t, repo)
	})
}
