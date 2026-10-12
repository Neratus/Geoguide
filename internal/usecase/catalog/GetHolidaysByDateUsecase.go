package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type GetHolidaysByDateUseCase struct {
	countryRepo interfaces.CountryRepository
	logger      *slog.Logger
}

func NewGetHolidaysByDateUseCase(countryRepo interfaces.CountryRepository, logger *slog.Logger) *GetHolidaysByDateUseCase {
	return &GetHolidaysByDateUseCase{countryRepo: countryRepo, logger: logger}
}

func (uc *GetHolidaysByDateUseCase) Execute(ctx context.Context, req dto.GetHolidaysByDateRequest) ([]dto.HolidaySearchDateResponse, error) {
	uc.logger.Info("getting holidays by date", "date", req.Date)
	holidays, err := uc.countryRepo.FindHolidaysByDate(ctx, req.Date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	result := make([]dto.HolidaySearchDateResponse, len(holidays))
	for i, h := range holidays {
		result[i] = dto.HolidaySearchDateResponse{
			ID:          h.GetId(),
			Name:        h.GetName(),
			Date:        h.GetDate(),
			Description: h.GetDescription(),
			Traditions:  h.GetTraditions(),
			History:     h.GetHistory(),
			IsNational:  h.IsNational(),
			CountryID:   h.GetCountryID(),
			ImageID:     h.GetImageID(),
		}
	}
	return result, nil
}
