//go:build redis
// +build redis

package session_repo

import (
	"log"
	"os"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	"github.com/stretchr/testify/require"
)

var testCfg *config.Config

func TestMain(m *testing.M) {
	var err error
	testCfg, err = config.LoadTest()
	if err != nil {
		log.Fatalf("Failed to load test config: %v", err)
	}
	domain.SetConfig(domain.GetConfig())
	code := m.Run()
	os.Exit(code)
}

func TestRedisSessionRepository(t *testing.T) {
	repo, err := NewSessionRepo(testCfg)
	require.NoError(t, err)

	t.Run("SetAndGet", func(t *testing.T) {
		sessionRepository_SetAndGet(t, repo)
	})
	t.Run("GetNotFound", func(t *testing.T) {
		sessionRepository_GetNotFound(t, repo)
	})
	t.Run("Delete", func(t *testing.T) {
		sessionRepository_Delete(t, repo)
	})
	t.Run("Exists", func(t *testing.T) {
		sessionRepository_Exists(t, repo)
	})
}
