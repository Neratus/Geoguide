//go:build cassandra
// +build cassandra

package country_repo

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
	if err := ensureTablesCountry(cassandraSession); err != nil {
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

func ensureTablesCountry(session *gocql.Session) error {
	dropQueries := []string{
		"DROP TABLE IF EXISTS country_by_id",
		"DROP TABLE IF EXISTS country_by_name",
		"DROP TABLE IF EXISTS holiday_by_country",
		"DROP TABLE IF EXISTS holiday_by_date",
		"DROP TABLE IF EXISTS holiday_by_id",
	}
	for _, q := range dropQueries {
		_ = session.Query(q).Exec()
	}

	queries := []string{
		`CREATE TABLE country_by_id (
            id UUID PRIMARY KEY,
            name TEXT,
            area DOUBLE,
            population BIGINT,
            gdp DOUBLE,
            currency TEXT,
            visa_requirements TEXT,
            description TEXT,
            safety_tips TEXT,
            best_season TEXT,
            language TEXT,
            phone_code TEXT,
            religion TEXT,
            capital_id UUID,
            image_id UUID
        )`,
		`CREATE TABLE country_by_name (
            name TEXT PRIMARY KEY,
            id UUID
        )`,
		`CREATE TABLE holiday_by_country (
            country_id UUID,
            date DATE,
            id UUID,
            name TEXT,
            description TEXT,
            traditions TEXT,
            history TEXT,
            is_national BOOLEAN,
            image_id UUID,
            PRIMARY KEY (country_id, date, id)
        ) WITH CLUSTERING ORDER BY (date ASC)`,
		`CREATE TABLE holiday_by_date (
            date DATE,
            country_id UUID,
            id UUID,
            name TEXT,
            description TEXT,
            is_national BOOLEAN,
            PRIMARY KEY (date, country_id, id)
        )`,
		`CREATE TABLE holiday_by_id (
            id UUID PRIMARY KEY,
            country_id UUID,
            date DATE,
            name TEXT,
            description TEXT,
            traditions TEXT,
            history TEXT,
            is_national BOOLEAN,
            image_id UUID
        )`,
	}
	for _, q := range queries {
		if err := session.Query(q).Exec(); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	log.Println("Cassandra tables for country ready")
	return nil
}

func cleanCassandraTables(session *gocql.Session) {
	tables := []string{"country_by_id", "country_by_name", "holiday_by_country", "holiday_by_date"}
	for _, t := range tables {
		if err := session.Query(fmt.Sprintf("TRUNCATE TABLE %s", t)).Exec(); err != nil {
			log.Printf("Warning truncate %s: %v", t, err)
		}
	}
}

func TestCassandraCountryRepository(t *testing.T) {
	repo := NewCassandraCountryRepo(cassandraSession)

	t.Run("SaveNewCountry", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		countryRepository_SaveNewCountry(t, repo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		countryRepository_FindByID_NotFound(t, repo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		countryRepository_FindByID_Found(t, repo)
	})
	t.Run("FindAll", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		countryRepository_FindAll(t, repo)
	})
	t.Run("FindByName_NotFound", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		countryRepository_FindByName_NotFound(t, repo)
	})
	t.Run("FindByName_Found", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		countryRepository_FindByName_Found(t, repo)
	})
	t.Run("UpdateCountry_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		countryRepository_UpdateCountry_Success(t, repo)
	})
	t.Run("DeleteCountry_Success", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		countryRepository_DeleteCountry_Success(t, repo)
	})

}
