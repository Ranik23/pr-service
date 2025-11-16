package redis

import (
	"app/internal/domain"
	"app/internal/repository/cache"
	"app/internal/repository/errs"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
)

type statsCache struct {
	redisClient *redis.Client
}

func NewStatsCache(redisClient *redis.Client) cache.StatsCache {
	return &statsCache{
		redisClient: redisClient,
	}
}

func (s *statsCache) DecrementAssignCountByUserID(ctx context.Context, userID domain.UserID) error {
	_, err := s.redisClient.Decr(ctx, userID.String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errs.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *statsCache) GetAssignCountByUserID(ctx context.Context, userID domain.UserID) (int, error) {
	count, err := s.redisClient.Get(ctx, userID.String()).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, errs.ErrNotFound
		}
		return 0, err
	}
	return count, nil
}

func (s *statsCache) IncrementAssignCountByUserID(ctx context.Context, userID domain.UserID) error {
	_, err := s.redisClient.Incr(ctx, userID.String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errs.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *statsCache) SetAssignCountByUserID(ctx context.Context, userID domain.UserID, count int) error {
	err := s.redisClient.Set(ctx, userID.String(), count, 0).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errs.ErrNotFound
		}
		return err
	}
	return nil
}


