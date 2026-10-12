package resolvers

import (
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/delivery/dto"
	gen "github.com/Neratus/geoguide/internal/delivery/gen/graphql"
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/google/uuid"
)

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func derefFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func floatPtr(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}

func parseTimeOrZero(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Time{}
}

func paginationOrDefault(p *gen.PaginationInput) (int, int) {
	limit, offset := 20, 0
	if p != nil {
		if p.Limit != nil {
			limit = *p.Limit
		}
		if p.Offset != nil {
			offset = *p.Offset
		}
	}
	return limit, offset
}
func userFromUpdateProfile(p *dto.UpdateUserProfileResponse) *gen.User {
	if p == nil {
		return nil
	}
	u := &gen.User{
		ID:                 uuid.UUID(p.ID),
		Username:           p.Username,
		Email:              p.Email,
		RegisteredAt:       parseTimeOrZero(p.RegisteredAt),
		FavoriteCategories: p.FavoriteCategories,
		IsBlocked:          p.IsBlocked,
	}
	if p.Phone != "" {
		u.Phone = &p.Phone
	}
	if p.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", p.BirthDate); err == nil {
			u.BirthDate = &t
		}
	}
	if p.CountryOfResidence != "" {
		u.CountryOfResidence = &p.CountryOfResidence
	}
	if p.AvatarURL != "" {
		u.AvatarURL = &p.AvatarURL
	}
	if u.FavoriteCategories == nil {
		u.FavoriteCategories = []string{}
	}
	return u
}

func userFromProfile(p *dto.GetUserProfileResponse) *gen.User {
	u := &gen.User{
		ID:                 uuid.UUID(p.ID),
		Username:           p.Username,
		Email:              p.Email,
		RegisteredAt:       parseTimeOrZero(p.RegisteredAt),
		FavoriteCategories: p.FavoriteCategories,
		IsBlocked:          p.IsBlocked,
		Role:               gen.UserRole(stringsToUpper(p.Role)),
	}
	if p.Phone != "" {
		u.Phone = &p.Phone
	}
	if p.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", p.BirthDate); err == nil {
			u.BirthDate = &t
		}
	}
	if p.CountryOfResidence != "" {
		u.CountryOfResidence = &p.CountryOfResidence
	}
	if p.AvatarURL != "" {
		u.AvatarURL = &p.AvatarURL
	}
	if u.FavoriteCategories == nil {
		u.FavoriteCategories = []string{}
	}
	return u
}

func stringsToUpper(s string) string {
	switch s {
	case "user":
		return "USER"
	case "moderator":
		return "MODERATOR"
	case "analyst":
		return "ANALYST"
	}
	return s
}

func countrySummary(c dto.CountryResponse) *gen.Country {
	out := &gen.Country{
		ID:   uuid.UUID(c.ID),
		Name: c.Name,
	}
	if c.Capital != "" {
		out.Capital = &c.Capital
	}
	if c.Area != 0 {
		out.Area = &c.Area
	}
	if c.Population != 0 {
		pop := domain.Long(c.Population)
		out.Population = &pop
	}
	if c.Currency != "" {
		out.Currency = &c.Currency
	}
	if c.Language != "" {
		out.Language = &c.Language
	}
	if c.PhoneCode != "" {
		out.PhoneCode = &c.PhoneCode
	}
	if c.ImageURL != "" {
		out.ImageURL = &c.ImageURL
	}
	return out
}

func countryDetail(c dto.GetCountryResponse) *gen.CountryDetails {
	out := &gen.CountryDetails{
		ID:       uuid.UUID(c.ID),
		Name:     c.Name,
		Holidays: []*gen.Holiday{},
	}
	if c.Area != 0 {
		out.Area = &c.Area
	}
	if c.Population != 0 {
		pop := domain.Long(c.Population)
		out.Population = &pop
	}
	if c.GDP != 0 {
		out.Gdp = &c.GDP
	}
	if c.Currency != "" {
		out.Currency = &c.Currency
	}
	if c.VisaRequirements != "" {
		out.VisaRequirements = &c.VisaRequirements
	}
	if c.Description != "" {
		out.Description = &c.Description
	}
	if c.SafetyTips != "" {
		out.SafetyTips = &c.SafetyTips
	}
	if c.BestSeason != "" {
		out.BestSeason = &c.BestSeason
	}
	if c.Language != "" {
		out.Language = &c.Language
	}
	if c.PhoneCode != "" {
		out.PhoneCode = &c.PhoneCode
	}
	if c.Religion != "" {
		out.Religion = &c.Religion
	}
	if c.CapitalName != "" {
		out.CapitalName = &c.CapitalName
	}
	if c.CapitalID != (uuid.Nil) {
		capID := uuid.UUID(c.CapitalID)
		out.CapitalID = &capID
	}
	if c.ImageURL != "" {
		out.ImageURL = &c.ImageURL
	}
	for _, h := range c.Holidays {
		hol := &gen.Holiday{
			ID:         uuid.UUID(h.ID),
			Name:       h.Name,
			Date:       h.Date,
			IsNational: h.IsNational,
		}
		if h.Description != "" {
			hol.Description = &h.Description
		}
		out.Holidays = append(out.Holidays, hol)
	}
	return out
}

func citySummary(c dto.CityResponse) *gen.City {
	out := &gen.City{
		ID:          uuid.UUID(c.ID),
		Name:        c.Name,
		IsCapital:   c.IsCapital,
		Coordinates: parseCoords(c.Coordinates),
	}
	if c.Population != 0 {
		pop := domain.Long(c.Population)
		out.Population = &pop
	}
	if c.Description != "" {
		out.Description = &c.Description
	}
	if c.Timezone != "" {
		out.Timezone = &c.Timezone
	}
	if c.ImageURL != "" {
		out.ImageURL = &c.ImageURL
	}
	return out
}

func cityDetail(c dto.GetCityResponse) *gen.CityDetails {
	out := &gen.CityDetails{
		ID:          uuid.UUID(c.ID),
		Name:        c.Name,
		IsCapital:   c.IsCapital,
		Coordinates: coordsToGraphQL(c.Coordinates),
		CountryID:   uuid.UUID(c.CountryID),
	}
	if c.Population != 0 {
		pop := domain.Long(c.Population)
		out.Population = &pop
	}
	if c.Description != "" {
		out.Description = &c.Description
	}
	if c.Timezone != "" {
		out.Timezone = &c.Timezone
	}
	if c.TravelTips != "" {
		out.TravelTips = &c.TravelTips
	}
	if c.ImageID != (uuid.Nil) {
		imgID := uuid.UUID(c.ImageID)
		out.ImageID = &imgID
	}
	if c.ImageURL != "" {
		out.ImageURL = &c.ImageURL
	}
	return out
}

func districtFromResponse(d dto.DistrictResponse) *gen.District {
	out := &gen.District{
		ID:          uuid.UUID(d.ID),
		Name:        d.Name,
		Coordinates: coordsToGraphQL(d.Coordinates),
		CityID:      uuid.UUID(d.CityID),
	}
	if d.Description != "" {
		out.Description = &d.Description
	}
	if d.ImageID != (uuid.Nil) {
		imgID := uuid.UUID(d.ImageID)
		out.ImageID = &imgID
	}
	return out
}

func transportNodeFromResponse(n dto.TransportNodeResponse) *gen.TransportNode {
	out := &gen.TransportNode{
		ID:          uuid.UUID(n.ID),
		Name:        n.Name,
		NodeType:    gen.NodeType(n.NodeType),
		Coordinates: coordsToGraphQL(n.Coordinates),
		CityID:      uuid.UUID(n.CityID),
	}
	if n.Address != "" {
		out.Address = &n.Address
	}
	if n.ImageID != (uuid.Nil) {
		imgID := uuid.UUID(n.ImageID)
		out.ImageID = &imgID
	}
	return out
}

func placesFromResponse(items []dto.PlaceResponse) []*gen.Place {
	out := make([]*gen.Place, 0, len(items))
	for _, p := range items {
		out = append(out, placeFromSummary(p))
	}
	return out
}

func placeFromSummary(p dto.PlaceResponse) *gen.Place {
	out := &gen.Place{
		ID:           uuid.UUID(p.ID),
		Name:         p.Name,
		Category:     p.Category,
		ReviewsCount: int(p.ReviewsCount),
		CityID:       uuid.UUID(p.CityID),
		Reviews:      []*gen.ReviewBrief{},
	}
	if p.Description != "" {
		out.Description = &p.Description
	}
	if p.Coordinates != "" {
		out.Coordinates = &p.Coordinates
	}
	if p.Address != "" {
		out.Address = &p.Address
	}
	if p.OpeningHours != "" {
		out.OpeningHours = &p.OpeningHours
	}
	if p.PriceInfo != "" {
		out.PriceInfo = &p.PriceInfo
	}
	if p.AvgVisitDurationMin != 0 {
		v := int(p.AvgVisitDurationMin)
		out.AvgVisitDurationMin = &v
	}
	if p.AvgRating != 0 {
		v := float64(p.AvgRating)
		out.AvgRating = &v
	}
	if p.ContactPhone != "" {
		out.ContactPhone = &p.ContactPhone
	}
	if p.Website != "" {
		out.Website = &p.Website
	}
	if p.CityName != "" {
		out.CityName = &p.CityName
	}
	if p.DistrictID != (uuid.Nil) {
		d := uuid.UUID(p.DistrictID)
		out.DistrictID = &d
	}
	if p.DistrictName != "" {
		out.DistrictName = &p.DistrictName
	}
	if p.ImageURL != "" {
		out.ImageURL = &p.ImageURL
	}

	for _, rv := range p.Reviews {
		out.Reviews = append(out.Reviews, &gen.ReviewBrief{
			ID:         uuid.UUID(rv.ID),
			Rating:     rv.Rating,
			Comment:    rv.Comment,
			VisitDate:  parseTimeOrZero(rv.VisitDate),
			CreatedAt:  parseTimeOrZero(rv.CreatedAt),
			Username:   rv.Username,
			IsApproved: rv.IsApproved,
		})
	}

	return out
}

func tripSummaryFromCreate(t dto.CreateTripResponse) *gen.Trip {
	out := &gen.Trip{
		ID:        uuid.UUID(t.ID),
		Title:     t.Title,
		StartDate: parseTimeOrZero(t.StartDate),
		EndDate:   parseTimeOrZero(t.EndDate),
		Places:    []*gen.TripPlace{},
	}
	if t.Budget != 0 {
		out.Budget = &t.Budget
	}
	if t.Status != "" {
		out.Status = &t.Status
	}
	return out
}

func tripFromDetail(t dto.GetTripResponse) *gen.Trip {
	out := &gen.Trip{
		ID:        uuid.UUID(t.ID),
		Title:     t.Title,
		StartDate: parseTimeOrZero(t.StartDate),
		EndDate:   parseTimeOrZero(t.EndDate),
		Places:    []*gen.TripPlace{},
	}
	if t.Budget != 0 {
		out.Budget = &t.Budget
	}
	if t.Status != "" {
		out.Status = &t.Status
	}
	if t.Notes != "" {
		out.Notes = &t.Notes
	}
	for _, p := range t.Places {
		tp := &gen.TripPlace{
			PlaceID:   uuid.UUID(p.PlaceID),
			DayNumber: p.DayNumber,
		}
		if p.ArrivalTime != "" {
			if t, err := time.Parse(time.RFC3339, p.ArrivalTime); err == nil {
				tp.ArrivalTime = &t
			}
		}
		if p.DurationMin != 0 {
			d := int(p.DurationMin)
			tp.DurationMin = &d
		}
		if p.Notes != "" {
			tp.Notes = &p.Notes
		}
		if p.VisitStatus != "" {
			tp.VisitStatus = &p.VisitStatus
		}
		if p.ActualCost != 0 {
			tp.ActualCost = &p.ActualCost
		}
		out.Places = append(out.Places, tp)
	}
	return out
}

func reviewFromCreateResponse(r dto.CreateReviewResponse) *gen.Review {
	out := &gen.Review{
		ID:          uuid.UUID(r.ID),
		Rating:      r.Rating,
		Comment:     r.Comment,
		VisitDate:   parseTimeOrZero(r.VisitDate),
		CreatedAt:   parseTimeOrZero(r.CreatedAt),
		IsApproved:  r.IsApproved,
		IsModerated: r.IsModerated,
	}
	return out
}

func parseCoords(s string) *gen.Coordinates {
	if s == "" {
		return &gen.Coordinates{}
	}
	var lat, lng float64
	if _, err := fmt.Sscan(s, &lat, &lng); err != nil {
		return &gen.Coordinates{}
	}
	return &gen.Coordinates{Lat: lat, Lng: lng}
}

func coordsToGraphQL(c domain.Coordinates) *gen.Coordinates {
	return &gen.Coordinates{Lat: c.Lat(), Lng: c.Lng()}
}
