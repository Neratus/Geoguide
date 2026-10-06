package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
)

type GetHolidaysByCountryUseCase struct {
	countryRepo interfaces.CountryRepository
	logger      *slog.Logger
}

func NewGetHolidaysByCountryUseCase(
	countryRepo interfaces.CountryRepository,
	logger *slog.Logger,
) *GetHolidaysByCountryUseCase {
	return &GetHolidaysByCountryUseCase{
		countryRepo: countryRepo,
		logger:      logger,
	}
}

func (uc *GetHolidaysByCountryUseCase) Execute(ctx context.Context, req requests.GetHolidaysByCountryRequest) ([]requests.HolidaySearchResponse, error) {
	uc.logger.Info("getting holidays by country", "country_id", req.CountryID)

	_, err := uc.countryRepo.FindByID(ctx, req.CountryID)
	if err != nil {
		uc.logger.Warn("country not found", "country_id", req.CountryID)
		return nil, usecase_errors.ErrCountryNotFound
	}

	holidays, err := uc.countryRepo.FindHolidaysByCountry(ctx, req.CountryID)
	if err != nil {
		uc.logger.Error("failed to get holidays", "error", err)
		return nil, err
	}

	response := make([]requests.HolidaySearchResponse, 0, len(holidays))
	for _, h := range holidays {
		response = append(response, requests.HolidaySearchResponse{
			ID:          h.GetId(),
			Name:        h.GetName(),
			Date:        h.GetDate(),
			Description: h.GetDescription(),
			Traditions:  h.GetTraditions(),
			History:     h.GetHistory(),
			IsNational:  h.IsNational(),
		})
	}

	uc.logger.Info("holidays retrieved", "count", len(response))
	return response, nil
}
