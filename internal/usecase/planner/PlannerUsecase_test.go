package planner

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/mocks"
	"github.com/Neratus/geoguide/internal/mocks/builders"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
func TestCreateTripUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	mockRepo := mocks.NewMockTripRepository()
	uc := NewCreateTripUseCase(mockRepo, logger)

	t.Run("Success_CreateTrip", func(t *testing.T) {
		userID := domain.UserID(uuid.New())
		req := requests.CreateTripRequest{
			UserID:    userID,
			Title:     "Поездка в Париж",
			StartDate: time.Now().Add(24 * time.Hour),
			EndDate:   time.Now().Add(48 * time.Hour),
			Budget:    5000.0,
			Notes:     "Отдых",
		}

		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "Поездка в Париж", resp.Title)
		require.Equal(t, "DRAFT", resp.Status)

		require.Len(t, mockRepo.Trips, 1)
		savedTrip := mockRepo.Trips[resp.ID]
		require.NotNil(t, savedTrip)
		require.Equal(t, userID, savedTrip.GetUserID())
	})
}

func TestGetTripUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	mockRepo := mocks.NewMockTripRepository()
	uc := NewGetTripUseCase(mockRepo, logger)

	t.Run("Success_GetTrip", func(t *testing.T) {
		userID := domain.UserID(uuid.New())
		tripID := domain.TripID(uuid.New())

		trip, _ := domain.NewTrip(tripID, "Тест", time.Now(), time.Now().Add(24*time.Hour), 1000, "DRAFT", "", userID, domain.ImageID{})
		mockRepo.Trips[tripID] = trip
		mockRepo.Places[tripID] = []*domain.TripPlace{}

		req := requests.GetTripRequest{TripID: tripID, UserID: userID}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "Тест", resp.Title)
	})

	t.Run("Error_TripNotFound", func(t *testing.T) {
		req := requests.GetTripRequest{TripID: domain.TripID(uuid.New()), UserID: domain.UserID(uuid.New())}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrTripNotFound)
		require.Nil(t, resp)
	})

	t.Run("Error_Unauthorized", func(t *testing.T) {
		ownerID := domain.UserID(uuid.New())
		requesterID := domain.UserID(uuid.New())
		tripID := domain.TripID(uuid.New())

		trip, _ := domain.NewTrip(tripID, "Чужая поездка", time.Now(), time.Now().Add(24*time.Hour), 1000, "DRAFT", "", ownerID, domain.ImageID{})
		mockRepo.Trips[tripID] = trip

		req := requests.GetTripRequest{TripID: tripID, UserID: requesterID}
		resp, err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrTripNotFound)
		require.Nil(t, resp)
	})
}

func TestGetUserTripsUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	mockRepo := mocks.NewMockTripRepository()
	uc := NewGetUserTripsUseCase(mockRepo, logger)

	t.Run("Success_GetUserTrips", func(t *testing.T) {
		userID := domain.UserID(uuid.New())

		trip1, _ := domain.NewTrip(domain.TripID(uuid.New()), "Trip 1", time.Now(), time.Now().Add(24*time.Hour), 1000, "DRAFT", "", userID, domain.ImageID{})
		trip2, _ := domain.NewTrip(domain.TripID(uuid.New()), "Trip 2", time.Now(), time.Now().Add(48*time.Hour), 2000, "DRAFT", "", userID, domain.ImageID{})

		mockRepo.Trips[trip1.GetId()] = trip1
		mockRepo.Trips[trip2.GetId()] = trip2

		req := requests.GetUserTripsRequest{UserID: userID}
		resp, err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp, 2)
	})
}

func TestAddPlaceToTripUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	mockTripRepo := mocks.NewMockTripRepository()
	mockPlaceRepo := mocks.NewMockPlaceRepository()
	uc := NewAddPlaceToTripUseCase(mockTripRepo, mockPlaceRepo, logger)

	t.Run("Success_AddPlace", func(t *testing.T) {
		userID := domain.UserID(uuid.New())
		tripID := domain.TripID(uuid.New())
		placeID := domain.PlaceID(uuid.New())

		trip, _ := domain.NewTrip(tripID, "Тест", time.Now(), time.Now().Add(24*time.Hour), 1000, "DRAFT", "", userID, domain.ImageID{})
		mockTripRepo.Trips[tripID] = trip

		place := builders.NewPlaceBuilder().Build()
		mockPlaceRepo.Places[placeID] = place

		req := requests.AddPlaceToTripRequest{
			TripID:      tripID,
			PlaceID:     placeID,
			UserID:      userID,
			DayNumber:   1,
			DurationMin: 120,
		}

		err := uc.Execute(ctx, req)

		require.NoError(t, err)
		require.Len(t, mockTripRepo.Places[tripID], 1)
		require.Equal(t, placeID, mockTripRepo.Places[tripID][0].GetPlaceID())
	})

	t.Run("Error_PlaceNotFound", func(t *testing.T) {
		req := requests.AddPlaceToTripRequest{
			TripID:  domain.TripID(uuid.New()),
			PlaceID: domain.PlaceID(uuid.New()),
			UserID:  domain.UserID(uuid.New()),
		}

		err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrPlaceNotFound)
	})

	t.Run("Error_TripNotFound", func(t *testing.T) {
		placeID := domain.PlaceID(uuid.New())
		mockPlaceRepo.Places[placeID] = builders.NewPlaceBuilder().Build()

		req := requests.AddPlaceToTripRequest{
			TripID:  domain.TripID(uuid.New()),
			PlaceID: placeID,
			UserID:  domain.UserID(uuid.New()),
		}

		err := uc.Execute(ctx, req)

		require.Error(t, err)
		require.ErrorIs(t, err, usecase_errors.ErrTripNotFound)
	})
}

func TestDeleteTripUseCase(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()
	mockRepo := mocks.NewMockTripRepository()
	uc := NewDeleteTripUseCase(mockRepo, logger)

	t.Run("Success_DeleteTrip", func(t *testing.T) {
		tripID := domain.TripID(uuid.New())
		trip, _ := domain.NewTrip(tripID, "Тест", time.Now(), time.Now().Add(24*time.Hour), 1000, "DRAFT", "", domain.UserID(uuid.New()), domain.ImageID{})
		mockRepo.Trips[tripID] = trip

		req := requests.DeleteTripRequest{TripID: tripID}
		err := uc.Execute(ctx, req)

		require.NoError(t, err)
		_, exists := mockRepo.Trips[tripID]
		require.False(t, exists, "Поездка должна быть удалена из репозитория")
	})

	t.Run("Error_TripNotFound", func(t *testing.T) {
		req := requests.DeleteTripRequest{TripID: domain.TripID(uuid.New())}
		err := uc.Execute(ctx, req)

		require.Error(t, err)
	})
}
