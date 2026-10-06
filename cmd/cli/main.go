package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/logger"
	"github.com/pressly/goose/v3"

	city_repo "github.com/Neratus/geoguide/internal/repository/city"
	config "github.com/Neratus/geoguide/internal/repository/config"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
	place_repo "github.com/Neratus/geoguide/internal/repository/place"
	session_repo "github.com/Neratus/geoguide/internal/repository/session"
	static_repo "github.com/Neratus/geoguide/internal/repository/static"
	trip_repo "github.com/Neratus/geoguide/internal/repository/trip"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
	"github.com/Neratus/geoguide/internal/services"
	analytics_usecase "github.com/Neratus/geoguide/internal/usecase/analytics"
	auth_usecase "github.com/Neratus/geoguide/internal/usecase/auth"
	catalog_usecase "github.com/Neratus/geoguide/internal/usecase/catalog"
	content_usecase "github.com/Neratus/geoguide/internal/usecase/content"
	planner_usecase "github.com/Neratus/geoguide/internal/usecase/planner"
	review_usecase "github.com/Neratus/geoguide/internal/usecase/review"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func cleanTablesUser(ctx context.Context, pool *pgxpool.Pool) {
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

	userRepo, err := user_repo.NewPostgresUserRepo(cfg)
	if err != nil {
		slog.Error("failed to init user repo", "error", err)
		os.Exit(1)
	}
	defer userRepo.Close()

	slog.Info("Redis config",
		"host", cfg.Database.Redis.Host,
		"port", cfg.Database.Redis.Port,
		"db", cfg.Database.Redis.DB)

	sessionRepo, err := session_repo.NewSessionRepo(cfg)
	if err != nil {
		slog.Error("failed to init session repo", "error", err)
		os.Exit(1)
	}

	placeRepo, err := place_repo.NewPostgresPlaceRepo(cfg)
	if err != nil {
		slog.Error("failed to init place repo", "error", err)
		os.Exit(1)
	}
	defer placeRepo.Close()

	cityRepo, err := city_repo.NewPostgresCityRepo(cfg)
	if err != nil {
		slog.Error("failed to init city repo", "error", err)
		os.Exit(1)
	}
	defer cityRepo.Close()

	countryRepo, err := country_repo.NewPostgresCountryRepo(cfg)
	if err != nil {
		slog.Error("failed to init country repo", "error", err)
		os.Exit(1)
	}
	defer countryRepo.Close()

	tripRepo, err := trip_repo.NewPostgresTripRepo(cfg)
	if err != nil {
		slog.Error("failed to init trip repo", "error", err)
		os.Exit(1)
	}
	defer tripRepo.Close()

	dbURL := cfg.PostgresConnString()
	sqlDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		slog.Error("failed to open db for migrations", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	minioCfg := cfg.Database.Minio
	minioClient, err := minio.New(minioCfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioCfg.AccessKey, minioCfg.SecretKey, ""),
		Secure: minioCfg.UseSSL,
	})
	if err != nil {
		slog.Error("failed to create minio client", "error", err)
		os.Exit(1)
	}

	bucketName := minioCfg.Bucket

	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		slog.Error("failed to check bucket existence", "error", err)
		os.Exit(1)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			slog.Error("failed to create bucket", "error", err)
			os.Exit(1)
		}
	}
	staticRepo, err := static_repo.NewPostgresStaticRepo(cfg, minioClient)
	if err != nil {
		slog.Error("failed to init static repo", "error", err)
		os.Exit(1)
	}
	defer staticRepo.Close()

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

	rateLimiter, err := services.NewRateLimiter(cfg)
	if err != nil {
		slog.Error("failed to init rate limiter", "error", err)
		os.Exit(1)
	}
	reportGenService := services.NewReportGeneratorService(slog.Default())
	generateReportUC := analytics_usecase.NewGenerateReportUseCase(placeRepo, userRepo, tripRepo, reportGenService, slog.Default())

	registerUC := auth_usecase.NewRegisterUserUseCase(userRepo, sessionRepo, emailNotifier, taskQueue, slog.Default(), rateLimiter)
	loginUC := auth_usecase.NewLoginUseCase(userRepo, sessionRepo, emailNotifier, taskQueue, slog.Default(), rateLimiter)
	updateProfileUC := auth_usecase.NewUpdateUserProfileUseCase(userRepo, sessionRepo, slog.Default())
	getProfileUC := auth_usecase.NewGetUserProfileUseCase(userRepo, sessionRepo, slog.Default())
	logoutUC := auth_usecase.NewLogoutUseCase(sessionRepo, slog.Default())
	moderateUserUC := auth_usecase.NewModerateUserUsecase(userRepo, emailNotifier, taskQueue, slog.Default())

	getCountriesUC := catalog_usecase.NewGetCountriesUseCase(countryRepo, cityRepo, slog.Default())
	getCountryUC := catalog_usecase.NewGetCountryUseCase(countryRepo, cityRepo, slog.Default())
	getCitiesUC := catalog_usecase.NewGetCitiesByCountryUseCase(cityRepo, slog.Default())
	getPlaceUC := catalog_usecase.NewGetPlaceUseCase(placeRepo, userRepo, slog.Default())
	getPlacesByCityUC := catalog_usecase.NewGetPlacesByCityUseCase(placeRepo, slog.Default())
	getPlacesByCategoryUC := catalog_usecase.NewGetPlacesByCategoryUseCase(placeRepo, slog.Default())

	createTripUC := planner_usecase.NewCreateTripUseCase(tripRepo, slog.Default())
	getTripUC := planner_usecase.NewGetTripUseCase(tripRepo, slog.Default())
	getUserTripsUC := planner_usecase.NewGetUserTripsUseCase(tripRepo, slog.Default())
	addPlaceToTripUC := planner_usecase.NewAddPlaceToTripUseCase(tripRepo, placeRepo, slog.Default())
	removePlaceFromTripUC := planner_usecase.NewRemovePlaceFromTripUseCase(tripRepo, slog.Default())
	updateTripUC := planner_usecase.NewUpdateTripUseCase(tripRepo, slog.Default())
	deleteTripUC := planner_usecase.NewDeleteTripUseCase(tripRepo, slog.Default())

	createReviewUC := review_usecase.NewCreateReviewUseCase(placeRepo, emailNotifier, taskQueue, slog.Default())
	getReviewsByPlaceUC := review_usecase.NewGetReviewsByPlaceUseCase(placeRepo, slog.Default())
	getUserReviewsUC := review_usecase.NewGetUserReviewsUseCase(placeRepo, slog.Default())
	moderateReviewUC := review_usecase.NewModerateReviewUseCase(placeRepo, userRepo, emailNotifier, taskQueue, slog.Default())
	deleteReviewUC := review_usecase.NewDeleteReviewUseCase(placeRepo, slog.Default())

	getStaticPageUC := content_usecase.NewGetStaticPageByIDUseCase(staticRepo, slog.Default())
	createStaticPageUC := content_usecase.NewCreateStaticPageUseCase(staticRepo, slog.Default())
	updateStaticPageUC := content_usecase.NewUpdateStaticPageUseCase(staticRepo, slog.Default())
	deleteStaticPageUC := content_usecase.NewDeleteStaticPageUseCase(staticRepo, slog.Default())

	verifyContactUC := auth_usecase.NewVerifyContactUseCase(userRepo, sessionRepo, emailNotifier, taskQueue, slog.Default())
	enable2FAUC := auth_usecase.NewEnableTwoFactorUseCase(userRepo, slog.Default())
	verify2FALoginUC := auth_usecase.NewVerifyTwoFactorLoginUseCase(userRepo, sessionRepo, slog.Default())
	disable2FAUC := auth_usecase.NewDisableTwoFactorUseCase(userRepo, slog.Default())

	cleanTablesUser(ctx, cityRepo.Pool)
	seedDatabase(ctx, countryRepo, cityRepo, placeRepo, slog.Default())
	fmt.Println("Test data seeded successfully!")

	if interactive {
		runInteractive(ctx, slog.Default(), []string{},
			registerUC, loginUC, getProfileUC, updateProfileUC, logoutUC, moderateUserUC,
			verifyContactUC, enable2FAUC, verify2FALoginUC, disable2FAUC,
			getCountriesUC, getCountryUC, getCitiesUC, getPlaceUC, getPlacesByCityUC, getPlacesByCategoryUC,
			createTripUC, getTripUC, getUserTripsUC, addPlaceToTripUC, removePlaceFromTripUC, updateTripUC, deleteTripUC,
			createReviewUC, getReviewsByPlaceUC, getUserReviewsUC, moderateReviewUC, deleteReviewUC,
			getStaticPageUC, createStaticPageUC, updateStaticPageUC, deleteStaticPageUC)
	} else {
		runCommandLine(ctx, slog.Default(), flag.Args(),
			registerUC, loginUC, getProfileUC, updateProfileUC, logoutUC, moderateUserUC,
			verifyContactUC, enable2FAUC, verify2FALoginUC, disable2FAUC,
			getCountriesUC, getCountryUC, getCitiesUC, getPlaceUC, getPlacesByCityUC, getPlacesByCategoryUC,
			createTripUC, getTripUC, getUserTripsUC, addPlaceToTripUC, removePlaceFromTripUC, updateTripUC, deleteTripUC,
			createReviewUC, getReviewsByPlaceUC, getUserReviewsUC, moderateReviewUC, deleteReviewUC,
			getStaticPageUC, createStaticPageUC, updateStaticPageUC, deleteStaticPageUC)
	}
}

func runCommandLine(ctx context.Context, logger *slog.Logger, args []string,
	registerUC *auth_usecase.RegisterUserUseCase,
	loginUC *auth_usecase.LoginUseCase,
	getProfileUC *auth_usecase.GetUserProfileUseCase,
	updateProfileUC *auth_usecase.UpdateUserProfileUseCase,
	logoutUC *auth_usecase.LogoutUseCase,
	moderateUserUC *auth_usecase.ModerateUserUsecase,
	verifyContactUC *auth_usecase.VerifyContactUseCase,
	enable2FAUC *auth_usecase.EnableTwoFactorUseCase,
	verify2FALoginUC *auth_usecase.VerifyTwoFactorLoginUseCase,
	disable2FAUC *auth_usecase.DisableTwoFactorUseCase,
	getCountriesUC *catalog_usecase.GetCountriesUseCase,
	getCountryUC *catalog_usecase.GetCountryUseCase,
	getCitiesUC *catalog_usecase.GetCitiesByCountryUseCase,
	getPlaceUC *catalog_usecase.GetPlaceUseCase,
	getPlacesByCityUC *catalog_usecase.GetPlacesByCityUseCase,
	getPlacesByCategoryUC *catalog_usecase.GetPlacesByCategoryUseCase,
	createTripUC *planner_usecase.CreateTripUseCase,
	getTripUC *planner_usecase.GetTripUseCase,
	getUserTripsUC *planner_usecase.GetUserTripsUseCase,
	addPlaceToTripUC *planner_usecase.AddPlaceToTripUseCase,
	removePlaceFromTripUC *planner_usecase.RemovePlaceFromTripUseCase,
	updateTripUC *planner_usecase.UpdateTripUseCase,
	deleteTripUC *planner_usecase.DeleteTripUseCase,
	createReviewUC *review_usecase.CreateReviewUseCase,
	getReviewsByPlaceUC *review_usecase.GetReviewsByPlaceUseCase,
	getUserReviewsUC *review_usecase.GetUserReviewsUseCase,
	moderateReviewUC *review_usecase.ModerateReviewUseCase,
	deleteReviewUC *review_usecase.DeleteReviewUseCase,
	getStaticPageUC *content_usecase.GetStaticPageByIDUseCase,
	createStaticPageUC *content_usecase.CreateStaticPageUseCase,
	updateStaticPageUC *content_usecase.UpdateStaticPageUseCase,
	deleteStaticPageUC *content_usecase.DeleteStaticPageUseCase,
) {
	if len(args) == 0 {
		fmt.Println("No command provided. Use -i for interactive mode or provide a command.")
		os.Exit(1)
	}

	cmd := args[0]
	params := args[1:]

	switch cmd {
	case "register":
		if len(params) < 4 {
			fmt.Println("Usage: register <username> <email> <password> <phone>")
			os.Exit(1)
		}
		req := requests.RegisterUserRequest{
			Username: params[0],
			Email:    params[1],
			Password: params[2],
			Phone:    params[3],
		}
		resp, err := registerUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Registered user: %s (ID: %s). Verification codes sent.\n", resp.Username, resp.ID)

	case "login":
		if len(params) < 2 {
			fmt.Println("Usage: login <username> <password>")
			os.Exit(1)
		}
		req := requests.LoginUserRequest{Username: params[0], Password: params[1]}
		resp, err := loginUC.Execute(ctx, req, "", "")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if resp.RequiresTwoFactor {
			fmt.Printf("Two-factor authentication required. Challenge ID: %s\n", resp.ChallengeID)
		} else {
			fmt.Printf("Logged in as %s (token: %s)\n", resp.Username, resp.Token)
		}

	case "verify":
		if len(params) < 3 {
			fmt.Println("Usage: verify <user_id> <contact> <code>")
			os.Exit(1)
		}
		userID, err := uuid.Parse(params[0])
		if err != nil {
			fmt.Printf("Invalid user_id: %v\n", err)
			os.Exit(1)
		}
		contact := params[1]
		code := params[2]
		req := requests.VerifyContactRequest{
			UserID:  domain.UserID(userID),
			Contact: contact,
			Code:    code,
		}
		err = verifyContactUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Contact verified successfully")

	case "enable_2fa":
		if len(params) < 2 {
			fmt.Println("Usage:")
			fmt.Println("  enable_2fa generate <user_id> <email>")
			fmt.Println("  enable_2fa verify <user_id> <secret> <code>")
			os.Exit(1)
		}
		subcmd := params[0]
		if subcmd == "generate" && len(params) >= 3 {
			userID, err := uuid.Parse(params[1])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				os.Exit(1)
			}
			email := params[2]
			secret, uri, err := enable2FAUC.GenerateSecret(ctx, domain.UserID(userID), email)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("TOTP Secret: %s\n", secret)
			fmt.Printf("OTPAuth URI: %s\n", uri)
			fmt.Println("Scan this QR code with Google Authenticator or similar app")
		} else if subcmd == "verify" && len(params) >= 4 {
			userID, err := uuid.Parse(params[1])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				os.Exit(1)
			}
			secret := params[2]
			code := params[3]
			backupCodes, err := enable2FAUC.VerifyAndEnable(ctx, requests.EnableTwoFactorRequest{
				UserID: domain.UserID(userID),
				Secret: secret,
				Code:   code,
			})
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("2FA successfully enabled. Save these backup codes (one-time use):")
			for i, bc := range backupCodes {
				fmt.Printf("  %d. %s\n", i+1, bc)
			}
		} else {
			fmt.Println("Invalid enable_2fa subcommand")
		}

	case "disable_2fa":
		if len(params) < 1 {
			fmt.Println("Usage: disable_2fa <user_id>")
			os.Exit(1)
		}
		userID, err := uuid.Parse(params[0])
		if err != nil {
			fmt.Printf("Invalid user_id: %v\n", err)
			os.Exit(1)
		}
		err = disable2FAUC.Execute(ctx, domain.UserID(userID))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("2FA disabled successfully")

	case "verify_2fa":
		if len(params) < 2 {
			fmt.Println("Usage: verify_2fa <challenge_id> <code>")
			os.Exit(1)
		}
		challengeID := params[0]
		code := params[1]
		token, err := verify2FALoginUC.Execute(ctx, requests.VerifyTwoFactorRequest{
			ChallengeID: challengeID,
			Code:        code,
		})
		if err != nil {
			fmt.Printf("2FA verification failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("2FA successful, session token: %s\n", token)

	case "profile":
		if len(params) < 1 {
			fmt.Println("Usage: profile <token>")
			os.Exit(1)
		}
		req := requests.GetUserProfileRequest{Token: params[0]}
		resp, err := getProfileUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User: %s, email: %s, phone: %s\n", resp.Username, resp.Email, resp.Phone)

	case "logout":
		if len(params) < 1 {
			fmt.Println("Usage: logout <token>")
			os.Exit(1)
		}
		req := requests.LogoutRequest{Token: params[0]}
		resp, err := logoutUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Logged out: %v\n", resp.Success)

	case "countries":
		req := requests.GetCountriesRequest{Limit: 20, Offset: 0}
		resp, err := getCountriesUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for _, c := range resp {
			fmt.Printf("%s (%s)\n", c.Name, c.Capital)
		}

	case "country":
		if len(params) < 1 {
			fmt.Println("Usage: country <id>")
			os.Exit(1)
		}
		id, err := uuid.Parse(params[0])
		if err != nil {
			fmt.Printf("Invalid ID: %v\n", err)
			os.Exit(1)
		}
		req := requests.GetCountryRequest{ID: domain.CountryID(id)}
		resp, err := getCountryUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s, capital: %s, area: %f\n", resp.Name, resp.CapitalName, resp.Area)

	case "cities":
		if len(params) < 1 {
			fmt.Println("Usage: cities <country_id>")
			os.Exit(1)
		}
		countryID, err := uuid.Parse(params[0])
		if err != nil {
			fmt.Printf("Invalid ID: %v\n", err)
			os.Exit(1)
		}
		req := requests.GetCitiesByCountryRequest{CountryID: domain.CountryID(countryID), Limit: 20, Offset: 0}
		resp, err := getCitiesUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for _, c := range resp {
			fmt.Printf("%s (population: %d)\n", c.Name, c.Population)
		}

	case "place":
		if len(params) < 1 {
			fmt.Println("Usage: place <id>")
			os.Exit(1)
		}
		id, err := uuid.Parse(params[0])
		if err != nil {
			fmt.Printf("Invalid ID: %v\n", err)
			os.Exit(1)
		}
		req := requests.GetPlaceRequest{ID: domain.PlaceID(id)}
		resp, err := getPlaceUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Place: %s, category: %s, rating: %.2f\n", resp.Name, resp.Category, resp.AvgRating)

	case "places_by_city":
		if len(params) < 1 {
			fmt.Println("Usage: places_by_city <city_id>")
			os.Exit(1)
		}
		cityID, err := uuid.Parse(params[0])
		if err != nil {
			fmt.Printf("Invalid ID: %v\n", err)
			os.Exit(1)
		}
		req := requests.GetPlacesByCityRequest{CityID: domain.CityID(cityID), Limit: 20, Offset: 0}
		resp, err := getPlacesByCityUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for _, p := range resp {
			fmt.Printf("%s (%s)\n", p.Name, p.Category)
		}

	case "places_by_category":
		if len(params) < 1 {
			fmt.Println("Usage: places_by_category <category>")
			os.Exit(1)
		}
		req := requests.GetPlacesByCategoryRequest{Category: params[0], Limit: 20, Offset: 0}
		resp, err := getPlacesByCategoryUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for _, p := range resp {
			fmt.Printf("%s\n", p.Name)
		}

	case "create_trip":
		if len(params) < 5 {
			fmt.Println("Usage: create_trip <title> <start_date> <end_date> <budget> <user_id>")
			os.Exit(1)
		}
		startDate, _ := time.Parse("2006-01-02", params[1])
		endDate, _ := time.Parse("2006-01-02", params[2])
		budget, _ := strconv.ParseFloat(params[3], 64)
		userID, _ := uuid.Parse(params[4])
		req := requests.CreateTripRequest{
			Title:     params[0],
			StartDate: startDate,
			EndDate:   endDate,
			Budget:    budget,
			UserID:    domain.UserID(userID),
		}
		resp, err := createTripUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Trip created: %s (ID: %s)\n", resp.Title, resp.ID)

	case "get_trip":
		if len(params) < 1 {
			fmt.Println("Usage: get_trip <trip_id>")
			os.Exit(1)
		}
		id, _ := uuid.Parse(params[0])
		req := requests.GetTripRequest{TripID: domain.TripID(id)}
		resp, err := getTripUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Trip: %s, %s - %s, %d places\n", resp.Title, resp.StartDate, resp.EndDate, len(resp.Places))

	case "user_trips":
		if len(params) < 1 {
			fmt.Println("Usage: user_trips <user_id>")
			os.Exit(1)
		}
		userID, _ := uuid.Parse(params[0])
		req := requests.GetUserTripsRequest{UserID: domain.UserID(userID)}
		resp, err := getUserTripsUC.Execute(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for _, t := range resp {
			fmt.Printf("%s (%s - %s)\n", t.Title, t.StartDate, t.EndDate)
		}

	default:
		fmt.Printf("Unknown command: %s. Use -i for interactive mode or 'help'.\n", cmd)
		os.Exit(1)
	}
}

func runInteractive(ctx context.Context, logger *slog.Logger, args []string,
	registerUC *auth_usecase.RegisterUserUseCase,
	loginUC *auth_usecase.LoginUseCase,
	getProfileUC *auth_usecase.GetUserProfileUseCase,
	updateProfileUC *auth_usecase.UpdateUserProfileUseCase,
	logoutUC *auth_usecase.LogoutUseCase,
	moderateUserUC *auth_usecase.ModerateUserUsecase,
	verifyContactUC *auth_usecase.VerifyContactUseCase,
	enable2FAUC *auth_usecase.EnableTwoFactorUseCase,
	verify2FALoginUC *auth_usecase.VerifyTwoFactorLoginUseCase,
	disable2FAUC *auth_usecase.DisableTwoFactorUseCase,
	getCountriesUC *catalog_usecase.GetCountriesUseCase,
	getCountryUC *catalog_usecase.GetCountryUseCase,
	getCitiesUC *catalog_usecase.GetCitiesByCountryUseCase,
	getPlaceUC *catalog_usecase.GetPlaceUseCase,
	getPlacesByCityUC *catalog_usecase.GetPlacesByCityUseCase,
	getPlacesByCategoryUC *catalog_usecase.GetPlacesByCategoryUseCase,
	createTripUC *planner_usecase.CreateTripUseCase,
	getTripUC *planner_usecase.GetTripUseCase,
	getUserTripsUC *planner_usecase.GetUserTripsUseCase,
	addPlaceToTripUC *planner_usecase.AddPlaceToTripUseCase,
	removePlaceFromTripUC *planner_usecase.RemovePlaceFromTripUseCase,
	updateTripUC *planner_usecase.UpdateTripUseCase,
	deleteTripUC *planner_usecase.DeleteTripUseCase,
	createReviewUC *review_usecase.CreateReviewUseCase,
	getReviewsByPlaceUC *review_usecase.GetReviewsByPlaceUseCase,
	getUserReviewsUC *review_usecase.GetUserReviewsUseCase,
	moderateReviewUC *review_usecase.ModerateReviewUseCase,
	deleteReviewUC *review_usecase.DeleteReviewUseCase,
	getStaticPageUC *content_usecase.GetStaticPageByIDUseCase,
	createStaticPageUC *content_usecase.CreateStaticPageUseCase,
	updateStaticPageUC *content_usecase.UpdateStaticPageUseCase,
	deleteStaticPageUC *content_usecase.DeleteStaticPageUseCase,
) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("GeoGuide CLI (interactive mode)")
	fmt.Println("Available commands: help, register, login, profile, logout, verify, enable_2fa, disable_2fa, verify_2fa, countries, country <id>, cities <country_id>, place <id>, places_by_city <city_id>, places_by_category <category>, create_trip, get_trip <id>, user_trips <user_id>, add_place_to_trip, remove_place_from_trip, update_trip, delete_trip, create_review, reviews_by_place <place_id>, user_reviews <user_id>, moderate_review, delete_review, get_page <slug>, create_page, update_page, delete_page, set_config, exit")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		cmd, cmdArgs := parseCommandLine(line)

		switch cmd {
		case "help":
			printHelp()
		case "exit", "quit":
			return

		case "register":
			if len(cmdArgs) < 4 {
				fmt.Println("Usage: register <username> <email> <password> <phone>")
				continue
			}
			req := requests.RegisterUserRequest{
				Username: cmdArgs[0],
				Email:    cmdArgs[1],
				Password: cmdArgs[2],
				Phone:    cmdArgs[3],
			}
			resp, err := registerUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Registered user: %s (ID: %s). Verification codes sent.\n", resp.Username, resp.ID)
			}

		case "login":
			if len(cmdArgs) < 2 {
				fmt.Println("Usage: login <username> <password>")
				continue
			}
			req := requests.LoginUserRequest{Username: cmdArgs[0], Password: cmdArgs[1]}
			resp, err := loginUC.Execute(ctx, req, "", "")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if resp.RequiresTwoFactor {
				fmt.Printf("Two-factor authentication required. Challenge ID: %s\n", resp.ChallengeID)
			} else {
				fmt.Printf("Logged in as %s (token: %s)\n", resp.Username, resp.Token)
			}

		case "verify":
			if len(cmdArgs) < 3 {
				fmt.Println("Usage: verify <user_id> <contact> <code>")
				continue
			}
			userID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				continue
			}
			contact := cmdArgs[1]
			code := cmdArgs[2]
			req := requests.VerifyContactRequest{
				UserID:  domain.UserID(userID),
				Contact: contact,
				Code:    code,
			}
			err = verifyContactUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Contact verified successfully")
			}

		case "enable_2fa":
			if len(cmdArgs) < 2 {
				fmt.Println("Usage:")
				fmt.Println("  enable_2fa generate <user_id> <email>")
				fmt.Println("  enable_2fa verify <user_id> <secret> <code>")
				continue
			}
			subcmd := cmdArgs[0]
			if subcmd == "generate" && len(cmdArgs) >= 3 {
				userID, err := uuid.Parse(cmdArgs[1])
				if err != nil {
					fmt.Printf("Invalid user_id: %v\n", err)
					continue
				}
				email := cmdArgs[2]
				secret, uri, err := enable2FAUC.GenerateSecret(ctx, domain.UserID(userID), email)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("TOTP Secret: %s\n", secret)
					fmt.Printf("OTPAuth URI: %s\n", uri)
					fmt.Println("Scan this QR code with Google Authenticator or similar app")
				}
			} else if subcmd == "verify" && len(cmdArgs) >= 4 {
				userID, err := uuid.Parse(cmdArgs[1])
				if err != nil {
					fmt.Printf("Invalid user_id: %v\n", err)
					continue
				}
				secret := cmdArgs[2]
				code := cmdArgs[3]
				backupCodes, err := enable2FAUC.VerifyAndEnable(ctx, requests.EnableTwoFactorRequest{
					UserID: domain.UserID(userID),
					Secret: secret,
					Code:   code,
				})
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println("2FA successfully enabled. Save these backup codes (one-time use):")
					for i, bc := range backupCodes {
						fmt.Printf("  %d. %s\n", i+1, bc)
					}
				}
			} else {
				fmt.Println("Invalid enable_2fa subcommand")
			}

		case "disable_2fa":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: disable_2fa <user_id>")
				continue
			}
			userID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				continue
			}
			err = disable2FAUC.Execute(ctx, domain.UserID(userID))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("2FA disabled successfully")
			}

		case "verify_2fa":
			if len(cmdArgs) < 2 {
				fmt.Println("Usage: verify_2fa <challenge_id> <code>")
				continue
			}
			challengeID := cmdArgs[0]
			code := cmdArgs[1]
			token, err := verify2FALoginUC.Execute(ctx, requests.VerifyTwoFactorRequest{
				ChallengeID: challengeID,
				Code:        code,
			})
			if err != nil {
				fmt.Printf("2FA verification failed: %v\n", err)
			} else {
				fmt.Printf("2FA successful, session token: %s\n", token)
			}

		case "profile":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: profile <token>")
				continue
			}
			req := requests.GetUserProfileRequest{Token: cmdArgs[0]}
			resp, err := getProfileUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("User profile:\n  ID: %s\n  Username: %s\n  Email: %s\n  Phone: %s\n  Registered: %s\n  Birth Date: %s\n  Country: %s\n  Avatar: %s\n  Categories: %v\n  Blocked: %t\n",
					resp.ID, resp.Username, resp.Email, resp.Phone, resp.RegisteredAt, resp.BirthDate, resp.CountryOfResidence, resp.AvatarURL, resp.FavoriteCategories, resp.IsBlocked)
			}

		case "logout":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: logout <token>")
				continue
			}
			req := requests.LogoutRequest{Token: cmdArgs[0]}
			resp, err := logoutUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Logged out: %v\n", resp.Success)
			}

		case "countries":
			req := requests.GetCountriesRequest{Limit: 100, Offset: 0}
			resp, err := getCountriesUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Found %d countries:\n", len(resp))
				for _, c := range resp {
					fmt.Printf("  ID: %s | Name: %s | Capital: %s | Area: %.2f | Pop: %d | Currency: %s | Language: %s\n",
						c.ID, c.Name, c.Capital, c.Area, c.Population, c.Currency, c.Language)
				}
			}

		case "country":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: country <id>")
				continue
			}
			id, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid ID: %v\n", err)
				continue
			}
			req := requests.GetCountryRequest{ID: domain.CountryID(id)}
			resp, err := getCountryUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Country details:\n  ID: %s\n  Name: %s\n  Capital: %s (ID: %s)\n  Area: %.2f km²\n  Population: %d\n  GDP: %.2f\n  Currency: %s\n  Visa: %s\n  Description: %s\n  Safety: %s\n  Best season: %s\n  Language: %s\n  Phone: %s\n  Religion: %s\n  Image: %s\n  Holidays: %d\n",
					resp.ID, resp.Name, resp.CapitalName, resp.CapitalID, resp.Area, resp.Population, resp.GDP, resp.Currency,
					resp.VisaRequirements, resp.Description, resp.SafetyTips, resp.BestSeason, resp.Language, resp.PhoneCode,
					resp.Religion, resp.ImageURL, len(resp.Holidays))
				for i, h := range resp.Holidays {
					fmt.Printf("    %d. %s (%s): %s\n", i+1, h.Name, h.Date.Format("2006-01-02"), h.Description)
				}
			}

		case "cities":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: cities <country_id>")
				continue
			}
			countryID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid ID: %v\n", err)
				continue
			}
			req := requests.GetCitiesByCountryRequest{CountryID: domain.CountryID(countryID), Limit: 100, Offset: 0}
			resp, err := getCitiesUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Cities in country:\n")
				for _, c := range resp {
					fmt.Printf("  ID: %s | Name: %s | Population: %d | Capital: %t | Timezone: %s\n",
						c.ID, c.Name, c.Population, c.IsCapital, c.Timezone)
				}
			}

		case "place":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: place <id>")
				continue
			}
			id, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid ID: %v\n", err)
				continue
			}
			req := requests.GetPlaceRequest{ID: domain.PlaceID(id)}
			resp, err := getPlaceUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Place details:\n  ID: %s\n  Name: %s\n  Category: %s\n  Description: %s\n  Coordinates: %s\n  Address: %s\n  Hours: %s\n  Price: %s\n  Duration: %d min\n  Rating: %.2f (%d reviews)\n  Phone: %s\n  Website: %s\n  City: %s (ID: %s)\n  District: %s (ID: %s)\n  Image: %s\n  Reviews: %d\n",
					resp.ID, resp.Name, resp.Category, resp.Description, resp.Coordinates, resp.Address, resp.OpeningHours,
					resp.PriceInfo, resp.AvgVisitDurationMin, resp.AvgRating, resp.ReviewsCount, resp.ContactPhone, resp.Website,
					resp.CityName, resp.CityID, resp.DistrictName, resp.DistrictID, resp.ImageURL, len(resp.Reviews))
				for i, r := range resp.Reviews {
					fmt.Printf("    %d. Rating: %d | User: %s | Date: %s | Comment: %s\n", i+1, r.Rating, r.Username, r.VisitDate, r.Comment)
				}
			}

		case "places_by_city":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: places_by_city <city_id>")
				continue
			}
			cityID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid ID: %v\n", err)
				continue
			}
			req := requests.GetPlacesByCityRequest{CityID: domain.CityID(cityID), Limit: 100, Offset: 0}
			resp, err := getPlacesByCityUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Places in city:\n")
				for _, p := range resp {
					fmt.Printf("  ID: %s | Name: %s | Category: %s | Rating: %.2f\n", p.ID, p.Name, p.Category, p.AvgRating)
				}
			}

		case "places_by_category":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: places_by_category <category>")
				continue
			}
			req := requests.GetPlacesByCategoryRequest{Category: cmdArgs[0], Limit: 100, Offset: 0}
			resp, err := getPlacesByCategoryUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Places in category '%s':\n", cmdArgs[0])
				for _, p := range resp {
					fmt.Printf("  ID: %s | Name: %s | Rating: %.2f\n", p.ID, p.Name, p.AvgRating)
				}
			}
		case "set_config":
			if len(cmdArgs) < 2 {
				fmt.Println("Usage: set_config <key> <value>")
				fmt.Println("Example: set_config MaxReviewCommentLength 100")
				continue
			}
			key := cmdArgs[0]
			value := cmdArgs[1]
			var converted interface{}
			if i, err := strconv.Atoi(value); err == nil {
				converted = i
			} else {
				converted = value
			}
			if err := domain.UpdateConfigFromMap(map[string]interface{}{key: converted}); err != nil {
				fmt.Printf("Error updating config: %v\n", err)
			} else {
				fmt.Printf("Config updated: %s = %v\n", key, converted)
			}

		case "create_trip":
			if len(cmdArgs) < 5 {
				fmt.Println("Usage: create_trip <title> <start_date> <end_date> <budget> <user_id>")
				fmt.Println("Example: create_trip \"Paris Adventure\" 2025-07-01 2025-07-07 1500.00 6f3a82d8-e8f8-48f4-ad35-67585b5d4942")
				continue
			}
			title := cmdArgs[0]
			startDate, err := time.Parse("2006-01-02", cmdArgs[1])
			if err != nil {
				fmt.Printf("Invalid start_date: %v\n", err)
				continue
			}
			endDate, err := time.Parse("2006-01-02", cmdArgs[2])
			if err != nil {
				fmt.Printf("Invalid end_date: %v\n", err)
				continue
			}
			budget, err := strconv.ParseFloat(cmdArgs[3], 64)
			if err != nil {
				fmt.Printf("Invalid budget: %v\n", err)
				continue
			}
			userID, err := uuid.Parse(cmdArgs[4])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				continue
			}
			req := requests.CreateTripRequest{
				Title:     title,
				StartDate: startDate,
				EndDate:   endDate,
				Budget:    budget,
				UserID:    domain.UserID(userID),
				Notes:     "",
			}
			resp, err := createTripUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Trip created:\n  ID: %s\n  Title: %s\n  Dates: %s – %s\n  Budget: %.2f\n  Status: %s\n",
					resp.ID, resp.Title, resp.StartDate, resp.EndDate, resp.Budget, resp.Status)
			}

		case "get_trip":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: get_trip <trip_id>")
				continue
			}
			id, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid ID: %v\n", err)
				continue
			}
			req := requests.GetTripRequest{TripID: domain.TripID(id)}
			resp, err := getTripUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Trip details:\n  ID: %s\n  Title: %s\n  Dates: %s – %s\n  Budget: %.2f\n  Status: %s\n  Notes: %s\n  Places: %d\n",
					resp.ID, resp.Title, resp.StartDate, resp.EndDate, resp.Budget, resp.Status, resp.Notes, len(resp.Places))
				for i, p := range resp.Places {
					fmt.Printf("    %d. Day %d | Place ID: %s | Arrival: %s | Duration: %d min | Status: %s | Cost: %.2f\n",
						i+1, p.DayNumber, p.PlaceID, p.ArrivalTime, p.DurationMin, p.VisitStatus, p.ActualCost)
				}
			}

		case "user_trips":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: user_trips <user_id>")
				continue
			}
			userID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				continue
			}
			req := requests.GetUserTripsRequest{UserID: domain.UserID(userID)}
			resp, err := getUserTripsUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("User trips:\n")
				for _, t := range resp {
					fmt.Printf("  ID: %s | %s | %s – %s | %s\n", t.ID, t.Title, t.StartDate, t.EndDate, t.Status)
				}
			}

		case "add_place_to_trip":
			if len(cmdArgs) < 4 {
				fmt.Println("Usage: add_place_to_trip <trip_id> <place_id> <day_number> <duration_min>")
				continue
			}
			tripID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid trip_id: %v\n", err)
				continue
			}
			placeID, err := uuid.Parse(cmdArgs[1])
			if err != nil {
				fmt.Printf("Invalid place_id: %v\n", err)
				continue
			}
			dayNum, err := strconv.Atoi(cmdArgs[2])
			if err != nil {
				fmt.Printf("Invalid day number: %v\n", err)
				continue
			}
			duration, err := strconv.Atoi(cmdArgs[3])
			if err != nil {
				fmt.Printf("Invalid duration: %v\n", err)
				continue
			}
			req := requests.AddPlaceToTripRequest{
				TripID:      domain.TripID(tripID),
				PlaceID:     domain.PlaceID(placeID),
				DayNumber:   dayNum,
				ArrivalTime: nil,
				DurationMin: duration,
				Notes:       "",
			}
			err = addPlaceToTripUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Place added to trip successfully")
			}

		case "remove_place_from_trip":
			if len(cmdArgs) < 2 {
				fmt.Println("Usage: remove_place_from_trip <trip_id> <place_id>")
				continue
			}
			tripID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid trip_id: %v\n", err)
				continue
			}
			placeID, err := uuid.Parse(cmdArgs[1])
			if err != nil {
				fmt.Printf("Invalid place_id: %v\n", err)
				continue
			}
			req := requests.RemovePlaceFromTripRequest{TripID: domain.TripID(tripID), PlaceID: domain.PlaceID(placeID)}
			err = removePlaceFromTripUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Place removed from trip successfully")
			}

		case "update_trip":
			if len(cmdArgs) < 2 {
				fmt.Println("Usage: update_trip <trip_id> [title] [start_date] [end_date] [budget] [status] [notes]")
				continue
			}
			id, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid trip_id: %v\n", err)
				continue
			}
			req := requests.UpdateTripRequest{TripID: domain.TripID(id)}
			if len(cmdArgs) > 1 && cmdArgs[1] != "" {
				req.Title = &cmdArgs[1]
			}
			if len(cmdArgs) > 2 && cmdArgs[2] != "" {
				t, err := time.Parse("2006-01-02", cmdArgs[2])
				if err == nil {
					req.StartDate = &t
				}
			}
			if len(cmdArgs) > 3 && cmdArgs[3] != "" {
				t, err := time.Parse("2006-01-02", cmdArgs[3])
				if err == nil {
					req.EndDate = &t
				}
			}
			if len(cmdArgs) > 4 && cmdArgs[4] != "" {
				budget, err := strconv.ParseFloat(cmdArgs[4], 64)
				if err == nil {
					req.Budget = &budget
				}
			}
			if len(cmdArgs) > 5 && cmdArgs[5] != "" {
				req.Status = &cmdArgs[5]
			}
			if len(cmdArgs) > 6 && cmdArgs[6] != "" {
				req.Notes = &cmdArgs[6]
			}
			err = updateTripUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Trip updated successfully")
			}

		case "delete_trip":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: delete_trip <trip_id>")
				continue
			}
			id, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid trip_id: %v\n", err)
				continue
			}
			req := requests.DeleteTripRequest{TripID: domain.TripID(id)}
			err = deleteTripUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Trip deleted successfully")
			}

		case "create_review":
			if len(cmdArgs) < 5 {
				fmt.Println("Usage: create_review <user_id> <place_id> <rating> <comment> <visit_date>")
				fmt.Println("Example: create_review 6f3a82d8-e8f8-48f4-ad35-67585b5d4942 44444444-4444-4444-4444-444444444444 5 \"Great place!\" 2025-06-15")
				continue
			}
			userID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				continue
			}
			placeID, err := uuid.Parse(cmdArgs[1])
			if err != nil {
				fmt.Printf("Invalid place_id: %v\n", err)
				continue
			}
			rating, err := strconv.Atoi(cmdArgs[2])
			if err != nil || rating < 1 || rating > 5 {
				fmt.Println("Rating must be between 1 and 5")
				continue
			}
			comment := cmdArgs[3]
			visitDate, err := time.Parse("2006-01-02", cmdArgs[4])
			if err != nil {
				fmt.Printf("Invalid visit_date: %v\n", err)
				continue
			}
			req := requests.CreateReviewRequest{
				UserID:    domain.UserID(userID),
				PlaceID:   domain.PlaceID(placeID),
				Rating:    rating,
				Comment:   comment,
				VisitDate: visitDate,
				ImageURL:  "",
			}
			resp, err := createReviewUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Review created:\n  ID: %s\n  Rating: %d\n  Comment: %s\n  Visit: %s\n  Created: %s\n  Moderated: %t\n  Approved: %t\n",
					resp.ID, resp.Rating, resp.Comment, resp.VisitDate, resp.CreatedAt, resp.IsModerated, resp.IsApproved)
			}

		case "reviews_by_place":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: reviews_by_place <place_id>")
				continue
			}
			placeID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid place_id: %v\n", err)
				continue
			}
			req := requests.GetReviewsByPlaceRequest{PlaceID: domain.PlaceID(placeID), OnlyApproved: false, Limit: 50, Offset: 0}
			resp, err := getReviewsByPlaceUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Reviews for place:\n")
				for _, r := range resp {
					fmt.Printf("  ID: %s | Rating: %d | User: %s | Date: %s | Comment: %s | Approved: %t\n",
						r.ID, r.Rating, r.Username, r.VisitDate, r.Comment, r.IsApproved)
				}
			}

		case "user_reviews":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: user_reviews <user_id>")
				continue
			}
			userID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				continue
			}
			req := requests.GetUserReviewsRequest{UserID: domain.UserID(userID), Limit: 50, Offset: 0}
			resp, err := getUserReviewsUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("User reviews:\n")
				for _, r := range resp {
					fmt.Printf("  ID: %s | Place: %s | Rating: %d | Visit: %s | Comment: %s | Approved: %t\n",
						r.ID, r.PlaceName, r.Rating, r.VisitDate, r.Comment, r.IsApproved)
				}
			}

		case "moderate_review":
			if len(cmdArgs) < 3 {
				fmt.Println("Usage: moderate_review <review_id> <approved> <comment>")
				fmt.Println("Example: moderate_review 123e4567-e89b-12d3-a456-426614174000 true \"Good review\"")
				continue
			}
			reviewID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid review_id: %v\n", err)
				continue
			}
			approved, err := strconv.ParseBool(cmdArgs[1])
			if err != nil {
				fmt.Println("Approved must be true or false")
				continue
			}
			comment := cmdArgs[2]
			req := requests.ModerateReviewRequest{
				ReviewID:    domain.ReviewID(reviewID),
				Approved:    approved,
				Comment:     comment,
				ModeratorID: domain.UserID{},
			}
			resp, err := moderateReviewUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Review moderated:\n  ID: %s\n  Approved: %t\n  Moderated: %t\n  Comment: %s\n",
					resp.ID, resp.IsApproved, resp.IsModerated, resp.ModerationComment)
			}

		case "delete_review":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: delete_review <review_id>")
				continue
			}
			reviewID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid review_id: %v\n", err)
				continue
			}
			req := requests.DeleteReviewRequest{
				ReviewID:    domain.ReviewID(reviewID),
				UserID:      domain.UserID{},
				IsModerator: true,
			}
			resp, err := deleteReviewUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Review deleted: %v\n", resp.Success)
			}

		case "get_page":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: get_page <slug>")
				continue
			}

			ID, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid user_id: %v\n", err)
				continue
			}
			req := content_usecase.GetStaticPageByIDRequest{ID: ID}
			resp, err := getStaticPageUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Static page:\n  ID: %s\n  Slug: %s\n  Title: %s\n  Content: %s\n  Meta: %s\n  Image URL: %s\n  Image Alt: %s\n  Published: %t\n  Published at: %s\n",
					resp.ID, resp.Slug, resp.Title, resp.Content, resp.MetaDesc, resp.ImageURL, resp.ImageAlt, resp.IsPublished, resp.PublishedAt)
			}

		case "create_page":
			if len(cmdArgs) < 3 {
				fmt.Println("Usage: create_page <slug> <title> <content>")
				fmt.Println("Example: create_page about \"About Us\" \"<h1>About</h1><p>Text</p>\"")
				continue
			}
			req := content_usecase.CreateStaticPageRequest{
				Slug:      cmdArgs[0],
				Title:     cmdArgs[1],
				Content:   cmdArgs[2],
				MetaDesc:  "",
				ImageAlt:  "",
				ImageFile: nil,
			}
			resp, err := createStaticPageUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Page created with ID: %s\n", resp.ID)
			}

		case "update_page":
			if len(cmdArgs) < 4 {
				fmt.Println("Usage: update_page <id> <slug> <title> <content>")
				continue
			}
			id, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid ID: %v\n", err)
				continue
			}
			req := content_usecase.UpdateStaticPageRequest{
				ID:      domain.StaticPageID(id),
				Slug:    &cmdArgs[1],
				Title:   &cmdArgs[2],
				Content: &cmdArgs[3],
			}
			err = updateStaticPageUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Page updated successfully")
			}

		case "delete_page":
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: delete_page <id>")
				continue
			}
			id, err := uuid.Parse(cmdArgs[0])
			if err != nil {
				fmt.Printf("Invalid ID: %v\n", err)
				continue
			}
			req := content_usecase.DeleteStaticPageRequest{ID: domain.StaticPageID(id)}
			err = deleteStaticPageUC.Execute(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Page deleted successfully")
			}

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
		}
	}
}

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
	fmt.Println(`GeoGuide CLI - available commands:

  === AUTH ===
  register <username> <email> <password> <phone>
  login <username> <password>
  profile <token>
  logout <token>

  === CATALOG ===
  countries
  country <id>
  cities <country_id>
  place <id>
  places_by_city <city_id>
  places_by_category <category>

  === PLANNER ===
  create_trip <title> <start_date> <end_date> <budget> <user_id>
  get_trip <trip_id>
  user_trips <user_id>
  add_place_to_trip <trip_id> <place_id> <day_number> <duration_min>
  remove_place_from_trip <trip_id> <place_id>
  update_trip <trip_id> [title] [start_date] [end_date] [budget] [status] [notes]
  delete_trip <trip_id>

  === REVIEWS ===
  create_review <user_id> <place_id> <rating> <comment> <visit_date>
  reviews_by_place <place_id>
  user_reviews <user_id>
  moderate_review <review_id> <approved> <comment>
  delete_review <review_id>

  === CONTENT (static pages) ===
  get_page <slug>
  create_page <slug> <title> <content>
  update_page <id> <slug> <title> <content>
  delete_page <id>

  === UTILITIES ===
  seed
  help
  exit
`)
}

func seedDatabase(ctx context.Context, countryRepo interfaces.CountryRepository, cityRepo interfaces.CityRepository, placeRepo interfaces.PlaceRepository, logger *slog.Logger) {
	russia := domain.NewCountryFromDB(
		domain.CountryID(uuid.New()),
		"Russia",
		17100000.0,
		1.7e12,
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
	if err := countryRepo.Save(ctx, russia); err != nil {
		logger.Error("failed to save country", "error", err)
		return
	}
	fmt.Printf("Created country: %s (ID: %s)\n", russia.GetName(), russia.GetId())

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
	if err := cityRepo.Save(ctx, moscow); err != nil {
		logger.Error("failed to save city", "error", err)
		return
	}
	fmt.Printf("Created city: %s (ID: %s)\n", moscow.GetName(), moscow.GetId())

	russia.SetCapitalID(moscow.GetId())
	if err := countryRepo.Update(ctx, russia); err != nil {
		logger.Error("failed to update country capital", "error", err)
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
		"https://example.com/redsquare",
		moscow.GetId(),
		domain.CityDistrictID(uuid.Nil),
		domain.ImageID(uuid.Nil),
	)
	if err := placeRepo.Save(ctx, redSquare); err != nil {
		logger.Error("failed to save place", "error", err)
		return
	}
	fmt.Printf("Created place: %s (ID: %s)\n", redSquare.GetName(), redSquare.GetId())

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
	if err := cityRepo.Save(ctx, petersburg); err != nil {
		logger.Error("failed to save city", "error", err)
		return
	}
	fmt.Printf("Created city: %s (ID: %s)\n", petersburg.GetName(), petersburg.GetId())

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
	if err := placeRepo.Save(ctx, hermitage); err != nil {
		logger.Error("failed to save place", "error", err)
		return
	}
	fmt.Printf("Created place: %s (ID: %s)\n", hermitage.GetName(), hermitage.GetId())
}
