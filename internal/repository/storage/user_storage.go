package storage

import (
	"app/internal/domain"
	"app/internal/repository/models"
	"context"
)

//go:generate mockgen -source=user_storage.go -destination=mock/user_storage_mock.go -package=mock
type UserStorage interface {
	CreateUser(ctx context.Context, userID domain.UserID, name string) (*models.User, error)
	GetUserByID(ctx context.Context, userID domain.UserID) (*models.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamID domain.TeamID) ([]models.User, error)
	UpdateActivity(ctx context.Context, userID domain.UserID, isActive domain.UserActivityStatus) error
}