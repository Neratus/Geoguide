package domain

import (
	"fmt"
	"regexp"
	"strings"
)

type Place struct {
	id                  PlaceID
	name                string
	category            string
	description         string
	coordinates         Coordinates
	address             string
	openingHours        string
	priceInfo           string
	avgVisitDurationMin int32
	avgRating           float64
	reviewCnt           int32
	contactPhone        string
	website             string
	cityID              CityID
	districtID          CityDistrictID
	imageID             ImageID
	imageURL            string
	cityName            string
	districtName        string
}

func (p *Place) GetId() PlaceID                { return p.id }
func (p *Place) GetName() string               { return p.name }
func (p *Place) GetCategory() string           { return p.category }
func (p *Place) GetDescription() string        { return p.description }
func (p *Place) GetCoordinates() Coordinates   { return p.coordinates }
func (p *Place) GetAddress() string            { return p.address }
func (p *Place) GetOpeningHours() string       { return p.openingHours }
func (p *Place) GetPriceInfo() string          { return p.priceInfo }
func (p *Place) GetAvgVisitDurationMin() int32 { return p.avgVisitDurationMin }
func (p *Place) GetAvgRating() float64         { return p.avgRating }
func (p *Place) GetReviewCnt() int32           { return p.reviewCnt }
func (p *Place) GetContactPhone() string       { return p.contactPhone }
func (p *Place) GetWebsite() string            { return p.website }
func (p *Place) GetCityID() CityID             { return p.cityID }
func (p *Place) GetDistrictID() CityDistrictID { return p.districtID }
func (p *Place) GetImageID() ImageID           { return p.imageID }
func (p *Place) GetCityName() string {
	return p.cityName
}

func (p *Place) GetDistrictName() string {
	return p.districtName
}

func (p *Place) SetId(id PlaceID)                 { p.id = id }
func (p *Place) SetName(name string)              { p.name = name }
func (p *Place) SetCategory(cat string)           { p.category = cat }
func (p *Place) SetDescription(desc string)       { p.description = desc }
func (p *Place) SetCoordinates(c Coordinates)     { p.coordinates = c }
func (p *Place) SetAddress(addr string)           { p.address = addr }
func (p *Place) SetOpeningHours(oh string)        { p.openingHours = oh }
func (p *Place) SetPriceInfo(pi string)           { p.priceInfo = pi }
func (p *Place) SetAvgVisitDurationMin(d int32)   { p.avgVisitDurationMin = d }
func (p *Place) SetAvgRating(r float64)           { p.avgRating = r }
func (p *Place) SetReviewCnt(cnt int32)           { p.reviewCnt = cnt }
func (p *Place) SetContactPhone(phone string)     { p.contactPhone = phone }
func (p *Place) SetWebsite(web string)            { p.website = web }
func (p *Place) SetCityID(cid CityID)             { p.cityID = cid }
func (p *Place) SetDistrictID(did CityDistrictID) { p.districtID = did }
func (p *Place) SetImageID(iid ImageID)           { p.imageID = iid }
func (p *Place) SetImageURL(url string)           { p.imageURL = url }

var phoneRegex = regexp.MustCompile(`^\+?[1-9][0-9]{7,14}$`)

func ValidatePhone(phone string) bool {
	if phone == "" {
		return true
	}
	return phoneRegex.MatchString(phone)
}

var urlRegex = regexp.MustCompile(`^(https?://)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(:\d+)?(/.*)?$`)

func ValidateURL(url string) bool {
	if url == "" {
		return true
	}
	return urlRegex.MatchString(url)
}

func (p *Place) CoordStr() string {
	return fmt.Sprintf("%f,%f", p.coordinates.Lat(), p.coordinates.Lng())
}

func NewPlace(
	id PlaceID,
	name, category, description string,
	coordinates Coordinates,
	address, openingHours, priceInfo string,
	avgVisitDurationMin int32,
	contactPhone, website string,
	cityID CityID,
	districtID CityDistrictID,
	imageID ImageID,
	cityName string,
	districtName string,
) (*Place, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyPlaceName
	}
	if len(name) > GetConfig().MaxPlaceNameLength {
		return nil, ErrPlaceNameTooLong
	}

	category = strings.TrimSpace(category)
	if category == "" {
		return nil, ErrEmptyPlaceCategory
	}
	if len(category) > GetConfig().MaxPlaceCategoryLength {
		return nil, ErrPlaceCategoryTooLong
	}

	if len(description) > GetConfig().MaxPlaceDescriptionLength {
		return nil, ErrPlaceDescriptionTooLong
	}

	if len(address) > GetConfig().MaxPlaceAddressLength {
		return nil, ErrPlaceAddressTooLong
	}

	if avgVisitDurationMin < 0 {
		return nil, ErrNegativeVisitDuration
	}

	if len(contactPhone) > GetConfig().MaxContactPhoneLength {
		return nil, ErrPlaceContactPhoneTooLong
	}
	if len(website) > GetConfig().MaxWebsiteLength {
		return nil, ErrPlaceWebsiteTooLong
	}

	if cityID == (CityID{}) {
		return nil, ErrInvalidCityID
	}

	if err := validateCoordinates(coordinates); err != nil {
		return nil, err
	}

	if !ValidatePhone(contactPhone) {
		return nil, ErrInvalidPhone
	}

	if !ValidateURL(website) {
		return nil, ErrInvalidWebsite
	}

	return &Place{
		id:                  id,
		name:                name,
		category:            category,
		description:         description,
		coordinates:         coordinates,
		address:             address,
		openingHours:        openingHours,
		priceInfo:           priceInfo,
		avgVisitDurationMin: avgVisitDurationMin,
		avgRating:           0.0,
		reviewCnt:           0,
		contactPhone:        contactPhone,
		website:             website,
		cityID:              cityID,
		districtID:          districtID,
		imageID:             imageID,
		cityName:            cityName,
		districtName:        districtName,
	}, nil
}

func NewPlaceFromDB(
	id PlaceID,
	name, category, description string,
	coordinates Coordinates,
	address, openingHours, priceInfo string,
	avgVisitDurationMin int32,
	avgRating float64,
	reviewCnt int32,
	contactPhone, website string,
	cityID CityID,
	districtID CityDistrictID,
	imageID ImageID,
) *Place {
	return &Place{
		id:                  id,
		name:                name,
		category:            category,
		description:         description,
		coordinates:         coordinates,
		address:             address,
		openingHours:        openingHours,
		priceInfo:           priceInfo,
		avgVisitDurationMin: avgVisitDurationMin,
		avgRating:           avgRating,
		reviewCnt:           reviewCnt,
		contactPhone:        contactPhone,
		website:             website,
		cityID:              cityID,
		districtID:          districtID,
		imageID:             imageID,
	}
}
