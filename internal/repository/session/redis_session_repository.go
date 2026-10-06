package session_repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	config "github.com/Neratus/geoguide/internal/repository/config"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisSessionRepo struct {
	client *redis.Client
	ctx    context.Context
}

func NewSessionRepoWithClient(client *redis.Client) *RedisSessionRepo {
	return &RedisSessionRepo{
		client: client,
		ctx:    context.Background(),
	}
}

func NewSessionRepo(cfg *config.Config) (*RedisSessionRepo, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Database.Redis.Host, cfg.Database.Redis.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Database.Redis.Password,
		DB:       cfg.Database.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisSessionRepo{client: client, ctx: context.Background()}, nil
}

func (r *RedisSessionRepo) Close() error {
	return r.client.Close()
}

func (r *RedisSessionRepo) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	switch v := value.(type) {
	case domain.UserID:
		value = uuid.UUID(v)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, ttl).Err()
}

func (r *RedisSessionRepo) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("key not found")
		}
		return err
	}
	if ptr, ok := dest.(*domain.UserID); ok {
		var u uuid.UUID
		if err := json.Unmarshal(data, &u); err != nil {
			return err
		}
		*ptr = domain.UserID(u)
		return nil
	}

	return json.Unmarshal(data, dest)
}

func (r *RedisSessionRepo) Delete(ctx context.Context, key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisSessionRepo) Exists(ctx context.Context, key string) bool {
	val, err := r.client.Exists(r.ctx, key).Result()
	return err == nil && val > 0
}

func (r *RedisSessionRepo) Save_code(ctx context.Context, userID domain.UserID, contact, purpose, code string, ttl time.Duration) error {
	key := fmt.Sprintf("verif:%s:%s:%s", userID.String(), purpose, contact)

	err := r.client.Set(ctx, key, code, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save verification code to redis: %w", err)
	}

	return nil
}

func (r *RedisSessionRepo) Validate_code(ctx context.Context, userID domain.UserID, contact, purpose, code string) (bool, error) {
	key := fmt.Sprintf("verif:%s:%s:%s", userID.String(), purpose, contact)

	storedCode, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get verification code from redis: %w", err)
	}

	return storedCode == code, nil
}

func (r *RedisSessionRepo) Delete_code(ctx context.Context, userID domain.UserID, contact, purpose string) error {
	key := fmt.Sprintf("verif:%s:%s:%s", userID.String(), purpose, contact)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return fmt.Errorf("failed to delete verification code from redis: %w", err)
		}
	}

	return nil
}

func (r *RedisSessionRepo) CreateTwoFactorChallenge(ctx context.Context, userID domain.UserID) (string, error) {
	challengeID := uuid.New().String()
	key := fmt.Sprintf("2fa_challenge:%s", challengeID)
	err := r.client.Set(ctx, key, userID.String(), 5*time.Minute).Err()
	if err != nil {
		return "", err
	}
	return challengeID, nil
}

func (r *RedisSessionRepo) GetUserIDByChallenge(ctx context.Context, challengeID string) (domain.UserID, error) {
	key := fmt.Sprintf("2fa_challenge:%s", challengeID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return uuid.Nil, errors.New("challenge not found or expired")
	}
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(val)
}

func (r *RedisSessionRepo) DeleteChallenge(ctx context.Context, challengeID string) error {
	key := fmt.Sprintf("2fa_challenge:%s", challengeID)
	return r.client.Del(ctx, key).Err()
}
