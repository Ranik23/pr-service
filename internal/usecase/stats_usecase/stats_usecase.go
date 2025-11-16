package stats_usecase

import (
	"app/internal/domain"
	"app/internal/repository/cache"
	repoerrs "app/internal/repository/errs"
	"app/internal/repository/storage"
	"app/internal/usecase/errs"
	"context"
	"errors"
)

type StatsUseCase interface {
	GetAssignCountByUserID(ctx context.Context, userID domain.UserID) (*domain.UserStats, error)
}

type statsUseCase struct {
	statsCache  cache.StatsCache
	userStorage storage.UserStorage
}

func NewStatsUseCase(statsCache cache.StatsCache, userStorage storage.UserStorage) StatsUseCase {
	return &statsUseCase{
		statsCache:  statsCache,
		userStorage: userStorage,	
	}
}

func (s *statsUseCase) GetAssignCountByUserID(ctx context.Context, userID domain.UserID) (*domain.UserStats, error) {

	_, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}

	count, err := s.statsCache.GetAssignCountByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return &domain.UserStats{UserID: userID, AssignedCount: 0}, nil
		}
		return nil, err
	}
	
	return &domain.UserStats{
		UserID: 	   userID,
		AssignedCount: count,
	}, nil
}

