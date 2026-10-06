package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	city_repo "github.com/Neratus/geoguide/internal/repository/city"
	config "github.com/Neratus/geoguide/internal/repository/config"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
	favour_repo "github.com/Neratus/geoguide/internal/repository/favouritePlace"
	place_repo "github.com/Neratus/geoguide/internal/repository/place"
	session_repo "github.com/Neratus/geoguide/internal/repository/session"
	static_repo "github.com/Neratus/geoguide/internal/repository/static"
	trip_repo "github.com/Neratus/geoguide/internal/repository/trip"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
	"github.com/gocql/gocql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/tarantool/go-tarantool"
)

type Repositories struct {
	UserRepo      interfaces.UserRepository
	SessionRepo   interfaces.SessionRepository
	PlaceRepo     interfaces.PlaceRepository
	CityRepo      interfaces.CityRepository
	CountryRepo   interfaces.CountryRepository
	TripRepo      interfaces.TripRepository
	FavouriteRepo interfaces.FavouriteRepository
	StaticRepo    interfaces.StaticPageRepository
	MinioClient   *minio.Client

	PostgresDB       *sql.DB
	CassandraSession *gocql.Session
	TarantoolConn    *tarantool.Connection
}

func NewRepositories(cfg *config.Config) (*Repositories, error) {
	minioClient, err := initMinio(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init minio: %w", err)
	}

	var userRepo interfaces.UserRepository
	var placeRepo interfaces.PlaceRepository
	var cityRepo interfaces.CityRepository
	var countryRepo interfaces.CountryRepository
	var tripRepo interfaces.TripRepository
	var favouriteRepo interfaces.FavouriteRepository
	var staticRepo interfaces.StaticPageRepository
	var postgresDB *sql.DB
	var cassSession *gocql.Session

	switch cfg.Database.PrimaryType {
	case "postgres":
		postgresDB, err = initPostgresDB(cfg)
		if err != nil {
			return nil, err
		}
		userRepo, err = user_repo.NewPostgresUserRepo(cfg)
		if err != nil {
			return nil, err
		}
		placeRepo, err = place_repo.NewPostgresPlaceRepo(cfg)
		if err != nil {
			return nil, err
		}
		cityRepo, err = city_repo.NewPostgresCityRepo(cfg)
		if err != nil {
			return nil, err
		}
		countryRepo, err = country_repo.NewPostgresCountryRepo(cfg)
		if err != nil {
			return nil, err
		}
		tripRepo, err = trip_repo.NewPostgresTripRepo(cfg)
		if err != nil {
			return nil, err
		}
		favouriteRepo, err = favour_repo.NewPostgresFavouriteRepository(cfg)
		if err != nil {
			return nil, err
		}
		staticRepo, err = static_repo.NewPostgresStaticRepo(cfg, minioClient)
		if err != nil {
			return nil, err
		}
	case "cassandra":
		cassSession, err = initCassandra(cfg)
		if err != nil {
			return nil, err
		}
		userRepo = user_repo.NewCassandraUserRepo(cassSession)
		placeRepo = place_repo.NewCassandraPlaceRepo(cassSession)
		cityRepo = city_repo.NewCassandraCityRepo(cassSession)
		countryRepo = country_repo.NewCassandraCountryRepo(cassSession)
		tripRepo = trip_repo.NewCassandraTripRepo(cassSession)
		favouriteRepo = favour_repo.NewCassandraFavouriteRepository(cassSession)
		staticRepo = static_repo.NewCassandraStaticRepo(cassSession, minioClient, cfg.Database.Minio.Bucket)
	default:
		return nil, fmt.Errorf("unsupported primary database type: %s", cfg.Database.PrimaryType)
	}

	var sessionRepo interfaces.SessionRepository
	switch cfg.Database.CacheType {
	case "redis":
		sessionRepo, err = session_repo.NewSessionRepo(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to init redis session repo: %w", err)
		}
	case "tarantool":
		sessionRepo, err = session_repo.NewTarantoolSessionRepo(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to init tarantool: %w", err)
		}
		err = ensureTarantoolSchema(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to init tarantool: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported cache type: %s", cfg.Database.CacheType)
	}

	return &Repositories{
		UserRepo:         userRepo,
		SessionRepo:      sessionRepo,
		PlaceRepo:        placeRepo,
		CityRepo:         cityRepo,
		CountryRepo:      countryRepo,
		TripRepo:         tripRepo,
		FavouriteRepo:    favouriteRepo,
		StaticRepo:       staticRepo,
		MinioClient:      minioClient,
		PostgresDB:       postgresDB,
		CassandraSession: cassSession,
	}, nil
}

func initMinio(cfg *config.Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Database.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Database.Minio.AccessKey, cfg.Database.Minio.SecretKey, ""),
		Secure: cfg.Database.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

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
	return minioClient, nil
}

func initPostgresDB(cfg *config.Config) (*sql.DB, error) {
	connStr := cfg.PostgresConnString()
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	return db, nil
}

func initCassandra(cfg *config.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Database.Cassandra.Hosts...)
	cluster.Keyspace = cfg.Database.Cassandra.Keyspace
	cluster.Consistency = gocql.ParseConsistency(cfg.Database.Cassandra.Consistency)
	if cfg.Database.Cassandra.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Database.Cassandra.Username,
			Password: cfg.Database.Cassandra.Password,
		}
	}
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create cassandra session: %w", err)
	}
	return session, nil
}

func (r *Repositories) GetPostgresDB() *sql.DB {
	return r.PostgresDB
}

func (r *Repositories) GetCassandraSession() *gocql.Session {
	return r.CassandraSession
}

func ensureTarantoolSchema(cfg *config.Config) error {
	cfgTar := cfg.Database.Tarantool
	if len(cfgTar.Hosts) == 0 {
		return errors.New("no tarantool hosts provided")
	}
	opts := tarantool.Opts{
		User: cfgTar.Username,
		Pass: cfgTar.Password,
	}
	conn, err := tarantool.Connect(cfgTar.Hosts[0], opts)
	if err != nil {
		return fmt.Errorf("failed to connect to tarantool: %w", err)
	}
	_, err = conn.Eval(`
        if box.space['sessions'] == nil then
            box.schema.space.create('sessions', { if_not_exists = true })
            box.space.sessions:format({
                {name = 'key', type = 'string'},
                {name = 'value', type = 'string'},
                {name = 'expires_at', type = 'unsigned'}
            })
            box.space.sessions:create_index('primary', { type = 'hash', parts = {'key'}, if_not_exists = true })
        end
    `, []interface{}{})

	if err != nil {
		return fmt.Errorf("failed to create sessions space: %w", err)
	}
	return nil
}

func (r *Repositories) Close() error {
	var errs []error

	if r.PostgresDB != nil {
		if err := r.PostgresDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close PostgresDB: %w", err))
		}
	}
	if r.CassandraSession != nil {
		r.CassandraSession.Close()
	}
	if closer, ok := r.UserRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close UserRepo: %w", err))
		}
	}
	if closer, ok := r.PlaceRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close PlaceRepo: %w", err))
		}
	}
	if closer, ok := r.CityRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close CityRepo: %w", err))
		}
	}
	if closer, ok := r.CountryRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close CountryRepo: %w", err))
		}
	}
	if closer, ok := r.TripRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close TripRepo: %w", err))
		}
	}
	if closer, ok := r.FavouriteRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close FavouriteRepo: %w", err))
		}
	}
	if closer, ok := r.StaticRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close StaticRepo: %w", err))
		}
	}
	if r.SessionRepo != nil {
		if closer, ok := r.SessionRepo.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close SessionRepo: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing repositories: %v", errs)
	}
	return nil
}
