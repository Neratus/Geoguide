package analytics

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GenerateReportUseCase struct {
	placeRepo interfaces.PlaceRepository
	userRepo  interfaces.UserRepository
	tripRepo  interfaces.TripRepository
	reportGen interfaces.ReportGenerator
	logger    *slog.Logger
}

func NewGenerateReportUseCase(
	placeRepo interfaces.PlaceRepository,
	userRepo interfaces.UserRepository,
	tripRepo interfaces.TripRepository,
	reportGen interfaces.ReportGenerator,
	logger *slog.Logger,
) *GenerateReportUseCase {
	return &GenerateReportUseCase{
		placeRepo: placeRepo,
		userRepo:  userRepo,
		tripRepo:  tripRepo,
		reportGen: reportGen,
		logger:    logger,
	}
}
func (uc *GenerateReportUseCase) Execute(ctx context.Context, req requests.GenerateReportRequest) ([]byte, error) {
	uc.logger.Info("generating report", "type", req.ReportType, "format", req.Format, "limit", req.Limit)

	switch req.ReportType {
	case "popular_places":
		uc.logger.Info("fetching popular places report")
		places, err := uc.placeRepo.GetPopularPlacesReport(ctx, req.Limit, req.DateFrom, req.DateTo)
		if err != nil {
			uc.logger.Error("GetPopularPlacesReport failed", "error", err)
			return nil, err
		}
		uc.logger.Info("got places", "count", len(places))
		return uc.reportGen.GeneratePopularPlacesReport(places, req.Format, req.Options)

	case "user_activity":
		users, err := uc.userRepo.GetUserActivityReport(ctx, req.Limit, req.DateFrom, req.DateTo)
		if err != nil {
			uc.logger.Error("GetUserActivityReport failed", "error", err)
			return nil, err
		}
		return uc.reportGen.GenerateUserActivityReport(users, req.Format, req.Options)

	case "trip_statistics":
		trips, err := uc.tripRepo.GetTripStatisticsReport(ctx, req.Limit, req.DateFrom, req.DateTo)
		if err != nil {
			uc.logger.Error("GetTripStatisticsReport failed", "error", err)
			return nil, err
		}
		return uc.reportGen.GenerateTripStatisticsReport(trips, req.Format, req.Options)

	default:
		return nil, usecase_errors.ErrInvalidReportType
	}
}
