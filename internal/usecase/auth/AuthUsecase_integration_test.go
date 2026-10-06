//go:build postgres
// +build postgres

package auth_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	configPkg "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	session_repo "github.com/Neratus/geoguide/internal/repository/session"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/Neratus/geoguide/internal/usecase/auth"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

var testPool *pgxpool.Pool
var testQueries *postgreSQL.Queries
var testRedisClient *redis.Client
var testSession *session_repo.RedisSessionRepo

func setupPostgresForMain(baseConnStr, migrationsPath string) (*pgxpool.Pool, func(), error) {
	ctx := context.Background()
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	admin, err := pgx.Connect(ctx, baseConnStr)
	if err != nil {
		return nil, nil, fmt.Errorf("admin connect: %w", err)
	}
	if _, err := admin.Exec(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		_ = admin.Close(ctx)
		return nil, nil, fmt.Errorf("create schema: %w", err)
	}
	_ = admin.Close(ctx)

	poolCfg, err := pgxpool.ParseConfig(baseConnStr)
	if err != nil {
		return nil, nil, fmt.Errorf("parse config: %w", err)
	}
	if poolCfg.ConnConfig.RuntimeParams == nil {
		poolCfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolCfg.ConnConfig.RuntimeParams["search_path"] = schema

	sqlDB := stdlib.OpenDB(*poolCfg.ConnConfig)
	if err := goose.SetDialect("postgres"); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(sqlDB, migrationsPath); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("goose up: %w", err)
	}
	sqlDB.Close()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("pool: %w", err)
	}

	teardown := func() {
		pool.Close()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		conn, err := pgx.Connect(c, baseConnStr)
		if err == nil {
			_, _ = conn.Exec(c, `DROP SCHEMA "`+schema+`" CASCADE`)
			_ = conn.Close(c)
		}
	}
	return pool, teardown, nil
}

func TestMain(m *testing.M) {
	cfg, err := configPkg.LoadTest()
	if err != nil {
		log.Fatalf("Failed to load test config: %v", err)
	}

	pool, teardownPG, err := setupPostgresForMain(cfg.PostgresConnString(), "../../../migrations")
	if err != nil {
		log.Fatalf("setup postgres: %v", err)
	}
	testPool = pool
	testQueries = postgreSQL.New(pool)

	var testImageID domain.ImageID
	testSlug := "test-auth-image-" + uuid.NewString()[:8]
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO "StaticPage" (slug, title) VALUES ($1, 'Test') RETURNING id`, testSlug).
		Scan(&testImageID); err != nil {
		log.Fatalf("Failed to create test StaticPage: %v", err)
	}
	builders.DefaultImageID = testImageID

	addr := fmt.Sprintf("%s:%d", cfg.Database.Redis.Host, cfg.Database.Redis.Port)
	dbIndex := 1 + int(uuid.New()[15])%15

	testRedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Database.Redis.Password,
		DB:       dbIndex,
	})
	if err := testRedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis ping (db=%d): %v", dbIndex, err)
	}
	if err := testRedisClient.FlushDB(context.Background()).Err(); err != nil {
		log.Fatalf("redis flushdb (db=%d): %v", dbIndex, err)
	}
	log.Printf("auth integration tests use redis db=%d", dbIndex)

	testSession = session_repo.NewSessionRepoWithClient(testRedisClient)

	domain.SetConfig(domain.GetConfig())

	code := m.Run()

	_ = testRedisClient.FlushDB(context.Background()).Err()
	_ = testRedisClient.Close()
	teardownPG()
	os.Exit(code)
}

func setupTest(t *testing.T) *user_repo.PostgresUserRepo {
	t.Helper()

	if err := testRedisClient.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("redis flush: %v", err)
	}

	repo := &user_repo.PostgresUserRepo{Pool: testPool, Queries: testQueries}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tables := []string{"Review", "TripPlace", "Trip", "Place", "City", "Country", "User"}
		for _, table := range tables {
			_, _ = testPool.Exec(ctx, "TRUNCATE TABLE \""+table+"\" RESTART IDENTITY CASCADE;")
		}
		_ = testRedisClient.FlushDB(context.Background()).Err()
	})

	return repo
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type notifierSpy struct {
	EmailSent     bool
	LastCode      string
	LastRecipient string
}

func (n *notifierSpy) Send(ctx context.Context, msg interfaces.NotifyMessage) error { return nil }
func (n *notifierSpy) SendTemplate(ctx context.Context, to, templateName string, data interface{}) error {
	return nil
}
func (n *notifierSpy) SendWelcomeEmail(ctx context.Context, to, username string) error { return nil }
func (n *notifierSpy) SendUserBlockedNotification(ctx context.Context, to, username, reason string) error {
	return nil
}
func (n *notifierSpy) SendUserUnblockedNotification(ctx context.Context, to, username string) error {
	return nil
}
func (n *notifierSpy) SendLoginNotification(ctx context.Context, to, username, ip, userAgent string) error {
	return nil
}
func (n *notifierSpy) SendVerificationEmail(ctx context.Context, to, code string) error {
	n.EmailSent = true
	n.LastCode = code
	n.LastRecipient = to
	return nil
}

type staticStub struct{}

func (s *staticStub) Save(ctx context.Context, page *domain.StaticPage, reader io.Reader) error {
	return nil
}
func (s *staticStub) FindByID(ctx context.Context, id uuid.UUID) (*domain.StaticPage, error) {
	return nil, nil
}
func (s *staticStub) Update(ctx context.Context, page *domain.StaticPage, reader io.Reader) error {
	return nil
}
func (s *staticStub) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	return nil
}
func (s *staticStub) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (s *staticStub) DeleteFile(ctx context.Context, fileID string) error { return nil }
func (s *staticStub) FindBySlug(ctx context.Context, slug string) (*domain.StaticPage, error) {
	return nil, nil
}
func (s *staticStub) GetFile(ctx context.Context, fileID domain.ImageID) (io.ReadCloser, error) {
	return nil, nil
}
func (s *staticStub) GetFileURL(ctx context.Context, fileID string) (string, error) {
	return "https://cdn.example.com/" + fileID, nil
}

func TestRegisterUserUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	userRepo := setupTest(t)
	notifier := &notifierSpy{}
	taskQueue := &taskQueueSpy{}
	rateLimiter := &allowAllRateLimiter{}

	uc := auth.NewRegisterUserUseCase(userRepo, testSession, notifier, taskQueue, logger, rateLimiter)

	t.Run("Success_Register", func(t *testing.T) {
		req := requests.RegisterUserRequest{
			Username: "newuser_int",
			Email:    "new_int@example.com",
			Password: "StrongPass123!",
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Equal(t, "verification codes sent", resp.Message)
		require.True(t, taskQueue.PublishCalled)
		require.Len(t, taskQueue.PublishedTasks, 1)

		savedUser, err := userRepo.FindByEmail(ctx, req.Email)
		require.NoError(t, err)
		require.Equal(t, "newuser_int", savedUser.GetUsername())
		require.False(t, savedUser.IsEmailVerified())
	})

	t.Run("Error_UserAlreadyExists", func(t *testing.T) {
		existingUser := builders.NewUserBuilder().
			WithUsername("existing_user").
			WithEmail("exist@example.com").
			Build()
		require.NoError(t, userRepo.Save(ctx, existingUser))

		req := requests.RegisterUserRequest{
			Username: "newuser2",
			Email:    "exist@example.com",
			Password: "StrongPass123!",
		}

		resp, err := uc.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrUserAlreadyExists)
		require.Nil(t, resp)
	})

	t.Run("Error_RateLimited", func(t *testing.T) {
		ucRestricted := auth.NewRegisterUserUseCase(
			userRepo, testSession, notifier, taskQueue, logger, &restrictedRateLimiter{},
		)

		req := requests.RegisterUserRequest{
			Username: "spammer",
			Email:    "spam@example.com",
			Password: "StrongPass123!",
		}

		resp, err := ucRestricted.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrTooManyAttempts)
		require.Nil(t, resp)
	})
}

func TestVerifyContactUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	userRepo := setupTest(t)
	notifier := &notifierSpy{}
	taskQueue := &taskQueueSpy{}

	registerUC := auth.NewRegisterUserUseCase(
		userRepo, testSession, notifier, taskQueue, logger, &allowAllRateLimiter{},
	)
	verifyUC := auth.NewVerifyContactUseCase(userRepo, testSession, notifier, taskQueue, logger)

	t.Run("Success_VerifyEmail", func(t *testing.T) {
		registerReq := requests.RegisterUserRequest{
			Username: "verify_user",
			Email:    "verify@example.com",
			Password: "StrongPass123!",
		}
		_, err := registerUC.Execute(ctx, registerReq)
		require.NoError(t, err)

		user, err := userRepo.FindByEmail(ctx, registerReq.Email)
		require.NoError(t, err)

		var code string
		prefix := "verif:" + user.GetId().String() + ":"
		iter := testRedisClient.Scan(ctx, 0, prefix+"*", 10).Iterator()
		for iter.Next(ctx) {
			val, err := testRedisClient.Get(ctx, iter.Val()).Result()
			if err == nil {
				code = val
				break
			}
		}
		require.NoError(t, iter.Err())
		require.NotEmpty(t, code, "registration must save verification code into Redis")

		verifyReq := requests.VerifyContactRequest{
			UserID:  user.GetId(),
			Contact: user.GetEmail(),
			Code:    code,
		}
		require.NoError(t, verifyUC.Execute(ctx, verifyReq))

		updatedUser, err := userRepo.FindByID(ctx, user.GetId())
		require.NoError(t, err)
		require.True(t, updatedUser.IsEmailVerified())
	})
}

func TestGetUserProfileUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	userRepo := setupTest(t)
	staticRepo := &staticStub{}

	uc := auth.NewGetUserProfileUseCase(userRepo, testSession, staticRepo, logger)

	t.Run("Success_GetProfile", func(t *testing.T) {
		user := builders.NewUserBuilder().
			WithUsername("profile_user").
			WithEmail("profile@example.com").
			Build()
		require.NoError(t, userRepo.Save(ctx, user))

		req := requests.GetUserProfileRequest{UserID: user.GetId()}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Equal(t, "profile_user", resp.Username)
		require.Equal(t, "profile@example.com", resp.Email)
	})

	t.Run("Error_UserNotFound", func(t *testing.T) {
		req := requests.GetUserProfileRequest{UserID: domain.UserID(uuid.New())}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func TestUpdateUserProfileUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	userRepo := setupTest(t)

	uc := auth.NewUpdateUserProfileUseCase(userRepo, testSession, logger)

	t.Run("Success_UpdateProfile", func(t *testing.T) {
		user := builders.NewUserBuilder().
			WithUsername("old_name").
			WithEmail("update@example.com").
			Build()
		require.NoError(t, userRepo.Save(ctx, user))

		require.NoError(t, testSession.Set(ctx, "valid-token", uuid.UUID(user.GetId()), time.Hour))

		var rt uuid.UUID
		require.NoError(t, testSession.Get(ctx, "valid-token", &rt))
		require.Equal(t, uuid.UUID(user.GetId()), rt)

		newUsername := "updated_user"
		req := requests.UpdateUserProfileRequest{
			Token:    "valid-token",
			Username: &newUsername,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Equal(t, "updated_user", resp.Username)

		updatedUser, err := userRepo.FindByID(ctx, user.GetId())
		require.NoError(t, err)
		require.Equal(t, "updated_user", updatedUser.GetUsername())
	})

	t.Run("Error_InvalidSession", func(t *testing.T) {
		newUsername := "hacker"
		req := requests.UpdateUserProfileRequest{
			Token:    "non-existent-token",
			Username: &newUsername,
		}

		resp, err := uc.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrInvalidSession)
		require.Nil(t, resp)
	})

	t.Run("Error_UserBlocked", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))
		require.NoError(t, userRepo.BlockUser(ctx, user.GetId(), "Spam"))

		require.NoError(t, testSession.Set(ctx, "blocked-token", uuid.UUID(user.GetId()), time.Hour))

		newUsername := "blocked_user"
		req := requests.UpdateUserProfileRequest{
			Token:    "blocked-token",
			Username: &newUsername,
		}

		resp, err := uc.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrUserBlocked)
		require.Nil(t, resp)
	})
}

func TestLogoutUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	uc := auth.NewLogoutUseCase(testSession, logger)

	t.Run("Success_Logout", func(t *testing.T) {
		token := "session-token-xyz"
		require.NoError(t, testSession.Set(ctx, token, "any-user", time.Hour))

		req := requests.LogoutRequest{Token: token}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.True(t, resp.Success)
		require.False(t, testSession.Exists(ctx, token))
	})

	t.Run("Error_InvalidSession", func(t *testing.T) {
		req := requests.LogoutRequest{Token: "fake-token"}
		resp, err := uc.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrInvalidSession)
		require.Nil(t, resp)
	})
}
