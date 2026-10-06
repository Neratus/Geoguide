package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
)

type GetPlaceUseCase struct {
	placeRepo  interfaces.PlaceRepository
	userRepo   interfaces.UserRepository
	staticRepo interfaces.StaticPageRepository
	logger     *slog.Logger
}

func NewGetPlaceUseCase(
	placeRepo interfaces.PlaceRepository,
	userRepo interfaces.UserRepository,
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *GetPlaceUseCase {
	return &GetPlaceUseCase{
		placeRepo:  placeRepo,
		userRepo:   userRepo,
		staticRepo: staticRepo,
		logger:     logger,
	}
}

func (uc *GetPlaceUseCase) Execute(ctx context.Context, req requests.GetPlaceRequest) (*requests.PlaceResponse, error) {
	uc.logger.Info("getting place", "place_id", req.ID)

	place, err := uc.placeRepo.FindByID(ctx, req.ID)
	if err != nil || place == nil {
		uc.logger.Warn("place not found", "place_id", req.ID)
		return nil, usecase_errors.ErrPlaceNotFound
	}

	reviews, err := uc.placeRepo.FindReviewsByPlace(ctx, req.ID)
	if err != nil {
		uc.logger.Warn("failed to get reviews", "error", err)
		reviews = []*domain.Review{}
	}
	imageURL := ""
	if place.GetImageID() != uuid.Nil {
		url, err := uc.staticRepo.GetFileURL(ctx, place.GetImageID().String())
		if err == nil {
			imageURL = url
		}
	}
	place.SetImageURL(imageURL)

	reviewResponses := make([]requests.ReviewResponse, 0, len(reviews))
	for _, r := range reviews {
		username := ""
		user, err := uc.userRepo.FindByID(ctx, r.GetUserID())
		if err == nil && user != nil {
			username = user.GetUsername()
		}

		reviewResponses = append(reviewResponses, requests.ReviewResponse{
			ID:         r.GetId(),
			Rating:     r.GetRating(),
			Comment:    r.GetComment(),
			VisitDate:  r.GetVisitDate().Format("2006-01-02"),
			CreatedAt:  r.GetCreatedAt().Format("2006-01-02 15:04:05"),
			Username:   username,
			IsApproved: r.IsApproved(),
		})
	}

	uc.logger.Info("place retrieved", "place_id", req.ID)

	coords := place.GetCoordinates()
	coordStr := coords.String()

	return &requests.PlaceResponse{
		ID:                  place.GetId(),
		Name:                place.GetName(),
		Category:            place.GetCategory(),
		Description:         place.GetDescription(),
		Coordinates:         coordStr,
		Address:             place.GetAddress(),
		OpeningHours:        place.GetOpeningHours(),
		PriceInfo:           place.GetPriceInfo(),
		AvgVisitDurationMin: int(place.GetAvgVisitDurationMin()),
		AvgRating:           place.GetAvgRating(),
		ReviewsCount:        int(place.GetReviewCnt()),
		ContactPhone:        place.GetContactPhone(),
		Website:             place.GetWebsite(),
		CityID:              place.GetCityID(),
		CityName:            "",
		DistrictID:          place.GetDistrictID(),
		DistrictName:        "",
		ImageURL:            place.GetImageID().String(),
		Reviews:             reviewResponses,
	}, nil
}
