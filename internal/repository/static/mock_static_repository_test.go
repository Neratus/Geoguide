package static_repo

import (
	"testing"
)

func TestMockStaticPageRepository(t *testing.T) {
	t.Run("SaveNew", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_SaveNew(t, repo)
	})
	t.Run("SaveExisting", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_SaveExisting(t, repo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_FindByID_NotFound(t, repo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_FindByID_Found(t, repo)
	})
	t.Run("Update_NotFound", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_Update_NotFound(t, repo)
	})
	t.Run("Update_Success", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_Update_Success(t, repo)
	})
	t.Run("Delete_NotFound", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_Delete_NotFound(t, repo)
	})
	t.Run("Delete_Success", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_Delete_Success(t, repo)
	})
	t.Run("GetFile_Success", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_GetFile_Success(t, repo)
	})
	t.Run("GetFile_NotFound", func(t *testing.T) {
		repo, _ := NewMockStaticPageRepo()
		staticRepository_GetFile_NotFound(t, repo)
	})
}
