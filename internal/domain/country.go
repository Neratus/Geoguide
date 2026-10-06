package domain

import "strings"

type Country struct {
	id               CountryID
	name             string
	area             float64
	population       int64
	gdp              float64
	currency         string
	visaRequirements string
	description      string
	safetyTips       string
	bestSeason       string
	language         string
	phoneCode        string
	religion         string
	capitalID        CityID
	imageID          ImageID
}

func (c *Country) GetId() CountryID            { return c.id }
func (c *Country) GetName() string             { return c.name }
func (c *Country) GetArea() float64            { return c.area }
func (c *Country) GetPopulation() int64        { return c.population }
func (c *Country) GetGdp() float64             { return c.gdp }
func (c *Country) GetCurrency() string         { return c.currency }
func (c *Country) GetVisaRequirements() string { return c.visaRequirements }
func (c *Country) GetDescription() string      { return c.description }
func (c *Country) GetSafetyTips() string       { return c.safetyTips }
func (c *Country) GetBestSeason() string       { return c.bestSeason }
func (c *Country) GetLanguage() string         { return c.language }
func (c *Country) GetPhoneCode() string        { return c.phoneCode }
func (c *Country) GetReligion() string         { return c.religion }
func (c *Country) GetCapitalID() CityID        { return c.capitalID }
func (c *Country) GetImageID() ImageID         { return c.imageID }

func (c *Country) SetId(id CountryID)           { c.id = id }
func (c *Country) SetName(name string)          { c.name = name }
func (c *Country) SetArea(area float64)         { c.area = area }
func (c *Country) SetPopulation(pop int64)      { c.population = pop }
func (c *Country) SetGdp(gdp float64)           { c.gdp = gdp }
func (c *Country) SetCurrency(currency string)  { c.currency = currency }
func (c *Country) SetVisaRequirements(v string) { c.visaRequirements = v }
func (c *Country) SetDescription(desc string)   { c.description = desc }
func (c *Country) SetSafetyTips(tips string)    { c.safetyTips = tips }
func (c *Country) SetBestSeason(season string)  { c.bestSeason = season }
func (c *Country) SetLanguage(lang string)      { c.language = lang }
func (c *Country) SetPhoneCode(code string)     { c.phoneCode = code }
func (c *Country) SetReligion(rel string)       { c.religion = rel }
func (c *Country) SetCapitalID(cid CityID)      { c.capitalID = cid }
func (c *Country) SetImageID(iid ImageID)       { c.imageID = iid }

func NewCountry(
	id CountryID,
	name string,
	area, gdp float64,
	population int64,
	currency, visaRequirements, description, safetyTips, bestSeason, language, phoneCode, religion string,
	capitalID CityID,
	imageID ImageID,
) (*Country, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyCountryName
	}
	if len(name) > GetConfig().MaxCountryNameLength {
		return nil, ErrCountryNameTooLong
	}
	if currency != "" && len(currency) > GetConfig().MaxCurrencyLength {
		return nil, ErrCurrencyTooLong
	}
	if language != "" && len(language) > GetConfig().MaxLanguageLength {
		return nil, ErrLanguageTooLong
	}
	if phoneCode != "" && len(phoneCode) > GetConfig().MaxPhoneCodeLength {
		return nil, ErrPhoneCodeTooLong
	}
	if bestSeason != "" && len(bestSeason) > GetConfig().MaxBestSeasonLength {
		return nil, ErrBestSeasonTooLong
	}
	if religion != "" && len(religion) > GetConfig().MaxReligionLength {
		return nil, ErrReligionTooLong
	}
	if area < 0 {
		return nil, ErrNegativeArea
	}
	if population < 0 {
		return nil, ErrCityNegativePopulation
	}
	if gdp < 0 {
		return nil, ErrNegativeGDP
	}
	return &Country{
		id:               id,
		name:             name,
		area:             area,
		population:       population,
		gdp:              gdp,
		currency:         currency,
		visaRequirements: visaRequirements,
		description:      description,
		safetyTips:       safetyTips,
		bestSeason:       bestSeason,
		language:         language,
		phoneCode:        phoneCode,
		religion:         religion,
		capitalID:        capitalID,
		imageID:          imageID,
	}, nil
}

func NewCountryFromDB(
	id CountryID,
	name string,
	area, gdp float64,
	population int64,
	currency, visaRequirements, description, safetyTips, bestSeason, language, phoneCode, religion string,
	capitalID CityID,
	imageID ImageID,
) *Country {
	return &Country{
		id:               id,
		name:             name,
		area:             area,
		population:       population,
		gdp:              gdp,
		currency:         currency,
		visaRequirements: visaRequirements,
		description:      description,
		safetyTips:       safetyTips,
		bestSeason:       bestSeason,
		language:         language,
		phoneCode:        phoneCode,
		religion:         religion,
		capitalID:        capitalID,
		imageID:          imageID,
	}
}
