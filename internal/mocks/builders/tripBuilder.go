package builders

import (
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type TripBuilder struct {
	trip *domain.Trip
}

func NewTripBuilder() *TripBuilder {
	trip, _ := domain.NewTrip(
		domain.TripID(uuid.New()),
		"Тестовая поездка",
		time.Now().Add(24*time.Hour),
		time.Now().Add(48*time.Hour),
		1000.0,
		"DRAFT",
		"Тестовые заметки",
		domain.UserID(uuid.New()),
		DefaultImageID,
	)
	return &TripBuilder{trip: trip}
}

func (b *TripBuilder) WithID(id domain.TripID) *TripBuilder {
	b.trip.SetId(id)
	return b
}

func (b *TripBuilder) WithUserID(userID domain.UserID) *TripBuilder {
	b.trip.SetUserID(userID)
	return b
}

func (b *TripBuilder) WithTitle(title string) *TripBuilder {
	b.trip.SetTitle(title)
	return b
}

func (b *TripBuilder) Build() *domain.Trip {
	return b.trip
}
