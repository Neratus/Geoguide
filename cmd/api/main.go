package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Neratus/geoguide/internal/delivery"
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
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
	flag.BoolVar(&interactive, "i", true, "interactive mode")
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

	slog.Info("Starting GeoGuide CLI", "version", cfg.App.Version, "config", configPath)

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
	router := delivery.NewRouter(server, jwtSecret)

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
	fmt.Printf("Created test user: testuser / TestPass123! (ID: %s)\n", testUserID)

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
