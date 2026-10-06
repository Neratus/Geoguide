package auth_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/Neratus/geoguide/internal/domain/requests"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	user_repo "github.com/Neratus/geoguide/internal/repository/user"
	usecase_errors "github.com/Neratus/geoguide/internal/usecase"
	"github.com/Neratus/geoguide/internal/usecase/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

type allowAllRateLimiter struct{}

func (r *allowAllRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}

type restrictedRateLimiter struct{}

func (r *restrictedRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return false, nil
}

type testDeps struct {
	ctx         context.Context
	logger      *slog.Logger
	userRepo    *user_repo.PostgresUserRepo
	pgxMock     pgxmock.PgxPoolIface
	sessionSpy  *sessionRepoSpy
	taskSpy     *taskQueueSpy
	rateLimiter *rateLimiterSpy
	staticMock  *staticRepoMock
}

func setupTestDeps(t *testing.T) *testDeps {
	t.Helper()

	pgxMock, err := pgxmock.NewPool()
	require.NoError(t, err)

	userRepo := &user_repo.PostgresUserRepo{
		Pool:    pgxMock,
		Queries: postgreSQL.New(pgxMock),
	}

	t.Cleanup(func() {
		if err := pgxMock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet pgx expectations: %v", err)
		}
		pgxMock.Close()
	})

	return &testDeps{
		ctx:         context.Background(),
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		userRepo:    userRepo,
		pgxMock:     pgxMock,
		sessionSpy:  &sessionRepoSpy{},
		taskSpy:     &taskQueueSpy{},
		rateLimiter: &rateLimiterSpy{AllowResult: true},
		staticMock:  &staticRepoMock{GetFileURLResult: "https://cdn.example.com/avatar.jpg"},
	}
}

type taskQueueSpy struct {
	PublishedTasks []interfaces.Task
	PublishCalled  bool
}

func (q *taskQueueSpy) Publish(ctx context.Context, task interfaces.Task) error {
	q.PublishedTasks = append(q.PublishedTasks, task)
	q.PublishCalled = true
	return nil
}
func (q *taskQueueSpy) Subscribe(taskName string, handler interfaces.TaskHandler) error { return nil }
func (q *taskQueueSpy) Run(ctx context.Context) error                                   { return nil }
func (q *taskQueueSpy) Close() error                                                    { return nil }

func defaultRegisterRequest() requests.RegisterUserRequest {
	return requests.RegisterUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "StrongPass123!",
	}
}

func defaultVerifyContactRequest(userID domain.UserID) requests.VerifyContactRequest {
	return requests.VerifyContactRequest{
		UserID:  userID,
		Contact: "test@example.com",
		Code:    "123456",
	}
}

func defaultGetProfileRequest(userID domain.UserID) requests.GetUserProfileRequest {
	return requests.GetUserProfileRequest{UserID: userID}
}

func defaultUpdateProfileRequest(token string) requests.UpdateUserProfileRequest {
	username := "updated_user"
	return requests.UpdateUserProfileRequest{
		Token:    token,
		Username: &username,
	}
}

func defaultLogoutRequest(token string) requests.LogoutRequest {
	return requests.LogoutRequest{Token: token}
}

func defaultMockUserRow(userID uuid.UUID) *pgxmock.Rows {
	return buildMockUserRow(userID, "testuser", "test@example.com", "hashed_password", false, nil, "")
}

func buildMockUserRow(
	userID uuid.UUID,
	username, email, passwordHash string,
	isBlocked bool,
	blockedAt *time.Time,
	blockReason string,
) *pgxmock.Rows {
	now := time.Now()
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	blockedAtPg := pgtype.Timestamp{Valid: false}
	if blockedAt != nil {
		blockedAtPg = pgtype.Timestamp{Time: *blockedAt, Valid: true}
	}
	blockReasonPg := pgtype.Text{Valid: false}
	if blockReason != "" {
		blockReasonPg = pgtype.Text{String: blockReason, Valid: true}
	}

	return pgxmock.NewRows([]string{
		"id", "role", "username", "email", "password_hash", "phone", "registered_at",
		"birth_date", "country_of_residence", "avatar_url", "favorite_categories",
		"is_blocked", "blocked_at", "block_reason", "email_verified", "phone_verified",
		"two_factor_enabled", "two_factor_secret", "backup_codes",
	}).AddRow(
		pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		"USER",
		username,
		email,
		passwordHash,
		pgtype.Text{Valid: false},
		pgtype.Timestamp{Time: now, Valid: true},
		pgtype.Date{Time: birthDate, Valid: true},
		pgtype.Text{Valid: false},
		pgtype.Text{String: "test-avatar-id", Valid: true},
		[]string{},
		pgtype.Bool{Bool: isBlocked, Valid: true},
		blockedAtPg,
		blockReasonPg,
		pgtype.Bool{Bool: true, Valid: true},
		pgtype.Bool{Bool: true, Valid: true},
		pgtype.Bool{Bool: false, Valid: true},
		pgtype.Text{Valid: false},
		[]string{},
	)
}

const (
	sqlMarkerFindUserByEmail     = `FindUserByEmail`
	sqlMarkerSaveUser            = `SaveUser`
	sqlMarkerUpdateEmailVerified = `UpdateEmailVerified`
	sqlMarkerUpdatePhoneVerified = `UpdatePhoneVerified`
	sqlMarkerFindUserByID        = `FindUserByID`
	sqlMarkerUpdateUser          = `UpdateUser`
	sqlMarkerFindUserByUsername  = `FindUserByUsername`
)

type noopNotifier struct{}

func (n *noopNotifier) Send(ctx context.Context, msg interfaces.NotifyMessage) error { return nil }
func (n *noopNotifier) SendTemplate(ctx context.Context, to, templateName string, data interface{}) error {
	return nil
}
func (n *noopNotifier) SendWelcomeEmail(ctx context.Context, to, username string) error { return nil }
func (n *noopNotifier) SendUserBlockedNotification(ctx context.Context, to, username, reason string) error {
	return nil
}
func (n *noopNotifier) SendUserUnblockedNotification(ctx context.Context, to, username string) error {
	return nil
}
func (n *noopNotifier) SendLoginNotification(ctx context.Context, to, username, ip, userAgent string) error {
	return nil
}
func (n *noopNotifier) SendVerificationEmail(ctx context.Context, to, code string) error { return nil }

type noopTaskQueue struct{}

func (q *noopTaskQueue) Publish(ctx context.Context, task interfaces.Task) error         { return nil }
func (q *noopTaskQueue) Subscribe(taskName string, handler interfaces.TaskHandler) error { return nil }
func (q *noopTaskQueue) Run(ctx context.Context) error                                   { return nil }
func (q *noopTaskQueue) Close() error                                                    { return nil }

type sessionRepoSpy struct {
	GetCalled, ExistsCalled, DeleteCalled, SaveCodeCalled, ValidateCodeCalled, DeleteCodeCalled bool
	CapturedToken                                                                               string
	CapturedUserID                                                                              domain.UserID
	CapturedDest                                                                                interface{}
}

func (s *sessionRepoSpy) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	s.CapturedToken = key
	if uid, ok := value.(domain.UserID); ok {
		s.CapturedUserID = uid
	}
	return nil
}
func (s *sessionRepoSpy) Get(ctx context.Context, key string, dest interface{}) error {
	s.GetCalled, s.CapturedToken, s.CapturedDest = true, key, dest
	if ptr, ok := dest.(*uuid.UUID); ok {
		*ptr = s.CapturedUserID
	}
	return nil
}
func (s *sessionRepoSpy) Delete(ctx context.Context, key string) error {
	s.DeleteCalled, s.CapturedToken = true, key
	return nil
}
func (s *sessionRepoSpy) Exists(ctx context.Context, key string) bool {
	s.ExistsCalled, s.CapturedToken = true, key
	return true
}
func (s *sessionRepoSpy) Save_code(ctx context.Context, userID domain.UserID, contact, purpose, code string, ttl time.Duration) error {
	s.SaveCodeCalled, s.CapturedUserID = true, userID
	return nil
}
func (s *sessionRepoSpy) Validate_code(ctx context.Context, userID domain.UserID, contact, purpose, code string) (bool, error) {
	s.ValidateCodeCalled, s.CapturedUserID = true, userID
	return true, nil
}
func (s *sessionRepoSpy) Delete_code(ctx context.Context, userID domain.UserID, contact, purpose string) error {
	s.DeleteCodeCalled, s.CapturedUserID = true, userID
	return nil
}
func (s *sessionRepoSpy) CreateTwoFactorChallenge(ctx context.Context, userID domain.UserID) (string, error) {
	return "", nil
}
func (s *sessionRepoSpy) GetUserIDByChallenge(ctx context.Context, challengeID string) (domain.UserID, error) {
	return uuid.Nil, nil
}
func (s *sessionRepoSpy) DeleteChallenge(ctx context.Context, challengeID string) error { return nil }

type failingValidateSessionRepo struct{ sessionRepoSpy }

func (f *failingValidateSessionRepo) Validate_code(ctx context.Context, userID domain.UserID, contact, purpose, code string) (bool, error) {
	return false, nil
}

type staticRepoMock struct {
	GetFileURLResult string
	GetFileURLError  error
}

func (m *staticRepoMock) Save(context.Context, *domain.StaticPage, io.Reader) error { return nil }
func (m *staticRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*domain.StaticPage, error) {
	return nil, nil
}
func (m *staticRepoMock) Update(context.Context, *domain.StaticPage, io.Reader) error { return nil }
func (m *staticRepoMock) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	return nil
}
func (m *staticRepoMock) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *staticRepoMock) DeleteFile(ctx context.Context, fileID string) error { return nil }
func (m *staticRepoMock) FindBySlug(ctx context.Context, slug string) (*domain.StaticPage, error) {
	return nil, nil
}
func (m *staticRepoMock) GetFile(context.Context, domain.ImageID) (io.ReadCloser, error) {
	return nil, nil
}
func (m *staticRepoMock) GetFileURL(ctx context.Context, fileID string) (string, error) {
	return m.GetFileURLResult, m.GetFileURLError
}

type rateLimiterSpy struct {
	AllowResult bool
	AllowError  error
}

func (r *rateLimiterSpy) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return r.AllowResult, r.AllowError
}

type mockSessionForUpdate struct{ getError error }

func (m *mockSessionForUpdate) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}
func (m *mockSessionForUpdate) Get(ctx context.Context, key string, dest interface{}) error {
	if m.getError != nil {
		return m.getError
	}
	return nil
}
func (m *mockSessionForUpdate) Delete(ctx context.Context, key string) error { return nil }
func (m *mockSessionForUpdate) Exists(ctx context.Context, key string) bool  { return true }
func (m *mockSessionForUpdate) Save_code(ctx context.Context, userID domain.UserID, contact, purpose, code string, ttl time.Duration) error {
	return nil
}
func (m *mockSessionForUpdate) Validate_code(ctx context.Context, userID domain.UserID, contact, purpose, code string) (bool, error) {
	return false, nil
}
func (m *mockSessionForUpdate) Delete_code(ctx context.Context, userID domain.UserID, contact, purpose string) error {
	return nil
}
func (m *mockSessionForUpdate) CreateTwoFactorChallenge(ctx context.Context, userID domain.UserID) (string, error) {
	return "", nil
}
func (m *mockSessionForUpdate) GetUserIDByChallenge(ctx context.Context, challengeID string) (domain.UserID, error) {
	return uuid.Nil, nil
}
func (m *mockSessionForUpdate) DeleteChallenge(ctx context.Context, challengeID string) error {
	return nil
}

type mockSessionForLogout struct{ existsResult bool }

func (m *mockSessionForLogout) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}
func (m *mockSessionForLogout) Get(ctx context.Context, key string, dest interface{}) error {
	return nil
}
func (m *mockSessionForLogout) Delete(ctx context.Context, key string) error { return nil }
func (m *mockSessionForLogout) Exists(ctx context.Context, key string) bool  { return m.existsResult }
func (m *mockSessionForLogout) Save_code(ctx context.Context, userID domain.UserID, contact, purpose, code string, ttl time.Duration) error {
	return nil
}
func (m *mockSessionForLogout) Validate_code(ctx context.Context, userID domain.UserID, contact, purpose, code string) (bool, error) {
	return false, nil
}
func (m *mockSessionForLogout) Delete_code(ctx context.Context, userID domain.UserID, contact, purpose string) error {
	return nil
}
func (m *mockSessionForLogout) CreateTwoFactorChallenge(ctx context.Context, userID domain.UserID) (string, error) {
	return "", nil
}
func (m *mockSessionForLogout) GetUserIDByChallenge(ctx context.Context, challengeID string) (domain.UserID, error) {
	return uuid.Nil, nil
}
func (m *mockSessionForLogout) DeleteChallenge(ctx context.Context, challengeID string) error {
	return nil
}

func TestRegisterUserUseCase_Classical(t *testing.T) {
	t.Run("Success_Register", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewRegisterUserUseCase(deps.userRepo, deps.sessionSpy, &noopNotifier{}, deps.taskSpy, deps.logger, deps.rateLimiter)

		deps.pgxMock.ExpectQuery(sqlMarkerFindUserByEmail).WithArgs(pgxmock.AnyArg()).WillReturnError(errors.New("no rows"))
		deps.pgxMock.ExpectQuery(sqlMarkerSaveUser).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(pgtype.UUID{Bytes: [16]byte(uuid.New()), Valid: true}))

		resp, err := uc.Execute(deps.ctx, defaultRegisterRequest())

		require.NoError(t, err)
		require.Equal(t, "verification codes sent", resp.Message)
		require.True(t, deps.sessionSpy.SaveCodeCalled)
		require.True(t, deps.taskSpy.PublishCalled)
	})

	t.Run("Error_RateLimited", func(t *testing.T) {
		deps := setupTestDeps(t)
		deps.rateLimiter.AllowResult = false
		uc := auth.NewRegisterUserUseCase(deps.userRepo, deps.sessionSpy, &noopNotifier{}, deps.taskSpy, deps.logger, deps.rateLimiter)

		req := defaultRegisterRequest()
		req.Email = "spam@example.com"

		resp, err := uc.Execute(deps.ctx, req)
		require.ErrorIs(t, err, usecase_errors.ErrTooManyAttempts)
		require.Nil(t, resp)
	})

	t.Run("Error_UserAlreadyExists", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewRegisterUserUseCase(deps.userRepo, deps.sessionSpy, &noopNotifier{}, deps.taskSpy, deps.logger, deps.rateLimiter)

		req := defaultRegisterRequest()
		req.Email = "exist@example.com"
		deps.pgxMock.ExpectQuery(sqlMarkerFindUserByEmail).WithArgs(pgxmock.AnyArg()).WillReturnRows(defaultMockUserRow(uuid.New()))

		resp, err := uc.Execute(deps.ctx, req)
		require.ErrorIs(t, err, usecase_errors.ErrUserAlreadyExists)
		require.Nil(t, resp)
	})
}

func TestVerifyContactUseCase_Classical(t *testing.T) {
	t.Run("Success_VerifyEmail", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewVerifyContactUseCase(deps.userRepo, deps.sessionSpy, &noopNotifier{}, deps.taskSpy, deps.logger)
		userID := uuid.New()

		deps.pgxMock.ExpectExec(sqlMarkerUpdateEmailVerified).
			WithArgs(pgxmock.AnyArg(), pgtype.Bool{Bool: true, Valid: true}).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := uc.Execute(deps.ctx, defaultVerifyContactRequest(userID))

		require.NoError(t, err)
		require.True(t, deps.sessionSpy.ValidateCodeCalled)
		require.True(t, deps.sessionSpy.DeleteCodeCalled)
	})

	t.Run("Error_InvalidCode", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewVerifyContactUseCase(deps.userRepo, &failingValidateSessionRepo{}, &noopNotifier{}, deps.taskSpy, deps.logger)

		req := defaultVerifyContactRequest(uuid.New())
		req.Code = "wrong"

		require.ErrorIs(t, uc.Execute(deps.ctx, req), usecase_errors.ErrInvalidVerificationCode)
	})
}

func TestGetUserProfileUseCase_Classical(t *testing.T) {
	t.Run("Success_GetProfile", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewGetUserProfileUseCase(deps.userRepo, deps.sessionSpy, deps.staticMock, deps.logger)
		userID := uuid.New()

		deps.pgxMock.ExpectQuery(sqlMarkerFindUserByID).WithArgs(pgxmock.AnyArg()).WillReturnRows(defaultMockUserRow(userID))

		resp, err := uc.Execute(deps.ctx, defaultGetProfileRequest(userID))

		require.NoError(t, err)
		require.Equal(t, "https://cdn.example.com/avatar.jpg", resp.AvatarURL)
		require.Equal(t, "testuser", resp.Username)
	})

	t.Run("Error_UserNotFound", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewGetUserProfileUseCase(deps.userRepo, deps.sessionSpy, deps.staticMock, deps.logger)

		deps.pgxMock.ExpectQuery(sqlMarkerFindUserByID).WithArgs(pgxmock.AnyArg()).WillReturnError(errors.New("no rows"))

		resp, err := uc.Execute(deps.ctx, defaultGetProfileRequest(uuid.New()))
		require.Error(t, err)
		require.Contains(t, err.Error(), "no rows")
		require.Nil(t, resp)
	})
}

func TestUpdateUserProfileUseCase_Classical(t *testing.T) {
	t.Run("Success_UpdateProfile", func(t *testing.T) {
		deps := setupTestDeps(t)
		userID := uuid.New()
		deps.sessionSpy.CapturedUserID = userID

		uc := auth.NewUpdateUserProfileUseCase(deps.userRepo, deps.sessionSpy, deps.logger)
		req := defaultUpdateProfileRequest("valid-token-123")

		deps.pgxMock.ExpectQuery(sqlMarkerFindUserByID).WithArgs(pgxmock.AnyArg()).WillReturnRows(defaultMockUserRow(userID))
		deps.pgxMock.ExpectQuery(sqlMarkerFindUserByUsername).WithArgs(pgxmock.AnyArg()).WillReturnError(errors.New("no rows"))
		deps.pgxMock.ExpectExec(sqlMarkerUpdateUser).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		resp, err := uc.Execute(deps.ctx, req)
		require.NoError(t, err)
		require.Equal(t, "updated_user", resp.Username)
	})

	t.Run("Error_InvalidSession", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewUpdateUserProfileUseCase(deps.userRepo, &mockSessionForUpdate{getError: errors.New("invalid")}, deps.logger)

		resp, err := uc.Execute(deps.ctx, defaultUpdateProfileRequest("bad-token"))
		require.ErrorIs(t, err, usecase_errors.ErrInvalidSession)
		require.Nil(t, resp)
	})

	t.Run("Error_UserBlocked", func(t *testing.T) {
		deps := setupTestDeps(t)
		userID := uuid.New()
		deps.sessionSpy.CapturedUserID = userID
		uc := auth.NewUpdateUserProfileUseCase(deps.userRepo, deps.sessionSpy, deps.logger)

		blockedTime := time.Now().Add(-time.Hour)
		deps.pgxMock.ExpectQuery(sqlMarkerFindUserByID).WithArgs(pgxmock.AnyArg()).
			WillReturnRows(buildMockUserRow(userID, "blocked_user", "test@example.com", "hash", true, &blockedTime, "Spam"))

		resp, err := uc.Execute(deps.ctx, defaultUpdateProfileRequest("token"))
		require.ErrorIs(t, err, usecase_errors.ErrUserBlocked)
		require.Nil(t, resp)
	})
}

func TestLogoutUseCase_Classical(t *testing.T) {
	t.Run("Success_Logout", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewLogoutUseCase(deps.sessionSpy, deps.logger)
		token := "session-token-xyz"

		resp, err := uc.Execute(deps.ctx, defaultLogoutRequest(token))

		require.NoError(t, err)
		require.True(t, resp.Success)
		require.True(t, deps.sessionSpy.ExistsCalled)
		require.True(t, deps.sessionSpy.DeleteCalled)
		require.Equal(t, token, deps.sessionSpy.CapturedToken)
	})

	t.Run("Error_InvalidSession", func(t *testing.T) {
		deps := setupTestDeps(t)
		uc := auth.NewLogoutUseCase(&mockSessionForLogout{existsResult: false}, deps.logger)

		resp, err := uc.Execute(deps.ctx, defaultLogoutRequest("fake-token"))
		require.ErrorIs(t, err, usecase_errors.ErrInvalidSession)
		require.Nil(t, resp)
	})
}
