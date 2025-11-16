package storage

import (
	"context"

	"app/internal/domain"
	"app/internal/repository/models"
)

//go:generate mockgen -source=pr_storage.go -destination=mock/pr_storage_mock.go -package=mock
type PRStorage interface {
	CreatePullRequest(ctx context.Context, prID domain.PRID, prName string, prAuthorID domain.UserID) (*models.PullRequest, error)
	GetPullRequestByID(ctx context.Context, prID domain.PRID) (*models.PullRequest, error)
	GetPullRequestsByReviewerID(ctx context.Context, reviewerID domain.UserID) ([]models.PullRequest, error)
	GetAllOpenPullRequests(ctx context.Context) ([]models.PullRequest, error)
	UpdatePullRequestStatus(ctx context.Context, prID domain.PRID,  status domain.PRStatus) error
	CreatePRReviewerInstance(ctx context.Context, prID domain.PRID, reviewerID domain.UserID) error
	DeletePRReviewerInstance(ctx context.Context, prID domain.PRID, reviewerID domain.UserID) error
	GetReviewersFromPR(ctx context.Context, prID domain.PRID) ([]models.User, error)
}
