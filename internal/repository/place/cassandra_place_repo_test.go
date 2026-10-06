//go:build cassandra
// +build cassandra

package place_repo

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	city_repo "github.com/Neratus/geoguide/internal/repository/city"
	config "github.com/Neratus/geoguide/internal/repository/config"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
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

	if err := ensureTablesPlace(cassandraSession); err != nil {
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

func ensureTablesPlace(session *gocql.Session) error {
	dropQueries := []string{
		"DROP TABLE IF EXISTS trip_by_id",
		"DROP TABLE IF EXISTS trip_by_user",
		"DROP TABLE IF EXISTS trip_place_by_trip",
		"DROP TABLE IF EXISTS place_by_id",
		"DROP TABLE IF EXISTS place_by_city",
		"DROP TABLE IF EXISTS place_by_category",
		"DROP TABLE IF EXISTS review_by_id",
		"DROP TABLE IF EXISTS review_by_place",
		"DROP TABLE IF EXISTS review_by_user",
		"DROP TABLE IF EXISTS city_by_id",
		"DROP TABLE IF EXISTS city_by_country",
		"DROP TABLE IF EXISTS city_by_name",
		"DROP TABLE IF EXISTS country_by_id",
		"DROP TABLE IF EXISTS country_by_name",
		"DROP TABLE IF EXISTS user_by_id",
		"DROP TABLE IF EXISTS user_by_username",
		"DROP TABLE IF EXISTS user_by_email",
	}
	for _, q := range dropQueries {
		_ = session.Query(q).Exec()
	}

	queries := []string{
		`CREATE TABLE trip_by_id (
            id UUID PRIMARY KEY,
            title TEXT,
            start_date DATE,
            end_date DATE,
            budget DOUBLE,
            status TEXT,
            notes TEXT,
            created_at TIMESTAMP,
            user_id UUID,
            image_id UUID
        )`,
		`CREATE TABLE trip_by_user (
            user_id UUID,
            start_date DATE,
            id UUID,
            title TEXT,
            end_date DATE,
            budget DOUBLE,
            status TEXT,
            notes TEXT,
            created_at TIMESTAMP,
            image_id UUID,
            PRIMARY KEY (user_id, start_date, id)
        ) WITH CLUSTERING ORDER BY (start_date DESC)`,
		`CREATE TABLE trip_place_by_trip (
    trip_id UUID,
    place_id UUID,
    day_number INT,
    id UUID,
    arrival_time TIME,
    duration_min INT,
    notes TEXT,
    visit_status TEXT,
    actual_cost DOUBLE,
    PRIMARY KEY ((trip_id, place_id), day_number)
) WITH CLUSTERING ORDER BY (day_number ASC)`,
		`CREATE INDEX IF NOT EXISTS idx_trip_place_by_trip_trip ON trip_place_by_trip (trip_id);`,
		`CREATE TABLE place_by_id (
            id UUID PRIMARY KEY,
            name TEXT,
            category TEXT,
            description TEXT,
            coordinates TEXT,
            address TEXT,
            opening_hours TEXT,
            price_info TEXT,
            avg_visit_duration_min INT,
            avg_rating DOUBLE,
            reviews_count INT,
            contact_phone TEXT,
            website TEXT,
            city_id UUID,
            city_name TEXT,
            district_id UUID,
            district_name TEXT,
            image_id UUID,
            created_at TIMESTAMP
        )`,
		`CREATE TABLE place_by_city (
            city_id UUID,
            avg_rating DOUBLE,
            id UUID,
            name TEXT,
            category TEXT,
            description TEXT,
            coordinates TEXT,
            address TEXT,
            opening_hours TEXT,
            price_info TEXT,
            avg_visit_duration_min INT,
            reviews_count INT,
            contact_phone TEXT,
            website TEXT,
            district_id UUID,
            district_name TEXT,
            image_id UUID,
            created_at TIMESTAMP,
            PRIMARY KEY (city_id, avg_rating, id)
        ) WITH CLUSTERING ORDER BY (avg_rating DESC)`,
		`CREATE TABLE place_by_category (
            category TEXT,
            avg_rating DOUBLE,
            id UUID,
            name TEXT,
            description TEXT,
            city_id UUID,
            city_name TEXT,
            image_id UUID,
            PRIMARY KEY (category, avg_rating, id)
        ) WITH CLUSTERING ORDER BY (avg_rating DESC)`,
		`CREATE TABLE review_by_id (
            id UUID PRIMARY KEY,
            place_id UUID,
            user_id UUID,
            rating INT,
            comment TEXT,
            visit_date DATE,
            created_at TIMESTAMP,
            is_moderated BOOLEAN,
            is_approved BOOLEAN,
            moderation_comment TEXT,
            image_id UUID,
            username TEXT
        )`,
		`CREATE TABLE review_by_place (
            place_id UUID,
            created_at TIMESTAMP,
            id UUID,
            rating INT,
            comment TEXT,
            visit_date DATE,
            username TEXT,
            is_approved BOOLEAN,
            is_moderated BOOLEAN,
            moderation_comment TEXT,
            image_id UUID,
            PRIMARY KEY (place_id, created_at, id)
        ) WITH CLUSTERING ORDER BY (created_at DESC)`,
		`CREATE TABLE review_by_user (
            user_id UUID,
            created_at TIMESTAMP,
            id UUID,
            place_id UUID,
            rating INT,
            comment TEXT,
            visit_date DATE,
            is_approved BOOLEAN,
            is_moderated BOOLEAN,
            PRIMARY KEY (user_id, created_at, id)
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
		`CREATE TABLE user_by_id (
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
		`CREATE TABLE user_by_username (
            username TEXT PRIMARY KEY,
            id UUID
        )`,
		`CREATE TABLE user_by_email (
            email TEXT PRIMARY KEY,
            id UUID
        )`,
	}
	for _, q := range queries {
		if err := session.Query(q).Exec(); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	log.Println("Cassandra tables for trip tests ready")
	return nil
}

func cleanCassandraTables(session *gocql.Session) {
	tables := []string{
		"place_by_id", "place_by_city", "place_by_category",
		"review_by_id", "review_by_place", "review_by_user",
		"city_by_id", "city_by_country", "city_by_name",
		"country_by_id", "country_by_name",
		"user_by_id", "user_by_username", "user_by_email",
	}
	for _, t := range tables {
		if err := session.Query(fmt.Sprintf("TRUNCATE TABLE %s", t)).Exec(); err != nil {
			log.Printf("Warning truncate %s: %v", t, err)
		}
	}
}

func NewTestCassandraPlaceRepo(session *gocql.Session) *CassandraPlaceRepo {
	return NewCassandraPlaceRepo(session)
}

func NewTestCassandraCountryRepo(session *gocql.Session) *country_repo.CassandraCountryRepo {
	return country_repo.NewCassandraCountryRepo(session)
}

func NewTestCassandraCityRepo(session *gocql.Session) *city_repo.CassandraCityRepo {
	return city_repo.NewCassandraCityRepo(session)
}

func NewTestCassandraUserRepo(session *gocql.Session) *user_repo.CassandraUserRepo {
	return user_repo.NewCassandraUserRepo(session)
}

func TestCassandraPlaceRepository(t *testing.T) {
	placeRepo := NewTestCassandraPlaceRepo(cassandraSession)
	countryRepo := NewTestCassandraCountryRepo(cassandraSession)
	cityRepo := NewTestCassandraCityRepo(cassandraSession)
	userRepo := NewTestCassandraUserRepo(cassandraSession)

	t.Run("SaveNewPlace", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_SaveNewPlace(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_FindByID_NotFound(t, placeRepo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_FindByID_Found(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindByCity", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_FindByCity(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindByCategory", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_FindByCategory(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("UpdatePlace", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_UpdatePlace_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("DeletePlace", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_DeletePlace_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("UpdateRating", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_UpdateRating_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})

	t.Run("SaveNewReview", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_SaveNewReview(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindReviewByID_NotFound", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_FindReviewByID_NotFound(t, placeRepo)
	})
	t.Run("FindReviewByID_Found", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_FindReviewByID_Found(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindReviewsByPlace", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_FindReviewsByPlace(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("FindReviewsByUser", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_FindReviewsByUser(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("UpdateReview", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_UpdateReview_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("DeleteReview", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_DeleteReview_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
	t.Run("ModerateReview", func(t *testing.T) {
		cleanCassandraTables(cassandraSession)
		placeRepository_ModerateReview_Success(t, placeRepo, countryRepo, cityRepo, userRepo)
	})
}
