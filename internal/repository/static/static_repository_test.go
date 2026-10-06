package static_repo

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func generateTestStaticPage(slug, title string) *domain.StaticPage {
	page, err := domain.NewStaticPage(
		uuid.Nil,
		slug,
		title,
		"content",
		"meta desc",
		"http://example.com/image.jpg",
		"image alt",
	)
	if err != nil {
		panic(err)
	}
	return page
}

func staticRepository_SaveNew(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	page := generateTestStaticPage("test-page", "Test Page")
	file := strings.NewReader("fake image data")
	err := repo.Save(ctx, page, file)
	require.NoError(t, err)
	require.NotZero(t, page.GetId())

	found, err := repo.FindByID(ctx, page.GetId())
	require.NoError(t, err)
	require.Equal(t, page.GetId(), found.GetId())
	require.Equal(t, "test-page", found.GetSlug())

	rc, err := repo.GetFile(ctx, page.GetId())
	require.NoError(t, err)
	defer rc.Close()
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "fake image data", string(data))
}

func staticRepository_SaveExisting(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	page := generateTestStaticPage("test-page", "Test Page")
	err := repo.Save(ctx, page, strings.NewReader("data"))
	require.NoError(t, err)
	id := page.GetId()

	page.SetTitle("Updated Title")
	err = repo.Save(ctx, page, strings.NewReader("new data"))
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Updated Title", found.GetTitle())

	rc, err := repo.GetFile(ctx, id)
	require.NoError(t, err)
	defer rc.Close()
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "new data", string(data))
}

func staticRepository_FindByID_NotFound(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	_, err := repo.FindByID(ctx, uuid.New())
	require.Error(t, err)
}

func staticRepository_FindByID_Found(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	page := generateTestStaticPage("test-page", "Test Page")
	err := repo.Save(ctx, page, strings.NewReader("data"))
	require.NoError(t, err)
	id := page.GetId()

	found, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, page.GetId(), found.GetId())
}

func staticRepository_Update_NotFound(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	page := generateTestStaticPage("test-page", "Test Page")
	page.SetId(uuid.New())
	err := repo.Update(ctx, page, strings.NewReader("data"))
	require.Error(t, err)
}

func staticRepository_Update_Success(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	page := generateTestStaticPage("test-page", "Test Page")
	err := repo.Save(ctx, page, strings.NewReader("old"))
	require.NoError(t, err)
	id := page.GetId()

	page.SetTitle("New Title")
	err = repo.Update(ctx, page, strings.NewReader("new"))
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "New Title", found.GetTitle())

	rc, err := repo.GetFile(ctx, id)
	require.NoError(t, err)
	defer rc.Close()
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "new", string(data))
}

func staticRepository_Delete_NotFound(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
}

func staticRepository_Delete_Success(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	page := generateTestStaticPage("test-page", "Test Page")
	err := repo.Save(ctx, page, strings.NewReader("data"))
	require.NoError(t, err)
	id := page.GetId()

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, id)
	require.Error(t, err)

	_, err = repo.GetFile(ctx, id)
	require.Error(t, err)
}

func staticRepository_GetFile_Success(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	page := generateTestStaticPage("test-page", "Test Page")
	fileData := "test file content"
	err := repo.Save(ctx, page, strings.NewReader(fileData))
	require.NoError(t, err)
	id := page.GetId()

	rc, err := repo.GetFile(ctx, id)
	require.NoError(t, err)
	defer rc.Close()
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, fileData, string(data))
}

func staticRepository_GetFile_NotFound(t *testing.T, repo interfaces.StaticPageRepository) {
	ctx := context.Background()
	_, err := repo.GetFile(ctx, uuid.New())
	require.Error(t, err)
}
