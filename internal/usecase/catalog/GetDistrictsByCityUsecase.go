package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
)

type GetDistrictsByCityUseCase struct {
	cityRepo interfaces.CityRepository
	logger   *slog.Logger
}

func NewGetDistrictsByCityUseCase(
	districtRepo interfaces.CityRepository,
	logger *slog.Logger,
) *GetDistrictsByCityUseCase {
	return &GetDistrictsByCityUseCase{
		cityRepo: districtRepo,
		logger:   logger,
	}
}

func (uc *GetDistrictsByCityUseCase) Execute(ctx context.Context, req requests.GetDistrictsByCityRequest) ([]requests.DistrictResponse, error) {
	uc.logger.Info("getting districts by city", "city_id", req.CityID)
	districts, err := uc.cityRepo.FindDistrictsByCity(ctx, req.CityID)
	if err != nil {
		return nil, err
	}
	start := req.Offset
	if start > len(districts) {
		start = len(districts)
	}
	end := start + req.Limit
	if end > len(districts) {
		end = len(districts)
	}
	result := make([]requests.DistrictResponse, 0, end-start)
	for _, d := range districts[start:end] {
		result = append(result, requests.DistrictResponse{
			ID:          d.GetId(),
			Name:        d.GetName(),
			Description: d.GetDescription(),
			Coordinates: d.GetCoordinates(),
			CityID:      d.GetCityId(),
			ImageID:     d.GetImageId(),
		})
	}
	return result, nil
}
