package redis

import (
	"app/internal/domain"
	"app/internal/repository/cache"
	"app/internal/repository/errs"
	"app/pkg/logger"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
)

type statsCache struct {
	redisClient *redis.Client
	logger      logger.Logger
}

func NewStatsCache(redisClient *redis.Client, logger logger.Logger) cache.StatsCache {
	return &statsCache{
		logger:      logger,
		redisClient: redisClient,
	}
}

func (s *statsCache) DecrementAssignCountByUserID(ctx context.Context, userID domain.UserID) error {
	_, err := s.redisClient.Decr(ctx, userID.String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.logger.Errorw("User not found in Redis when decrementing assign count", "userID", userID)
			return errs.ErrNotFound
		}
		s.logger.Errorw("Failed to decrement assign count", "userID", userID, "error", err)
		return err
	}
	return nil
}

func (s *statsCache) GetAssignCountByUserID(ctx context.Context, userID domain.UserID) (int, error) {
	count, err := s.redisClient.Get(ctx, userID.String()).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.logger.Errorw("User not found in Redis when getting assign count", "userID", userID)
			return 0, errs.ErrNotFound
		}
		s.logger.Errorw("Failed to get assign count", "userID", userID, "error", err)
		return 0, err
	}
	return count, nil
}

func (s *statsCache) IncrementAssignCountByUserID(ctx context.Context, userID domain.UserID) error {
	_, err := s.redisClient.Incr(ctx, userID.String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.logger.Errorw("User not found in Redis when incrementing assign count", "userID", userID)
			return errs.ErrNotFound
		}
		s.logger.Errorw("Failed to increment assign count", "userID", userID, "error", err)
		return err
	}
	return nil
}

func (s *statsCache) SetAssignCountByUserID(ctx context.Context, userID domain.UserID, count int) error {
	err := s.redisClient.Set(ctx, userID.String(), count, 0).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.logger.Errorw("User not found in Redis when setting assign count", "userID", userID)
			return errs.ErrNotFound
		}
		s.logger.Errorw("Failed to set assign count", "userID", userID, "error", err)
		return err
	}
	return nil
}


