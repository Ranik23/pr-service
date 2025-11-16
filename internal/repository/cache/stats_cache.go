package cache

import (
	"context"
	
	"app/internal/domain"
)

//go:generate mockgen -source=stats_cache.go -destination=mock/stats_cache_mock.go -package=mock
type StatsCache interface {
	GetAssignCountByUserID(ctx context.Context, userID domain.UserID) (int, error)
	SetAssignCountByUserID(ctx context.Context, userID domain.UserID, count int) error
	IncrementAssignCountByUserID(ctx context.Context, userID domain.UserID) error
	DecrementAssignCountByUserID(ctx context.Context, userID domain.UserID) error
}