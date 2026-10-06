package domain

import "strings"

type CityDistrict struct {
	id          CityDistrictID
	name        string
	description string
	coordinates Coordinates
	city_id     CityID
	image_id    ImageID
}

func (c *CityDistrict) GetId() CityDistrictID {
	return c.id
}

func (c *CityDistrict) GetName() string {
	return c.name
}

func (c *CityDistrict) GetDescription() string {
	return c.description
}

func (c *CityDistrict) GetCoordinates() Coordinates {
	return c.coordinates
}

func (c *CityDistrict) GetCityId() CityID {
	return c.city_id
}

func (c *CityDistrict) GetImageId() ImageID {
	return c.image_id
}

func (c *CityDistrict) SetId(id CityDistrictID) {
	c.id = id
}

func (c *CityDistrict) SetName(name string) {
	c.name = name
}

func (c *CityDistrict) SetDescription(description string) {
	c.description = description
}

func (c *CityDistrict) SetCoordinates(coordinates Coordinates) {
	c.coordinates = coordinates
}

func (c *CityDistrict) SetCityId(city_id CityID) {
	c.city_id = city_id
}

func (c *CityDistrict) SetImageId(image_id ImageID) {
	c.image_id = image_id
}

func NewCityDistrict(name, description string, coordinates Coordinates, cityID CityID, imageID ImageID) (*CityDistrict, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyCityDistrictName
	}
	if len(name) > GetConfig().MaxDistrictNameLength {
		return nil, ErrCityDistrictNameTooLong
	}

	if description == "" {
		return nil, ErrEmptyCityDistrictDescription
	}

	if len(description) > GetConfig().MaxDistrictDescriptionLength {
		return nil, ErrCityDistrictDescriptionTooLong
	}

	if cityID == (CityID{}) {
		return nil, ErrInvalidCityID
	}

	if err := validateCoordinates(coordinates); err != nil {
		return nil, err
	}

	return &CityDistrict{
		id:          (CityID{}),
		name:        name,
		description: description,
		coordinates: coordinates,
		city_id:     cityID,
		image_id:    imageID,
	}, nil
}

func NewCityDistrictFromDB(id CityDistrictID, name string, description string, coordinates Coordinates, cityID CityID, imageID ImageID) *CityDistrict {
	return &CityDistrict{
		id:          id,
		name:        name,
		description: description,
		coordinates: coordinates,
		city_id:     cityID,
		image_id:    imageID,
	}
}
