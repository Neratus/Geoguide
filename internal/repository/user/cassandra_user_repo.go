package user_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type CassandraUserRepo struct {
	session *gocql.Session
}

func NewCassandraUserRepo(session *gocql.Session) *CassandraUserRepo {
	return &CassandraUserRepo{session: session}
}

func (r *CassandraUserRepo) Close() error {
	return nil
}

func (r *CassandraUserRepo) Save(ctx context.Context, user *domain.User) error {
	id := user.GetId()
	if id == uuid.Nil {
		id = domain.UserID(uuid.New())
		user.SetId(id)
	}
	idq := gocql.UUID(id)
	username := user.GetUsername()
	email := user.GetEmail()

	queryID := r.session.Query(`
        INSERT INTO user_by_id (
            id, role, username, email, password_hash, phone, registered_at, birth_date,
            country_of_residence, avatar_url, favorite_categories, is_blocked, blocked_at,
            block_reason, email_verified, phone_verified, two_factor_enabled,
            two_factor_secret, backup_codes
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
		idq, user.GetRole(), username, email, user.GetPasswordHash(),
		user.GetPhone(), user.GetRegisteredAt(), user.GetBirthDate(),
		user.GetCountryOfResidence(), user.GetAvatarUrl(), user.GetFavoriteCategories(),
		user.IsBlocked(), user.GetBlockedAt(), user.GetBlockReason(),
		user.IsEmailVerified(), user.IsPhoneVerified(), user.IsTwoFactorEnabled(),
		user.GetTwoFactorSecret(), user.GetBackupCodes(),
	)
	if err := queryID.WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to insert into user_by_id: %w", err)
	}

	queryUsername := r.session.Query(`
        INSERT INTO user_by_username (username, id) VALUES (?, ?)
    `, username, idq)
	if err := queryUsername.WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM user_by_id WHERE id = ?`, idq).WithContext(ctx).Exec()
		return fmt.Errorf("failed to insert into user_by_username: %w", err)
	}

	queryEmail := r.session.Query(`
        INSERT INTO user_by_email (email, id) VALUES (?, ?)
    `, email, idq)
	if err := queryEmail.WithContext(ctx).Exec(); err != nil {
		_ = r.session.Query(`DELETE FROM user_by_id WHERE id = ?`, idq).WithContext(ctx).Exec()
		_ = r.session.Query(`DELETE FROM user_by_username WHERE username = ?`, username).WithContext(ctx).Exec()
		return fmt.Errorf("failed to insert into user_by_email: %w", err)
	}
	return nil
}

func (r *CassandraUserRepo) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	var (
		userID                                         gocql.UUID
		role, username, email, passwordHash, phone     string
		registeredAt                                   time.Time
		birthDate                                      *time.Time
		countryOfResidence, avatarUrl                  string
		favoriteCategories                             []string
		isBlocked                                      bool
		blockedAt                                      *time.Time
		blockReason                                    string
		emailVerified, phoneVerified, twoFactorEnabled bool
		twoFactorSecret                                string
		backupCodes                                    []string
	)
	query := r.session.Query(`
        SELECT id, role, username, email, password_hash, phone, registered_at, birth_date,
               country_of_residence, avatar_url, favorite_categories, is_blocked, blocked_at,
               block_reason, email_verified, phone_verified, two_factor_enabled,
               two_factor_secret, backup_codes
        FROM user_by_id WHERE id = ? LIMIT 1
    `, gocql.UUID(id))
	iter := query.WithContext(ctx).Iter()
	if !iter.Scan(
		&userID, &role, &username, &email, &passwordHash, &phone, &registeredAt, &birthDate,
		&countryOfResidence, &avatarUrl, &favoriteCategories, &isBlocked, &blockedAt,
		&blockReason, &emailVerified, &phoneVerified, &twoFactorEnabled,
		&twoFactorSecret, &backupCodes,
	) {
		if err := iter.Close(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	user := domain.NewUserFromDB(
		domain.UserID(userID),
		username, email, passwordHash, phone,
		registeredAt,
		birthDate,
		countryOfResidence, avatarUrl,
		favoriteCategories,
		isBlocked,
		blockedAt,
		blockReason,
		role,
		emailVerified, phoneVerified, twoFactorEnabled,
		twoFactorSecret,
		backupCodes,
	)
	return user, nil
}

func (r *CassandraUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var id gocql.UUID
	query := r.session.Query(`SELECT id FROM user_by_username WHERE username = ? LIMIT 1`, username)
	if err := query.WithContext(ctx).Scan(&id); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("failed to find user by username: %w", err)
		}
		return nil, err
	}
	return r.FindByID(ctx, domain.UserID(id))
}

func (r *CassandraUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var id gocql.UUID
	query := r.session.Query(`SELECT id FROM user_by_email WHERE email = ? LIMIT 1`, email)
	if err := query.WithContext(ctx).Scan(&id); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("failed to find user by email: %w", err)
		}
		return nil, err
	}
	return r.FindByID(ctx, domain.UserID(id))
}

func (r *CassandraUserRepo) Update(ctx context.Context, user *domain.User) error {
	queryID := r.session.Query(`
		UPDATE user_by_id SET
			role = ?, username = ?, email = ?, password_hash = ?, phone = ?,
			birth_date = ?, country_of_residence = ?, avatar_url = ?, favorite_categories = ?
		WHERE id = ?
	`, user.GetRole(), user.GetUsername(), user.GetEmail(), user.GetPasswordHash(),
		user.GetPhone(), user.GetBirthDate(), user.GetCountryOfResidence(),
		user.GetAvatarUrl(), user.GetFavoriteCategories(), gocql.UUID(user.GetId()))
	if err := queryID.WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to update user_by_id: %w", err)
	}

	return nil
}

func (r *CassandraUserRepo) Delete(ctx context.Context, id domain.UserID) error {
	user, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	if err := r.session.Query(`DELETE FROM user_by_id WHERE id = ?`, gocql.UUID(id)).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM user_by_username WHERE username = ?`, user.GetUsername()).WithContext(ctx).Exec(); err != nil {
		return err
	}
	if err := r.session.Query(`DELETE FROM user_by_email WHERE email = ?`, user.GetEmail()).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *CassandraUserRepo) BlockUser(ctx context.Context, id domain.UserID, reason string) error {
	now := time.Now()
	query := r.session.Query(`
		UPDATE user_by_id SET is_blocked = true, blocked_at = ?, block_reason = ? WHERE id = ?
	`, now, reason, gocql.UUID(id))
	return query.WithContext(ctx).Exec()
}

func (r *CassandraUserRepo) UnblockUser(ctx context.Context, id domain.UserID) error {
	query := r.session.Query(`
		UPDATE user_by_id SET is_blocked = false, blocked_at = null, block_reason = null WHERE id = ?
	`, gocql.UUID(id))
	return query.WithContext(ctx).Exec()
}

func (r *CassandraUserRepo) UpdateEmailVerified(ctx context.Context, userID domain.UserID, verified bool) error {
	query := r.session.Query(`UPDATE user_by_id SET email_verified = ? WHERE id = ?`, verified, gocql.UUID(userID))
	return query.WithContext(ctx).Exec()
}

func (r *CassandraUserRepo) UpdatePhoneVerified(ctx context.Context, userID domain.UserID, verified bool) error {
	query := r.session.Query(`UPDATE user_by_id SET phone_verified = ? WHERE id = ?`, verified, gocql.UUID(userID))
	return query.WithContext(ctx).Exec()
}

func (r *CassandraUserRepo) EnableTwoFactor(ctx context.Context, userID domain.UserID, secret string, backupCodes []string) error {
	query := r.session.Query(`
		UPDATE user_by_id SET two_factor_enabled = true, two_factor_secret = ?, backup_codes = ?
		WHERE id = ?
	`, secret, backupCodes, gocql.UUID(userID))
	return query.WithContext(ctx).Exec()
}

func (r *CassandraUserRepo) DisableTwoFactor(ctx context.Context, userID domain.UserID) error {
	query := r.session.Query(`
		UPDATE user_by_id SET two_factor_enabled = false, two_factor_secret = null, backup_codes = null
		WHERE id = ?
	`, gocql.UUID(userID))
	return query.WithContext(ctx).Exec()
}

func (r *CassandraUserRepo) UpdateAvatar(ctx context.Context, userID domain.UserID, avatarURL string) error {
	query := r.session.Query(`UPDATE user_by_id SET avatar_url = ? WHERE id = ?`, avatarURL, gocql.UUID(userID))
	return query.WithContext(ctx).Exec()
}

func (r *CassandraUserRepo) FindAll(ctx context.Context, limit, offset int, search string) ([]*domain.User, error) {
	var users []*domain.User
	query := r.session.Query(`SELECT id FROM user_by_id`)
	iter := query.WithContext(ctx).Iter()
	var id gocql.UUID
	for iter.Scan(&id) {
		user, err := r.FindByID(ctx, domain.UserID(id))
		if err != nil {
			continue
		}
		if user != nil {
			users = append(users, user)
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *CassandraUserRepo) GetUserActivityReport(ctx context.Context, limit int, dateFrom, dateTo string) ([]domain.UserActivityData, error) {
	users, err := r.FindAll(ctx, 0, 0, "")
	if err != nil {
		return nil, err
	}
	result := make([]domain.UserActivityData, 0, len(users))
	for _, u := range users {
		tripsCount := 0
		reviewsCount := 0
		var count int
		if err := r.session.Query(`SELECT COUNT(*) FROM trip_by_user WHERE user_id = ?`, gocql.UUID(u.GetId())).WithContext(ctx).Scan(&count); err == nil {
			tripsCount = count
		}
		if err := r.session.Query(`SELECT COUNT(*) FROM review_by_user WHERE user_id = ?`, gocql.UUID(u.GetId())).WithContext(ctx).Scan(&count); err == nil {
			reviewsCount = count
		}
		result = append(result, domain.UserActivityData{
			UserID:         u.GetId(),
			Username:       u.GetUsername(),
			TripsCreated:   tripsCount,
			ReviewsWritten: reviewsCount,
			LastActive:     time.Now(),
		})
		if len(result) >= limit && limit > 0 {
			break
		}
	}
	return result, nil
}
