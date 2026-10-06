//go:build cassandra
// +build cassandra

package city_repo

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
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
	if err := ensureTablesCity(cassandraSession); err != nil {
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
	cluster.Consistency = gocql.Quorum
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
	clusterWithKeyspace.Consistency = gocql.Quorum
	clusterWithKeyspace.Timeout = 10 * time.Second
	return clusterWithKeyspace.CreateSession()
}

func ensureTablesCity(session *gocql.Session) error {
	dropQueries := []string{
		"DROP TABLE IF EXISTS country_by_id",
		"DROP TABLE IF EXISTS country_by_name",
		"DROP TABLE IF EXISTS city_by_id",
		"DROP TABLE IF EXISTS city_by_country",
		"DROP TABLE IF EXISTS city_by_name",
		"DROP TABLE IF EXISTS district_by_city",
		"DROP TABLE IF EXISTS transport_node_by_city",
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
		`CREATE TABLE city_by_id (
            id UUID PRIMARY KEY,
            name TEXT,
            population BIGINT,
            is_capital BOOLEAN,
            coordinates TEXT,
            description TEXT,
            timezone TEXT,
            travel_tips TEXT,
            country_id UUID,
            image_id UUID
        )`,
		`CREATE TABLE city_by_country (
            country_id UUID,
            name TEXT,
            id UUID,
            population BIGINT,
            is_capital BOOLEAN,
            coordinates TEXT,
            description TEXT,
            timezone TEXT,
            travel_tips TEXT,
            image_id UUID,
            PRIMARY KEY (country_id, name, id)
        ) WITH CLUSTERING ORDER BY (name ASC)`,
		`CREATE TABLE city_by_name (
            name TEXT PRIMARY KEY,
            id UUID,
            country_id UUID
        )`,
		`CREATE TABLE district_by_city (
            city_id UUID,
            id UUID,
            name TEXT,
            description TEXT,
            coordinates TEXT,
            image_id UUID,
            PRIMARY KEY (city_id, id)
        )`,
		`CREATE INDEX IF NOT EXISTS ON district_by_city (id)`,
		`CREATE TABLE transport_node_by_city (
            city_id UUID,
            type TEXT,
            id UUID,
            name TEXT,
            coordinates TEXT,
            address TEXT,
            image_id UUID,
            PRIMARY KEY (city_id, type, id)
        ) WITH CLUSTERING ORDER BY (type ASC)`,
		`CREATE INDEX IF NOT EXISTS ON transport_node_by_city (id)`,
	}
	for _, q := range queries {
		if err := session.Query(q).Exec(); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	log.Println("Cassandra tables for city and country ready")
	return nil
}

func cleanCassandraTables(session *gocql.Session) {
	tables := []string{
		"city_by_id", "city_by_country", "city_by_name",
		"district_by_city", "transport_node_by_city",
		"country_by_id", "country_by_name",
	}
	for _, t := range tables {
		if err := session.Query(fmt.Sprintf("TRUNCATE TABLE %s", t)).Exec(); err != nil {
			log.Printf("Warning truncate %s: %v", t, err)
		}
	}
}

func NewTestCassandraCityRepo(session *gocql.Session) *CassandraCityRepo {
	return NewCassandraCityRepo(session)
}

func NewTestCassandraCountryRepo(session *gocql.Session) *country_repo.CassandraCountryRepo {
	return country_repo.NewCassandraCountryRepo(session)
}

func TestCassandraCityRepository(t *testing.T) {
	cityRepo := NewTestCassandraCityRepo(cassandraSession)
	countryRepo := NewTestCassandraCountryRepo(cassandraSession)

	t.Run("CityNotFound", func(t *testing.T) {
		cityRepository_CityNotFound(t, cityRepo)
	})
	t.Run("CityFound", func(t *testing.T) {
		cityRepository_CityFound(t, cityRepo, countryRepo)
	})
	t.Run("CityFindByCountry", func(t *testing.T) {
		cityRepository_CityFindByCountry(t, cityRepo, countryRepo)
	})
	t.Run("CityFindByName", func(t *testing.T) {
		cityRepository_CityFindByName(t, cityRepo, countryRepo)
	})
	t.Run("CityUpdate", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		cityRepository_CityUpdate(t, cityRepo, countryRepo)
	})
	t.Run("CityDelete", func(t *testing.T) {
		cityRepository_CityDelete(t, cityRepo, countryRepo)
	})
	t.Run("DistrictNotFound", func(t *testing.T) {
		cityRepository_DistrictNotFound(t, cityRepo)
	})
	t.Run("DistrictSaveAndFind", func(t *testing.T) {
		cityRepository_DistrictSaveAndFind(t, cityRepo, countryRepo)
	})
	t.Run("DistrictUpdate", func(t *testing.T) {
		cityRepository_DistrictUpdate(t, cityRepo, countryRepo)
	})
	t.Run("DistrictDelete", func(t *testing.T) {
		cityRepository_DistrictDelete(t, cityRepo, countryRepo)
		time.Sleep(100 * time.Millisecond)
	})
	t.Run("TransportNodeNotFound", func(t *testing.T) {
		cityRepository_TransportNodeNotFound(t, cityRepo)
	})
	t.Run("TransportNodeSaveAndFind", func(t *testing.T) {
		cityRepository_TransportNodeSaveAndFind(t, cityRepo, countryRepo)
		time.Sleep(50 * time.Millisecond)
	})
	t.Run("TransportNodeFindByCity", func(t *testing.T) {
		cityRepository_TransportNodeFindByCity(t, cityRepo, countryRepo)
	})
	t.Run("TransportNodeUpdate", func(t *testing.T) {
		cityRepository_TransportNodeUpdate(t, cityRepo, countryRepo)
		time.Sleep(50 * time.Millisecond)
	})
	t.Run("TransportNodeDelete", func(t *testing.T) {
		cityRepository_TransportNodeDelete(t, cityRepo, countryRepo)
	})
}
