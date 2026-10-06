package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/delivery"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/repository"
	"github.com/Neratus/geoguide/internal/repository/config"
	"github.com/Neratus/geoguide/internal/services"
	"github.com/Neratus/geoguide/internal/usecase/admin"
	"github.com/Neratus/geoguide/internal/usecase/auth"
	"github.com/Neratus/geoguide/internal/usecase/catalog"
	"github.com/Neratus/geoguide/internal/usecase/favourites"
	"github.com/Neratus/geoguide/internal/usecase/planner"
	"github.com/Neratus/geoguide/internal/usecase/review"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type noopNotifier struct{}

func (n *noopNotifier) Send(ctx context.Context, msg interfaces.NotifyMessage) error { return nil }
func (n *noopNotifier) SendTemplate(ctx context.Context, to, templateName string, data interface{}) error {
	return nil
}
func (n *noopNotifier) SendWelcomeEmail(ctx context.Context, to, username string) error { return nil }
func (n *noopNotifier) SendUserBlockedNotification(ctx context.Context, to, username, reason string) error {
	return nil
}
func (n *noopNotifier) SendUserUnblockedNotification(ctx context.Context, to, username string) error {
	return nil
}
func (n *noopNotifier) SendLoginNotification(ctx context.Context, to, username, ip, userAgent string) error {
	return nil
}
func (n *noopNotifier) SendVerificationEmail(ctx context.Context, to, code string) error { return nil }

type noopTaskQueue struct{}

func (q *noopTaskQueue) Publish(ctx context.Context, task interfaces.Task) error         { return nil }
func (q *noopTaskQueue) Subscribe(taskName string, handler interfaces.TaskHandler) error { return nil }
func (q *noopTaskQueue) Run(ctx context.Context) error                                   { return nil }
func (q *noopTaskQueue) Close() error                                                    { return nil }

type testEnv struct {
	dbURL  string
	appURL string
	cfg    *config.Config
	repos  *repository.Repositories
}

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		log.Fatalf("failed to chdir to project root: %v", err)
	}
	log.Println("Starting GeoGuide E2E test suite with Testcontainers...")
	code := m.Run()
	log.Println("GeoGuide E2E test suite finished.")
	os.Exit(code)
}

func startTestEnv(t *testing.T) testEnv {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	postgresContainer, err := tcpostgres.Run(ctx,
		"postgres:15-alpine",
		tcpostgres.WithDatabase("geoguide_e2e"),
		tcpostgres.WithUsername("app_user"),
		tcpostgres.WithPassword("app_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	localstackContainer, err := localstack.Run(ctx,
		"localstack/localstack:3.0.2",
		testcontainers.WithEnv(map[string]string{
			"LOCALSTACK_SERVICES": "s3",
			"DEBUG":               "1",
		}),
	)
	if err != nil {
		t.Fatalf("failed to start localstack container: %v", err)
	}
	t.Cleanup(func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate localstack container: %v", err)
		}
	})

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start Redis: %v", err)
	}

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get Redis host: %v", err)
	}

	redisPort, err := redisContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("failed to get Redis mapped port: %v", err)
	}

	t.Logf("e2e: Redis endpoint = %s:%s", redisHost, redisPort.Port())

	host, err := localstackContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get localstack host: %v", err)
	}

	mappedPort, err := localstackContainer.MappedPort(ctx, "4566/tcp")
	if err != nil {
		t.Fatalf("failed to get localstack mapped port: %v", err)
	}

	endpoint := fmt.Sprintf("%s:%s", host, mappedPort.Port())

	if strings.TrimSpace(endpoint) == "" {
		t.Fatal("localstack endpoint is empty")
	}

	t.Logf("e2e: LocalStack endpoint = %q", endpoint)

	dbURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	if err := applyMigrations(dbURL); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	log.Println("e2e: Database migrations applied successfully")

	tmpDir := t.TempDir()
	configPath := fmt.Sprintf("%s/config.yaml", tmpDir)
	parsed, err := pgx.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("failed to parse dbURL %q: %v", dbURL, err)
	}

	tmpl, err := template.ParseFiles("internal/e2e/testdata/config_template.yaml")
	if err != nil {
		t.Fatalf("failed to parse config template: %v", err)
	}

	data := map[string]interface{}{
		"Host":      parsed.Host,
		"Port":      parsed.Port,
		"User":      parsed.User,
		"Password":  parsed.Password,
		"Database":  parsed.Database,
		"Endpoint":  endpoint,
		"RedisHost": redisHost,
		"RedisPort": func() int {
			p, err := strconv.Atoi(redisPort.Port())
			if err != nil {
				t.Fatalf("failed to parse Redis port %q: %v", redisPort.Port(), err)
			}
			return p
		}(),
	}

	var configContent bytes.Buffer
	if err := tmpl.Execute(&configContent, data); err != nil {
		t.Fatalf("failed to execute config template: %v", err)
	}
	t.Logf("e2e: generated config:\n%s", configContent.String())

	if err := os.WriteFile(configPath, configContent.Bytes(), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.LoadFile(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	t.Logf("e2e: loaded MinIO endpoint = %q", cfg.Database.Minio.Endpoint)

	repos, err := repository.NewRepositories(cfg)
	if err != nil {
		t.Fatalf("failed to init repositories: %v", err)
	}
	t.Cleanup(func() { repos.Close() })

	if err := seedTestData(ctx, repos); err != nil {
		t.Logf("seed warning (non-fatal): %v", err)
	}

	services.InitJWT(cfg.Database.Auth.JWTSecret)
	userRepo := repos.UserRepo
	sessionRepo := repos.SessionRepo
	placeRepo := repos.PlaceRepo
	cityRepo := repos.CityRepo
	countryRepo := repos.CountryRepo
	tripRepo := repos.TripRepo
	favouriteRepo := repos.FavouriteRepo
	staticRepo := repos.StaticRepo
	rateLimiter, _ := services.NewRateLimiter(cfg)

	loginUC := auth.NewLoginUseCase(userRepo, sessionRepo, &noopNotifier{}, &noopTaskQueue{}, slog.Default(), rateLimiter)
	logoutUC := auth.NewLogoutUseCase(sessionRepo, slog.Default())
	getProfileUC := auth.NewGetUserProfileUseCase(userRepo, sessionRepo, staticRepo, slog.Default())
	getCountriesUC := catalog.NewGetCountriesUseCase(countryRepo, cityRepo, staticRepo, slog.Default())
	getCountryUC := catalog.NewGetCountryUseCase(countryRepo, cityRepo, staticRepo, slog.Default())
	getCitiesUC := catalog.NewGetCitiesByCountryUseCase(cityRepo, staticRepo, slog.Default())
	getCityUC := catalog.NewGetCityUseCase(cityRepo, staticRepo, slog.Default())
	getPlaceUC := catalog.NewGetPlaceUseCase(placeRepo, userRepo, staticRepo, slog.Default())
	getPlacesByCityUC := catalog.NewGetPlacesByCityUseCase(placeRepo, staticRepo, slog.Default())
	createTripUC := planner.NewCreateTripUseCase(tripRepo, slog.Default())
	getTripUC := planner.NewGetTripUseCase(tripRepo, slog.Default())
	getUserTripsUC := planner.NewGetUserTripsUseCase(tripRepo, slog.Default())
	addPlaceToTripUC := planner.NewAddPlaceToTripUseCase(tripRepo, placeRepo, slog.Default())
	removePlaceFromTripUC := planner.NewRemovePlaceFromTripUseCase(tripRepo, slog.Default())
	updateTripUC := planner.NewUpdateTripUseCase(tripRepo, slog.Default())
	deleteTripUC := planner.NewDeleteTripUseCase(tripRepo, slog.Default())
	createReviewUC := review.NewCreateReviewUseCase(placeRepo, nil, nil, slog.Default())
	getReviewsByPlaceUC := review.NewGetReviewsByPlaceUseCase(placeRepo, slog.Default())
	getUserReviewsUC := review.NewGetUserReviewsUseCase(placeRepo, slog.Default())
	deleteReviewUC := review.NewDeleteReviewUseCase(placeRepo, slog.Default())
	moderateReviewUC := review.NewModerateReviewUseCase(placeRepo, userRepo, nil, nil, slog.Default())
	addFavUC := favourites.NewAddFavouriteUseCase(favouriteRepo, placeRepo, slog.Default())
	removeFavUC := favourites.NewRemoveFavouriteUseCase(favouriteRepo, slog.Default())
	getFavUC := favourites.NewGetFavouritesUseCase(favouriteRepo, slog.Default())
	getUsersUC := admin.NewGetUsersUseCase(userRepo, slog.Default())
	getPendingReviewsUC := admin.NewGetPendingReviewsUseCase(placeRepo, slog.Default())

	server := delivery.NewServer(
		nil, nil,
		loginUC, logoutUC,
		nil, nil,
		getProfileUC, nil, nil,
		nil,
		getCountriesUC, getCountryUC, getCitiesUC, getCityUC, nil,
		nil, nil,
		getPlacesByCityUC, nil, getPlaceUC,
		createTripUC, getTripUC, getUserTripsUC, updateTripUC, deleteTripUC,
		addPlaceToTripUC, removePlaceFromTripUC,
		createReviewUC, getReviewsByPlaceUC, getUserReviewsUC, deleteReviewUC, moderateReviewUC,
		nil, nil,
		getUsersUC, getPendingReviewsUC,
		addFavUC, removeFavUC, getFavUC, nil,
		cfg,
	)

	jwtSecret := []byte(cfg.Database.Auth.JWTSecret)
	router := delivery.NewRouter(server, jwtSecret)
	appServer := httptest.NewServer(router)
	t.Cleanup(appServer.Close)

	return testEnv{
		dbURL:  dbURL,
		appURL: appServer.URL,
		cfg:    cfg,
		repos:  repos,
	}
}

func applyMigrations(dbURL string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()
	return goose.Up(db, "migrations")
}
