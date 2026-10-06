package interfaces

import (
	"context"

	"github.com/Neratus/geoguide/internal/domain"
)

type CityRepository interface {
	Save(ctx context.Context, city *domain.City) error
	FindByID(ctx context.Context, id domain.CityID) (*domain.City, error)
	FindByCountry(ctx context.Context, countryID domain.CountryID) ([]*domain.City, error)
	FindByName(ctx context.Context, name string) (*domain.City, error)
	SearchCities(ctx context.Context, query string, limit, offset int) ([]*domain.City, error)
	Update(ctx context.Context, city *domain.City) error
	Delete(ctx context.Context, id domain.CityID) error

	SaveDistrict(ctx context.Context, district *domain.CityDistrict) error
	FindDistrictsByCity(ctx context.Context, cityID domain.CityID) ([]*domain.CityDistrict, error)
	UpdateDistrict(ctx context.Context, district *domain.CityDistrict) error
	DeleteDistrict(ctx context.Context, id domain.CityDistrictID) error

	SaveTransportNode(ctx context.Context, node *domain.TransportNode) error
	FindTransportNodeByID(ctx context.Context, id domain.NodeID) (*domain.TransportNode, error)
	FindTransportNodesByCity(ctx context.Context, cityID domain.CityID) ([]*domain.TransportNode, error)
	UpdateTransportNode(ctx context.Context, node *domain.TransportNode) error
	DeleteTransportNode(ctx context.Context, id domain.NodeID) error
}
