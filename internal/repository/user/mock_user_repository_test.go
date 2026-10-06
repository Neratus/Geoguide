package user_repo

import (
	"testing"
)

func TestMockUserRepository(t *testing.T) {
	t.Run("SaveNew", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_SaveNew(t, repo)
	})
	t.Run("SaveExisting", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_SaveExisting(t, repo)
	})
	t.Run("FindByID_NotFound", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_FindByID_NotFound(t, repo)
	})
	t.Run("FindByID_Found", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_FindByID_Found(t, repo)
	})
	t.Run("FindByUsername_NotFound", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_FindByUsername_NotFound(t, repo)
	})
	t.Run("FindByUsername_Found", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_FindByUsername_Found(t, repo)
	})
	t.Run("FindByEmail_NotFound", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_FindByEmail_NotFound(t, repo)
	})
	t.Run("FindByEmail_Found", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_FindByEmail_Found(t, repo)
	})
	t.Run("Update_Success", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_Update_Success(t, repo)
	})
	t.Run("Delete_Success", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_Delete_Success(t, repo)
	})
	t.Run("BlockUser_Success", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_BlockUser_Success(t, repo)
	})
	t.Run("UnblockUser_Success", func(t *testing.T) {
		repo, _ := NewMockUserRepo()
		userRepository_UnblockUser_Success(t, repo)
	})
}
