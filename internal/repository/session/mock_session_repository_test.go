package session_repo

import "testing"

func TestMockSessionRepository(t *testing.T) {
	t.Run("SetAndGet", func(t *testing.T) {
		repo, _ := NewMockSessionRepo()
		sessionRepository_SetAndGet(t, repo)
	})
	t.Run("GetNotFound", func(t *testing.T) {
		repo, _ := NewMockSessionRepo()
		sessionRepository_GetNotFound(t, repo)
	})
	t.Run("Delete", func(t *testing.T) {
		repo, _ := NewMockSessionRepo()
		sessionRepository_Delete(t, repo)
	})
	t.Run("Exists", func(t *testing.T) {
		repo, _ := NewMockSessionRepo()
		sessionRepository_Exists(t, repo)
	})
}
