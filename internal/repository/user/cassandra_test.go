//go:build cassandra
// +build cassandra

package user_repo

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	"github.com/gocql/gocql"
	"github.com/spf13/viper"
)

var cassandraSession *gocql.Session

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
		log.Fatalf("Failed to load test config: %v", err)
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
	if err := ensureTables(cassandraSession); err != nil {
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
		return nil, fmt.Errorf("failed to connect to Cassandra: %w", err)
	}
	defer session.Close()

	keyspace := cfg.Database.Cassandra.Keyspace
	var count int
	iter := session.Query(`SELECT COUNT(*) FROM system_schema.keyspaces WHERE keyspace_name = ?`, keyspace).Iter()
	iter.Scan(&count)
	iter.Close()

	if count == 0 {
		createKeyspaceQuery := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s 
			WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`, keyspace)
		if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
			return nil, fmt.Errorf("failed to create keyspace: %w", err)
		}
		log.Printf("Keyspace %s created", keyspace)
	}

	clusterWithKeyspace := gocql.NewCluster(cfg.Database.Cassandra.Hosts...)
	clusterWithKeyspace.Keyspace = keyspace
	clusterWithKeyspace.Consistency = gocql.ParseConsistency(cfg.Database.Cassandra.Consistency)
	clusterWithKeyspace.Timeout = 10 * time.Second
	return clusterWithKeyspace.CreateSession()
}

func ensureTables(session *gocql.Session) error {
	log.Println("Creating tables if not exist...")
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user_by_id (
			id UUID PRIMARY KEY,
			role TEXT,
			username TEXT,
			email TEXT,
			password_hash TEXT,
			phone TEXT,
			registered_at TIMESTAMP,
			birth_date DATE,
			country_of_residence TEXT,
			avatar_url TEXT,
			favorite_categories SET<TEXT>,
			is_blocked BOOLEAN,
			blocked_at TIMESTAMP,
			block_reason TEXT,
			email_verified BOOLEAN,
			phone_verified BOOLEAN,
			two_factor_enabled BOOLEAN,
			two_factor_secret TEXT,
			backup_codes SET<TEXT>
		)`,
		`CREATE TABLE IF NOT EXISTS user_by_username (
			username TEXT PRIMARY KEY,
			id UUID
		)`,
		`CREATE TABLE IF NOT EXISTS user_by_email (
			email TEXT PRIMARY KEY,
			id UUID
		)`,
	}
	for _, q := range queries {
		if err := session.Query(q).Exec(); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	log.Println("Tables ensured")
	return nil
}

func cleanCassandraTables(session *gocql.Session) {
	tables := []string{"user_by_id", "user_by_username", "user_by_email"}
	for _, table := range tables {
		if err := session.Query(fmt.Sprintf("TRUNCATE TABLE %s", table)).Exec(); err != nil {
			log.Printf("Warning: truncate table %s failed: %v", table, err)
		}
	}
}

func NewTestCassandraUserRepo(session *gocql.Session) *CassandraUserRepo {
	return NewCassandraUserRepo(session)
}

func TestCassandraUserRepository(t *testing.T) {
	repo := NewTestCassandraUserRepo(cassandraSession)

	t.Run("SaveNew", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_SaveNew(t, repo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_FindByID_NotFound(t, repo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_FindByID_Found(t, repo)
	})
	t.Run("FindByUsername_NotFound", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_FindByUsername_NotFound(t, repo)
	})
	t.Run("FindByUsername_Found", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_FindByUsername_Found(t, repo)
	})
	t.Run("FindByEmail_NotFound", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_FindByEmail_NotFound(t, repo)
	})
	t.Run("FindByEmail_Found", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_FindByEmail_Found(t, repo)
	})
	t.Run("Update_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_Update_Success(t, repo)
	})
	t.Run("Delete_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_Delete_Success(t, repo)
	})
	t.Run("BlockUser_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_BlockUser_Success(t, repo)
	})
	t.Run("UnblockUser_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		userRepository_UnblockUser_Success(t, repo)
	})
}
