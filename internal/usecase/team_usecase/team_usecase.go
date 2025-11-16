package team_usecase

import (
	"app/internal/domain"
	"app/internal/mapper"
	repositoryerrs "app/internal/repository/errs"
	"app/internal/repository/storage"
	"app/internal/usecase/errs"
	"app/pkg/logger"
	"app/pkg/txmanager"
	"context"
	"errors"
)

//go:generate mockgen -source=team_usecase.go -destination=mock/mock_team_usecase.go -package=mock
type TeamUseCase interface {
	CreateTeam(ctx context.Context, teamName string, users []domain.TeamUser) (*domain.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)
}

type teamUseCase struct {
	teamStorage storage.TeamStorage
	userStorage storage.UserStorage
	txmanager   txmanager.TxManager
	logger      logger.Logger
}

func NewTeamUseCase(teamStorage storage.TeamStorage, userStorage storage.UserStorage,
	txmanager txmanager.TxManager, logger logger.Logger) TeamUseCase {
	return &teamUseCase{
		teamStorage: teamStorage,
		userStorage: userStorage,
		txmanager:   txmanager,
		logger:      logger,
	}
}

func (t *teamUseCase) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	var team *domain.Team

	if err := t.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadOnly,
		func(ctx context.Context) error {
			teamModel, err := t.teamStorage.GetTeamByName(ctx, teamName)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					t.logger.Errorw("Team not found", "teamName", teamName)
					return errs.ErrTeamNotFound
				}
				t.logger.Errorw("Failed to get team by name", "teamName", teamName, "error", err)
				return err
			}

			userModels, err := t.teamStorage.GetUsersByTeam(ctx, teamModel.ID)
			if err != nil {
				t.logger.Errorw("Failed to get users by team", "teamID", teamModel.ID, "error", err)
				return err
			}

			team = &domain.Team{
				ID:       teamModel.ID,
				TeamName: teamModel.TeamName,
				Users:    mapper.ModelsToDomainUsers(userModels),
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	t.logger.Infow("Successfully retrieved team by name", "teamName", teamName)

	return team, nil
}

func (t *teamUseCase) CreateTeam(ctx context.Context, teamName string, users []domain.TeamUser) (*domain.Team, error) {
	var team *domain.Team

	if len(users) == 0 {
		t.logger.Errorw("No users provided for team creation", "teamName", teamName)
		return nil, errs.ErrNoUsersProvided
	}

	if len(teamName) == 0 {
		t.logger.Errorw("Team name is empty")
		return nil, errs.ErrInvalidTeamName
	}

	for _, user := range users {
		if len(user.ID) == 0 || len(user.ID) > 255 || len(user.Name) == 0 {
			t.logger.Errorw("Invalid user ID provided", "userID", user.ID)
			return nil, errs.ErrInvalidUserID
		}
	}

	if err := t.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadWrite,
		func(ctx context.Context) error {

			teamModel, err := t.teamStorage.CreateTeam(ctx, teamName)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrAlreadyExists) {
					t.logger.Errorw("Team already exists", "teamName", teamName)
					return errs.ErrTeamAlreadyExists
				}
				t.logger.Errorw("Failed to create team", "teamName", teamName, "error", err)
				return err
			}

			for _, user := range users {

				_, err := t.userStorage.CreateUser(ctx, user.ID, user.Name)
				if err != nil {
					if errors.Is(err, repositoryerrs.ErrAlreadyExists) {
						t.logger.Errorw("Failed to create team", "teamName", teamName, "error", err)
						return errs.ErrUserAlreadyHasTeam
					}
					t.logger.Errorw("Failed to create user", "userID", user.ID, "error", err)
					return err
				}

				err = t.teamStorage.CreateUserTeamInstance(ctx, teamModel.ID, user.ID)
				if err != nil {
					t.logger.Errorw("Failed to create user-team instance", "teamID", teamModel.ID, "userID", user.ID, "error", err)
					return err
				}
			}

			var domainUsers []domain.User
			for _, user := range users {
				domainUsers = append(domainUsers, domain.User{ID: user.ID, Name: user.Name})
			}

			team = &domain.Team{
				ID:       teamModel.ID,
				TeamName: teamModel.TeamName,
				Users:    domainUsers,
			}

			return nil
		},
	); err != nil {
		t.logger.Errorw("Transaction failed while creating team", "teamName", teamName, "error", err)
		return nil, err
	}

	t.logger.Infow("Successfully created team", "teamName", teamName)

	return team, nil
}
