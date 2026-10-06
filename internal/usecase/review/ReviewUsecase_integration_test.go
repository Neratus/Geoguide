//go:build postgres
// +build postgres

package review

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	city_repo "github.com/Neratus/geoguide/internal/repository/city"
	configPkg "github.com/Neratus/geoguide/internal/repository/config"
	country_repo "github.com/Neratus/geoguide/internal/repository/country"
	place_repo "github.com/Neratus/geoguide/internal/repository/place"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

var testPool *pgxpool.Pool
var testQueries *postgreSQL.Queries

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
		return nil, nil, err
	}
	if poolCfg.ConnConfig.RuntimeParams == nil {
		poolCfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolCfg.ConnConfig.RuntimeParams["search_path"] = schema

	sqlDB := stdlib.OpenDB(*poolCfg.ConnConfig)
	if err := goose.SetDialect("postgres"); err != nil {
		sqlDB.Close()
		return nil, nil, err
	}
	if err := goose.Up(sqlDB, migrationsPath); err != nil {
		sqlDB.Close()
		return nil, nil, err
	}
	sqlDB.Close()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, nil, err
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

	pool, teardown, err := setupPostgresForMain(cfg.PostgresConnString(), "../../../migrations")
	if err != nil {
		log.Fatalf("setup postgres: %v", err)
	}

	testPool = pool
	testQueries = postgreSQL.New(pool)

	var testImageID domain.ImageID
	testSlug := "test-usecase-image-" + uuid.NewString()[:8]
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO "StaticPage" (slug, title) VALUES ($1, 'Test') RETURNING id`, testSlug).
		Scan(&testImageID); err != nil {
		log.Fatalf("Failed to create test StaticPage: %v", err)
	}
	builders.DefaultImageID = testImageID

	domain.SetConfig(domain.GetConfig())

	code := m.Run()
	teardown()
	os.Exit(code)
}

func setupTest(t *testing.T) (
	*place_repo.PostgresPlaceRepo,
	*user_repo.PostgresUserRepo,
	*city_repo.PostgresCityRepo,
	*country_repo.PostgresCountryRepo,
) {
	t.Helper()

	placeRepo := &place_repo.PostgresPlaceRepo{Pool: testPool, Queries: testQueries}
	userRepo := &user_repo.PostgresUserRepo{Pool: testPool, Queries: testQueries}
	cityRepo := &city_repo.PostgresCityRepo{Pool: testPool, Queries: testQueries}
	countryRepo := &country_repo.PostgresCountryRepo{Pool: testPool, Queries: testQueries}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tables := []string{"Review", "Place", "City", "Country", "User"}
		for _, table := range tables {
			_, _ = testPool.Exec(ctx, "TRUNCATE TABLE \""+table+"\" RESTART IDENTITY CASCADE;")
		}
	})

	return placeRepo, userRepo, cityRepo, countryRepo
}

type spyTaskQueue struct {
	PublishedTasks []interface{}
}

func (s *spyTaskQueue) Publish(ctx context.Context, task interfaces.Task) error {
	s.PublishedTasks = append(s.PublishedTasks, task)
	return nil
}
func (s *spyTaskQueue) Subscribe(taskName string, handler interfaces.TaskHandler) error { return nil }
func (s *spyTaskQueue) Run(ctx context.Context) error                                   { return nil }
func (s *spyTaskQueue) Close() error                                                    { return nil }

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

func TestCreateReviewUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	placeRepo, userRepo, cityRepo, countryRepo := setupTest(t)
	taskQueue := &spyTaskQueue{}

	uc := NewCreateReviewUseCase(placeRepo, nil, taskQueue, logger)

	t.Run("Success_CreateNewReview", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))

		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))

		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		req := requests.CreateReviewRequest{
			UserID:    user.GetId(),
			PlaceID:   place.GetId(),
			Rating:    5,
			Comment:   "Отлично!",
			VisitDate: time.Now(),
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, 5, resp.Rating)
		require.Equal(t, "Отлично!", resp.Comment)
		require.False(t, resp.IsApproved)
		require.Len(t, taskQueue.PublishedTasks, 1)

		savedReview, err := placeRepo.FindReviewByID(ctx, resp.ID)
		require.NoError(t, err)
		require.Equal(t, "Отлично!", savedReview.GetComment())
	})

	t.Run("Error_PlaceNotFound", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		req := requests.CreateReviewRequest{
			UserID:    user.GetId(),
			PlaceID:   domain.PlaceID(uuid.New()),
			Rating:    5,
			Comment:   "Отлично!",
			VisitDate: time.Now(),
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrPlaceNotFound)
		require.Nil(t, resp)
	})

	t.Run("Error_ReviewAlreadyExists", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		existingReview := builders.NewReviewBuilder().
			WithUserID(user.GetId()).
			WithPlaceID(place.GetId()).
			Build()
		require.NoError(t, placeRepo.SaveReview(ctx, existingReview))

		require.NoError(t, placeRepo.ModerateReview(ctx, existingReview.GetId(), true, ""))

		req := requests.CreateReviewRequest{
			UserID:    user.GetId(),
			PlaceID:   place.GetId(),
			Rating:    5,
			Comment:   "Еще один отзыв",
			VisitDate: time.Now(),
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrReviewAlreadyExists)
		require.Nil(t, resp)
	})
}

func TestDeleteReviewUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	placeRepo, userRepo, cityRepo, countryRepo := setupTest(t)

	uc := NewDeleteReviewUseCase(placeRepo, logger)

	t.Run("Success_DeleteByAuthor", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		author := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, author))

		review := builders.NewReviewBuilder().
			WithUserID(author.GetId()).
			WithPlaceID(place.GetId()).
			Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review))

		req := requests.DeleteReviewRequest{
			ReviewID:    review.GetId(),
			UserID:      author.GetId(),
			IsModerator: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)

		_, err = placeRepo.FindReviewByID(ctx, review.GetId())
		require.Error(t, err)
	})

	t.Run("Success_DeleteByModerator", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		author := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, author))
		moderator := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, moderator))

		review := builders.NewReviewBuilder().
			WithUserID(author.GetId()).
			WithPlaceID(place.GetId()).
			Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review))

		req := requests.DeleteReviewRequest{
			ReviewID:    review.GetId(),
			UserID:      moderator.GetId(),
			IsModerator: true,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)
	})

	t.Run("Error_ReviewNotFound", func(t *testing.T) {
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		req := requests.DeleteReviewRequest{
			ReviewID:    domain.ReviewID(uuid.New()),
			UserID:      user.GetId(),
			IsModerator: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrReviewNotFound)
		require.Nil(t, resp)
	})

	t.Run("Error_NotAuthorized", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		author := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, author))
		randomUser := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, randomUser))

		review := builders.NewReviewBuilder().
			WithUserID(author.GetId()).
			WithPlaceID(place.GetId()).
			Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review))

		req := requests.DeleteReviewRequest{
			ReviewID:    review.GetId(),
			UserID:      randomUser.GetId(),
			IsModerator: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrNotAuthorized)
		require.Nil(t, resp)
	})
}

func TestGetReviewsByPlaceUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	placeRepo, userRepo, cityRepo, countryRepo := setupTest(t)

	uc := NewGetReviewsByPlaceUseCase(placeRepo, logger)

	t.Run("Success_GetAllReviews", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		review1 := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place.GetId()).WithRating(5).Build()
		review1.SetApproved(true)
		require.NoError(t, placeRepo.SaveReview(ctx, review1))

		review2 := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place.GetId()).WithRating(2).Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review2))

		require.NoError(t, placeRepo.ModerateReview(ctx, review1.GetId(), true, ""))
		require.NoError(t, placeRepo.ModerateReview(ctx, review2.GetId(), true, ""))

		req := requests.GetReviewsByPlaceRequest{
			PlaceID:      place.GetId(),
			OnlyApproved: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 2)
	})

	t.Run("Success_GetOnlyApprovedReviews", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		review1 := builders.NewReviewBuilder().WithUserID(user.GetId()).WithPlaceID(place.GetId()).WithRating(5).Build()
		review1.SetApproved(true)
		require.NoError(t, placeRepo.SaveReview(ctx, review1))

		require.NoError(t, placeRepo.ModerateReview(ctx, review1.GetId(), true, ""))

		req := requests.GetReviewsByPlaceRequest{
			PlaceID:      place.GetId(),
			OnlyApproved: true,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, "Отличное место!", resp[0].Comment)
	})

	t.Run("Error_PlaceNotFound", func(t *testing.T) {
		req := requests.GetReviewsByPlaceRequest{
			PlaceID:      domain.PlaceID(uuid.New()),
			OnlyApproved: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrPlaceNotFound)
		require.Nil(t, resp)
	})
}

func TestModerateReviewUseCase_Integration(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	placeRepo, userRepo, cityRepo, countryRepo := setupTest(t)
	taskQueue := &spyTaskQueue{}

	uc := NewModerateReviewUseCase(placeRepo, userRepo, nil, taskQueue, logger)

	t.Run("Success_ModerateAndApprove", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))

		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		review := builders.NewReviewBuilder().
			WithUserID(user.GetId()).
			WithPlaceID(place.GetId()).
			Build()
		require.NoError(t, placeRepo.SaveReview(ctx, review))

		req := requests.ModerateReviewRequest{
			ReviewID: review.GetId(),
			Approved: true,
			Comment:  "Все отлично",
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsModerated)
		require.True(t, resp.IsApproved)
		require.Equal(t, "Все отлично", resp.ModerationComment)
		require.Len(t, taskQueue.PublishedTasks, 1)

		updatedReview, err := placeRepo.FindReviewByID(ctx, review.GetId())
		require.NoError(t, err)
		require.True(t, updatedReview.IsModerated())
		require.True(t, updatedReview.IsApproved())
	})

	t.Run("Error_ReviewNotFound", func(t *testing.T) {
		req := requests.ModerateReviewRequest{
			ReviewID: domain.ReviewID(uuid.New()),
			Approved: true,
			Comment:  "Ок",
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrReviewNotFound)
		require.Nil(t, resp)
	})

	t.Run("Error_ReviewAlreadyModerated", func(t *testing.T) {
		country := builders.NewCountryBuilder().Build()
		require.NoError(t, countryRepo.Save(ctx, country))
		city := builders.NewCityBuilder().WithCountryID(country.GetId()).Build()
		require.NoError(t, cityRepo.Save(ctx, city))
		place := builders.NewPlaceBuilder().WithCityID(city.GetId()).Build()
		require.NoError(t, placeRepo.Save(ctx, place))
		user := builders.NewUserBuilder().Build()
		require.NoError(t, userRepo.Save(ctx, user))

		alreadyModeratedReview := builders.NewReviewBuilder().
			WithUserID(user.GetId()).
			WithPlaceID(place.GetId()).
			Build()
		require.NoError(t, placeRepo.SaveReview(ctx, alreadyModeratedReview))

		require.NoError(t, placeRepo.ModerateReview(ctx, alreadyModeratedReview.GetId(), false, ""))

		req := requests.ModerateReviewRequest{
			ReviewID: alreadyModeratedReview.GetId(),
			Approved: false,
			Comment:  "Повторная модерация",
		}

		resp, err := uc.Execute(ctx, req)

		require.ErrorIs(t, err, usecase_errors.ErrReviewAlreadyModerated)
		require.Nil(t, resp)
	})
}
