package session_repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	"github.com/google/uuid"
	"github.com/tarantool/go-tarantool"
)

type TarantoolSessionRepo struct {
	conn  *tarantool.Connection
	space string
}

func NewTarantoolSessionRepo(cfg *config.Config) (*TarantoolSessionRepo, error) {
	cfgTar := cfg.Database.Tarantool
	if len(cfgTar.Hosts) == 0 {
		return nil, errors.New("no tarantool hosts provided")
	}
	opts := tarantool.Opts{
		User: cfgTar.Username,
		Pass: cfgTar.Password,
	}
	conn, err := tarantool.Connect(cfgTar.Hosts[0], opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tarantool: %w", err)
	}
	return &TarantoolSessionRepo{
		conn:  conn,
		space: cfgTar.Space,
	}, nil
}

func (r *TarantoolSessionRepo) Close() error {
	return r.conn.Close()
}

func (r *TarantoolSessionRepo) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(ttl).Unix()
	_, err = r.conn.Replace(r.space, []interface{}{key, string(data), expiresAt})
	return err
}

func (r *TarantoolSessionRepo) Get(ctx context.Context, key string, dest interface{}) error {
	resp, err := r.conn.Select(r.space, "primary", 0, 1, tarantool.IterEq, []interface{}{key})
	if err != nil {
		return fmt.Errorf("failed to select key %s: %w", key, err)
	}
	if len(resp.Tuples()) == 0 {
		return errors.New("key not found")
	}
	tuple := resp.Tuples()[0]
	if len(tuple) < 3 {
		return errors.New("invalid tuple format")
	}
	expiresAt, ok := tuple[2].(uint64)
	if !ok {
		return errors.New("expires_at field is not uint64")
	}
	if time.Now().Unix() > int64(expiresAt) {
		_ = r.Delete(ctx, key)
		return errors.New("key not found")
	}
	valueStr, ok := tuple[1].(string)
	if !ok {
		return errors.New("value field is not string")
	}
	return json.Unmarshal([]byte(valueStr), dest)
}

func (r *TarantoolSessionRepo) Delete(ctx context.Context, key string) error {
	_, err := r.conn.Delete(r.space, "primary", []interface{}{key})
	if err != nil {
		if tarantoolErr, ok := err.(*tarantool.Error); ok && tarantoolErr.Code == 44 {
			return nil
		}
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

func (r *TarantoolSessionRepo) Exists(ctx context.Context, key string) bool {
	var dummy string
	err := r.Get(ctx, key, &dummy)
	return err == nil
}

func (r *TarantoolSessionRepo) Save_code(ctx context.Context, userID domain.UserID, contact, purpose, code string, ttl time.Duration) error {
	key := fmt.Sprintf("verif:%s:%s:%s", userID.String(), purpose, contact)
	return r.Set(ctx, key, code, ttl)
}

func (r *TarantoolSessionRepo) Validate_code(ctx context.Context, userID domain.UserID, contact, purpose, code string) (bool, error) {
	key := fmt.Sprintf("verif:%s:%s:%s", userID.String(), purpose, contact)
	var storedCode string
	err := r.Get(ctx, key, &storedCode)
	if err != nil {
		if err.Error() == "key not found" {
			return false, nil
		}
		return false, err
	}
	return storedCode == code, nil
}

func (r *TarantoolSessionRepo) Delete_code(ctx context.Context, userID domain.UserID, contact, purpose string) error {
	key := fmt.Sprintf("verif:%s:%s:%s", userID.String(), purpose, contact)
	return r.Delete(ctx, key)
}

func (r *TarantoolSessionRepo) CreateTwoFactorChallenge(ctx context.Context, userID domain.UserID) (string, error) {
	challengeID := uuid.New().String()
	key := fmt.Sprintf("2fa_challenge:%s", challengeID)
	err := r.Set(ctx, key, userID.String(), 5*time.Minute)
	if err != nil {
		return "", err
	}
	return challengeID, nil
}

func (r *TarantoolSessionRepo) GetUserIDByChallenge(ctx context.Context, challengeID string) (domain.UserID, error) {
	key := fmt.Sprintf("2fa_challenge:%s", challengeID)
	var userIDStr string
	err := r.Get(ctx, key, &userIDStr)
	if err != nil {
		if err.Error() == "key not found" {
			return uuid.Nil, errors.New("challenge not found or expired")
		}
		return uuid.Nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (r *TarantoolSessionRepo) DeleteChallenge(ctx context.Context, challengeID string) error {
	key := fmt.Sprintf("2fa_challenge:%s", challengeID)
	return r.Delete(ctx, key)
}
