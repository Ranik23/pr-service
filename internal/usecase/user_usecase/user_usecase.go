package user_usecase

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

type UserUseCase interface {
	CreateUser(ctx context.Context, userID domain.UserID, name string) (*domain.User, error)
	GetUserByID(ctx context.Context, userID domain.UserID) (*domain.User, error)
	UpdateUserActivity(ctx context.Context, userID domain.UserID, isActive domain.UserActivityStatus) (*domain.User, error)
	DeactivateUsersByTeamName(ctx context.Context, teamName string) error
}

type userUseCase struct {
	userStorage storage.UserStorage
	teamStorage storage.TeamStorage
	txmanager   txmanager.TxManager
	logger      logger.Logger
}

func (u *userUseCase) DeactivateUsersByTeamName(ctx context.Context, teamName string) error {

	if err := u.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadWrite,
		func(ctx context.Context) error {

			team, err := u.teamStorage.GetTeamByName(ctx, teamName)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					u.logger.Errorw("Team not found", "teamName", teamName)
					return errs.ErrTeamNotFound
				}
				u.logger.Errorw("Failed to get team by name", "teamName", teamName, "error", err)
				return err
			}

			user, err := u.teamStorage.GetUsersByTeam(ctx, team.ID)
			if err != nil {
				u.logger.Errorw("Failed to get users by team", "teamName", teamName, "error", err)
				return err
			}

			if len(user) == 0 {
				u.logger.Infow("No users found for the team", "teamName", teamName)
				return errs.ErrNoUsersInTeam
			}

			for _, usr := range user {
				err := u.userStorage.UpdateActivity(ctx, usr.ID, domain.UserStatusInactive)
				if err != nil {
					u.logger.Errorw("Failed to deactivate user", "userID", usr.ID, "error", err)
					return err
				}
			}
			return nil

		}); err != nil {
		u.logger.Errorw("Transaction failed while deactivating users by team name", "teamName", teamName, "error", err)
		return err
	}

	u.logger.Infow("Successfully deactivated users for the team", "teamName", teamName)

	return nil
}

func NewUserUseCase(userStorage storage.UserStorage, txmanager txmanager.TxManager,
	teamStorage storage.TeamStorage, logger logger.Logger) UserUseCase {
	return &userUseCase{
		userStorage: userStorage,
		teamStorage: teamStorage,
		txmanager:   txmanager,
		logger:      logger,
	}
}
func (u *userUseCase) CreateUser(ctx context.Context, userID domain.UserID, name string) (*domain.User, error) {
	var user domain.User

	if len(userID) == 0 || len(userID) > 255 {
		u.logger.Errorw("Invalid user ID", "userID", userID)
		return nil, errs.ErrInvalidUserID
	}

	if err := u.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadWrite,
		func(ctx context.Context) error {
			userModel, err := u.userStorage.CreateUser(ctx, userID, name)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrAlreadyExists) {
					u.logger.Errorw("Failed to create user", "userID", userID, "error", err)
					return errs.ErrUserAlreadyExists
				}
				u.logger.Errorw("Failed to create user", "userID", userID, "error", err)
				return err
			}

			user = mapper.ModelToDomainUser(*userModel)
			return nil

		}); err != nil {
		u.logger.Errorw("Transaction failed while creating user", "userID", userID, "error", err)
		return nil, err
	}

	u.logger.Infow("Successfully created user", "userID", userID)

	return &user, nil
}

func (u *userUseCase) GetUserByID(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	var user domain.User

	if len(userID) == 0 {
		u.logger.Errorw("User ID is empty")
		return nil, errs.ErrInvalidUserID
	}

	if err := u.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadOnly,
		func(ctx context.Context) error {
			userModel, err := u.userStorage.GetUserByID(ctx, userID)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					u.logger.Errorw("User not found", "userID", userID)
					return errs.ErrUserNotFound
				}
				u.logger.Errorw("Failed to get user by ID", "userID", userID, "error", err)
				return err
			}

			user = mapper.ModelToDomainUser(*userModel)

			return nil
		}); err != nil {
		u.logger.Errorw("Transaction failed while getting user by ID", "userID", userID, "error", err)
		return nil, err
	}

	u.logger.Infow("Successfully retrieved user by ID", "userID", userID)

	return &user, nil
}

func (u *userUseCase) UpdateUserActivity(ctx context.Context, userID domain.UserID, isActive domain.UserActivityStatus) (*domain.User, error) {
	var updatedUser domain.User

	if err := u.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadWrite,
		func(ctx context.Context) error {
			user, err := u.userStorage.GetUserByID(ctx, userID)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					u.logger.Errorw("User not found", "userID", userID)
					return errs.ErrUserNotFound
				}
				u.logger.Errorw("Failed to get user by ID", "userID", userID, "error", err)
				return err
			}

			err = u.userStorage.UpdateActivity(ctx, user.ID, isActive)
			if err != nil {
				u.logger.Errorw("Failed to update user activity", "userID", userID, "error", err)
				return err
			}

			newUser := user
			newUser.StatusActivity = isActive.IsActive()

			updatedUser = mapper.ModelToDomainUser(*newUser)

			return nil
		}); err != nil {
		u.logger.Errorw("Transaction failed while updating user activity", "userID", userID, "error", err)
		return nil, err
	}

	u.logger.Infow("Successfully updated user activity", "userID", userID, "isActive", isActive)

	return &updatedUser, nil
}
