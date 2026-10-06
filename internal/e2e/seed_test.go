package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/repository"
	"github.com/google/uuid"
)

func seedTestData(ctx context.Context, repos *repository.Repositories) error {
	russia := domain.NewCountryFromDB(
		domain.CountryID(uuid.New()),
		"Russia",
		17100000.0,
		2.5e12,
		144000000,
		"RUB",
		"Visa required for some countries",
		"Largest country in the world",
		"Be careful in big cities",
		"June-August",
		"Russian",
		"+7",
		"Orthodox Christianity",
		domain.CityID(uuid.Nil),
		domain.ImageID(uuid.Nil),
	)
	if err := repos.CountryRepo.Save(ctx, russia); err != nil {
		return fmt.Errorf("save country: %w", err)
	}
	fmt.Printf("Created country: %s (ID: %s)\n", russia.GetName(), russia.GetId())

	moscowCoords := domain.NewCoordinates(55.7558, 37.6173)
	moscow := domain.NewCityFromDB(
		domain.CityID(uuid.New()),
		"Moscow",
		12500000,
		true,
		moscowCoords,
		"Capital of Russia",
		"UTC+3",
		"Visit Red Square",
		russia.GetId(),
		domain.ImageID(uuid.Nil),
	)
	if err := repos.CityRepo.Save(ctx, moscow); err != nil {
		return fmt.Errorf("save city: %w", err)
	}
	fmt.Printf("Created city: %s (ID: %s)\n", moscow.GetName(), moscow.GetId())

	russia.SetCapitalID(moscow.GetId())
	if err := repos.CountryRepo.Update(ctx, russia); err != nil {
		fmt.Printf("WARN: failed to update country capital: %v\n", err)
	}

	redSquareCoords := domain.NewCoordinates(55.7537, 37.6212)
	redSquare := domain.NewPlaceFromDB(
		domain.PlaceID(uuid.New()),
		"Red Square",
		"attraction",
		"Famous square in Moscow",
		redSquareCoords,
		"Red Square, Moscow",
		"Always open",
		"Free",
		60,
		4.8,
		1500,
		"",
		"https://kreml.ru/ru",
		moscow.GetId(),
		domain.CityDistrictID(uuid.Nil),
		domain.ImageID(uuid.Nil),
	)
	if err := repos.PlaceRepo.Save(ctx, redSquare); err != nil {
		return fmt.Errorf("save place: %w", err)
	}
	fmt.Printf("Created place: %s (ID: %s)\n", redSquare.GetName(), redSquare.GetId())

	testUserID := domain.UserID(uuid.New())
	testUserEmail := "testuser@example.com"
	testUserPass := "TestPass123!"
	hashedTestPass, _ := domain.HashPassword(testUserPass)
	testUser := domain.NewUserFromDB(
		testUserID, "testuser", testUserEmail, hashedTestPass, "+79991234567", time.Now(), nil, "Russia", "", []string{},
		false, nil, "", domain.UserRoleUser, true, true, false, "", nil,
	)
	if err := repos.UserRepo.Save(ctx, testUser); err != nil {
		return fmt.Errorf("save test user: %w", err)
	}
	if err := repos.UserRepo.UpdateEmailVerified(ctx, testUser.GetId(), true); err != nil {
		return fmt.Errorf("verify test user email: %w", err)
	}

	savedUser, err := repos.UserRepo.FindByUsername(ctx, "testuser")
	if err != nil {
		return fmt.Errorf("find test user after verification: %w", err)
	}
	if !savedUser.IsEmailVerified() {
		return fmt.Errorf("test user email is NOT verified after UpdateEmailVerified — check repository implementation")
	}
	fmt.Printf(" Test user email verified: %v\n", savedUser.IsEmailVerified())

	fmt.Printf("Created test user: testuser / TestPass123! (ID: %s)\n", testUser.GetId())
	fmt.Println("Test data seeded successfully!")
	return nil
}
