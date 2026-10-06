package review

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/mocks"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func createTestPlace(id domain.PlaceID, cityID domain.CityID) *domain.Place {
	coords := domain.NewCoordinates(48.8606, 2.3376)
	place, err := domain.NewPlace(
		id, "Лувр", "Музей", "Описание", coords,
		"Париж", "09:00-18:00", "15 EUR", 120,
		"+33123456789", "https://louvre.fr",
		cityID, domain.CityDistrictID(uuid.New()), domain.ImageID(uuid.New()),
		"Париж", "1-й округ",
	)
	if err != nil {
		panic(err)
	}
	return place
}

func createTestReview(id domain.ReviewID, rating int, comment string, visitDate time.Time, userID domain.UserID, placeID domain.PlaceID) *domain.Review {
	review, err := domain.NewReview(
		id, rating, comment, visitDate, userID, placeID, domain.ImageID(uuid.New()),
	)
	if err != nil {
		panic(err)
	}
	return review
}

func TestCreateReviewUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	placeRepo := mocks.NewMockPlaceRepository()
	notifier := mocks.NewMockNotifier()
	taskQueue := mocks.NewMockTaskQueue()

	uc := NewCreateReviewUseCase(placeRepo, notifier, taskQueue, logger)

	placeID := domain.PlaceID(uuid.New())
	userID := domain.UserID(uuid.New())
	cityID := domain.CityID(uuid.New())

	place := createTestPlace(placeID, cityID)
	placeRepo.Places[placeID] = place

	t.Run("Success_CreateNewReview", func(t *testing.T) {
		req := requests.CreateReviewRequest{
			UserID:    userID,
			PlaceID:   placeID,
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

		require.Len(t, placeRepo.Reviews, 1, "Отзыв должен быть сохранен в репозитории")
		require.Len(t, taskQueue.PublishedTasks, 1, "Задача должна быть отправлена в очередь")
	})

	t.Run("Error_PlaceNotFound", func(t *testing.T) {
		req := requests.CreateReviewRequest{
			UserID:    userID,
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
		existingReview := createTestReview(domain.ReviewID(uuid.New()), 4, "Норм", time.Now(), userID, placeID)
		placeRepo.Reviews[existingReview.GetId()] = existingReview

		req := requests.CreateReviewRequest{
			UserID:    userID,
			PlaceID:   placeID,
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
func TestDeleteReviewUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	placeRepo := mocks.NewMockPlaceRepository()

	uc := NewDeleteReviewUseCase(placeRepo, logger)

	placeID := domain.PlaceID(uuid.New())
	authorID := domain.UserID(uuid.New())
	randomUserID := domain.UserID(uuid.New())
	cityID := domain.CityID(uuid.New())
	reviewID := domain.ReviewID(uuid.New())

	place := createTestPlace(placeID, cityID)
	placeRepo.Places[placeID] = place

	review := createTestReview(reviewID, 5, "Супер", time.Now(), authorID, placeID)
	placeRepo.Reviews[reviewID] = review

	t.Run("Success_DeleteByAuthor", func(t *testing.T) {
		req := requests.DeleteReviewRequest{
			ReviewID:    reviewID,
			UserID:      authorID,
			IsModerator: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)

		_, err = placeRepo.FindReviewByID(ctx, reviewID)
		require.Error(t, err, "Отзыв должен быть удален из репозитория")
	})

	t.Run("Success_DeleteByModerator", func(t *testing.T) {
		placeRepo.Reviews[reviewID] = review

		req := requests.DeleteReviewRequest{
			ReviewID:    reviewID,
			UserID:      randomUserID,
			IsModerator: true,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Success)
	})

	t.Run("Error_ReviewNotFound", func(t *testing.T) {
		req := requests.DeleteReviewRequest{
			ReviewID:    domain.ReviewID(uuid.New()),
			UserID:      authorID,
			IsModerator: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrReviewNotFound)
		require.Nil(t, resp)
	})

	t.Run("Error_NotAuthorized", func(t *testing.T) {
		placeRepo.Reviews[reviewID] = review

		req := requests.DeleteReviewRequest{
			ReviewID:    reviewID,
			UserID:      randomUserID,
			IsModerator: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrNotAuthorized)
		require.Nil(t, resp)
	})
}

func TestGetReviewsByPlaceUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	placeRepo := mocks.NewMockPlaceRepository()

	uc := NewGetReviewsByPlaceUseCase(placeRepo, logger)

	placeID := domain.PlaceID(uuid.New())
	userID := domain.UserID(uuid.New())
	cityID := domain.CityID(uuid.New())

	place := createTestPlace(placeID, cityID)
	_ = placeRepo.Save(ctx, place)

	review1, err := domain.NewReview(
		domain.ReviewID(uuid.New()),
		5, "Отлично", time.Now(), userID, placeID, domain.ImageID(uuid.New()),
	)
	require.NoError(t, err)
	review1.SetApproved(true)
	_ = placeRepo.SaveReview(ctx, review1)

	review2, err := domain.NewReview(
		domain.ReviewID(uuid.New()),
		2, "Плохо", time.Now(), userID, placeID, domain.ImageID(uuid.New()),
	)
	require.NoError(t, err)
	_ = placeRepo.SaveReview(ctx, review2)

	t.Run("Success_GetAllReviews", func(t *testing.T) {
		req := requests.GetReviewsByPlaceRequest{
			PlaceID:      placeID,
			OnlyApproved: false,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 2, "Должны вернуться оба отзыва")
	})

	t.Run("Success_GetOnlyApprovedReviews", func(t *testing.T) {
		req := requests.GetReviewsByPlaceRequest{
			PlaceID:      placeID,
			OnlyApproved: true,
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 1, "Должен вернуться только одобренный отзыв")
		require.Equal(t, "Отлично", resp[0].Comment)
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

func TestModerateReviewUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	placeRepo := mocks.NewMockPlaceRepository()
	userRepo := mocks.NewMockUserRepository()
	notifier := mocks.NewMockNotifier()
	taskQueue := mocks.NewMockTaskQueue()

	uc := NewModerateReviewUseCase(placeRepo, userRepo, notifier, taskQueue, logger)

	placeID := domain.PlaceID(uuid.New())
	userID := domain.UserID(uuid.New())
	cityID := domain.CityID(uuid.New())
	reviewID := domain.ReviewID(uuid.New())

	place := createTestPlace(placeID, cityID)
	placeRepo.Places[placeID] = place

	user := &domain.User{}
	userRepo.Users[userID] = user

	review := createTestReview(reviewID, 4, "Хорошо", time.Now(), userID, placeID)
	placeRepo.Reviews[reviewID] = review

	t.Run("Success_ModerateAndApprove", func(t *testing.T) {
		req := requests.ModerateReviewRequest{
			ReviewID: reviewID,
			Approved: true,
			Comment:  "Все отлично",
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsModerated)
		require.True(t, resp.IsApproved)
		require.Equal(t, "Все отлично", resp.ModerationComment)

		require.Len(t, taskQueue.PublishedTasks, 1, "Задача о модерации должна быть опубликована")
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
		alreadyModeratedReview := createTestReview(domain.ReviewID(uuid.New()), 3, "Средне", time.Now(), userID, placeID)
		alreadyModeratedReview.SetModerated(true)
		placeRepo.Reviews[alreadyModeratedReview.GetId()] = alreadyModeratedReview

		req := requests.ModerateReviewRequest{
			ReviewID: alreadyModeratedReview.GetId(),
			Approved: false,
			Comment:  "Повторная модерация",
		}

		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrReviewAlreadyModerated)
		require.Nil(t, resp)
	})
}
