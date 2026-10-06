package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
)

type GetCitiesByCountryUseCase struct {
	cityRepo   interfaces.CityRepository
	staticRepo interfaces.StaticPageRepository
	logger     *slog.Logger
}

func NewGetCitiesByCountryUseCase(
	cityRepo interfaces.CityRepository,
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *GetCitiesByCountryUseCase {
	return &GetCitiesByCountryUseCase{
		cityRepo:   cityRepo,
		staticRepo: staticRepo,
		logger:     logger,
	}
}

func (uc *GetCitiesByCountryUseCase) Execute(ctx context.Context, req requests.GetCitiesByCountryRequest) ([]requests.CityResponse, error) {
	uc.logger.Info("getting cities by country", "country_id", req.CountryID, "limit", req.Limit, "offset", req.Offset)

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	cities, err := uc.cityRepo.FindByCountry(ctx, req.CountryID)
	if err != nil {
		uc.logger.Error("failed to get cities", "error", err)
		return nil, err
	}

	if len(cities) == 0 {
		return nil, usecase_errors.ErrCitiesNotFound
	}

	response := make([]requests.CityResponse, 0, len(cities))
	for _, city := range cities {
		coords := city.GetCoordinates()
		coordStr := coords.String()
		imageURL := ""
		if city.GetImageId() != uuid.Nil {
			url, err := uc.staticRepo.GetFileURL(ctx, city.GetImageId().String())
			if err == nil {
				imageURL = url
			}
		}
		response = append(response, requests.CityResponse{
			ID:          city.GetId(),
			Name:        city.GetName(),
			Population:  city.GetPopulation(),
			IsCapital:   city.IsCapital(),
			Coordinates: coordStr,
			Description: city.GetDescription(),
			Timezone:    city.GetTimezone(),
			ImageURL:    imageURL,
		})
	}

	uc.logger.Info("cities retrieved", "count", len(response))
	return response, nil
}
