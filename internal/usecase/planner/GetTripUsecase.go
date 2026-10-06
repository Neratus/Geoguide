package planner

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetTripUseCase struct {
	tripRepo interfaces.TripRepository
	logger   *slog.Logger
}

func NewGetTripUseCase(
	tripRepo interfaces.TripRepository,
	logger *slog.Logger,
) *GetTripUseCase {
	return &GetTripUseCase{
		tripRepo: tripRepo,
		logger:   logger,
	}
}

func (uc *GetTripUseCase) Execute(ctx context.Context, req requests.GetTripRequest) (*requests.GetTripResponse, error) {
	uc.logger.Info("getting trip", "trip_id", req.TripID)

	trip, err := uc.tripRepo.FindByID(ctx, req.TripID)
	if err != nil || trip == nil || trip.GetUserID() != req.UserID {
		uc.logger.Warn("trip not found", "trip_id", req.TripID)
		return nil, usecase_errors.ErrTripNotFound
	}

	places, err := uc.tripRepo.GetPlaces(ctx, req.TripID)
	if err != nil {
		uc.logger.Warn("failed to get trip places", "error", err)
		places = []*domain.TripPlace{}
	}

	resp := &requests.GetTripResponse{
		ID:        trip.GetId(),
		Title:     trip.GetTitle(),
		StartDate: trip.GetStartDate().Format("2006-01-02"),
		EndDate:   trip.GetEndDate().Format("2006-01-02"),
		Budget:    trip.GetBudget(),
		Status:    trip.GetStatus(),
		Notes:     trip.GetNotes(),
		Places:    make([]requests.TripPlaceResponse, 0, len(places)),
	}

	for _, tp := range places {
		arrivalTime := ""
		if tp.GetArrivalTime() != nil {
			arrivalTime = tp.GetArrivalTime().Format("15:04")
		}
		resp.Places = append(resp.Places, requests.TripPlaceResponse{
			PlaceID:     tp.GetPlaceID(),
			DayNumber:   tp.GetDayNumber(),
			ArrivalTime: arrivalTime,
			DurationMin: tp.GetDurationMin(),
			Notes:       tp.GetNotes(),
			VisitStatus: tp.GetVisitStatus(),
			ActualCost:  tp.GetActualCost(),
		})
	}

	uc.logger.Info("trip retrieved", "trip_id", req.TripID)
	return resp, nil
}
