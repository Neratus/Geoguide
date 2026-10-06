package builders

import (
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

type UserBuilder struct {
	user *domain.User
}

func NewUserBuilder() *UserBuilder {
	id := domain.UserID(uuid.New())
	hash, _ := domain.HashPassword("password123")

	uniqueEmail := fmt.Sprintf("test_%s@example.com", uuid.New().String()[:8])
	uniqueSuffix := uuid.New().String()[:8]
	username := fmt.Sprintf("testuser_%s", uniqueSuffix)
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	user, err := domain.NewUser(
		id,
		username,
		uniqueEmail,
		hash,
		"+1234567890",
		&birthDate,
		"France",
		domain.UserRoleUser,
	)
	if err != nil {
		panic(err)
	}
	return &UserBuilder{user: user}
}

func (b *UserBuilder) WithID(id domain.UserID) *UserBuilder {
	b.user.SetId(id)
	return b
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.user.SetUsername(username)
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.SetEmail(email)
	return b
}

func (b *UserBuilder) WithPhone(phone string) *UserBuilder {
	b.user.SetPhone(phone)
	return b
}

func (b *UserBuilder) WithCountry(country string) *UserBuilder {
	b.user.SetCountryOfResidence(country)
	return b
}

func (b *UserBuilder) Build() *domain.User {
	return b.user
}
