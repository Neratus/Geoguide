package builders

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type HolidayBuilder struct {
	holiday *domain.Holiday
}

func NewHolidayBuilder() *HolidayBuilder {
	id := domain.HolidayID(uuid.New())
	countryID := domain.CountryID(uuid.New())
	imageID := DefaultImageID

	dateStr := "2024-01-01"
	date, _ := time.Parse("2006-01-02", dateStr)

	holiday := domain.NewHolidayFromDB(
		id, "Новый Год", date, "Празднование", "Традиции", "История", true, countryID, imageID,
	)
	return &HolidayBuilder{holiday: holiday}
}

func (b *HolidayBuilder) WithDate(date string) *HolidayBuilder {
	return b
}

func (b *HolidayBuilder) Build() *domain.Holiday {
	return b.holiday
}
