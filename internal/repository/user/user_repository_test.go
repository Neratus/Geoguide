package user_repo

import (
	"context"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func generateTestUser() *domain.User {
	hash, _ := domain.HashPassword("password123")

	user, err := domain.NewUser(
		uuid.Nil,
		"testuser",
		"test@example.com",
		hash,
		"+1234567890",
		nil,
		"France",
		domain.UserRoleUser,
	)
	if err != nil {
		panic(err)
	}
	return user
}

func userRepository_SaveNew(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
	require.NotZero(t, user.GetId())
}

func userRepository_SaveExisting(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
	id := user.GetId()

	user.SetUsername("newname")
	err = repo.Save(context.Background(), user)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "newname", found.GetUsername())
}

func userRepository_FindByID_NotFound(t *testing.T, repo interfaces.UserRepository) {
	_, err := repo.FindByID(context.Background(), domain.Int64ToUUID(9999))
	require.NoError(t, err)
}

func userRepository_FindByID_Found(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
	id := user.GetId()

	found, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, user.GetId(), found.GetId())
	require.Equal(t, user.GetUsername(), found.GetUsername())
	require.Equal(t, user.GetEmail(), found.GetEmail())
}

func userRepository_FindByUsername_NotFound(t *testing.T, repo interfaces.UserRepository) {
	_, err := repo.FindByUsername(context.Background(), "nonexistent")
	require.Error(t, err)
}

func userRepository_FindByUsername_Found(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)

	found, err := repo.FindByUsername(context.Background(), user.GetUsername())
	require.NoError(t, err)
	require.Equal(t, user.GetId(), found.GetId())
}

func userRepository_FindByEmail_NotFound(t *testing.T, repo interfaces.UserRepository) {
	_, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")
	require.Error(t, err)
}

func userRepository_FindByEmail_Found(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)

	found, err := repo.FindByEmail(context.Background(), (user.GetEmail()))
	require.NoError(t, err)
	require.Equal(t, user.GetId(), found.GetId())
}

func userRepository_Update_Success(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
	id := user.GetId()

	user.SetUsername("updated")
	err = repo.Update(context.Background(), user)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "updated", found.GetUsername())
}

func userRepository_Delete_Success(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
	id := user.GetId()

	err = repo.Delete(context.Background(), id)
	require.NoError(t, err)

	_, err = repo.FindByID(context.Background(), id)
}

func userRepository_BlockUser_Success(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
	id := user.GetId()

	err = repo.BlockUser(context.Background(), id, "spam")
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	require.True(t, found.IsBlocked())
	require.Equal(t, "spam", found.GetBlockReason())
	require.NotNil(t, found.GetBlockedAt())
}

func userRepository_UnblockUser_Success(t *testing.T, repo interfaces.UserRepository) {
	user := generateTestUser()
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
	id := user.GetId()

	err = repo.BlockUser(context.Background(), id, "spam")
	require.NoError(t, err)

	err = repo.UnblockUser(context.Background(), id)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	require.False(t, found.IsBlocked())
	require.Empty(t, found.GetBlockReason())
	require.Nil(t, found.GetBlockedAt())
}
