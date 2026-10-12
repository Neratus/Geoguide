package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Neratus/geoguide/internal/delivery"
	"github.com/Neratus/geoguide/internal/delivery/graphql/resolvers"
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/logger"
	repository "github.com/Neratus/geoguide/internal/repository"
	"github.com/Neratus/geoguide/internal/repository/config"
	"github.com/Neratus/geoguide/internal/services"
	"github.com/Neratus/geoguide/internal/usecase/admin"
	"github.com/Neratus/geoguide/internal/usecase/analytics"
	"github.com/Neratus/geoguide/internal/usecase/auth"
	"github.com/Neratus/geoguide/internal/usecase/catalog"
	"github.com/Neratus/geoguide/internal/usecase/content"
	"github.com/Neratus/geoguide/internal/usecase/favourites"
	"github.com/Neratus/geoguide/internal/usecase/planner"
	"github.com/Neratus/geoguide/internal/usecase/review"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/pquerna/otp/totp"
	"github.com/pressly/goose/v3"
)

func main() {
	var configPath string
	var interactive bool
	flag.StringVar(&configPath, "config", "config/config.yaml", "path to config file")
	flag.BoolVar(&interactive, "i", true, "interactive CLI mode")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logFile, err := logger.Init(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	domainLoader := domain.NewViperDomainLoader("config/domain_config.yaml")
	domainLoader.StartPolling(30 * time.Second)

	slog.Info("Starting GeoGuide", "version", cfg.App.Version, "config", configPath)

	ctx := context.Background()

	if err := runMigrationsOnly(cfg); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	repos, err := repository.NewRepositories(cfg)
	if err != nil {
		slog.Error("failed to init repositories", "error", err)
		os.Exit(1)
	}
	defer repos.Close()

	if err := cleanAndSeed(ctx, cfg, repos); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}

	userRepo := repos.UserRepo
	sessionRepo := repos.SessionRepo
	placeRepo := repos.PlaceRepo
	cityRepo := repos.CityRepo
	countryRepo := repos.CountryRepo
	tripRepo := repos.TripRepo
	favouriteRepo := repos.FavouriteRepo
	staticRepo := repos.StaticRepo

	emailNotifier, err := services.NewSMTPSender(
		cfg.SMTP.From,
		cfg.SMTP.Password,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.TemplateDir,
	)
	if err != nil {
		slog.Warn("email notifier not available", "error", err)
		emailNotifier = nil
	}

	watermillLogger := watermill.NewStdLogger(false, false)
	taskQueue, err := services.NewWatermillQueue(watermillLogger)
	if err != nil {
		slog.Warn("task queue not available", "error", err)
		taskQueue = nil
	}
	if taskQueue != nil {
		go func() {
			if err := services.RegisterAllConsumers(taskQueue, emailNotifier); err != nil {
				slog.Error("failed to register consumers", "error", err)
			}
			if err := services.StartConsumers(context.Background(), taskQueue); err != nil {
				slog.Error("consumer error", "error", err)
			}
		}()
		defer taskQueue.Close()
	}
	services.InitJWT(cfg.Database.Auth.JWTSecret)

	rateLimiter, err := services.NewRateLimiter(cfg)
	if err != nil {
		slog.Error("failed to init rate limiter", "error", err)
		os.Exit(1)
	}
	reportGenService := services.NewReportGeneratorService(slog.Default())
	generateReportUC := analytics.NewGenerateReportUseCase(placeRepo, userRepo, tripRepo, reportGenService, slog.Default())

	registerUC := auth.NewRegisterUserUseCase(userRepo, sessionRepo, emailNotifier, taskQueue, slog.Default(), rateLimiter)
	loginUC := auth.NewLoginUseCase(userRepo, sessionRepo, emailNotifier, taskQueue, slog.Default(), rateLimiter)
	updateProfileUC := auth.NewUpdateUserProfileUseCase(userRepo, sessionRepo, slog.Default())
	getProfileUC := auth.NewGetUserProfileUseCase(userRepo, sessionRepo, staticRepo, slog.Default())
	logoutUC := auth.NewLogoutUseCase(sessionRepo, slog.Default())
	moderateUserUC := auth.NewModerateUserUsecase(userRepo, emailNotifier, taskQueue, slog.Default())

	getCountriesUC := catalog.NewGetCountriesUseCase(countryRepo, cityRepo, staticRepo, slog.Default())
	getCountryUC := catalog.NewGetCountryUseCase(countryRepo, cityRepo, staticRepo, slog.Default())
	getCitiesUC := catalog.NewGetCitiesByCountryUseCase(cityRepo, staticRepo, slog.Default())
	getPlaceUC := catalog.NewGetPlaceUseCase(placeRepo, userRepo, staticRepo, slog.Default())
	getPlacesByCityUC := catalog.NewGetPlacesByCityUseCase(placeRepo, staticRepo, slog.Default())
	getPlacesByCategoryUC := catalog.NewGetPlacesByCategoryUseCase(placeRepo, staticRepo, slog.Default())

	createTripUC := planner.NewCreateTripUseCase(tripRepo, slog.Default())
	getTripUC := planner.NewGetTripUseCase(tripRepo, slog.Default())
	getUserTripsUC := planner.NewGetUserTripsUseCase(tripRepo, slog.Default())
	addPlaceToTripUC := planner.NewAddPlaceToTripUseCase(tripRepo, placeRepo, slog.Default())
	removePlaceFromTripUC := planner.NewRemovePlaceFromTripUseCase(tripRepo, slog.Default())
	updateTripUC := planner.NewUpdateTripUseCase(tripRepo, slog.Default())
	deleteTripUC := planner.NewDeleteTripUseCase(tripRepo, slog.Default())

	createReviewUC := review.NewCreateReviewUseCase(placeRepo, emailNotifier, taskQueue, slog.Default())
	getReviewsByPlaceUC := review.NewGetReviewsByPlaceUseCase(placeRepo, slog.Default())
	getUserReviewsUC := review.NewGetUserReviewsUseCase(placeRepo, slog.Default())
	moderateReviewUC := review.NewModerateReviewUseCase(placeRepo, userRepo, emailNotifier, taskQueue, slog.Default())
	deleteReviewUC := review.NewDeleteReviewUseCase(placeRepo, slog.Default())

	getCityUC := catalog.NewGetCityUseCase(cityRepo, staticRepo, slog.Default())
	searchCitiesUC := catalog.NewSearchCitiesUseCase(cityRepo, slog.Default())
	getDistrictsUC := catalog.NewGetDistrictsByCityUseCase(cityRepo, slog.Default())
	getTransportNodesUC := catalog.NewGetTransportNodesByCityUseCase(cityRepo, slog.Default())

	getStaticPageUC := content.NewGetStaticPageByIDUseCase(staticRepo, slog.Default())

	verifyContactUC := auth.NewVerifyContactUseCase(userRepo, sessionRepo, emailNotifier, taskQueue, slog.Default())
	enable2FAUC := auth.NewEnableTwoFactorUseCase(userRepo, slog.Default())
	verify2FALoginUC := auth.NewVerifyTwoFactorLoginUseCase(userRepo, sessionRepo, slog.Default())
	getUsersUC := admin.NewGetUsersUseCase(userRepo, slog.Default())
	getPendingReviewsUC := admin.NewGetPendingReviewsUseCase(placeRepo, slog.Default())

	uploadAvatarUC := auth.NewUploadAvatarUseCase(userRepo, staticRepo, slog.Default())
	addFavUC := favourites.NewAddFavouriteUseCase(favouriteRepo, placeRepo, slog.Default())
	removeFavUC := favourites.NewRemoveFavouriteUseCase(favouriteRepo, slog.Default())
	getFavUC := favourites.NewGetFavouritesUseCase(favouriteRepo, slog.Default())

	getHolidays := catalog.NewGetHolidaysByDateUseCase(countryRepo, slog.Default())

	cli := &CLI{
		ctx:    ctx,
		logger: slog.Default(),

		RegisterUC:      registerUC,
		LoginUC:         loginUC,
		UpdateProfileUC: updateProfileUC,
		GetProfileUC:    getProfileUC,
		LogoutUC:        logoutUC,
		ModerateUserUC:  moderateUserUC,
		VerifyContactUC: verifyContactUC,
		Enable2FAUC:     enable2FAUC,
		Verify2FAUC:     verify2FALoginUC,
		UploadAvatarUC:  uploadAvatarUC,

		GetCountriesUC:        getCountriesUC,
		GetCountryUC:          getCountryUC,
		GetCitiesUC:           getCitiesUC,
		GetCityUC:             getCityUC,
		SearchCitiesUC:        searchCitiesUC,
		GetDistrictsUC:        getDistrictsUC,
		GetTransportNodesUC:   getTransportNodesUC,
		GetPlaceUC:            getPlaceUC,
		GetPlacesByCityUC:     getPlacesByCityUC,
		GetPlacesByCategoryUC: getPlacesByCategoryUC,
		GetHolidaysUC:         getHolidays,

		CreateTripUC:          createTripUC,
		GetTripUC:             getTripUC,
		GetUserTripsUC:        getUserTripsUC,
		AddPlaceToTripUC:      addPlaceToTripUC,
		RemovePlaceFromTripUC: removePlaceFromTripUC,
		UpdateTripUC:          updateTripUC,
		DeleteTripUC:          deleteTripUC,

		CreateReviewUC:      createReviewUC,
		GetReviewsByPlaceUC: getReviewsByPlaceUC,
		GetUserReviewsUC:    getUserReviewsUC,
		ModerateReviewUC:    moderateReviewUC,
		DeleteReviewUC:      deleteReviewUC,

		GetStaticPageUC: getStaticPageUC,

		AddFavUC:    addFavUC,
		RemoveFavUC: removeFavUC,
		GetFavUC:    getFavUC,

		GetUsersUC:          getUsersUC,
		GetPendingReviewsUC: getPendingReviewsUC,

		GenerateReportUC: generateReportUC,
	}

	gqlResolver := &resolvers.Resolver{
		Logger: slog.Default(),

		RegisterUC:      registerUC,
		LoginUC:         loginUC,
		UpdateProfileUC: updateProfileUC,
		GetProfileUC:    getProfileUC,
		LogoutUC:        logoutUC,
		ModerateUserUC:  moderateUserUC,
		VerifyContactUC: verifyContactUC,
		Enable2FAUC:     enable2FAUC,
		Verify2FAUC:     verify2FALoginUC,

		GetCountriesUC:        getCountriesUC,
		GetCountryUC:          getCountryUC,
		GetCitiesUC:           getCitiesUC,
		GetCityUC:             getCityUC,
		SearchCitiesUC:        searchCitiesUC,
		GetDistrictsUC:        getDistrictsUC,
		GetTransportNodesUC:   getTransportNodesUC,
		GetPlaceUC:            getPlaceUC,
		GetPlacesByCityUC:     getPlacesByCityUC,
		GetPlacesByCategoryUC: getPlacesByCategoryUC,
		GetHolidaysUC:         getHolidays,

		CreateTripUC:          createTripUC,
		GetTripUC:             getTripUC,
		GetUserTripsUC:        getUserTripsUC,
		AddPlaceToTripUC:      addPlaceToTripUC,
		RemovePlaceFromTripUC: removePlaceFromTripUC,
		UpdateTripUC:          updateTripUC,
		DeleteTripUC:          deleteTripUC,

		CreateReviewUC:      createReviewUC,
		GetReviewsByPlaceUC: getReviewsByPlaceUC,
		GetUserReviewsUC:    getUserReviewsUC,
		ModerateReviewUC:    moderateReviewUC,
		DeleteReviewUC:      deleteReviewUC,

		GetStaticPageUC: getStaticPageUC,

		AddFavUC:    addFavUC,
		RemoveFavUC: removeFavUC,
		GetFavUC:    getFavUC,

		GetUsersUC:          getUsersUC,
		GetPendingReviewsUC: getPendingReviewsUC,

		GenerateReportUC: generateReportUC,
	}

	if flag.NArg() > 0 {
		cli.RunCommandLine(flag.Args())
		return
	}
	if interactive {
		cli.RunInteractive()
		return
	}

	server := delivery.NewServer(
		registerUC, verifyContactUC, loginUC, logoutUC,
		enable2FAUC, verify2FALoginUC, getProfileUC, updateProfileUC, uploadAvatarUC,
		moderateUserUC,
		getCountriesUC, getCountryUC, getCitiesUC, getCityUC, searchCitiesUC,
		getDistrictsUC, getTransportNodesUC, getPlacesByCityUC, getPlacesByCategoryUC, getPlaceUC,
		createTripUC, getTripUC, getUserTripsUC, updateTripUC, deleteTripUC,
		addPlaceToTripUC, removePlaceFromTripUC,
		createReviewUC, getReviewsByPlaceUC, getUserReviewsUC, deleteReviewUC, moderateReviewUC,
		getStaticPageUC, generateReportUC, getUsersUC, getPendingReviewsUC,
		addFavUC, removeFavUC, getFavUC, getHolidays,
		cfg,
	)

	jwtSecret := []byte(cfg.Database.Auth.JWTSecret)
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("your-secret-key")
	}
	router := delivery.NewRouter(server, gqlResolver, jwtSecret)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		slog.Error("server forced shutdown", "error", err)
	}

	slog.Info("server stopped")
}

func runMigrationsOnly(cfg *config.Config) error {
	switch cfg.Database.PrimaryType {
	case "postgres":
		connStr := cfg.PostgresConnString()
		db, err := sql.Open("pgx", connStr)
		if err != nil {
			return fmt.Errorf("failed to open postgres: %w", err)
		}
		defer db.Close()
		if err := goose.Up(db, "migrations"); err != nil {
			return fmt.Errorf("postgres migrations failed: %w", err)
		}
		slog.Info("PostgreSQL migrations applied")
		return nil

	case "cassandra":
		cluster := gocql.NewCluster(cfg.Database.Cassandra.Hosts...)
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
			return fmt.Errorf("failed to create cassandra session: %w", err)
		}
		defer session.Close()

		keyspace := cfg.Database.Cassandra.Keyspace
		createKeyspace := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};`, keyspace)
		if err := session.Query(createKeyspace).Exec(); err != nil {
			return fmt.Errorf("failed to create keyspace: %w", err)
		}
		slog.Info("Keyspace ensured", "keyspace", keyspace)

		clusterWithKeyspace := gocql.NewCluster(cfg.Database.Cassandra.Hosts...)
		clusterWithKeyspace.Keyspace = keyspace
		clusterWithKeyspace.Consistency = gocql.ParseConsistency(cfg.Database.Cassandra.Consistency)
		if cfg.Database.Cassandra.Username != "" {
			clusterWithKeyspace.Authenticator = gocql.PasswordAuthenticator{
				Username: cfg.Database.Cassandra.Username,
				Password: cfg.Database.Cassandra.Password,
			}
		}
		clusterWithKeyspace.Timeout = 10 * time.Second
		clusterWithKeyspace.ConnectTimeout = 10 * time.Second

		sessionKeyspace, err := clusterWithKeyspace.CreateSession()
		if err != nil {
			return fmt.Errorf("failed to create session with keyspace: %w", err)
		}
		defer sessionKeyspace.Close()

		files, err := os.ReadDir("migrations_cassandra")
		if err != nil {
			return fmt.Errorf("failed to read migrations_cassandra: %w", err)
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".cql") {
				continue
			}
			content, err := os.ReadFile(filepath.Join("migrations_cassandra", file.Name()))
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", file.Name(), err)
			}
			for q := range strings.SplitSeq(string(content), ";") {
				q = strings.TrimSpace(q)
				if q == "" || strings.HasPrefix(strings.ToUpper(q), "USE ") {
					continue
				}
				if err := sessionKeyspace.Query(q).Exec(); err != nil {
					if !strings.Contains(err.Error(), "already exists") {
						return fmt.Errorf("error in %s: %w\nQuery: %s", file.Name(), err, q)
					}
				}
			}
			slog.Info("Executed migration", "file", file.Name())
		}
		slog.Info("Cassandra schema applied")
		return nil

	default:
		return fmt.Errorf("unsupported primary_type: %s", cfg.Database.PrimaryType)
	}
}

func cleanAndSeed(ctx context.Context, cfg *config.Config, repos *repository.Repositories) error {
	switch cfg.Database.PrimaryType {
	case "postgres":
		db := repos.GetPostgresDB()
		if db == nil {
			return fmt.Errorf("postgres connection not available")
		}
		if err := cleanTablesPostgres(ctx, db); err != nil {
			return fmt.Errorf("clean tables failed: %w", err)
		}
	case "cassandra":
		session := repos.GetCassandraSession()
		if session == nil {
			return fmt.Errorf("cassandra session not available")
		}
		if err := cleanTablesCassandra(ctx, session, cfg.Database.Cassandra.Keyspace); err != nil {
			return fmt.Errorf("clean tables failed: %w", err)
		}
	}
	return seedWithRepos(ctx, repos, repos.MinioClient, cfg.Database.Minio.Bucket)
}

func cleanTablesPostgres(ctx context.Context, db *sql.DB) error {
	tables := []string{
		"TripPlace", "Review", "Trip", "Place", "CityDistrict",
		"TransportNode", "City", "Holiday", "Country", "User", "StaticPage",
	}
	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE \"%s\" RESTART IDENTITY CASCADE;", table))
		if err != nil {
			return fmt.Errorf("truncate %s: %w", table, err)
		}
	}
	slog.Info("PostgreSQL tables truncated")
	return nil
}

func cleanTablesCassandra(ctx context.Context, session *gocql.Session, keyspace string) error {
	tables := []string{
		"trip_place_by_trip",
		"review_by_place", "review_by_user", "review_pending",
		"trip_by_id", "trip_by_user",
		"place_by_id", "place_by_city", "place_by_category",
		"district_by_city",
		"transport_node_by_city",
		"city_by_id", "city_by_country", "city_by_name",
		"holiday_by_country", "holiday_by_date",
		"country_by_id", "country_by_name",
		"user_by_id", "user_by_username", "user_by_email",
		"static_page_by_id", "static_page_by_slug",
		"favourite_by_user",
	}
	for _, table := range tables {
		if err := session.Query(fmt.Sprintf("TRUNCATE %s.%s;", keyspace, table)).Exec(); err != nil {
			slog.Warn("truncate failed (table may not exist)", "table", table, "error", err)
		}
	}
	slog.Info("Cassandra tables truncated")
	return nil
}

func runMigrationsAndSeed(ctx context.Context, cfg *config.Config, repos *repository.Repositories, minioClient *minio.Client) error {
	switch cfg.Database.PrimaryType {
	case "postgres":
		db := repos.GetPostgresDB()
		if db == nil {
			return fmt.Errorf("postgres connection not available")
		}
		if err := goose.Up(db, "migrations"); err != nil {
			return fmt.Errorf("postgres migrations failed: %w", err)
		}
		slog.Info("PostgreSQL migrations applied")

		if err := cleanTablesPostgres(ctx, db); err != nil {
			return fmt.Errorf("clean tables failed: %w", err)
		}

		if err := seedWithRepos(ctx, repos, minioClient, cfg.Database.Minio.Bucket); err != nil {
			return fmt.Errorf("seed failed: %w", err)
		}
		return nil

	case "cassandra":
		session := repos.GetCassandraSession()
		if session == nil {
			return fmt.Errorf("cassandra session not available")
		}
		keyspace := cfg.Database.Cassandra.Keyspace
		if err := applyCassandraMigrations(cfg); err != nil {
			return fmt.Errorf("cassandra migrations failed: %w", err)
		}
		slog.Info("Cassandra schema applied")

		if err := cleanTablesCassandra(ctx, session, keyspace); err != nil {
			return fmt.Errorf("clean tables failed: %w", err)
		}

		if err := seedWithRepos(ctx, repos, minioClient, cfg.Database.Minio.Bucket); err != nil {
			return fmt.Errorf("seed failed: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported primary_type: %s", cfg.Database.PrimaryType)
	}
}

func applyCassandraMigrations(cfg *config.Config) error {
	clusterNoKeyspace := gocql.NewCluster(cfg.Database.Cassandra.Hosts...)
	clusterNoKeyspace.Consistency = gocql.ParseConsistency(cfg.Database.Cassandra.Consistency)
	if cfg.Database.Cassandra.Username != "" {
		clusterNoKeyspace.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Database.Cassandra.Username,
			Password: cfg.Database.Cassandra.Password,
		}
	}
	clusterNoKeyspace.Timeout = 10 * time.Second
	clusterNoKeyspace.ConnectTimeout = 10 * time.Second

	sessionNoKeyspace, err := clusterNoKeyspace.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create temporary session: %w", err)
	}
	defer sessionNoKeyspace.Close()

	keyspace := cfg.Database.Cassandra.Keyspace

	createKeyspaceQuery := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};`, keyspace)
	if err := sessionNoKeyspace.Query(createKeyspaceQuery).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace '%s': %w", keyspace, err)
	}
	slog.Info("Keyspace ensured", "keyspace", keyspace)

	cluster := gocql.NewCluster(cfg.Database.Cassandra.Hosts...)
	cluster.Keyspace = keyspace
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
		return fmt.Errorf("failed to create session for keyspace '%s': %w", keyspace, err)
	}
	defer session.Close()

	files, err := os.ReadDir("migrations_cassandra")
	if err != nil {
		return fmt.Errorf("failed to read migrations_cassandra directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".cql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join("migrations_cassandra", file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		queries := strings.SplitSeq(string(content), ";")
		for q := range queries {
			q = strings.TrimSpace(q)
			if q == "" {
				continue
			}
			if err := session.Query(q).Exec(); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return fmt.Errorf("failed to execute query from %s: %w\nQuery: %s", file.Name(), err, q)
				}
				slog.Warn("Table already exists, skipping", "query", q)
			}
		}
		slog.Info("Executed migration", "file", file.Name())
	}

	return nil
}

func seedWithRepos(ctx context.Context, repos *repository.Repositories, minioClient *minio.Client, bucketName string) error {
	logger := slog.Default()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		slog.Info("Created bucket", "bucket", bucketName)
	}

	uploadImageAndCreateStaticPage := func(path, contentType, entityType, entityName string) (domain.ImageID, error) {
		file, err := os.Open(path)
		if err != nil {
			return domain.ImageID(uuid.Nil), err
		}
		defer file.Close()

		imageUUID := uuid.New()
		objectName := imageUUID.String()

		_, err = minioClient.PutObject(ctx, bucketName, objectName, file, -1, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			return domain.ImageID(uuid.Nil), err
		}
		fmt.Printf("Uploaded %s to MinIO: %s\n", entityName, objectName)

		staticPage, err := domain.NewStaticPage(
			domain.StaticPageID(imageUUID),
			"img-"+imageUUID.String(),
			entityType+": "+entityName,
			"Image page for "+entityType+" "+entityName,
			"",
			"",
			"",
		)
		if err != nil {
			fmt.Printf("WARN: failed to create StaticPage object for %s: %v\n", entityName, err)
			return domain.ImageID(uuid.Nil), err
		}
		if err := repos.StaticRepo.Save(ctx, staticPage, nil); err != nil {
			fmt.Printf("WARN: failed to save StaticPage for %s: %v\n", entityName, err)
		} else {
			fmt.Printf("StaticPage record created for %s (ID: %s)\n", entityName, imageUUID)
		}
		return domain.ImageID(imageUUID), nil
	}

	russia := domain.NewCountryFromDB(
		domain.CountryID(uuid.New()),
		"Russia",
		17100000.0,
		2.5e12,
		144000000,
		"RUB",
		"Visa required for some countries",
		"Largest country in the world",
		"Be careful in big cities",
		"June-August",
		"Russian",
		"+7",
		"Orthodox Christianity",
		domain.CityID(uuid.Nil),
		domain.ImageID(uuid.Nil),
	)
	fmt.Println("russia", russia)
	if err := repos.CountryRepo.Save(ctx, russia); err != nil {
		return fmt.Errorf("save country: %w", err)
	}
	fmt.Printf("Created country: %s (ID: %s)\n", russia.GetName(), russia.GetId())

	imgID, err := uploadImageAndCreateStaticPage("images/a.jpg", "image/jpeg", "Country", russia.GetName())
	if err == nil {
		russia.SetImageID(imgID)
		if err := repos.CountryRepo.Update(ctx, russia); err != nil {
			logger.Warn("failed to update country image", "error", err)
		}
	} else {
		logger.Warn("failed to upload country image", "error", err)
	}

	moscowCoords := domain.NewCoordinates(55.7558, 37.6173)
	moscow := domain.NewCityFromDB(
		domain.CityID(uuid.New()),
		"Moscow",
		12500000,
		true,
		moscowCoords,
		"Capital of Russia",
		"UTC+3",
		"Visit Red Square",
		russia.GetId(),
		domain.ImageID(uuid.Nil),
	)
	if err := repos.CityRepo.Save(ctx, moscow); err != nil {
		return fmt.Errorf("save city: %w", err)
	}
	fmt.Printf("Created city: %s (ID: %s)\n", moscow.GetName(), moscow.GetId())

	imgID, err = uploadImageAndCreateStaticPage("images/b.jpg", "image/jpeg", "City", moscow.GetName())
	if err == nil {
		moscow.SetImageId(imgID)
		if err := repos.CityRepo.Update(ctx, moscow); err != nil {
			logger.Warn("failed to update city image", "error", err)
		}
	} else {
		logger.Warn("failed to upload Moscow image", "error", err)
	}

	russia.SetCapitalID(moscow.GetId())
	if err := repos.CountryRepo.Update(ctx, russia); err != nil {
		logger.Warn("failed to update country capital", "error", err)
	}

	spbCoords := domain.NewCoordinates(59.9343, 30.3351)
	petersburg := domain.NewCityFromDB(
		domain.CityID(uuid.New()),
		"Saint Petersburg",
		5400000,
		false,
		spbCoords,
		"Cultural capital",
		"UTC+3",
		"Visit Hermitage",
		russia.GetId(),
		domain.ImageID(uuid.Nil),
	)
	if err := repos.CityRepo.Save(ctx, petersburg); err != nil {
		return fmt.Errorf("save city: %w", err)
	}
	fmt.Printf("Created city: %s (ID: %s)\n", petersburg.GetName(), petersburg.GetId())

	imgID, err = uploadImageAndCreateStaticPage("images/c.png", "image/png", "City", petersburg.GetName())
	if err == nil {
		petersburg.SetImageId(imgID)
		if err := repos.CityRepo.Update(ctx, petersburg); err != nil {
			logger.Warn("failed to update city image", "error", err)
		}
	} else {
		logger.Warn("failed to upload Saint Petersburg image", "error", err)
	}

	redSquareCoords := domain.NewCoordinates(55.7537, 37.6212)
	redSquare := domain.NewPlaceFromDB(
		domain.PlaceID(uuid.New()),
		"Red Square",
		"attraction",
		"Famous square in Moscow",
		redSquareCoords,
		"Red Square, Moscow",
		"Always open",
		"Free",
		60,
		4.8,
		1500,
		"",
		"https://kreml.ru/ru",
		moscow.GetId(),
		domain.CityDistrictID(uuid.Nil),
		domain.ImageID(uuid.Nil),
	)
	if err := repos.PlaceRepo.Save(ctx, redSquare); err != nil {
		return fmt.Errorf("save place: %w", err)
	}
	fmt.Printf("Created place: %s (ID: %s)\n", redSquare.GetName(), redSquare.GetId())

	hermitageCoords := domain.NewCoordinates(59.9405, 30.3149)
	hermitage := domain.NewPlaceFromDB(
		domain.PlaceID(uuid.New()),
		"Hermitage Museum",
		"museum",
		"World-famous museum",
		hermitageCoords,
		"Palace Square, 2",
		"10:30-18:00",
		"500 RUB",
		180,
		4.9,
		2000,
		"+7-812-710-90-79",
		"https://hermitagemuseum.org",
		petersburg.GetId(),
		domain.CityDistrictID(uuid.Nil),
		domain.ImageID(uuid.Nil),
	)
	if err := repos.PlaceRepo.Save(ctx, hermitage); err != nil {
		return fmt.Errorf("save place: %w", err)
	}
	fmt.Printf("Created place: %s (ID: %s)\n", hermitage.GetName(), hermitage.GetId())

	testUserID := domain.UserID(uuid.New())
	testUserEmail := "testuser@example.com"
	testUserPass := "TestPass123!"
	hashedTestPass, _ := domain.HashPassword(testUserPass)
	testUser := domain.NewUserFromDB(
		testUserID, "testuser", testUserEmail, hashedTestPass, "+79991234567", time.Now(), nil, "Russia", "", []string{},
		false, nil, "", domain.UserRoleUser, true, true, false, "", nil,
	)
	if err := repos.UserRepo.Save(ctx, testUser); err != nil {
		return fmt.Errorf("save test user: %w", err)
	}
	if err := repos.UserRepo.UpdateEmailVerified(ctx, testUser.GetId(), true); err != nil {
		return fmt.Errorf("verify test user email: %w", err)
	}

	savedUser, err := repos.UserRepo.FindByUsername(ctx, "testuser")
	if err != nil {
		return fmt.Errorf("find test user after verification: %w", err)
	}
	if !savedUser.IsEmailVerified() {
		return fmt.Errorf("test user email is NOT verified after UpdateEmailVerified — check repository implementation")
	}
	fmt.Printf("Created test user: testuser / TestPass123! (ID: %s)\n", testUserID)

	testUserID = domain.UserID(uuid.New())
	testUserEmail = "testuser1@example.com"
	testUserPass = "TestPass123!"
	hashedTestPass, _ = domain.HashPassword(testUserPass)
	testUser = domain.NewUserFromDB(
		testUserID, "testuser1", testUserEmail, hashedTestPass, "+79991234567", time.Now(), nil, "Russia", "", []string{},
		false, nil, "", domain.UserRoleUser, true, true, false, "", nil,
	)
	if err := repos.UserRepo.Save(ctx, testUser); err != nil {
		return fmt.Errorf("save test user: %w", err)
	}

	moderatorID := domain.UserID(uuid.New())
	moderatorEmail := "moderator@example.com"
	moderatorPass := "Moderator123!"
	hashedPass, _ := domain.HashPassword(moderatorPass)
	moderator := domain.NewUserFromDB(
		moderatorID, "moderator", moderatorEmail, hashedPass, "", time.Now(), nil, "", "", []string{},
		false, nil, "", domain.UserRoleModerator, false, false, false, "", nil,
	)
	if err := repos.UserRepo.Save(ctx, moderator); err != nil {
		return fmt.Errorf("save moderator: %w", err)
	}
	if err := repos.UserRepo.UpdateEmailVerified(ctx, moderatorID, true); err != nil {
		logger.Error("failed to verify moderator email", "error", err)
	}
	enable2FAForUser(ctx, repos.UserRepo, moderatorID, moderatorEmail)

	analystID := domain.UserID(uuid.New())
	analystEmail := "analyst@example.com"
	analystPass := "Analyst123!"
	hashedPass2, _ := domain.HashPassword(analystPass)
	analyst := domain.NewUserFromDB(
		analystID, "analyst", analystEmail, hashedPass2, "", time.Now(), nil, "", "", []string{},
		false, nil, "", domain.UserRoleAnalyst, false, false, false, "", nil,
	)
	if err := repos.UserRepo.Save(ctx, analyst); err != nil {
		return fmt.Errorf("save analyst: %w", err)
	}
	if err := repos.UserRepo.UpdateEmailVerified(ctx, analystID, true); err != nil {
		logger.Warn("failed to verify analyst email", "error", err)
	}
	enable2FAForUser(ctx, repos.UserRepo, analystID, analystEmail)

	fmt.Println("Test data seeded successfully!")
	return nil
}

func enable2FAForUser(ctx context.Context, userRepo interfaces.UserRepository, userID domain.UserID, email string) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GeoGuide",
		AccountName: email,
	})
	if err != nil {
		slog.Error("failed to generate TOTP secret", "user", email, "error", err)
		return
	}
	backupCodes := generateBackupCodes(10)
	fmt.Printf("Backup codes for %s: %v\n", email, backupCodes)
	hashedBackups := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedBackups[i], _ = domain.HashPassword(code)
	}
	if err := userRepo.EnableTwoFactor(ctx, userID, key.Secret(), hashedBackups); err != nil {
		slog.Error("failed to enable 2FA", "user", email, "error", err)
		return
	}
	fmt.Printf("2FA enabled for %s (secret: %s)\n", email, key.Secret())
}

func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := range count {
		codes[i] = generateRandomCode(8)
	}
	return codes
}

func generateRandomCode(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}

// =========================================================
//                            CLI
// =========================================================

type CLI struct {
	ctx    context.Context
	logger *slog.Logger

	RegisterUC      *auth.RegisterUserUseCase
	LoginUC         *auth.LoginUseCase
	UpdateProfileUC *auth.UpdateUserProfileUseCase
	GetProfileUC    *auth.GetUserProfileUseCase
	LogoutUC        *auth.LogoutUseCase
	ModerateUserUC  *auth.ModerateUserUsecase
	VerifyContactUC *auth.VerifyContactUseCase
	Enable2FAUC     *auth.EnableTwoFactorUseCase
	Verify2FAUC     *auth.VerifyTwoFactorLoginUseCase
	UploadAvatarUC  *auth.UploadAvatarUseCase

	GetCountriesUC        *catalog.GetCountriesUseCase
	GetCountryUC          *catalog.GetCountryUseCase
	GetCitiesUC           *catalog.GetCitiesByCountryUseCase
	GetCityUC             *catalog.GetCityUseCase
	SearchCitiesUC        *catalog.SearchCitiesUseCase
	GetDistrictsUC        *catalog.GetDistrictsByCityUseCase
	GetTransportNodesUC   *catalog.GetTransportNodesByCityUseCase
	GetPlaceUC            *catalog.GetPlaceUseCase
	GetPlacesByCityUC     *catalog.GetPlacesByCityUseCase
	GetPlacesByCategoryUC *catalog.GetPlacesByCategoryUseCase
	GetHolidaysUC         *catalog.GetHolidaysByDateUseCase

	CreateTripUC          *planner.CreateTripUseCase
	GetTripUC             *planner.GetTripUseCase
	GetUserTripsUC        *planner.GetUserTripsUseCase
	AddPlaceToTripUC      *planner.AddPlaceToTripUseCase
	RemovePlaceFromTripUC *planner.RemovePlaceFromTripUseCase
	UpdateTripUC          *planner.UpdateTripUseCase
	DeleteTripUC          *planner.DeleteTripUseCase

	CreateReviewUC      *review.CreateReviewUseCase
	GetReviewsByPlaceUC *review.GetReviewsByPlaceUseCase
	GetUserReviewsUC    *review.GetUserReviewsUseCase
	ModerateReviewUC    *review.ModerateReviewUseCase
	DeleteReviewUC      *review.DeleteReviewUseCase

	GetStaticPageUC *content.GetStaticPageByIDUseCase

	AddFavUC    *favourites.AddFavouriteUseCase
	RemoveFavUC *favourites.RemoveFavouriteUseCase
	GetFavUC    *favourites.GetFavouritesUseCase

	GetUsersUC          *admin.GetUsersUseCase
	GetPendingReviewsUC *admin.GetPendingReviewsUseCase

	GenerateReportUC *analytics.GenerateReportUseCase
}

func (c *CLI) RunInteractive() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("GeoGuide CLI (interactive mode). Type 'help' for commands, 'exit' to quit.")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "input error: %v\n", err)
			}
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		cmd, args := parseCommandLine(line)
		if cmd == "exit" || cmd == "quit" {
			return
		}
		c.dispatch(cmd, args)
	}
}

func (c *CLI) RunCommandLine(args []string) {
	if len(args) == 0 {
		fmt.Println("No command provided. Use -i for interactive mode or 'help'.")
		os.Exit(1)
	}
	c.dispatch(args[0], args[1:])
}

func (c *CLI) dispatch(cmd string, args []string) {
	switch cmd {
	case "":
		return
	case "help":
		printHelp()

	case "register":
		c.cmdRegister(args)
	case "login":
		c.cmdLogin(args)
	case "verify":
		c.cmdVerify(args)
	case "profile":
		c.cmdProfile(args)
	case "update_profile":
		c.cmdUpdateProfile(args)
	case "upload_avatar":
		c.cmdUploadAvatar(args)
	case "logout":
		c.cmdLogout(args)
	case "moderate_user":
		c.cmdModerateUser(args)
	case "enable_2fa":
		c.cmdEnable2FA(args)
	case "verify_2fa":
		c.cmdVerify2FA(args)

	case "countries":
		c.cmdCountries(args)
	case "country":
		c.cmdCountry(args)
	case "cities":
		c.cmdCities(args)
	case "city":
		c.cmdCity(args)
	case "search_cities":
		c.cmdSearchCities(args)
	case "districts":
		c.cmdDistricts(args)
	case "transport_nodes":
		c.cmdTransportNodes(args)
	case "place":
		c.cmdPlace(args)
	case "places_by_city":
		c.cmdPlacesByCity(args)
	case "places_by_category":
		c.cmdPlacesByCategory(args)
	case "holidays":
		c.cmdHolidays(args)

	case "create_trip":
		c.cmdCreateTrip(args)
	case "get_trip":
		c.cmdGetTrip(args)
	case "user_trips":
		c.cmdUserTrips(args)
	case "add_place_to_trip":
		c.cmdAddPlaceToTrip(args)
	case "remove_place_from_trip":
		c.cmdRemovePlaceFromTrip(args)
	case "update_trip":
		c.cmdUpdateTrip(args)
	case "delete_trip":
		c.cmdDeleteTrip(args)

	case "create_review":
		c.cmdCreateReview(args)
	case "reviews_by_place":
		c.cmdReviewsByPlace(args)
	case "user_reviews":
		c.cmdUserReviews(args)
	case "moderate_review":
		c.cmdModerateReview(args)
	case "delete_review":
		c.cmdDeleteReview(args)
	case "pending_reviews":
		c.cmdPendingReviews(args)

	case "get_page":
		c.cmdGetPage(args)

	case "add_favourite":
		c.cmdAddFavourite(args)
	case "remove_favourite":
		c.cmdRemoveFavourite(args)
	case "favourites":
		c.cmdFavourites(args)

	case "users":
		c.cmdUsers(args)

	case "report":
		c.cmdReport(args)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

// ---- AUTH ----

func (c *CLI) cmdRegister(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: register <username> <email> <password> <phone>")
		return
	}
	req := requests.RegisterUserRequest{
		Username: args[0],
		Email:    args[1],
		Password: args[2],
		Phone:    args[3],
	}
	resp, err := c.RegisterUC.Execute(c.ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Registered user: %s (ID: %s). Verification email sent.\n", resp.Username, resp.ID)
}

func (c *CLI) cmdLogin(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: login <username> <password>")
		return
	}
	req := requests.LoginUserRequest{Username: args[0], Password: args[1]}
	resp, err := c.LoginUC.Execute(c.ctx, req, "", "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if resp.RequiresTwoFactor {
		fmt.Printf("Two-factor authentication required. Challenge ID: %s\n", resp.ChallengeID)
		return
	}
	fmt.Printf("Logged in as %s (token: %s)\n", resp.Username, resp.Token)
}

func (c *CLI) cmdVerify(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: verify <user_id> <contact> <code>")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	req := requests.VerifyContactRequest{
		UserID:  domain.UserID(userID),
		Contact: args[1],
		Code:    args[2],
	}
	if err := c.VerifyContactUC.Execute(c.ctx, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Contact verified successfully")
}

func (c *CLI) cmdProfile(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: profile <user_id>")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	resp, err := c.GetProfileUC.Execute(c.ctx, requests.GetUserProfileRequest{UserID: domain.UserID(userID)})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("User profile:\n  ID: %s\n  Username: %s\n  Email: %s\n  Phone: %s\n  Registered: %s\n  Birth Date: %s\n  Country: %s\n  Avatar: %s\n  Categories: %v\n  Role: %s\n  Blocked: %t\n",
		resp.ID, resp.Username, resp.Email, resp.Phone, resp.RegisteredAt, resp.BirthDate,
		resp.CountryOfResidence, resp.AvatarURL, resp.FavoriteCategories, resp.Role, resp.IsBlocked)
}

func (c *CLI) cmdUpdateProfile(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: update_profile <token> field=value [field=value...]")
		fmt.Println("Fields: username, phone, country, avatar_url, birth_date, categories (comma-sep)")
		return
	}
	req := requests.UpdateUserProfileRequest{Token: args[0]}
	for _, kv := range args[1:] {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k, v := parts[0], parts[1]
		switch k {
		case "username":
			req.Username = &v
		case "phone":
			req.Phone = &v
		case "country":
			req.CountryOfResidence = &v
		case "avatar_url":
			req.AvatarURL = &v
		case "birth_date":
			if t, err := time.Parse("2006-01-02", v); err == nil {
				req.BirthDate = &t
			}
		case "categories":
			req.FavoriteCategories = strings.Split(v, ",")
		}
	}
	if _, err := c.UpdateProfileUC.Execute(c.ctx, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Profile updated successfully")
}

func (c *CLI) cmdUploadAvatar(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: upload_avatar <user_id> <file_path>")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}

	path := args[1]
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Cannot open file: %v\n", err)
		return
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		fmt.Printf("Cannot stat file: %v\n", err)
		return
	}
	if st.Size() > 5<<20 {
		fmt.Println("File too large (max 5 MB)")
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	avatarURL, err := c.UploadAvatarUC.Execute(c.ctx, domain.UserID(userID), f, st.Size(), filepath.Base(path), contentType)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Avatar uploaded: %s\n", avatarURL)
}

func (c *CLI) cmdLogout(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: logout <token>")
		return
	}
	resp, err := c.LogoutUC.Execute(c.ctx, requests.LogoutRequest{Token: args[0]})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Logged out: %v\n", resp.Success)
}

func (c *CLI) cmdModerateUser(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: moderate_user <user_id> <true|false> <reason> <moderator_id>")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	block, err := strconv.ParseBool(args[1])
	if err != nil {
		fmt.Println("Second arg must be true|false")
		return
	}
	moderatorID, err := uuid.Parse(args[3])
	if err != nil {
		fmt.Printf("Invalid moderator_id: %v\n", err)
		return
	}
	req := requests.ModerateUserRequest{
		UserID:      domain.UserID(userID),
		Block:       block,
		Reason:      args[2],
		ModeratorID: domain.UserID(moderatorID),
	}
	res, err := c.ModerateUserUC.Execute(c.ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("User moderation applied. IsBlocked=%v Reason=%s\n", res.IsBlocked, res.BlockReason)
}

func (c *CLI) cmdEnable2FA(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  enable_2fa generate <user_id> <email>")
		fmt.Println("  enable_2fa verify <user_id> <secret> <code>")
		return
	}
	switch args[0] {
	case "generate":
		if len(args) < 3 {
			fmt.Println("Usage: enable_2fa generate <user_id> <email>")
			return
		}
		userID, err := uuid.Parse(args[1])
		if err != nil {
			fmt.Printf("Invalid user_id: %v\n", err)
			return
		}
		secret, uri, err := c.Enable2FAUC.GenerateSecret(c.ctx, domain.UserID(userID), args[2])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("TOTP Secret: %s\nOTPAuth URI: %s\n", secret, uri)
	case "verify":
		if len(args) < 4 {
			fmt.Println("Usage: enable_2fa verify <user_id> <secret> <code>")
			return
		}
		userID, err := uuid.Parse(args[1])
		if err != nil {
			fmt.Printf("Invalid user_id: %v\n", err)
			return
		}
		codes, err := c.Enable2FAUC.VerifyAndEnable(c.ctx, requests.EnableTwoFactorRequest{
			UserID: domain.UserID(userID),
			Secret: args[2],
			Code:   args[3],
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("2FA enabled. Backup codes (save them!):")
		for i, bc := range codes {
			fmt.Printf("  %d. %s\n", i+1, bc)
		}
	default:
		fmt.Println("Unknown enable_2fa subcommand")
	}
}

func (c *CLI) cmdVerify2FA(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: verify_2fa <challenge_id> <code>")
		return
	}
	token, err := c.Verify2FAUC.Execute(c.ctx, requests.VerifyTwoFactorRequest{
		ChallengeID: args[0],
		Code:        args[1],
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("2FA successful, session token: %s\n", token)
}

// ---- CATALOG ----

func (c *CLI) cmdCountries(args []string) {
	limit, offset := 100, 0
	if len(args) >= 1 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			limit = n
		}
	}
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetCountriesUC.Execute(c.ctx, requests.GetCountriesRequest{Limit: limit, Offset: offset})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Found %d countries:\n", len(resp))
	for _, co := range resp {
		fmt.Printf("  ID: %s | Name: %s | Capital: %s | Area: %.2f | Pop: %d | Currency: %s | Language: %s | Phone: %s\n",
			co.ID, co.Name, co.Capital, co.Area, co.Population, co.Currency, co.Language, co.PhoneCode)
	}
}

func (c *CLI) cmdCountry(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: country <id>")
		return
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		return
	}
	resp, err := c.GetCountryUC.Execute(c.ctx, requests.GetCountryRequest{ID: domain.CountryID(id)})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Country details:\n  ID: %s\n  Name: %s\n  Capital: %s (ID: %s)\n  Area: %.2f km²\n  Population: %d\n  GDP: %.2f\n  Currency: %s\n  Visa: %s\n  Description: %s\n  Safety: %s\n  Best season: %s\n  Language: %s\n  Phone: %s\n  Religion: %s\n  Image: %s\n  Holidays: %d\n",
		resp.ID, resp.Name, resp.CapitalName, resp.CapitalID, resp.Area, resp.Population, resp.GDP, resp.Currency,
		resp.VisaRequirements, resp.Description, resp.SafetyTips, resp.BestSeason, resp.Language, resp.PhoneCode,
		resp.Religion, resp.ImageURL, len(resp.Holidays))
	for i, h := range resp.Holidays {
		fmt.Printf("    %d. %s (%s) national=%t: %s\n", i+1, h.Name, h.Date.Format("2006-01-02"), h.IsNational, h.Description)
	}
}

func (c *CLI) cmdCities(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cities <country_id> [limit] [offset]")
		return
	}
	countryID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		return
	}
	limit, offset := 100, 0
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			limit = n
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetCitiesUC.Execute(c.ctx, requests.GetCitiesByCountryRequest{
		CountryID: domain.CountryID(countryID), Limit: limit, Offset: offset,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Cities in country:")
	for _, city := range resp {
		fmt.Printf("  ID: %s | Name: %s | Population: %d | Capital: %t | Timezone: %s | Coords: %s\n",
			city.ID, city.Name, city.Population, city.IsCapital, city.Timezone, city.Coordinates)
	}
}

func (c *CLI) cmdCity(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: city <id>")
		return
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		return
	}
	resp, err := c.GetCityUC.Execute(c.ctx, requests.GetCityRequest{ID: domain.CityID(id)})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("City:\n  ID: %s\n  Name: %s\n  Population: %d\n  Capital: %t\n  Coordinates: %.4f, %.4f\n  Description: %s\n  Timezone: %s\n  Travel tips: %s\n  Country ID: %s\n  Image: %s\n",
		resp.ID, resp.Name, resp.Population, resp.IsCapital,
		resp.Coordinates.Lat(), resp.Coordinates.Lng(),
		resp.Description, resp.Timezone, resp.TravelTips, resp.CountryID, resp.ImageURL)
}

func (c *CLI) cmdSearchCities(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: search_cities <query> [limit] [offset]")
		return
	}
	limit, offset := 20, 0
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			limit = n
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			offset = n
		}
	}
	resp, err := c.SearchCitiesUC.Execute(c.ctx, requests.SearchCitiesRequest{Query: args[0], Limit: limit, Offset: offset})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, city := range resp {
		fmt.Printf("  ID: %s | Name: %s | Population: %d | Coords: %s\n",
			city.ID, city.Name, city.Population, city.Coordinates)
	}
}

func (c *CLI) cmdDistricts(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: districts <city_id> [limit] [offset]")
		return
	}
	cityID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		return
	}
	limit, offset := 100, 0
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			limit = n
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetDistrictsUC.Execute(c.ctx, requests.GetDistrictsByCityRequest{
		CityID: domain.CityID(cityID), Limit: limit, Offset: offset,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, d := range resp {
		fmt.Printf("  ID: %s | Name: %s | Coords: %.4f,%.4f | City: %s\n",
			d.ID, d.Name, d.Coordinates.Lat(), d.Coordinates.Lng(), d.CityID)
	}
}

func (c *CLI) cmdTransportNodes(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: transport_nodes <city_id> [limit] [offset]")
		return
	}
	cityID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		return
	}
	limit, offset := 100, 0
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			limit = n
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetTransportNodesUC.Execute(c.ctx, requests.GetTransportNodesByCityRequest{
		CityID: domain.CityID(cityID), Limit: limit, Offset: offset,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, n := range resp {
		fmt.Printf("  ID: %s | Type: %s | Name: %s | Address: %s | Coords: %.4f,%.4f\n",
			n.ID, n.NodeType, n.Name, n.Address, n.Coordinates.Lat(), n.Coordinates.Lng())
	}
}

func (c *CLI) cmdPlace(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: place <id>")
		return
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		return
	}
	resp, err := c.GetPlaceUC.Execute(c.ctx, requests.GetPlaceRequest{ID: domain.PlaceID(id)})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Place details:\n  ID: %s\n  Name: %s\n  Category: %s\n  Description: %s\n  Coordinates: %s\n  Address: %s\n  Hours: %s\n  Price: %s\n  Duration: %d min\n  Rating: %.2f (%d reviews)\n  Phone: %s\n  Website: %s\n  City: %s (ID: %s)\n  District: %s (ID: %s)\n  Image: %s\n  Reviews: %d\n",
		resp.ID, resp.Name, resp.Category, resp.Description, resp.Coordinates, resp.Address, resp.OpeningHours,
		resp.PriceInfo, resp.AvgVisitDurationMin, resp.AvgRating, resp.ReviewsCount, resp.ContactPhone, resp.Website,
		resp.CityName, resp.CityID, resp.DistrictName, resp.DistrictID, resp.ImageURL, len(resp.Reviews))
	for i, r := range resp.Reviews {
		fmt.Printf("    %d. Rating: %d | User: %s | Date: %s | Approved: %t | Comment: %s\n",
			i+1, r.Rating, r.Username, r.VisitDate, r.IsApproved, r.Comment)
	}
}

func (c *CLI) cmdPlacesByCity(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: places_by_city <city_id> [category] [limit] [offset] [sortBy]")
		return
	}
	cityID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		return
	}
	req := requests.GetPlacesByCityRequest{
		CityID: domain.CityID(cityID),
		Limit:  100,
	}
	if len(args) >= 2 {
		req.Category = args[1]
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			req.Limit = n
		}
	}
	if len(args) >= 4 {
		if n, err := strconv.Atoi(args[3]); err == nil {
			req.Offset = n
		}
	}
	if len(args) >= 5 {
		req.SortBy = args[4]
	}
	resp, err := c.GetPlacesByCityUC.Execute(c.ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, p := range resp {
		fmt.Printf("  ID: %s | Name: %s | Category: %s | Rating: %.2f | Reviews: %d\n",
			p.ID, p.Name, p.Category, p.AvgRating, p.ReviewsCount)
	}
}

func (c *CLI) cmdPlacesByCategory(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: places_by_category <category> [limit] [offset]")
		return
	}
	limit, offset := 100, 0
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			limit = n
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetPlacesByCategoryUC.Execute(c.ctx, requests.GetPlacesByCategoryRequest{
		Category: args[0], Limit: limit, Offset: offset,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, p := range resp {
		fmt.Printf("  ID: %s | Name: %s | Rating: %.2f | Reviews: %d\n",
			p.ID, p.Name, p.AvgRating, p.ReviewsCount)
	}
}

func (c *CLI) cmdHolidays(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: holidays <YYYY-MM-DD>")
		return
	}
	date, err := time.Parse("2006-01-02", args[0])
	if err != nil {
		fmt.Printf("Invalid date: %v\n", err)
		return
	}
	resp, err := c.GetHolidaysUC.Execute(c.ctx, requests.GetHolidaysByDateRequest{Date: date})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, h := range resp {
		fmt.Printf("  %s | Date: %s | National: %t | %s\n",
			h.Name, h.Date.Format("2006-01-02"), h.IsNational, h.Description)
	}
}

// ---- PLANNER ----

func (c *CLI) cmdCreateTrip(args []string) {
	if len(args) < 5 {
		fmt.Println(`Usage: create_trip <title> <start_date> <end_date> <budget> <user_id> [notes]`)
		return
	}
	startDate, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		fmt.Printf("Invalid start_date: %v\n", err)
		return
	}
	endDate, err := time.Parse("2006-01-02", args[2])
	if err != nil {
		fmt.Printf("Invalid end_date: %v\n", err)
		return
	}
	budget, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		fmt.Printf("Invalid budget: %v\n", err)
		return
	}
	userID, err := uuid.Parse(args[4])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	req := requests.CreateTripRequest{
		Title:     args[0],
		StartDate: startDate,
		EndDate:   endDate,
		Budget:    budget,
		UserID:    domain.UserID(userID),
	}
	if len(args) >= 6 {
		req.Notes = args[5]
	}
	resp, err := c.CreateTripUC.Execute(c.ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Trip created:\n  ID: %s\n  Title: %s\n  Dates: %s – %s\n  Budget: %.2f\n  Status: %s\n",
		resp.ID, resp.Title, resp.StartDate, resp.EndDate, resp.Budget, resp.Status)
}

func (c *CLI) cmdGetTrip(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: get_trip <trip_id> <user_id>")
		return
	}
	tripID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid trip_id: %v\n", err)
		return
	}
	userID, err := uuid.Parse(args[1])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	resp, err := c.GetTripUC.Execute(c.ctx, requests.GetTripRequest{
		TripID: domain.TripID(tripID),
		UserID: domain.UserID(userID),
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Trip:\n  ID: %s\n  Title: %s\n  Dates: %s – %s\n  Budget: %.2f\n  Status: %s\n  Notes: %s\n  Places: %d\n",
		resp.ID, resp.Title, resp.StartDate, resp.EndDate, resp.Budget, resp.Status, resp.Notes, len(resp.Places))
	for i, p := range resp.Places {
		fmt.Printf("    %d. Day %d | Place ID: %s | Arrival: %s | Duration: %d min | Status: %s | Cost: %.2f | Notes: %s\n",
			i+1, p.DayNumber, p.PlaceID, p.ArrivalTime, p.DurationMin, p.VisitStatus, p.ActualCost, p.Notes)
	}
}

func (c *CLI) cmdUserTrips(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: user_trips <user_id>")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	resp, err := c.GetUserTripsUC.Execute(c.ctx, requests.GetUserTripsRequest{UserID: domain.UserID(userID)})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, t := range resp {
		fmt.Printf("  ID: %s | %s | %s – %s | %s\n", t.ID, t.Title, t.StartDate, t.EndDate, t.Status)
	}
}

func (c *CLI) cmdAddPlaceToTrip(args []string) {
	if len(args) < 5 {
		fmt.Println("Usage: add_place_to_trip <trip_id> <place_id> <day_number> <duration_min> <user_id> [arrival_RFC3339] [notes]")
		return
	}
	tripID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid trip_id: %v\n", err)
		return
	}
	placeID, err := uuid.Parse(args[1])
	if err != nil {
		fmt.Printf("Invalid place_id: %v\n", err)
		return
	}
	day, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Invalid day: %v\n", err)
		return
	}
	dur, err := strconv.Atoi(args[3])
	if err != nil {
		fmt.Printf("Invalid duration: %v\n", err)
		return
	}
	userID, err := uuid.Parse(args[4])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	req := requests.AddPlaceToTripRequest{
		TripID:      domain.TripID(tripID),
		PlaceID:     domain.PlaceID(placeID),
		DayNumber:   day,
		DurationMin: dur,
		UserID:      domain.UserID(userID),
	}
	if len(args) >= 6 && args[5] != "" {
		if t, err := time.Parse(time.RFC3339, args[5]); err == nil {
			req.ArrivalTime = &t
		}
	}
	if len(args) >= 7 {
		req.Notes = args[6]
	}
	if err := c.AddPlaceToTripUC.Execute(c.ctx, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Place added to trip successfully")
}

func (c *CLI) cmdRemovePlaceFromTrip(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: remove_place_from_trip <trip_id> <place_id> <user_id>")
		return
	}
	tripID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid trip_id: %v\n", err)
		return
	}
	placeID, err := uuid.Parse(args[1])
	if err != nil {
		fmt.Printf("Invalid place_id: %v\n", err)
		return
	}
	userID, err := uuid.Parse(args[2])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	req := requests.RemovePlaceFromTripRequest{
		TripID:  domain.TripID(tripID),
		PlaceID: domain.PlaceID(placeID),
		UserID:  domain.UserID(userID),
	}
	if err := c.RemovePlaceFromTripUC.Execute(c.ctx, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Place removed from trip successfully")
}

func (c *CLI) cmdUpdateTrip(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: update_trip <trip_id> <user_id> [title] [start_date] [end_date] [budget] [status] [notes]")
		return
	}
	tripID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid trip_id: %v\n", err)
		return
	}
	userID, err := uuid.Parse(args[1])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	req := requests.UpdateTripRequest{
		TripID: domain.TripID(tripID),
		UserID: domain.UserID(userID),
	}
	if len(args) > 2 && args[2] != "" {
		req.Title = &args[2]
	}
	if len(args) > 3 && args[3] != "" {
		if t, err := time.Parse("2006-01-02", args[3]); err == nil {
			req.StartDate = &t
		}
	}
	if len(args) > 4 && args[4] != "" {
		if t, err := time.Parse("2006-01-02", args[4]); err == nil {
			req.EndDate = &t
		}
	}
	if len(args) > 5 && args[5] != "" {
		if b, err := strconv.ParseFloat(args[5], 64); err == nil {
			req.Budget = &b
		}
	}
	if len(args) > 6 && args[6] != "" {
		req.Status = &args[6]
	}
	if len(args) > 7 && args[7] != "" {
		req.Notes = &args[7]
	}
	if err := c.UpdateTripUC.Execute(c.ctx, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Trip updated successfully")
}

func (c *CLI) cmdDeleteTrip(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: delete_trip <trip_id>")
		return
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid trip_id: %v\n", err)
		return
	}
	if err := c.DeleteTripUC.Execute(c.ctx, requests.DeleteTripRequest{TripID: domain.TripID(id)}); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Trip deleted successfully")
}

// ---- REVIEWS ----

func (c *CLI) cmdCreateReview(args []string) {
	if len(args) < 5 {
		fmt.Println(`Usage: create_review <user_id> <place_id> <rating> <comment> <visit_date> [image_url]`)
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	placeID, err := uuid.Parse(args[1])
	if err != nil {
		fmt.Printf("Invalid place_id: %v\n", err)
		return
	}
	rating, err := strconv.Atoi(args[2])
	if err != nil || rating < 1 || rating > 5 {
		fmt.Println("Rating must be 1..5")
		return
	}
	visitDate, err := time.Parse("2006-01-02", args[4])
	if err != nil {
		fmt.Printf("Invalid visit_date: %v\n", err)
		return
	}
	req := requests.CreateReviewRequest{
		UserID:    domain.UserID(userID),
		PlaceID:   domain.PlaceID(placeID),
		Rating:    rating,
		Comment:   args[3],
		VisitDate: visitDate,
	}
	if len(args) >= 6 {
		req.ImageURL = args[5]
	}
	resp, err := c.CreateReviewUC.Execute(c.ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Review created: ID=%s Rating=%d Moderated=%t Approved=%t\n",
		resp.ID, resp.Rating, resp.IsModerated, resp.IsApproved)
}

func (c *CLI) cmdReviewsByPlace(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: reviews_by_place <place_id> [onlyApproved:true|false] [limit] [offset]")
		return
	}
	placeID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid place_id: %v\n", err)
		return
	}
	onlyApproved := false
	limit, offset := 50, 0
	if len(args) >= 2 {
		if b, err := strconv.ParseBool(args[1]); err == nil {
			onlyApproved = b
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			limit = n
		}
	}
	if len(args) >= 4 {
		if n, err := strconv.Atoi(args[3]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetReviewsByPlaceUC.Execute(c.ctx, requests.GetReviewsByPlaceRequest{
		PlaceID: domain.PlaceID(placeID), OnlyApproved: onlyApproved, Limit: limit, Offset: offset,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, r := range resp {
		fmt.Printf("  ID: %s | Rating: %d | User: %s | Date: %s | Approved: %t | Comment: %s\n",
			r.ID, r.Rating, r.Username, r.VisitDate, r.IsApproved, r.Comment)
	}
}

func (c *CLI) cmdUserReviews(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: user_reviews <user_id> [limit] [offset]")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	limit, offset := 50, 0
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			limit = n
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetUserReviewsUC.Execute(c.ctx, requests.GetUserReviewsRequest{
		UserID: domain.UserID(userID), Limit: limit, Offset: offset,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, r := range resp {
		fmt.Printf("  ID: %s | Place: %s (%s) | Rating: %d | Visit: %s | Approved: %t | Moderated: %t\n",
			r.ID, r.PlaceName, r.PlaceID, r.Rating, r.VisitDate, r.IsApproved, r.IsModerated)
	}
}

func (c *CLI) cmdModerateReview(args []string) {
	if len(args) < 4 {
		fmt.Println(`Usage: moderate_review <review_id> <true|false> <comment> <moderator_id>`)
		return
	}
	reviewID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid review_id: %v\n", err)
		return
	}
	approved, err := strconv.ParseBool(args[1])
	if err != nil {
		fmt.Println("Second arg must be true|false")
		return
	}
	modID, err := uuid.Parse(args[3])
	if err != nil {
		fmt.Printf("Invalid moderator_id: %v\n", err)
		return
	}
	resp, err := c.ModerateReviewUC.Execute(c.ctx, requests.ModerateReviewRequest{
		ReviewID:    domain.ReviewID(reviewID),
		Approved:    approved,
		Comment:     args[2],
		ModeratorID: domain.UserID(modID),
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Review moderated: ID=%s Approved=%t Moderated=%t Comment=%s\n",
		resp.ID, resp.IsApproved, resp.IsModerated, resp.ModerationComment)
}

func (c *CLI) cmdDeleteReview(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: delete_review <review_id> <user_id> <isModerator:true|false>")
		return
	}
	reviewID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid review_id: %v\n", err)
		return
	}
	userID, err := uuid.Parse(args[1])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	isModerator, err := strconv.ParseBool(args[2])
	if err != nil {
		fmt.Println("Third arg must be true|false")
		return
	}
	resp, err := c.DeleteReviewUC.Execute(c.ctx, requests.DeleteReviewRequest{
		ReviewID:    domain.ReviewID(reviewID),
		UserID:      domain.UserID(userID),
		IsModerator: isModerator,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Review deleted: %v\n", resp.Success)
}

func (c *CLI) cmdPendingReviews(args []string) {
	limit, offset := 50, 0
	if len(args) >= 1 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			limit = n
		}
	}
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetPendingReviewsUC.Execute(c.ctx, requests.GetPendingReviewsRequest{
		Limit:       limit,
		Offset:      offset,
		IsModerator: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, r := range resp.Reviews {
		fmt.Printf("  ID: %s | Rating: %d | User: %s | Visit: %s | Approved: %t | Comment: %s\n",
			r.GetId(), r.GetRating(), r.GetUsername(), r.GetVisitDate().Format("2006-01-02"),
			r.IsApproved(), r.GetComment())
	}
}

// ---- CONTENT ----

func (c *CLI) cmdGetPage(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: get_page <id>")
		return
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		return
	}
	resp, err := c.GetStaticPageUC.Execute(c.ctx, content.GetStaticPageByIDRequest{ID: domain.StaticPageID(id)})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Static page:\n  ID: %s\n  Slug: %s\n  Title: %s\n  Content: %s\n  Meta: %s\n  Image URL: %s\n  Image Alt: %s\n  Published: %t\n",
		resp.ID, resp.Slug, resp.Title, resp.Content, resp.MetaDesc, resp.ImageURL, resp.ImageAlt, resp.IsPublished)
}

// ---- FAVOURITES ----

func (c *CLI) cmdAddFavourite(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: add_favourite <user_id> <place_id>")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	placeID, err := uuid.Parse(args[1])
	if err != nil {
		fmt.Printf("Invalid place_id: %v\n", err)
		return
	}
	if err := c.AddFavUC.Execute(c.ctx, favourites.AddFavouriteRequest{
		UserID:  domain.UserID(userID),
		PlaceID: domain.PlaceID(placeID),
	}); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Place added to favourites")
}

func (c *CLI) cmdRemoveFavourite(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: remove_favourite <user_id> <place_id>")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	placeID, err := uuid.Parse(args[1])
	if err != nil {
		fmt.Printf("Invalid place_id: %v\n", err)
		return
	}
	if err := c.RemoveFavUC.Execute(c.ctx, favourites.RemoveFavouriteRequest{
		UserID:  domain.UserID(userID),
		PlaceID: domain.PlaceID(placeID),
	}); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Place removed from favourites")
}

func (c *CLI) cmdFavourites(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: favourites <user_id> [limit] [offset]")
		return
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		fmt.Printf("Invalid user_id: %v\n", err)
		return
	}
	limit, offset := 100, 0
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			limit = n
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			offset = n
		}
	}
	resp, err := c.GetFavUC.Execute(c.ctx, favourites.GetFavouritesRequest{
		UserID: domain.UserID(userID), Limit: limit, Offset: offset,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, p := range resp.Places {
		fmt.Printf("  ID: %s | Name: %s | Category: %s | Rating: %.2f | City: %s\n",
			p.GetId(), p.GetName(), p.GetCategory(), p.GetAvgRating(), p.GetCityName())
	}
}

// ---- ADMIN ----

func (c *CLI) cmdUsers(args []string) {
	limit, offset := 50, 0
	search := ""
	if len(args) >= 1 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			limit = n
		} else {
			search = args[0]
		}
	}
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			offset = n
		}
	}
	if len(args) >= 3 {
		search = args[2]
	}
	resp, err := c.GetUsersUC.Execute(c.ctx, requests.GetUsersRequest{
		Limit: limit, Offset: offset, Search: search, IsModerator: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, u := range resp.Users {
		fmt.Printf("  ID: %s | Username: %s | Email: %s | Role: %s | Blocked: %t | Registered: %s\n",
			u.GetId(), u.GetUsername(), u.GetEmail(), u.GetRole(), u.IsBlocked(), u.GetRegisteredAt().Format(time.RFC3339))
	}
}

// ---- ANALYTICS ----

func (c *CLI) cmdReport(args []string) {
	if len(args) < 2 {
		fmt.Println(`Usage: report <report_type> <format:json|csv> [date_from] [date_to] [limit]`)
		fmt.Println("  report_type: places | users | trips")
		return
	}
	req := requests.GenerateReportRequest{
		ReportType: args[0],
		Format:     domain.ReportFormat(args[1]),
		Options: domain.ReportOptions{
			IncludeHeaders: true,
			PrettyPrint:    args[1] == "json",
		},
		Limit: 1000,
	}
	if len(args) >= 3 {
		req.DateFrom = args[2]
	}
	if len(args) >= 4 {
		req.DateTo = args[3]
	}
	if len(args) >= 5 {
		if n, err := strconv.Atoi(args[4]); err == nil {
			req.Limit = n
		}
	}
	data, err := c.GenerateReportUC.Execute(c.ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// ---- HELPERS ----

func parseCommandLine(line string) (string, []string) {
	var args []string
	var current strings.Builder
	inQuote := false
	for _, r := range line {
		switch r {
		case '"':
			inQuote = !inQuote
		case ' ':
			if inQuote {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	if len(args) == 0 {
		return "", nil
	}
	return args[0], args[1:]
}

func printHelp() {
	fmt.Print(`GeoGuide CLI — available commands:

  === AUTH ===
  register <username> <email> <password> <phone>
  login <username> <password>
  verify <user_id> <contact> <code>
  profile <user_id>
  update_profile <token> field=value [field=value...]
      fields: username, phone, country, avatar_url, birth_date, categories
  upload_avatar <user_id> <file_path>
  logout <token>
  moderate_user <user_id> <true|false> <reason> <moderator_id>
  enable_2fa generate <user_id> <email>
  enable_2fa verify <user_id> <secret> <code>
  verify_2fa <challenge_id> <code>

  === CATALOG ===
  countries [limit] [offset]
  country <id>
  cities <country_id> [limit] [offset]
  city <id>
  search_cities <query> [limit] [offset]
  districts <city_id> [limit] [offset]
  transport_nodes <city_id> [limit] [offset]
  place <id>
  places_by_city <city_id> [category] [limit] [offset] [sortBy]
  places_by_category <category> [limit] [offset]
  holidays <YYYY-MM-DD>

  === PLANNER ===
  create_trip <title> <start_date> <end_date> <budget> <user_id> [notes]
  get_trip <trip_id> <user_id>
  user_trips <user_id>
  add_place_to_trip <trip_id> <place_id> <day> <duration_min> <user_id> [arrival_RFC3339] [notes]
  remove_place_from_trip <trip_id> <place_id> <user_id>
  update_trip <trip_id> <user_id> [title] [start_date] [end_date] [budget] [status] [notes]
  delete_trip <trip_id>

  === REVIEWS ===
  create_review <user_id> <place_id> <rating> <comment> <visit_date> [image_url]
  reviews_by_place <place_id> [true|false] [limit] [offset]
  user_reviews <user_id> [limit] [offset]
  moderate_review <review_id> <true|false> <comment> <moderator_id>
  delete_review <review_id> <user_id> <true|false>
  pending_reviews [limit] [offset]

  === CONTENT ===
  get_page <id>

  === FAVOURITES ===
  add_favourite <user_id> <place_id>
  remove_favourite <user_id> <place_id>
  favourites <user_id> [limit] [offset]

  === ADMIN ===
  users [limit] [offset] [search]

  === ANALYTICS ===
  report <report_type> <json|csv> [date_from] [date_to] [limit]

  === UTILITIES ===
  help
  exit | quit
`)
}
