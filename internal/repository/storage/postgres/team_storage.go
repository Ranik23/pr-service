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

type teamStorage struct {
	txmanager txmanager.TxManager
	sq        squirrel.StatementBuilderType
	logger    logger.Logger
}

func NewTeamStorage(txmanager txmanager.TxManager, logger logger.Logger) storage.TeamStorage {
	return &teamStorage{
		txmanager: txmanager,
		sq:        squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:    logger,
	}
}

func (t *teamStorage) GetTeamByUserID(ctx context.Context, userID domain.UserID) (*models.Team, error) {
	tx := t.txmanager.GetExecutor(ctx)
	query, args, err := t.sq.
		Select("t.id", "t.team_name", "t.created_at").
		From("teams t").
		Join("user_teams ut ON t.id = ut.team_id").
		Where(squirrel.Eq{"ut.user_id": userID.String()}).
		ToSql()
	if err != nil {
		t.logger.Errorw("Failed to build SQL query for getting team by user ID", "error", err)
		return nil, err
	}

	var team models.Team
	err = tx.QueryRow(ctx, query, args...).Scan(&team.ID, &team.TeamName, &team.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			t.logger.Warnw("Team not found for user", "user_id", userID)
			return nil, errs.ErrNotFound
		}
		t.logger.Errorw("Failed to get team by user ID", "user_id", userID, "error", err)
		return nil, err
	}

	t.logger.Infow("Successfully retrieved team by user ID", "user_id", userID, "team_id", team.ID)
	return &team, nil
}

func (t *teamStorage) GetUsersByTeam(ctx context.Context, teamID domain.TeamID) ([]models.User, error) {
	tx := t.txmanager.GetExecutor(ctx)
	query, args, err := t.sq.
		Select("u.id", "u.is_active", "u.name").
		From("users u").
		Join("user_teams ut ON u.id = ut.user_id").
		Where(squirrel.Eq{"ut.team_id": teamID.Int64()}).
		ToSql()
	if err != nil {
		t.logger.Errorw("Failed to build SQL query for getting users by team", "error", err)
		return nil, err
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		t.logger.Errorw("Failed to get users by team", "team_id", teamID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.StatusActivity, &user.Name); err != nil {
			t.logger.Errorw("Failed to scan user row for team", "team_id", teamID, "error", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		t.logger.Errorw("Error during rows iteration for team users", "team_id", teamID, "error", err)
		return nil, err
	}

	t.logger.Infow("Successfully retrieved users by team", "team_id", teamID, "count", len(users))
	return users, nil
}

func (t *teamStorage) CreateTeam(ctx context.Context, teamName string) (*models.Team, error) {
	tx := t.txmanager.GetExecutor(ctx)

	query, args, err := t.sq.
		Insert("teams").
		Columns("team_name", "created_at").
		Values(teamName, squirrel.Expr("NOW()")).
		Suffix("RETURNING id, team_name, created_at").
		ToSql()
	if err != nil {
		t.logger.Errorw("Failed to build SQL query for creating team", "error", err)
		return nil, err
	}

	var team models.Team
	err = tx.QueryRow(ctx, query, args...).Scan(&team.ID, &team.TeamName, &team.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				t.logger.Warnw("Team already exists", "team_name", teamName)
				return nil, errs.ErrAlreadyExists
			}
		}
		t.logger.Errorw("Failed to create team", "team_name", teamName, "error", err)
		return nil, err
	}

	t.logger.Infow("Successfully created team", "team_id", team.ID, "team_name", teamName)
	return &team, nil
}

func (t *teamStorage) CreateUserTeamInstance(ctx context.Context, teamID domain.TeamID, userID domain.UserID) error {
	tx := t.txmanager.GetExecutor(ctx)

	query, args, err := t.sq.
		Insert("user_teams").
		Columns("team_id", "user_id", "joined_at").
		Values(teamID.Int64(), userID.String(), squirrel.Expr("NOW()")).
		ToSql()
	if err != nil {
		t.logger.Errorw("Failed to build SQL query for creating user-team instance", "error", err)
		return err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				t.logger.Warnw("User-team instance already exists", "team_id", teamID, "user_id", userID)
				return errs.ErrAlreadyExists
			}
		}
		t.logger.Errorw("Failed to create user-team instance", "team_id", teamID, "user_id", userID, "error", err)
		return err
	}

	t.logger.Infow("Successfully created user-team instance", "team_id", teamID, "user_id", userID)
	return nil
}

func (t *teamStorage) GetTeamByID(ctx context.Context, teamID domain.TeamID) (*models.Team, error) {
	tx := t.txmanager.GetExecutor(ctx)
	query, args, err := t.sq.
		Select("id", "team_name", "created_at").
		From("teams").
		Where(squirrel.Eq{"id": teamID.Int64()}).
		ToSql()
	if err != nil {
		t.logger.Errorw("Failed to build SQL query for getting team by ID", "error", err)
		return nil, err
	}

	var team models.Team
	err = tx.QueryRow(ctx, query, args...).Scan(&team.ID, &team.TeamName, &team.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			t.logger.Warnw("Team not found by ID", "team_id", teamID)
			return nil, errs.ErrNotFound
		}
		t.logger.Errorw("Failed to get team by ID", "team_id", teamID, "error", err)
		return nil, err
	}

	t.logger.Infow("Successfully retrieved team by ID", "team_id", teamID, "team_name", team.TeamName)
	return &team, nil
}

func (t *teamStorage) GetTeamByName(ctx context.Context, teamName string) (*models.Team, error) {
	tx := t.txmanager.GetExecutor(ctx)
	query, args, err := t.sq.
		Select("id", "team_name", "created_at").
		From("teams").
		Where(squirrel.Eq{"team_name": teamName}).
		ToSql()
	if err != nil {
		t.logger.Errorw("Failed to build SQL query for getting team by name", "error", err)
		return nil, err
	}

	var team models.Team
	err = tx.QueryRow(ctx, query, args...).Scan(&team.ID, &team.TeamName, &team.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			t.logger.Warnw("Team not found by name", "team_name", teamName)
			return nil, errs.ErrNotFound
		}
		t.logger.Errorw("Failed to get team by name", "team_name", teamName, "error", err)
		return nil, err
	}

	t.logger.Infow("Successfully retrieved team by name", "team_id", team.ID, "team_name", teamName)
	return &team, nil
}
