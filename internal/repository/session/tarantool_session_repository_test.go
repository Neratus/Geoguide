//go:build tarantool
// +build tarantool

package session_repo

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

var tarantoolCfg *config.Config

func TestMain(m *testing.M) {
	v := viper.New()
	v.SetConfigFile("../../../config/config.cassandra.yaml")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	if cfg.Database.CacheType != "tarantool" {
		log.Println("Skipping Tarantool tests: cache_type != tarantool")
		os.Exit(0)
	}
	tarantoolCfg = &cfg

	repo, err := NewTarantoolSessionRepo(tarantoolCfg)
	if err != nil {
		log.Fatalf("Failed to create Tarantool session repo: %v", err)
	}
	defer repo.Close()

	if err := ensureTarantoolSchema(repo); err != nil {
		log.Fatalf("Failed to ensure Tarantool schema: %v", err)
	}

	domain.SetConfig(domain.GetConfig())
	code := m.Run()
	os.Exit(code)
}

func ensureTarantoolSchema(repo *TarantoolSessionRepo) error {
	_, err := repo.conn.Eval(`
        if box.space['sessions'] == nil then
            box.schema.space.create('sessions', { if_not_exists = true })
            box.space.sessions:format({
                {name = 'key', type = 'string'},
                {name = 'value', type = 'string'},
                {name = 'expires_at', type = 'unsigned'}
            })
            box.space.sessions:create_index('primary', { type = 'hash', parts = {'key'}, if_not_exists = true })
        end
    `, []interface{}{})
	if err != nil {
		return fmt.Errorf("failed to create sessions space: %w", err)
	}
	return nil
}
func TestTarantoolSessionRepository(t *testing.T) {
	repo, err := NewTarantoolSessionRepo(tarantoolCfg)
	require.NoError(t, err)
	defer repo.Close()

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
