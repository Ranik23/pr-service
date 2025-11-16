package stats_usecase

import (
	"app/internal/domain"
	"app/internal/repository/cache"
	repoerrs "app/internal/repository/errs"
	"app/internal/usecase/errs"
	"context"
	"errors"
)

type StatsUseCase interface {
	GetAssignCountByUserID(ctx context.Context, userID domain.UserID) (*domain.UserStats, error)
}

type statsUseCase struct {
	statsCache cache.StatsCache
}

func NewStatsUseCase(statsCache cache.StatsCache) StatsUseCase {
	return &statsUseCase{
		statsCache: statsCache,	
	}
}

func (s *statsUseCase) GetAssignCountByUserID(ctx context.Context, userID domain.UserID) (*domain.UserStats, error) {
	count, err := s.statsCache.GetAssignCountByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	
	return &domain.UserStats{
		UserID: 	   userID,
		AssignedCount: count,
	}, nil
}

