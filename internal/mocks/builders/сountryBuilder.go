package builders

import (
	"math/rand"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

var cntRnd = rand.New(rand.NewSource(42))

var DefaultImageID domain.ImageID

var countryNames = []string{"Франция", "Италия", "Япония", "Бразилия", "Египет"}
var cityNames = []string{"Париж", "Рим", "Токио", "Рио-де-Жанейро", "Каир"}

type CountryBuilder struct {
	country *domain.Country
}

func NewCountryBuilder() *CountryBuilder {
	id := domain.CountryID(uuid.New())
	capitalID := domain.CityID(uuid.Nil)
	imageID := DefaultImageID

	country := domain.NewCountryFromDB(
		id, countryNames[cntRnd.Intn(len(countryNames))],
		1000000.0, 50000.0, 50000000,
		"USD", "Нет", "Описание страны", "Безопасно", "Лето",
		"Английский", "+1", "Христианство", capitalID, imageID,
	)
	return &CountryBuilder{country: country}
}

func (b *CountryBuilder) WithID(id domain.CountryID) *CountryBuilder {
	b.country.SetId(id)
	return b
}

func (b *CountryBuilder) WithName(name string) *CountryBuilder {
	b.country.SetName(name)
	return b
}

func (b *CountryBuilder) WithCapitalID(capitalID domain.CityID) *CountryBuilder {
	b.country.SetCapitalID(capitalID)
	return b
}

func (b *CountryBuilder) Build() *domain.Country {
	return b.country
}
