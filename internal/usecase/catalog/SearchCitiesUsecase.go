package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type SearchCitiesUseCase struct {
	cityRepo interfaces.CityRepository
	logger   *slog.Logger
}

func NewSearchCitiesUseCase(
	cityRepo interfaces.CityRepository,
	logger *slog.Logger,
) *SearchCitiesUseCase {
	return &SearchCitiesUseCase{
		cityRepo: cityRepo,
		logger:   logger,
	}
}

func (uc *SearchCitiesUseCase) Execute(ctx context.Context, req dto.SearchCitiesRequest) ([]dto.CityResponse, error) {
	uc.logger.Info("searching cities", "query", req.Query, "limit", req.Limit, "offset", req.Offset)

	cities, err := uc.cityRepo.SearchCities(ctx, req.Query, req.Limit, req.Offset)
	if err != nil {
		uc.logger.Error("failed to search cities", "error", err)
		return nil, err
	}

	response := make([]dto.CityResponse, 0, len(cities))
	for _, city := range cities {
		coords := city.GetCoordinates()
		coordStr := coords.String()
		response = append(response, dto.CityResponse{
			ID:          city.GetId(),
			Name:        city.GetName(),
			Population:  city.GetPopulation(),
			IsCapital:   city.IsCapital(),
			Coordinates: coordStr,
			Description: city.GetDescription(),
			Timezone:    city.GetTimezone(),
			ImageURL:    city.GetImageId().String(),
		})
	}

	return response, nil
}
