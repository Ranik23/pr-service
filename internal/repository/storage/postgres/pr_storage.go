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

type prStorage struct {
	txmanager txmanager.TxManager
	sq        squirrel.StatementBuilderType
	logger    logger.Logger
}

func NewPRStorage(txmanager txmanager.TxManager, logger logger.Logger) storage.PRStorage {
	return &prStorage{
		logger:    logger,
		txmanager: txmanager,
		sq:        squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (p *prStorage) GetAllOpenPullRequests(ctx context.Context) ([]models.PullRequest, error) {
	tx := p.txmanager.GetExecutor(ctx)

	query, args, err := p.sq.
		Select("id", "name", "author_id", "status", "need_more_reviewers", "created_at", "merged_at").
		From("pull_requests").
		Where(squirrel.Eq{"status": domain.PRStatusOpen}).
		ToSql()
	if err != nil {
		p.logger.Errorw("Failed to build SQL query for getting open pull requests", "error", err)
		return nil, err
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		p.logger.Errorw("Failed to get all open pull requests", "error", err)
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.NeedMoreReviewers, &pr.CreatedAt, &pr.MergedAt); err != nil {
			p.logger.Errorw("Failed to scan pull request row", "error", err)
			return nil, err
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		p.logger.Errorw("Error during rows iteration", "error", err)
		return nil, err
	}

	p.logger.Infow("Successfully retrieved open pull requests", "count", len(prs))
	return prs, nil
}

func (p *prStorage) CreatePRReviewerInstance(ctx context.Context, prID domain.PRID, reviewerID domain.UserID) error {
	tx := p.txmanager.GetExecutor(ctx)

	query, args, err := p.sq.
		Insert("pr_reviewers").
		Columns("pr_id", "reviewer_id", "assigned_at").
		Values(prID.String(), reviewerID.String(), squirrel.Expr("NOW()")).
		ToSql()
	if err != nil {
		p.logger.Errorw("Failed to build SQL query for creating PR reviewer", "error", err)
		return err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				p.logger.Warnw("PR reviewer already exists", "pr_id", prID, "reviewer_id", reviewerID)
				return errs.ErrAlreadyExists
			}
		}
		p.logger.Errorw("Failed to create PR reviewer instance", "pr_id", prID, "reviewer_id", reviewerID, "error", err)
		return err
	}

	p.logger.Infow("Successfully created PR reviewer instance", "pr_id", prID, "reviewer_id", reviewerID)
	return nil
}

func (p *prStorage) CreatePullRequest(ctx context.Context, prID domain.PRID, prName string, prAuthorID domain.UserID) (*models.PullRequest, error) {
	tx := p.txmanager.GetExecutor(ctx)

	query, args, err := p.sq.
		Insert("pull_requests").
		Columns("id", "name", "author_id", "status", "need_more_reviewers").
		Values(prID.String(), prName, prAuthorID.String(), domain.PRStatusOpen, true).
		Suffix("RETURNING id, name, author_id, status, need_more_reviewers, created_at").
		ToSql()
	if err != nil {
		p.logger.Errorw("Failed to build SQL query for creating pull request", "error", err)
		return nil, err
	}

	var pr models.PullRequest
	err = tx.QueryRow(ctx, query, args...).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.NeedMoreReviewers, &pr.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				p.logger.Warnw("Pull request already exists", "pr_id", prID, "pr_name", prName)
				return nil, errs.ErrAlreadyExists
			}
		}
		p.logger.Errorw("Failed to create pull request", "pr_id", prID, "pr_name", prName, "author_id", prAuthorID, "error", err)
		return nil, err
	}

	p.logger.Infow("Successfully created pull request", "pr_id", prID, "pr_name", prName, "author_id", prAuthorID)
	return &pr, nil
}

func (p *prStorage) GetPullRequestByID(ctx context.Context, prID domain.PRID) (*models.PullRequest, error) {
	tx := p.txmanager.GetExecutor(ctx)

	query, args, err := p.sq.
		Select("id", "name", "author_id", "status", "need_more_reviewers", "created_at", "merged_at").
		From("pull_requests").
		Where(squirrel.Eq{"id": prID.String()}).
		ToSql()
	if err != nil {
		p.logger.Errorw("Failed to build SQL query for getting pull request by ID", "error", err)
		return nil, err
	}

	var pr models.PullRequest
	err = tx.QueryRow(ctx, query, args...).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.NeedMoreReviewers, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.logger.Warnw("Pull request not found", "pr_id", prID)
			return nil, errs.ErrNotFound
		}
		p.logger.Errorw("Failed to get pull request by ID", "pr_id", prID, "error", err)
		return nil, err
	}

	p.logger.Infow("Successfully retrieved pull request by ID", "pr_id", prID, "status", pr.Status)
	return &pr, nil
}

func (p *prStorage) GetPullRequestsByReviewerID(ctx context.Context, reviewerID domain.UserID) ([]models.PullRequest, error) {
	tx := p.txmanager.GetExecutor(ctx)

	query, args, err := p.sq.
		Select("pr.id", "pr.name", "pr.author_id", "pr.status", "pr.need_more_reviewers", "pr.created_at", "pr.merged_at").
		From("pull_requests pr").
		Join("pr_reviewers prr ON pr.id = prr.pr_id").
		Where(squirrel.Eq{"prr.reviewer_id": reviewerID.String()}).
		ToSql()
	if err != nil {
		p.logger.Errorw("Failed to build SQL query for getting PRs by reviewer ID", "error", err)
		return nil, err
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		p.logger.Errorw("Failed to get pull requests by reviewer ID", "reviewer_id", reviewerID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.NeedMoreReviewers, &pr.CreatedAt, &pr.MergedAt); err != nil {
			p.logger.Errorw("Failed to scan pull request row for reviewer", "reviewer_id", reviewerID, "error", err)
			return nil, err
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		p.logger.Errorw("Error during rows iteration for reviewer PRs", "reviewer_id", reviewerID, "error", err)
		return nil, err
	}

	p.logger.Infow("Successfully retrieved pull requests for reviewer", "reviewer_id", reviewerID, "count", len(prs))
	return prs, nil
}

func (p *prStorage) GetReviewersFromPR(ctx context.Context, prID domain.PRID) ([]models.User, error) {
	tx := p.txmanager.GetExecutor(ctx)

	query, args, err := p.sq.
		Select("u.id", "u.is_active", "u.name").
		From("users u").
		Join("pr_reviewers prr ON u.id = prr.reviewer_id").
		Where(squirrel.Eq{"prr.pr_id": prID.String()}).
		ToSql()
	if err != nil {
		p.logger.Errorw("Failed to build SQL query for getting reviewers from PR", "error", err)
		return nil, err
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		p.logger.Errorw("Failed to get reviewers from PR", "pr_id", prID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var reviewers []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.StatusActivity, &user.Name); err != nil {
			p.logger.Errorw("Failed to scan reviewer row", "pr_id", prID, "error", err)
			return nil, err
		}
		reviewers = append(reviewers, user)
	}

	if err := rows.Err(); err != nil {
		p.logger.Errorw("Error during rows iteration for reviewers", "pr_id", prID, "error", err)
		return nil, err
	}

	p.logger.Infow("Successfully retrieved reviewers from PR", "pr_id", prID, "count", len(reviewers))
	return reviewers, nil
}

func (p *prStorage) DeletePRReviewerInstance(ctx context.Context, prID domain.PRID, reviewerID domain.UserID) error {
	tx := p.txmanager.GetExecutor(ctx)

	query, args, err := p.sq.
		Delete("pr_reviewers").
		Where(squirrel.Eq{"pr_id": prID.String()}).
		Where(squirrel.Eq{"reviewer_id": reviewerID.String()}).
		ToSql()
	if err != nil {
		p.logger.Errorw("Failed to build SQL query for deleting PR reviewer", "error", err)
		return err
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		p.logger.Errorw("Failed to delete PR reviewer instance", "pr_id", prID, "reviewer_id", reviewerID, "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		p.logger.Warnw("No PR reviewer instance found to delete", "pr_id", prID, "reviewer_id", reviewerID)
		return errs.ErrNotFound
	}

	p.logger.Infow("Successfully deleted PR reviewer instance", "pr_id", prID, "reviewer_id", reviewerID)
	return nil
}

func (p *prStorage) UpdatePullRequestStatus(ctx context.Context, prID domain.PRID, status domain.PRStatus) error {
	tx := p.txmanager.GetExecutor(ctx)

	query, args, err := p.sq.
		Update("pull_requests").
		Set("status", status).
		Where(squirrel.Eq{"id": prID.String()}).
		ToSql()
	if err != nil {
		p.logger.Errorw("Failed to build SQL query for updating PR status", "error", err)
		return err
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		p.logger.Errorw("Failed to update pull request status", "pr_id", prID, "status", status, "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		p.logger.Warnw("No pull request found to update status", "pr_id", prID)
		return errs.ErrNotFound
	}

	p.logger.Infow("Successfully updated pull request status", "pr_id", prID, "status", status)
	return nil
}
