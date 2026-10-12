package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
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

func (uc *GetHolidaysByCountryUseCase) Execute(ctx context.Context, req dto.GetHolidaysByCountryRequest) ([]dto.HolidaySearchResponse, error) {
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

	response := make([]dto.HolidaySearchResponse, 0, len(holidays))
	for _, h := range holidays {
		response = append(response, dto.HolidaySearchResponse{
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
