package catalog

import (
	"context"
	"log/slog"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/google/uuid"
)

type GetCountryUseCase struct {
	countryRepo interfaces.CountryRepository
	cityRepo    interfaces.CityRepository
	staticRepo  interfaces.StaticPageRepository
	logger      *slog.Logger
}

func NewGetCountryUseCase(
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	staticRepo interfaces.StaticPageRepository,
	logger *slog.Logger,
) *GetCountryUseCase {
	return &GetCountryUseCase{
		countryRepo: countryRepo,
		cityRepo:    cityRepo,
		staticRepo:  staticRepo,
		logger:      logger,
	}
}

func (uc *GetCountryUseCase) Execute(ctx context.Context, req requests.GetCountryRequest) (*requests.GetCountryResponse, error) {
	uc.logger.Info("getting country", "country_id", req.ID)

	country, err := uc.countryRepo.FindByID(ctx, req.ID)
	if err != nil || country == nil {
		uc.logger.Warn("country not found", "country_id", req.ID)
		return nil, usecase_errors.ErrCountryNotFound
	}

	capitalName := ""
	if country.GetCapitalID() != uuid.Nil {
		capital, _ := uc.cityRepo.FindByID(ctx, country.GetCapitalID())
		if capital != nil {
			capitalName = capital.GetName()
		}
	}

	holidays, _ := uc.countryRepo.FindHolidaysByCountry(ctx, req.ID)
	holidayResponses := make([]requests.HolidayResponse, 0, len(holidays))
	for _, h := range holidays {
		holidayResponses = append(holidayResponses, requests.HolidayResponse{
			ID:          h.GetId(),
			Name:        h.GetName(),
			Date:        h.GetDate(),
			Description: h.GetDescription(),
			IsNational:  h.IsNational(),
		})
	}

	imageURL := ""
	if country.GetImageID() != uuid.Nil {
		url, err := uc.staticRepo.GetFileURL(ctx, country.GetImageID().String())
		if err == nil {
			imageURL = url
		} else {
			uc.logger.Warn("failed to get image URL", "image_id", country.GetImageID(), "error", err)
		}
	}

	uc.logger.Info("country retrieved", "country_id", req.ID)

	return &requests.GetCountryResponse{
		ID:               country.GetId(),
		Name:             country.GetName(),
		Area:             country.GetArea(),
		Population:       country.GetPopulation(),
		GDP:              country.GetGdp(),
		Currency:         country.GetCurrency(),
		VisaRequirements: country.GetVisaRequirements(),
		Description:      country.GetDescription(),
		SafetyTips:       country.GetSafetyTips(),
		BestSeason:       country.GetBestSeason(),
		Language:         country.GetLanguage(),
		PhoneCode:        country.GetPhoneCode(),
		Religion:         country.GetReligion(),
		CapitalID:        country.GetCapitalID(),
		CapitalName:      capitalName,
		ImageURL:         imageURL,
		Holidays:         holidayResponses,
	}, nil
}
