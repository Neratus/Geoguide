package place_repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func generateTestCountry(id domain.CountryID) *domain.Country {
	country, _ := domain.NewCountry(
		id, "TestCountry", 1000, 1000, 1000,
		"USD", "none", "desc", "safe", "summer",
		"en", "+1", "none", uuid.Nil, uuid.Nil,
	)
	return country
}

func generateTestCity(id domain.CityID, countryID domain.CountryID) *domain.City {
	coords := domain.NewCoordinates(0, 0)
	city, _ := domain.NewCity(
		id, "TestCity", 1000, false, coords,
		"desc", "UTC", "tips", countryID, uuid.Nil,
	)
	return city
}

var userCounter int

func generateTestUser(id domain.UserID) *domain.User {
	userCounter++
	username := fmt.Sprintf("testuser%d", userCounter)
	email := fmt.Sprintf("test%d@example.com", userCounter)
	user, err := domain.NewUser(
		id,
		username,
		email,
		"hash",
		"+123",
		nil,
		"US",
		domain.UserRoleUser,
	)
	if err != nil {
		panic(err)
	}
	return user
}

func generateTestPlace(id domain.PlaceID, cityID domain.CityID) *domain.Place {
	if cityID == uuid.Nil {
		panic("cityID is required and must be non-zero")
	}
	coords := domain.NewCoordinates(48.8566, 2.3522)
	place, err := domain.NewPlace(
		id,
		"Eiffel Tower",
		"Landmark",
		"Iron tower",
		coords,
		"Champ de Mars",
		"09:00-23:45",
		"€17-€26",
		90,
		"+33144112323",
		"https://example.com",
		cityID,
		uuid.Nil,
		uuid.Nil,
		"",
		"",
	)
	if err != nil {
		panic(err)
	}
	return place
}

func generateTestReview(id domain.ReviewID, placeID domain.PlaceID, userID domain.UserID) *domain.Review {
	if placeID == uuid.Nil {
		panic("placeID is required and must be non-zero")
	}
	if userID == uuid.Nil {
		panic("userID is required and must be non-zero")
	}
	visitDate, _ := time.Parse("2006-01-02", "2024-07-01")
	review, err := domain.NewReview(
		id,
		5,
		"Great experience!",
		visitDate,
		userID,
		placeID,
		uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
	return review
}

func placeRepository_SaveNewPlace(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	err := countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCity(uuid.Nil, country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	place := generateTestPlace(uuid.Nil, city.GetId())
	err = placeRepo.Save(ctx, place)
	require.NoError(t, err)
	require.NotZero(t, place.GetId())

	found, err := placeRepo.FindByID(ctx, place.GetId())
	require.NoError(t, err)
	require.Equal(t, place.GetId(), found.GetId())
}

func placeRepository_SaveExistingPlace(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	place := generateTestPlace(uuid.Nil, city.GetId())
	err := placeRepo.Save(ctx, place)
	require.NoError(t, err)
	id := place.GetId()

	place.SetName("New Name")
	err = placeRepo.Save(ctx, place)
	require.NoError(t, err)

	found, err := placeRepo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "New Name", found.GetName())
}

func placeRepository_FindByID_NotFound(t *testing.T, placeRepo interfaces.PlaceRepository) {
	_, err := placeRepo.FindByID(context.Background(), uuid.New())
	require.Error(t, err)
}

func placeRepository_FindByID_Found(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	place := generateTestPlace(uuid.Nil, city.GetId())
	err := placeRepo.Save(ctx, place)
	require.NoError(t, err)
	id := place.GetId()

	found, err := placeRepo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, place.GetId(), found.GetId())
}

func placeRepository_FindByCity(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city1 := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city1)
	city2 := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city2)

	p1 := generateTestPlace(uuid.Nil, city1.GetId())
	p1.SetName("Place1")
	_ = placeRepo.Save(ctx, p1)

	p2 := generateTestPlace(uuid.Nil, city1.GetId())
	p2.SetName("Place2")
	_ = placeRepo.Save(ctx, p2)

	p3 := generateTestPlace(uuid.Nil, city2.GetId())
	p3.SetName("Place3")
	_ = placeRepo.Save(ctx, p3)

	found, err := placeRepo.FindByCity(ctx, city1.GetId())
	require.NoError(t, err)
	require.Len(t, found, 2)

	found, err = placeRepo.FindByCity(ctx, city2.GetId())
	require.NoError(t, err)
	require.Len(t, found, 1)
}

func placeRepository_FindByCategory(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	cat1 := "Landmark"
	cat2 := "Museum"

	p1 := generateTestPlace(uuid.Nil, city.GetId())
	p1.SetCategory(cat1)
	p1.SetName("Eiffel")
	_ = placeRepo.Save(ctx, p1)

	p2 := generateTestPlace(uuid.Nil, city.GetId())
	p2.SetCategory(cat1)
	p2.SetName("Arc")
	_ = placeRepo.Save(ctx, p2)

	p3 := generateTestPlace(uuid.Nil, city.GetId())
	p3.SetCategory(cat2)
	p3.SetName("Louvre")
	_ = placeRepo.Save(ctx, p3)

	found, err := placeRepo.FindByCategory(ctx, cat1)
	require.NoError(t, err)
	require.Len(t, found, 2)

	found, err = placeRepo.FindByCategory(ctx, cat2)
	require.NoError(t, err)
	require.Len(t, found, 1)
}

func placeRepository_UpdatePlace_Success(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	place := generateTestPlace(uuid.Nil, city.GetId())
	err := placeRepo.Save(ctx, place)
	require.NoError(t, err)
	id := place.GetId()

	place.SetName("Updated")
	err = placeRepo.Update(ctx, place)
	require.NoError(t, err)

	found, err := placeRepo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Updated", found.GetName())
}

func placeRepository_DeletePlace_Success(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	place := generateTestPlace(uuid.Nil, city.GetId())
	err := placeRepo.Save(ctx, place)
	require.NoError(t, err)
	id := place.GetId()

	err = placeRepo.Delete(ctx, id)
	require.NoError(t, err)

	_, err = placeRepo.FindByID(ctx, id)
	require.Error(t, err)
}

func placeRepository_UpdateRating_Success(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	place := generateTestPlace(uuid.Nil, city.GetId())
	err := placeRepo.Save(ctx, place)
	require.NoError(t, err)
	id := place.GetId()

	err = placeRepo.UpdateRating(ctx, id, 4.7, 15)
	require.NoError(t, err)

	found, err := placeRepo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, 4.7, found.GetAvgRating())
	require.Equal(t, int32(15), found.GetReviewCnt())
}

func placeRepository_SaveNewReview(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)
	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)

	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	review := generateTestReview(uuid.Nil, place.GetId(), user.GetId())
	err := placeRepo.SaveReview(ctx, review)
	require.NoError(t, err)
	require.NotZero(t, review.GetId())

	found, err := placeRepo.FindReviewByID(ctx, review.GetId())
	require.NoError(t, err)
	require.Equal(t, review.GetId(), found.GetId())
}

func placeRepository_SaveExistingReview(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)
	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)
	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	review := generateTestReview(uuid.Nil, place.GetId(), user.GetId())
	err := placeRepo.SaveReview(ctx, review)
	require.NoError(t, err)
	id := review.GetId()

	review.SetComment("Updated comment")
	err = placeRepo.SaveReview(ctx, review)
	require.NoError(t, err)

	found, err := placeRepo.FindReviewByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Updated comment", found.GetComment())
}

func placeRepository_FindReviewByID_NotFound(t *testing.T, placeRepo interfaces.PlaceRepository) {
	_, err := placeRepo.FindReviewByID(context.Background(), uuid.New())
	require.Error(t, err)
}

func placeRepository_FindReviewByID_Found(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)
	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)
	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	review := generateTestReview(uuid.Nil, place.GetId(), user.GetId())
	_ = placeRepo.SaveReview(ctx, review)
	id := review.GetId()

	found, err := placeRepo.FindReviewByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, review.GetId(), found.GetId())
}
func placeRepository_FindReviewsByPlace(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	err := countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCity(uuid.Nil, country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	place1 := generateTestPlace(uuid.Nil, city.GetId())
	err = placeRepo.Save(ctx, place1)
	require.NoError(t, err)

	place2 := generateTestPlace(uuid.Nil, city.GetId())
	err = placeRepo.Save(ctx, place2)
	require.NoError(t, err)

	user1 := generateTestUser(uuid.Nil)
	err = userRepo.Save(ctx, user1)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user1.GetId())

	user2 := generateTestUser(uuid.Nil)
	err = userRepo.Save(ctx, user2)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user2.GetId())

	r1 := generateTestReview(uuid.Nil, place1.GetId(), user1.GetId())
	r2 := generateTestReview(uuid.Nil, place1.GetId(), user2.GetId())
	r3 := generateTestReview(uuid.Nil, place2.GetId(), user1.GetId())

	err = placeRepo.SaveReview(ctx, r1)
	require.NoError(t, err)
	err = placeRepo.SaveReview(ctx, r2)
	require.NoError(t, err)
	err = placeRepo.SaveReview(ctx, r3)
	require.NoError(t, err)

	err = placeRepo.ModerateReview(ctx, r1.GetId(), true, "")
	require.NoError(t, err)
	err = placeRepo.ModerateReview(ctx, r2.GetId(), true, "")
	require.NoError(t, err)
	err = placeRepo.ModerateReview(ctx, r3.GetId(), true, "")
	require.NoError(t, err)

	reviews, err := placeRepo.FindReviewsByPlace(ctx, place1.GetId())
	require.NoError(t, err)
	require.Len(t, reviews, 2)

	reviews, err = placeRepo.FindReviewsByPlace(ctx, place2.GetId())
	require.NoError(t, err)
	require.Len(t, reviews, 1)
}

func placeRepository_FindReviewsByUser(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)
	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)

	user1 := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user1)
	user2 := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user2)

	r1 := generateTestReview(uuid.Nil, place.GetId(), user1.GetId())
	r2 := generateTestReview(uuid.Nil, place.GetId(), user1.GetId())
	r3 := generateTestReview(uuid.Nil, place.GetId(), user2.GetId())
	_ = placeRepo.SaveReview(ctx, r1)
	_ = placeRepo.SaveReview(ctx, r2)
	_ = placeRepo.SaveReview(ctx, r3)

	reviews, err := placeRepo.FindReviewsByUser(ctx, user1.GetId())
	require.NoError(t, err)
	require.Len(t, reviews, 2)

	reviews, err = placeRepo.FindReviewsByUser(ctx, user2.GetId())
	require.NoError(t, err)
	require.Len(t, reviews, 1)
}

func placeRepository_UpdateReview_Success(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)
	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)
	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	review := generateTestReview(uuid.Nil, place.GetId(), user.GetId())
	_ = placeRepo.SaveReview(ctx, review)
	id := review.GetId()

	review.SetRating(4)
	err := placeRepo.UpdateReview(ctx, review)
	require.NoError(t, err)

	found, err := placeRepo.FindReviewByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, 4, found.GetRating())
}

func placeRepository_DeleteReview_Success(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)
	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)
	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	review := generateTestReview(uuid.Nil, place.GetId(), user.GetId())
	_ = placeRepo.SaveReview(ctx, review)
	id := review.GetId()

	err := placeRepo.DeleteReview(ctx, id)
	require.NoError(t, err)

	_, err = placeRepo.FindReviewByID(ctx, id)
	require.Error(t, err)
}

func placeRepository_ModerateReview_Success(t *testing.T,
	placeRepo interfaces.PlaceRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	userRepo interfaces.UserRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)
	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)
	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)
	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	review := generateTestReview(uuid.Nil, place.GetId(), user.GetId())
	_ = placeRepo.SaveReview(ctx, review)
	id := review.GetId()

	err := placeRepo.ModerateReview(ctx, id, true, "looks good")
	require.NoError(t, err)

}
