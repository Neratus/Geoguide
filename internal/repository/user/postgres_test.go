//go:build postgres
// +build postgres

package user_repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	configPkg "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

var testPool *pgxpool.Pool
var testQueries *postgreSQL.Queries

func TestMain(m *testing.M) {
	cfg, err := configPkg.LoadTest()
	if err != nil {
		log.Fatalf("Failed to load test config: %v", err)
	}

	connStr := cfg.PostgresConnString()
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	if err := goose.Up(sqlDB, "../../../migrations"); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}
	sqlDB.Close()

	testPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	testQueries = postgreSQL.New(testPool)

	var testImageID domain.ImageID
	testSlug := fmt.Sprintf("test-image-slug-%s", uuid.New().String()[:8])

	err = testPool.QueryRow(context.Background(),
		`INSERT INTO "StaticPage" (slug, title) VALUES ($1, 'Test Image') RETURNING id`, testSlug).
		Scan(&testImageID)
	if err != nil {
		log.Fatalf("Failed to create test StaticPage: %v", err)
	}
	builders.DefaultImageID = testImageID

	domain.SetConfig(domain.GetConfig())

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func setupTest(t *testing.T) *PostgresUserRepo {

	t.Helper()
	tx, err := testPool.Begin(context.Background())
	require.NoError(t, err)

	queries := postgreSQL.New(tx)
	repo := &PostgresUserRepo{
		Pool:    testPool,
		Queries: queries,
	}

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	return repo
}

func TestPostgresUserRepository(t *testing.T) {

	t.Run("Save_Valid", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().
			WithUsername("newuser").
			WithEmail("new@example.com").
			Build()

		err := repo.Save(ctx, user)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, user.GetId())
	})

	t.Run("FindByID_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().WithUsername("findme").Build()
		require.NoError(t, repo.Save(ctx, user))

		found, err := repo.FindByID(ctx, user.GetId())

		require.NoError(t, err)
		require.Equal(t, user.GetId(), found.GetId())
		require.Equal(t, "findme", found.GetUsername())
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		nonExistentID := domain.UserID(uuid.New())

		found, _ := repo.FindByID(ctx, nonExistentID)

		require.Nil(t, found)
	})

	t.Run("FindByID_InvalidID", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		invalidID := domain.UserID(uuid.Nil)

		found, _ := repo.FindByID(ctx, invalidID)

		require.Nil(t, found)
	})

	t.Run("FindByUsername_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().WithUsername("unique_username").Build()
		require.NoError(t, repo.Save(ctx, user))

		found, err := repo.FindByUsername(ctx, "unique_username")

		require.NoError(t, err)
		require.Equal(t, user.GetId(), found.GetId())
	})

	t.Run("FindByUsername_NotFound", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()

		found, err := repo.FindByUsername(ctx, "nonexistent_user")

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("FindByEmail_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().WithEmail("unique@email.com").Build()
		require.NoError(t, repo.Save(ctx, user))

		found, err := repo.FindByEmail(ctx, "unique@email.com")

		require.NoError(t, err)
		require.Equal(t, user.GetId(), found.GetId())
	})

	t.Run("FindByEmail_NotFound", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()

		found, err := repo.FindByEmail(ctx, "nonexistent@email.com")

		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("Update_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().WithUsername("toupdate").Build()
		require.NoError(t, repo.Save(ctx, user))

		user.SetUsername("updated_name")
		err := repo.Update(ctx, user)

		require.NoError(t, err)
		updated, err := repo.FindByID(ctx, user.GetId())
		require.NoError(t, err)
		require.Equal(t, "updated_name", updated.GetUsername())
	})

	t.Run("Update_PartialSuccess", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().WithUsername("partial").WithCountry("France").Build()
		require.NoError(t, repo.Save(ctx, user))
		originalUsername := user.GetUsername()

		user.SetCountryOfResidence("Germany")
		err := repo.Update(ctx, user)

		require.NoError(t, err)
		updated, err := repo.FindByID(ctx, user.GetId())
		require.NoError(t, err)
		require.Equal(t, "Germany", updated.GetCountryOfResidence())
		require.Equal(t, originalUsername, updated.GetUsername())
	})

	t.Run("Delete_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().WithUsername("todelete").Build()
		require.NoError(t, repo.Save(ctx, user))

		err := repo.Delete(ctx, user.GetId())

		require.NoError(t, err)
		user, _ = repo.FindByID(ctx, user.GetId())
		require.Nil(t, user)
	})

	t.Run("BlockUser_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().Build()
		require.NoError(t, repo.Save(ctx, user))

		err := repo.BlockUser(ctx, user.GetId(), "spam")

		require.NoError(t, err)
		found, err := repo.FindByID(ctx, user.GetId())
		require.NoError(t, err)
		require.True(t, found.IsBlocked())
		require.Equal(t, "spam", found.GetBlockReason())
		require.NotNil(t, found.GetBlockedAt())
	})

	t.Run("UnblockUser_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().Build()
		require.NoError(t, repo.Save(ctx, user))
		require.NoError(t, repo.BlockUser(ctx, user.GetId(), "spam"))

		err := repo.UnblockUser(ctx, user.GetId())

		require.NoError(t, err)
		found, err := repo.FindByID(ctx, user.GetId())
		require.NoError(t, err)
		require.False(t, found.IsBlocked())
		require.Empty(t, found.GetBlockReason())
		require.Nil(t, found.GetBlockedAt())
	})

	t.Run("FindAll_Found", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		require.NoError(t, repo.Save(ctx, builders.NewUserBuilder().WithUsername("user1").Build()))
		require.NoError(t, repo.Save(ctx, builders.NewUserBuilder().WithUsername("user2").Build()))

		users, err := repo.FindAll(ctx, 10, 0, "")

		require.NoError(t, err)
		require.Len(t, users, 2)
	})

	t.Run("FindAll_Empty", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()

		users, err := repo.FindAll(ctx, 10, 0, "nonexistent_search_term")

		require.NoError(t, err)
		require.Empty(t, users)
	})

	t.Run("UpdateEmailVerified_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().Build()
		require.NoError(t, repo.Save(ctx, user))

		err := repo.UpdateEmailVerified(ctx, user.GetId(), true)

		require.NoError(t, err)

	})

	t.Run("UpdatePhoneVerified_Success", func(t *testing.T) {
		repo := setupTest(t)
		ctx := context.Background()
		user := builders.NewUserBuilder().Build()
		require.NoError(t, repo.Save(ctx, user))

		err := repo.UpdatePhoneVerified(ctx, user.GetId(), true)

		require.NoError(t, err)
		found, err := repo.FindByID(ctx, user.GetId())
		require.NoError(t, err)
		require.True(t, found.IsPhoneVerified())
	})
}
