package e2e

import (
	"net/http"
	"testing"
)

func TestE2E_DemoScenario(t *testing.T) {
	env := startTestEnv(t)
	WithUser(t, &env, TestUser, TestPassword, func(c *TestClient) {
		t.Run("Step1_Login_Success", func(t *testing.T) {
			if c.token == "" {
				t.Fatal("token is empty after successful login")
			}
			t.Log(" User logged in successfully")
		})

		var russiaID string
		t.Run("Step2_GetCountries", func(t *testing.T) {
			resp, err := c.DoAPI("GET", "/api/v1/countries?limit=50&offset=0", nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			var countries []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			DecodeJSON(t, resp, &countries)
			if len(countries) == 0 {
				t.Fatal("expected at least one country")
			}
			for _, country := range countries {
				if country.Name == "Russia" {
					russiaID = country.ID
					break
				}
			}
			if russiaID == "" {
				t.Fatalf("Russia not found in countries list. Available: %+v", countries)
			}
			t.Logf(" Found %d countries, Russia ID: %s", len(countries), russiaID)
		})

		t.Run("Step3_GetCountry_Russia", func(t *testing.T) {
			resp, err := c.DoAPI("GET", "/api/v1/countries/"+russiaID, nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			var country struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				CapitalName string `json:"capitalName"`
				Currency    string `json:"currency"`
				Language    string `json:"language"`
			}
			DecodeJSON(t, resp, &country)
			if country.Name != "Russia" {
				t.Fatalf("expected country name 'Russia', got '%s'", country.Name)
			}
			if country.Currency != "RUB" {
				t.Fatalf("expected currency 'RUB', got '%s'", country.Currency)
			}
			t.Logf(" Russia details: currency=%s, language=%s, capital=%s",
				country.Currency, country.Language, country.CapitalName)
		})

		var moscowID string
		t.Run("Step4_GetCity_Moscow", func(t *testing.T) {
			resp, err := c.DoAPI("GET", "/api/v1/countries/"+russiaID+"/cities?limit=50", nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			var cities []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			DecodeJSON(t, resp, &cities)
			for _, city := range cities {
				if city.Name == "Moscow" {
					moscowID = city.ID
					break
				}
			}
			if moscowID == "" {
				t.Fatalf("Moscow not found in cities list. Available: %+v", cities)
			}
			resp, err = c.DoAPI("GET", "/api/v1/cities/"+moscowID, nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			var city struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Population  int64  `json:"population"`
				IsCapital   bool   `json:"isCapital"`
				Description string `json:"description"`
			}
			DecodeJSON(t, resp, &city)
			if city.Name != "Moscow" {
				t.Fatalf("expected city name 'Moscow', got '%s'", city.Name)
			}
			if !city.IsCapital {
				t.Fatal("Moscow should be marked as capital")
			}
			t.Logf(" Moscow details: population=%d, isCapital=%v, description=%s",
				city.Population, city.IsCapital, city.Description)
		})

		var redSquareID string
		t.Run("Step5_GetPlace_RedSquare", func(t *testing.T) {
			resp, err := c.DoAPI("GET", "/api/v1/places?cityId="+moscowID+"&limit=50", nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}

			var places []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			DecodeJSON(t, resp, &places)

			for _, place := range places {
				if place.Name == "Red Square" {
					redSquareID = place.ID
					break
				}
			}
			if redSquareID == "" {
				t.Fatalf("Red Square not found in places list. Available: %+v", places)
			}

			resp, err = c.DoAPI("GET", "/api/v1/places/"+redSquareID, nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}

			var place struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Category    string  `json:"category"`
				Description string  `json:"description"`
				AvgRating   float64 `json:"avgRating"`
				CityName    string  `json:"city_name"`
				CityID      string  `json:"city_id"`
			}
			DecodeJSON(t, resp, &place)

			if place.Name != "Red Square" {
				t.Fatalf("expected place name 'Red Square', got '%s'", place.Name)
			}

			if place.CityName != "Moscow" && place.CityID != moscowID {
				t.Logf(" Warning: expected city 'Moscow' or city_id '%s', got cityName='%s', cityId='%s'",
					moscowID, place.CityName, place.CityID)
			} else {
				t.Logf(" Red Square details: category=%s, rating=%.1f, city=%s",
					place.Category, place.AvgRating, place.CityName)
			}
		})

		t.Run("Step6_AddToFavourites", func(t *testing.T) {
			resp, err := c.DoAPI("POST", "/api/v1/favourites", map[string]string{
				"placeId": redSquareID,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 201/200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			resp.Body.Close()

			resp, err = c.DoAPI("GET", "/api/v1/favourites?limit=50", nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			var favourites []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			DecodeJSON(t, resp, &favourites)
			found := false
			for _, fav := range favourites {
				if fav.ID == redSquareID {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Red Square not found in favourites. Available: %+v", favourites)
			}
			t.Logf(" Red Square added to favourites (%d total)", len(favourites))
		})

		var tripID string
		t.Run("Step7_CreateTripAndAddPlace", func(t *testing.T) {
			resp, err := c.DoAPI("POST", "/api/v1/trips", map[string]interface{}{
				"title":     "E2E Test Trip to Moscow",
				"startDate": "2026-11-01",
				"endDate":   "2026-11-05",
				"budget":    50000.0,
				"notes":     "Demo trip created by E2E test",
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 201/200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			var tripResp struct {
				ID     string  `json:"id"`
				Title  string  `json:"title"`
				Budget float64 `json:"budget"`
				Status string  `json:"status"`
			}
			DecodeJSON(t, resp, &tripResp)
			if tripResp.ID == "" {
				t.Fatal("trip ID is empty")
			}
			if tripResp.Title != "E2E Test Trip to Moscow" {
				t.Fatalf("expected title 'E2E Test Trip to Moscow', got '%s'", tripResp.Title)
			}
			tripID = tripResp.ID
			t.Logf(" Trip created: id=%s, title=%s, status=%s", tripID, tripResp.Title, tripResp.Status)

			resp, err = c.DoAPI("POST", "/api/v1/trips/"+tripID+"/places", map[string]interface{}{
				"placeId":     redSquareID,
				"dayNumber":   1,
				"durationMin": 120,
				"notes":       "Must-see landmark",
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
				t.Fatalf("expected 201/200/204, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			resp.Body.Close()

			resp, err = c.DoAPI("GET", "/api/v1/trips/"+tripID, nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", resp.StatusCode, ReadBody(t, resp))
			}
			var tripDetails struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Places []struct {
					PlaceID   string `json:"placeId"`
					DayNumber int    `json:"dayNumber"`
					Notes     string `json:"notes"`
				} `json:"places"`
			}
			DecodeJSON(t, resp, &tripDetails)
			if len(tripDetails.Places) != 1 {
				t.Fatalf("expected 1 place in trip, got %d", len(tripDetails.Places))
			}
			if tripDetails.Places[0].PlaceID != redSquareID {
				t.Fatalf("expected place ID %s, got %s", redSquareID, tripDetails.Places[0].PlaceID)
			}
			if tripDetails.Places[0].DayNumber != 1 {
				t.Fatalf("expected day number 1, got %d", tripDetails.Places[0].DayNumber)
			}
			t.Logf(" Red Square added to trip '%s' (day %d, %d places total)",
				tripDetails.Title, tripDetails.Places[0].DayNumber, len(tripDetails.Places))
		})
	})
}
