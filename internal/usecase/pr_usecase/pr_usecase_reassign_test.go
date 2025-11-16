package pr_usecase

import (
	"app/internal/domain"
	cachemock "app/internal/repository/cache/mock"
	repositoryerrs "app/internal/repository/errs"
	"app/internal/repository/models"
	mock "app/internal/repository/storage/mock"
	"app/internal/usecase/errs"
	mocklog "app/pkg/logger/mock"
	txmock "app/pkg/txmanager/mock"
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestReAssignSuccess(t *testing.T) {
	Convey("ReAssign: success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
		mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("100")
		reviewerIDToChange := domain.UserID("200")
		authorID := domain.UserID("author-123")
		teamID := domain.TeamID(1)
		newReviewerID := domain.UserID("new-reviewer-300")

		statsCache.EXPECT().IncrementAssignCountByUserID(gomock.Any(), gomock.Any()).AnyTimes()

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(&models.PullRequest{
				ID:       prID,
				AuthorID: authorID,
				Status:   domain.PRStatusOpen,
			}, nil)

		prStorage.EXPECT().DeletePRReviewerInstance(gomock.Any(), prID, reviewerIDToChange).
			Return(nil)

		teamStorage.EXPECT().GetTeamByUserID(gomock.Any(), reviewerIDToChange).
			Return(&models.Team{
				ID:       teamID,
				TeamName: "test-team",
			}, nil)

		userID1 := domain.UserID("1")
		userID2 := domain.UserID("2")
		

		userStorage.EXPECT().GetActiveUsersByTeam(gomock.Any(), teamID).
			Return([]models.User{
				{ID: userID1, StatusActivity: true},
				{ID: userID2, StatusActivity: true},
				{ID: newReviewerID, StatusActivity: true},
				{ID: authorID, StatusActivity: true},
			}, nil)

		prStorage.EXPECT().GetReviewersFromPR(gomock.Any(), prID).
			Return([]models.User{
				{ID: reviewerIDToChange, StatusActivity: true},
				{ID: "other-reviewer", StatusActivity: true},
			}, nil)

		prStorage.EXPECT().CreatePRReviewerInstance(gomock.Any(), prID, gomock.Any()).
			Return(nil)

		err := uc.ReassignReviewer(context.Background(), prID, reviewerIDToChange)
		So(err, ShouldBeNil)
	})
}

func TestReAssign_UserHasNoTeam(t *testing.T) {
	Convey("ReAssign: author has no team", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("100")
		reviewerIDToChange := domain.UserID("200")
		authorID := domain.UserID("author-123")

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(&models.PullRequest{
				ID:       prID,
				AuthorID: authorID,
				Status:   domain.PRStatusOpen,
			}, nil)

		prStorage.EXPECT().DeletePRReviewerInstance(gomock.Any(), prID, reviewerIDToChange).
			Return(nil)

		teamStorage.EXPECT().GetTeamByUserID(gomock.Any(), reviewerIDToChange).
			Return(nil, repositoryerrs.ErrNotFound)

		err := uc.ReassignReviewer(context.Background(), prID, reviewerIDToChange)
		So(err, ShouldEqual, errs.ErrUserHasNoTeam)
	})
}

func TestReAssign_AllUsersAreReviewers(t *testing.T) {
	Convey("ReAssign: all active users are already reviewers", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("100")
		reviewerIDToChange := domain.UserID("200")
		authorID := domain.UserID("author-123")
		teamID := domain.TeamID(1)
		otherReviewer := domain.UserID("other-reviewer-300")

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(&models.PullRequest{
				ID:       prID,
				AuthorID: authorID,
				Status:   domain.PRStatusOpen,
			}, nil)

		prStorage.EXPECT().DeletePRReviewerInstance(gomock.Any(), prID, reviewerIDToChange).
			Return(nil)

		teamStorage.EXPECT().GetTeamByUserID(gomock.Any(), reviewerIDToChange).
			Return(&models.Team{
				ID:       teamID,
				TeamName: "test-team",
			}, nil)

		userStorage.EXPECT().GetActiveUsersByTeam(gomock.Any(), teamID).
			Return([]models.User{
				{ID: authorID, StatusActivity: true},
				{ID: reviewerIDToChange, StatusActivity: true},
				{ID: otherReviewer, StatusActivity: true},
			}, nil)

		prStorage.EXPECT().GetReviewersFromPR(gomock.Any(), prID).
			Return([]models.User{
				{ID: reviewerIDToChange, StatusActivity: true},
				{ID: otherReviewer, StatusActivity: true},
			}, nil)

		err := uc.ReassignReviewer(context.Background(), prID, reviewerIDToChange)
		So(err, ShouldEqual, errs.ErrNoAvailableActiveUserToAssign)
	})
}

func TestReAssign_CreateReviewerFails(t *testing.T) {
	Convey("ReAssign: failed to create new reviewer instance", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("100")
		reviewerIDToChange := domain.UserID("200")
		authorID := domain.UserID("author-123")
		teamID := domain.TeamID(1)
		newReviewerID := domain.UserID("new-reviewer-300")

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(&models.PullRequest{
				ID:       prID,
				AuthorID: authorID,
				Status:   domain.PRStatusOpen,
			}, nil)

		prStorage.EXPECT().DeletePRReviewerInstance(gomock.Any(), prID, reviewerIDToChange).
			Return(nil)

		teamStorage.EXPECT().GetTeamByUserID(gomock.Any(), reviewerIDToChange).
			Return(&models.Team{
				ID:       teamID,
				TeamName: "test-team",
			}, nil)

		userStorage.EXPECT().GetActiveUsersByTeam(gomock.Any(), teamID).
			Return([]models.User{
				{ID: authorID, StatusActivity: true},
				{ID: newReviewerID, StatusActivity: true},
			}, nil)

		prStorage.EXPECT().GetReviewersFromPR(gomock.Any(), prID).
			Return([]models.User{
				{ID: reviewerIDToChange, StatusActivity: true},
			}, nil)

		prStorage.EXPECT().CreatePRReviewerInstance(gomock.Any(), prID, newReviewerID).
			Return(errors.New("database error"))

		err := uc.ReassignReviewer(context.Background(), prID, reviewerIDToChange)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "database error")
	})
}

func TestReAssign_OnlyAuthorInTeam(t *testing.T) {
	Convey("ReAssign: only author in team", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("100")
		reviewerIDToChange := domain.UserID("200")
		authorID := domain.UserID("author-123")
		teamID := domain.TeamID(1)

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(&models.PullRequest{
				ID:       prID,
				AuthorID: authorID,
				Status:   domain.PRStatusOpen,
			}, nil)

		prStorage.EXPECT().DeletePRReviewerInstance(gomock.Any(), prID, reviewerIDToChange).
			Return(nil)

		teamStorage.EXPECT().GetTeamByUserID(gomock.Any(), reviewerIDToChange).
			Return(&models.Team{
				ID:       teamID,
				TeamName: "test-team",
			}, nil)

		userStorage.EXPECT().GetActiveUsersByTeam(gomock.Any(), teamID).
			Return([]models.User{
				{ID: authorID, StatusActivity: true},
			}, nil)

		prStorage.EXPECT().GetReviewersFromPR(gomock.Any(), prID).
			Return([]models.User{
				{ID: reviewerIDToChange, StatusActivity: true},
			}, nil)

		err := uc.ReassignReviewer(context.Background(), prID, reviewerIDToChange)
		So(err, ShouldEqual, errs.ErrNoAvailableActiveUserToAssign)
	})
}

func TestReAssignNoActiveUsersInTeam(t *testing.T) {
	Convey("ReAssign: no active users in team", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("100")
		reviewerIDToChange := domain.UserID("200")
		authorID := domain.UserID("author-123")
		teamID := domain.TeamID(1)

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(&models.PullRequest{
				ID:       prID,
				AuthorID: authorID,
				Status:   domain.PRStatusOpen,
			}, nil)

		prStorage.EXPECT().DeletePRReviewerInstance(gomock.Any(), prID, reviewerIDToChange).
			Return(nil)

		teamStorage.EXPECT().GetTeamByUserID(gomock.Any(), reviewerIDToChange).
			Return(&models.Team{
				ID:       teamID,
				TeamName: "test-team",
			}, nil)

		userStorage.EXPECT().GetActiveUsersByTeam(gomock.Any(), teamID).
			Return([]models.User{
				{ID: authorID, StatusActivity: true},
				{ID: "user-1", StatusActivity: false},
				{ID: "user-2", StatusActivity: false},
			}, nil)

		prStorage.EXPECT().GetReviewersFromPR(gomock.Any(), prID).
			Return([]models.User{
				{ID: authorID, StatusActivity: true},
			}, nil)

		err := uc.ReassignReviewer(context.Background(), prID, reviewerIDToChange)
		So(err, ShouldEqual, errs.ErrNoAvailableActiveUserToAssign)
	})
}

func TestReAssignAlreadyMerged(t *testing.T) {
	Convey("ReAssign: already merged", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("100")
		authorID := domain.UserID("author-123")
		reviewerToChangeID := domain.UserID("132")

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(&models.PullRequest{
				ID:       prID,
				AuthorID: authorID,
				Status:   domain.PRStatusMerged,
			}, nil)

		err := uc.ReassignReviewer(context.Background(), prID, reviewerToChangeID)
		So(err, ShouldEqual, errs.ErrPRAlreadyMerged)
	})
}
