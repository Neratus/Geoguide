package catalog

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/google/uuid"
)

type GetCountriesUseCase struct {
	countryRepo interfaces.CountryRepository
	cityRepo    interfaces.CityRepository
	staticRepo  interfaces.StaticPageRepository
	logger      *slog.Logger
}

func NewGetCountriesUseCase(
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *GetCountriesUseCase {
	return &GetCountriesUseCase{
		countryRepo: countryRepo,
		cityRepo:    cityRepo,
		staticRepo:  staticRepo,
		logger:      logger,
	}
}

func (uc *GetCountriesUseCase) Execute(ctx context.Context, req requests.GetCountriesRequest) ([]requests.CountryResponse, error) {
	uc.logger.Info("getting countries", "limit", req.Limit, "offset", req.Offset)

	countries, err := uc.countryRepo.FindAll(ctx)
	if err != nil {
		uc.logger.Error("failed to get countries", "error", err)
		return nil, err
	}

	resp := make([]requests.CountryResponse, 0, len(countries))
	for _, c := range countries {
		capitalName := ""
		if c.GetCapitalID() != uuid.Nil {
			capital, _ := uc.cityRepo.FindByID(ctx, c.GetCapitalID())
			if capital != nil {
				capitalName = capital.GetName()
			}
		}
		imageURL := ""
		if c.GetImageID() != uuid.Nil {
			url, err := uc.staticRepo.GetFileURL(ctx, c.GetImageID().String())
			if err == nil {
				imageURL = url
			} else {
				uc.logger.Warn("failed to get image URL", "country_id", c.GetId(), "error", err)
			}
		}
		if req.Search != "" {
			if strings.Contains(strings.ToLower(c.GetName()), strings.ToLower(req.Search)) {
				resp = append(resp, requests.CountryResponse{
					ID:         c.GetId(),
					Name:       c.GetName(),
					Capital:    capitalName,
					Area:       c.GetArea(),
					Population: c.GetPopulation(),
					Currency:   c.GetCurrency(),
					Language:   c.GetLanguage(),
					PhoneCode:  c.GetPhoneCode(),
					ImageURL:   imageURL,
				})

			} else {
				continue
			}
		}
		resp = append(resp, requests.CountryResponse{
			ID:         c.GetId(),
			Name:       c.GetName(),
			Capital:    capitalName,
			Area:       c.GetArea(),
			Population: c.GetPopulation(),
			Currency:   c.GetCurrency(),
			Language:   c.GetLanguage(),
			PhoneCode:  c.GetPhoneCode(),
			ImageURL:   imageURL,
		})
	}
	return resp, nil
}
