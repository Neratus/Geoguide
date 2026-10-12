package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
)

type GetCityUseCase struct {
	cityRepo   interfaces.CityRepository
	staticRepo interfaces.StaticPageRepository
	logger     *slog.Logger
}

func NewGetCityUseCase(
	cityRepo interfaces.CityRepository,
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *GetCityUseCase {
	return &GetCityUseCase{
		cityRepo:   cityRepo,
		staticRepo: staticRepo,
		logger:     logger,
	}
}

func (uc *GetCityUseCase) Execute(ctx context.Context, req dto.GetCityRequest) (*dto.GetCityResponse, error) {
	uc.logger.Info("getting city", "city_id", req.ID)

	city, err := uc.cityRepo.FindByID(ctx, req.ID)
	if err != nil || city == nil {
		uc.logger.Warn("city not found", "city_id", req.ID)
		return nil, usecase_errors.ErrCityNotFound
	}

	imageURL := ""
	if city.GetImageId() != uuid.Nil {
		url, err := uc.staticRepo.GetFileURL(ctx, city.GetImageId().String())
		if err == nil {
			imageURL = url
		} else {
			uc.logger.Warn("failed to get image URL", "city_id", city.GetId(), "error", err)
		}
	}
	uc.logger.Info("city image_id", "id", city.GetImageId())
	if city.GetImageId() != uuid.Nil {
		url, err := uc.staticRepo.GetFileURL(ctx, city.GetImageId().String())
		uc.logger.Info("generated image URL", "url", url, "error", err)
		if err == nil {
			imageURL = url
		}
	}

	return &dto.GetCityResponse{
		ID:          city.GetId(),
		Name:        city.GetName(),
		Population:  city.GetPopulation(),
		IsCapital:   city.IsCapital(),
		Coordinates: city.GetCoordinates(),
		Description: city.GetDescription(),
		Timezone:    city.GetTimezone(),
		TravelTips:  city.GetTravelTips(),
		CountryID:   city.GetCountryId(),
		ImageURL:    imageURL,
		ImageID:     city.GetImageId(),
	}, nil
}
