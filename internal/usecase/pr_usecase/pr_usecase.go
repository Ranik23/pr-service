package pr_usecase

import (
	"app/internal/domain"
	"app/internal/mapper"
	"app/internal/repository/cache"
	repositoryerrs "app/internal/repository/errs"
	"app/internal/repository/models"
	"app/internal/repository/storage"
	"app/internal/usecase/errs"
	"app/pkg/logger"
	"app/pkg/txmanager"
	"context"
	"errors"
	"strings"

	"math/rand"
)

//go:generate mockgen -source=pr_usecase.go -destination=mock/mock_pr_usecase.go -package=mock
type PullRequestUseCase interface {
	CreatePR(ctx context.Context, prAuthorID domain.UserID, prID domain.PRID, prName string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID domain.PRID, reviewerIDToChange domain.UserID) error
	MergePR(ctx context.Context, prID domain.PRID) error
	GetPRByUserID(ctx context.Context, userID domain.UserID) ([]domain.PullRequest, error)
}

type pullRequestUseCase struct {
	statsCache  cache.StatsCache
	prStorage   storage.PRStorage
	userStorage storage.UserStorage
	teamStorage storage.TeamStorage
	txmanager   txmanager.TxManager
	logger      logger.Logger
}

func NewPRUseCase(prStorage storage.PRStorage, userStorage storage.UserStorage, cache cache.StatsCache,
	teamStorage storage.TeamStorage, txmanager txmanager.TxManager, logger logger.Logger) PullRequestUseCase {
	return &pullRequestUseCase{
		prStorage:   prStorage,
		statsCache:  cache,
		userStorage: userStorage,
		teamStorage: teamStorage,
		txmanager:   txmanager,
		logger:      logger,
	}
}

func (p *pullRequestUseCase) GetPRByUserID(ctx context.Context, userID domain.UserID) ([]domain.PullRequest, error) {
	var prs []domain.PullRequest

	if len(userID) == 0 {
		p.logger.Errorw("User ID is empty")
		return nil, errs.ErrInvalidUserID
	}

	err := p.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadOnly,
		func(ctx context.Context) error {

			prModels, err := p.prStorage.GetPullRequestsByReviewerID(ctx, userID)
			if err != nil {
				p.logger.Errorw("Failed to get pull requests by reviewer ID", "userID", userID, "error", err)
				return err
			}

			for _, pm := range prModels {
				authorModel, err := p.userStorage.GetUserByID(ctx, pm.AuthorID)
				if err != nil {
					p.logger.Errorw("Failed to get user by ID", "userID", pm.AuthorID, "error", err)
					return err
				}

				reviewers, err := p.prStorage.GetReviewersFromPR(ctx, pm.ID)
				if err != nil {
					p.logger.Errorw("Failed to get reviewers from pull request", "prID", pm.ID, "error", err)
					return err
				}

				prs = append(prs, domain.PullRequest{
					ID:        pm.ID,
					Name:      pm.Name,
					Reviewers: mapper.ModelsToDomainUsers(reviewers),
					Author:    mapper.ModelToDomainUser(*authorModel),
					Status:    pm.Status,
					CreatedAt: pm.CreatedAt,
					MergedAt:  pm.MergedAt,
				})
			}

			return nil
		})

	if err != nil {
		p.logger.Errorw("Failed to get pull requests by reviewer ID", "userID", userID, "error", err)
		return nil, err
	}

	p.logger.Infow("Successfully retrieved pull requests for user", "userID", userID, "count", len(prs))

	return prs, nil
}

func (p *pullRequestUseCase) CreatePR(ctx context.Context, prAuthorID domain.UserID, prID domain.PRID, prName string) (*domain.PullRequest, error) {
	var pr *domain.PullRequest

	if len(prName) == 0 {
		p.logger.Errorw("Pull request name is empty")
		return nil, errs.ErrInvalidPullRequestName
	}

	if len(prAuthorID) == 0 {
		p.logger.Errorw("Pull request author ID is empty")
		return nil, errs.ErrInvalidUserID
	}

	if len(prID) == 0 {
		p.logger.Errorw("Pull request ID is invalid", "prID", prID)
		return nil, errs.ErrInvalidPullRequestID
	}

	if err := p.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadWrite,
		func(ctx context.Context) error {
			author, err := p.userStorage.GetUserByID(ctx, prAuthorID)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					p.logger.Errorw("Failed to get user by ID", "userID", prAuthorID, "error", err)
					return errs.ErrUserNotFound
				}
				p.logger.Errorw("Failed to get user by ID", "userID", prAuthorID, "error", err)
				return err
			}

			prModel, err := p.prStorage.CreatePullRequest(ctx, prID, prName, prAuthorID)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrAlreadyExists) {
					p.logger.Errorw("Pull request already exists", "prName", prName)
					return errs.ErrPullRequestAlreadyExists
				}
				p.logger.Errorw("Failed to create pull request", "prName", prName, "error", err)
				return err
			}

			team, err := p.teamStorage.GetTeamByUserID(ctx, prAuthorID)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					p.logger.Errorw("User has no team", "userID", prAuthorID)
					return errs.ErrUserHasNoTeam // что невозможно раз он уже был создан
				}
				p.logger.Errorw("Failed to get team by user ID", "userID", prAuthorID, "error", err)
				return err
			}

			users, err := p.teamStorage.GetUsersByTeam(ctx, team.ID)
			if err != nil {
				p.logger.Errorw("Failed to get users by team", "teamID", team.ID, "error", err)
				return err
			}

			n := len(users)

			count := rand.Intn(2) + 1
			if count > n {
				count = n
			}

			shuffled := make([]models.User, n)
			copy(shuffled, users)
			rand.Shuffle(n, func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

			selectedReviewers := make([]models.User, 0, count)
			for _, u := range shuffled {
				if !strings.EqualFold(u.ID.String(), prAuthorID.String()) && u.StatusActivity {
					selectedReviewers = append(selectedReviewers, u)
					if len(selectedReviewers) == count {
						break
					}
				}
			}

			for _, reviewer := range selectedReviewers {
				if err := p.prStorage.CreatePRReviewerInstance(ctx, prModel.ID, reviewer.ID); err != nil {
					p.logger.Errorw("Failed to create PR reviewer instance", "prID", prModel.ID, "reviewerID", reviewer.ID, "error", err)
					return err
				}
			}

			for _, reviewer := range selectedReviewers {
				if err := p.statsCache.IncrementAssignCountByUserID(ctx, reviewer.ID); err != nil {
					p.logger.Errorw("Failed to increment assign count in stats cache", "userID", reviewer.ID, "error", err)
					return err
				}
			}

			pr = &domain.PullRequest{
				ID:        prModel.ID,
				Name:      prName,
				Author:    mapper.ModelToDomainUser(*author),
				Status:    prModel.Status,
				CreatedAt: prModel.CreatedAt,
				MergedAt:  prModel.MergedAt,
			}

			return nil
		}); err != nil {
		p.logger.Errorw("Transaction failed while creating pull request", "prName", prName, "error", err)
		return nil, err
	}

	p.logger.Infow("Successfully created pull request", "prID", pr.ID, "prName", pr.Name)

	return pr, nil
}

func (p *pullRequestUseCase) ReassignReviewer(ctx context.Context, prID domain.PRID, reviewerIDToRemove domain.UserID) error {
	return p.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadWrite,
		func(ctx context.Context) error {

			pr, err := p.prStorage.GetPullRequestByID(ctx, prID)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					p.logger.Errorw("Failed to get pull request by ID", "prID", prID, "error", err)
					return errs.ErrPullRequestNotFound
				}
				p.logger.Errorw("Failed to get pull request by ID", "prID", prID, "error", err)
				return err
			}

			if pr.Status == domain.PRStatusMerged {
				p.logger.Errorw("Pull Request already merged! Can't reassign", "prID", prID)
				return errs.ErrPRAlreadyMerged
			}

			if err := p.prStorage.DeletePRReviewerInstance(ctx, prID, reviewerIDToRemove); err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					p.logger.Errorw("Failed to delete PR reviewer instance", "prID", prID, "reviewerID", reviewerIDToRemove, "error", err)
					return errs.ErrReviewerNotFoundInPullRequest
				}
				p.logger.Errorw("Failed to delete PR reviewer instance", "prID", prID, "reviewerID", reviewerIDToRemove, "error", err)
				return err
			}

			team, err := p.teamStorage.GetTeamByUserID(ctx, reviewerIDToRemove)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					p.logger.Errorw("User has no team", "userID", reviewerIDToRemove)
					return errs.ErrUserHasNoTeam
				}
				p.logger.Errorw("Failed to get team by user ID", "userID", pr.AuthorID, "error", err)
				return err
			}

			activeUsers, err := p.userStorage.GetActiveUsersByTeam(ctx, team.ID)
			if err != nil {
				p.logger.Errorw("Failed to get active users by team", "teamID", team.ID, "error", err)
				return err
			}

			reviewers, err := p.prStorage.GetReviewersFromPR(ctx, prID)
			if err != nil {
				p.logger.Errorw("Failed to get reviewers from pull request", "prID", prID, "error", err)
				return err
			}

			user := p.pickFirstAvailableActiveUser(activeUsers, reviewers, pr.AuthorID)
			if user == nil {
				p.logger.Errorw("No available active user to assign as reviewer", "prID", prID)
				return errs.ErrNoAvailableActiveUserToAssign
			}

			if err := p.prStorage.CreatePRReviewerInstance(ctx, prID, user.ID); err != nil {
				p.logger.Errorw("Failed to create PR reviewer instance", "prID", prID, "reviewerID", user.ID, "error", err)
				return err
			}

			if err := p.statsCache.IncrementAssignCountByUserID(ctx, user.ID); err != nil {
				p.logger.Errorw("Failed to increment assign count in stats cache", "userID", user.ID, "error", err)
				return err
			}

			p.logger.Infow("Successfully reassigned reviewer", "prID", prID, "reviewerID", user.ID)

			return nil
		},
	)
}

func (p *pullRequestUseCase) pickFirstAvailableActiveUser(activeUsers, reviewers []models.User, authorID domain.UserID) *models.User {
	reviewerIDs := make(map[string]struct{}, len(reviewers))
	for _, r := range reviewers {
		reviewerIDs[r.ID.String()] = struct{}{}
	}

	for _, u := range activeUsers {
		if !strings.EqualFold(u.ID.String(), authorID.String()) && u.StatusActivity {
			if _, exists := reviewerIDs[u.ID.String()]; !exists {
				return &u
			}
		}
	}

	p.logger.Errorw("No available active user to assign as reviewer")

	return nil
}

func (p *pullRequestUseCase) MergePR(ctx context.Context, prID domain.PRID) error {
	return p.txmanager.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadWrite,
		func(ctx context.Context) error {
			pr, err := p.prStorage.GetPullRequestByID(ctx, prID)
			if err != nil {
				if errors.Is(err, repositoryerrs.ErrNotFound) {
					p.logger.Errorw("Failed to get pull request by ID", "prID", prID, "error", err)
					return errs.ErrPullRequestNotFound
				}
				p.logger.Errorw("Failed to get pull request by ID", "prID", prID, "error", err)
				return err
			}

			if pr.Status == domain.PRStatusMerged {
				p.logger.Warnw("Pull Request already merged, skipping merge operation", "prID", prID)
				return nil
			}

			if err := p.prStorage.UpdatePullRequestStatus(ctx, pr.ID, domain.PRStatusMerged); err != nil {
				p.logger.Errorw("Failed to update pull request status", "prID", pr.ID, "error", err)
				return err
			}

			p.logger.Infow("Successfully merged pull request", "prID", pr.ID)

			return nil
		},
	)
}
