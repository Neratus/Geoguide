package session_repo

import (
	"context"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/stretchr/testify/require"
)

func sessionRepository_SetAndGet(t *testing.T, repo interfaces.SessionRepository) {
	key := "testKey"
	value := "testValue"
	err := repo.Set(context.Background(), key, value, time.Minute)
	require.NoError(t, err)

	var got string
	err = repo.Get(context.Background(), key, &got)
	require.NoError(t, err)
	require.Equal(t, value, got)
}

func sessionRepository_GetNotFound(t *testing.T, repo interfaces.SessionRepository) {
	var got string
	err := repo.Get(context.Background(), "nonexistent", &got)
	require.Error(t, err)
}

func sessionRepository_Delete(t *testing.T, repo interfaces.SessionRepository) {
	key := "testKey"
	value := "testValue"
	err := repo.Set(context.Background(), key, value, time.Minute)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), key)
	require.NoError(t, err)

	var got string
	err = repo.Get(context.Background(), key, &got)
	require.Error(t, err)
}

func sessionRepository_Exists(t *testing.T, repo interfaces.SessionRepository) {
	key := "testKey"
	value := "testValue"
	err := repo.Set(context.Background(), key, value, time.Minute)
	require.NoError(t, err)

	ok := repo.Exists(context.Background(), key)
	require.True(t, ok)

	ok = repo.Exists(context.Background(), "nonexistent")
	require.False(t, ok)
}
