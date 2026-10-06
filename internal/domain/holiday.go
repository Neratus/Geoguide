package domain

import (
	"strings"
	"time"
)

type Holiday struct {
	id          HolidayID
	name        string
	date        time.Time
	description string
	traditions  string
	history     string
	isNational  bool
	countryID   CountryID
	imageID     ImageID
}

func (h *Holiday) GetId() HolidayID        { return h.id }
func (h *Holiday) GetName() string         { return h.name }
func (h *Holiday) GetDate() time.Time      { return h.date }
func (h *Holiday) GetDescription() string  { return h.description }
func (h *Holiday) GetTraditions() string   { return h.traditions }
func (h *Holiday) GetHistory() string      { return h.history }
func (h *Holiday) IsNational() bool        { return h.isNational }
func (h *Holiday) GetCountryID() CountryID { return h.countryID }
func (h *Holiday) GetImageID() ImageID     { return h.imageID }

func (h *Holiday) SetId(id HolidayID)         { h.id = id }
func (h *Holiday) SetName(name string)        { h.name = name }
func (h *Holiday) SetDate(d time.Time)        { h.date = d }
func (h *Holiday) SetDescription(desc string) { h.description = desc }
func (h *Holiday) SetTraditions(trad string)  { h.traditions = trad }
func (h *Holiday) SetHistory(hist string)     { h.history = hist }
func (h *Holiday) SetNational(nat bool)       { h.isNational = nat }
func (h *Holiday) SetCountryID(cid CountryID) { h.countryID = cid }
func (h *Holiday) SetImageID(iid ImageID)     { h.imageID = iid }

func NewHoliday(
	id HolidayID,
	name string,
	date time.Time,
	description, traditions, history string,
	isNational bool,
	countryID CountryID,
	imageID ImageID,
) (*Holiday, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyHolidayName
	}
	if len(name) > GetConfig().MaxHolidayNameLength {
		return nil, ErrHolidayNameTooLong
	}
	if date.IsZero() {
		return nil, ErrInvalidHolidayDate
	}
	if countryID == (CountryID{}) {
		return nil, ErrInvalidCountryID
	}
	return &Holiday{
		id:          id,
		name:        name,
		date:        date,
		description: description,
		traditions:  traditions,
		history:     history,
		isNational:  isNational,
		countryID:   countryID,
		imageID:     imageID,
	}, nil
}

func NewHolidayFromDB(
	id HolidayID,
	name string,
	date time.Time,
	description, traditions, history string,
	isNational bool,
	countryID CountryID,
	imageID ImageID,
) *Holiday {
	return &Holiday{
		id:          id,
		name:        name,
		date:        date,
		description: description,
		traditions:  traditions,
		history:     history,
		isNational:  isNational,
		countryID:   countryID,
		imageID:     imageID,
	}
}
