package catalog

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/google/uuid"
)

type SearchCountriesUseCase struct {
	countryRepo interfaces.CountryRepository
	logger      *slog.Logger
}

func NewSearchCountriesUseCase(countryRepo interfaces.CountryRepository, logger *slog.Logger) *SearchCountriesUseCase {
	return &SearchCountriesUseCase{countryRepo: countryRepo, logger: logger}
}

func (uc *SearchCountriesUseCase) Execute(ctx context.Context, req requests.SearchCountriesRequest) ([]requests.CountrySearchResponse, error) {
	uc.logger.Info("searching countries", "query", req.Query)
	allCountries, err := uc.countryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.Country, 0)
	q := strings.ToLower(req.Query)
	for _, c := range allCountries {
		if strings.Contains(strings.ToLower(c.GetName()), q) {
			filtered = append(filtered, *c)
		}
	}
	start := req.Offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + req.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	result := make([]requests.CountrySearchResponse, 0, end-start)
	for _, c := range filtered[start:end] {
		capital := ""
		if capID := c.GetCapitalID(); capID != domain.CityID(uuid.Nil) {
			capital = capID.String()
		}
		result = append(result, requests.CountrySearchResponse{
			ID:       c.GetId(),
			Name:     c.GetName(),
			Capital:  capital,
			Currency: c.GetCurrency(),
		})
	}
	return result, nil
}
