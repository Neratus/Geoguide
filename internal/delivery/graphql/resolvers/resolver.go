package resolvers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Neratus/geoguide/internal/delivery/middleware"
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/usecase/admin"
	"github.com/Neratus/geoguide/internal/usecase/analytics"
	"github.com/Neratus/geoguide/internal/usecase/auth"
	"github.com/Neratus/geoguide/internal/usecase/catalog"
	"github.com/Neratus/geoguide/internal/usecase/content"
	"github.com/Neratus/geoguide/internal/usecase/favourites"
	"github.com/Neratus/geoguide/internal/usecase/planner"
	"github.com/Neratus/geoguide/internal/usecase/review"
	"github.com/google/uuid"
)

type Resolver struct {
	Logger *slog.Logger

	RegisterUC      *auth.RegisterUserUseCase
	LoginUC         *auth.LoginUseCase
	UpdateProfileUC *auth.UpdateUserProfileUseCase
	GetProfileUC    *auth.GetUserProfileUseCase
	LogoutUC        *auth.LogoutUseCase
	ModerateUserUC  *auth.ModerateUserUsecase
	VerifyContactUC *auth.VerifyContactUseCase
	Enable2FAUC     *auth.EnableTwoFactorUseCase
	Verify2FAUC     *auth.VerifyTwoFactorLoginUseCase

	GetCountriesUC        *catalog.GetCountriesUseCase
	GetCountryUC          *catalog.GetCountryUseCase
	GetCitiesUC           *catalog.GetCitiesByCountryUseCase
	GetCityUC             *catalog.GetCityUseCase
	SearchCitiesUC        *catalog.SearchCitiesUseCase
	GetDistrictsUC        *catalog.GetDistrictsByCityUseCase
	GetTransportNodesUC   *catalog.GetTransportNodesByCityUseCase
	GetPlaceUC            *catalog.GetPlaceUseCase
	GetPlacesByCityUC     *catalog.GetPlacesByCityUseCase
	GetPlacesByCategoryUC *catalog.GetPlacesByCategoryUseCase
	GetHolidaysUC         *catalog.GetHolidaysByDateUseCase

	CreateTripUC          *planner.CreateTripUseCase
	GetTripUC             *planner.GetTripUseCase
	GetUserTripsUC        *planner.GetUserTripsUseCase
	AddPlaceToTripUC      *planner.AddPlaceToTripUseCase
	RemovePlaceFromTripUC *planner.RemovePlaceFromTripUseCase
	UpdateTripUC          *planner.UpdateTripUseCase
	DeleteTripUC          *planner.DeleteTripUseCase

	CreateReviewUC      *review.CreateReviewUseCase
	GetReviewsByPlaceUC *review.GetReviewsByPlaceUseCase
	GetUserReviewsUC    *review.GetUserReviewsUseCase
	ModerateReviewUC    *review.ModerateReviewUseCase
	DeleteReviewUC      *review.DeleteReviewUseCase

	GetStaticPageUC *content.GetStaticPageByIDUseCase

	AddFavUC    *favourites.AddFavouriteUseCase
	RemoveFavUC *favourites.RemoveFavouriteUseCase
	GetFavUC    *favourites.GetFavouritesUseCase

	GetUsersUC          *admin.GetUsersUseCase
	GetPendingReviewsUC *admin.GetPendingReviewsUseCase

	GenerateReportUC *analytics.GenerateReportUseCase
}

func CurrentUserID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(middleware.UserIDKey).(domain.UserID)
	if !ok {
		return uuid.Nil, errors.New("unauthenticated")
	}
	return uuid.UUID(v), nil
}

func CurrentRawToken(ctx context.Context) string {
	v, _ := ctx.Value(middleware.RawTokenKey).(string)
	return v
}

func IsModerator(ctx context.Context) bool {
	v, _ := ctx.Value(middleware.IsModeratorKey).(bool)
	return v
}
