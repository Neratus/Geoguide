package domain

import (
	"strings"
)

type City struct {
	id          CityID
	name        string
	population  int64
	isCapital   bool
	coordinates Coordinates
	description string
	timezone    string
	travelTips  string
	countryID   CountryID
	imageID     ImageID
}

func (c *City) GetId() CityID {
	return c.id
}

func (c *City) GetName() string {
	return c.name
}

func (c *City) GetPopulation() int64 {
	return c.population
}

func (c *City) IsCapital() bool {
	return c.isCapital
}

func (c *City) GetCoordinates() Coordinates {
	return c.coordinates
}

func (c *City) GetDescription() string {
	return c.description
}

func (c *City) GetTimezone() string {
	return c.timezone
}

func (c *City) GetTravelTips() string {
	return c.travelTips
}

func (c *City) GetCountryId() CountryID {
	return c.countryID
}

func (c *City) GetImageId() ImageID {
	return c.imageID
}

func (c *City) SetId(id CityID) {
	c.id = id
}

func (c *City) SetName(name string) {
	c.name = name
}

func (c *City) SetPopulation(population int64) {
	c.population = population
}

func (c *City) SetIsCapital(isCapital bool) {
	c.isCapital = isCapital
}

func (c *City) SetCoordinates(coordinates Coordinates) {
	c.coordinates = coordinates
}

func (c *City) SetDescription(description string) {
	c.description = description
}

func (c *City) SetTimezone(timezone string) {
	c.timezone = timezone
}

func (c *City) SetTravelTips(travelTips string) {
	c.travelTips = travelTips
}

func (c *City) SetCountryId(countryID CountryID) {
	c.countryID = countryID
}

func (c *City) SetImageId(imageID ImageID) {
	c.imageID = imageID
}

func NewCity(
	id CityID,
	name string,
	population int64,
	isCapital bool,
	coordinates Coordinates,
	description string,
	timezone string,
	travelTips string,
	countryID CountryID,
	imageID ImageID,
) (*City, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyCityName
	}
	if len(name) > GetConfig().MaxCityNameLength {
		return nil, ErrCityNameTooLong
	}

	if description == "" {
		return nil, ErrEmptyCityDescription
	}

	if len(description) > GetConfig().MaxCityDescriptionLength {
		return nil, ErrCityDescriptionTooLong
	}

	if timezone != "" && len(timezone) > GetConfig().MaxCityTimezoneLength {
		return nil, ErrCityTimezoneTooLong
	}

	if len(travelTips) > GetConfig().MaxCityTravelTipsLength {
		return nil, ErrCityTravelTipsTooLong
	}

	if population < 0 {
		return nil, ErrCityNegativePopulation
	}

	if countryID == (CountryID{}) {
		return nil, ErrInvalidCountryID
	}

	if err := validateCoordinates(coordinates); err != nil {
		return nil, err
	}

	return &City{
		id:          id,
		name:        name,
		population:  population,
		isCapital:   isCapital,
		coordinates: coordinates,
		description: description,
		timezone:    timezone,
		travelTips:  travelTips,
		countryID:   countryID,
		imageID:     imageID,
	}, nil
}

func NewCityFromDB(
	id CityID,
	name string,
	population int64,
	isCapital bool,
	coordinates Coordinates,
	description string,
	timezone string,
	travelTips string,
	countryID CountryID,
	imageID ImageID,
) *City {
	return &City{
		id:          id,
		name:        name,
		population:  population,
		isCapital:   isCapital,
		coordinates: coordinates,
		description: description,
		timezone:    timezone,
		travelTips:  travelTips,
		countryID:   countryID,
		imageID:     imageID,
	}
}
