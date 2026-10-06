package delivery

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Neratus/geoguide/internal/api"
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/requests"
	config "github.com/Neratus/geoguide/internal/repository/config"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/Neratus/geoguide/internal/usecase/admin"
	"github.com/Neratus/geoguide/internal/usecase/analytics"
	"github.com/Neratus/geoguide/internal/usecase/auth"
	"github.com/Neratus/geoguide/internal/usecase/catalog"
	"github.com/Neratus/geoguide/internal/usecase/content"
	"github.com/Neratus/geoguide/internal/usecase/favourites"
	"github.com/Neratus/geoguide/internal/usecase/planner"
	"github.com/Neratus/geoguide/internal/usecase/review"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime/types"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Server struct {
	registerUC          *auth.RegisterUserUseCase
	verifyContactUC     *auth.VerifyContactUseCase
	loginUC             *auth.LoginUseCase
	logoutUC            *auth.LogoutUseCase
	enable2FAUC         *auth.EnableTwoFactorUseCase
	verify2FAUC         *auth.VerifyTwoFactorLoginUseCase
	getUserProfileUC    *auth.GetUserProfileUseCase
	updateUserProfileUC *auth.UpdateUserProfileUseCase
	moderateUserUC      *auth.ModerateUserUsecase
	uploadAvatarUC      *auth.UploadAvatarUseCase

	getCountriesUC        *catalog.GetCountriesUseCase
	getCountryUC          *catalog.GetCountryUseCase
	getCitiesByCountryUC  *catalog.GetCitiesByCountryUseCase
	getCityUC             *catalog.GetCityUseCase
	searchCitiesUC        *catalog.SearchCitiesUseCase
	getDistrictsUC        *catalog.GetDistrictsByCityUseCase
	getTransportNodesUC   *catalog.GetTransportNodesByCityUseCase
	getPlacesByCityUC     *catalog.GetPlacesByCityUseCase
	getPlacesByCategoryUC *catalog.GetPlacesByCategoryUseCase
	getPlaceUC            *catalog.GetPlaceUseCase
	getHolidaysByDateUC   *catalog.GetHolidaysByDateUseCase

	createTripUC   *planner.CreateTripUseCase
	getTripUC      *planner.GetTripUseCase
	getUserTripsUC *planner.GetUserTripsUseCase
	updateTripUC   *planner.UpdateTripUseCase
	deleteTripUC   *planner.DeleteTripUseCase
	addPlaceUC     *planner.AddPlaceToTripUseCase
	removePlaceUC  *planner.RemovePlaceFromTripUseCase

	createReviewUC      *review.CreateReviewUseCase
	getReviewsByPlaceUC *review.GetReviewsByPlaceUseCase
	getUserReviewsUC    *review.GetUserReviewsUseCase
	deleteReviewUC      *review.DeleteReviewUseCase
	moderateReviewUC    *review.ModerateReviewUseCase

	getStaticPageUC *content.GetStaticPageByIDUseCase

	generateReportUC *analytics.GenerateReportUseCase

	getUsersUC          *admin.GetUsersUseCase
	getPendingReviewsUC *admin.GetPendingReviewsUseCase

	addFavouriteUC    *favourites.AddFavouriteUseCase
	removeFavouriteUC *favourites.RemoveFavouriteUseCase
	getFavouritesUC   *favourites.GetFavouritesUseCase

	cfg *config.Config
}

func NewServer(
	registerUC *auth.RegisterUserUseCase,
	verifyContactUC *auth.VerifyContactUseCase,
	loginUC *auth.LoginUseCase,
	logoutUC *auth.LogoutUseCase,
	enable2FAUC *auth.EnableTwoFactorUseCase,
	verify2FAUC *auth.VerifyTwoFactorLoginUseCase,
	getUserProfileUC *auth.GetUserProfileUseCase,
	updateUserProfileUC *auth.UpdateUserProfileUseCase,
	uploadAvatarUC *auth.UploadAvatarUseCase,
	moderateUserUC *auth.ModerateUserUsecase,
	getCountriesUC *catalog.GetCountriesUseCase,
	getCountryUC *catalog.GetCountryUseCase,
	getCitiesByCountryUC *catalog.GetCitiesByCountryUseCase,
	getCityUC *catalog.GetCityUseCase,
	searchCitiesUC *catalog.SearchCitiesUseCase,
	getDistrictsUC *catalog.GetDistrictsByCityUseCase,
	getTransportNodesUC *catalog.GetTransportNodesByCityUseCase,
	getPlacesByCityUC *catalog.GetPlacesByCityUseCase,
	getPlacesByCategoryUC *catalog.GetPlacesByCategoryUseCase,
	getPlaceUC *catalog.GetPlaceUseCase,
	createTripUC *planner.CreateTripUseCase,
	getTripUC *planner.GetTripUseCase,
	getUserTripsUC *planner.GetUserTripsUseCase,
	updateTripUC *planner.UpdateTripUseCase,
	deleteTripUC *planner.DeleteTripUseCase,
	addPlaceUC *planner.AddPlaceToTripUseCase,
	removePlaceUC *planner.RemovePlaceFromTripUseCase,
	createReviewUC *review.CreateReviewUseCase,
	getReviewsByPlaceUC *review.GetReviewsByPlaceUseCase,
	getUserReviewsUC *review.GetUserReviewsUseCase,
	deleteReviewUC *review.DeleteReviewUseCase,
	moderateReviewUC *review.ModerateReviewUseCase,
	getStaticPageUC *content.GetStaticPageByIDUseCase,
	generateReportUC *analytics.GenerateReportUseCase,
	getUsersUC *admin.GetUsersUseCase,
	getPendingReviewsUC *admin.GetPendingReviewsUseCase,
	addFavouriteUC *favourites.AddFavouriteUseCase,
	removeFavouriteUC *favourites.RemoveFavouriteUseCase,
	getFavouritesUC *favourites.GetFavouritesUseCase,
	getHolidaysByDateUC *catalog.GetHolidaysByDateUseCase,
	cfg *config.Config,
) *Server {
	return &Server{
		registerUC:            registerUC,
		verifyContactUC:       verifyContactUC,
		loginUC:               loginUC,
		logoutUC:              logoutUC,
		enable2FAUC:           enable2FAUC,
		verify2FAUC:           verify2FAUC,
		getUserProfileUC:      getUserProfileUC,
		updateUserProfileUC:   updateUserProfileUC,
		moderateUserUC:        moderateUserUC,
		uploadAvatarUC:        uploadAvatarUC,
		getCountriesUC:        getCountriesUC,
		getCountryUC:          getCountryUC,
		getCitiesByCountryUC:  getCitiesByCountryUC,
		getCityUC:             getCityUC,
		searchCitiesUC:        searchCitiesUC,
		getDistrictsUC:        getDistrictsUC,
		getTransportNodesUC:   getTransportNodesUC,
		getPlacesByCityUC:     getPlacesByCityUC,
		getPlacesByCategoryUC: getPlacesByCategoryUC,
		getPlaceUC:            getPlaceUC,
		createTripUC:          createTripUC,
		getTripUC:             getTripUC,
		getUserTripsUC:        getUserTripsUC,
		updateTripUC:          updateTripUC,
		deleteTripUC:          deleteTripUC,
		addPlaceUC:            addPlaceUC,
		removePlaceUC:         removePlaceUC,
		createReviewUC:        createReviewUC,
		getReviewsByPlaceUC:   getReviewsByPlaceUC,
		getUserReviewsUC:      getUserReviewsUC,
		deleteReviewUC:        deleteReviewUC,
		moderateReviewUC:      moderateReviewUC,
		getStaticPageUC:       getStaticPageUC,
		generateReportUC:      generateReportUC,
		getUsersUC:            getUsersUC,
		getPendingReviewsUC:   getPendingReviewsUC,
		addFavouriteUC:        addFavouriteUC,
		removeFavouriteUC:     removeFavouriteUC,
		getFavouritesUC:       getFavouritesUC,
		getHolidaysByDateUC:   getHolidaysByDateUC,
		cfg:                   cfg,
	}
}

func (s *Server) RegisterUser(ctx echo.Context) error {
	var req api.RegisterUserRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	var birthDate *time.Time
	if req.BirthDate != nil {
		birthDate = &req.BirthDate.Time
	}

	domainReq := requests.RegisterUserRequest{
		Username:           req.Username,
		Email:              string(req.Email),
		Password:           req.Password,
		Phone:              derefString(req.Phone),
		BirthDate:          birthDate,
		CountryOfResidence: derefString(req.CountryOfResidence),
	}

	res, err := s.registerUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	var respID openapi_types.UUID
	copy(respID[:], res.ID[:])

	str := "Verification email sent"
	resp := api.RegisterUserResponse{
		Id:       &respID,
		Username: &res.Username,
		Email:    ptrEmail(req.Email),
		Message:  &str,
	}

	return ctx.JSON(http.StatusCreated, resp)
}

func (s *Server) VerifyContact(ctx echo.Context) error {
	var req api.VerifyContactRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	domainReq := requests.VerifyContactRequest{
		UserID:  domain.UserID(req.UserId),
		Contact: req.Contact,
		Code:    req.Code,
	}

	if err := s.verifyContactUC.Execute(ctx.Request().Context(), domainReq); err != nil {
		fmt.Println(err)
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "verified successfully"})
}

func (s *Server) Login(ctx echo.Context) error {
	var req api.LoginUserRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	ip := getClientIP(ctx.Request())
	userAgent := ctx.Request().UserAgent()

	domainReq := requests.LoginUserRequest{
		Username: req.Username,
		Password: req.Password,
	}

	res, err := s.loginUC.Execute(ctx.Request().Context(), domainReq, ip, userAgent)
	if err != nil {
		return respondError(ctx, http.StatusUnauthorized, err.Error())
	}

	var token *string
	var challengeID *string
	requiresTwoFactor := false

	if res.Token != "" {
		token = &res.Token
	} else if res.RequiresTwoFactor {
		requiresTwoFactor = true
		challengeID = &res.ChallengeID
	}
	email := types.Email(res.Email)

	resp := api.LoginUserResponse{
		Id:                (*openapi_types.UUID)(&res.ID),
		Username:          &res.Username,
		Email:             &email,
		Token:             token,
		RequiresTwoFactor: &requiresTwoFactor,
		ChallengeId:       challengeID,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) Logout(ctx echo.Context) error {
	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader == "" {
		return respondError(ctx, http.StatusUnauthorized, "missing authorization header")
	}

	parts := splitBearer(authHeader)
	if parts == nil {
		return respondError(ctx, http.StatusUnauthorized, "invalid authorization header format")
	}

	token := parts[1]
	_, err := s.logoutUC.Execute(ctx.Request().Context(), requests.LogoutRequest{Token: token})
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (s *Server) EnableTwoFactor(ctx echo.Context) error {
	var req api.EnableTwoFactorRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	domainReq := requests.EnableTwoFactorRequest{
		UserID: userID,
		Secret: req.Secret,
		Code:   req.Code,
	}

	backupCodes, err := s.enable2FAUC.VerifyAndEnable(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	otpauthUri := ""
	resp := api.EnableTwoFactorResponse{
		BackupCodes: &backupCodes,
		OtpauthUri:  &otpauthUri,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) VerifyTwoFactor(ctx echo.Context) error {
	var req api.VerifyTwoFactorRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	domainReq := requests.VerifyTwoFactorRequest{
		ChallengeID: req.ChallengeId,
		Code:        req.Code,
	}

	token, err := s.verify2FAUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusUnauthorized, err.Error())
	}

	resp := api.TwoFactorResponse{Token: &token}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetUserProfile(ctx echo.Context) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	profile, err := s.getUserProfileUC.Execute(ctx.Request().Context(), requests.GetUserProfileRequest{UserID: userID})
	if err != nil {
		return respondError(ctx, http.StatusNotFound, err.Error())
	}

	var birthDate *openapi_types.Date
	if profile.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", profile.BirthDate); err == nil {
			d := openapi_types.Date{Time: t}
			birthDate = &d
		}
	}

	registeredAt, _ := time.Parse(time.RFC3339, profile.RegisteredAt)
	roleValue := api.GetUserProfileResponseRole(profile.Role)

	email := types.Email(profile.Email)
	resp := api.GetUserProfileResponse{
		Id:                 (*openapi_types.UUID)(&profile.ID),
		Username:           &profile.Username,
		Email:              &email,
		Phone:              &profile.Phone,
		RegisteredAt:       &registeredAt,
		BirthDate:          birthDate,
		CountryOfResidence: &profile.CountryOfResidence,
		AvatarUrl:          &profile.AvatarURL,
		FavoriteCategories: &profile.FavoriteCategories,
		IsBlocked:          &profile.IsBlocked,
		Role:               &roleValue,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) UpdateUserProfile(ctx echo.Context) error {
	_, err := getUserID(ctx)
	if err != nil {
		return err
	}

	var req api.UpdateUserProfileRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader == "" {
		return respondError(ctx, http.StatusUnauthorized, "missing authorization header")
	}
	parts := splitBearer(authHeader)
	if parts == nil {
		return respondError(ctx, http.StatusUnauthorized, "invalid authorization header format")
	}
	token := parts[1]

	domainReq := requests.UpdateUserProfileRequest{
		Token:              token,
		Username:           req.Username,
		Phone:              req.Phone,
		CountryOfResidence: req.CountryOfResidence,
		AvatarURL:          req.AvatarUrl,
		FavoriteCategories: derefStringSlice(req.FavoriteCategories),
	}
	if req.BirthDate != nil {
		domainReq.BirthDate = &req.BirthDate.Time
	}

	updated, err := s.updateUserProfileUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	var birthDate *openapi_types.Date
	if updated.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", updated.BirthDate); err == nil {
			d := openapi_types.Date{Time: t}
			birthDate = &d
		}
	}
	registeredAt, _ := time.Parse(time.RFC3339, updated.RegisteredAt)

	mail := types.Email(updated.Email)
	resp := api.GetUserProfileResponse{
		Id:                 (*openapi_types.UUID)(&updated.ID),
		Username:           &updated.Username,
		Email:              &mail,
		Phone:              &updated.Phone,
		RegisteredAt:       &registeredAt,
		BirthDate:          birthDate,
		CountryOfResidence: &updated.CountryOfResidence,
		AvatarUrl:          &updated.AvatarURL,
		FavoriteCategories: &updated.FavoriteCategories,
		IsBlocked:          &updated.IsBlocked,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) ModerateUser(ctx echo.Context) error {
	var req api.ModerateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	moderatorID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	domainReq := requests.ModerateUserRequest{
		UserID:      domain.UserID(req.UserId),
		Block:       req.Block,
		Reason:      derefString(req.Reason),
		ModeratorID: moderatorID,
	}

	res, err := s.moderateUserUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	resp := api.ModerateUserResponse{
		UserId:      &req.UserId,
		IsBlocked:   &res.IsBlocked,
		BlockedAt:   res.BlockedAt,
		BlockReason: &res.BlockReason,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetUsers(ctx echo.Context, params api.GetUsersParams) error {
	if !getIsModerator(ctx) {
		return respondError(ctx, http.StatusForbidden, "forbidden")
	}

	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)
	search := derefStringWithDefault(params.Search, "")

	req := requests.GetUsersRequest{
		Limit:       limit,
		Offset:      offset,
		Search:      search,
		IsModerator: true,
	}

	resp, err := s.getUsersUC.Execute(ctx.Request().Context(), req)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	result := make([]api.UserResponse, 0, len(resp.Users))
	for _, u := range resp.Users {
		if u.GetRole() != domain.UserRoleUser {
			continue
		}
		userID := u.GetId()
		username := u.GetUsername()
		email := types.Email(u.GetEmail())

		isBlocked := u.IsBlocked()
		registeredAt := u.GetRegisteredAt()
		result = append(result, api.UserResponse{
			Id:           (*openapi_types.UUID)(&userID),
			Username:     &username,
			Email:        &email,
			IsBlocked:    &isBlocked,
			RegisteredAt: &registeredAt,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

func (s *Server) GetPendingReviews(ctx echo.Context, params api.GetPendingReviewsParams) error {
	if !getIsModerator(ctx) {
		return respondError(ctx, http.StatusForbidden, "forbidden")
	}

	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)
	isApproved := derefBoolWithDefault(params.IsApproved, false)

	req := requests.GetPendingReviewsRequest{
		Limit:       limit,
		Offset:      offset,
		IsModerator: true,
		IsApproved:  isApproved,
	}

	resp, err := s.getPendingReviewsUC.Execute(ctx.Request().Context(), req)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	result := make([]api.ReviewBriefResponse, 0, len(resp.Reviews))
	for _, rev := range resp.Reviews {
		rating := rev.GetRating()
		comment := rev.GetComment()
		visitDate := openapi_types.Date{Time: rev.GetVisitDate()}
		createdAt := rev.GetCreatedAt()
		isApproved := rev.IsApproved()
		username := rev.GetUsername()
		id := rev.GetId()
		result = append(result, api.ReviewBriefResponse{
			Id:         (*openapi_types.UUID)(&id),
			Rating:     &rating,
			Comment:    &comment,
			VisitDate:  &visitDate,
			CreatedAt:  &createdAt,
			Username:   &username,
			IsApproved: &isApproved,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

func (s *Server) GetCountries(ctx echo.Context, params api.GetCountriesParams) error {
	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)
	search := derefStringWithDefault(params.Search, "")

	domainReq := requests.GetCountriesRequest{Limit: limit, Offset: offset, Search: search}
	countries, err := s.getCountriesUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.CountryResponse, 0, len(countries))
	for _, c := range countries {
		capitalName := c.Capital
		area := float32(c.Area)
		resp = append(resp, api.CountryResponse{
			Id:         (*openapi_types.UUID)(&c.ID),
			Name:       &c.Name,
			Capital:    &capitalName,
			Area:       &area,
			Population: &c.Population,
			Currency:   &c.Currency,
			Language:   &c.Language,
			PhoneCode:  &c.PhoneCode,
			ImageUrl:   &c.ImageURL,
		})
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetCountry(ctx echo.Context, id openapi_types.UUID) error {
	domainReq := requests.GetCountryRequest{ID: domain.CountryID(id)}
	country, err := s.getCountryUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusNotFound, err.Error())
	}

	holidays := make([]api.HolidayResponse, 0, len(country.Holidays))
	for _, h := range country.Holidays {
		date := openapi_types.Date{Time: h.Date}
		holidays = append(holidays, api.HolidayResponse{
			Id:          (*openapi_types.UUID)(&h.ID),
			Name:        &h.Name,
			Date:        &date,
			Description: &h.Description,
			IsNational:  &h.IsNational,
		})
	}

	area := float32(country.Area)
	gdp := float32(country.GDP)

	resp := api.GetCountryResponse{
		Id:               (*openapi_types.UUID)(&country.ID),
		Name:             &country.Name,
		Area:             &area,
		Population:       &country.Population,
		Gdp:              &gdp,
		Currency:         &country.Currency,
		VisaRequirements: &country.VisaRequirements,
		Description:      &country.Description,
		SafetyTips:       &country.SafetyTips,
		BestSeason:       &country.BestSeason,
		Language:         &country.Language,
		PhoneCode:        &country.PhoneCode,
		Religion:         &country.Religion,
		CapitalId:        (*openapi_types.UUID)(&country.CapitalID),
		CapitalName:      &country.CapitalName,
		ImageUrl:         &country.ImageURL,
		Holidays:         &holidays,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetCitiesByCountry(ctx echo.Context, countryId openapi_types.UUID, params api.GetCitiesByCountryParams) error {
	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)

	domainReq := requests.GetCitiesByCountryRequest{CountryID: domain.CountryID(countryId), Limit: limit, Offset: offset}
	cities, err := s.getCitiesByCountryUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.CityResponse, 0, len(cities))
	for _, city := range cities {
		coord := parseCoordinatesString(city.Coordinates)
		resp = append(resp, api.CityResponse{
			Id:          (*openapi_types.UUID)(&city.ID),
			Name:        &city.Name,
			Population:  &city.Population,
			IsCapital:   &city.IsCapital,
			Coordinates: coord,
			Description: &city.Description,
			Timezone:    &city.Timezone,
			ImageUrl:    &city.ImageURL,
		})
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetCity(ctx echo.Context, id openapi_types.UUID) error {
	domainReq := requests.GetCityRequest{ID: domain.CityID(id)}
	city, err := s.getCityUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusNotFound, err.Error())
	}

	coord := convertCoordinatesToAPI(city.Coordinates)
	resp := api.GetCityResponse{
		Id:          (*openapi_types.UUID)(&city.ID),
		Name:        &city.Name,
		Population:  &city.Population,
		IsCapital:   &city.IsCapital,
		Coordinates: coord,
		Description: &city.Description,
		Timezone:    &city.Timezone,
		TravelTips:  &city.TravelTips,
		CountryId:   (*openapi_types.UUID)(&city.CountryID),
		ImageId:     (*openapi_types.UUID)(&city.ImageID),
		ImageUrl:    &city.ImageURL,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) SearchCities(ctx echo.Context, params api.SearchCitiesParams) error {
	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)
	query := derefStringWithDefault(params.Query, "")

	domainReq := requests.SearchCitiesRequest{Query: query, Limit: limit, Offset: offset}
	cities, err := s.searchCitiesUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.CityResponse, 0, len(cities))
	for _, city := range cities {
		coord := parseCoordinatesString(city.Coordinates)
		resp = append(resp, api.CityResponse{
			Id:          (*openapi_types.UUID)(&city.ID),
			Name:        &city.Name,
			Population:  &city.Population,
			IsCapital:   &city.IsCapital,
			Coordinates: coord,
			Description: &city.Description,
			Timezone:    &city.Timezone,
			ImageUrl:    &city.ImageURL,
		})
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetDistrictsByCity(ctx echo.Context, cityId openapi_types.UUID, params api.GetDistrictsByCityParams) error {
	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)

	domainReq := requests.GetDistrictsByCityRequest{CityID: domain.CityID(cityId), Limit: limit, Offset: offset}
	districts, err := s.getDistrictsUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.DistrictResponse, 0, len(districts))
	for _, d := range districts {
		coord := api.Coordinates{Lat: d.Coordinates.Lat(), Lng: d.Coordinates.Lng()}
		resp = append(resp, api.DistrictResponse{
			Id:          (*openapi_types.UUID)(&d.ID),
			Name:        &d.Name,
			Description: &d.Description,
			Coordinates: &coord,
			CityId:      (*openapi_types.UUID)(&d.CityID),
			ImageId:     (*openapi_types.UUID)(&d.ImageID),
		})
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetTransportNodesByCity(ctx echo.Context, cityId openapi_types.UUID, params api.GetTransportNodesByCityParams) error {
	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)

	domainReq := requests.GetTransportNodesByCityRequest{CityID: domain.CityID(cityId), Limit: limit, Offset: offset}
	nodes, err := s.getTransportNodesUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.TransportNodeResponse, 0, len(nodes))
	for _, n := range nodes {
		coord := convertCoordinatesToAPI(n.Coordinates)
		nodeType := api.TransportNodeResponseNodeType(n.NodeType)
		resp = append(resp, api.TransportNodeResponse{
			Id:          (*openapi_types.UUID)(&n.ID),
			Name:        &n.Name,
			NodeType:    &nodeType,
			Coordinates: coord,
			Address:     &n.Address,
			CityId:      (*openapi_types.UUID)(&n.CityID),
			ImageId:     (*openapi_types.UUID)(&n.ImageID),
		})
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetPlaces(ctx echo.Context, params api.GetPlacesParams) error {
	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)
	category := derefStringWithDefault(params.Category, "")
	sortBy := ""
	if params.SortBy != nil {
		sortBy = string(*params.SortBy)
	}

	if params.CityId != nil {
		domainReq := requests.GetPlacesByCityRequest{
			CityID:   domain.CityID(*params.CityId),
			Category: category,
			Limit:    limit,
			Offset:   offset,
			SortBy:   sortBy,
		}
		places, err := s.getPlacesByCityUC.Execute(ctx.Request().Context(), domainReq)
		if err != nil {
			return respondError(ctx, http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(http.StatusOK, buildPlaceResponses(places))
	}

	if category == "" {
		return respondError(ctx, http.StatusBadRequest, "either cityId or category must be provided")
	}

	domainReq := requests.GetPlacesByCategoryRequest{
		Category: category,
		Limit:    limit,
		Offset:   offset,
	}
	places, err := s.getPlacesByCategoryUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, buildPlaceResponses(places))
}

func (s *Server) GetPlace(ctx echo.Context, id openapi_types.UUID) error {
	domainReq := requests.GetPlaceRequest{ID: domain.PlaceID(id)}
	place, err := s.getPlaceUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusNotFound, err.Error())
	}

	coordStr := place.Coordinates
	reviews := buildReviewBriefResponses(place.Reviews)
	rating := float32(place.AvgRating)

	resp := api.PlaceResponse{
		Id:                  (*openapi_types.UUID)(&place.ID),
		Name:                &place.Name,
		Category:            &place.Category,
		Description:         &place.Description,
		Coordinates:         &coordStr,
		Address:             &place.Address,
		OpeningHours:        &place.OpeningHours,
		PriceInfo:           &place.PriceInfo,
		AvgVisitDurationMin: &place.AvgVisitDurationMin,
		AvgRating:           &rating,
		ReviewsCount:        &place.ReviewsCount,
		ContactPhone:        &place.ContactPhone,
		Website:             &place.Website,
		CityId:              (*openapi_types.UUID)(&place.CityID),
		CityName:            &place.CityName,
		DistrictId:          (*openapi_types.UUID)(&place.DistrictID),
		DistrictName:        &place.DistrictName,
		ImageUrl:            &place.ImageURL,
		Reviews:             &reviews,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) CreateTrip(ctx echo.Context) error {
	var req api.CreateTripRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}
	budget := req.Budget

	domainReq := requests.CreateTripRequest{
		UserID:    userID,
		Title:     req.Title,
		StartDate: req.StartDate.Time,
		EndDate:   req.EndDate.Time,
		Budget:    float64(*budget),
		Notes:     derefString(req.Notes),
	}

	res, err := s.createTripUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	resp := api.CreateTripResponse{
		Id:        (*openapi_types.UUID)(&res.ID),
		Title:     &res.Title,
		StartDate: datePtr(parseDate(res.StartDate)),
		EndDate:   datePtr(parseDate(res.EndDate)),
		Budget:    float32Ptr(res.Budget),
		Status:    &res.Status,
	}

	return ctx.JSON(http.StatusCreated, resp)
}

func (s *Server) GetUserTrips(ctx echo.Context) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	trips, err := s.getUserTripsUC.Execute(ctx.Request().Context(), requests.GetUserTripsRequest{UserID: userID})
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.UserTripResponse, 0, len(trips))
	for _, t := range trips {
		resp = append(resp, api.UserTripResponse{
			Id:        (*openapi_types.UUID)(&t.ID),
			Title:     &t.Title,
			StartDate: datePtr(parseDate(t.StartDate)),
			EndDate:   datePtr(parseDate(t.EndDate)),
			Status:    &t.Status,
		})
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetTrip(ctx echo.Context, tripId openapi_types.UUID) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	domainReq := requests.GetTripRequest{TripID: domain.TripID(tripId), UserID: userID}
	trip, err := s.getTripUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusNotFound, err.Error())
	}

	places := make([]api.TripPlaceResponse, 0, len(trip.Places))
	for _, p := range trip.Places {
		var arrivalTimePtr *time.Time
		if p.ArrivalTime != "" {
			if t, err := time.Parse(time.RFC3339, p.ArrivalTime); err == nil {
				arrivalTimePtr = &t
			}
		}
		places = append(places, api.TripPlaceResponse{
			PlaceId:     (*openapi_types.UUID)(&p.PlaceID),
			DayNumber:   &p.DayNumber,
			ArrivalTime: arrivalTimePtr,
			DurationMin: &p.DurationMin,
			Notes:       &p.Notes,
			VisitStatus: &p.VisitStatus,
			ActualCost:  float32Ptr(p.ActualCost),
		})
	}

	resp := api.GetTripResponse{
		Id:        (*openapi_types.UUID)(&trip.ID),
		Title:     &trip.Title,
		StartDate: datePtr(parseDate(trip.StartDate)),
		EndDate:   datePtr(parseDate(trip.EndDate)),
		Budget:    float32Ptr(trip.Budget),
		Status:    &trip.Status,
		Notes:     &trip.Notes,
		Places:    &places,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) UpdateTrip(ctx echo.Context, tripId openapi_types.UUID) error {
	var req api.UpdateTripRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	domainReq := requests.UpdateTripRequest{
		TripID:    domain.TripID(tripId),
		UserID:    userID,
		Title:     req.Title,
		StartDate: mapOptionalDate(req.StartDate),
		EndDate:   mapOptionalDate(req.EndDate),
		Budget:    mapOptionalFloat32(req.Budget),
		Status:    req.Status,
		Notes:     req.Notes,
	}

	if err := s.updateTripUC.Execute(ctx.Request().Context(), domainReq); err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (s *Server) DeleteTrip(ctx echo.Context, tripId openapi_types.UUID) error {
	domainReq := requests.DeleteTripRequest{TripID: domain.TripID(tripId)}
	if err := s.deleteTripUC.Execute(ctx.Request().Context(), domainReq); err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (s *Server) AddPlaceToTrip(ctx echo.Context, tripId openapi_types.UUID) error {
	var req api.AddPlaceToTripRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	dur := req.DurationMin

	domainReq := requests.AddPlaceToTripRequest{
		TripID:      domain.TripID(tripId),
		PlaceID:     domain.PlaceID(req.PlaceId),
		DayNumber:   req.DayNumber,
		ArrivalTime: req.ArrivalTime,
		DurationMin: *dur,
		Notes:       derefString(req.Notes),
		UserID:      userID,
	}

	if err := s.addPlaceUC.Execute(ctx.Request().Context(), domainReq); err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.NoContent(http.StatusCreated)
}

func (s *Server) RemovePlaceFromTrip(ctx echo.Context, tripId openapi_types.UUID, params api.RemovePlaceFromTripParams) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	domainReq := requests.RemovePlaceFromTripRequest{
		TripID:  domain.TripID(tripId),
		PlaceID: domain.PlaceID(params.PlaceId),
		UserID:  userID,
	}

	if err := s.removePlaceUC.Execute(ctx.Request().Context(), domainReq); err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (s *Server) CreateReview(ctx echo.Context, placeId openapi_types.UUID) error {
	var req api.CreateReviewRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	domainReq := requests.CreateReviewRequest{
		UserID:    userID,
		PlaceID:   domain.PlaceID(placeId),
		Rating:    req.Rating,
		Comment:   req.Comment,
		VisitDate: req.VisitDate.Time,
		ImageURL:  derefString(req.ImageUrl),
	}

	res, err := s.createReviewUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	resp := api.CreateReviewResponse{
		Id:          (*openapi_types.UUID)(&res.ID),
		Rating:      &res.Rating,
		Comment:     &res.Comment,
		VisitDate:   datePtr(parseDate(res.VisitDate)),
		CreatedAt:   parseTimePtr(res.CreatedAt),
		IsModerated: &res.IsModerated,
		IsApproved:  &res.IsApproved,
	}

	return ctx.JSON(http.StatusCreated, resp)
}

func (s *Server) GetReviewsByPlace(ctx echo.Context, placeId openapi_types.UUID, params api.GetReviewsByPlaceParams) error {
	onlyApproved := derefBoolWithDefault(params.OnlyApproved, true)
	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)

	domainReq := requests.GetReviewsByPlaceRequest{
		PlaceID: domain.PlaceID(placeId), OnlyApproved: onlyApproved, Limit: limit, Offset: offset,
	}
	reviews, err := s.getReviewsByPlaceUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := buildReviewByPlaceResponses(reviews)
	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetUserReviews(ctx echo.Context, params api.GetUserReviewsParams) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)

	reviews, err := s.getUserReviewsUC.Execute(ctx.Request().Context(), requests.GetUserReviewsRequest{
		UserID: userID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.UserReviewResponse, 0, len(reviews))
	for _, rv := range reviews {
		resp = append(resp, api.UserReviewResponse{
			Id:          (*openapi_types.UUID)(&rv.ID),
			Rating:      &rv.Rating,
			Comment:     &rv.Comment,
			VisitDate:   datePtr(parseDate(rv.VisitDate)),
			CreatedAt:   parseTimePtr(rv.CreatedAt),
			PlaceId:     (*openapi_types.UUID)(&rv.PlaceID),
			PlaceName:   &rv.PlaceName,
			IsApproved:  &rv.IsApproved,
			IsModerated: &rv.IsModerated,
		})
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) DeleteReview(ctx echo.Context, reviewId openapi_types.UUID) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}
	isModerator := getIsModerator(ctx)

	domainReq := requests.DeleteReviewRequest{
		ReviewID: domain.ReviewID(reviewId), UserID: userID, IsModerator: isModerator,
	}
	res, err := s.deleteReviewUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}
	if !res.Success {
		return respondError(ctx, http.StatusInternalServerError, "failed to delete review")
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (s *Server) ModerateReview(ctx echo.Context) error {
	var req api.ModerateReviewRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	moderatorID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	domainReq := requests.ModerateReviewRequest{
		ReviewID:    domain.ReviewID(req.ReviewId),
		Approved:    req.Approved,
		Comment:     derefString(req.Comment),
		ModeratorID: moderatorID,
	}

	res, err := s.moderateReviewUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, err.Error())
	}

	resp := api.ModerateReviewResponse{
		Id:                (*openapi_types.UUID)(&res.ID),
		IsApproved:        &res.IsApproved,
		IsModerated:       &res.IsModerated,
		ModerationComment: &res.ModerationComment,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) GetHolidaysByDate(ctx echo.Context, params api.GetHolidaysByDateParams) error {
	req := requests.GetHolidaysByDateRequest{Date: params.Date.Time}
	holidays, err := s.getHolidaysByDateUC.Execute(ctx.Request().Context(), req)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, "failed to fetch holidays: "+err.Error())
	}

	response := make([]api.HolidayResponse, len(holidays))
	for i, h := range holidays {
		dateResp := openapi_types.Date{Time: h.Date}
		response[i] = api.HolidayResponse{
			Id:          (*openapi_types.UUID)(&h.ID),
			Name:        &h.Name,
			Date:        &dateResp,
			Description: &h.Description,
			IsNational:  &h.IsNational,
		}
	}

	return ctx.JSON(http.StatusOK, response)
}

func (s *Server) GetStaticPage(ctx echo.Context, id openapi_types.UUID) error {
	page, err := s.getStaticPageUC.Execute(ctx.Request().Context(), content.GetStaticPageByIDRequest{ID: domain.StaticPageID(id)})
	if err != nil {
		return respondError(ctx, http.StatusNotFound, "page not found")
	}

	var publishedAtPtr *time.Time
	if page.PublishedAt != "" {
		now := time.Now()
		publishedAtPtr = &now
	}
	updatedAt := time.Now()

	resp := api.StaticPageResponse{
		Id:              (*openapi_types.UUID)(&page.ID),
		Slug:            &page.Slug,
		Title:           &page.Title,
		Content:         &page.Content,
		MetaDescription: &page.MetaDesc,
		ImageUrl:        &page.ImageURL,
		ImageAlt:        &page.ImageAlt,
		UpdatedAt:       &updatedAt,
		PublishedAt:     publishedAtPtr,
		IsPublished:     &page.IsPublished,
	}

	return ctx.JSON(http.StatusOK, resp)
}

// TODO: ассинхронная генерация
func (s *Server) GenerateReport(ctx echo.Context) error {
	var req api.GenerateReportRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	limit := req.Limit
	domainReq := requests.GenerateReportRequest{
		ReportType: string(req.ReportType),
		Format:     domain.ReportFormat(req.Format),
		Options: domain.ReportOptions{
			IncludeHeaders: true,
			PrettyPrint:    req.Format == api.Json,
		},
		Limit: *limit,
	}
	if req.DateFrom != nil {
		domainReq.DateFrom = req.DateFrom.String()
	}
	if req.DateTo != nil {
		domainReq.DateTo = req.DateTo.String()
	}

	reportData, err := s.generateReportUC.Execute(ctx.Request().Context(), domainReq)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	if req.Format == api.Json {
		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.Blob(http.StatusOK, "application/json", reportData)
	}

	ctx.Response().Header().Set("Content-Type", "text/csv")
	ctx.Response().Header().Set("Content-Disposition", "attachment; filename=report.csv")
	return ctx.Blob(http.StatusOK, "text/csv", []byte(reportData))
}

// TODO: убрать заглушку
func (s *Server) StreamReportStatus(ctx echo.Context, reportId openapi_types.UUID) error {
	ctx.Response().Header().Set("Content-Type", "text/event-stream")
	ctx.Response().Header().Set("Cache-Control", "no-cache")
	ctx.Response().Header().Set("Connection", "keep-alive")

	_, err := fmt.Fprintf(ctx.Response(), "data: {\"status\": \"completed\", \"reportId\": \"%s\"}\n\n", reportId.String())
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, "stream write error")
	}
	ctx.Response().Flush()
	return nil
}

func (s *Server) GetFavourites(ctx echo.Context, params api.GetFavouritesParams) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	limit := derefIntWithDefault(params.Limit, 20)
	offset := derefIntWithDefault(params.Offset, 0)

	resp, err := s.getFavouritesUC.Execute(ctx.Request().Context(), favourites.GetFavouritesRequest{
		UserID: userID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, "internal error")
	}

	items := make([]api.PlaceResponse, len(resp.Places))
	for i, p := range resp.Places {
		coords := p.GetCoordinates()
		coordStr := fmt.Sprintf("%f,%f", coords.Lat(), coords.Lng())
		name := p.GetName()
		cat := p.GetCategory()
		id := p.GetId()
		cityID := p.GetCityID()
		distrID := p.GetDistrictID()
		dur := (int(p.GetAvgVisitDurationMin()))
		cnt := int(p.GetReviewCnt())
		items[i] = api.PlaceResponse{
			Id:                  (*openapi_types.UUID)(&id),
			Name:                &name,
			Category:            &cat,
			Description:         stringPtr(p.GetDescription()),
			Coordinates:         &coordStr,
			Address:             stringPtr(p.GetAddress()),
			OpeningHours:        stringPtr(p.GetOpeningHours()),
			PriceInfo:           stringPtr(p.GetPriceInfo()),
			AvgVisitDurationMin: &dur,
			AvgRating:           float32Ptr(p.GetAvgRating()),
			ReviewsCount:        &cnt,
			ContactPhone:        stringPtr(p.GetContactPhone()),
			Website:             stringPtr(p.GetWebsite()),
			CityId:              (*openapi_types.UUID)(&cityID),
			CityName:            stringPtr(p.GetCityName()),
			DistrictId:          (*openapi_types.UUID)(&distrID),
			DistrictName:        stringPtr(p.GetDistrictName()),
			Reviews:             &[]api.ReviewBriefResponse{},
		}
	}

	return ctx.JSON(http.StatusOK, items)
}

func (s *Server) AddFavourite(ctx echo.Context) error {
	var req api.FavouriteRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	if err := s.addFavouriteUC.Execute(ctx.Request().Context(), favourites.AddFavouriteRequest{
		UserID:  userID,
		PlaceID: req.PlaceId,
	}); err != nil {
		if errors.Is(err, usecase_errors.ErrPlaceNotFound) {
			return respondError(ctx, http.StatusNotFound, "place not found")
		}
		return respondError(ctx, http.StatusInternalServerError, "internal error")
	}

	return ctx.NoContent(http.StatusCreated)
}

func (s *Server) RemoveFavourite(ctx echo.Context) error {
	var req api.FavouriteRequest
	if err := ctx.Bind(&req); err != nil {
		return respondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	if err := s.removeFavouriteUC.Execute(ctx.Request().Context(), favourites.RemoveFavouriteRequest{
		UserID:  userID,
		PlaceID: req.PlaceId,
	}); err != nil {
		return respondError(ctx, http.StatusInternalServerError, "internal error")
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (s *Server) UploadAvatar(ctx echo.Context) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, "missing file")
	}
	if file.Size > 5<<20 {
		return respondError(ctx, http.StatusBadRequest, "file too large")
	}

	src, err := file.Open()
	if err != nil {
		return respondError(ctx, http.StatusBadRequest, "cannot open file")
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return respondError(ctx, http.StatusBadRequest, "only JPEG, PNG, WEBP allowed")
	}

	avatarURL, err := s.uploadAvatarUC.Execute(ctx.Request().Context(), userID, src, file.Size, file.Filename, contentType)
	if err != nil {
		return respondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"avatarUrl": avatarURL})
}

func buildReviewBriefResponses(reviews []requests.ReviewResponse) []api.ReviewBriefResponse {
	out := make([]api.ReviewBriefResponse, 0, len(reviews))
	for _, rv := range reviews {
		visitDate := parseDate(rv.VisitDate)
		createdAt := parseTimePtr(rv.CreatedAt)
		rating := rv.Rating

		out = append(out, api.ReviewBriefResponse{
			Id:         (*openapi_types.UUID)(&rv.ID),
			Rating:     &rating,
			Comment:    &rv.Comment,
			VisitDate:  visitDate,
			CreatedAt:  createdAt,
			Username:   &rv.Username,
			IsApproved: &rv.IsApproved,
		})
	}
	return out
}

func buildReviewByPlaceResponses(reviews []requests.ReviewByPlaceResponse) []api.ReviewBriefResponse {
	out := make([]api.ReviewBriefResponse, 0, len(reviews))
	for _, rv := range reviews {
		visitDate := parseDate(rv.VisitDate)
		createdAt := parseTimePtr(rv.CreatedAt)
		rating := rv.Rating

		out = append(out, api.ReviewBriefResponse{
			Id:         (*openapi_types.UUID)(&rv.ID),
			Rating:     &rating,
			Comment:    &rv.Comment,
			VisitDate:  visitDate,
			CreatedAt:  createdAt,
			Username:   &rv.Username,
			IsApproved: &rv.IsApproved,
		})
	}
	return out
}

func buildPlaceResponses(places []requests.PlaceResponse) []api.PlaceResponse {
	resp := make([]api.PlaceResponse, 0, len(places))
	for _, p := range places {
		reviews := buildReviewBriefResponses(p.Reviews)

		avgRating := float32(p.AvgRating)
		reviewsCount := int(p.ReviewsCount)
		avgVisitDuration := int(p.AvgVisitDurationMin)

		resp = append(resp, api.PlaceResponse{
			Id:                  (*openapi_types.UUID)(&p.ID),
			Name:                &p.Name,
			Category:            &p.Category,
			Description:         &p.Description,
			Coordinates:         &p.Coordinates,
			Address:             &p.Address,
			OpeningHours:        &p.OpeningHours,
			PriceInfo:           &p.PriceInfo,
			AvgVisitDurationMin: &avgVisitDuration,
			AvgRating:           &avgRating,
			ReviewsCount:        &reviewsCount,
			ContactPhone:        &p.ContactPhone,
			Website:             &p.Website,
			CityId:              (*openapi_types.UUID)(&p.CityID),
			CityName:            &p.CityName,
			DistrictId:          (*openapi_types.UUID)(&p.DistrictID),
			DistrictName:        &p.DistrictName,
			ImageUrl:            &p.ImageURL,
			Reviews:             &reviews,
		})
	}
	return resp
}
