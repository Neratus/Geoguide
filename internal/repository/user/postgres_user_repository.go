package user_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	postgreSQL "github.com/Neratus/geoguide/internal/repository/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Close()
}

type PostgresUserRepo struct {
	Pool    DBPool
	Queries *postgreSQL.Queries
}

func NewPostgresUserRepo(cfg *config.Config) (*PostgresUserRepo, error) {
	connStr := cfg.PostgresConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresUserRepo{
		Pool:    pool,
		Queries: postgreSQL.New(pool),
	}, nil
}

func (r *PostgresUserRepo) Close() error {
	if r.Pool != nil {
		r.Pool.Close()
	}
	return nil
}

func (r *PostgresUserRepo) Save(ctx context.Context, user *domain.User) error {
	userParams := toSaveUserParams(user)
	id, err := r.Queries.SaveUser(ctx, userParams)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	user.SetId(domain.FromPgUUID(id))
	return err
}

func (r *PostgresUserRepo) FindAll(ctx context.Context, limit, offset int, search string) ([]*domain.User, error) {
	rows, err := r.Queries.FindAllUsers(ctx, postgreSQL.FindAllUsersParams{
		Column1: search,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, 0, len(rows))
	for _, row := range rows {
		user, err := toDomainUser(row)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *PostgresUserRepo) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	row, err := r.Queries.FindUserByID(ctx, domain.ToPgUUID(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainUser(row)
}

func (r *PostgresUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	dbuser, err := r.Queries.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	return toDomainUser(dbuser)
}

func (r *PostgresUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	dbuser, err := r.Queries.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return toDomainUser(dbuser)
}

func (r *PostgresUserRepo) Update(ctx context.Context, user *domain.User) error {
	var birthDate pgtype.Date
	if bd := user.GetBirthDate(); bd != nil {
		birthDate = pgtype.Date{Time: *bd, Valid: true}
	}
	params := postgreSQL.UpdateUserParams{
		ID:                 domain.ToPgUUID(user.GetId()),
		Role:               user.GetRole(),
		Username:           user.GetUsername(),
		Email:              user.GetEmail(),
		PasswordHash:       user.GetPasswordHash(),
		Phone:              pgtype.Text{String: user.GetPhone(), Valid: user.GetPhone() != ""},
		BirthDate:          birthDate,
		CountryOfResidence: pgtype.Text{String: user.GetCountryOfResidence(), Valid: user.GetCountryOfResidence() != ""},
		AvatarUrl:          pgtype.Text{String: user.GetAvatarUrl(), Valid: user.GetAvatarUrl() != ""},
		FavoriteCategories: user.GetFavoriteCategories(),
	}
	err := r.Queries.UpdateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) Delete(ctx context.Context, id domain.UserID) error {
	err := r.Queries.DeleteUser(ctx, domain.ToPgUUID(id))
	return err
}

func (r *PostgresUserRepo) BlockUser(ctx context.Context, id domain.UserID, reason string) error {
	params := postgreSQL.BlockUserParams{
		ID:          domain.ToPgUUID(id),
		BlockReason: pgtype.Text{String: reason, Valid: true},
	}
	err := r.Queries.BlockUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) UnblockUser(ctx context.Context, id domain.UserID) error {
	err := r.Queries.UnblockUser(ctx, domain.ToPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) UpdateEmailVerified(ctx context.Context, userID domain.UserID, verified bool) error {
	err := r.Queries.UpdateEmailVerified(ctx, postgreSQL.UpdateEmailVerifiedParams{
		ID:            domain.ToPgUUID(userID),
		EmailVerified: pgtype.Bool{Bool: verified, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update email verified: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) GetUserActivityReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.UserActivityData, error) {
	params := postgreSQL.GetUserActivityReportParams{
		DateFrom: pgtype.Text{String: dateFrom, Valid: true},
		DateTo:   pgtype.Text{String: dateTo, Valid: true},
		Limit:    int32(limit),
	}
	rows, err := r.Queries.GetUserActivityReport(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get user activity report: %w", err)
	}

	result := make([]domain.UserActivityData, 0, len(rows))
	for _, row := range rows {
		var lastActive time.Time
		switch v := row.LastActive.(type) {
		case time.Time:
			lastActive = v
		case pgtype.Timestamp:
			if v.Valid {
				lastActive = v.Time
			}
		}

		result = append(result, domain.UserActivityData{
			UserID:         domain.FromPgUUID(row.ID),
			Username:       row.Username,
			TripsCreated:   int(row.TripsCreated),
			ReviewsWritten: int(row.ReviewsWritten),
			LastActive:     lastActive,
		})
	}
	return result, nil
}

func (r *PostgresUserRepo) UpdatePhoneVerified(ctx context.Context, userID domain.UserID, verified bool) error {
	err := r.Queries.UpdatePhoneVerified(ctx, postgreSQL.UpdatePhoneVerifiedParams{
		ID:            domain.ToPgUUID(userID),
		PhoneVerified: pgtype.Bool{Bool: verified, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update phone verified: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) EnableTwoFactor(ctx context.Context, userID domain.UserID, secret string, backupCodes []string) error {
	err := r.Queries.EnableTwoFactor(ctx, postgreSQL.EnableTwoFactorParams{
		ID:              domain.ToPgUUID(userID),
		TwoFactorSecret: pgtype.Text{String: secret, Valid: true},
		BackupCodes:     backupCodes,
	})
	if err != nil {
		return fmt.Errorf("failed to enable two factor: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) DisableTwoFactor(ctx context.Context, userID domain.UserID) error {
	err := r.Queries.DisableTwoFactor(ctx, domain.ToPgUUID(userID))
	if err != nil {
		return fmt.Errorf("failed to disable two factor: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) UpdateAvatar(ctx context.Context, userID domain.UserID, avatarURL string) error {
	_, err := r.Pool.Exec(ctx, `UPDATE "User" SET avatar_url = $1 WHERE id = $2`, avatarURL, userID)
	return err
}

func toSaveUserParams(user *domain.User) postgreSQL.SaveUserParams {
	var birthDate pgtype.Date
	if bd := user.GetBirthDate(); bd != nil {
		birthDate = pgtype.Date{Time: *bd, Valid: true}
	}
	return postgreSQL.SaveUserParams{
		Role:               user.GetRole(),
		Username:           user.GetUsername(),
		Email:              user.GetEmail(),
		PasswordHash:       user.GetPasswordHash(),
		Phone:              pgtype.Text{String: user.GetPhone(), Valid: user.GetPhone() != ""},
		BirthDate:          birthDate,
		CountryOfResidence: pgtype.Text{String: user.GetCountryOfResidence(), Valid: user.GetCountryOfResidence() != ""},
		AvatarUrl:          pgtype.Text{String: user.GetAvatarUrl(), Valid: user.GetAvatarUrl() != ""},
		FavoriteCategories: user.GetFavoriteCategories(),
	}
}

func toDomainUser(dbUser postgreSQL.User) (*domain.User, error) {
	id := domain.FromPgUUID(dbUser.ID)
	if id == uuid.Nil {
		return nil, errors.New("user id is null")
	}

	var birthDate *time.Time
	if dbUser.BirthDate.Valid {
		birthDate = &dbUser.BirthDate.Time
	}

	var blockedAt *time.Time
	if dbUser.BlockedAt.Valid {
		blockedAt = &dbUser.BlockedAt.Time
	}

	emailVerified := dbUser.EmailVerified.Bool
	phoneVerified := dbUser.PhoneVerified.Bool
	twoFactorEnabled := dbUser.TwoFactorEnabled.Bool
	twoFactorSecret := ""
	if dbUser.TwoFactorSecret.Valid {
		twoFactorSecret = dbUser.TwoFactorSecret.String
	}
	backupCodes := dbUser.BackupCodes

	return domain.NewUserFromDB(
		id,
		dbUser.Username,
		dbUser.Email,
		dbUser.PasswordHash,
		dbUser.Phone.String,
		dbUser.RegisteredAt.Time,
		birthDate,
		dbUser.CountryOfResidence.String,
		dbUser.AvatarUrl.String,
		dbUser.FavoriteCategories,
		dbUser.IsBlocked.Bool,
		blockedAt,
		dbUser.BlockReason.String,
		dbUser.Role,
		emailVerified,
		phoneVerified,
		twoFactorEnabled,
		twoFactorSecret,
		backupCodes,
	), nil
}
