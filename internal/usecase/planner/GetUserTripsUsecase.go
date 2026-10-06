package planner

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
)

type GetUserTripsUseCase struct {
	tripRepo interfaces.TripRepository
	logger   *slog.Logger
}

func NewGetUserTripsUseCase(
	tripRepo interfaces.TripRepository,
	logger *slog.Logger,
) *GetUserTripsUseCase {
	return &GetUserTripsUseCase{
		tripRepo: tripRepo,
		logger:   logger,
	}
}

func (uc *GetUserTripsUseCase) Execute(ctx context.Context, req requests.GetUserTripsRequest) ([]requests.UserTripResponse, error) {
	uc.logger.Info("getting user trips", "user_id", req.UserID)

	trips, err := uc.tripRepo.FindByUser(ctx, req.UserID)
	if err != nil {
		uc.logger.Error("failed to get user trips", "error", err)
		return nil, err
	}

	response := make([]requests.UserTripResponse, 0, len(trips))
	for _, t := range trips {
		response = append(response, requests.UserTripResponse{
			ID:        t.GetId(),
			Title:     t.GetTitle(),
			StartDate: t.GetStartDate().Format("2006-01-02"),
			EndDate:   t.GetEndDate().Format("2006-01-02"),
			Status:    t.GetStatus(),
		})
	}

	uc.logger.Info("user trips retrieved", "count", len(response))
	return response, nil
}
