package builders

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

var rnd = rand.New(rand.NewSource(42))

var categories = []string{"Музей", "Парк", "Ресторан", "Памятник", "Театр", "Торговый центр"}

func NewPlaceBuilder() *PlaceBuilder {
	id := domain.PlaceID(uuid.New())
	name := fmt.Sprintf("Достопримечательность_%d", rnd.Intn(10000))
	category := categories[rnd.Intn(len(categories))]
	description := fmt.Sprintf("Описание для %s.", name)
	coords := domain.NewCoordinates(40.0+rnd.Float64()*15.0, -10.0+rnd.Float64()*40.0)
	cityID := domain.CityID(uuid.New())
	districtID := domain.CityDistrictID(uuid.Nil)
	imageID := DefaultImageID

	place, err := domain.NewPlace(
		id, name, category, description, coords,
		"ул. Тестовая, 1", "09:00-18:00", "Free", 60,
		"+1234567890", "https://example.com", cityID, districtID, imageID,
		"TestCity", "TestDistrict",
	)
	if err != nil {
		panic(err)
	}
	return &PlaceBuilder{place: place}
}

type PlaceBuilder struct {
	place *domain.Place
}

func (b *PlaceBuilder) WithID(id domain.PlaceID) *PlaceBuilder {
	b.place.SetId(id)
	return b
}

func (b *PlaceBuilder) WithCityID(cityID domain.CityID) *PlaceBuilder {
	b.place.SetCityID(cityID)
	return b
}

func (b *PlaceBuilder) WithCategory(category string) *PlaceBuilder {
	b.place.SetCategory(category)
	return b
}

func (b *PlaceBuilder) Build() *domain.Place {
	return b.place
}

type ReviewBuilder struct {
	review *domain.Review
}

func NewReviewBuilder() *ReviewBuilder {
	id := domain.ReviewID(uuid.New())
	userID := domain.UserID(uuid.New())
	placeID := domain.PlaceID(uuid.New())
	imageID := DefaultImageID

	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	review, err := domain.NewReview(
		id, 5, "Отличное место!", date, userID, placeID, imageID,
	)
	if err != nil {
		panic(err)
	}
	return &ReviewBuilder{review: review}
}

func (b *ReviewBuilder) WithID(id domain.ReviewID) *ReviewBuilder {
	b.review.SetId(id)
	return b
}

func (b *ReviewBuilder) WithUserID(userID domain.UserID) *ReviewBuilder {
	b.review.SetUserID(userID)
	return b
}

func (b *ReviewBuilder) WithPlaceID(placeID domain.PlaceID) *ReviewBuilder {
	b.review.SetPlaceID(placeID)
	return b
}

func (b *ReviewBuilder) WithRating(rating int) *ReviewBuilder {
	b.review.SetRating(rating)
	return b
}

func (b *ReviewBuilder) WithImageID(imageID domain.ImageID) *ReviewBuilder {
	b.review.SetImageID(imageID)
	return b
}

func (b *ReviewBuilder) Build() *domain.Review {
	return b.review
}
