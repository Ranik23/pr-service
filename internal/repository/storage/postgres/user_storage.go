package postgres

import (
	"app/internal/domain"
	"app/internal/repository/errs"
	"app/internal/repository/models"
	"app/internal/repository/storage"
	"app/pkg/logger"
	"app/pkg/txmanager"
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type userStorage struct {
	txmanager txmanager.TxManager
	sq        squirrel.StatementBuilderType
	logger    logger.Logger
}

func NewUserStorage(txmanager txmanager.TxManager, logger logger.Logger) storage.UserStorage {
	return &userStorage{
		txmanager: txmanager,
		sq:        squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:    logger,
	}
}

func (u *userStorage) CreateUser(ctx context.Context, userID domain.UserID, name string) (*models.User, error) {
	tx := u.txmanager.GetExecutor(ctx)

	query, args, err := u.sq.
		Insert("users").
		Columns("id", "name").
		Values(userID.String(), name).
		Suffix("RETURNING id, is_active, name").
		ToSql()
	if err != nil {
		u.logger.Errorw("Failed to build SQL query for creating user", "error", err)
		return nil, err
	}

	var user models.User
	err = tx.QueryRow(ctx, query, args...).Scan(&user.ID, &user.StatusActivity, &user.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				u.logger.Warnw("User already exists", "user_id", userID, "name", name)
				return nil, errs.ErrAlreadyExists
			}
		}
		u.logger.Errorw("Failed to create user", "user_id", userID, "name", name, "error", err)
		return nil, err
	}

	u.logger.Infow("Successfully created user", "user_id", userID, "name", name)
	return &user, nil
}

func (u *userStorage) GetActiveUsersByTeam(ctx context.Context, teamID domain.TeamID) ([]models.User, error) {
	tx := u.txmanager.GetExecutor(ctx)

	query, args, err := u.sq.
		Select("u.id", "u.is_active", "u.name").
		From("users u").
		Join("user_teams ut ON u.id = ut.user_id").
		Where(squirrel.Eq{"ut.team_id": teamID.Int64()}).
		Where(squirrel.Eq{"u.is_active": true}).
		ToSql()
	if err != nil {
		u.logger.Errorw("Failed to build SQL query for getting active users by team", "error", err)
		return nil, err
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		u.logger.Errorw("Failed to get active users by team", "team_id", teamID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.StatusActivity, &user.Name); err != nil {
			u.logger.Errorw("Failed to scan user row for active team users", "team_id", teamID, "error", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		u.logger.Errorw("Error during rows iteration for active team users", "team_id", teamID, "error", err)
		return nil, err
	}

	u.logger.Infow("Successfully retrieved active users by team", "team_id", teamID, "count", len(users))
	return users, nil
}

func (u *userStorage) GetUserByID(ctx context.Context, userID domain.UserID) (*models.User, error) {
	tx := u.txmanager.GetExecutor(ctx)

	query, args, err := u.sq.
		Select("id", "is_active", "name").
		From("users").
		Where(squirrel.Eq{"id": userID.String()}).
		ToSql()
	if err != nil {
		u.logger.Errorw("Failed to build SQL query for getting user by ID", "error", err)
		return nil, err
	}

	var user models.User
	err = tx.QueryRow(ctx, query, args...).Scan(&user.ID, &user.StatusActivity, &user.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u.logger.Warnw("User not found by ID", "user_id", userID)
			return nil, errs.ErrNotFound
		}
		u.logger.Errorw("Failed to get user by ID", "user_id", userID, "error", err)
		return nil, err
	}

	u.logger.Infow("Successfully retrieved user by ID", "user_id", userID, "is_active", user.StatusActivity)
	return &user, nil
}

func (u *userStorage) UpdateActivity(ctx context.Context, userID domain.UserID, statusActivity domain.UserActivityStatus) error {
	tx := u.txmanager.GetExecutor(ctx)

	query, args, err := u.sq.
		Update("users").
		Set("is_active", statusActivity.IsActive()).
		Where(squirrel.Eq{"id": userID.String()}).
		ToSql()
	if err != nil {
		u.logger.Errorw("Failed to build SQL query for updating user activity", "error", err)
		return err
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		u.logger.Errorw("Failed to update user activity", "user_id", userID, "status", statusActivity, "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		u.logger.Warnw("No user found to update activity", "user_id", userID)
		return errs.ErrNotFound
	}

	u.logger.Infow("Successfully updated user activity", "user_id", userID, "status", statusActivity)
	return nil
}
