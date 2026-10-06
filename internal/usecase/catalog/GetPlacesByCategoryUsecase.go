package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/google/uuid"
)

type GetPlacesByCategoryUseCase struct {
	placeRepo  interfaces.PlaceRepository
	staticRepo interfaces.StaticPageRepository
	logger     *slog.Logger
}

func NewGetPlacesByCategoryUseCase(
	placeRepo interfaces.PlaceRepository,
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *GetPlacesByCategoryUseCase {
	return &GetPlacesByCategoryUseCase{
		placeRepo:  placeRepo,
		staticRepo: staticRepo,
		logger:     logger,
	}
}

func (uc *GetPlacesByCategoryUseCase) Execute(ctx context.Context, req requests.GetPlacesByCategoryRequest) ([]requests.PlaceResponse, error) {
	uc.logger.Info("getting places by category", "category", req.Category)

	places, err := uc.placeRepo.FindByCategory(ctx, req.Category)
	if err != nil {
		uc.logger.Error("failed to get places by category", "error", err)
		return nil, err
	}

	resp := make([]requests.PlaceResponse, 0, len(places))
	for _, p := range places {
		imageURL := ""
		if p.GetImageID() != uuid.Nil {
			url, err := uc.staticRepo.GetFileURL(ctx, p.GetImageID().String())
			if err == nil {
				imageURL = url
			} else {
				uc.logger.Warn("failed to get image URL for place", "place_id", p.GetId(), "error", err)
			}
		}
		coord := p.GetCoordinates()

		resp = append(resp, requests.PlaceResponse{
			ID:                  p.GetId(),
			Name:                p.GetName(),
			Category:            p.GetCategory(),
			Description:         p.GetDescription(),
			Coordinates:         coord.String(),
			Address:             p.GetAddress(),
			OpeningHours:        p.GetOpeningHours(),
			PriceInfo:           p.GetPriceInfo(),
			AvgVisitDurationMin: int(p.GetAvgVisitDurationMin()),
			AvgRating:           p.GetAvgRating(),
			ReviewsCount:        int(p.GetReviewCnt()),
			ContactPhone:        p.GetContactPhone(),
			Website:             p.GetWebsite(),
			CityID:              p.GetCityID(),
			CityName:            p.GetCityName(),
			DistrictID:          p.GetDistrictID(),
			DistrictName:        p.GetDistrictName(),
			ImageURL:            imageURL,
			Reviews:             nil,
		})
	}
	return resp, nil
}
