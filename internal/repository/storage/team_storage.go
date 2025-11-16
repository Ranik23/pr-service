package storage

import (
	"app/internal/domain"
	"app/internal/repository/models"
	"context"
)

//go:generate mockgen -source=team_storage.go -destination=mock/team_storage_mock.go -package=mock
type TeamStorage interface {
	CreateTeam(ctx context.Context, teamName string) (*models.Team, error)
	GetTeamByID(ctx context.Context, teamID domain.TeamID) (*models.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (*models.Team, error)
	GetTeamByUserID(ctx context.Context, userID domain.UserID) (*models.Team, error)
	CreateUserTeamInstance(ctx context.Context, teamID domain.TeamID, userID domain.UserID) error
	GetUsersByTeam(ctx context.Context, teamID domain.TeamID) ([]models.User, error)
}