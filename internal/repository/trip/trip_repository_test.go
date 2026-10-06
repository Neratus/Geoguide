package trip_repo

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
	country, err := domain.NewCountry(
		id,
		"TestCountry",
		100000, 1000000, 10000000,
		"USD", "None", "desc", "safe", "summer",
		"English", "+1", "none",
		uuid.Nil, uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
	return country
}

func generateTestCity(id domain.CityID, countryID domain.CountryID) *domain.City {
	coords := domain.NewCoordinates(0, 0)
	city, err := domain.NewCity(
		id, "TestCity", 100000, false, coords,
		"desc", "UTC", "tips", countryID, uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
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
	place, err := domain.NewPlace(
		id,
		"Test Place",
		"attraction",
		"",
		domain.NewCoordinates(0, 0),
		"",
		"",
		"",
		0,
		"",
		"",
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

func generateTestTrip(id domain.TripID, userID domain.UserID) *domain.Trip {
	start, _ := time.Parse("2006-01-02", "2024-06-01")
	end, _ := time.Parse("2006-01-02", "2024-06-10")
	trip, err := domain.NewTrip(
		id,
		"Summer Vacation",
		start, end,
		1500.00,
		domain.TripStatusPlanned,
		"",
		userID,
		uuid.Nil,
	)
	if err != nil {
		panic(err)
	}
	return trip
}

func tripRepository_SaveNewTrip(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	err := countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCity(uuid.Nil, country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	user := generateTestUser(uuid.Nil)
	err = userRepo.Save(ctx, user)
	require.NoError(t, err)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	err = tripRepo.Save(ctx, trip)
	require.NoError(t, err)
	require.NotZero(t, trip.GetId())
}

func tripRepository_SaveExistingTrip(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	err := tripRepo.Save(ctx, trip)
	require.NoError(t, err)
	id := trip.GetId()

	trip.SetTitle("Updated Title")
	err = tripRepo.Save(ctx, trip)
	require.NoError(t, err)

	found, err := tripRepo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Updated Title", found.GetTitle())
}

func tripRepository_FindByID_NotFound(t *testing.T, tripRepo interfaces.TripRepository) {
	_, err := tripRepo.FindByID(context.Background(), uuid.New())
	require.Error(t, err)
}

func tripRepository_FindByID_Found(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	err := tripRepo.Save(ctx, trip)
	require.NoError(t, err)
	id := trip.GetId()

	found, err := tripRepo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, trip.GetId(), found.GetId())
}

func tripRepository_FindByUser(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user1 := generateTestUser(uuid.Nil)
	err := userRepo.Save(ctx, user1)
	require.NoError(t, err)

	user2 := generateTestUser(uuid.Nil)
	err = userRepo.Save(ctx, user2)
	require.NoError(t, err)

	trip1 := generateTestTrip(uuid.Nil, user1.GetId())
	trip1.SetTitle("Trip 1")
	_ = tripRepo.Save(ctx, trip1)

	trip2 := generateTestTrip(uuid.Nil, user1.GetId())
	trip2.SetTitle("Trip 2")
	_ = tripRepo.Save(ctx, trip2)

	trip3 := generateTestTrip(uuid.Nil, user2.GetId())
	trip3.SetTitle("Trip 3")
	_ = tripRepo.Save(ctx, trip3)

	trips, err := tripRepo.FindByUser(ctx, user1.GetId())
	require.NoError(t, err)
	require.Len(t, trips, 2)

	trips, err = tripRepo.FindByUser(ctx, user2.GetId())
	require.NoError(t, err)
	require.Len(t, trips, 1)
}

func tripRepository_UpdateTrip_Success(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user := generateTestUser(uuid.Nil)
	err := userRepo.Save(ctx, user)
	require.NoError(t, err)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	err = tripRepo.Save(ctx, trip)
	require.NoError(t, err)
	id := trip.GetId()

	trip.SetTitle("New Title")
	err = tripRepo.Update(ctx, trip)
	require.NoError(t, err)

	found, err := tripRepo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "New Title", found.GetTitle())
}

func tripRepository_DeleteTrip_Success(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user := generateTestUser(uuid.Nil)
	err := userRepo.Save(ctx, user)
	require.NoError(t, err)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	err = tripRepo.Save(ctx, trip)
	require.NoError(t, err)
	id := trip.GetId()

	err = tripRepo.Delete(ctx, id)
	require.NoError(t, err)

	_, err = tripRepo.FindByID(ctx, id)
	require.Error(t, err)
}

func tripRepository_AddPlace_Success(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	err := countryRepo.Save(ctx, country)
	require.NoError(t, err)

	city := generateTestCity(uuid.Nil, country.GetId())
	err = cityRepo.Save(ctx, city)
	require.NoError(t, err)

	user := generateTestUser(uuid.Nil)
	err = userRepo.Save(ctx, user)
	require.NoError(t, err)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	err = tripRepo.Save(ctx, trip)
	require.NoError(t, err)

	place := generateTestPlace(uuid.Nil, city.GetId())
	err = placeRepo.Save(ctx, place)
	require.NoError(t, err)

	err = tripRepo.AddPlace(ctx, trip.GetId(), place.GetId(), 2, nil, 60, "nice view")
	require.NoError(t, err)

	places, err := tripRepo.GetPlaces(ctx, trip.GetId())
	require.NoError(t, err)
	require.Len(t, places, 1)
	require.Equal(t, place.GetId(), places[0].GetPlaceID())
}

func tripRepository_AddPlace_TripNotFound(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)

	err := tripRepo.AddPlace(ctx, uuid.New(), place.GetId(), 1, nil, 60, "")
	require.Error(t, err)
}

func tripRepository_RemovePlace_Success(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user := generateTestUser(uuid.Nil)
	err := userRepo.Save(ctx, user)
	require.NoError(t, err)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	err = tripRepo.Save(ctx, trip)
	require.NoError(t, err)

	place := generateTestPlace(uuid.Nil, city.GetId())
	err = placeRepo.Save(ctx, place)
	require.NoError(t, err)

	_ = tripRepo.AddPlace(ctx, trip.GetId(), place.GetId(), 2, nil, 60, "")

	err = tripRepo.RemovePlace(ctx, trip.GetId(), place.GetId())
	require.NoError(t, err)

	places, _ := tripRepo.GetPlaces(ctx, trip.GetId())
	require.Empty(t, places)
}

func tripRepository_GetPlaces_TripNotFound(t *testing.T, tripRepo interfaces.TripRepository) {
	places, err := tripRepo.GetPlaces(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Empty(t, places)
}

func tripRepository_GetPlaces_Success(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user := generateTestUser(uuid.Nil)
	err := userRepo.Save(ctx, user)
	require.NoError(t, err)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	_ = tripRepo.Save(ctx, trip)

	place1 := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place1)
	place2 := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place2)

	_ = tripRepo.AddPlace(ctx, trip.GetId(), place1.GetId(), 1, nil, 60, "")
	_ = tripRepo.AddPlace(ctx, trip.GetId(), place2.GetId(), 2, nil, 90, "")

	places, err := tripRepo.GetPlaces(ctx, trip.GetId())
	require.NoError(t, err)
	require.Len(t, places, 2)
}
func tripRepository_UpdateTripPlace_Success(t *testing.T,
	tripRepo interfaces.TripRepository,
	userRepo interfaces.UserRepository,
	countryRepo interfaces.CountryRepository,
	cityRepo interfaces.CityRepository,
	placeRepo interfaces.PlaceRepository) {

	ctx := context.Background()

	country := generateTestCountry(uuid.Nil)
	_ = countryRepo.Save(ctx, country)

	city := generateTestCity(uuid.Nil, country.GetId())
	_ = cityRepo.Save(ctx, city)

	user := generateTestUser(uuid.Nil)
	_ = userRepo.Save(ctx, user)

	trip := generateTestTrip(uuid.Nil, user.GetId())
	_ = tripRepo.Save(ctx, trip)

	place := generateTestPlace(uuid.Nil, city.GetId())
	_ = placeRepo.Save(ctx, place)

	_ = tripRepo.AddPlace(ctx, trip.GetId(), place.GetId(), 1, nil, 60, "")

	places, err := tripRepo.GetPlaces(ctx, trip.GetId())
	require.NoError(t, err)
	if len(places) == 0 {
		t.Fatal("no places found, cannot update")
	}
	tp := places[0]
	tp.SetNotes("updated notes")
	err = tripRepo.UpdateTripPlace(ctx, tp)
	require.NoError(t, err)

	updated, err := tripRepo.GetPlaces(ctx, trip.GetId())
	require.NoError(t, err)
	require.Len(t, updated, 1)
	require.Equal(t, "updated notes", updated[0].GetNotes())
}
