package resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	gen "github.com/Neratus/geoguide/internal/delivery/gen/graphql"
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/requests"
	"github.com/Neratus/geoguide/internal/usecase/content"
	"github.com/Neratus/geoguide/internal/usecase/favourites"
	"github.com/google/uuid"
)

// =========================================================
//                       MUTATION
// =========================================================

// RegisterUser is the resolver for the registerUser field.
func (r *mutationResolver) RegisterUser(ctx context.Context, input gen.RegisterUserInput) (*gen.User, error) {
	req := requests.RegisterUserRequest{
		Username:           input.Username,
		Email:              input.Email,
		Password:           input.Password,
		Phone:              derefString(input.Phone),
		BirthDate:          input.BirthDate,
		CountryOfResidence: derefString(input.CountryOfResidence),
	}
	resp, err := r.RegisterUC.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	// После регистрации тянем полный профиль — он уже есть в БД.
	profile, err := r.GetProfileUC.Execute(ctx, requests.GetUserProfileRequest{UserID: resp.ID})
	if err != nil {
		// fallback — минимальный User, если профиль почему-то недоступен
		return &gen.User{
			ID:                 uuid.UUID(resp.ID),
			Username:           resp.Username,
			Email:              resp.Email,
			RegisteredAt:       time.Now(),
			FavoriteCategories: []string{},
		}, nil
	}
	return userFromProfile(profile), nil
}

// VerifyContact is the resolver for the verifyContact field.
func (r *mutationResolver) VerifyContact(ctx context.Context, input gen.VerifyContactInput) (*gen.OperationResult, error) {
	err := r.VerifyContactUC.Execute(ctx, requests.VerifyContactRequest{
		UserID:  domain.UserID(input.UserID),
		Contact: input.Contact,
		Code:    input.Code,
	})
	if err != nil {
		return &gen.OperationResult{
			Success: false,
			Error:   &gen.Error{Message: err.Error()},
		}, nil
	}
	return &gen.OperationResult{Success: true}, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input gen.LoginInput) (*gen.LoginPayload, error) {
	res, err := r.LoginUC.Execute(ctx, requests.LoginUserRequest{
		Username: input.Username,
		Password: input.Password,
	}, "", "")
	if err != nil {
		return nil, err
	}

	payload := &gen.LoginPayload{
		RequiresTwoFactor: res.RequiresTwoFactor,
	}
	if res.Token != "" {
		tok := res.Token
		payload.Token = &tok
	}
	if res.RequiresTwoFactor {
		ch := res.ChallengeID
		payload.ChallengeID = &ch
		return payload, nil
	}

	// Подтянем профиль, чтобы вернуть User-объект
	if profile, err := r.GetProfileUC.Execute(ctx, requests.GetUserProfileRequest{UserID: res.ID}); err == nil {
		payload.User = userFromProfile(profile)
	} else {
		id := uuid.UUID(res.ID)
		payload.User = &gen.User{
			ID:                 id,
			Username:           res.Username,
			Email:              res.Email,
			RegisteredAt:       time.Now(),
			FavoriteCategories: []string{},
		}
	}
	return payload, nil
}

// Logout is the resolver for the logout field.
func (r *mutationResolver) Logout(ctx context.Context) (*gen.OperationResult, error) {
	token := CurrentRawToken(ctx)
	if token == "" {
		return &gen.OperationResult{
			Success: false,
			Error:   &gen.Error{Message: "missing token"},
		}, nil
	}
	if _, err := r.LogoutUC.Execute(ctx, requests.LogoutRequest{Token: token}); err != nil {
		return &gen.OperationResult{
			Success: false,
			Error:   &gen.Error{Message: err.Error()},
		}, nil
	}
	return &gen.OperationResult{Success: true}, nil
}

// EnableTwoFactor is the resolver for the enableTwoFactor field.
func (r *mutationResolver) EnableTwoFactor(ctx context.Context, input gen.EnableTwoFactorInput) (*gen.TwoFactorSetup, error) {
	codes, err := r.Enable2FAUC.VerifyAndEnable(ctx, requests.EnableTwoFactorRequest{
		UserID: domain.UserID(input.UserID),
		Secret: input.Secret,
		Code:   input.Code,
	})
	if err != nil {
		return nil, err
	}
	return &gen.TwoFactorSetup{
		BackupCodes: codes,
		OtpauthURI:  "", // заполним ниже, если есть GenerateSecret-подход
	}, nil
}

// VerifyTwoFactor is the resolver for the verifyTwoFactor field.
func (r *mutationResolver) VerifyTwoFactor(ctx context.Context, input gen.VerifyTwoFactorInput) (*gen.AuthToken, error) {
	tok, err := r.Verify2FAUC.Execute(ctx, requests.VerifyTwoFactorRequest{
		ChallengeID: input.ChallengeID,
		Code:        input.Code,
	})
	if err != nil {
		return nil, err
	}
	return &gen.AuthToken{Token: tok}, nil
}

func (r *mutationResolver) UpdateUserProfile(ctx context.Context, input gen.UpdateUserProfileInput) (*gen.User, error) {
	if _, err := CurrentUserID(ctx); err != nil {
		return nil, err
	}
	rawToken := CurrentRawToken(ctx)
	if rawToken == "" {
		return nil, fmt.Errorf("missing token")
	}

	req := requests.UpdateUserProfileRequest{Token: rawToken}
	if input.Username != nil {
		req.Username = input.Username
	}
	if input.Phone != nil {
		req.Phone = input.Phone
	}
	if input.BirthDate != nil {
		req.BirthDate = input.BirthDate
	}
	if input.CountryOfResidence != nil {
		req.CountryOfResidence = input.CountryOfResidence
	}
	if input.AvatarURL != nil {
		req.AvatarURL = input.AvatarURL
	}
	if input.FavoriteCategories != nil {
		req.FavoriteCategories = input.FavoriteCategories
	}

	updated, err := r.UpdateProfileUC.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	return userFromUpdateProfile(updated), nil
}

// ModerateUser is the resolver for the moderateUser field.
func (r *mutationResolver) ModerateUser(ctx context.Context, input gen.ModerateUserInput) (*gen.User, error) {
	modID, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	res, err := r.ModerateUserUC.Execute(ctx, requests.ModerateUserRequest{
		UserID:      domain.UserID(input.UserID),
		Block:       input.Block,
		Reason:      derefString(input.Reason),
		ModeratorID: domain.UserID(modID),
	})
	if err != nil {
		return nil, err
	}
	// Вернём минимальный User с актуальными полями блокировки.
	return &gen.User{
		ID:        input.UserID,
		Username:  "",
		Email:     "",
		IsBlocked: res.IsBlocked,
	}, nil
}

// ModerateReview is the resolver for the moderateReview field.
func (r *mutationResolver) ModerateReview(ctx context.Context, input gen.ModerateReviewInput) (*gen.Review, error) {
	modID, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	res, err := r.ModerateReviewUC.Execute(ctx, requests.ModerateReviewRequest{
		ReviewID:    domain.ReviewID(input.ReviewID),
		Approved:    input.Approved,
		Comment:     derefString(input.Comment),
		ModeratorID: domain.UserID(modID),
	})
	if err != nil {
		return nil, err
	}
	// Минимальный Review с обновлёнными полями.
	return &gen.Review{
		ID:          uuid.UUID(res.ID),
		IsApproved:  res.IsApproved,
		IsModerated: res.IsModerated,
	}, nil
}

// CreateTrip is the resolver for the createTrip field.
func (r *mutationResolver) CreateTrip(ctx context.Context, input gen.CreateTripInput) (*gen.Trip, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	res, err := r.CreateTripUC.Execute(ctx, requests.CreateTripRequest{
		Title:     input.Title,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Budget:    derefFloat64(input.Budget),
		UserID:    domain.UserID(uid),
		Notes:     derefString(input.Notes),
	})
	if err != nil {
		return nil, err
	}
	return tripSummaryFromCreate(*res), nil
}

// UpdateTrip is the resolver for the updateTrip field.
func (r *mutationResolver) UpdateTrip(ctx context.Context, id uuid.UUID, input gen.UpdateTripInput) (*gen.Trip, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	req := requests.UpdateTripRequest{
		TripID:    domain.TripID(id),
		UserID:    domain.UserID(uid),
		Title:     input.Title,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Budget:    input.Budget,
		Status:    input.Status,
		Notes:     input.Notes,
	}
	if err := r.UpdateTripUC.Execute(ctx, req); err != nil {
		return nil, err
	}
	// Подтянем актуальный Trip
	trip, err := r.GetTripUC.Execute(ctx, requests.GetTripRequest{
		TripID: domain.TripID(id),
		UserID: domain.UserID(uid),
	})
	if err != nil {
		return &gen.Trip{ID: id}, nil
	}
	return tripFromDetail(*trip), nil
}

// DeleteTrip is the resolver for the deleteTrip field.
func (r *mutationResolver) DeleteTrip(ctx context.Context, id uuid.UUID) (*gen.OperationResult, error) {
	if err := r.DeleteTripUC.Execute(ctx, requests.DeleteTripRequest{TripID: domain.TripID(id)}); err != nil {
		return &gen.OperationResult{
			Success: false,
			Error:   &gen.Error{Message: err.Error()},
		}, nil
	}
	return &gen.OperationResult{Success: true}, nil
}

// AddPlaceToTrip is the resolver for the addPlaceToTrip field.
func (r *mutationResolver) AddPlaceToTrip(ctx context.Context, tripID uuid.UUID, input gen.AddPlaceToTripInput) (*gen.TripPlace, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	req := requests.AddPlaceToTripRequest{
		TripID:      domain.TripID(tripID),
		PlaceID:     domain.PlaceID(input.PlaceID),
		DayNumber:   input.DayNumber,
		ArrivalTime: input.ArrivalTime,
		DurationMin: derefInt(input.DurationMin),
		Notes:       derefString(input.Notes),
		UserID:      domain.UserID(uid),
	}
	if err := r.AddPlaceToTripUC.Execute(ctx, req); err != nil {
		return nil, err
	}
	return &gen.TripPlace{
		PlaceID:     input.PlaceID,
		DayNumber:   input.DayNumber,
		ArrivalTime: input.ArrivalTime,
		DurationMin: input.DurationMin,
		Notes:       input.Notes,
	}, nil
}

// RemovePlaceFromTrip is the resolver for the removePlaceFromTrip field.
func (r *mutationResolver) RemovePlaceFromTrip(ctx context.Context, tripID uuid.UUID, placeID uuid.UUID) (*gen.OperationResult, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	req := requests.RemovePlaceFromTripRequest{
		TripID:  domain.TripID(tripID),
		PlaceID: domain.PlaceID(placeID),
		UserID:  domain.UserID(uid),
	}
	if err := r.RemovePlaceFromTripUC.Execute(ctx, req); err != nil {
		return &gen.OperationResult{
			Success: false,
			Error:   &gen.Error{Message: err.Error()},
		}, nil
	}
	return &gen.OperationResult{Success: true}, nil
}

// CreateReview is the resolver for the createReview field.
func (r *mutationResolver) CreateReview(ctx context.Context, input gen.CreateReviewInput) (*gen.Review, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	res, err := r.CreateReviewUC.Execute(ctx, requests.CreateReviewRequest{
		UserID:    domain.UserID(uid),
		PlaceID:   domain.PlaceID(input.PlaceID),
		Rating:    input.Rating,
		Comment:   input.Comment,
		VisitDate: input.VisitDate,
		ImageURL:  derefString(input.ImageURL),
	})
	if err != nil {
		return nil, err
	}
	return reviewFromCreateResponse(*res), nil
}

// DeleteReview is the resolver for the deleteReview field.
func (r *mutationResolver) DeleteReview(ctx context.Context, id uuid.UUID) (*gen.OperationResult, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	isMod := IsModerator(ctx)
	res, err := r.DeleteReviewUC.Execute(ctx, requests.DeleteReviewRequest{
		ReviewID:    domain.ReviewID(id),
		UserID:      domain.UserID(uid),
		IsModerator: isMod,
	})
	if err != nil {
		return &gen.OperationResult{
			Success: false,
			Error:   &gen.Error{Message: err.Error()},
		}, nil
	}
	return &gen.OperationResult{Success: res.Success}, nil
}

// AddFavourite is the resolver for the addFavourite field.
func (r *mutationResolver) AddFavourite(ctx context.Context, placeID uuid.UUID) (*gen.OperationResult, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := r.AddFavUC.Execute(ctx, favourites.AddFavouriteRequest{
		UserID:  domain.UserID(uid),
		PlaceID: domain.PlaceID(placeID),
	}); err != nil {
		return &gen.OperationResult{
			Success: false,
			Error:   &gen.Error{Message: err.Error()},
		}, nil
	}
	return &gen.OperationResult{Success: true}, nil
}

// RemoveFavourite is the resolver for the removeFavourite field.
func (r *mutationResolver) RemoveFavourite(ctx context.Context, placeID uuid.UUID) (*gen.OperationResult, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := r.RemoveFavUC.Execute(ctx, favourites.RemoveFavouriteRequest{
		UserID:  domain.UserID(uid),
		PlaceID: domain.PlaceID(placeID),
	}); err != nil {
		return &gen.OperationResult{
			Success: false,
			Error:   &gen.Error{Message: err.Error()},
		}, nil
	}
	return &gen.OperationResult{Success: true}, nil
}

// GenerateReport is the resolver for the generateReport field.
func (r *mutationResolver) GenerateReport(ctx context.Context, input gen.GenerateReportInput) (*gen.Report, error) {
	// Синхронно выполняем UC и возвращаем готовый Report.
	req := requests.GenerateReportRequest{
		ReportType: string(input.ReportType),
		Format:     domain.ReportFormat(strings.ToLower(string(input.Format))),
		Options: domain.ReportOptions{
			IncludeHeaders: true,
			PrettyPrint:    input.Format == gen.ReportFormatJSON,
		},
		Limit: derefInt(input.Limit),
	}
	if input.DateFrom != nil {
		req.DateFrom = input.DateFrom.Format("2006-01-02")
	}
	if input.DateTo != nil {
		req.DateTo = input.DateTo.Format("2006-01-02")
	}

	if _, err := r.GenerateReportUC.Execute(ctx, req); err != nil {
		return nil, err
	}
	// Заглушка: возвращаем готовый отчёт.
	now := time.Now()
	return &gen.Report{
		ID:          uuid.New(),
		ReportType:  input.ReportType,
		Format:      input.Format,
		Status:      gen.ReportStatusCompleted,
		Progress:    100,
		CreatedAt:   now,
		CompletedAt: &now,
	}, nil
}

// =========================================================
//                        QUERY
// =========================================================

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*gen.User, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	profile, err := r.GetProfileUC.Execute(ctx, requests.GetUserProfileRequest{UserID: domain.UserID(uid)})
	if err != nil {
		return nil, err
	}
	return userFromProfile(profile), nil
}

// Countries is the resolver for the countries field.
func (r *queryResolver) Countries(ctx context.Context, search *string, pagination *gen.PaginationInput) ([]*gen.Country, error) {
	limit, offset := paginationOrDefault(pagination)
	items, err := r.GetCountriesUC.Execute(ctx, requests.GetCountriesRequest{
		Limit:  limit,
		Offset: offset,
		Search: derefString(search),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.Country, 0, len(items))
	for _, c := range items {
		out = append(out, countrySummary(c))
	}
	return out, nil
}

// Country is the resolver for the country field.
func (r *queryResolver) Country(ctx context.Context, id uuid.UUID) (*gen.CountryDetails, error) {
	c, err := r.GetCountryUC.Execute(ctx, requests.GetCountryRequest{ID: domain.CountryID(id)})
	if err != nil {
		return nil, err
	}
	return countryDetail(*c), nil
}

// CitiesByCountry is the resolver for the citiesByCountry field.
func (r *queryResolver) CitiesByCountry(ctx context.Context, countryID uuid.UUID, pagination *gen.PaginationInput) ([]*gen.City, error) {
	limit, offset := paginationOrDefault(pagination)
	items, err := r.GetCitiesUC.Execute(ctx, requests.GetCitiesByCountryRequest{
		CountryID: domain.CountryID(countryID),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.City, 0, len(items))
	for _, c := range items {
		out = append(out, citySummary(c))
	}
	return out, nil
}

// City is the resolver for the city field.
func (r *queryResolver) City(ctx context.Context, id uuid.UUID) (*gen.CityDetails, error) {
	c, err := r.GetCityUC.Execute(ctx, requests.GetCityRequest{ID: domain.CityID(id)})
	if err != nil {
		return nil, err
	}
	return cityDetail(*c), nil
}

// Cities is the resolver for the cities field.
func (r *queryResolver) Cities(ctx context.Context, search *string, pagination *gen.PaginationInput) ([]*gen.City, error) {
	limit, offset := paginationOrDefault(pagination)
	items, err := r.SearchCitiesUC.Execute(ctx, requests.SearchCitiesRequest{
		Query:  derefString(search),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.City, 0, len(items))
	for _, c := range items {
		out = append(out, citySummary(c))
	}
	return out, nil
}

// DistrictsByCity is the resolver for the districtsByCity field.
func (r *queryResolver) DistrictsByCity(ctx context.Context, cityID uuid.UUID, pagination *gen.PaginationInput) ([]*gen.District, error) {
	limit, offset := paginationOrDefault(pagination)
	items, err := r.GetDistrictsUC.Execute(ctx, requests.GetDistrictsByCityRequest{
		CityID: domain.CityID(cityID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.District, 0, len(items))
	for _, d := range items {
		out = append(out, districtFromResponse(d))
	}
	return out, nil
}

// TransportsByCity is the resolver for the transportsByCity field.
func (r *queryResolver) TransportsByCity(ctx context.Context, cityID uuid.UUID, pagination *gen.PaginationInput) ([]*gen.TransportNode, error) {
	limit, offset := paginationOrDefault(pagination)
	items, err := r.GetTransportNodesUC.Execute(ctx, requests.GetTransportNodesByCityRequest{
		CityID: domain.CityID(cityID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.TransportNode, 0, len(items))
	for _, n := range items {
		out = append(out, transportNodeFromResponse(n))
	}
	return out, nil
}

// PlacesByCity is the resolver for the placesByCity field.
func (r *queryResolver) PlacesByCity(ctx context.Context, cityID uuid.UUID, category *string, sortBy *gen.PlaceSortBy, pagination *gen.PaginationInput) ([]*gen.Place, error) {
	limit, offset := paginationOrDefault(pagination)
	sort := ""
	if sortBy != nil {
		sort = strings.ToLower(string(*sortBy))
	}
	items, err := r.GetPlacesByCityUC.Execute(ctx, requests.GetPlacesByCityRequest{
		CityID:   domain.CityID(cityID),
		Category: derefString(category),
		SortBy:   sort,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	return placesFromResponse(items), nil
}

// PlacesByCategory is the resolver for the placesByCategory field.
func (r *queryResolver) PlacesByCategory(ctx context.Context, category string, pagination *gen.PaginationInput) ([]*gen.Place, error) {
	limit, offset := paginationOrDefault(pagination)
	items, err := r.GetPlacesByCategoryUC.Execute(ctx, requests.GetPlacesByCategoryRequest{
		Category: category,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	return placesFromResponse(items), nil
}

// Place is the resolver for the place field.
func (r *queryResolver) Place(ctx context.Context, id uuid.UUID) (*gen.Place, error) {
	p, err := r.GetPlaceUC.Execute(ctx, requests.GetPlaceRequest{ID: domain.PlaceID(id)})
	if err != nil {
		return nil, err
	}
	return placeFromSummary(*p), nil
}

// MyTrips is the resolver for the myTrips field.
func (r *queryResolver) MyTrips(ctx context.Context) ([]*gen.TripSummary, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := r.GetUserTripsUC.Execute(ctx, requests.GetUserTripsRequest{UserID: domain.UserID(uid)})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.TripSummary, 0, len(items))
	for _, t := range items {
		out = append(out, &gen.TripSummary{
			ID:        uuid.UUID(t.ID),
			Title:     t.Title,
			StartDate: parseTimeOrZero(t.StartDate),
			EndDate:   parseTimeOrZero(t.EndDate),
			Status:    stringPtr(t.Status),
		})
	}
	return out, nil
}

// Trip is the resolver for the trip field.
func (r *queryResolver) Trip(ctx context.Context, id uuid.UUID) (*gen.Trip, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	t, err := r.GetTripUC.Execute(ctx, requests.GetTripRequest{
		TripID: domain.TripID(id),
		UserID: domain.UserID(uid),
	})
	if err != nil {
		return nil, err
	}
	return tripFromDetail(*t), nil
}

// HolidaysByDate is the resolver for the holidaysByDate field.
func (r *queryResolver) HolidaysByDate(ctx context.Context, date time.Time) ([]*gen.Holiday, error) {
	items, err := r.GetHolidaysUC.Execute(ctx, requests.GetHolidaysByDateRequest{Date: date})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.Holiday, 0, len(items))
	for _, h := range items {
		out = append(out, &gen.Holiday{
			ID:          uuid.UUID(h.ID),
			Name:        h.Name,
			Date:        h.Date,
			Description: stringPtr(h.Description),
			IsNational:  h.IsNational,
		})
	}
	return out, nil
}

// File is the resolver for the file field.
func (r *queryResolver) File(ctx context.Context, id uuid.UUID) (*gen.StaticPage, error) {
	p, err := r.GetStaticPageUC.Execute(ctx, content.GetStaticPageByIDRequest{ID: domain.StaticPageID(id)})
	if err != nil {
		return nil, err
	}
	updated := time.Now()
	var published *time.Time
	if p.PublishedAt != "" {
		now := time.Now()
		published = &now
	}
	return &gen.StaticPage{
		ID:              uuid.UUID(p.ID),
		Slug:            p.Slug,
		Title:           p.Title,
		Content:         p.Content,
		MetaDescription: stringPtr(p.MetaDesc),
		ImageURL:        stringPtr(p.ImageURL),
		ImageAlt:        stringPtr(p.ImageAlt),
		UpdatedAt:       updated,
		PublishedAt:     published,
		IsPublished:     p.IsPublished,
	}, nil
}

// ReviewsByPlace is the resolver for the reviewsByPlace field.
func (r *queryResolver) ReviewsByPlace(ctx context.Context, placeID uuid.UUID, onlyApproved *bool, pagination *gen.PaginationInput) ([]*gen.ReviewBrief, error) {
	limit, offset := paginationOrDefault(pagination)
	only := true
	if onlyApproved != nil {
		only = *onlyApproved
	}
	items, err := r.GetReviewsByPlaceUC.Execute(ctx, requests.GetReviewsByPlaceRequest{
		PlaceID:      domain.PlaceID(placeID),
		OnlyApproved: only,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.ReviewBrief, 0, len(items))
	for _, rv := range items {
		out = append(out, &gen.ReviewBrief{
			ID:         uuid.UUID(rv.ID),
			Rating:     rv.Rating,
			Comment:    rv.Comment,
			VisitDate:  parseTimeOrZero(rv.VisitDate),
			CreatedAt:  parseTimeOrZero(rv.CreatedAt),
			Username:   rv.Username,
			IsApproved: rv.IsApproved,
		})
	}
	return out, nil
}

// MyReviews is the resolver for the myReviews field.
func (r *queryResolver) MyReviews(ctx context.Context, pagination *gen.PaginationInput) ([]*gen.Review, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset := paginationOrDefault(pagination)
	items, err := r.GetUserReviewsUC.Execute(ctx, requests.GetUserReviewsRequest{
		UserID: domain.UserID(uid),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.Review, 0, len(items))
	for _, rv := range items {
		out = append(out, &gen.Review{
			ID:          uuid.UUID(rv.ID),
			Rating:      rv.Rating,
			Comment:     rv.Comment,
			VisitDate:   parseTimeOrZero(rv.VisitDate),
			CreatedAt:   parseTimeOrZero(rv.CreatedAt),
			PlaceID:     uuid.UUID(rv.PlaceID),
			PlaceName:   stringPtr(rv.PlaceName),
			IsApproved:  rv.IsApproved,
			IsModerated: rv.IsModerated,
		})
	}
	return out, nil
}

// ModerationUsers is the resolver for the moderationUsers field.
func (r *queryResolver) ModerationUsers(ctx context.Context, search *string, pagination *gen.PaginationInput) ([]*gen.User, error) {
	if !IsModerator(ctx) {
		return nil, fmt.Errorf("forbidden")
	}
	limit, offset := paginationOrDefault(pagination)
	resp, err := r.GetUsersUC.Execute(ctx, requests.GetUsersRequest{
		Limit:       limit,
		Offset:      offset,
		Search:      derefString(search),
		IsModerator: true,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.User, 0, len(resp.Users))
	for _, u := range resp.Users {
		out = append(out, &gen.User{
			ID:                 u.GetId(),
			Username:           u.GetUsername(),
			Email:              u.GetEmail(),
			RegisteredAt:       u.GetRegisteredAt(),
			IsBlocked:          u.IsBlocked(),
			Role:               gen.UserRole(strings.ToUpper(u.GetRole())),
			FavoriteCategories: []string{},
		})
	}
	return out, nil
}

// ModerationPendingReviews is the resolver for the moderationPendingReviews field.
func (r *queryResolver) ModerationPendingReviews(ctx context.Context, isApproved *bool, pagination *gen.PaginationInput) ([]*gen.ReviewBrief, error) {
	if !IsModerator(ctx) {
		return nil, fmt.Errorf("forbidden")
	}
	limit, offset := paginationOrDefault(pagination)
	approved := false
	if isApproved != nil {
		approved = *isApproved
	}
	resp, err := r.GetPendingReviewsUC.Execute(ctx, requests.GetPendingReviewsRequest{
		Limit:       limit,
		Offset:      offset,
		IsModerator: true,
		IsApproved:  approved,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.ReviewBrief, 0, len(resp.Reviews))
	for _, rv := range resp.Reviews {
		out = append(out, &gen.ReviewBrief{
			ID:         rv.GetId(),
			Rating:     rv.GetRating(),
			Comment:    rv.GetComment(),
			VisitDate:  rv.GetVisitDate(),
			CreatedAt:  rv.GetCreatedAt(),
			Username:   rv.GetUsername(),
			IsApproved: rv.IsApproved(),
		})
	}
	return out, nil
}

// Favourites is the resolver for the favourites field.
func (r *queryResolver) Favourites(ctx context.Context, pagination *gen.PaginationInput) ([]*gen.Place, error) {
	uid, err := CurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset := paginationOrDefault(pagination)
	resp, err := r.GetFavUC.Execute(ctx, favourites.GetFavouritesRequest{
		UserID: domain.UserID(uid),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*gen.Place, 0, len(resp.Places))
	for _, p := range resp.Places {
		coords := p.GetCoordinates()
		cStr := fmt.Sprintf("%f,%f", coords.Lat(), coords.Lng())
		out = append(out, &gen.Place{
			ID:           p.GetId(),
			Name:         p.GetName(),
			Category:     p.GetCategory(),
			Description:  stringPtr(p.GetDescription()),
			Coordinates:  &cStr,
			Address:      stringPtr(p.GetAddress()),
			OpeningHours: stringPtr(p.GetOpeningHours()),
			PriceInfo:    stringPtr(p.GetPriceInfo()),
			AvgRating:    floatPtr(p.GetAvgRating()),
			ReviewsCount: int(p.GetReviewCnt()),
			ContactPhone: stringPtr(p.GetContactPhone()),
			Website:      stringPtr(p.GetWebsite()),
			CityID:       p.GetCityID(),
			CityName:     stringPtr(p.GetCityName()),
			DistrictName: stringPtr(p.GetDistrictName()),
			Reviews:      []*gen.ReviewBrief{},
		})
	}
	return out, nil
}

// Report is the resolver for the report field.
func (r *queryResolver) Report(ctx context.Context, id uuid.UUID) (*gen.Report, error) {
	// Асинхронные отчёты в этой итерации не хранятся — возвращаем stub.
	return nil, fmt.Errorf("report %s not found", id)
}

// =========================================================
//                     SUBSCRIPTION
// =========================================================

// ReportStatus is the resolver for the reportStatus field.
func (r *subscriptionResolver) ReportStatus(ctx context.Context, reportID uuid.UUID) (<-chan *gen.Report, error) {
	ch := make(chan *gen.Report)
	now := time.Now()
	go func() {
		defer close(ch)
		ch <- &gen.Report{
			ID:          reportID,
			Status:      gen.ReportStatusCompleted,
			Progress:    100,
			CreatedAt:   now,
			CompletedAt: &now,
		}
	}()
	return ch, nil
}

// =========================================================
//                 ResolverRoot impl
// =========================================================

// Mutation returns gen.MutationResolver implementation.
func (r *Resolver) Mutation() gen.MutationResolver { return &mutationResolver{r} }

// Query returns gen.QueryResolver implementation.
func (r *Resolver) Query() gen.QueryResolver { return &queryResolver{r} }

// Subscription returns gen.SubscriptionResolver implementation.
func (r *Resolver) Subscription() gen.SubscriptionResolver { return &subscriptionResolver{r} }

type (
	mutationResolver     struct{ *Resolver }
	queryResolver        struct{ *Resolver }
	subscriptionResolver struct{ *Resolver }
)
