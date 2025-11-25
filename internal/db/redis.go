package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/TOPfiIT/auth-service/internal/config"
	"github.com/TOPfiIT/auth-service/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func InitRedisClient(cfg *config.Config) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.RedisAddr(),
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Failed to connect to redis:", err)
	}

	return &Redis{
		client: client,
	}
}

func (c *Redis) SaveSession(ctx context.Context, companyID uuid.UUID, session *models.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("[Redis] marshal Session: %w", err)
	}

	key := fmt.Sprintf("session:%s", companyID)
	ttl := time.Until(session.ExpiresAt)

	if ttl <= 0 {
		return fmt.Errorf("[Redis] session already expired")
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *Redis) GetSession(ctx context.Context, companyID uuid.UUID) (*models.Session, error) {
	key := fmt.Sprintf("session:%s", companyID)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("[Redis] get Session: %w", err)
	}

	var session models.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("[Redis] unmarshal Session: %w", err)
	}

	return &session, nil
}

func (c *Redis) DeleteSession(ctx context.Context, companyID uuid.UUID) error {
	key := fmt.Sprintf("session:%s", companyID)
	return c.client.Del(ctx, key).Err()
}

func (c *Redis) SaveRefreshToken(ctx context.Context, companyID uuid.UUID, refreshToken string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:%s", companyID)
	return c.client.Set(ctx, key, refreshToken, ttl).Err()
}

func (c *Redis) GetRefreshToken(ctx context.Context, companyID uuid.UUID) (string, error) {
	key := fmt.Sprintf("refresh:%s", companyID)
	return c.client.Get(ctx, key).Result()
}

func (c *Redis) DeleteRefreshToken(ctx context.Context, companyID uuid.UUID) error {
	key := fmt.Sprintf("refresh:%s", companyID)
	return c.client.Del(ctx, key).Err()
}

func (c *Redis) Close() error {
	return c.client.Close()
}
