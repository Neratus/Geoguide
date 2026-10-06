package builders

import (
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type CityBuilder struct {
	city *domain.City
}

func NewCityBuilder() *CityBuilder {
	id := domain.CityID(uuid.New())
	countryID := domain.CountryID(uuid.Nil)
	imageID := DefaultImageID
	coords := domain.NewCoordinates(40.0+rnd.Float64()*15.0, -10.0+rnd.Float64()*40.0)

	city := domain.NewCityFromDB(
		id, cityNames[rnd.Intn(len(cityNames))],
		1000000, true, coords,
		"Описание города", "UTC+3", "Советы путешественникам",
		countryID, imageID,
	)
	return &CityBuilder{city: city}
}

func (b *CityBuilder) WithID(id domain.CityID) *CityBuilder {
	b.city.SetId(id)
	return b
}

func (b *CityBuilder) WithName(name string) *CityBuilder {
	b.city.SetName(name)
	return b
}

func (b *CityBuilder) WithCountryID(countryID domain.CountryID) *CityBuilder {
	b.city.SetCountryId(countryID)
	return b
}

func (b *CityBuilder) WithImageID(imageID domain.ImageID) *CityBuilder {
	b.city.SetImageId(imageID)
	return b
}

func (b *CityBuilder) Build() *domain.City {
	return b.city
}

type CityDistrictBuilder struct {
	district *domain.CityDistrict
}

func NewCityDistrictBuilder() *CityDistrictBuilder {
	id := domain.CityDistrictID(uuid.New())
	cityID := domain.CityID(uuid.New())
	imageID := DefaultImageID
	coords := domain.NewCoordinates(55.7558, 37.6173)

	district := domain.NewCityDistrictFromDB(
		id, "Центральный район", "Описание района", coords, cityID, imageID,
	)
	return &CityDistrictBuilder{district: district}
}

func (b *CityDistrictBuilder) WithCityID(cityID domain.CityID) *CityDistrictBuilder {
	b.district.SetCityId(cityID)
	return b
}

func (b *CityDistrictBuilder) Build() *domain.CityDistrict {
	return b.district
}
